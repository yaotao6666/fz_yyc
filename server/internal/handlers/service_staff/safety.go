package service_staff

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================
// 阶段三：服务过程安全（定位上报 / 录音提交 / SOS / 服务区域）
// ============================================

// LocationReportRequest 定位上报请求
type LocationReportRequest struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

// ReportLocation 服务中工单定时上报定位（前端 60s 一次）
func ReportLocation(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || orderID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req LocationReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	// 仅服务中（biz_status=3）且本人的工单可上报
	var order models.Order
	if err := database.DB.Where("id = ? AND assigned_staff_id = ?", orderID, staff.ID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "订单不存在")
		return
	}
	if order.BizStatus != 3 {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderStatusError, "仅服务中的工单可上报定位")
		return
	}

	track := models.ServiceLocationTrack{
		OrderID:    orderID,
		StaffID:   staff.ID,
		Lat:       req.Lat,
		Lng:       req.Lng,
		ReportedAt: time.Now(),
	}
	if err := database.DB.Create(&track).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "定位上报失败")
		return
	}

	response.Success(c, gin.H{"reported_at": track.ReportedAt})
}

// AudioSubmitRequest 录音提交请求（客户端七牛直传后提交 URL）
type AudioSubmitRequest struct {
	AudioURL string `json:"audio_url" binding:"required"`
}

// SubmitAudio 提交服务录音 URL（写入/更新服务记录）
func SubmitAudio(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || orderID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req AudioSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	// 仅本人接的订单可提交录音
	var order models.Order
	if err := database.DB.Where("id = ? AND assigned_staff_id = ?", orderID, staff.ID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "订单不存在")
		return
	}

	now := time.Now()
	var record models.ServiceRecord
	dbErr := database.DB.Where("order_id = ?", orderID).First(&record).Error
	if dbErr == nil {
		// 已有记录：仅当录音为空或已删除时可覆盖
		if record.AudioURL != "" && record.AudioDeletedAt == nil {
			response.Fail(c, http.StatusBadRequest, response.CodeServiceRecordExist, "该订单已提交过录音")
			return
		}
		updates := map[string]interface{}{
			"audio_url":         req.AudioURL,
			"audio_uploaded_at": now,
			"audio_deleted_at":  nil,
		}
		if err := database.DB.Model(&record).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "录音提交失败")
			return
		}
	} else {
		// 无记录：创建（签到前提交录音的场景）
		record = models.ServiceRecord{
			OrderID:         orderID,
			StaffID:         staff.ID,
			AudioURL:        req.AudioURL,
			AudioUploadedAt: &now,
			Status:          utils.ServiceRecordStatusNormal,
		}
		if err := database.DB.Create(&record).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "录音提交失败")
			return
		}
	}

	audioURL := req.AudioURL
	response.SuccessWithMessage(c, "录音提交成功", gin.H{
		"order_id":          orderID,
		"audio_url":         qiniu.GetService().BuildPrivateURL(audioURL),
		"audio_uploaded_at": now,
	})
}

// SOSRequest 一键SOS请求
type SOSRequest struct {
	OrderID *uint64 `json:"order_id"`
	Lat     float64 `json:"lat" binding:"required"`
	Lng     float64 `json:"lng" binding:"required"`
	Address string  `json:"address"`
}

// SOS 一键SOS（写预警事件）
func SOS(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	var req SOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	// 若携带订单ID，校验归属并同步标记服务记录
	sosTriggered := false
	if req.OrderID != nil && *req.OrderID > 0 {
		var order models.Order
		if err := database.DB.Where("id = ? AND assigned_staff_id = ?", *req.OrderID, staff.ID).
			First(&order).Error; err == nil {
			sosTriggered = true
			if err := database.DB.Model(&models.ServiceRecord{}).
				Where("order_id = ?", order.ID).
				Updates(map[string]interface{}{
					"sos_triggered": 1,
					"status":        utils.ServiceRecordStatusAbnormal,
				}).Error; err != nil {
				// 标记失败不影响SOS主流程
				_ = err
			}
		}
	}

	event := models.ServiceAlertEvent{
		OrderID:   req.OrderID,
		StaffID:   &staff.ID,
		AlertType: utils.AlertTypeSOS,
		Lat:       req.Lat,
		Lng:       req.Lng,
		Address:   req.Address,
		Status:    utils.AlertStatusPending,
	}
	if err := database.DB.Create(&event).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "SOS提交失败，请立即电话联系商家")
		return
	}

	response.SuccessWithMessage(c, "SOS已发出，商家将尽快处理", gin.H{
		"alert_event_id": event.ID,
		"created_at":     event.CreatedAt,
		"sos_marked":     sosTriggered,
	})
}

// GetMyRegion 查看我的服务区域（变更走资料变更审核通道）
func GetMyRegion(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	regions := parseRegions(staff.ServiceRegion)
	response.Success(c, gin.H{
		"service_region":    staff.ServiceRegion,
		"region_list":      regions,
		"region_limitless": len(regions) == 0, // 空=不限区域
		"pending_fields":   staff.PendingFields,
	})
}

// parseRegions 解析逗号分隔的服务区域
func parseRegions(region string) []string {
	if strings.TrimSpace(region) == "" {
		return []string{}
	}
	parts := strings.Split(region, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			result = append(result, p)
		}
	}
	return result
}

// StaffRegionMatch 判断订单区县是否落在人员服务区域内（空=不限）
func StaffRegionMatch(staffRegion string, deliveryDistrict string) bool {
	regions := parseRegions(staffRegion)
	if len(regions) == 0 {
		return true
	}
	district := strings.TrimSpace(deliveryDistrict)
	if district == "" {
		return false
	}
	for _, r := range regions {
		if r == district || strings.Contains(district, r) || strings.Contains(r, district) {
			return true
		}
	}
	return false
}
