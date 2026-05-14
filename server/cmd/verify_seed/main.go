package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 连接数据库
	dsn := "root:root123456@tcp(127.0.0.1:3306)/fz_yyc_api?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal("无法连接到数据库:", pingErr)
	}

	fmt.Println("✓ 数据库连接成功")

	// 验证服务商管理员密码
	var adminPwd string
	queryErr := db.QueryRow("SELECT password FROM service_provider_admins WHERE username = ?", "admin").Scan(&adminPwd)
	if queryErr != nil {
		log.Fatal("查询服务商管理员失败:", queryErr)
	}

	err = bcrypt.CompareHashAndPassword([]byte(adminPwd), []byte("admin123"))
	if err != nil {
		fmt.Println("✗ 服务商管理员密码验证失败: admin123")
	} else {
		fmt.Println("✓ 服务商管理员密码验证成功: admin123")
	}

	// 验证商家员工密码
	var merchantPwd string
	queryErr = db.QueryRow("SELECT password FROM merchant_staffs WHERE username = ?", "merchant").Scan(&merchantPwd)
	if queryErr != nil {
		log.Fatal("查询商家员工失败:", queryErr)
	}

	err = bcrypt.CompareHashAndPassword([]byte(merchantPwd), []byte("merchant123"))
	if err != nil {
		fmt.Println("✗ 商家员工密码验证失败: merchant123")
	} else {
		fmt.Println("✓ 商家员工密码验证成功: merchant123")
	}

	fmt.Println("\n数据完整性检查:")

	type countCheck struct {
		label    string
		query    string
		minCount int
	}

	checks := []countCheck{
		{label: "服务商数量", query: "SELECT COUNT(*) FROM service_providers", minCount: 1},
		{label: "服务商管理员数量", query: "SELECT COUNT(*) FROM service_provider_admins", minCount: 1},
		{label: "商家数量", query: "SELECT COUNT(*) FROM merchants", minCount: 10},
		{label: "商家员工数量", query: "SELECT COUNT(*) FROM merchant_staffs", minCount: 19},
		{label: "配送设置数量", query: "SELECT COUNT(*) FROM merchant_delivery_settings", minCount: 10},
		{label: "已配置子商户号的商家数量", query: "SELECT COUNT(*) FROM merchants WHERE COALESCE(sub_mch_id, '') <> ''", minCount: 9},
		{label: "支付配置完成商家数量", query: "SELECT COUNT(*) FROM merchants WHERE payment_config_status = 1", minCount: 9},
		{label: "分类数量", query: "SELECT COUNT(*) FROM categories", minCount: 40},
		{label: "商品数量", query: "SELECT COUNT(*) FROM products", minCount: 80},
		{label: "商品规格数量", query: "SELECT COUNT(*) FROM product_specs", minCount: 80},
		{label: "C端用户数量", query: "SELECT COUNT(*) FROM users", minCount: 30},
		{label: "收货地址数量", query: "SELECT COUNT(*) FROM user_addresses", minCount: 40},
		{label: "访问记录数量", query: "SELECT COUNT(*) FROM user_visits", minCount: 120},
		{label: "行为事件数量", query: "SELECT COUNT(*) FROM user_behavior_events", minCount: 200},
		{label: "订单数量", query: "SELECT COUNT(*) FROM orders", minCount: 120},
		{label: "订单商品数量", query: "SELECT COUNT(*) FROM order_items", minCount: 240},
		{label: "退款数量", query: "SELECT COUNT(*) FROM refunds", minCount: 10},
		{label: "平台活动数量", query: "SELECT COUNT(*) FROM activities", minCount: 4},
		{label: "系统公告数量", query: "SELECT COUNT(*) FROM announcements", minCount: 5},
		{label: "分账记录数量", query: "SELECT COUNT(*) FROM merchant_profit_sharing_records", minCount: 3},
		{label: "年费记录数量", query: "SELECT COUNT(*) FROM merchant_fees", minCount: 20},
		{label: "费率记录数量", query: "SELECT COUNT(*) FROM merchant_rates", minCount: 20},
		{label: "优惠券数量", query: "SELECT COUNT(*) FROM coupons", minCount: 20},
		{label: "优惠券领取记录数量", query: "SELECT COUNT(*) FROM coupon_records", minCount: 20},
		{label: "云打印机数量", query: "SELECT COUNT(*) FROM cloud_printers", minCount: 10},
		{label: "打印日志数量", query: "SELECT COUNT(*) FROM print_logs", minCount: 20},
	}

	for _, check := range checks {
		var count int
		if err := db.QueryRow(check.query).Scan(&count); err != nil {
			log.Fatalf("检查%s失败: %v", check.label, err)
		}
		if count >= check.minCount {
			fmt.Printf("✓ %s: %d (期望 >= %d)\n", check.label, count, check.minCount)
		} else {
			fmt.Printf("✗ %s: %d (期望 >= %d)\n", check.label, count, check.minCount)
		}
	}

	fmt.Println("\n关键关系检查:")

	relationshipChecks := []countCheck{
		{
			label:    "每个商家都有配送设置",
			query:    "SELECT COUNT(*) FROM merchants m LEFT JOIN merchant_delivery_settings s ON s.merchant_id = m.id WHERE s.id IS NULL",
			minCount: 0,
		},
		{
			label:    "至少存在分账成功记录",
			query:    "SELECT COUNT(*) FROM merchant_profit_sharing_records WHERE status = 1",
			minCount: 1,
		},
		{
			label:    "至少存在分账失败或跳过记录",
			query:    "SELECT COUNT(*) FROM merchant_profit_sharing_records WHERE status IN (2, 3)",
			minCount: 2,
		},
		{
			label:    "至少存在使用 1112649854 的商家支付配置",
			query:    "SELECT COUNT(*) FROM merchants WHERE sub_mch_id = '1112649854' AND payment_config_status = 1",
			minCount: 1,
		},
		{
			label:    "至少存在带商品的商家",
			query:    "SELECT COUNT(DISTINCT merchant_id) FROM products",
			minCount: 10,
		},
		{
			label:    "至少存在带订单项的订单",
			query:    "SELECT COUNT(DISTINCT order_id) FROM order_items",
			minCount: 120,
		},
		{
			label:    "至少存在使用优惠券的订单",
			query:    "SELECT COUNT(*) FROM coupon_records WHERE status = 1 AND order_id IS NOT NULL",
			minCount: 10,
		},
	}

	for _, check := range relationshipChecks {
		var count int
		if err := db.QueryRow(check.query).Scan(&count); err != nil {
			log.Fatalf("检查%s失败: %v", check.label, err)
		}
		if check.minCount == 0 {
			if count == 0 {
				fmt.Printf("✓ %s: %d\n", check.label, count)
			} else {
				fmt.Printf("✗ %s: %d\n", check.label, count)
			}
			continue
		}
		if count >= check.minCount {
			fmt.Printf("✓ %s: %d (期望 >= %d)\n", check.label, count, check.minCount)
		} else {
			fmt.Printf("✗ %s: %d (期望 >= %d)\n", check.label, count, check.minCount)
		}
	}

	fmt.Println("\n初始化与模拟数据验证完成！")
}
