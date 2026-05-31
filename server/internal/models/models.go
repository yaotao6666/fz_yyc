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
// 服务商表 (service_providers)
// 用途：存储微信支付服务商基础信息与支付密钥配置
// ============================================
type ServiceProvider struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:服务商ID" json:"id"`
	Name         string    `gorm:"size:128;not null;comment:服务商名称" json:"name"`
	ContactName  string    `gorm:"size:64;comment:联系人姓名" json:"contact_name"`
	ContactPhone string    `gorm:"size:20;comment:联系人电话" json:"contact_phone"`
	MchID        string    `gorm:"size:32;comment:微信支付服务商商户号" json:"mch_id"`
	APIKey       string    `gorm:"size:128;comment:微信支付API密钥" json:"api_key"`
	APIV3Key     string    `gorm:"size:128;comment:微信支付APIv3密钥" json:"api_v3_key"`
	CertSerialNo string    `gorm:"size:64;comment:微信支付证书序列号" json:"cert_serial_no"`
	PrivateKey   string    `gorm:"type:text;comment:微信支付私钥PEM" json:"private_key"`
	PublicKey    string    `gorm:"type:text;comment:微信支付公钥PEM" json:"public_key"`
	CallbackURL  string    `gorm:"size:256;comment:支付回调通知地址" json:"callback_url"`
	Status       uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ServiceProvider) TableName() string {
	return "service_providers"
}

// ============================================
// 服务商账号表 (service_provider_sps)
// 用途：服务商管理端登录账号，支持多角色
// ============================================
type ServiceProviderSp struct {
	ID                uint64           `gorm:"primaryKey;autoIncrement;comment:账号ID" json:"id"`
	ServiceProviderID uint64           `gorm:"not null;index;comment:所属服务商ID" json:"service_provider_id"`
	Username          string           `gorm:"size:64;uniqueIndex;comment:登录用户名" json:"username"`
	Password          string           `gorm:"size:128;not null;comment:加密密码(不返回)" json:"-"`
	Name              string           `gorm:"size:64;comment:账号显示名称" json:"name"`
	Phone             string           `gorm:"size:20;comment:账号手机号" json:"phone"`
	Role              string           `gorm:"size:32;not null;default:operator;comment:角色: owner=所有者 operator=操作员" json:"role"`
	Status            uint8            `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	LastLoginAt       *time.Time       `gorm:"comment:最后登录时间" json:"last_login_at"`
	CreatedAt         time.Time        `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	ServiceProvider   *ServiceProvider `gorm:"foreignKey:ServiceProviderID" json:"service_provider,omitempty"`
}

func (ServiceProviderSp) TableName() string {
	return "service_provider_sps"
}

// ============================================
// 商家表 (merchants)
// 用途：商家核心信息，包含基础资料、支付配置、分账配置与营业状态
// ============================================
type Merchant struct {
	ID                   uint64           `gorm:"primaryKey;autoIncrement;comment:商家ID" json:"id"`
	ServiceProviderID    uint64           `gorm:"not null;index;comment:所属服务商ID" json:"service_provider_id"`
	Name                 string           `gorm:"size:128;not null;comment:商家名称" json:"name"`
	Logo                 string           `gorm:"size:512;comment:商家Logo图片地址(七牛私有路径)" json:"logo"`
	CoverImage           string           `gorm:"size:512;comment:商家背景/封面图地址(七牛私有路径)" json:"cover_image"`
	ContactName          string           `gorm:"size:64;comment:联系人姓名" json:"contact_name"`
	ContactPhone         string           `gorm:"size:20;comment:联系电话(用于用户端拨打退款)" json:"contact_phone"`
	ContactEmail         string           `gorm:"size:128;comment:联系邮箱" json:"contact_email"`
	Address              string           `gorm:"size:256;comment:商家地址" json:"address"`
	Lat                  float64          `gorm:"type:decimal(10,6);comment:纬度" json:"lat"`
	Lng                  float64          `gorm:"type:decimal(10,6);comment:经度" json:"lng"`
	BusinessCategory     string           `gorm:"size:64;comment:经营类目" json:"business_category"`
	BusinessHours        string           `gorm:"size:64;comment:营业时间描述" json:"business_hours"`
	Announcement         string           `gorm:"type:text;comment:商家公告" json:"announcement"`
	MinOrderAmount       float64          `gorm:"type:decimal(10,2);not null;default:0;comment:最低起送金额" json:"min_order_amount"`
	TakeoutEnabled       bool             `gorm:"not null;default:true;comment:是否开启配送: true=开启 false=关闭" json:"takeout_enabled"`
	DineInEnabled        bool             `gorm:"not null;default:true;comment:是否开启堂食: true=开启 false=关闭" json:"dine_in_enabled"`
	PickupEnabled        bool             `gorm:"not null;default:true;comment:是否开启自提: true=开启 false=关闭" json:"pickup_enabled"`
	SubMchID             string           `gorm:"size:32;comment:微信支付子商户号(线下进件后回填)" json:"sub_mch_id"`
	ProfitSharingEnabled bool             `gorm:"not null;default:false;comment:是否开启分账: true=开启 false=关闭" json:"profit_sharing_enabled"`
	ProfitSharingRatio   float64          `gorm:"type:decimal(5,2);not null;default:0;comment:分账比例(百分比0-100)" json:"profit_sharing_ratio"`
	PaymentConfigStatus  uint8            `gorm:"not null;default:0;comment:支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)" json:"payment_config_status"`
	Status               uint8            `gorm:"not null;default:1;comment:营业状态: 1=营业中 0=休息中" json:"status"`
	Rating               float64          `gorm:"type:decimal(2,1);not null;default:5.0;comment:商家评分(1.0-5.0)" json:"rating"`
	SalesCount           uint             `gorm:"not null;default:0;comment:累计销量" json:"sales_count"`
	QRCodeURL            string           `gorm:"size:512;comment:商家小程序码图片地址" json:"qrcode_url"`
	CreatedAt            time.Time        `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time        `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	ServiceProvider      *ServiceProvider `gorm:"foreignKey:ServiceProviderID" json:"service_provider,omitempty"`
}

func (Merchant) TableName() string {
	return "merchants"
}

// ============================================
// 商家配送设置表 (merchant_delivery_settings)
// 用途：商家配送费规则配置，包含基础运费、满减免与距离阶梯计费
// ============================================
type MerchantDeliverySettings struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement;comment:设置ID" json:"id"`
	MerchantID         uint64    `gorm:"uniqueIndex;not null;comment:商家ID(一对一)" json:"merchant_id"`
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
	MerchantID          uint64     `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
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
	LastLoginAt         *time.Time `gorm:"comment:最后账号密码登录时间" json:"last_login_at"`
	LastWechatLoginAt   *time.Time `gorm:"comment:最后微信快捷登录时间" json:"last_wechat_login_at"`
	CreatedAt           time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Merchant            *Merchant  `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (MerchantStaff) TableName() string {
	return "merchant_staffs"
}

// ============================================
// 商品分类表 (categories)
// 用途：商家商品分类管理，支持排序与上下架
// ============================================
type Category struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:分类ID" json:"id"`
	MerchantID uint64    `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
	Name       string    `gorm:"size:64;not null;comment:分类名称" json:"name"`
	Sort       uint      `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status     uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Merchant   *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// ============================================
// 商品表 (products)
// 用途：商家商品信息，包含价格、库存、规格与上下架状态
// ============================================
type Product struct {
	ID            uint64        `gorm:"primaryKey;autoIncrement;comment:商品ID" json:"id"`
	MerchantID    uint64        `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
	CategoryID    *uint64       `gorm:"index;comment:所属分类ID(可为空)" json:"category_id"`
	Name          string        `gorm:"size:128;not null;comment:商品名称" json:"name"`
	Description   string        `gorm:"type:text;comment:商品描述" json:"description"`
	Images        JSON          `gorm:"type:json;comment:商品图片URL数组JSON" json:"images"`
	Price         float64       `gorm:"type:decimal(10,2);not null;comment:商品基础价格(元)" json:"price"`
	OriginalPrice float64       `gorm:"type:decimal(10,2);comment:划线原价(元,0表示不展示)" json:"original_price"`
	Stock         uint          `gorm:"not null;default:0;comment:库存数量" json:"stock"`
	Unit          string        `gorm:"size:16;not null;default:份;comment:计量单位" json:"unit"`
	Sales         uint          `gorm:"not null;default:0;comment:累计销量" json:"sales"`
	Sort          uint          `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status        uint8         `gorm:"not null;default:1;comment:状态: 1=上架 2=下架" json:"status"`
	DeletedAt     *time.Time    `gorm:"comment:软删除时间" json:"deleted_at"`
	CreatedAt     time.Time     `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Merchant      *Merchant     `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Category      *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Specs         []ProductSpec `gorm:"foreignKey:ProductID;references:ID" json:"specs,omitempty"`
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
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:记录ID" json:"id"`
	UserID     uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	MerchantID uint64    `gorm:"not null;index;comment:商家ID" json:"merchant_id"`
	OpenID     string    `gorm:"column:openid;size:64;index;comment:微信OpenID" json:"openid"`
	VisitTime  time.Time `gorm:"autoCreateTime;comment:访问时间" json:"visit_time"`
	Source     string    `gorm:"size:32;comment:访问来源: scan=扫码 direct=直接进入" json:"source"`
}

func (UserVisit) TableName() string {
	return "user_visits"
}

// ============================================
// 用户行为事件表 (user_behavior_events)
// 用途：记录C端用户在店铺内的行为事件(页面浏览/商品查看/下单/支付)，用于转化分析
// ============================================
type UserBehaviorEvent struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:事件ID" json:"id"`
	MerchantID uint64    `gorm:"not null;index;comment:商家ID" json:"merchant_id"`
	UserID     uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	OpenID     string    `gorm:"column:openid;size:64;index;comment:微信OpenID" json:"openid"`
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
// 用途：C端用户订单核心数据，包含金额、配送信息、状态流转与分账结果
// ============================================
type Order struct {
	ID                   uint64      `gorm:"primaryKey;autoIncrement;comment:订单ID" json:"id"`
	OrderNo              string      `gorm:"size:32;uniqueIndex;not null;comment:订单编号" json:"order_no"`
	UserID               uint64      `gorm:"not null;index;comment:下单用户ID" json:"user_id"`
	MerchantID           uint64      `gorm:"not null;index;comment:商家ID" json:"merchant_id"`
	TotalAmount          float64     `gorm:"type:decimal(10,2);not null;default:0;comment:商品总金额(元)" json:"total_amount"`
	DeliveryFee          float64     `gorm:"type:decimal(10,2);not null;default:0;comment:配送费(元)" json:"delivery_fee"`
	DiscountAmount       float64     `gorm:"type:decimal(10,2);not null;default:0;comment:优惠减免金额(元)" json:"discount_amount"`
	PayAmount            float64     `gorm:"type:decimal(10,2);not null;default:0;comment:实付金额(元)=总金额+配送费-优惠" json:"pay_amount"`
	DeliveryType         uint8       `gorm:"not null;default:1;comment:取餐方式: 1=配送 2=堂食 3=自提" json:"delivery_type"`
	DeliveryDistance     float64     `gorm:"type:decimal(5,2);comment:配送距离(km,配送时用户选择)" json:"delivery_distance"`
	DeliveryAddress      string      `gorm:"size:256;comment:收货地址(配送时填写)" json:"delivery_address"`
	ContactName          string      `gorm:"size:64;comment:联系人姓名(配送时填写)" json:"contact_name"`
	ContactPhone         string      `gorm:"size:20;comment:联系电话(配送时填写)" json:"contact_phone"`
	Status               uint8       `gorm:"not null;default:1;comment:订单状态: 1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款" json:"status"`
	Remark               string      `gorm:"size:256;comment:用户备注" json:"remark"`
	VerifyCode           string      `gorm:"size:16;comment:核销码(已支付订单出示给商家)" json:"verify_code"`
	VerifyQRCodeURL      string      `gorm:"-" json:"verify_qrcode_url,omitempty"`
	TransactionID        string      `gorm:"size:64;comment:微信支付交易单号" json:"transaction_id"`
	PaidAt               *time.Time  `gorm:"comment:支付完成时间" json:"paid_at"`
	PayNotifyPayload     JSON        `gorm:"type:json;comment:微信支付回调原始报文JSON" json:"pay_notify_payload"`
	CompletedAt          *time.Time  `gorm:"comment:核销完成时间" json:"completed_at"`
	CompletedByName      string      `gorm:"size:64;comment:核销操作人姓名" json:"completed_by_name"`
	CancelledAt          *time.Time  `gorm:"comment:取消时间" json:"cancelled_at"`
	RefundedAt           *time.Time  `gorm:"comment:退款完成时间(status=6时写入)" json:"refunded_at"`
	ProfitSharingStatus  uint8       `gorm:"not null;default:0;comment:分账状态: 0=未分账 1=分账中 2=分账成功 3=分账失败 4=已跳过" json:"profit_sharing_status"`
	ProfitSharingAmount  float64     `gorm:"type:decimal(10,2);not null;default:0;comment:分账金额-服务商抽佣(元)" json:"profit_sharing_amount"`
	ProfitSharingOrderNo string      `gorm:"size:64;comment:微信分账单号" json:"profit_sharing_order_no"`
	ProfitSharingAt      *time.Time  `gorm:"comment:分账完成时间" json:"profit_sharing_at"`
	ProfitSharingError   string      `gorm:"size:256;comment:分账失败原因" json:"profit_sharing_error"`
	CreatedAt            time.Time   `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time   `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	User                 *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Merchant             *Merchant   `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
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
	ID          uint64    `gorm:"primaryKey;autoIncrement;comment:明细ID" json:"id"`
	OrderID     uint64    `gorm:"not null;index;comment:所属订单ID" json:"order_id"`
	MerchantID  uint64    `gorm:"not null;index;comment:商家ID(冗余,便于按商家查询)" json:"merchant_id"`
	ProductID   uint64    `gorm:"not null;index;comment:商品ID" json:"product_id"`
	ProductName string    `gorm:"size:128;not null;comment:商品名称(下单时快照)" json:"product_name"`
	Image       string    `gorm:"size:512;comment:商品图片URL(下单时快照)" json:"image"`
	Price       float64   `gorm:"type:decimal(10,2);not null;comment:商品单价(基础价+规格加价,下单时快照)" json:"price"`
	Quantity    uint      `gorm:"not null;default:1;comment:购买数量" json:"quantity"`
	SpecInfo    JSON      `gorm:"type:json;comment:规格信息JSON(下单时快照)" json:"spec_info"`
	Subtotal    float64   `gorm:"type:decimal(10,2);not null;comment:小计金额(元)=单价×数量" json:"subtotal"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	Order       *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Merchant    *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Product     *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
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
// 商家分账记录表 (merchant_profit_sharing_records)
// 用途：记录每笔订单的分账结果，包含抽佣金额、比例、商家实收与状态
// ============================================
type MerchantProfitSharingRecord struct {
	ID                     uint64    `gorm:"primaryKey;autoIncrement;comment:分账记录ID" json:"id"`
	ServiceProviderID      uint64    `gorm:"not null;index;comment:服务商ID" json:"service_provider_id"`
	MerchantID             uint64    `gorm:"not null;index;comment:商家ID" json:"merchant_id"`
	OrderID                uint64    `gorm:"not null;index;comment:关联订单ID" json:"order_id"`
	OrderNo                string    `gorm:"size:32;not null;index;comment:订单编号" json:"order_no"`
	TransactionID          string    `gorm:"size:64;index;comment:微信支付交易单号" json:"transaction_id"`
	ProfitSharingOrderNo   string    `gorm:"size:64;not null;index;comment:微信分账单号" json:"profit_sharing_order_no"`
	ProfitSharingDate      time.Time `gorm:"not null;index;comment:分账日期" json:"profit_sharing_date"`
	PayAmount              float64   `gorm:"type:decimal(10,2);not null;default:0;comment:订单支付金额(元)" json:"pay_amount"`
	ProfitSharingRatio     float64   `gorm:"type:decimal(5,2);not null;default:0;comment:分账比例(百分比)" json:"profit_sharing_ratio"`
	ProfitSharingAmount    float64   `gorm:"type:decimal(10,2);not null;default:0;comment:分账金额-服务商抽佣(元)" json:"profit_sharing_amount"`
	MerchantReceivedAmount float64   `gorm:"type:decimal(10,2);not null;default:0;comment:商家实收金额(元)" json:"merchant_received_amount"`
	Status                 uint8     `gorm:"not null;default:0;index;comment:分账状态: 0=处理中 1=分账成功 2=分账失败 3=已跳过" json:"status"`
	ErrorMessage           string    `gorm:"size:256;comment:分账失败原因" json:"error_message"`
	CreatedAt              time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Order                  *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Merchant               *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (MerchantProfitSharingRecord) TableName() string {
	return "merchant_profit_sharing_records"
}

// ============================================
// 平台活动表 (activities)
// 用途：服务商发布的运营活动，支持多种类型与跳转链接
// ============================================
type Activity struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement;comment:活动ID" json:"id"`
	ServiceProviderID uint64     `gorm:"not null;index;comment:所属服务商ID" json:"service_provider_id"`
	Type              string     `gorm:"size:16;not null;comment:活动类型" json:"type"`
	Title             string     `gorm:"size:128;comment:活动标题" json:"title"`
	Content           string     `gorm:"type:text;comment:活动内容" json:"content"`
	Image             string     `gorm:"size:512;comment:活动封面图URL" json:"image"`
	LinkType          string     `gorm:"size:16;comment:跳转类型: page=小程序页面 url=外部链接" json:"link_type"`
	LinkValue         string     `gorm:"size:256;comment:跳转目标值(页面路径或URL)" json:"link_value"`
	Sort              uint       `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status            uint8      `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	StartTime         *time.Time `gorm:"comment:活动开始时间" json:"start_time"`
	EndTime           *time.Time `gorm:"comment:活动结束时间" json:"end_time"`
	CreatedAt         time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Activity) TableName() string {
	return "activities"
}

// ============================================
// 系统公告表 (announcements)
// 用途：服务商发布的系统公告，所有商家可见
// ============================================
type Announcement struct {
	ID                uint64           `gorm:"primaryKey;autoIncrement;comment:公告ID" json:"id"`
	ServiceProviderID uint64           `gorm:"not null;index;comment:所属服务商ID" json:"service_provider_id"`
	Title             string           `gorm:"size:128;not null;comment:公告标题" json:"title"`
	Content           string           `gorm:"type:text;comment:公告内容" json:"content"`
	Status            uint8            `gorm:"not null;default:1;comment:状态: 1=已发布 0=草稿/禁用" json:"status"`
	CreatedAt         time.Time        `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	ServiceProvider   *ServiceProvider `gorm:"foreignKey:ServiceProviderID" json:"service_provider,omitempty"`
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
	MerchantID uint64     `gorm:"not null;index;comment:商家ID" json:"merchant_id"`
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
	MerchantID    uint64     `gorm:"not null;index;comment:商家ID" json:"merchant_id"`
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
// 优惠券表 (coupons)
// 用途：商家发放的优惠券，支持满减类型与每人限领
// ============================================
type Coupon struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;comment:优惠券ID" json:"id"`
	MerchantID     uint64    `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
	Name           string    `gorm:"size:64;not null;comment:优惠券名称" json:"name"`
	Type           string    `gorm:"size:16;not null;comment:优惠券类型: discount=满减" json:"type"`
	DiscountAmount float64   `gorm:"type:decimal(10,2);comment:优惠金额(元)" json:"discount_amount"`
	MinOrderAmount float64   `gorm:"type:decimal(10,2);not null;default:0;comment:最低使用金额(元)" json:"min_order_amount"`
	TotalCount     int       `gorm:"not null;comment:发行总量" json:"total_count"`
	RemainingCount int       `gorm:"not null;comment:剩余数量" json:"remaining_count"`
	PerUserLimit   int       `gorm:"not null;default:1;comment:每人限领数量" json:"per_user_limit"`
	StartTime      time.Time `gorm:"comment:生效时间" json:"start_time"`
	EndTime        time.Time `gorm:"comment:失效时间" json:"end_time"`
	Status         uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt      time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Coupon) TableName() string {
	return "coupons"
}

// ============================================
// 优惠券领取记录表 (coupon_records)
// 用途：记录用户领取和使用优惠券的情况
// ============================================
type CouponRecord struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;comment:领取记录ID" json:"id"`
	UserID    uint64     `gorm:"not null;index;comment:用户ID" json:"user_id"`
	CouponID  uint64     `gorm:"not null;index;comment:优惠券ID" json:"coupon_id"`
	Status    uint8      `gorm:"not null;default:0;comment:状态: 0=未使用 1=已使用 2=已过期" json:"status"`
	UsedAt    *time.Time `gorm:"comment:使用时间" json:"used_at"`
	OrderID   uint64     `gorm:"index;comment:使用时关联的订单ID" json:"order_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime;comment:领取时间" json:"created_at"`
}

func (CouponRecord) TableName() string {
	return "coupon_records"
}

// ============================================
// 商家满减规则表 (merchant_full_reduction_rules)
// 用途：商家配置自动满减规则，下单时按门槛自动匹配最优优惠
// ============================================
type MerchantFullReductionRule struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;comment:规则ID" json:"id"`
	MerchantID     uint64    `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
	ThresholdAmount float64  `gorm:"type:decimal(10,2);not null;comment:满减门槛金额(元)" json:"threshold_amount"`
	DiscountAmount float64   `gorm:"type:decimal(10,2);not null;comment:减免金额(元)" json:"discount_amount"`
	Sort           int       `gorm:"not null;default:0;comment:排序值" json:"sort"`
	Status         uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=停用" json:"status"`
	CreatedAt      time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (MerchantFullReductionRule) TableName() string {
	return "merchant_full_reduction_rules"
}

// ============================================
// 云打印机表 (cloud_printers)
// 用途：商家绑定的云打印机设备，支持自动打印与多打印类型
// ============================================
type CloudPrinter struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;comment:打印机ID" json:"id"`
	MerchantID  uint64     `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
	Name        string     `gorm:"size:64;not null;comment:打印机名称" json:"name"`
	Brand       string     `gorm:"size:32;comment:打印机品牌/型号" json:"type"`
	DeviceNo    string     `gorm:"size:64;not null;comment:打印机设备编号" json:"device_no"`
	APIKey      string     `gorm:"size:64;comment:打印机API密钥(不返回)" json:"-"`
	APIURL      string     `gorm:"size:256;comment:打印机API地址" json:"api_url"`
	FeieUser    string     `gorm:"size:64;comment:飞鹅账号" json:"feie_user"`
	FeieUKey    string     `gorm:"size:128;comment:飞鹅UKey(不返回)" json:"-"`
	FeieSN      string     `gorm:"size:64;comment:飞鹅打印机终端号" json:"feie_sn"`
	PrintTypes  JSON       `gorm:"type:json;comment:自动打印类型JSON(如[new_order]=新订单自动打印)" json:"print_types"`
	Status      uint8      `gorm:"not null;default:1;comment:状态: 1=在线 0=离线" json:"status"`
	AutoPrint   bool       `gorm:"not null;default:false;comment:是否自动打印: true=开启 false=关闭" json:"auto_print"`
	IsDefault   bool       `gorm:"not null;default:false;comment:是否默认打印机: true=默认 false=非默认" json:"is_default"`
	PrintCount  int        `gorm:"not null;default:0;comment:累计打印次数" json:"print_count"`
	LastPrintAt *time.Time `gorm:"comment:最后打印时间" json:"last_print_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (CloudPrinter) TableName() string {
	return "cloud_printers"
}

// ============================================
// 打印记录表 (print_logs)
// 用途：记录每次打印任务的执行状态与错误信息
// ============================================
type PrintLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:打印记录ID" json:"id"`
	MerchantID   uint64    `gorm:"not null;index;comment:商家ID" json:"merchant_id"`
	PrinterID    uint64    `gorm:"not null;index;comment:打印机ID" json:"printer_id"`
	OrderID      uint64    `gorm:"index;comment:关联订单ID" json:"order_id"`
	Type         string    `gorm:"size:16;not null;comment:打印类型: new_order=新订单 refund=退款" json:"type"`
	Status       uint8     `gorm:"not null;default:0;comment:打印状态: 0=待打印 1=打印成功 2=打印失败" json:"status"`
	ErrorMessage string    `gorm:"size:256;comment:打印失败原因" json:"error_message"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (PrintLog) TableName() string {
	return "print_logs"
}
