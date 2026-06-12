package sp

import (
	"net/http"
	"strconv"

	merchantHandlers "fz_yyc_api/internal/handlers/merchant"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func runAsMerchant(c *gin.Context, fn func(*gin.Context)) {
	merchantID, _ := strconv.ParseUint(c.Param("merchant_id"), 10, 64)
	if merchantID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	var merchant models.Merchant
	if err := database.DB.Select("id").
		Where("id = ? AND service_provider_id = ?", merchantID, serviceProviderID).
		First(&merchant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询商家失败")
		return
	}

	originalUserID, hasUserID := c.Get("user_id")
	originalUserType, hasUserType := c.Get("user_type")
	originalUsername, hasUsername := c.Get("username")

	c.Set("user_id", merchantID)
	c.Set("user_type", "merchant")

	defer func() {
		if hasUserID {
			c.Set("user_id", originalUserID)
		} else {
			if c.Keys != nil {
				delete(c.Keys, "user_id")
			}
		}
		if hasUserType {
			c.Set("user_type", originalUserType)
		} else {
			if c.Keys != nil {
				delete(c.Keys, "user_type")
			}
		}
		if hasUsername {
			c.Set("username", originalUsername)
		} else {
			if c.Keys != nil {
				delete(c.Keys, "username")
			}
		}
	}()

	fn(c)
}

func GetMerchantCategories(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.GetCategories)
}

func CreateMerchantCategory(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.CreateCategory)
}

func UpdateMerchantCategory(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.UpdateCategory)
}

func DeleteMerchantCategory(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.DeleteCategory)
}

func SortMerchantCategories(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.SortCategories)
}

func GetMerchantProducts(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.GetProducts)
}

func GetMerchantProduct(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.GetProduct)
}

func CreateMerchantProduct(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.CreateProduct)
}

func UpdateMerchantProduct(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.UpdateProduct)
}

func MerchantProductOnSale(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.ProductOnSale)
}

func MerchantProductOffSale(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.ProductOffSale)
}

func BatchUpdateMerchantProductStatus(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.BatchUpdateProductStatus)
}

func DeleteMerchantProduct(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.DeleteProduct)
}

func UpdateMerchantProductStock(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.UpdateStock)
}

func GetMerchantProductSpecs(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.GetProductSpecs)
}

func UpdateMerchantProductSpecs(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.UpdateProductSpecs)
}

func DeleteMerchantProductSpecs(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.DeleteProductSpecs)
}

func GetMerchantPickupPoints(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.GetPickupPoints)
}

func CreateMerchantPickupPoint(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.CreatePickupPoint)
}

func UpdateMerchantPickupPoint(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.UpdatePickupPoint)
}

func DeleteMerchantPickupPoint(c *gin.Context) {
	runAsMerchant(c, merchantHandlers.DeletePickupPoint)
}
