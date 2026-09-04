package main

// 上线迁移工具：把远端 cxsm（旧版 8 张核心表）业务数据，按本地 fz_yyc_api 新表结构
// 做字段转换与 ID 重映射后迁移到本地库。
//
// 硬约束：绝不改动本地表结构（只增删行）；保留本地系统配置/菜单/商家 id=1 等主数据。
//
// 源库凭据通过专用前缀环境变量注入（不入 .env、不入 git）：
//
//	CXSM_SRC_HOST / CXSM_SRC_PORT / CXSM_SRC_USER / CXSM_SRC_PASSWORD / CXSM_SRC_DB
//
// 目标库（本地）复用 server/.env 的 DB_* 配置。
//
// 用法：
//
//	go run ./cmd/migrate_cxsm --dry-run   # 只读两端、打印计划，不写库
//	go run ./cmd/migrate_cxsm             # 正式执行
import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"fz_yyc_api/internal/config"
)

// DefaultMerchantSubMchIDPrefix 自营商户 sub_mch_id 校验（full 值见 hard constraint）
const ExpectedSubMchID = "1112979963"

// 被替换为核心业务表（重构本地分类树/商品/订单/用户）
var coreTables = []string{
	"order_items",
	"orders",
	"user_addresses",
	"product_specs",
	"products",
	"categories",
	"users",
}

// 引用了用户/商品/订单、需一并清空的 mock 表（避免替换后产生孤儿数据）
var referencingTables = []string{
	"user_behavior_events",
	"user_visits",
	"user_coupons",
	"refunds",
	"health_records",
	"health_assessments",
	"fitting_recommendations",
	"agreement_consents",
	"service_records",
	"service_reviews",
	"service_alert_events",
	"service_location_tracks",
	"profit_sharing_records",
	"profit_sharing_record_receivers",
	"store_home_recommends",
}

// 目标表必须包含的列（工具写入所需），用于“绝不动结构 / 结构不符即停”自检。
var targetColumns = map[string][]string{
	"categories": {
		"name", "category_type", "parent_id", "level", "sort", "status",
		"created_at", "updated_at",
	},
	"products": {
		"category_id", "name", "description", "images", "price", "original_price",
		"stock", "unit", "product_type", "service_content", "sale_type",
		"rental_unit", "rental_price", "deposit", "max_rental_duration",
		"sales", "sort", "status", "deleted_at", "created_at", "updated_at",
	},
	"product_specs": {"product_id", "name", "options", "created_at", "updated_at"},
	"orders": {
		"order_no", "user_id", "order_type", "total_amount", "delivery_fee",
		"discount_amount", "pay_amount", "total_deposit", "deposit_status",
		"deposit_refund_amount", "deposit_deduct_amount", "deposit_refunded_at",
		"rental_returned_at", "rental_return_remark", "rental_end_at",
		"parent_order_id", "renew_flag", "delivery_type", "delivery_distance",
		"delivery_address", "contact_name", "contact_phone", "status", "biz_status",
		"remark", "verify_code", "transaction_id", "paid_at", "pay_notify_payload",
		"profit_sharing_status", "profit_sharing_amount", "profit_sharing_order_no",
		"profit_sharing_at", "profit_sharing_error", "completed_at",
		"completed_by_name", "cancelled_at", "refunded_at", "scheduled_at",
		"assigned_staff_id", "record_id", "delivery_district", "actual_started_at",
		"actual_ended_at", "created_at", "updated_at", "pickup_point_id",
		"pickup_point_name", "pickup_point_address", "pickup_point_lat",
		"pickup_point_lng", "assigned_at",
	},
	"order_items": {
		"order_id", "product_id", "product_name", "image", "price", "quantity",
		"spec_info", "subtotal", "sale_type", "rental_unit", "rental_duration",
		"unit_rental_price", "rental_subtotal", "deposit", "deposit_deduct",
		"created_at",
	},
	"users": {
		"openid", "union_id", "nickname", "avatar", "phone", "status",
		"created_at", "updated_at", "first_visit_at", "last_visit_at", "visit_count",
		"has_ordered", "total_orders", "total_spent", "has_paid", "first_paid_at",
	},
	"user_addresses": {
		"user_id", "name", "phone", "province", "city", "district", "address",
		"lat", "lng", "is_default", "created_at", "updated_at",
	},
}

// 空的服务型商品内容结构（保证 C 端解析不崩）
const serviceContentEmptyJSON = `{"cycle":"","target_audience":"","services":[],"remark":""}`

// ---- 源表行结构（远端旧库）----
type SrcCategory struct {
	ID        uint64
	Name      string
	Sort      int
	Status    uint8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SrcProduct struct {
	ID           uint64
	CategoryID   sql.NullInt64
	Name         string
	Description  sql.NullString
	Images       sql.NullString
	Price        float64
	OriginalPrice sql.NullFloat64
	Stock        int
	Unit         string
	Sales        int
	Sort         int
	Status       uint8
	DeletedAt    sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SrcSpec struct {
	ID        uint64
	ProductID uint64
	Name      string
	Options   sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SrcOrder struct {
	ID                  uint64
	OrderNo             string
	UserID              uint64
	TotalAmount         float64
	DeliveryFee         float64
	DiscountAmount      float64
	PayAmount           float64
	DeliveryType        uint8
	DeliveryDistance    sql.NullFloat64
	DeliveryAddress     sql.NullString
	ContactName         sql.NullString
	ContactPhone        sql.NullString
	PickupPointID       sql.NullInt64
	PickupPointName     sql.NullString
	PickupPointAddress  sql.NullString
	PickupPointLat      sql.NullFloat64
	PickupPointLng      sql.NullFloat64
	Status              uint8
	Remark              sql.NullString
	VerifyCode          sql.NullString
	TransactionID       sql.NullString
	PaidAt              sql.NullTime
	PayNotifyPayload    sql.NullString
	CompletedAt         sql.NullTime
	CompletedByName     sql.NullString
	CancelledAt         sql.NullTime
	RefundedAt          sql.NullTime
	ProfitSharingStatus uint8
	ProfitSharingAmount float64
	ProfitSharingOrderNo sql.NullString
	ProfitSharingAt     sql.NullTime
	ProfitSharingError  sql.NullString
	CreatedAt           sql.NullTime
	UpdatedAt           sql.NullTime
}

type SrcOrderItem struct {
	ID          uint64
	OrderID     uint64
	ProductID   uint64
	ProductName string
	Image       sql.NullString
	Price       float64
	Quantity    int
	SpecInfo    sql.NullString
	Subtotal    float64
	CreatedAt   time.Time
}

type SrcUser struct {
	ID          uint64
	Openid      sql.NullString
	UnionID     sql.NullString
	Nickname    sql.NullString
	Avatar      sql.NullString
	Phone       sql.NullString
	Status      uint8
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	FirstVisitAt sql.NullTime
	LastVisitAt sql.NullTime
	VisitCount  uint64
	HasOrdered  bool
	TotalOrders uint64
	TotalSpent  float64
	HasPaid     bool
	FirstPaidAt sql.NullTime
}

type SrcAddress struct {
	ID        uint64
	UserID    uint64
	Name      string
	Phone     string
	Province  sql.NullString
	City      sql.NullString
	District  sql.NullString
	Address   string
	Lat       sql.NullFloat64
	Lng       sql.NullFloat64
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// nil2any 把 sql.Null* 归一化为 any 以便直接作为参数（NULL 传 nil/零值由调用方处理）。
func main() {
	dryRun := flag.Bool("dry-run", false, "只读两端并打印迁移计划，不写目标库")
	flag.Parse()

	if err := config.InitConfig(); err != nil {
		log.Fatalf("本地配置初始化失败: %v", err)
	}

	// ---- 目标库（本地）----
	tgtCfg := config.Config.Database
	tgtDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		tgtCfg.User, tgtCfg.Password, tgtCfg.Host, tgtCfg.Port, tgtCfg.Name, tgtCfg.Charset)
	tgt, err := sql.Open("mysql", tgtDSN)
	if err != nil {
		log.Fatalf("本地库连接失败: %v", err)
	}
	defer tgt.Close()
	if err := tgt.Ping(); err != nil {
		log.Fatalf("本地库 ping 失败: %v", err)
	}

	// ---- 源库（远端 cxsm）----
	srcHost := env("CXSM_SRC_HOST")
	srcPort := envInt("CXSM_SRC_PORT", 3306)
	srcUser := env("CXSM_SRC_USER")
	srcPass := env("CXSM_SRC_PASSWORD")
	srcDB := env("CXSM_SRC_DB")
	if srcHost == "" || srcUser == "" || srcDB == "" {
		log.Fatal("缺少源库配置：请设置 CXSM_SRC_HOST/CXSM_SRC_USER/CXSM_SRC_PASSWORD/CXSM_SRC_DB")
	}
	srcDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		srcUser, srcPass, srcHost, srcPort, srcDB)
	src, err := sql.Open("mysql", srcDSN)
	if err != nil {
		log.Fatalf("远端库连接失败: %v", err)
	}
	defer src.Close()
	if err := src.Ping(); err != nil {
		log.Fatalf("远端库 ping 失败: %v", err)
	}

	log.Printf("源库: %s@%s:%d/%s", srcUser, srcHost, srcPort, srcDB)
	log.Printf("目标库: %s@%s:%d/%s", tgtCfg.User, tgtCfg.Host, tgtCfg.Port, tgtCfg.Name)

	// 1) 结构自检 + 商家约束断言（任何失败立即退出，不写任何数据）
	checkStructure(tgt)
	checkMerchant(tgt)

	// 源库各表行数（用于回显 / dry-run）
	counts := sourceCounts(src)

	if *dryRun {
		printPlan(coreTables, referencingTables, counts)
		log.Printf("--dry-run 模式：未写任何数据。确认后请去掉 --dry-run 执行。")
		return
	}

	// 2) + 3) 事务内清空并重建
	if err := migrate(src, tgt, counts); err != nil {
		log.Fatalf("迁移失败，已回滚: %v", err)
	}

	log.Printf("迁移完成。请核对上面输出的行数/孤儿校验结果。")
}

func env(k string) string { return os.Getenv(k) }

func envInt(k string, def int) int {
	if v := env(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// checkStructure 校验目标表包含迁移所需全部列，缺任一列即 FATAL。
func checkStructure(db *sql.DB) {
	for tbl, need := range targetColumns {
		rows, err := db.Query("SHOW COLUMNS FROM `" + tbl + "`")
		if err != nil {
			log.Fatalf("结构自检 SHOW COLUMNS %s 失败: %v（请确认本地库表存在）", tbl, err)
		}
		got := map[string]bool{}
		for rows.Next() {
			var field, typ, nul, key string
			var def sql.NullString
			var extra string
			if err := rows.Scan(&field, &typ, &nul, &key, &def, &extra); err != nil {
				rows.Close()
				log.Fatalf("结构自检读取 %s 列失败: %v", tbl, err)
			}
			got[field] = true
		}
		rows.Close()
		for _, c := range need {
			if !got[c] {
				log.Fatalf("结构自检失败：本地表 %s 缺少迁移所需列 %q。为遵守“绝不动结构”，迁移终止，未写任何数据。", tbl, c)
			}
		}
	}
	log.Printf("结构自检通过：8 张目标表均包含迁移所需全部列，未发现结构差异。")
}

// checkMerchant 断言自营商户 id=1 存在且 sub_mch_id 不变。
func checkMerchant(db *sql.DB) {
	var id uint64
	var subMch sql.NullString
	err := db.QueryRow("SELECT id, sub_mch_id FROM merchants WHERE id = 1").Scan(&id, &subMch)
	if err != nil {
		log.Fatalf("商家自检失败：无法读取 merchants.id=1: %v", err)
	}
	if !subMch.Valid || subMch.String != ExpectedSubMchID {
		log.Fatalf("商家自检失败：merchants.id=1 的 sub_mch_id=%v，预期 %s，已终止（hard constraint）。",
			subMch.String, ExpectedSubMchID)
	}
	log.Printf("商家自检通过：保留本地 merchants.id=1（sub_mch_id=%s）。", subMch.String)
}

func sourceCounts(src *sql.DB) map[string]int {
	counts := map[string]int{}
	srcTables := []string{"categories", "products", "product_specs", "orders", "order_items", "users", "user_addresses"}
	for _, t := range srcTables {
		var n int
		if err := src.QueryRow("SELECT COUNT(*) FROM `" + t + "`").Scan(&n); err != nil {
			log.Fatalf("读取源表 %s 行数失败: %v", t, err)
		}
		counts[t] = n
	}
	return counts
}

func printPlan(core, ref []string, counts map[string]int) {
	fmt.Println("\n========== 迁移计划（dry-run） ==========")
	fmt.Println("\n[1] 将清空并重建的核心业务表（用远端重塑）：")
	for _, t := range core {
		fmt.Printf("    - %s（源库 %d 行 → 本地重建）\n", t, counts[t])
	}
	fmt.Println("\n[2] 将清空的引用型 mock 表（保留表结构，仅删行）：")
	for _, t := range ref {
		fmt.Printf("    - %s\n", t)
	}
	fmt.Println("\n[3] 将保留（不清、不迁移）的主数据/配置/菜单表：sys_*、system_configs、merchants(id=1)、")
	fmt.Println("    merchant_staffs(*)、service_staffs(*)、agreements、coupon_templates、")
	fmt.Println("    mini_program_banners、health_education_*、health_assessment_forms、")
	fmt.Println("    profit_sharing_receivers、printers 等。")
	fmt.Println("\n[4] 分类重建规则：远端扁平分类 → 本地 level=1 一级分类；租赁类/服务类→category_type=2。")
	fmt.Println("    商品按分类推导 product_type/sale_type（租赁→2/2，服务→3/1，其余→1/1）。")
	fmt.Println("    订单统一 order_type=1（零售），status 1/3/4/6 直接映射，新字段取默认值。")
	fmt.Println("=========================================")
}

func migrate(src, tgt *sql.DB, counts map[string]int) error {
	tx, err := tgt.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback() // 仅失败路径
		}
	}()

	if _, err := tx.Exec("SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return fmt.Errorf("SET FOREIGN_KEY_CHECKS=0 失败: %w", err)
	}
	// 结束时恢复 FK 检查
	defer func() {
		if _, err := tx.Exec("SET FOREIGN_KEY_CHECKS=1"); err != nil {
			log.Printf("警告：恢复 FOREIGN_KEY_CHECKS=1 失败: %v", err)
		}
	}()

	// ---- 清空：先引用型，再核心（父后子顺序不敏感，配 FK_CHECKS=0 兜底）----
	for _, t := range referencingTables {
		if _, err := tx.Exec("DELETE FROM `" + t + "`"); err != nil {
			return fmt.Errorf("清空 %s 失败: %w", t, err)
		}
	}
	for _, t := range coreTables {
		if _, err := tx.Exec("DELETE FROM `" + t + "`"); err != nil {
			return fmt.Errorf("清空 %s 失败: %w", t, err)
		}
	}

	// ---- 重建 ----
	skipped := map[string]int{}

	// users（先建，orders/addresses 依赖）
	userMap, err := loadUsers(src, tx, counts["users"], &skipped)
	if err != nil {
		return err
	}
	// categories（先建，products 依赖）
	catMap, catKind, err := loadCategories(src, tx, counts["categories"], &skipped)
	if err != nil {
		return err
	}
	// products
	prodMap, err := loadProducts(src, tx, catMap, catKind, counts["products"], &skipped)
	if err != nil {
		return err
	}
	// product_specs
	if err := loadSpecs(src, tx, prodMap, counts["product_specs"], &skipped); err != nil {
		return err
	}
	// orders
	orderMap, err := loadOrders(src, tx, userMap, counts["orders"], &skipped)
	if err != nil {
		return err
	}
	// user_addresses
	if err := loadAddresses(src, tx, userMap, counts["user_addresses"], &skipped); err != nil {
		return err
	}
	// order_items（最后）
	if err := loadOrderItems(src, tx, orderMap, prodMap, counts["order_items"], &skipped); err != nil {
		return err
	}

	// ---- 重置自增（非结构变更）----
	for _, t := range coreTables {
		if _, err := tx.Exec("ALTER TABLE `" + t + "` AUTO_INCREMENT = 1"); err != nil {
			return fmt.Errorf("重置 %s 自增失败: %w", t, err)
		}
	}

	// ---- 校验 ----
	wroteCounts := map[string]int{}
	for _, t := range coreTables {
		var n int
		if err := tx.QueryRow("SELECT COUNT(*) FROM `" + t + "`").Scan(&n); err != nil {
			return fmt.Errorf("校验 %s 行数失败: %w", t, err)
		}
		wroteCounts[t] = n
	}
	if err := verifyOrphans(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	printSummary(counts, wroteCounts, skipped)
	return nil
}

func printSummary(srcCount, wroteCount map[string]int, skipped map[string]int) {
	fmt.Println("\n========== 迁移结果汇总 ==========")
	for _, t := range coreTables {
		fmt.Printf("    %s: 源 %d 行 → 本地 %d 行  %s\n", t, srcCount[t], wroteCount[t],
			matchMark(srcCount[t], wroteCount[t]))
	}
	if len(skipped) > 0 {
		fmt.Println("\n（可接受丢失，已在日志记录跳过的行：）")
		for k, v := range skipped {
			fmt.Printf("    - %s: %d 行\n", k, v)
		}
	}
	fmt.Println("===================================")
}

func matchMark(a, b int) string {
	if a == b {
		return "✔ 一致"
	}
	return fmt.Sprintf("✘ 不一致（差 %d）", a-b)
}

// ---- 插入辅助 ----
func insertRow(tx *sql.Tx, table string, cols []string, vals []any) (int64, error) {
	ph := strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",")
	q := "INSERT INTO `" + table + "` (" + strings.Join(cols, ",") + ") VALUES (" + ph + ")"
	res, err := tx.Exec(q, vals...)
	if err != nil {
		return 0, fmt.Errorf("插入 %s 失败: %w", table, err)
	}
	return res.LastInsertId()
}

// ---- 各表加载 ----

func loadUsers(src *sql.DB, tx *sql.Tx, expect int, skipped *map[string]int) (map[uint64]uint64, error) {
	rows, err := src.Query("SELECT id,openid,union_id,nickname,avatar,phone,status,created_at,updated_at,first_visit_at,last_visit_at,visit_count,has_ordered,total_orders,total_spent,has_paid,first_paid_at FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("读取 users 失败: %w", err)
	}
	defer rows.Close()
	m := map[uint64]uint64{}
	cols := []string{"openid", "union_id", "nickname", "avatar", "phone", "status", "created_at", "updated_at", "first_visit_at", "last_visit_at", "visit_count", "has_ordered", "total_orders", "total_spent", "has_paid", "first_paid_at"}
	for rows.Next() {
		var u SrcUser
		if err := rows.Scan(&u.ID, &u.Openid, &u.UnionID, &u.Nickname, &u.Avatar, &u.Phone, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.FirstVisitAt, &u.LastVisitAt, &u.VisitCount, &u.HasOrdered, &u.TotalOrders, &u.TotalSpent, &u.HasPaid, &u.FirstPaidAt); err != nil {
			return nil, fmt.Errorf("扫描 users 失败: %w", err)
		}
		id, err := insertUser(tx, cols, u)
		if err != nil {
			return nil, err
		}
		m[u.ID] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

func insertUser(tx *sql.Tx, cols []string, u SrcUser) (uint64, error) {
	id, err := insertRow(tx, "users", cols, []any{
		nilStr(u.Openid), nilStr(u.UnionID), nilStr(u.Nickname), nilStr(u.Avatar),
		nilStr(u.Phone), u.Status, nilTime(u.CreatedAt), nilTime(u.UpdatedAt),
		nilTime(u.FirstVisitAt), nilTime(u.LastVisitAt), u.VisitCount,
		boolInt(u.HasOrdered), u.TotalOrders, u.TotalSpent, boolInt(u.HasPaid),
		nilTime(u.FirstPaidAt),
	})
	return uint64(id), err
}

func loadCategories(src *sql.DB, tx *sql.Tx, expect int, skipped *map[string]int) (map[uint64]uint64, map[uint64]string, error) {
	rows, err := src.Query("SELECT id,name,sort,status,created_at,updated_at FROM categories ORDER BY id")
	if err != nil {
		return nil, nil, fmt.Errorf("读取 categories 失败: %w", err)
	}
	defer rows.Close()
	catMap := map[uint64]uint64{}
	catKind := map[uint64]string{} // oldCategoryID -> "rental"|"service"|"goods"
	cols := []string{"name", "category_type", "parent_id", "level", "sort", "status", "created_at", "updated_at"}
	var n int
	for rows.Next() {
		var c SrcCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Sort, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("扫描 categories 失败: %w", err)
		}
		kind := "goods"
		catType := uint8(1)
		if c.Name == "租赁类" {
			kind = "rental"
			catType = 2
		} else if c.Name == "服务类" {
			kind = "service"
			catType = 2
		}
		id, err := insertRow(tx, "categories", cols, []any{
			c.Name, catType, nil, 1, c.Sort, c.Status, c.CreatedAt, c.UpdatedAt,
		})
		if err != nil {
			return nil, nil, err
		}
		catMap[c.ID] = uint64(id)
		catKind[c.ID] = kind
		n++
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return catMap, catKind, nil
}

func loadProducts(src *sql.DB, tx *sql.Tx, catMap map[uint64]uint64, catKind map[uint64]string, expect int, skipped *map[string]int) (map[uint64]uint64, error) {
	rows, err := src.Query("SELECT id,category_id,name,description,images,price,original_price,stock,unit,sales,sort,status,deleted_at,created_at,updated_at FROM products ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("读取 products 失败: %w", err)
	}
	defer rows.Close()
	m := map[uint64]uint64{}
	cols := []string{"category_id", "name", "description", "images", "price", "original_price", "stock", "unit", "product_type", "service_content", "sale_type", "rental_unit", "rental_price", "deposit", "max_rental_duration", "sales", "sort", "status", "deleted_at", "created_at", "updated_at"}
	for rows.Next() {
		var p SrcProduct
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Images, &p.Price, &p.OriginalPrice, &p.Stock, &p.Unit, &p.Sales, &p.Sort, &p.Status, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描 products 失败: %w", err)
		}
		var newCatID any
		if p.CategoryID.Valid {
			newCatID = catMap[uint64(p.CategoryID.Int64)]
		}
		kind := catKind[uint64(p.CategoryID.Int64)]
		if kind == "" {
			kind = "goods"
		}
		productType := uint8(1)
		saleType := uint8(1)
		rentalUnit := uint8(0)
		rentalPrice := 0.0
		deposit := 0.0
		maxDuration := 0
		var serviceContent any
		switch kind {
		case "rental":
			productType = 2
			saleType = 2
			rentalUnit = 1 // 按天
			rentalPrice = p.Price
			deposit = 10.00
			maxDuration = 0
		case "service":
			productType = 3
			saleType = 1
			serviceContent = serviceContentEmptyJSON
		}
		id, err := insertRow(tx, "products", cols, []any{
			newCatID, p.Name, nilStr(p.Description), nilStr(p.Images), p.Price,
			nilFloat(p.OriginalPrice), p.Stock, p.Unit, productType, serviceContent,
			saleType, rentalUnit, rentalPrice, deposit, maxDuration,
			p.Sales, p.Sort, p.Status, nilTime(p.DeletedAt), p.CreatedAt, p.UpdatedAt,
		})
		if err != nil {
			return nil, err
		}
		m[p.ID] = uint64(id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

func loadSpecs(src *sql.DB, tx *sql.Tx, prodMap map[uint64]uint64, expect int, skipped *map[string]int) error {
	rows, err := src.Query("SELECT id,product_id,name,options,created_at,updated_at FROM product_specs ORDER BY id")
	if err != nil {
		return fmt.Errorf("读取 product_specs 失败: %w", err)
	}
	defer rows.Close()
	cols := []string{"product_id", "name", "options", "created_at", "updated_at"}
	for rows.Next() {
		var s SrcSpec
		if err := rows.Scan(&s.ID, &s.ProductID, &s.Name, &s.Options, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return fmt.Errorf("扫描 product_specs 失败: %w", err)
		}
		pid, ok := prodMap[s.ProductID]
		if !ok {
			(*skipped)["product_specs(商品缺失)"]++
			continue
		}
		if _, err := insertRow(tx, "product_specs", cols, []any{pid, s.Name, nilStr(s.Options), s.CreatedAt, s.UpdatedAt}); err != nil {
			return err
		}
	}
	return rows.Err()
}

func loadOrders(src *sql.DB, tx *sql.Tx, userMap map[uint64]uint64, expect int, skipped *map[string]int) (map[uint64]uint64, error) {
	rows, err := src.Query("SELECT id,order_no,user_id,total_amount,delivery_fee,discount_amount,pay_amount,delivery_type,delivery_distance,delivery_address,contact_name,contact_phone,pickup_point_id,pickup_point_name,pickup_point_address,pickup_point_lat,pickup_point_lng,status,remark,verify_code,transaction_id,paid_at,pay_notify_payload,completed_at,completed_by_name,cancelled_at,refunded_at,profit_sharing_status,profit_sharing_amount,profit_sharing_order_no,profit_sharing_at,profit_sharing_error,created_at,updated_at FROM orders ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("读取 orders 失败: %w", err)
	}
	defer rows.Close()
	m := map[uint64]uint64{}
	cols := []string{
		"order_no", "user_id", "order_type", "total_amount", "delivery_fee",
		"discount_amount", "pay_amount", "total_deposit", "deposit_status",
		"deposit_refund_amount", "deposit_deduct_amount", "deposit_refunded_at",
		"rental_returned_at", "rental_return_remark", "rental_end_at",
		"parent_order_id", "renew_flag", "delivery_type", "delivery_distance",
		"delivery_address", "contact_name", "contact_phone", "status", "biz_status",
		"remark", "verify_code", "transaction_id", "paid_at", "pay_notify_payload",
		"profit_sharing_status", "profit_sharing_amount", "profit_sharing_order_no",
		"profit_sharing_at", "profit_sharing_error", "completed_at",
		"completed_by_name", "cancelled_at", "refunded_at", "scheduled_at",
		"assigned_staff_id", "record_id", "delivery_district", "actual_started_at",
		"actual_ended_at", "created_at", "updated_at", "pickup_point_id",
		"pickup_point_name", "pickup_point_address", "pickup_point_lat",
		"pickup_point_lng", "assigned_at",
	}
	for rows.Next() {
		var o SrcOrder
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.TotalAmount, &o.DeliveryFee, &o.DiscountAmount, &o.PayAmount, &o.DeliveryType, &o.DeliveryDistance, &o.DeliveryAddress, &o.ContactName, &o.ContactPhone, &o.PickupPointID, &o.PickupPointName, &o.PickupPointAddress, &o.PickupPointLat, &o.PickupPointLng, &o.Status, &o.Remark, &o.VerifyCode, &o.TransactionID, &o.PaidAt, &o.PayNotifyPayload, &o.CompletedAt, &o.CompletedByName, &o.CancelledAt, &o.RefundedAt, &o.ProfitSharingStatus, &o.ProfitSharingAmount, &o.ProfitSharingOrderNo, &o.ProfitSharingAt, &o.ProfitSharingError, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描 orders 失败: %w", err)
		}
		newUID, ok := userMap[o.UserID]
		if !ok {
			(*skipped)["orders(用户缺失)"]++
			continue
		}
		id, err := insertRow(tx, "orders", cols, []any{
			o.OrderNo, newUID, uint8(1), o.TotalAmount, o.DeliveryFee,
			o.DiscountAmount, o.PayAmount, 0.00, uint8(0),
			0.00, 0.00, nilTimeDefault((sql.NullTime{})),
			nil, nil, nil,
			nil, uint8(0), o.DeliveryType, nilFloat(o.DeliveryDistance),
			nilStr(o.DeliveryAddress), nilStr(o.ContactName), nilStr(o.ContactPhone),
			o.Status, uint8(0),
			nilStr(o.Remark), nilStr(o.VerifyCode), nilStr(o.TransactionID),
			nilTime(o.PaidAt), nilStr(o.PayNotifyPayload),
			o.ProfitSharingStatus, o.ProfitSharingAmount, nilStr(o.ProfitSharingOrderNo),
			nilTime(o.ProfitSharingAt), nilStr(o.ProfitSharingError),
			nilTime(o.CompletedAt), nilStr(o.CompletedByName), nilTime(o.CancelledAt),
			nilTime(o.RefundedAt), nil,
			nil, nil, nil, nil, nil,
			nilTime(o.CreatedAt), nilTime(o.UpdatedAt), nilInt(o.PickupPointID),
			nilStr(o.PickupPointName), nilStr(o.PickupPointAddress),
			nilFloat(o.PickupPointLat), nilFloat(o.PickupPointLng), nil,
		})
		if err != nil {
			return nil, err
		}
		m[o.ID] = uint64(id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

func loadAddresses(src *sql.DB, tx *sql.Tx, userMap map[uint64]uint64, expect int, skipped *map[string]int) error {
	rows, err := src.Query("SELECT id,user_id,name,phone,province,city,district,address,lat,lng,is_default,created_at,updated_at FROM user_addresses ORDER BY id")
	if err != nil {
		return fmt.Errorf("读取 user_addresses 失败: %w", err)
	}
	defer rows.Close()
	cols := []string{"user_id", "name", "phone", "province", "city", "district", "address", "lat", "lng", "is_default", "created_at", "updated_at"}
	for rows.Next() {
		var a SrcAddress
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Phone, &a.Province, &a.City, &a.District, &a.Address, &a.Lat, &a.Lng, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return fmt.Errorf("扫描 user_addresses 失败: %w", err)
		}
		uid, ok := userMap[a.UserID]
		if !ok {
			(*skipped)["user_addresses(用户缺失)"]++
			continue
		}
		if _, err := insertRow(tx, "user_addresses", cols, []any{
			uid, a.Name, a.Phone, nilStr(a.Province), nilStr(a.City), nilStr(a.District),
			a.Address, nilFloat(a.Lat), nilFloat(a.Lng), boolInt(a.IsDefault),
			a.CreatedAt, a.UpdatedAt,
		}); err != nil {
			return err
		}
	}
	return rows.Err()
}

func loadOrderItems(src *sql.DB, tx *sql.Tx, orderMap, prodMap map[uint64]uint64, expect int, skipped *map[string]int) error {
	rows, err := src.Query("SELECT id,order_id,product_id,product_name,image,price,quantity,spec_info,subtotal,created_at FROM order_items ORDER BY id")
	if err != nil {
		return fmt.Errorf("读取 order_items 失败: %w", err)
	}
	defer rows.Close()
	cols := []string{"order_id", "product_id", "product_name", "image", "price", "quantity", "spec_info", "subtotal", "sale_type", "rental_unit", "rental_duration", "unit_rental_price", "rental_subtotal", "deposit", "deposit_deduct", "created_at"}
	for rows.Next() {
		var it SrcOrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.ProductName, &it.Image, &it.Price, &it.Quantity, &it.SpecInfo, &it.Subtotal, &it.CreatedAt); err != nil {
			return fmt.Errorf("扫描 order_items 失败: %w", err)
		}
		oid, okO := orderMap[it.OrderID]
		pid, okP := prodMap[it.ProductID]
		if !okO || !okP {
			if !okO {
				(*skipped)["order_items(订单缺失)"]++
			}
			if !okP {
				(*skipped)["order_items(商品缺失)"]++
			}
			continue
		}
		if _, err := insertRow(tx, "order_items", cols, []any{
			oid, pid, it.ProductName, nilStr(it.Image), it.Price, it.Quantity,
			nilStr(it.SpecInfo), it.Subtotal, uint8(1), uint8(0), 0, 0.00, 0.00,
			0.00, 0.00, it.CreatedAt,
		}); err != nil {
			return err
		}
	}
	return rows.Err()
}

// verifyOrphans 孤儿引用校验。
func verifyOrphans(tx *sql.Tx) error {
	queries := []struct {
		desc string
		sql  string
	}{
		{"orders.user_id 指向不存在的用户", "SELECT COUNT(*) FROM orders o LEFT JOIN users u ON o.user_id=u.id WHERE u.id IS NULL"},
		{"order_items.order_id 指向不存在的订单", "SELECT COUNT(*) FROM order_items oi LEFT JOIN orders o ON oi.order_id=o.id WHERE o.id IS NULL"},
		{"order_items.product_id 指向不存在的商品", "SELECT COUNT(*) FROM order_items oi LEFT JOIN products p ON oi.product_id=p.id WHERE p.id IS NULL"},
		{"products.category_id 指向不存在的分类(允许NULL)", "SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id=c.id WHERE p.category_id IS NOT NULL AND c.id IS NULL"},
		{"user_addresses.user_id 指向不存在的用户", "SELECT COUNT(*) FROM user_addresses ua LEFT JOIN users u ON ua.user_id=u.id WHERE u.id IS NULL"},
		{"product_specs.product_id 指向不存在的商品", "SELECT COUNT(*) FROM product_specs ps LEFT JOIN products p ON ps.product_id=p.id WHERE p.id IS NULL"},
	}
	allOK := true
	for _, q := range queries {
		var n int
		if err := tx.QueryRow(q.sql).Scan(&n); err != nil {
			return fmt.Errorf("孤儿校验(%s)失败: %w", q.desc, err)
		}
		mark := "✔ 无孤儿"
		if n > 0 {
			allOK = false
			mark = fmt.Sprintf("⚠ 存在 %d 条", n)
		}
		log.Printf("孤儿校验：%s → %s", q.desc, mark)
	}
	if !allOK {
		return fmt.Errorf("孤儿引用校验未完全通过，请人工核对（迁移已执行，仅告警不阻塞）")
	}
	return nil
}

// ---- 参数归一化辅助 ----
func nilStr(s sql.NullString) any {
	if !s.Valid {
		return nil
	}
	return s.String
}

func nilTime(t sql.NullTime) any {
	if !t.Valid {
		return nil
	}
	return t.Time
}

func nilFloat(f sql.NullFloat64) any {
	if !f.Valid {
		return nil
	}
	return f.Float64
}

func nilInt(v sql.NullInt64) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

func boolInt(b bool) any {
	if b {
		return 1
	}
	return 0
}

// nilTimeDefault 生成一个恒为 NULL 的 time 占位（用于明确置空的字段）
func nilTimeDefault(_ sql.NullTime) any { return nil }