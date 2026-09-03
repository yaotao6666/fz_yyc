package upload

import (
	"fmt"
	"net/http"

	"fz_yyc_api/internal/middleware"
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
	} else if userType == "service_staff" {
		prefix = fmt.Sprintf("uploads/service-staff/%d", userID)
	}

	// 录音等非图片上传：?mime=audio/*（默认 image/*）
	// 注意：微信开发者工具上 getRecorderManager 即使声明 mp3，实际产出 video/webm；
	//      真机上也可能产出 silk/mp3/m4a 等。录音凭证统一放宽为 audio/* + video/*。
	mimeLimit := c.DefaultQuery("mime", "image/*")
	switch mimeLimit {
	case "audio/*":
		mimeLimit = "audio/*;video/*"
	default:
		mimeLimit = "image/*"
	}

	token, err := qiniu.GetService().GetUploadTokenWithMime(mimeLimit)
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

// SignRequest 单资源签名请求
type SignRequest struct {
	URL string `json:"url" binding:"required"`
}

// Sign 对单个七牛资源地址进行私有签名，供富文本编辑器实时预览粘贴/上传的图片使用
func (h *UploadHandler) Sign(c *gin.Context) {
	var req SignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	signed := qiniu.GetService().BuildPrivateURL(req.URL)
	response.Success(c, gin.H{"url": signed})
}
