package sp

import (
	"errors"
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/orderquery"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetOrders(c *gin.Context) {
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	statusParam := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	keyword := c.Query("keyword")
	merchantIDParam := c.Query("merchant_id")

	var merchantID uint64
	if merchantIDParam != "" {
		parsedMerchantID, err := strconv.ParseUint(merchantIDParam, 10, 64)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "商家参数错误")
			return
		}
		merchantID = parsedMerchantID

		var merchant models.Merchant
		if err := database.DB.Select("id").
			Where("id = ? AND service_provider_id = ?", merchantID, serviceProviderID).
			First(&merchant).Error; err != nil {
			response.Fail(c, http.StatusNotFound, response.CodeMerchantNotFound, "商家不存在")
			return
		}
	}

	var status *int
	if statusParam != "" {
		parsedStatus, err := strconv.Atoi(statusParam)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单状态参数错误")
			return
		}
		status = &parsedStatus
	}

	result, err := orderquery.GetOrderList(c.Request.Context(), orderquery.ListOptions{
		MerchantID:        merchantID,
		ServiceProviderID: serviceProviderID,
		Status:            status,
		StartDate:         startDate,
		EndDate:           endDate,
		Keyword:           keyword,
		Page:              page,
		PageSize:          pageSize,
		IncludeMerchant:   true,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取订单列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": result.List,
		"pagination": gin.H{
			"total":     result.Total,
			"page":      result.Page,
			"page_size": result.PageSize,
		},
	})
}

func GetOrderDetail(c *gin.Context) {
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单参数错误")
		return
	}

	order, err := orderquery.GetOrderDetail(orderID, orderquery.DetailOptions{
		ServiceProviderID: serviceProviderID,
		IncludeMerchant:   true,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取订单详情失败")
		return
	}

	response.Success(c, order)
}
