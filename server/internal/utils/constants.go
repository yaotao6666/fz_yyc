package utils

// DefaultMerchantID 全局单例商家ID。
// 系统已深度重构为单商户模式，merchants 表仅保留 id=1 一行作为全局配置，
// 其余业务表不再携带 merchant_id，业务侧统一使用该常量读取单例配置。
const DefaultMerchantID uint64 = 1

// ============================================
// 订单分类（PRD V2.0 口径：实物订单/服务订单二分）
// 实物订单 order_type ∈ {1,2}：主状态走 orders.status
// 服务订单 order_type ∈ {3,4,5,6}：工单状态走 orders.biz_status
// ============================================
const (
	OrderCategoryGoods   uint8 = 1 // 实物订单（1=普通商品 2=租赁商品）
	OrderCategoryService uint8 = 2 // 服务订单（3=即时 4=预约 5=上门 6=到店）
)

// OrderTypeGoodsMin/OrderTypeGoodsMax 实物订单 order_type 范围
const (
	OrderTypeGoodsMin uint8 = 1 // 普通商品
	OrderTypeGoodsMax uint8 = 2 // 租赁商品
)

// OrderTypeServiceMin/OrderTypeServiceMax 服务订单 order_type 范围
const (
	OrderTypeServiceMin uint8 = 3 // 即时服务
	OrderTypeServiceMax uint8 = 6 // 到店服务
)

// OrderCategory 按 order_type 返回订单分类：1=实物订单 2=服务订单
func OrderCategory(orderType uint8) uint8 {
	if orderType >= OrderTypeServiceMin && orderType <= OrderTypeServiceMax {
		return OrderCategoryService
	}
	return OrderCategoryGoods
}

// ============================================
// 居民健康档案状态
// ============================================
const (
	HealthRecordStatusNone     uint8 = 0 // 未建档
	HealthRecordStatusNormal   uint8 = 1 // 正常
	HealthRecordStatusArchived uint8 = 2 // 已归档
)

// ============================================
// 健康档案 - 与当前账号的亲属关系（阶段五 8.2 多档案）
// ============================================
const (
	HealthRecordRelationSelf       uint8 = 1 // 本人
	HealthRecordRelationParents    uint8 = 2 // 父母
	HealthRecordRelationOtherKin   uint8 = 3 // 其他亲属
)

// MaxHealthRecordsPerUser 单个用户账号下健康档案数量上限
const MaxHealthRecordsPerUser int = 5

// ============================================
// 健康评估登记类型
// ============================================
const (
	AssessmentTypeSelf  uint8 = 1 // 自助（C端用户）
	AssessmentTypeStaff uint8 = 2 // 服务人员登记
)

// ============================================
// 健康评估量表状态
// ============================================
const (
	AssessmentFormStatusDraft   uint8 = 0 // 草稿
	AssessmentFormStatusEnabled uint8 = 1 // 启用
)

// ============================================
// 康复辅具适配建议状态
// ============================================
const (
	FittingStatusDraft     uint8 = 0 // 草稿
	FittingStatusConfirmed uint8 = 1 // 已确认
	FittingStatusOrdered   uint8 = 2 // 已下单
)

// ============================================
// 健康宣教文章状态
// ============================================
const (
	EducationStatusDraft     uint8 = 0 // 草稿
	EducationStatusPublished uint8 = 1 // 发布
)

// ============================================
// 优惠券（PRD V2.0 阶段二：优惠券营销闭环）
// ============================================
const (
	CouponTypeThreshold uint8 = 1 // 满减券
	CouponTypeDiscount  uint8 = 2 // 折扣券
)

const (
	CouponValidTypeFixed  uint8 = 1 // 固定期限
	CouponValidTypeAfter  uint8 = 2 // 领取后N天有效
	CouponValidDaysDefault int = 7 // 领取后默认有效天数（兜底）
)

const (
	CouponScopeAll      uint8 = 1 // 全场通用
	CouponScopeCategory uint8 = 2 // 指定分类
	CouponScopeProduct  uint8 = 3 // 指定商品
)

const (
	CouponTemplateStatusEnabled  uint8 = 1 // 启用
	CouponTemplateStatusDisabled uint8 = 0 // 停用
)

const (
	UserCouponStatusUnused   uint8 = 1 // 未使用
	UserCouponStatusUsed     uint8 = 2 // 已使用
	UserCouponStatusExpired  uint8 = 3 // 已过期
	UserCouponStatusInvalid  uint8 = 4 // 已作废
)

const (
	UserCouponSourceSelf   uint8 = 1 // 自主领取
	UserCouponSourceSystem uint8 = 2 // 系统发放(30天唤回)
	UserCouponSourceManual uint8 = 3 // 运营手动发放
)

// ============================================
// 服务过程安全（PRD V2.0 阶段三）
// ============================================
const (
	AlertTypeSOS     uint8 = 1 // SOS求助
	AlertTypeTimeout uint8 = 2 // 服务超时未结束
)

const (
	AlertStatusPending    uint8 = 1 // 待处理
	AlertStatusProcessing uint8 = 2 // 处理中
	AlertStatusHandled    uint8 = 3 // 已处理
)

const (
	ServiceRecordStatusNormal  uint8 = 1 // 正常
	ServiceRecordStatusAbnormal uint8 = 2 // 异常
)

const (
	AgreementTypeUser      uint8 = 1 // 用户协议
	AgreementTypePrivacy  uint8 = 2 // 隐私政策
	AgreementTypeAuth     uint8 = 3 // 录音/定位授权协议
)

const (
	AgreementStatusPublished uint8 = 1 // 已发布(当前生效)
	AgreementStatusDraft     uint8 = 0 // 草稿/停用
)

const (
	ConsentUserTypeUser  uint8 = 1 // C端用户
	ConsentUserTypeStaff uint8 = 2 // 服务人员
)

// ServiceTimeoutAlertHours 服务超时预警阈值（小时）：签到后超过该时长未签退则触发预警
const ServiceTimeoutAlertHours = 4

// ServiceAudioRetainDays 服务录音保留天数（超过自动删除）
const ServiceAudioRetainDays = 30

// ============================================
// 服务评价（PRD V2.0 阶段四）
// ============================================
const (
	ReviewStatusVisible uint8 = 1 // 正常展示
	ReviewStatusHidden  uint8 = 0 // 后台隐藏
)

const (
	ReviewScoreMin uint8 = 1 // 评分下限
	ReviewScoreMax uint8 = 5 // 评分上限
)

// ReviewNearLimit 质量分计算取用的近 N 条评价数
const ReviewNearLimit = 100
