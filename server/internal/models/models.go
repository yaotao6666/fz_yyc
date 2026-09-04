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
	TakeoutEnabled       bool      `gorm:"not null;default:true;comment:是否支持配送" json:"takeout_enabled"`
	DineInEnabled        bool      `gorm:"not null;default:true;comment:是否支持堂食" json:"dine_in_enabled"`
	PickupEnabled        bool      `gorm:"not null;default:true;comment:是否支持自提" json:"pickup_enabled"`
	SubMchID             string    `gorm:"size:32;comment:微信支付子商户号(线下进件后回填)" json:"sub_mch_id"`
	PaymentConfigStatus  uint8     `gorm:"not null;default:0;comment:支付配置状态: 0=未完成配置 1=已完成配置(已回填sub_mch_id)" json:"payment_config_status"`
	ProfitSharingEnabled bool      `gorm:"not null;default:false;comment:是否开启自动分账: true=开启 false=关闭" json:"profit_sharing_enabled"`
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
// 商家员工表 (merchant_staffs)
// 用途：商家端登录账号，支持owner/staff角色与微信快捷登录
// ============================================
type MerchantStaff struct {
	ID                  uint64         `gorm:"primaryKey;autoIncrement;comment:员工ID" json:"id"`
	Username            string         `gorm:"size:64;not null;comment:登录用户名" json:"username"`
	Password            string         `gorm:"size:128;not null;comment:加密密码(不返回)" json:"-"`
	Name                string         `gorm:"size:64;comment:员工显示名称" json:"name"`
	Phone               string         `gorm:"size:20;comment:员工手机号" json:"phone"`
	OpenID              string         `gorm:"column:openid;size:64;comment:微信OpenID(用于快捷登录)" json:"openid"`
	UnionID             string         `gorm:"column:unionid;size:64;comment:微信UnionID" json:"unionid"`
	WechatBoundAt       *time.Time     `gorm:"comment:微信绑定时间" json:"wechat_bound_at"`
	Role                string         `gorm:"size:32;not null;default:staff;comment:角色: owner=店主(可管理员工) staff=普通员工" json:"role"`
	NotifyEnabled       bool           `gorm:"not null;default:true;comment:是否开启新订单声音提醒: true=开启 false=关闭" json:"notify_enabled"`
	BrowseNotifyEnabled bool           `gorm:"not null;default:true;comment:是否开启用户进店声音提醒: true=开启 false=关闭" json:"browse_notify_enabled"`
	Status              uint8          `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	DepartmentID        *uint64        `gorm:"comment:部门ID(sys_departments.id)" json:"department_id"`
	LastLoginAt         *time.Time     `gorm:"comment:最后账号密码登录时间" json:"last_login_at"`
	LastWechatLoginAt   *time.Time     `gorm:"comment:最后微信快捷登录时间" json:"last_wechat_login_at"`
	CreatedAt           time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
	ID           uint64      `gorm:"primaryKey;autoIncrement;comment:分类ID" json:"id"`
	Name         string      `gorm:"size:64;not null;comment:分类名称" json:"name"`
	ParentID     *uint64     `gorm:"index;comment:父分类ID(空=一级)" json:"parent_id"`
	Level        uint8       `gorm:"not null;default:1;comment:层级: 1=一级 2=二级 3=三级" json:"level"`
	CategoryType uint8       `gorm:"not null;default:1;comment:分类类型: 1=商品分类 2=服务分类" json:"category_type"`
	Sort         uint        `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status       uint8       `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt    time.Time   `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time   `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Children     []*Category `gorm:"-" json:"children,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// ============================================
// 商品表 (products)
// 用途：商家商品信息，包含价格、库存、规格与上下架状态
// ============================================
type Product struct {
	ID                uint64        `gorm:"primaryKey;autoIncrement;comment:商品ID" json:"id"`
	CategoryID        *uint64       `gorm:"index;comment:所属分类ID(可为空)" json:"category_id"`
	Name              string        `gorm:"size:128;not null;comment:商品名称" json:"name"`
	Description       string        `gorm:"type:text;comment:商品描述" json:"description"`
	Images            JSON          `gorm:"type:json;comment:商品图片URL数组JSON" json:"images"`
	Price             float64       `gorm:"type:decimal(10,2);not null;comment:商品基础价格(元)" json:"price"`
	OriginalPrice     float64       `gorm:"type:decimal(10,2);comment:划线原价(元,0表示不展示)" json:"original_price"`
	Stock             uint          `gorm:"not null;default:0;comment:库存数量" json:"stock"`
	Unit              string        `gorm:"size:16;not null;default:份;comment:计量单位" json:"unit"`
	ProductType       uint8         `gorm:"not null;default:1;comment:商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务" json:"product_type"`
	ServiceContent    JSON          `gorm:"type:json;comment:服务内容配置JSON" json:"service_content"`
	SaleType          uint8         `gorm:"not null;default:1;comment:销售类型: 1=一口价 2=租赁" json:"sale_type"`
	RentalUnit        uint8         `gorm:"not null;default:0;comment:租赁计费周期: 0=非租赁 1=按天 2=按周 3=按月" json:"rental_unit"`
	RentalPrice       float64       `gorm:"type:decimal(10,2);not null;default:0;comment:单位租金(元)" json:"rental_price"`
	Deposit           float64       `gorm:"type:decimal(10,2);not null;default:0;comment:押金(元)" json:"deposit"`
	MaxRentalDuration uint          `gorm:"not null;default:0;comment:最大租赁时长(0=不限)" json:"max_rental_duration"`
	Sales             uint          `gorm:"not null;default:0;comment:累计销量" json:"sales"`
	Sort              uint          `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status            uint8         `gorm:"not null;default:1;comment:状态: 1=上架 2=下架" json:"status"`
	DeletedAt         *time.Time    `gorm:"comment:软删除时间" json:"deleted_at"`
	CreatedAt         time.Time     `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time     `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Category          *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Specs             []ProductSpec `gorm:"foreignKey:ProductID;references:ID" json:"specs,omitempty"`
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
	EventType string    `gorm:"size:32;not null;index;comment:事件类型: page_view=页面浏览 product_view=商品查看 submit_order=提交订单 pay_success=支付成功" json:"event_type"`
	Page      string    `gorm:"size:64;comment:页面标识(如store_home/store_product)" json:"page"`
	ProductID *uint64   `gorm:"index;comment:关联商品ID(商品查看事件)" json:"product_id"`
	OrderID   *uint64   `gorm:"index;comment:关联订单ID(下单/支付事件)" json:"order_id"`
	Source    string    `gorm:"size:32;comment:事件来源: scan=扫码 direct=直接进入" json:"source"`
	Payload   JSON      `gorm:"type:json;comment:事件附加数据JSON" json:"payload"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (UserBehaviorEvent) TableName() string {
	return "user_behavior_events"
}

// ============================================
// 订单表 (orders)
// 用途：C端用户订单核心数据，包含金额、配送信息与状态流转
// ============================================
type Order struct {
	ID                   uint64        `gorm:"primaryKey;autoIncrement;comment:订单ID" json:"id"`
	OrderNo              string        `gorm:"size:32;uniqueIndex;not null;comment:订单编号" json:"order_no"`
	UserID               uint64        `gorm:"not null;index;comment:下单用户ID" json:"user_id"`
	OrderType            uint8         `gorm:"not null;default:1;comment:订单类型: 1=普通商品 2=租赁商品 3=即时服务 4=预约服务 5=上门服务 6=到店服务" json:"order_type"`
	BizStatus            uint8         `gorm:"not null;default:0;comment:业务状态: 0=无 1=待接单 2=已接单 3=服务中 5=已完成 6=已取消 (4 为历史保留位)" json:"biz_status"`
	ScheduledAt          *time.Time    `gorm:"comment:预约服务时间" json:"scheduled_at"`
	AssignedStaffID      *uint64       `gorm:"index;comment:指派服务员工ID" json:"assigned_staff_id"`
	AssignedAt           *time.Time    `gorm:"comment:指派/接单时间" json:"assigned_at"`
	ActualStartedAt      *time.Time    `gorm:"comment:实际开始时间" json:"actual_started_at"`
	ActualEndedAt        *time.Time    `gorm:"comment:实际结束时间" json:"actual_ended_at"`
	TotalAmount          float64       `gorm:"type:decimal(10,2);not null;default:0;comment:商品总金额(元)" json:"total_amount"`
	DeliveryFee          float64       `gorm:"type:decimal(10,2);not null;default:0;comment:配送费(元)" json:"delivery_fee"`
	DiscountAmount       float64       `gorm:"type:decimal(10,2);not null;default:0;comment:优惠减免金额(元)" json:"discount_amount"`
	PayAmount            float64       `gorm:"type:decimal(10,2);not null;default:0;comment:实付金额(元)=总金额+配送费-优惠+押金" json:"pay_amount"`
	TotalDeposit         float64       `gorm:"type:decimal(10,2);not null;default:0;comment:总押金(元)" json:"total_deposit"`
	DepositStatus        uint8         `gorm:"not null;default:0;comment:押金状态: 0=无押金 1=已收 2=已退 3=部分扣除" json:"deposit_status"`
	DepositRefundAmount  float64       `gorm:"type:decimal(10,2);not null;default:0;comment:押金退还金额" json:"deposit_refund_amount"`
	DepositDeductAmount  float64       `gorm:"type:decimal(10,2);not null;default:0;comment:押金扣除总额(损坏赔偿)" json:"deposit_deduct_amount"`
	DepositRefundedAt    *time.Time    `gorm:"comment:押金退还时间" json:"deposit_refunded_at"`
	RentalReturnedAt     *time.Time    `gorm:"comment:租赁归还时间" json:"rental_returned_at"`
	RentalReturnRemark   string        `gorm:"size:256;comment:归还备注(验机情况)" json:"rental_return_remark"`
	RentalEndAt          *time.Time    `gorm:"index;comment:租赁到期时间(支付时=paid_at+租赁时长)" json:"rental_end_at"`
	ParentOrderID        *uint64       `gorm:"index;comment:续租关联原订单ID(0/空=普通订单)" json:"parent_order_id,omitempty"`
	RenewFlag            uint8         `gorm:"not null;default:0;comment:是否续租单: 1=续租 0=非" json:"renew_flag"`
	AssignedStaffName    string        `gorm:"-" json:"assigned_staff_name"`
	RecordID             *uint64       `gorm:"index;comment:服务订单绑定的健康档案ID(服务单必填)" json:"record_id,omitempty"`
	RecordName           string        `gorm:"-" json:"record_name,omitempty"`
	RecordGender         uint8         `gorm:"-" json:"record_gender,omitempty"`
	RecordBirthDate      string        `gorm:"-" json:"record_birth_date,omitempty"`
	DeliveryDistrict     string        `gorm:"size:32;comment:收货区县(服务订单区域匹配用)" json:"delivery_district"`
	DeliveryAddress      string        `gorm:"size:256;comment:收货地址(配送时填写)" json:"delivery_address"`
	ContactName          string        `gorm:"size:64;comment:联系人姓名(配送时填写)" json:"contact_name"`
	ContactPhone         string        `gorm:"size:20;comment:联系电话(配送时填写)" json:"contact_phone"`
	Status               uint8         `gorm:"not null;default:1;comment:订单状态: 1=待支付 2=已支付 3=已完成 4=已取消 5=退款中 6=已退款" json:"status"`
	CanReview            bool          `gorm:"-" json:"can_review,omitempty"` // 待评价标记（服务订单已完成且未评价，瞬时字段）
	Remark               string        `gorm:"size:256;comment:用户备注" json:"remark"`
	TransactionID        string        `gorm:"size:64;comment:微信支付交易单号" json:"transaction_id"`
	PaidAt               *time.Time    `gorm:"comment:支付完成时间" json:"paid_at"`
	PayNotifyPayload     JSON          `gorm:"type:json;comment:微信支付回调原始报文JSON" json:"pay_notify_payload"`
	ProfitSharingStatus  uint8         `gorm:"not null;default:0;comment:分账状态:0未分账1分账中2分账成功3分账失败4已跳过" json:"profit_sharing_status"`
	ProfitSharingAmount  float64       `gorm:"type:decimal(10,2);not null;default:0;comment:分账总额(元)" json:"profit_sharing_amount"`
	ProfitSharingOrderNo string        `gorm:"size:64;comment:微信分账单号" json:"profit_sharing_order_no"`
	ProfitSharingAt      *time.Time    `gorm:"comment:分账完成时间" json:"profit_sharing_at"`
	ProfitSharingError   string        `gorm:"size:512;comment:分账失败原因" json:"profit_sharing_error"`
	CompletedAt          *time.Time    `gorm:"comment:核销完成时间" json:"completed_at"`
	CancelledAt          *time.Time    `gorm:"comment:取消时间" json:"cancelled_at"`
	RefundedAt           *time.Time    `gorm:"comment:退款完成时间(status=6时写入)" json:"refunded_at"`
	CreatedAt            time.Time     `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time     `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	User                 *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items                []OrderItem   `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Customer             *CustomerInfo `gorm:"-" json:"customer,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// CustomerInfo 服务人员工单详情增强展示信息（瞬态字段，不含身份证等敏感信息）
type CustomerInfo struct {
	User   *CustomerUser   `json:"user,omitempty"`
	Record *CustomerRecord `json:"record,omitempty"`
}

// CustomerUser 下单用户增强展示信息
type CustomerUser struct {
	Nickname    string  `json:"nickname"`
	Avatar      string  `json:"avatar"`
	Phone       string  `json:"phone"`
	TotalOrders uint    `json:"total_orders"`
	TotalSpent  float64 `json:"total_spent"`
}

// CustomerRecord 服务订单绑定的健康档案展示信息（不含 id_card 等敏感字段）
type CustomerRecord struct {
	ID               uint64   `json:"id"`
	RealName         string   `json:"real_name"`
	Gender           uint8    `json:"gender"`
	BirthDate        string   `json:"birth_date"`
	Relation         uint8    `json:"relation"`
	AssessmentLevel  string   `json:"assessment_level"`
	Phone            string   `json:"phone"`
	Address          string   `json:"address"`
	AllergyHistory   []string `json:"allergy_history"`
	ChronicTags      []string `json:"chronic_tags"`
	MedicationList   []string `json:"medication_list"`
	EmergencyContact string   `json:"emergency_contact"`
	EmergencyPhone   string   `json:"emergency_phone"`
}

// ============================================
// 订单商品表 (order_items)
// 用途：订单中的商品快照，记录下单时的商品名称、价格、规格与数量
// ============================================
type OrderItem struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement;comment:明细ID" json:"id"`
	OrderID         uint64    `gorm:"not null;index;comment:所属订单ID" json:"order_id"`
	ProductID       uint64    `gorm:"not null;index;comment:商品ID" json:"product_id"`
	ProductName     string    `gorm:"size:128;not null;comment:商品名称(下单时快照)" json:"product_name"`
	Image           string    `gorm:"size:512;comment:商品图片URL(下单时快照)" json:"image"`
	Price           float64   `gorm:"type:decimal(10,2);not null;comment:商品单价(基础价+规格加价,下单时快照)" json:"price"`
	Quantity        uint      `gorm:"not null;default:1;comment:购买数量" json:"quantity"`
	SpecInfo        JSON      `gorm:"type:json;comment:规格信息JSON(下单时快照)" json:"spec_info"`
	Subtotal        float64   `gorm:"type:decimal(10,2);not null;comment:小计金额(元)=单价×数量" json:"subtotal"`
	SaleType        uint8     `gorm:"not null;default:1;comment:销售类型快照: 1=一口价 2=租赁" json:"sale_type"`
	RentalUnit      uint8     `gorm:"not null;default:0;comment:租赁计费周期快照" json:"rental_unit"`
	RentalDuration  uint      `gorm:"not null;default:0;comment:租赁时长" json:"rental_duration"`
	UnitRentalPrice float64   `gorm:"type:decimal(10,2);not null;default:0;comment:单位租金快照" json:"unit_rental_price"`
	RentalSubtotal  float64   `gorm:"type:decimal(10,2);not null;default:0;comment:租金小计" json:"rental_subtotal"`
	Deposit         float64   `gorm:"type:decimal(10,2);not null;default:0;comment:单商品押金快照" json:"deposit"`
	DepositDeduct   float64   `gorm:"type:decimal(10,2);not null;default:0;comment:押金扣除金额" json:"deposit_deduct"`
	CreatedAt       time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	Order           *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product         *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
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
// 小程序轮播图表 (mini_program_banners)
// 用途：商家配置 C 端小程序首页金刚区轮播图
// ============================================
type MiniProgramBanner struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:轮播图ID" json:"id"`
	MerchantID uint64    `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
	Title      string    `gorm:"size:128;comment:轮播图标题(仅后台识别)" json:"title"`
	Image      string    `gorm:"size:512;not null;comment:轮播图图片URL(七牛私有路径)" json:"image"`
	LinkType   string    `gorm:"size:16;not null;default:'none';comment:跳转类型: none=无跳转 product=商品详情 category=分类页 url=外部链接" json:"link_type"`
	LinkValue  string    `gorm:"size:256;comment:跳转目标值: product=product_id category=空(跳分类tab) url=链接" json:"link_value"`
	Sort       uint      `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status     uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (MiniProgramBanner) TableName() string {
        return "mini_program_banners"
}

// ============================================
// 小程序首页推荐商品/服务表 (store_home_recommends)
// 用途：商家维度配置 C 端首页推荐项，可指向实物商品(产品类型1/2)或服务(产品类型3/4)
// ============================================
type HomeRecommend struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:推荐ID" json:"id"`
	MerchantID uint64    `gorm:"not null;index;comment:所属商家ID" json:"merchant_id"`
	ProductID  uint64    `gorm:"not null;comment:目标商品/服务ID(products.id)" json:"product_id"`
	TargetType uint8     `gorm:"not null;default:1;comment:推荐对象类型: 1=实物商品(产品类型1/2) 2=服务(产品类型3/4)" json:"target_type"`
	Title      string    `gorm:"size:128;comment:展示标题(空则用商品名)" json:"title"`
	Sort       uint      `gorm:"not null;default:0;comment:排序值(越小越靠前)" json:"sort"`
	Status     uint8     `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Product    *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (HomeRecommend) TableName() string {
	return "store_home_recommends"
}

// ============================================
// 打印机表 (printers)
// 用途：商家的云打印设备配置，供 PC 后台管理
// ============================================
type Printer struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;comment:打印机ID" json:"id"`
	Name        string     `gorm:"size:64;not null;comment:打印机名称" json:"name"`
	Type        uint8      `gorm:"not null;default:1;comment:类型: 1=飞鹅 2=通用云打印" json:"type"`
	FeieUser    string     `gorm:"size:128;comment:飞鹅云账号" json:"feie_user"`
	FeieUKey    string     `gorm:"size:128;comment:飞鹅云UKey(API密钥)" json:"feie_ukey"`
	FeieSN      string     `gorm:"size:32;comment:飞鹅云打印机编号" json:"feie_sn"`
	Status      uint8      `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	AutoPrint   bool       `gorm:"not null;default:false;comment:是否自动打印小票: true=自动 false=手动" json:"auto_print"`
	IsDefault   bool       `gorm:"not null;default:false;comment:是否为默认打印机: true=默认 false=非默认" json:"is_default"`
	PrintCount  uint       `gorm:"not null;default:0;comment:累计打印次数" json:"print_count"`
	LastPrintAt *time.Time `gorm:"comment:最后打印时间" json:"last_print_at"`
	Remark      string     `gorm:"size:256;comment:备注" json:"remark"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Printer) TableName() string {
	return "printers"
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
	ID             uint64     `gorm:"primaryKey;autoIncrement;comment:服务人员ID" json:"id"`
	Username       string     `gorm:"size:64;not null;comment:登录用户名" json:"username"`
	Password       string     `gorm:"size:128;not null;comment:加密密码(不返回)" json:"-"`
	Name           string     `gorm:"size:64;comment:姓名" json:"name"`
	Phone          string     `gorm:"size:20;comment:手机号" json:"phone"`
	OpenID         string     `gorm:"column:openid;size:64;index;comment:微信OpenID(用于快捷登录)" json:"openid"`
	Avatar         string     `gorm:"size:512;comment:头像URL" json:"avatar"`
	Qualifications JSON       `gorm:"type:json;comment:资质材料列表JSON[{type,name,url}]" json:"qualifications,omitempty"`
	ServiceRegion  string     `gorm:"size:256;comment:服务区域(区县,逗号分隔,空=不限)" json:"service_region"`
	QualityScore   float64    `gorm:"type:decimal(3,1);not null;default:5.0;comment:服务质量分(阶段四写入)" json:"quality_score"`
	Status         uint8      `gorm:"not null;default:0;comment:状态: 0=待审核 1=启用 2=禁用" json:"status"`
	AuditStatus    uint8      `gorm:"not null;default:0;comment:审核状态: 0=无/已通过 1=待审核(与status解耦)" json:"audit_status"`
	PendingFields  JSON       `gorm:"type:json;comment:审核中待变更字段快照" json:"pending_fields,omitempty"`
	LastLoginAt    *time.Time `gorm:"comment:最后登录时间" json:"last_login_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ServiceStaff) TableName() string {
	return "service_staffs"
}

// ============================================
// 服务人员审核记录表 (service_staff_audit_records)
// 用途：注册申请/信息变更/资质提交/状态变更的审核留痕
// 命名约定：服务人员相关表统一 `service_` 前缀（含 service_staffs）
// ============================================
type StaffAuditRecord struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;comment:审核记录ID" json:"id"`
	StaffID        uint64     `gorm:"not null;index;comment:服务人员ID" json:"staff_id"`
	AuditType      uint8      `gorm:"not null;default:1;comment:审核类型:1=注册申请 2=信息变更 3=资质提交 4=状态变更" json:"audit_type"`
	ApplyType      uint8      `gorm:"not null;default:1;comment:申请来源:1=自注册 2=PC添加 3=端上变更" json:"apply_type"`
	BeforeData     JSON       `gorm:"type:json;comment:变更前字段快照" json:"before_data,omitempty"`
	AfterData      JSON       `gorm:"type:json;comment:变更后字段快照" json:"after_data,omitempty"`
	Qualifications JSON       `gorm:"type:json;comment:资质材料URL列表" json:"qualifications,omitempty"`
	Status         uint8      `gorm:"not null;default:0;index;comment:审核状态:0=待审 1=通过 2=驳回" json:"status"`
	ReviewerID     *uint64    `gorm:"comment:审核人工员ID(merchant_staffs.id)" json:"reviewer_id,omitempty"`
	ReviewRemark   string     `gorm:"size:256;comment:审核备注" json:"review_remark"`
	ReviewAt       *time.Time `gorm:"comment:审核时间" json:"review_at,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (StaffAuditRecord) TableName() string {
	return "service_staff_audit_records"
}

// ============================================
// 服务记录表 (service_records)
// 用途：服务订单完结时写入的签到/签退/录音快照（PRD V2.0 阶段三：服务过程安全）
// ============================================
type ServiceRecord struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;comment:服务记录ID" json:"id"`
	OrderID         uint64     `gorm:"not null;uniqueIndex;comment:订单ID" json:"order_id"`
	StaffID         uint64     `gorm:"not null;index;comment:服务人员ID" json:"staff_id"`
	StartTime       *time.Time `gorm:"comment:签到时间(orders快照)" json:"start_time"`
	EndTime         *time.Time `gorm:"comment:签退时间(orders快照)" json:"end_time"`
	GPSTrackURL     string     `gorm:"size:512;comment:轨迹聚合文件URL(可选)" json:"gps_track_url"`
	AudioURL        string     `gorm:"size:512;comment:服务录音URL(七牛)" json:"audio_url"`
	AudioUploadedAt *time.Time `gorm:"comment:录音上传时间(30天清理依据)" json:"audio_uploaded_at"`
	AudioDeletedAt  *time.Time `gorm:"comment:录音删除标记时间" json:"audio_deleted_at"`
	SOSTriggered    uint8      `gorm:"not null;default:0;comment:是否触发SOS: 0=否 1=是" json:"sos_triggered"`
	Status          uint8      `gorm:"not null;default:1;comment:状态: 1=正常 2=异常" json:"status"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ServiceRecord) TableName() string {
	return "service_records"
}

// ============================================
// 服务轨迹点表 (service_location_tracks)
// 用途：服务中工单的高频定位上报（60s/次）
// ============================================
type ServiceLocationTrack struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:轨迹点ID" json:"id"`
	OrderID    uint64    `gorm:"not null;index;comment:订单ID" json:"order_id"`
	StaffID    uint64    `gorm:"not null;index;comment:服务人员ID" json:"staff_id"`
	Lat        float64   `gorm:"type:decimal(10,6);not null;comment:纬度" json:"lat"`
	Lng        float64   `gorm:"type:decimal(10,6);not null;comment:经度" json:"lng"`
	ReportedAt time.Time `gorm:"not null;index;comment:上报时间" json:"reported_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (ServiceLocationTrack) TableName() string {
	return "service_location_tracks"
}

// ============================================
// 服务预警事件表 (service_alert_events)
// 用途：统一承载 SOS 求助与服务超时预警（PRD V2.0 阶段三）
// ============================================
type ServiceAlertEvent struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;comment:预警事件ID" json:"id"`
	OrderID      *uint64    `gorm:"index;comment:关联订单ID(可空)" json:"order_id"`
	StaffID      *uint64    `gorm:"index;comment:服务人员ID(订单级预警可空)" json:"staff_id"`
	AlertType    uint8      `gorm:"not null;comment:预警类型: 1=SOS求助 2=服务超时未结束" json:"alert_type"`
	Lat          float64    `gorm:"type:decimal(10,6);comment:触发位置纬度" json:"lat"`
	Lng          float64    `gorm:"type:decimal(10,6);comment:触发位置经度" json:"lng"`
	Address      string     `gorm:"size:256;comment:触发位置地址" json:"address"`
	Summary      string     `gorm:"size:256;comment:预警摘要(订单级预警触发说明)" json:"summary"`
	Status       uint8      `gorm:"not null;default:1;index;comment:状态: 1=待处理 2=处理中 3=已处理" json:"status"`
	HandlerID    *uint64    `gorm:"comment:处理人ID(merchant_staffs.id)" json:"handler_id"`
	HandlerName  string     `gorm:"size:64;comment:处理人姓名快照" json:"handler_name"`
	HandleRemark string     `gorm:"size:512;comment:处理备注留痕" json:"handle_remark"`
	HandledAt    *time.Time `gorm:"comment:处理时间" json:"handled_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;index;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ServiceAlertEvent) TableName() string {
	return "service_alert_events"
}

// ============================================
// 预警配置表 (alert_settings)
// 用途：单行配置，控制预警总开关与各预警场景的触发阈值（PRD V2.0 阶段五）
// ============================================
type AlertSettings struct {
	ID                      uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Enabled                 bool      `gorm:"not null;default:true;comment:预警总开关" json:"enabled"`
	GoodsUnverifiedHours    int       `gorm:"not null;default:24;comment:实物超时未核销(小时)" json:"goods_unverified_hours"`
	ServiceUnassignedHours  int       `gorm:"not null;default:2;comment:服务超时未指派(小时)" json:"service_unassigned_hours"`
	EscortUnfinishedMinutes int       `gorm:"not null;default:120;comment:陪诊超时未完成(分钟)" json:"escort_unfinished_minutes"`
	ServiceUnstartedMinutes int       `gorm:"not null;default:30;comment:指派超时未签到(分钟)" json:"service_unstarted_minutes"`
	RentalOverdueHours      int       `gorm:"not null;default:24;comment:租赁逾期未归还(小时)" json:"rental_overdue_hours"`
	RefundStuckHours        int       `gorm:"not null;default:24;comment:退款卡在处理中(小时)" json:"refund_stuck_hours"`
	ServiceAudioRetainDays  int       `gorm:"not null;default:30;comment:服务录音保留天数(天)" json:"service_audio_retain_days"`
	CreatedAt               time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AlertSettings) TableName() string {
	return "alert_settings"
}

// ============================================
// 通用配置表 (system_configs)
// 用途：通用 key-value + JSON + 备注 配置存储，供任意业务域复用。
// 说明：原来 alert_settings 单行多列阈值已迁移为多条配置项（见 alertcfg 服务的 key 常量）。
// ============================================
type SystemConfig struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey   string    `gorm:"size:100;not null;uniqueIndex;comment:配置键" json:"config_key"`
	ConfigValue string    `gorm:"type:text;not null;comment:配置值(JSON)" json:"config_value"`
	Remark      string    `gorm:"size:255;default:null;comment:备注" json:"remark"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// ============================================
// 协议表 (agreements)
// 用途：用户协议/隐私政策/录音定位授权协议等版本化管理（PRD V2.0 阶段三 6.5）
// ============================================
type Agreement struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;comment:协议ID" json:"id"`
	Type        uint8      `gorm:"not null;index;comment:协议类型: 1=用户协议 2=隐私政策 3=录音/定位授权协议" json:"type"`
	Title       string     `gorm:"size:128;not null;comment:协议标题" json:"title"`
	Content     string     `gorm:"type:text;comment:协议正文(富文本)" json:"content"`
	Version     string     `gorm:"size:32;not null;comment:版本号(同类型递增,如v1.2)" json:"version"`
	Status      uint8      `gorm:"not null;default:0;comment:状态: 1=已发布(当前生效) 0=草稿/停用" json:"status"`
	PublishedAt *time.Time `gorm:"comment:发布时间" json:"published_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Agreement) TableName() string {
	return "agreements"
}

// ============================================
// 协议同意留痕表 (agreement_consents)
// 用途：记录用户/服务人员对某版本协议的同意行为
// ============================================
type AgreementConsent struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;comment:留痕ID" json:"id"`
	AgreementID uint64    `gorm:"not null;index;comment:协议ID" json:"agreement_id"`
	UserType    uint8     `gorm:"not null;default:1;comment:用户类型: 1=C端用户 2=服务人员" json:"user_type"`
	UserID      uint64    `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Version     string    `gorm:"size:32;not null;comment:同意时协议版本号快照" json:"version"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:同意时间" json:"created_at"`
}

func (AgreementConsent) TableName() string {
	return "agreement_consents"
}

// ============================================
// 服务评价表 (service_reviews)
// 用途：服务订单完结后的用户评价，一单一评，驱动服务质量分计算（PRD V2.0 阶段四）
// ============================================
type ServiceReview struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement;comment:评价ID" json:"id"`
	OrderID           uint64    `gorm:"not null;uniqueIndex;comment:订单ID(一单一评)" json:"order_id"`
	UserID            uint64    `gorm:"not null;index;comment:下单用户ID" json:"user_id"`
	StaffID           uint64    `gorm:"not null;index;comment:被评价服务人员ID" json:"staff_id"`
	Score             uint8     `gorm:"not null;comment:总体评分 1-5" json:"score"`
	AttitudeScore     uint8     `gorm:"not null;comment:服务态度分 1-5" json:"attitude_score"`
	ProfessionalScore uint8     `gorm:"not null;comment:专业技能分 1-5" json:"professional_score"`
	PunctualScore     uint8     `gorm:"not null;comment:准时守约分 1-5" json:"punctual_score"`
	Content           string    `gorm:"size:512;comment:评价内容" json:"content"`
	Images            JSON      `gorm:"type:json;comment:评价图片URL列表" json:"images"`
	Status            uint8     `gorm:"not null;default:1;comment:状态: 1=正常展示 0=后台隐藏" json:"status"`
	CreatedAt         time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ServiceReview) TableName() string {
	return "service_reviews"
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
	ID        uint64          `gorm:"primaryKey;autoIncrement;comment:部门ID" json:"id"`
	ParentID  uint64          `gorm:"not null;default:0;comment:父部门ID(0=顶级)" json:"parent_id"`
	Name      string          `gorm:"size:64;not null;comment:部门名称" json:"name"`
	Leader    string          `gorm:"size:64;comment:负责人" json:"leader"`
	Phone     string          `gorm:"size:20;comment:联系电话" json:"phone"`
	Sort      uint            `gorm:"not null;default:0;comment:排序值" json:"sort"`
	Status    uint8           `gorm:"not null;default:1;comment:状态: 1=启用 0=禁用" json:"status"`
	CreatedAt time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
	ID               uint64    `gorm:"primaryKey;autoIncrement;comment:档案ID" json:"id"`
	UserID           uint64    `gorm:"not null;index:idx_health_records_user_id;comment:用户ID（多档案：普通索引）" json:"user_id"`
	Relation         uint8     `gorm:"not null;default:1;comment:与账号关系:1本人2父母3其他亲属" json:"relation"`
	RealName         string    `gorm:"size:64;comment:真实姓名" json:"real_name"`
	Gender           uint8     `gorm:"not null;default:0;comment:性别:1男2女" json:"gender"`
	BirthDate        string    `gorm:"size:16;comment:出生日期" json:"birth_date"`
	IDCard           string    `gorm:"size:32;comment:身份证号" json:"id_card"`
	Phone            string    `gorm:"size:20;comment:联系电话" json:"phone"`
	EmergencyContact string    `gorm:"size:64;comment:紧急联系人" json:"emergency_contact"`
	EmergencyPhone   string    `gorm:"size:20;comment:紧急联系电话" json:"emergency_phone"`
	Address          string    `gorm:"size:256;comment:常住地址" json:"address"`
	HeightCm         float64   `gorm:"type:decimal(5,1);comment:身高(cm)" json:"height_cm"`
	WeightKg         float64   `gorm:"type:decimal(5,1);comment:体重(kg)" json:"weight_kg"`
	BloodType        string    `gorm:"size:8;comment:血型" json:"blood_type"`
	PastHistory      JSON      `gorm:"type:json;comment:既往病史数组" json:"past_history"`
	AllergyHistory   JSON      `gorm:"type:json;comment:过敏史数组" json:"allergy_history"`
	FamilyHistory    JSON      `gorm:"type:json;comment:家族病史数组" json:"family_history"`
	SurgeryHistory   JSON      `gorm:"type:json;comment:手术史数组" json:"surgery_history"`
	MedicationList   JSON      `gorm:"type:json;comment:长期用药数组" json:"medication_list"`
	ChronicTags      JSON      `gorm:"type:json;comment:慢病标签数组" json:"chronic_tags"`
	Smoking          string    `gorm:"size:32;comment:吸烟情况" json:"smoking"`
	Drinking         string    `gorm:"size:32;comment:饮酒情况" json:"drinking"`
	AssessmentLevel  string    `gorm:"size:32;comment:最近一次评估等级" json:"assessment_level"`
	Remark           string    `gorm:"size:512;comment:备注" json:"remark"`
	Status           uint8     `gorm:"not null;default:1;comment:0未建档1正常2已归档" json:"status"`
	CreatedAt        time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	User             *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
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
	ID           uint64                `gorm:"primaryKey;autoIncrement;comment:评估ID" json:"id"`
	UserID       uint64                `gorm:"not null;index:idx_health_assessments_user_id;comment:用户ID" json:"user_id"`
	RecordID     *uint64               `gorm:"index:idx_health_assessments_record_id;comment:关联档案ID（阶段五 8.2 多档案）" json:"record_id"`
	FormID       uint64                `gorm:"not null;index:idx_health_assessments_form_id;comment:量表ID" json:"form_id"`
	FormName     string                `gorm:"size:64;comment:量表名称快照" json:"form_name"`
	AssessorType uint8                 `gorm:"not null;default:1;comment:1自助2服务人员" json:"assessor_type"`
	StaffID      *uint64               `gorm:"comment:评估服务人员ID" json:"staff_id"`
	Answers      JSON                  `gorm:"type:json;comment:答案{key:选中label}" json:"answers"`
	TotalScore   float64               `gorm:"type:decimal(6,1);comment:总分" json:"total_score"`
	Level        string                `gorm:"size:32;comment:评估等级" json:"level"`
	Conclusion   string                `gorm:"size:512;comment:评估结论" json:"conclusion"`
	Suggestions  JSON                  `gorm:"type:json;comment:建议数组" json:"suggestions"`
	SymptomDesc  string                `gorm:"size:512;comment:症状描述" json:"symptom_desc"`
	CreatedAt    time.Time             `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	Form         *HealthAssessmentForm `gorm:"foreignKey:FormID" json:"form,omitempty"`
	User         *User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Record       *HealthRecord         `gorm:"foreignKey:RecordID" json:"record,omitempty"`
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
	RecordID            *uint64           `gorm:"index:idx_fitting_recommendations_record_id;comment:关联档案ID（阶段五 8.2 多档案）" json:"record_id"`
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
	Record              *HealthRecord     `gorm:"foreignKey:RecordID" json:"record,omitempty"`
	Assessment          *HealthAssessment `gorm:"foreignKey:AssessmentID" json:"assessment,omitempty"`
}

func (FittingRecommendation) TableName() string {
	return "fitting_recommendations"
}

// ============================================
// 健康宣教分类表 (health_education_categories)
// 用途：两级分类（parent_id=0 为一级），商家后台树形 CRUD，C 端仅展示启用项
// ============================================
type HealthEducationCategory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:分类ID" json:"id"`
	ParentID  uint64    `gorm:"not null;default:0;index:idx_hec_parent_id;comment:父分类ID，0=一级" json:"parent_id"`
	Name      string    `gorm:"size:64;not null;comment:分类名称" json:"name"`
	Sort      int32     `gorm:"not null;default:0;comment:排序（小在前）" json:"sort"`
	Status    uint8     `gorm:"not null;default:1;index:idx_hec_status;comment:状态:1启用0停用" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (HealthEducationCategory) TableName() string {
	return "health_education_categories"
}

// ============================================
// 健康宣教内容表 (health_education_articles)
// 用途：商家后台维护的慢病健康宣教文章；C 端按分类浏览，tags 作为附加属性
// ============================================
type HealthEducationArticle struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;comment:文章ID" json:"id"`
	Title      string     `gorm:"size:128;not null;comment:标题" json:"title"`
	CategoryID uint64     `gorm:"not null;default:0;index:idx_hea_category_id;comment:分类ID（关联health_education_categories）" json:"category_id"`
	Category   string     `gorm:"size:32;comment:【兼容保留】旧字符串分类，迁移后仅供历史参考" json:"category"`
	Cover      string     `gorm:"size:512;comment:封面图URL" json:"cover"`
	Content    string     `gorm:"type:text;comment:正文" json:"content"`
	Tags       JSON       `gorm:"type:json;comment:定向慢病标签数组（附加属性，不做板块组织）" json:"tags"`
	Status     uint8      `gorm:"not null;default:0;comment:状态:0草稿1发布" json:"status"`
	PublishAt  *time.Time `gorm:"comment:发布时间" json:"publish_at"`
	Views      uint       `gorm:"not null;default:0;comment:浏览量" json:"views"`
	CreatedAt  time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
	ID               uint64                        `gorm:"primaryKey;autoIncrement;comment:分账单ID" json:"id"`
	MerchantID       uint64                        `gorm:"not null;default:1;comment:商家ID(单商户恒为1)" json:"merchant_id"`
	OrderID          uint64                        `gorm:"not null;comment:订单ID" json:"order_id"`
	OrderNo          string                        `gorm:"size:32;not null;comment:订单编号" json:"order_no"`
	SPMchID          string                        `gorm:"size:32;comment:服务商商户号" json:"sp_mchid"`
	SubMchID         string                        `gorm:"size:32;comment:特约商户号(分账出资方)" json:"sub_mchid"`
	AppID            string                        `gorm:"size:64;comment:分账请求使用的appid(特约商户主体小程序)" json:"appid"`
	TransactionID    string                        `gorm:"size:64;comment:微信支付交易单号" json:"transaction_id"`
	OutOrderNo       string                        `gorm:"size:64;not null;comment:微信分账单号" json:"out_order_no"`
	TotalAmount      float64                       `gorm:"type:decimal(10,2);not null;default:0;comment:订单实付金额(元)" json:"total_amount"`
	TotalShareAmount float64                       `gorm:"type:decimal(10,2);not null;default:0;comment:本次分账总额(元)" json:"total_share_amount"`
	Status           uint8                         `gorm:"not null;default:0;comment:状态:0待分账1分账中2分账成功3分账失败4已跳过" json:"status"`
	ShareTime        *time.Time                    `gorm:"comment:分账完成时间" json:"share_time"`
	ErrorMessage     string                        `gorm:"size:512;comment:失败原因" json:"error_message"`
	CreatedAt        time.Time                     `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time                     `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Order            *Order                        `gorm:"foreignKey:OrderID" json:"order,omitempty"`
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

// ============================================
// 优惠券模板表 (coupon_templates)
// 用途：商户创建的满减券/折扣券模板，控制总量、限领、有效期与适用范围
// ============================================
type CouponTemplate struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement;comment:券模板ID" json:"id"`
	Name            string     `gorm:"size:64;not null;comment:券名称" json:"name"`
	Type            uint8      `gorm:"not null;comment:券类型: 1=满减券 2=折扣券" json:"type"`
	ThresholdAmount float64    `gorm:"type:decimal(10,2);not null;default:0;comment:使用门槛金额(0=无门槛)" json:"threshold_amount"`
	DiscountAmount  float64    `gorm:"type:decimal(10,2);not null;default:0;comment:满减面值(满减券)" json:"discount_amount"`
	DiscountRate    float64    `gorm:"type:decimal(3,2);not null;default:0;comment:折扣率如0.90(折扣券)" json:"discount_rate"`
	TotalCount      int        `gorm:"not null;default:0;comment:发行总量(0=不限)" json:"total_count"`
	ReceivedCount   int        `gorm:"not null;default:0;comment:已领取数量" json:"received_count"`
	PerUserLimit    int        `gorm:"not null;default:1;comment:每人限领数量" json:"per_user_limit"`
	ValidType       uint8      `gorm:"not null;comment:有效期类型: 1=固定期限 2=领取后N天有效" json:"valid_type"`
	ValidStartAt    *time.Time `gorm:"comment:固定期限开始时间" json:"valid_start_at"`
	ValidEndAt      *time.Time `gorm:"comment:固定期限结束时间" json:"valid_end_at"`
	ValidDays       int        `gorm:"comment:领取后有效天数" json:"valid_days"`
	ApplyScope      uint8      `gorm:"not null;default:1;comment:适用范围: 1=全场 2=指定分类 3=指定商品" json:"apply_scope"`
	ScopeIds        JSON       `gorm:"type:json;comment:适用范围ID列表" json:"scope_ids"`
	Status          uint8      `gorm:"not null;default:1;comment:状态: 1=启用 0=停用" json:"status"`
	Remark          string     `gorm:"size:512;comment:备注" json:"remark"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (CouponTemplate) TableName() string {
	return "coupon_templates"
}

// ============================================
// 用户优惠券表 (user_coupons)
// 用途：用户领取/被发放的券实例，记录核销与过期信息
// ============================================
type UserCoupon struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement;comment:用户券ID" json:"id"`
	UserID     uint64          `gorm:"not null;index;comment:用户ID" json:"user_id"`
	TemplateID uint64          `gorm:"not null;index;comment:券模板ID" json:"template_id"`
	Status     uint8           `gorm:"not null;default:1;comment:状态: 1=未使用 2=已使用 3=已过期 4=已作废" json:"status"`
	Source     uint8           `gorm:"not null;default:1;comment:来源: 1=自主领取 2=系统发放(30天唤回) 3=运营手动发放" json:"source"`
	ReceivedAt time.Time       `gorm:"autoCreateTime;comment:领取时间" json:"received_at"`
	ExpiredAt  time.Time       `gorm:"not null;comment:过期时间(领取时计算落库)" json:"expired_at"`
	UsedAt     *time.Time      `gorm:"comment:核销时间" json:"used_at"`
	OrderID    *uint64         `gorm:"index;comment:核销关联订单ID" json:"order_id"`
	OrderNo    string          `gorm:"size:32;comment:核销关联订单号" json:"order_no"`
	CreatedAt  time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Template   *CouponTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	User       *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserCoupon) TableName() string {
	return "user_coupons"
}

// ============================================
// 订单通知日志表 (order_notify_logs)
// 用途：服务人员接单等订单类通知的发送留痕，channel 区分公众号/小程序订阅消息、短信等
// ============================================
type OrderNotifyLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:日志ID" json:"id"`
	OrderID   uint64    `gorm:"not null;index:idx_onn_order_id;comment:订单ID" json:"order_id"`
	UserID    uint64    `gorm:"not null;default:0;comment:接收用户ID" json:"user_id"`
	Channel   string    `gorm:"size:16;not null;default:wechat;comment:发送渠道(wechat/sms)" json:"channel"`
	Content   string    `gorm:"size:255;not null;default:;comment:通知内容" json:"content"`
	Success   uint8     `gorm:"not null;default:0;comment:是否成功: 0=失败 1=成功" json:"success"`
	Message   string    `gorm:"size:255;not null;default:;comment:结果说明" json:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (OrderNotifyLog) TableName() string {
	return "order_notify_logs"
}
