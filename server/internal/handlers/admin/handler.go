package admin

import (
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.Where("username = ? AND status = ?", req.Username, 1).First(&admin).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	now := time.Now()
	database.DB.Model(&admin).Update("last_login_at", now)

	token, _ := utils.GenerateToken(admin.ID, "admin", admin.Username)
	response.Success(c, gin.H{
		"token": token,
		"admin": admin,
	})
}

func PaymentNotify(c *gin.Context) {
	// 微信支付回调处理
	// 具体实现待完成
	response.Success(c, gin.H{"message": "回调处理成功"})
}
