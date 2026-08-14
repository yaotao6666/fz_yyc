package merchant

import (
	"net/http"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

func resolveTargetMerchantID(c *gin.Context) (uint64, bool) {
	userType := middleware.GetUserType(c)
	if userType == "merchant" {
		return middleware.GetMerchantID(c), true
	}

	response.Fail(c, http.StatusForbidden, response.CodeForbidden, "需要商家权限")
	return 0, false
}
