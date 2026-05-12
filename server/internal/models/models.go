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
// ============================================
type ServiceProvider struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	ContactName  string    `gorm:"size:64" json:"contact_name"`
	ContactPhone string    `gorm:"size:20" json:"contact_phone"`
	MchID        string    `gorm:"size:32" json:"mch_id"`
	APIKey       string    `gorm:"size:128" json:"api_key"`
	APIV3Key     string    `gorm:"size:128" json:"api_v3_key"`
	CertSerialNo string    `gorm:"size:64" json:"cert_serial_no"`
	PrivateKey   string    `gorm:"type:text" json:"private_key"`
	PublicKey    string    `gorm:"type:text" json:"public_key"`
	CallbackURL  string    `gorm:"size:256" json:"callback_url"`
	Status       uint8     `gorm:"not null;default:1" json:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ServiceProvider) TableName() string {
	return "service_providers"
}

// ============================================
// 服务商管理员表 (service_provider_admins)
// ============================================
type ServiceProviderAdmin struct {
	ID                uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceProviderID uint64           `gorm:"not null;index" json:"service_provider_id"`
	Username          string           `gorm:"size:64;uniqueIndex" json:"username"`
	Password          string           `gorm:"size:128;not null" json:"-"`
	Name              string           `gorm:"size:64" json:"name"`
	Phone             string           `gorm:"size:20" json:"phone"`
	Role              string           `gorm:"size:32;not null;default:operator" json:"role"`
	Status            uint8            `gorm:"not null;default:1" json:"status"`
	LastLoginAt       *time.Time       `json:"last_login_at"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	ServiceProvider   *ServiceProvider `gorm:"foreignKey:ServiceProviderID" json:"service_provider,omitempty"`
}

func (ServiceProviderAdmin) TableName() string {
	return "service_provider_admins"
}

// ============================================
// 商家表 (merchants)
// ============================================
type Merchant struct {
	ID                uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceProviderID uint64           `gorm:"not null;index" json:"service_provider_id"`
	Name              string           `gorm:"size:128;not null" json:"name"`
	Logo              string           `gorm:"size:512" json:"logo"`
	ContactName       string           `gorm:"size:64" json:"contact_name"`
	ContactPhone      string           `gorm:"size:20" json:"contact_phone"`
	ContactEmail      string           `gorm:"size:128" json:"contact_email"`
	Address           string           `gorm:"size:256" json:"address"`
	Lat               float64          `gorm:"type:decimal(10,6)" json:"lat"`
	Lng               float64          `gorm:"type:decimal(10,6)" json:"lng"`
	BusinessCategory  string           `gorm:"size:64" json:"business_category"`
	BusinessHours     string           `gorm:"size:64" json:"business_hours"`
	Announcement      string           `gorm:"type:text" json:"announcement"`
	MinOrderAmount    float64          `gorm:"type:decimal(10,2);not null;default:0" json:"min_order_amount"`
	TakeoutEnabled    bool             `gorm:"not null;default:true" json:"takeout_enabled"`
	DineInEnabled     bool             `gorm:"not null;default:true" json:"dine_in_enabled"`
	SubMchID          string           `gorm:"size:32" json:"sub_mch_id"`
	SubMchStatus      uint8            `gorm:"not null;default:0" json:"sub_mch_status"`
	ApplymentStatus   uint8            `gorm:"not null;default:0" json:"applyment_status"`
	AuditStatus       uint8            `gorm:"not null;default:0" json:"audit_status"`
	AuditRemark       string           `gorm:"size:256" json:"audit_remark"`
	Status            uint8            `gorm:"not null;default:1" json:"status"`
	Rating            float64          `gorm:"type:decimal(2,1);not null;default:5.0" json:"rating"`
	SalesCount        uint             `gorm:"not null;default:0" json:"sales_count"`
	QRCodeURL         string           `gorm:"size:512" json:"qrcode_url"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	ServiceProvider   *ServiceProvider `gorm:"foreignKey:ServiceProviderID" json:"service_provider,omitempty"`
}

func (Merchant) TableName() string {
	return "merchants"
}

// ============================================
// 商家进件申请表 (merchant_applications)
// ============================================
type MerchantApplication struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID          uint64     `gorm:"not null;index" json:"merchant_id"`
	MerchantName        string     `gorm:"size:128;not null" json:"merchant_name"`
	BusinessLicenseInfo JSON       `gorm:"type:json" json:"business_license_info"`
	LegalPersonInfo     JSON       `gorm:"type:json" json:"legal_person_info"`
	BankAccountInfo     JSON       `gorm:"type:json" json:"bank_account_info"`
	StoreInfo           JSON       `gorm:"type:json" json:"store_info"`
	ContactInfo         JSON       `gorm:"type:json" json:"contact_info"`
	ApplymentID         string     `gorm:"size:64" json:"applyment_id"`
	SubMchID            string     `gorm:"size:32" json:"sub_mch_id"`
	Status              uint8      `gorm:"not null;default:0" json:"status"`
	AuditDetail         JSON       `gorm:"type:json" json:"audit_detail"`
	SubmitTime          *time.Time `json:"submit_time"`
	AuditTime           *time.Time `json:"audit_time"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Merchant            *Merchant  `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (MerchantApplication) TableName() string {
	return "merchant_applications"
}

// ============================================
// 商家配送设置表 (merchant_delivery_settings)
// ============================================
type MerchantDeliverySettings struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID         uint64    `gorm:"uniqueIndex;not null" json:"merchant_id"`
	Enabled            bool      `gorm:"not null;default:true" json:"enabled"`
	BaseFee            float64   `gorm:"type:decimal(10,2);not null;default:0" json:"base_fee"`
	FreeDeliveryAmount float64   `gorm:"type:decimal(10,2);not null;default:0" json:"free_delivery_amount"`
	MaxDistance        uint      `gorm:"not null;default:10" json:"max_distance"`
	DistanceRules      JSON      `gorm:"type:json" json:"distance_rules"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MerchantDeliverySettings) TableName() string {
	return "merchant_delivery_settings"
}

// ============================================
// 商家营业执照表 (merchant_licenses)
// ============================================
type MerchantLicense struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID         uint64     `gorm:"uniqueIndex;not null" json:"merchant_id"`
	LicenseNo          string     `gorm:"size:64" json:"license_no"`
	LicenseName        string     `gorm:"size:128" json:"license_name"`
	LicenseImage       string     `gorm:"size:512" json:"license_image"`
	LegalPerson        string     `gorm:"size:64" json:"legal_person"`
	LegalPersonID      string     `gorm:"size:32" json:"legal_person_id"`
	LegalPersonIDFront string     `gorm:"size:512" json:"legal_person_id_front"`
	LegalPersonIDBack  string     `gorm:"size:512" json:"legal_person_id_back"`
	ValidFrom          *time.Time `gorm:"type:date" json:"valid_from"`
	ValidTo            *time.Time `gorm:"type:date" json:"valid_to"`
	Status             uint8      `gorm:"not null;default:1" json:"status"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MerchantLicense) TableName() string {
	return "merchant_licenses"
}

// ============================================
// 商家员工表 (merchant_staffs)
// ============================================
type MerchantStaff struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID          uint64     `gorm:"not null;index" json:"merchant_id"`
	Username            string     `gorm:"size:64;not null" json:"username"`
	Password            string     `gorm:"size:128;not null" json:"-"`
	Name                string     `gorm:"size:64" json:"name"`
	Phone               string     `gorm:"size:20" json:"phone"`
	OpenID              string     `gorm:"column:openid;size:64" json:"openid"`
	UnionID             string     `gorm:"column:unionid;size:64" json:"unionid"`
	WechatBoundAt       *time.Time `json:"wechat_bound_at"`
	Role                string     `gorm:"size:32;not null;default:staff" json:"role"`
	NotifyEnabled       bool       `gorm:"not null;default:true" json:"notify_enabled"`
	BrowseNotifyEnabled bool       `gorm:"not null;default:true" json:"browse_notify_enabled"`
	Status              uint8      `gorm:"not null;default:1" json:"status"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	LastWechatLoginAt   *time.Time `json:"last_wechat_login_at"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Merchant            *Merchant  `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (MerchantStaff) TableName() string {
	return "merchant_staffs"
}

// ============================================
// 商品分类表 (categories)
// ============================================
type Category struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID uint64    `gorm:"not null;index" json:"merchant_id"`
	Name       string    `gorm:"size:64;not null" json:"name"`
	Sort       uint      `gorm:"not null;default:0" json:"sort"`
	Status     uint8     `gorm:"not null;default:1" json:"status"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Merchant   *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// ============================================
// 商品表 (products)
// ============================================
type Product struct {
	ID            uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID    uint64        `gorm:"not null;index" json:"merchant_id"`
	CategoryID    *uint64       `gorm:"index" json:"category_id"`
	Name          string        `gorm:"size:128;not null" json:"name"`
	Description   string        `gorm:"type:text" json:"description"`
	Images        JSON          `gorm:"type:json" json:"images"`
	Price         float64       `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice float64       `gorm:"type:decimal(10,2)" json:"original_price"`
	Stock         uint          `gorm:"not null;default:0" json:"stock"`
	Unit          string        `gorm:"size:16;not null;default:份" json:"unit"`
	Sales         uint          `gorm:"not null;default:0" json:"sales"`
	Sort          uint          `gorm:"not null;default:0" json:"sort"`
	Status        uint8         `gorm:"not null;default:1" json:"status"`
	DeletedAt     *time.Time    `json:"deleted_at"`
	CreatedAt     time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	Merchant      *Merchant     `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Category      *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Specs         []ProductSpec `gorm:"foreignKey:ProductID;references:ID" json:"specs,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

// ============================================
// 商品规格表 (product_specs)
// ============================================
type ProductSpec struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID uint64    `gorm:"not null;index" json:"product_id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Options   JSON      `gorm:"type:json" json:"options"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Product   *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (ProductSpec) TableName() string {
	return "product_specs"
}

// ============================================
// C端用户表 (users)
// ============================================
type User struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OpenID       string     `gorm:"column:openid;size:64;uniqueIndex" json:"openid"`
	UnionID      string     `gorm:"size:64;index" json:"union_id"`
	Nickname     string     `gorm:"size:64" json:"nickname"`
	Avatar       string     `gorm:"size:512" json:"avatar"`
	Phone        string     `gorm:"size:20;index" json:"phone"`
	Status       uint8      `gorm:"not null;default:1" json:"status"`
	FirstVisitAt *time.Time `json:"first_visit_at"`
	LastVisitAt  *time.Time `json:"last_visit_at"`
	VisitCount   uint       `gorm:"not null;default:1" json:"visit_count"`
	HasOrdered   bool       `gorm:"not null;default:false" json:"has_ordered"`
	TotalOrders  uint       `gorm:"not null;default:0" json:"total_orders"`
	TotalSpent   float64    `gorm:"type:decimal(10,2);not null;default:0" json:"total_spent"`
	HasPaid      bool       `gorm:"not null;default:false" json:"has_paid"`
	FirstPaidAt  *time.Time `json:"first_paid_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// ============================================
// 用户访问记录表 (user_visits)
// ============================================
type UserVisit struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null;index" json:"user_id"`
	MerchantID uint64    `gorm:"not null;index" json:"merchant_id"`
	OpenID     string    `gorm:"column:openid;size:64;index" json:"openid"`
	VisitTime  time.Time `gorm:"autoCreateTime" json:"visit_time"`
	Source     string    `gorm:"size:32" json:"source"`
}

func (UserVisit) TableName() string {
	return "user_visits"
}

// ============================================
// 用户行为事件表 (user_behavior_events)
// ============================================
type UserBehaviorEvent struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID uint64    `gorm:"not null;index" json:"merchant_id"`
	UserID     uint64    `gorm:"not null;index" json:"user_id"`
	OpenID     string    `gorm:"column:openid;size:64;index" json:"openid"`
	EventType  string    `gorm:"size:32;not null;index" json:"event_type"`
	Page       string    `gorm:"size:64" json:"page"`
	ProductID  uint64    `gorm:"index" json:"product_id"`
	OrderID    uint64    `gorm:"index" json:"order_id"`
	Source     string    `gorm:"size:32" json:"source"`
	Payload    JSON      `gorm:"type:json" json:"payload"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserBehaviorEvent) TableName() string {
	return "user_behavior_events"
}

// ============================================
// 订单表 (orders)
// ============================================
type Order struct {
	ID               uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo          string      `gorm:"size:32;uniqueIndex;not null" json:"order_no"`
	UserID           uint64      `gorm:"not null;index" json:"user_id"`
	MerchantID       uint64      `gorm:"not null;index" json:"merchant_id"`
	TotalAmount      float64     `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	DeliveryFee      float64     `gorm:"type:decimal(10,2);not null;default:0" json:"delivery_fee"`
	DiscountAmount   float64     `gorm:"type:decimal(10,2);not null;default:0" json:"discount_amount"`
	PayAmount        float64     `gorm:"type:decimal(10,2);not null;default:0" json:"pay_amount"`
	DeliveryType     uint8       `gorm:"not null;default:1" json:"delivery_type"`
	DeliveryDistance float64     `gorm:"type:decimal(5,2)" json:"delivery_distance"`
	DeliveryAddress  string      `gorm:"size:256" json:"delivery_address"`
	ContactName      string      `gorm:"size:64" json:"contact_name"`
	ContactPhone     string      `gorm:"size:20" json:"contact_phone"`
	Status           uint8       `gorm:"not null;default:1" json:"status"`
	Remark           string      `gorm:"size:256" json:"remark"`
	VerifyCode       string      `gorm:"size:16" json:"verify_code"`
	TransactionID    string      `gorm:"size:64" json:"transaction_id"`
	PaidAt           *time.Time  `json:"paid_at"`
	CompletedAt      *time.Time  `json:"completed_at"`
	CompletedByName  string      `gorm:"size:64" json:"completed_by_name"`
	CancelledAt      *time.Time  `json:"cancelled_at"`
	RefundedAt       *time.Time  `json:"refunded_at"`
	CreatedAt        time.Time   `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt        time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
	User             *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Merchant         *Merchant   `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Items            []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// ============================================
// 订单商品表 (order_items)
// ============================================
type OrderItem struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     uint64    `gorm:"not null;index" json:"order_id"`
	MerchantID  uint64    `gorm:"not null;index" json:"merchant_id"`
	ProductID   uint64    `gorm:"not null;index" json:"product_id"`
	ProductName string    `gorm:"size:128;not null" json:"product_name"`
	Image       string    `gorm:"size:512" json:"image"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity    uint      `gorm:"not null;default:1" json:"quantity"`
	SpecInfo    JSON      `gorm:"type:json" json:"spec_info"`
	Subtotal    float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	Order       *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Merchant    *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Product     *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// ============================================
// 退款记录表 (refunds)
// ============================================
type Refund struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID      uint64     `gorm:"not null;index" json:"order_id"`
	RefundNo     string     `gorm:"size:32;uniqueIndex;not null" json:"refund_no"`
	RefundAmount float64    `gorm:"type:decimal(10,2);not null" json:"refund_amount"`
	RefundReason string     `gorm:"size:256" json:"refund_reason"`
	Status       uint8      `gorm:"not null;default:0" json:"status"`
	RefundID     string     `gorm:"size:64" json:"refund_id"`
	RefundedAt   *time.Time `json:"refunded_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Order        *Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (Refund) TableName() string {
	return "refunds"
}

// ============================================
// 邀请记录表 (invite_records)
// ============================================
type InviteRecord struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	InviterID    uint64     `gorm:"not null;index" json:"inviter_id"`
	InviteeID    uint64     `gorm:"index" json:"invitee_id"`
	InviteCode   string     `gorm:"size:32;index" json:"invite_code"`
	Status       uint8      `gorm:"not null;default:0" json:"status"`
	RewardType   string     `gorm:"size:32" json:"reward_type"`
	RewardStatus uint8      `gorm:"not null;default:0" json:"reward_status"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	Inviter      *Merchant  `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
	Invitee      *Merchant  `gorm:"foreignKey:InviteeID" json:"invitee,omitempty"`
}

func (InviteRecord) TableName() string {
	return "invite_records"
}

// ============================================
// 平台活动表 (activities)
// ============================================
type Activity struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string     `gorm:"size:16;not null" json:"type"`
	Title     string     `gorm:"size:128" json:"title"`
	Content   string     `gorm:"type:text" json:"content"`
	Image     string     `gorm:"size:512" json:"image"`
	LinkType  string     `gorm:"size:16" json:"link_type"`
	LinkValue string     `gorm:"size:256" json:"link_value"`
	Sort      uint       `gorm:"not null;default:0" json:"sort"`
	Status    uint8      `gorm:"not null;default:1" json:"status"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Activity) TableName() string {
	return "activities"
}

// ============================================
// 邀请奖励规则表 (invite_rewards)
// ============================================
type InviteReward struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        string    `gorm:"size:32;not null" json:"type"`
	Condition   string    `gorm:"size:32;not null" json:"condition"`
	Description string    `gorm:"size:256" json:"description"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (InviteReward) TableName() string {
	return "invite_rewards"
}

// ============================================
// 系统公告表 (announcements)
// ============================================
type Announcement struct {
	ID                uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceProviderID uint64           `gorm:"not null;index" json:"service_provider_id"`
	Title             string           `gorm:"size:128;not null" json:"title"`
	Content           string           `gorm:"type:text" json:"content"`
	Status            uint8            `gorm:"not null;default:1" json:"status"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	ServiceProvider   *ServiceProvider `gorm:"foreignKey:ServiceProviderID" json:"service_provider,omitempty"`
}

func (Announcement) TableName() string {
	return "announcements"
}

// ============================================
// 商家审核记录表 (merchant_audit_records)
// ============================================
type MerchantAuditRecord struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID   uint64    `gorm:"not null;index" json:"merchant_id"`
	AuditorID    uint64    `gorm:"not null;index" json:"auditor_id"`
	Action       string    `gorm:"size:32;not null" json:"action"`
	BeforeStatus uint8     `gorm:"not null" json:"before_status"`
	AfterStatus  uint8     `gorm:"not null" json:"after_status"`
	Remark       string    `gorm:"size:256" json:"remark"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	Merchant     *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (MerchantAuditRecord) TableName() string {
	return "merchant_audit_records"
}

// ============================================
// 商家年费表 (merchant_fees)
// ============================================
type MerchantFee struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID uint64     `gorm:"not null;index" json:"merchant_id"`
	Year       uint       `gorm:"not null" json:"year"`
	Amount     float64    `gorm:"type:decimal(10,2);not null;default:0" json:"amount"`
	Status     string     `gorm:"size:16;not null" json:"status"`
	PayTime    *time.Time `json:"pay_time"`
	FreeReason string     `gorm:"size:256" json:"free_reason"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MerchantFee) TableName() string {
	return "merchant_fees"
}

// ============================================
// 商家手续费率表 (merchant_rates)
// ============================================
type MerchantRate struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID    uint64     `gorm:"not null;index" json:"merchant_id"`
	RateType      string     `gorm:"size:32;not null" json:"rate_type"`
	Rate          float64    `gorm:"type:decimal(5,4);not null" json:"rate"`
	EffectiveTime time.Time  `json:"effective_time"`
	ExpireTime    *time.Time `json:"expire_time"`
	Remark        string     `gorm:"size:256" json:"remark"`
	Status        uint8      `gorm:"not null;default:1" json:"status"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MerchantRate) TableName() string {
	return "merchant_rates"
}

// ============================================
// 用户收货地址表 (user_addresses)
// ============================================
type UserAddress struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Phone     string    `gorm:"size:20;not null" json:"phone"`
	Province  string    `gorm:"size:32" json:"province"`
	City      string    `gorm:"size:32" json:"city"`
	District  string    `gorm:"size:32" json:"district"`
	Address   string    `gorm:"size:256;not null" json:"address"`
	Lat       float64   `gorm:"type:decimal(10,6)" json:"lat"`
	Lng       float64   `gorm:"type:decimal(10,6)" json:"lng"`
	IsDefault bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserAddress) TableName() string {
	return "user_addresses"
}

// ============================================
// 优惠券表 (coupons)
// ============================================
type Coupon struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID     uint64    `gorm:"not null;index" json:"merchant_id"`
	Name           string    `gorm:"size:64;not null" json:"name"`
	Type           string    `gorm:"size:16;not null" json:"type"`
	DiscountAmount float64   `gorm:"type:decimal(10,2)" json:"discount_amount"`
	MinOrderAmount float64   `gorm:"type:decimal(10,2);not null;default:0" json:"min_order_amount"`
	TotalCount     int       `gorm:"not null" json:"total_count"`
	RemainingCount int       `gorm:"not null" json:"remaining_count"`
	PerUserLimit   int       `gorm:"not null;default:1" json:"per_user_limit"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         uint8     `gorm:"not null;default:1" json:"status"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Coupon) TableName() string {
	return "coupons"
}

// ============================================
// 优惠券领取记录表 (coupon_records)
// ============================================
type CouponRecord struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64     `gorm:"not null;index" json:"user_id"`
	CouponID  uint64     `gorm:"not null;index" json:"coupon_id"`
	Status    uint8      `gorm:"not null;default:0" json:"status"`
	UsedAt    *time.Time `json:"used_at"`
	OrderID   uint64     `gorm:"index" json:"order_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (CouponRecord) TableName() string {
	return "coupon_records"
}

// ============================================
// 云打印机表 (cloud_printers)
// ============================================
type CloudPrinter struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID  uint64     `gorm:"not null;index" json:"merchant_id"`
	Name        string     `gorm:"size:64;not null" json:"name"`
	Brand       string     `gorm:"size:32" json:"type"`
	DeviceNo    string     `gorm:"size:64;not null" json:"device_no"`
	APIKey      string     `gorm:"size:64" json:"-"`
	APIURL      string     `gorm:"size:256" json:"api_url"`
	PrintTypes  JSON       `gorm:"type:json" json:"print_types"`
	Status      uint8      `gorm:"not null;default:1" json:"status"`
	AutoPrint   bool       `gorm:"not null;default:false" json:"auto_print"`
	IsDefault   bool       `gorm:"not null;default:false" json:"is_default"`
	PrintCount  int        `gorm:"not null;default:0" json:"print_count"`
	LastPrintAt *time.Time `json:"last_print_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CloudPrinter) TableName() string {
	return "cloud_printers"
}

// ============================================
// 打印记录表 (print_logs)
// ============================================
type PrintLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID   uint64    `gorm:"not null;index" json:"merchant_id"`
	PrinterID    uint64    `gorm:"not null;index" json:"printer_id"`
	OrderID      uint64    `gorm:"index" json:"order_id"`
	Type         string    `gorm:"size:16;not null" json:"type"`
	Status       uint8     `gorm:"not null;default:0" json:"status"`
	ErrorMessage string    `gorm:"size:256" json:"error_message"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PrintLog) TableName() string {
	return "print_logs"
}
