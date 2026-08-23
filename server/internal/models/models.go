package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSON JSON类型，用于处理数据库JSON字段
type JSON json.RawMessage

// MarshalJSON 保持 JSON 字段按原始 JSON 输出，避免被编码成字节数组字符串。
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON 允许请求体直接反序列化到 JSON 字段。
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("json target is nil")
	}
	if data == nil {
		*j = nil
		return nil
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Scan 实现 sql.Scanner 接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*j = append((*j)[0:0], bytes...)
	return nil
}

// Value 实现 driver.Valuer 接口
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// ============================================
// 商家表 (merchants)
// 用途：商家核心信息，包含基础资料、支付配置与营业状态
// ============================================
type Merchant struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement;comment:商家ID" json:"id"`
	Name                 string    `gorm:"size:128;not null;comment:商家名称" json:"name"`
	Logo                 string    `gorm:"size:512;comment:商家Logo图片地址(七牛私有路径)" json:"logo"`
	CoverImage           string    `gorm:"size:512;comment:商家背景/封面图地址(七牛私有路径)" json:"cover_image"`
	ContactName          string    `gorm:"size:64;comment:联系人姓名" json:"contact_name"`
	ContactPhone         string    `gorm:"size:20;comment:联系电话(用于用户端拨打退款)" json:"contact_phone"`
	ContactEmail         string    `gorm:"size:128;comment:联系邮箱" json:"contact_email"`
	Address              string    `gorm:"size:256;comment:商家地址" json:"address"`
	Lat                  float64   `gorm:"type:decimal(10,6);comment:纬度" json:"lat"`
	Lng                  float64   `gorm:"type:decimal(10,6);comment:经度" json:"lng"`
	BusinessCategory     string    `gorm:"size:64;comment:经营类目" json:"business_category"`
	BusinessHours        string    `gorm:"size:64;comment:营业时间描述" json:"business_hours"`
	Announcement         string    `gorm:"type:text;comment:商家公告" json:"announcement"`
	SubMchID            string  `gorm:"size:32;comment:微信支付子商户号(线下进件后回填)" json:"sub_mch_id"`
	PaymentConfigStatus uint8   `gorm:"not null;default:0;comment:支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)" json:"payment_config_status"`
	ProfitSharingEnabled bool  `gorm:"not null;default:false;comment:是否开启自动分账: true=开启 false=关闭" json:"profit_sharing_enabled"`
	Status               uint8     `gorm:"not null;default:1;comment:营业状态: 1=营业中 0=休息中" json:"status"`
	Rating               float64   `gorm:"type:decimal(2,1);not null;default:5.0;comment:商家评分(1.0-5.0)" json:"rating"`
	SalesCount           uint      `gorm:"not null;default:0;comment:累计销量" json:"sales_count"`
	CreatedAt            time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Merchant) TableName() string {
	return "merchants"
}

// ============================================
// 商家配送设置表 (merchant_delivery_settings)
// 用途：商家配送费规则配置（单例，仅一行），包含基础运费、满减免与距离阶梯计费
// ============================================
type MerchantDeliverySettings struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement;comment:设置ID" json:"id"`
	Enabled            bool      `gorm:"not null;default:true;comment:是否启用配送费规则: true=启用 false=停用(仅控制费用规则,配送开关取merchants.takeout_enabled)" json:"enabled"`
	BaseFee            float64   `gorm:"type:decimal(10,2);not null;default:0;comment:基础配送费(元)" json:"base_fee"`
	FreeDeliveryAmount float64   `gorm:"type:decimal(10,2);not null;default:0;comment:满减免配送费金额(0=不启用)" json:"free_delivery_amount"`
	MaxDistance        uint      `gorm:"not null;default:10;comment:最大配送距离(km)" json:"max_distance"`
	DistanceRules      JSON      `gorm:"type:json;comment:距离阶梯计费规则JSON [{min_distance,max_distance,fee}]" json:"distance_rules"`
	CreatedAt          time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (MerchantDeliverySettings) TableName() string {
	return "merchant_delivery_settings"
}

// ============================================
// 商家员工表 (merchant_staffs)
// 用途：商家端登录账号，支持owner/staff角色与微信快捷登录
// ============================================
type MerchantStaff struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement;comment:员工ID" json:"id"`
	Username            string     `gorm:"size:64;not null;comment:登录用户名" json:"username"`
	Password            string     `gorm:"size:128;not null;comment:加密密码(不返回)" json:"-"`
	Name                string     `gorm:"size:64;comment:员工显示名称" json:"name"`
	Phone               string     `gorm:"size:20;comment:员工手机号" json:"phone"`
	OpenID              string     `gorm:"column:openid;size:64;comment:微信OpenID(用于快捷登录)" json:"openid"`
	UnionID             string     `gorm:"column:unionid;size:64;comment:微信UnionID" json:"unionid"`
	WechatBoundAt       *time.Time `gorm:"comment:微信绑定时间" json:"wechat_bound_at"`
	Role                string     `gorm:"size:32;not null;default:staff;comment:角色: owner=店主(可管理员工) staff=普通员工" json:"role"`
	NotifyEnabled       bool       `gorm:"not null;default:true;comment:是否开启新订单声音提醒: true=开启 false=关闭" json:"notify_enabled"`
	BrowseNotifyEnabled bool       `gorm:"not null;default:true;comment:是否开启用户进店声音提醒: true=开启 false=关闭" json:"browse_notify_enabled"`
	Status              uint8      `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	DepartmentID        *uint64    `gorm:"comment:部门ID(sys_departments.id)" json:"department_id"`
	LastLoginAt         *time.Time `gorm:"comment:最后账号密码登录时间" json:"last_login_at"`
	LastWechatLoginAt   *time.Time `gorm:"comment:最后微信快捷登录时间" json:"last_wechat_login_at"`
	CreatedAt           time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Department          *SysDepartment `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Roles               []SysRole      `gorm:"many2many:merchant_staff_roles;joinForeignKey:staff_id;joinReferences:role_id" json:"roles,omitempty"`
}

func (MerchantStaff) TableName() string {
	return "merchant_staffs"
}

// ============================================
// 商品分类表 (categories)
// 用途：商家商品分类管理，支持排序与上下架
// ============================================
type Category struct {
	ID           uint64        `gorm:"primaryKey;autoIncrement;comment:分类ID" json:"id"`
	Name         string        `gorm:"size:64;not null;comment:分类名称" json:"name"`
	ParentID     *uint64       `gorm:"index;comment:父分类ID(空=一级)" json:"parent_id"`
	Level        uint8         `gorm:"not null;default:1;comment:层级: 1=一级 2=二级 3=三级" json:"level"`
	CategoryType uint8         `gorm:"not null;default:1;comment:分类类型: 1=商品分类 2=服务分类" json:"category_type"`
	Sort         uint          `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status       uint8         `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt    time.Time     `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Children     []*Category   `gorm:"-" json:"children,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// ============================================
// 商品表 (products)
// 用途：商家商品信息，包含价格、库存、规格与上下架状态
// ============================================
type Product struct {
	ID                 uint64        `gorm:"primaryKey;autoIncrement;comment:商品ID" json:"id"`
	CategoryID         *uint64       `gorm:"index;comment:所属分类ID(可为空)" json:"category_id"`
	Name               string        `gorm:"size:128;not null;comment:商品名称" json:"name"`
	Description        string        `gorm:"type:text;comment:商品描述" json:"description"`
	Images             JSON          `gorm:"type:json;comment:商品图片URL数组JSON" json:"images"`
	Price              float64       `gorm:"type:decimal(10,2);not null;comment:商品基础价格(元)" json:"price"`
	OriginalPrice      float64       `gorm:"type:decimal(10,2);comment:划线原价(元,0表示不展示)" json:"original_price"`
	Stock              uint          `gorm:"not null;default:0;comment:库存数量" json:"stock"`
	Unit               string        `gorm:"size:16;not null;default:份;comment:计量单位" json:"unit"`
	ProductType        uint8         `gorm:"not null;default:1;comment:商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普资讯" json:"product_type"`
	ServiceContent     JSON          `gorm:"type:json;comment:服务内容配置JSON" json:"service_content"`
	SaleType           uint8         `gorm:"not null;default:1;comment:销售类型: 1=一口价 2=租赁" json:"sale_type"`
	RentalUnit         uint8         `gorm:"not null;default:0;comment:租赁计费周期: 0=非租赁 1=按天 2=按周 3=按月" json:"rental_unit"`
	RentalPrice        float64       `gorm:"type:decimal(10,2);not null;default:0;comment:单位租金(元)" json:"rental_price"`
	Deposit            float64       `gorm:"type:decimal(10,2);not null;default:0;comment:押金(元)" json:"deposit"`
	MaxRentalDuration  uint          `gorm:"not null;default:0;comment:最大租赁时长(0=不限)" json:"max_rental_duration"`
	Sales              uint          `gorm:"not null;default:0;comment:累计销量" json:"sales"`
	Sort               uint          `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status             uint8         `gorm:"not null;default:1;comment:状态: 1=上架 2=下架" json:"status"`
	DeletedAt          *time.Time    `gorm:"comment:软删除时间" json:"deleted_at"`
	CreatedAt          time.Time     `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt          time.Time     `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Category           *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Specs              []ProductSpec `gorm:"foreignKey:ProductID;references:ID" json:"specs,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

// ============================================
// 商品规格表 (product_specs)
// 用途：商品规格定义，如尺寸/口味等，每个规格包含多个选项(名称+加价+库存)
// ============================================
type ProductSpec struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:规格ID" json:"id"`
	ProductID uint64    `gorm:"not null;index;comment:所属商品ID" json:"product_id"`
	Name      string    `gorm:"size:64;not null;comment:规格名称(如:尺寸/口味)" json:"name"`
	Options   JSON      `gorm:"type:json;comment:规格选项JSON [{id,name,price,stock}]" json:"options"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Product   *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (ProductSpec) TableName() string {
	return "product_specs"
}

// ============================================
// C端用户表 (users)
// 用途：C端微信用户信息，以openid为唯一标识，记录访问与消费统计
// ============================================
type User struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;comment:用户ID" json:"id"`
	OpenID       string     `gorm:"column:openid;size:64;uniqueIndex;comment:微信OpenID(用户唯一标识)" json:"openid"`
	UnionID      string     `gorm:"size:64;index;comment:微信UnionID(跨小程序唯一)" json:"union_id"`
	Nickname     string     `gorm:"size:64;comment:用户昵称(默认微信用户)" json:"nickname"`
	Avatar       string     `gorm:"size:512;comment:用户头像URL" json:"avatar"`
	Phone        string     `gorm:"size:20;index;comment:用户手机号" json:"phone"`
	Status       uint8      `gorm:"not null;default:1;comment:状态: 1=正常 0=禁用" json:"status"`
	FirstVisitAt *time.Time `gorm:"comment:首次访问时间" json:"first_visit_at"`
	LastVisitAt  *time.Time `gorm:"comment:最后访问时间" json:"last_visit_at"`
	VisitCount   uint       `gorm:"not null;default:1;comment:累计访问次数" json:"visit_count"`
	HasOrdered   bool       `gorm:"not null;default:false;comment:是否下过单: true=是 false=否" json:"has_ordered"`
	TotalOrders  uint       `gorm:"not null;default:0;comment:累计订单数" json:"total_orders"`
	TotalSpent   float64    `gorm:"type:decimal(10,2);not null;default:0;comment:累计消费金额(元)" json:"total_spent"`
	HasPaid      bool       `gorm:"not null;default:false;comment:是否完成过支付: true=是 false=否" json:"has_paid"`
	FirstPaidAt  *time.Time `gorm:"comment:首次支付时间" json:"first_paid_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// ============================================
// 用户访问记录表 (user_visits)
// 用途：记录C端用户每次访问商家店铺的行为，用于统计分析
// ============================================
type UserVisit struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:记录ID" json:"id"`
	UserID    uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	OpenID    string    `gorm:"column:openid;size:64;index;comment:微信OpenID" json:"openid"`
	VisitTime time.Time `gorm:"autoCreateTime;comment:访问时间" json:"visit_time"`
	Source    string    `gorm:"size:32;comment:访问来源: scan=扫码 direct=直接进入" json:"source"`
}

func (UserVisit) TableName() string {
	return "user_visits"
}

// ============================================
// 用户行为事件表 (user_behavior_events)
// 用途：记录C端用户在店铺内的行为事件(页面浏览/商品查看/下单/支付)，用于转化分析
// ============================================
type UserBehaviorEvent struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:事件ID" json:"id"`
	UserID    uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	OpenID    string    `gorm:"column:openid;size:64;index;comment:微信OpenID" json:"openid"`
	EventType  string    `gorm:"size:32;not null;index;comment:事件类型: page_view=页面浏览 product_view=商品查看 submit_order=提交订单 pay_success=支付成功" json:"event_type"`
	Page       string    `gorm:"size:64;comment:页面标识(如store_home/store_product)" json:"page"`
	ProductID  *uint64   `gorm:"index;comment:关联商品ID(商品查看事件)" json:"product_id"`
	OrderID    *uint64   `gorm:"index;comment:关联订单ID(下单/支付事件)" json:"order_id"`
	Source     string    `gorm:"size:32;comment:事件来源: scan=扫码 direct=直接进入" json:"source"`
	Payload    JSON      `gorm:"type:json;comment:事件附加数据JSON" json:"payload"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (UserBehaviorEvent) TableName() string {
	return "user_behavior_events"
}

// ============================================
// 订单表 (orders)
// 用途：C端用户订单核心数据，包含金额、配送信息与状态流转
// ============================================
type Order struct {
	ID                   uint64      `gorm:"primaryKey;autoIncrement;comment:订单ID" json:"id"`
	OrderNo              string      `gorm:"size:32;uniqueIndex;not null;comment:订单编号" json:"order_no"`
	UserID               uint64      `gorm:"not null;index;comment:下单用户ID" json:"user_id"`
	OrderType            uint8       `gorm:"not null;default:1;comment:订单类型: 1=普通商品 2=租赁商品 3=即时服务 4=预约服务 5=上门服务 6=到店服务" json:"order_type"`
	BizStatus            uint8       `gorm:"not null;default:0;comment:业务状态: 0=无 1=待接单 2=已接单 3=服务中 4=待支付(尾款) 5=已完成 6=已取消" json:"biz_status"`
	ScheduledAt          *time.Time  `gorm:"comment:预约服务时间" json:"scheduled_at"`
	AssignedStaffID      *uint64     `gorm:"index;comment:指派服务员工ID" json:"assigned_staff_id"`
	ActualStartedAt      *time.Time  `gorm:"comment:实际开始时间" json:"actual_started_at"`
	ActualEndedAt        *time.Time  `gorm:"comment:实际结束时间" json:"actual_ended_at"`
	TotalAmount          float64     `gorm:"type:decimal(10,2);not null;default:0;comment:商品总金额(元)" json:"total_amount"`
	DeliveryFee          float64     `gorm:"type:decimal(10,2);not null;default:0;comment:配送费(元)" json:"delivery_fee"`
	DiscountAmount       float64     `gorm:"type:decimal(10,2);not null;default:0;comment:优惠减免金额(元)" json:"discount_amount"`
	PayAmount            float64     `gorm:"type:decimal(10,2);not null;default:0;comment:实付金额(元)=总金额+配送费-优惠+押金" json:"pay_amount"`
	TotalDeposit         float64     `gorm:"type:decimal(10,2);not null;default:0;comment:总押金(元)" json:"total_deposit"`
	DepositStatus        uint8       `gorm:"not null;default:0;comment:押金状态: 0=无押金 1=已收 2=已退 3=部分扣除" json:"deposit_status"`
	DepositRefundAmount  float64     `gorm:"type:decimal(10,2);not null;default:0;comment:押金退还金额" json:"deposit_refund_amount"`
	DepositDeductAmount  float64     `gorm:"type:decimal(10,2);not null;default:0;comment:押金扣除总额(损坏赔偿)" json:"deposit_deduct_amount"`
	DepositRefundedAt    *time.Time  `gorm:"comment:押金退还时间" json:"deposit_refunded_at"`
	RentalReturnedAt     *time.Time  `gorm:"comment:租赁归还时间" json:"rental_returned_at"`
	RentalReturnRemark   string      `gorm:"size:256;comment:归还备注(验机情况)" json:"rental_return_remark"`
	DeliveryAddress      string      `gorm:"size:256;comment:收货地址(配送时填写)" json:"delivery_address"`
	ContactName          string      `gorm:"size:64;comment:联系人姓名(配送时填写)" json:"contact_name"`
	ContactPhone         string      `gorm:"size:20;comment:联系电话(配送时填写)" json:"contact_phone"`
	Status               uint8       `gorm:"not null;default:1;comment:订单状态: 1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款" json:"status"`
	Remark               string      `gorm:"size:256;comment:用户备注" json:"remark"`
	TransactionID        string      `gorm:"size:64;comment:微信支付交易单号" json:"transaction_id"`
	PaidAt               *time.Time  `gorm:"comment:支付完成时间" json:"paid_at"`
	PayNotifyPayload     JSON        `gorm:"type:json;comment:微信支付回调原始报文JSON" json:"pay_notify_payload"`
	ProfitSharingStatus  uint8       `gorm:"not null;default:0;comment:分账状态:0未分账1分账中2分账成功3分账失败4已跳过" json:"profit_sharing_status"`
	ProfitSharingAmount  float64     `gorm:"type:decimal(10,2);not null;default:0;comment:分账总额(元)" json:"profit_sharing_amount"`
	ProfitSharingOrderNo string      `gorm:"size:64;comment:微信分账单号" json:"profit_sharing_order_no"`
	ProfitSharingAt      *time.Time  `gorm:"comment:分账完成时间" json:"profit_sharing_at"`
	ProfitSharingError   string      `gorm:"size:512;comment:分账失败原因" json:"profit_sharing_error"`
	CompletedAt          *time.Time  `gorm:"comment:核销完成时间" json:"completed_at"`
	CancelledAt          *time.Time  `gorm:"comment:取消时间" json:"cancelled_at"`
	RefundedAt           *time.Time  `gorm:"comment:退款完成时间(status=6时写入)" json:"refunded_at"`
	CreatedAt            time.Time   `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time   `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	User                 *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items                []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// ============================================
// 订单商品表 (order_items)
// 用途：订单中的商品快照，记录下单时的商品名称、价格、规格与数量
// ============================================
type OrderItem struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement;comment:明细ID" json:"id"`
	OrderID          uint64    `gorm:"not null;index;comment:所属订单ID" json:"order_id"`
	ProductID        uint64    `gorm:"not null;index;comment:商品ID" json:"product_id"`
	ProductName      string    `gorm:"size:128;not null;comment:商品名称(下单时快照)" json:"product_name"`
	Image            string    `gorm:"size:512;comment:商品图片URL(下单时快照)" json:"image"`
	Price            float64   `gorm:"type:decimal(10,2);not null;comment:商品单价(基础价+规格加价,下单时快照)" json:"price"`
	Quantity         uint      `gorm:"not null;default:1;comment:购买数量" json:"quantity"`
	SpecInfo         JSON      `gorm:"type:json;comment:规格信息JSON(下单时快照)" json:"spec_info"`
	Subtotal         float64   `gorm:"type:decimal(10,2);not null;comment:小计金额(元)=单价×数量" json:"subtotal"`
	SaleType         uint8     `gorm:"not null;default:1;comment:销售类型快照: 1=一口价 2=租赁" json:"sale_type"`
	RentalUnit       uint8     `gorm:"not null;default:0;comment:租赁计费周期快照" json:"rental_unit"`
	RentalDuration   uint      `gorm:"not null;default:0;comment:租赁时长" json:"rental_duration"`
	UnitRentalPrice  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:单位租金快照" json:"unit_rental_price"`
	RentalSubtotal   float64   `gorm:"type:decimal(10,2);not null;default:0;comment:租金小计" json:"rental_subtotal"`
	Deposit          float64   `gorm:"type:decimal(10,2);not null;default:0;comment:单商品押金快照" json:"deposit"`
	DepositDeduct    float64   `gorm:"type:decimal(10,2);not null;default:0;comment:押金扣除金额" json:"deposit_deduct"`
	CreatedAt        time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	Order            *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product          *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// ============================================
// 退款记录表 (refunds)
// 用途：记录每笔退款的退款单号、金额、原因与微信退款状态
// ============================================
type Refund struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;comment:退款ID" json:"id"`
	OrderID      uint64     `gorm:"not null;index;comment:关联订单ID" json:"order_id"`
	RefundNo     string     `gorm:"size:32;uniqueIndex;not null;comment:退款单号" json:"refund_no"`
	RefundAmount float64    `gorm:"type:decimal(10,2);not null;comment:退款金额(元)" json:"refund_amount"`
	RefundReason string     `gorm:"size:256;comment:退款原因" json:"refund_reason"`
	Status       uint8      `gorm:"not null;default:0;comment:退款状态: 0=处理中 1=退款成功 2=退款失败" json:"status"`
	RefundID     string     `gorm:"size:64;comment:微信退款单号" json:"refund_id"`
	RefundedAt   *time.Time `gorm:"comment:退款完成时间" json:"refunded_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Order        *Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (Refund) TableName() string {
	return "refunds"
}

// ============================================
// 平台活动表 (activities)
// 用途：服务商发布的运营活动，支持多种类型与跳转链接
// ============================================
type Activity struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;comment:活动ID" json:"id"`
	Type      string     `gorm:"size:16;not null;comment:活动类型" json:"type"`
	Title     string     `gorm:"size:128;comment:活动标题" json:"title"`
	Content   string     `gorm:"type:text;comment:活动内容" json:"content"`
	Image     string     `gorm:"size:512;comment:活动封面图URL" json:"image"`
	LinkType  string     `gorm:"size:16;comment:跳转类型: page=小程序页面 url=外部链接" json:"link_type"`
	LinkValue string     `gorm:"size:256;comment:跳转目标值(页面路径或URL)" json:"link_value"`
	Sort      uint       `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status    uint8      `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	StartTime *time.Time `gorm:"comment:活动开始时间" json:"start_time"`
	EndTime   *time.Time `gorm:"comment:活动结束时间" json:"end_time"`
	CreatedAt time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Activity) TableName() string {
	return "activities"
}

// ============================================
// 系统公告表 (announcements)
// 用途：发布的系统公告，所有商家可见
// ============================================
type Announcement struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:公告ID" json:"id"`
	Title     string    `gorm:"size:128;not null;comment:公告标题" json:"title"`
	Content   string    `gorm:"type:text;comment:公告内容" json:"content"`
	Status    uint8     `gorm:"not null;default:1;comment:状态: 1=已发布 0=草稿/禁用" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}

// ============================================
// 商家年费表 (merchant_fees)
// 用途：记录商家年费缴纳情况，支持已缴/免费/待缴等状态
// ============================================
type MerchantFee struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;comment:年费ID" json:"id"`
	Year       uint       `gorm:"not null;comment:缴费年度" json:"year"`
	Amount     float64    `gorm:"type:decimal(10,2);not null;default:0;comment:年费金额(元)" json:"amount"`
	Status     string     `gorm:"size:16;not null;comment:缴费状态: paid=已缴 free=免费 pending=待缴" json:"status"`
	PayTime    *time.Time `gorm:"comment:缴费时间" json:"pay_time"`
	FreeReason string     `gorm:"size:256;comment:免费原因(免费时填写)" json:"free_reason"`
	CreatedAt  time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (MerchantFee) TableName() string {
	return "merchant_fees"
}

// ============================================
// 商家手续费率表 (merchant_rates)
// 用途：记录商家不同类型的手续费率及有效期
// ============================================
type MerchantRate struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;comment:费率ID" json:"id"`
	RateType      string     `gorm:"size:32;not null;comment:费率类型(如payment=支付手续费)" json:"rate_type"`
	Rate          float64    `gorm:"type:decimal(5,4);not null;comment:费率(小数,如0.0060=0.6%)" json:"rate"`
	EffectiveTime time.Time  `gorm:"comment:生效时间" json:"effective_time"`
	ExpireTime    *time.Time `gorm:"comment:失效时间(空=永久有效)" json:"expire_time"`
	Remark        string     `gorm:"size:256;comment:备注" json:"remark"`
	Status        uint8      `gorm:"not null;default:1;comment:状态: 1=生效 0=失效" json:"status"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (MerchantRate) TableName() string {
	return "merchant_rates"
}

// ============================================
// 用户收货地址表 (user_addresses)
// 用途：C端用户收货地址管理，支持设置默认地址
// ============================================
type UserAddress struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:地址ID" json:"id"`
	UserID    uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Name      string    `gorm:"size:64;not null;comment:收货人姓名" json:"name"`
	Phone     string    `gorm:"size:20;not null;comment:收货人电话" json:"phone"`
	Province  string    `gorm:"size:32;comment:省份" json:"province"`
	City      string    `gorm:"size:32;comment:城市" json:"city"`
	District  string    `gorm:"size:32;comment:区县" json:"district"`
	Address   string    `gorm:"size:256;not null;comment:详细地址" json:"address"`
	Lat       float64   `gorm:"type:decimal(10,6);comment:纬度" json:"lat"`
	Lng       float64   `gorm:"type:decimal(10,6);comment:经度" json:"lng"`
	IsDefault bool      `gorm:"not null;default:false;comment:是否默认地址: true=默认 false=非默认" json:"is_default"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (UserAddress) TableName() string {
	return "user_addresses"
}

// ============================================
// 服务人员表 (service_staffs)
// 用途：独立于商户员工的服务人员账号，用于接单小程序
// ============================================
type ServiceStaff struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;comment:服务人员ID" json:"id"`
	Username      string     `gorm:"size:64;not null;comment:登录用户名" json:"username"`
	Password      string     `gorm:"size:128;not null;comment:加密密码(不返回)" json:"-"`
	Name          string     `gorm:"size:64;comment:姓名" json:"name"`
	Phone         string     `gorm:"size:20;comment:手机号" json:"phone"`
	OpenID        string     `gorm:"column:openid;size:64;index;comment:微信OpenID(用于快捷登录)" json:"openid"`
	Avatar        string     `gorm:"size:512;comment:头像URL" json:"avatar"`
	Status        uint8      `gorm:"not null;default:0;comment:状态: 0=待审核 1=启用 2=禁用" json:"status"`
	LastLoginAt   *time.Time `gorm:"comment:最后登录时间" json:"last_login_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ServiceStaff) TableName() string {
	return "service_staffs"
}

// ============================================
// 系统菜单表 (sys_menus)
// 用途：PC 管理端菜单/按钮定义，全局共享（不按商家隔离）
// ============================================
type SysMenu struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:菜单ID" json:"id"`
	ParentID   uint64    `gorm:"not null;default:0;comment:父菜单ID(0=顶级)" json:"parent_id"`
	MenuType   uint8     `gorm:"not null;default:1;comment:类型: 1=菜单/目录 2=按钮" json:"menu_type"`
	Name       string    `gorm:"size:64;not null;comment:菜单名称" json:"name"`
	Path       string    `gorm:"size:128;comment:前端路由路径(菜单时)" json:"path"`
	Icon       string    `gorm:"size:64;comment:菜单图标" json:"icon"`
	Sort       uint      `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status     uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	Visible    uint8     `gorm:"not null;default:1;comment:是否显示: 1=显示 0=隐藏" json:"visible"`
	Permission string    `gorm:"size:128;comment:权限标识(如 order:view)" json:"permission"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Children   []SysMenu `gorm:"-" json:"children,omitempty"`
}

func (SysMenu) TableName() string {
	return "sys_menus"
}

// ============================================
// 系统角色表 (sys_roles)
// 用途：PC 管理端角色定义，全局共享
// ============================================
type SysRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:角色ID" json:"id"`
	Name      string    `gorm:"size:64;not null;comment:角色名称" json:"name"`
	Code      string    `gorm:"size:64;not null;uniqueIndex;comment:角色编码(唯一)" json:"code"`
	Remark    string    `gorm:"size:256;comment:备注" json:"remark"`
	Status    uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Menus     []SysMenu `gorm:"many2many:sys_role_menus;joinForeignKey:role_id;joinReferences:menu_id" json:"menus,omitempty"`
}

func (SysRole) TableName() string {
	return "sys_roles"
}

// ============================================
// 角色-菜单关联表 (sys_role_menus)
// ============================================
type SysRoleMenu struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:关联ID" json:"id"`
	RoleID    uint64    `gorm:"not null;uniqueIndex:uk_sys_role_menus;comment:角色ID" json:"role_id"`
	MenuID    uint64    `gorm:"not null;uniqueIndex:uk_sys_role_menus;comment:菜单ID" json:"menu_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (SysRoleMenu) TableName() string {
	return "sys_role_menus"
}

// ============================================
// 系统部门表 (sys_departments)
// 用途：PC 管理端部门（树形），全局共享
// ============================================
type SysDepartment struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:部门ID" json:"id"`
	ParentID  uint64    `gorm:"not null;default:0;comment:父部门ID(0=顶级)" json:"parent_id"`
	Name      string    `gorm:"size:64;not null;comment:部门名称" json:"name"`
	Leader    string    `gorm:"size:64;comment:负责人" json:"leader"`
	Phone     string    `gorm:"size:20;comment:联系电话" json:"phone"`
	Sort      uint      `gorm:"not null;default:0;comment:排序值" json:"sort"`
	Status    uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Children  []SysDepartment `gorm:"-" json:"children,omitempty"`
}

func (SysDepartment) TableName() string {
	return "sys_departments"
}

// ============================================
// 员工-角色关联表 (merchant_staff_roles)
// ============================================
type MerchantStaffRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:关联ID" json:"id"`
	StaffID   uint64    `gorm:"not null;uniqueIndex:uk_merchant_staff_roles;comment:员工ID" json:"staff_id"`
	RoleID    uint64    `gorm:"not null;uniqueIndex:uk_merchant_staff_roles;comment:角色ID" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (MerchantStaffRole) TableName() string {
	return "merchant_staff_roles"
}

// ============================================
// 居民健康档案表 (health_records)
// 用途：C端用户/服务人员维护的基础健康档案，评估后回写 assessment_level
// ============================================
type HealthRecord struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement;comment:档案ID" json:"id"`
	UserID           uint64     `gorm:"not null;uniqueIndex;comment:用户ID" json:"user_id"`
	RealName         string     `gorm:"size:64;comment:真实姓名" json:"real_name"`
	Gender           uint8      `gorm:"not null;default:0;comment:性别:1男2女" json:"gender"`
	BirthDate        string     `gorm:"size:16;comment:出生日期" json:"birth_date"`
	IDCard           string     `gorm:"size:32;comment:身份证号" json:"id_card"`
	Phone            string     `gorm:"size:20;comment:联系电话" json:"phone"`
	EmergencyContact string     `gorm:"size:64;comment:紧急联系人" json:"emergency_contact"`
	EmergencyPhone   string     `gorm:"size:20;comment:紧急联系电话" json:"emergency_phone"`
	Address          string     `gorm:"size:256;comment:常住地址" json:"address"`
	HeightCm         float64    `gorm:"type:decimal(5,1);comment:身高(cm)" json:"height_cm"`
	WeightKg         float64    `gorm:"type:decimal(5,1);comment:体重(kg)" json:"weight_kg"`
	BloodType        string     `gorm:"size:8;comment:血型" json:"blood_type"`
	PastHistory      JSON       `gorm:"type:json;comment:既往病史数组" json:"past_history"`
	AllergyHistory   JSON       `gorm:"type:json;comment:过敏史数组" json:"allergy_history"`
	FamilyHistory    JSON       `gorm:"type:json;comment:家族病史数组" json:"family_history"`
	SurgeryHistory   JSON       `gorm:"type:json;comment:手术史数组" json:"surgery_history"`
	MedicationList   JSON       `gorm:"type:json;comment:长期用药数组" json:"medication_list"`
	ChronicTags      JSON       `gorm:"type:json;comment:慢病标签数组" json:"chronic_tags"`
	Smoking          string     `gorm:"size:32;comment:吸烟情况" json:"smoking"`
	Drinking         string     `gorm:"size:32;comment:饮酒情况" json:"drinking"`
	AssessmentLevel  string     `gorm:"size:32;comment:最近一次评估等级" json:"assessment_level"`
	Remark           string     `gorm:"size:512;comment:备注" json:"remark"`
	Status           uint8      `gorm:"not null;default:1;comment:0未建档1正常2已归档" json:"status"`
	CreatedAt        time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	User             *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (HealthRecord) TableName() string {
	return "health_records"
}

// ============================================
// 健康评估量表表 (health_assessment_forms)
// 用途：管理端配置的评估量表，含题目与评分规则，启用后才可被用户/服务人员使用
// ============================================
type HealthAssessmentForm struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;comment:量表ID" json:"id"`
	Name        string    `gorm:"size:64;not null;comment:量表名称" json:"name"`
	Dimension   string    `gorm:"size:32;comment:评估维度" json:"dimension"`
	Description string    `gorm:"type:text;comment:量表说明" json:"description"`
	Questions   JSON      `gorm:"type:json;comment:题目数组[{key,title,options:[{label,score}]}]" json:"questions"`
	ScoreRule   JSON      `gorm:"type:json;comment:评分规则[{min,max,level,conclusion}]" json:"score_rule"`
	Version     uint      `gorm:"not null;default:1;comment:版本号" json:"version"`
	Status      uint8     `gorm:"not null;default:0;comment:0草稿1启用" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (HealthAssessmentForm) TableName() string {
	return "health_assessment_forms"
}

// ============================================
// 健康评估记录表 (health_assessments)
// 用途：用户/服务人员每次评估的答卷与结果快照，量表名称/答案随记录保存
// ============================================
type HealthAssessment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:评估ID" json:"id"`
	UserID       uint64    `gorm:"not null;index:idx_health_assessments_user_id;comment:用户ID" json:"user_id"`
	FormID       uint64    `gorm:"not null;index:idx_health_assessments_form_id;comment:量表ID" json:"form_id"`
	FormName     string    `gorm:"size:64;comment:量表名称快照" json:"form_name"`
	AssessorType uint8     `gorm:"not null;default:1;comment:1自助2服务人员" json:"assessor_type"`
	StaffID      *uint64   `gorm:"comment:评估服务人员ID" json:"staff_id"`
	Answers      JSON      `gorm:"type:json;comment:答案{key:选中label}" json:"answers"`
	TotalScore   float64   `gorm:"type:decimal(6,1);comment:总分" json:"total_score"`
	Level        string    `gorm:"size:32;comment:评估等级" json:"level"`
	Conclusion   string    `gorm:"size:512;comment:评估结论" json:"conclusion"`
	Suggestions  JSON      `gorm:"type:json;comment:建议数组" json:"suggestions"`
	SymptomDesc  string    `gorm:"size:512;comment:症状描述" json:"symptom_desc"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	Form         *HealthAssessmentForm `gorm:"foreignKey:FormID" json:"form,omitempty"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (HealthAssessment) TableName() string {
	return "health_assessments"
}

// ============================================
// 康复辅具适配建议表 (fitting_recommendations)
// 用途：服务人员依据评估结果生成的辅具适配建议，推荐商品以快照形式保存，支持确认与下单状态流转
// ============================================
type FittingRecommendation struct {
	ID                  uint64            `gorm:"primaryKey;autoIncrement;comment:建议ID" json:"id"`
	UserID              uint64            `gorm:"not null;index:idx_fitting_recommendations_user_id;comment:居民用户ID" json:"user_id"`
	AssessmentID        *uint64           `gorm:"comment:关联评估记录ID" json:"assessment_id"`
	SymptomDesc         string            `gorm:"size:512;comment:症状/需求描述" json:"symptom_desc"`
	FittingResult       string            `gorm:"size:512;comment:适配结论" json:"fitting_result"`
	RecommendedProducts JSON              `gorm:"type:json;comment:推荐商品快照[{product_id,name,reason,sale_type}]" json:"recommended_products"`
	StaffID             *uint64           `gorm:"index:idx_fitting_recommendations_staff_id;comment:生成建议的服务人员ID" json:"staff_id"`
	Status              uint8             `gorm:"not null;default:0;comment:状态:0草稿1已确认2已下单" json:"status"`
	OrderID             *uint64           `gorm:"comment:关联订单ID" json:"order_id"`
	CreatedAt           time.Time         `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time         `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	User                *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Assessment          *HealthAssessment `gorm:"foreignKey:AssessmentID" json:"assessment,omitempty"`
}

func (FittingRecommendation) TableName() string {
	return "fitting_recommendations"
}

// ============================================
// 照护计划表 (care_plans)
// 用途：居家康养照护计划，指派服务人员按计划上门执行护理项
// ============================================
type CarePlan struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement;comment:计划ID" json:"id"`
	UserID          uint64         `gorm:"not null;index:idx_care_plans_user_id;comment:居民用户ID" json:"user_id"`
	Name            string         `gorm:"size:64;not null;comment:计划名称" json:"name"`
	PlanType        uint8          `gorm:"not null;default:1;comment:类型:1生活照料2基础护理3康复训练4综合康养" json:"plan_type"`
	StartDate       string         `gorm:"size:16;comment:开始日期" json:"start_date"`
	EndDate         string         `gorm:"size:16;comment:结束日期" json:"end_date"`
	Frequency       string         `gorm:"size:64;comment:照护频次" json:"frequency"`
	Goals           string         `gorm:"size:512;comment:照护目标" json:"goals"`
	Items           JSON           `gorm:"type:json;comment:护理项配置[{name,desc}]" json:"items"`
	AssignedStaffID *uint64        `gorm:"index:idx_care_plans_assigned_staff_id;comment:指派服务人员ID" json:"assigned_staff_id"`
	OrderID         *uint64        `gorm:"comment:关联服务订单ID(可空)" json:"order_id"`
	Status          uint8          `gorm:"not null;default:0;comment:状态:0草稿1执行中2已暂停3已完成" json:"status"`
	CreatedAt       time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	User            *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	AssignedStaff   *ServiceStaff  `gorm:"foreignKey:AssignedStaffID" json:"assigned_staff,omitempty"`
}

func (CarePlan) TableName() string {
	return "care_plans"
}

// ============================================
// 上门照护记录表 (care_visits)
// 用途：服务人员上门照护执行后的记录，含护理项、生命体征与照片，作为服务闭环的历史凭证
// ============================================
type CareVisit struct {
	ID             uint64        `gorm:"primaryKey;autoIncrement;comment:记录ID" json:"id"`
	PlanID         *uint64       `gorm:"index:idx_care_visits_plan_id;comment:关联照护计划ID(可空)" json:"plan_id"`
	OrderID        *uint64       `gorm:"comment:关联服务订单ID(可空)" json:"order_id"`
	UserID         uint64        `gorm:"not null;index:idx_care_visits_user_id;comment:居民用户ID" json:"user_id"`
	StaffID        uint64        `gorm:"not null;comment:录入服务人员ID" json:"staff_id"`
	VisitAt        *time.Time    `gorm:"comment:到访时间" json:"visit_at"`
	NursingItems   JSON          `gorm:"type:json;comment:完成的护理项[{name,done,remark}]" json:"nursing_items"`
	Vitals         JSON          `gorm:"type:json;comment:生命体征{blood_pressure,blood_glucose,heart_rate,oxygen,weight}" json:"vitals"`
	Photos         JSON          `gorm:"type:json;comment:照片URL数组" json:"photos"`
	Remark         string        `gorm:"size:512;comment:备注" json:"remark"`
	FollowUpAdvice string        `gorm:"size:512;comment:下次随访建议" json:"follow_up_advice"`
	CreatedAt      time.Time     `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	Plan           *CarePlan     `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	User           *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Staff          *ServiceStaff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (CareVisit) TableName() string {
	return "care_visits"
}

// ============================================
// 随访任务表 (follow_up_tasks)
// 用途：服务完成/租赁归还/评估完成后自动生成的随访任务，服务人员可认领执行，
// 管理端可手动创建与代执行，作为基层健康服务闭环的持续跟进
// ============================================
type FollowUpTask struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement;comment:任务ID" json:"id"`
	UserID         uint64         `gorm:"not null;index:idx_follow_up_tasks_user_id;comment:居民用户ID" json:"user_id"`
	TaskType       uint8          `gorm:"not null;default:1;comment:类型:1康复随访2租后回访3慢病随访4评估回访" json:"task_type"`
	SourceType     uint8          `gorm:"not null;default:1;comment:来源:1服务完成2租赁归还3评估完成4手动" json:"source_type"`
	SourceID       *uint64        `gorm:"comment:来源ID(订单/评估ID)" json:"source_id"`
	PlanFollowTime *time.Time     `gorm:"comment:计划随访时间" json:"plan_follow_time"`
	StaffID        *uint64        `gorm:"index:idx_follow_up_tasks_staff_id;comment:执行服务人员ID(可空=待认领)" json:"staff_id"`
	ContactMethod  uint8          `gorm:"not null;default:0;comment:随访方式:1电话2上门3微信" json:"contact_method"`
	Status         uint8          `gorm:"not null;default:0;index:idx_follow_up_tasks_status;comment:状态:0待执行1已完成2已跳过" json:"status"`
	Result         JSON           `gorm:"type:json;comment:随访结果{contact_method,content,education_article_ids,satisfaction,remark}" json:"result"`
	CompletedAt    *time.Time     `gorm:"comment:完成时间" json:"completed_at"`
	Remark         string         `gorm:"size:512;comment:备注" json:"remark"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	User           *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Staff          *ServiceStaff  `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (FollowUpTask) TableName() string {
	return "follow_up_tasks"
}

// ============================================
// 生命体征监测表 (health_monitoring)
// 用途：居民血压/血糖/心率/血氧/体重等体征数据，用户自助或服务人员代录，支持持续跟踪
// ============================================
type HealthMonitoring struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:记录ID" json:"id"`
	UserID     uint64    `gorm:"not null;index:idx_health_monitoring_user_id;comment:用户ID" json:"user_id"`
	RecordType uint8     `gorm:"not null;default:1;index:idx_health_monitoring_record_type;comment:类型:1血压2血糖3心率4血氧5体重" json:"record_type"`
	Value      float64   `gorm:"type:decimal(8,2);not null;default:0;comment:测量值" json:"value"`
	Unit       string    `gorm:"size:16;comment:单位" json:"unit"`
	Extra      JSON      `gorm:"type:json;comment:扩展(如血压高低压)" json:"extra"`
	RecordedBy uint64    `gorm:"not null;default:0;comment:录入人(0=用户本人)" json:"recorded_by"`
	RecordedAt *time.Time `gorm:"comment:测量时间" json:"recorded_at"`
	Remark     string    `gorm:"size:512;comment:备注" json:"remark"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (HealthMonitoring) TableName() string {
	return "health_monitoring"
}

// ============================================
// 健康宣教内容表 (health_education_articles)
// 用途：管理端维护的慢病健康宣教文章，可按慢病标签定向推荐给居民
// ============================================
type HealthEducationArticle struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;comment:文章ID" json:"id"`
	Title     string     `gorm:"size:128;not null;comment:标题" json:"title"`
	Category  string     `gorm:"size:32;comment:分类" json:"category"`
	Cover     string     `gorm:"size:512;comment:封面图URL" json:"cover"`
	Content   string     `gorm:"type:text;comment:正文" json:"content"`
	Tags      JSON       `gorm:"type:json;comment:定向慢病标签数组" json:"tags"`
	Status    uint8      `gorm:"not null;default:0;comment:状态:0草稿1发布" json:"status"`
	PublishAt *time.Time `gorm:"comment:发布时间" json:"publish_at"`
	Views     uint       `gorm:"not null;default:0;comment:浏览量" json:"views"`
	CreatedAt time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (HealthEducationArticle) TableName() string {
	return "health_education_articles"
}

// ============================================
// 分账接收方表 (profit_sharing_receivers)
// 用途：后台可视化管理分账接收方，支持商户号与个人微信两类账号
// ============================================
type ProfitSharingReceiver struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:接收方ID" json:"id"`
	MerchantID   uint64    `gorm:"not null;default:1;comment:商家ID(单商户恒为1)" json:"merchant_id"`
	ReceiverType uint8     `gorm:"not null;comment:接收方类型: 1=商户号 2=个人微信openid" json:"receiver_type"`
	Name         string    `gorm:"size:64;not null;comment:显示名称" json:"name"`
	Account      string    `gorm:"size:64;not null;comment:接收方账号(商户号或个人openid)" json:"account"`
	PersonalName string    `gorm:"size:64;comment:个人真实姓名(个人类型时, 微信实名校验)" json:"personal_name"`
	RelationType string    `gorm:"size:32;not null;default:SERVICE_PROVIDER;comment:与特约商户关系, 默认SERVICE_PROVIDER" json:"relation_type"`
	DefaultRatio float64   `gorm:"type:decimal(5,2);not null;default:0;comment:自动分账默认比例(%)" json:"default_ratio"`
	WechatBound  uint8     `gorm:"not null;default:0;comment:微信接收方关系是否已建立: 0=未建立 1=已建立" json:"wechat_bound"`
	WechatError  string    `gorm:"size:256;comment:微信建立接收方关系失败原因" json:"wechat_error"`
	Status       uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=停用" json:"status"`
	Sort         uint      `gorm:"not null;default:0;comment:排序值" json:"sort"`
	Remark       string    `gorm:"size:256;comment:备注" json:"remark"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ProfitSharingReceiver) TableName() string {
	return "profit_sharing_receivers"
}

// ============================================
// 分账单表 (profit_sharing_records)
// 用途：记录每笔支付对应的分账单及其整体状态
// ============================================
type ProfitSharingRecord struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement;comment:分账单ID" json:"id"`
	MerchantID       uint64    `gorm:"not null;default:1;comment:商家ID(单商户恒为1)" json:"merchant_id"`
	OrderID          uint64    `gorm:"not null;comment:订单ID" json:"order_id"`
	OrderNo          string    `gorm:"size:32;not null;comment:订单编号" json:"order_no"`
	SPMchID          string    `gorm:"size:32;comment:服务商商户号" json:"sp_mchid"`
	SubMchID         string    `gorm:"size:32;comment:特约商户号(分账出资方)" json:"sub_mchid"`
	AppID            string    `gorm:"size:64;comment:分账请求使用的appid(特约商户主体小程序)" json:"appid"`
	TransactionID    string    `gorm:"size:64;comment:微信支付交易单号" json:"transaction_id"`
	OutOrderNo       string    `gorm:"size:64;not null;comment:微信分账单号" json:"out_order_no"`
	TotalAmount      float64   `gorm:"type:decimal(10,2);not null;default:0;comment:订单实付金额(元)" json:"total_amount"`
	TotalShareAmount float64   `gorm:"type:decimal(10,2);not null;default:0;comment:本次分账总额(元)" json:"total_share_amount"`
	Status           uint8     `gorm:"not null;default:0;comment:状态:0待分账1分账中2分账成功3分账失败4已跳过" json:"status"`
	ShareTime        *time.Time `gorm:"comment:分账完成时间" json:"share_time"`
	ErrorMessage     string    `gorm:"size:512;comment:失败原因" json:"error_message"`
	CreatedAt        time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Order            *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Receivers        []ProfitSharingRecordReceiver `gorm:"foreignKey:RecordID" json:"receivers,omitempty"`
}

func (ProfitSharingRecord) TableName() string {
	return "profit_sharing_records"
}

// ============================================
// 分账单明细表 (profit_sharing_record_receivers)
// 用途：记录分账单内每一方的分账金额与微信回执
// ============================================
type ProfitSharingRecordReceiver struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;comment:明细ID" json:"id"`
	RecordID     uint64     `gorm:"not null;comment:分账单ID" json:"record_id"`
	ReceiverID   *uint64    `gorm:"comment:接收方ID(可空)" json:"receiver_id"`
	ReceiverType uint8      `gorm:"not null;comment:接收方类型快照: 1=商户号 2=个人微信openid" json:"receiver_type"`
	ReceiverName string     `gorm:"size:64;not null;comment:接收方名称快照" json:"receiver_name"`
	Account      string     `gorm:"size:64;not null;comment:接收方账号快照" json:"account"`
	Amount       float64    `gorm:"type:decimal(10,2);not null;default:0;comment:分账金额(元)" json:"amount"`
	ResultStatus string     `gorm:"size:32;comment:微信分账结果: PROCESSING/SUCCESS/CLOSED/FAILED/FINISHED" json:"result_status"`
	DetailID     string     `gorm:"size:64;comment:微信分账明细单号" json:"detail_id"`
	FailReason   string     `gorm:"size:128;comment:分账失败原因" json:"fail_reason"`
	FinishTime   *time.Time `gorm:"comment:分账完成时间" json:"finish_time"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ProfitSharingRecordReceiver) TableName() string {
	return "profit_sharing_record_receivers"
}
