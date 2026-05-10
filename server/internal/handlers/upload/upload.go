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
	} else if userType == "admin" {
		prefix = "uploads/admin"
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
