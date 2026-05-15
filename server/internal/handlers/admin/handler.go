package admin

import (
	"fz_yyc_api/internal/middleware"
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

func GetAdminProfile(c *gin.Context) {
	adminID := middleware.GetUserID(c)

	var admin models.ServiceProviderAdmin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "管理员不存在")
		return
	}

	response.Success(c, gin.H{
		"id":       admin.ID,
		"username": admin.Username,
		"name":     admin.Name,
		"phone":    admin.Phone,
		"role":     admin.Role,
	})
}

type UpdateAdminProfileRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func UpdateAdminProfile(c *gin.Context) {
	adminID := middleware.GetUserID(c)

	var req UpdateAdminProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "管理员不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&admin).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新失败")
			return
		}
	}

	database.DB.First(&admin, adminID)
	response.Success(c, gin.H{
		"id":       admin.ID,
		"username": admin.Username,
		"name":     admin.Name,
		"phone":    admin.Phone,
		"role":     admin.Role,
	})
}

type ChangeAdminPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func ChangeAdminPassword(c *gin.Context) {
	adminID := middleware.GetUserID(c)

	var req ChangeAdminPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改失败")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "旧密码错误")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改失败")
		return
	}

	if err := database.DB.Model(&admin).Update("password", string(hashed)).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改失败")
		return
	}

	response.Success(c, gin.H{"message": "修改成功"})
}
