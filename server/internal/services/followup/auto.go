package followup

import (
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
)

// CreateFollowUpTask 自动生成随访任务（在服务完成/租赁归还/评估完成时调用）。
// 仅负责插入一条待执行的随访任务：计划随访时间 = 当前时间 + 72 小时，
// staff_id 为 0 时存 NULL（表示待认领），status 默认待执行，remark 默认空串。
// 错误原样返回，由调用方忽略或记录日志，不影响业务主流程（函数内部不 panic）。
func CreateFollowUpTask(userID uint64, taskType, sourceType uint8, sourceID, staffID uint64) error {
	var sourceIDPtr *uint64
	if sourceID > 0 {
		sourceIDPtr = &sourceID
	}
	var staffIDPtr *uint64
	if staffID > 0 {
		staffIDPtr = &staffID
	}

	planFollowTime := time.Now().Add(72 * time.Hour)
	task := models.FollowUpTask{
		UserID:         userID,
		TaskType:       taskType,
		SourceType:     sourceType,
		SourceID:       sourceIDPtr,
		PlanFollowTime: &planFollowTime,
		StaffID:        staffIDPtr,
		ContactMethod:  0,
		Status:         0,
		Remark:         "",
	}
	return database.DB.Create(&task).Error
}
