package upload

import (
	"fmt"
	"net/http"
	"strconv"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) GetToken(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)

	prefix := "uploads/common"
	if userType == "merchant" {
		prefix = fmt.Sprintf("uploads/merchant/%d", userID)
	} else if userType == "sp" {
		prefix = "uploads/sp"

		merchantID, err := strconv.ParseUint(c.Query("merchant_id"), 10, 64)
		if err == nil && merchantID > 0 {
			var merchant models.Merchant
			if err := database.DB.Select("id").
				Where("id = ? AND service_provider_id = ?", merchantID, userID).
				First(&merchant).Error; err != nil {
				response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权为该商家上传图片")
				return
			}

			// 服务商代理商家上传时，资源应归属到目标商家目录，避免商品图落到 uploads/sp。
			prefix = fmt.Sprintf("uploads/merchant/%d", merchantID)
		}
	}

	token, err := qiniu.GetService().GetUploadToken()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeQiniuUploadFailed, "获取上传凭证失败")
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"domain":     qiniu.GetService().Domain,
		"prefix":     prefix,
		"upload_url": qiniu.GetService().UploadURL,
	})
}

func (h *UploadHandler) Callback(c *gin.Context) {
	response.Success(c, gin.H{"message": "ok"})
}
