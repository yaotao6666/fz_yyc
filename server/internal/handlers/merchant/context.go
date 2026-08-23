package merchant

import (
	"fz_yyc_api/internal/utils"

	"github.com/gin-gonic/gin"
)

// resolveTargetMerchantID 单商户模式下固定返回全局单例商家ID
func resolveTargetMerchantID(c *gin.Context) (uint64, bool) {
	return utils.DefaultMerchantID, true
}
