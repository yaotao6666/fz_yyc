package merchant

import (
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================
// 阶段三：订单服务记录查询（服务记录 + 轨迹 + 录音）
// ============================================

// GetOrderServiceRecord 订单服务记录详情（含轨迹点，录音为私有链接）
func GetOrderServiceRecord(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil || orderID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var order models.Order
	if err := database.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	var record models.ServiceRecord
	if err := database.DB.Where("order_id = ?", orderID).First(&record).Error; err != nil {
		response.Success(c, gin.H{
			"order_id":       orderID,
			"service_record": nil,
			"tracks":         []models.ServiceLocationTrack{},
		})
		return
	}

	// 轨迹点（按上报时间正序）
	var tracks []models.ServiceLocationTrack
	database.DB.Where("order_id = ?", orderID).Order("reported_at ASC").Limit(1000).Find(&tracks)

	// 服务人员信息快照
	var staff models.ServiceStaff
	staffInfo := gin.H{}
	if err := database.DB.Select("id, name, phone, service_region, quality_score").
		Where("id = ?", record.StaffID).First(&staff).Error; err == nil {
		staffInfo = gin.H{
			"id":             staff.ID,
			"name":           staff.Name,
			"phone":          staff.Phone,
			"service_region": staff.ServiceRegion,
			"quality_score":  staff.QualityScore,
		}
	}

	// 录音私有链接（已删除则不返回）
	audioURL := ""
	if record.AudioURL != "" && record.AudioDeletedAt == nil {
		audioURL = qiniu.GetService().BuildPrivateURL(record.AudioURL)
	}

	response.Success(c, gin.H{
		"order_id": orderID,
		"service_record": gin.H{
			"id":                record.ID,
			"staff_id":          record.StaffID,
			"staff":             staffInfo,
			"start_time":        record.StartTime,
			"end_time":          record.EndTime,
			"duration_minutes":  serviceDurationMinutes(record),
			"gps_track_url":     record.GPSTrackURL,
			"audio_url":         audioURL,
			"audio_uploaded_at": record.AudioUploadedAt,
			"audio_deleted_at":  record.AudioDeletedAt,
			"sos_triggered":     record.SOSTriggered,
			"status":            record.Status,
			"created_at":         record.CreatedAt,
		},
		"tracks":        tracks,
		"track_count":   len(tracks),
	})
}

// serviceDurationMinutes 服务时长（分钟），未结束返回-1
func serviceDurationMinutes(record models.ServiceRecord) int {
	if record.StartTime == nil || record.EndTime == nil {
		return -1
	}
	return int(record.EndTime.Sub(*record.StartTime).Minutes())
}
