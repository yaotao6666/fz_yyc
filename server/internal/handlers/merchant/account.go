package merchant

import (
	"net/http"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func getCurrentMerchantStaff(c *gin.Context) (*models.MerchantStaff, error) {
	return middleware.GetCurrentStaff(c)
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	staff, err := getCurrentMerchantStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工信息失败")
		return
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(req.OldPassword)); compareErr != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "原密码错误")
		return
	}

	if req.OldPassword == req.NewPassword {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "新旧密码不能相同")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "密码加密失败")
		return
	}

	if err := database.DB.Model(&models.MerchantStaff{}).
		Where("id = ?", staff.ID).
		Update("password", string(hashedPassword)).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改密码失败")
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}