package utils

// DefaultMerchantID 全局单例商家ID。
// 系统已深度重构为单商户模式，merchants 表仅保留 id=1 一行作为全局配置，
// 其余业务表不再携带 merchant_id，业务侧统一使用该常量读取单例配置。
const DefaultMerchantID uint64 = 1

// ============================================
// 居民健康档案状态
// ============================================
const (
	HealthRecordStatusNone     uint8 = 0 // 未建档
	HealthRecordStatusNormal   uint8 = 1 // 正常
	HealthRecordStatusArchived uint8 = 2 // 已归档
)

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
// 照护计划类型
// ============================================
const (
	CarePlanTypeLifeCare      uint8 = 1 // 生活照料
	CarePlanTypeBaseNursing   uint8 = 2 // 基础护理
	CarePlanTypeRehabTraining uint8 = 3 // 康复训练
	CarePlanTypeComprehensive uint8 = 4 // 综合康养
)

// ============================================
// 照护计划状态
// ============================================
const (
	CarePlanStatusDraft     uint8 = 0 // 草稿
	CarePlanStatusActive    uint8 = 1 // 执行中
	CarePlanStatusPaused    uint8 = 2 // 已暂停
	CarePlanStatusCompleted uint8 = 3 // 已完成
)

// ============================================
// 随访任务类型
// ============================================
const (
	FollowUpTypeRehab       uint8 = 1 // 康复随访
	FollowUpTypeReturnVisit uint8 = 2 // 租后回访
	FollowUpTypeChronic     uint8 = 3 // 慢病随访
	FollowUpTypeAssessment  uint8 = 4 // 评估回访
)

// ============================================
// 随访任务来源
// ============================================
const (
	FollowUpSourceService      uint8 = 1 // 服务完成
	FollowUpSourceRentalReturn uint8 = 2 // 租赁归还
	FollowUpSourceAssessment   uint8 = 3 // 评估完成
	FollowUpSourceManual       uint8 = 4 // 手动
)

// ============================================
// 随访任务状态
// ============================================
const (
	FollowUpStatusPending   uint8 = 0 // 待执行
	FollowUpStatusCompleted uint8 = 1 // 已完成
	FollowUpStatusSkipped   uint8 = 2 // 已跳过
)

// ============================================
// 随访方式
// ============================================
const (
	FollowUpMethodPhone  uint8 = 1 // 电话
	FollowUpMethodVisit  uint8 = 2 // 上门
	FollowUpMethodWechat uint8 = 3 // 微信
)

// ============================================
// 生命体征监测类型
// ============================================
const (
	MonitoringTypeBP     uint8 = 1 // 血压
	MonitoringTypeBS     uint8 = 2 // 血糖
	MonitoringTypeHR     uint8 = 3 // 心率
	MonitoringTypeSpO2   uint8 = 4 // 血氧
	MonitoringTypeWeight uint8 = 5 // 体重
)

// ============================================
// 健康宣教文章状态
// ============================================
const (
	EducationStatusDraft     uint8 = 0 // 草稿
	EducationStatusPublished uint8 = 1 // 发布
)
