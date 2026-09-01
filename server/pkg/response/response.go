package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码定义
const (
	CodeSuccess      = 0    // 成功
	CodeParamError   = 1001 // 参数错误
	CodeUnauthorized = 1002 // 未授权
	CodeForbidden    = 1003 // 禁止访问
	CodeNotFound     = 1004 // 资源不存在
	CodeServerError  = 9001 // 服务器内部错误

	// 用户相关
	CodeUserNotFound  = 2001 // 用户不存在
	CodeUserExists    = 2002 // 用户已存在
	CodePasswordError = 2003 // 密码错误

	// 商家相关
	CodeMerchantNotFound = 3001 // 商家不存在
	CodeMerchantNotAudit = 3002 // 商家未审核
	CodeMerchantDisabled = 3003 // 商家已禁用
	CodeDeliveryDisabled = 3004 // 商家未开通配送
	CodeOutOfRange       = 3005 // 超出配送范围

	// 商品相关
	CodeProductNotFound = 4001 // 商品不存在
	CodeProductOffSale  = 4002 // 商品已下架
	CodeStockNotEnough  = 4003 // 库存不足

	// 订单相关
	CodeOrderNotFound    = 5001 // 订单不存在
	CodeOrderStatusError = 5002 // 订单状态错误
	CodeOrderPaid        = 5003 // 订单已支付
	CodeOrderCancelled   = 5004 // 订单已取消
	CodeOrderNotPermit   = 5005 // 无权操作此订单
	CodeVerifyCodeError  = 5006 // 核销码错误

	// 支付相关
	CodePayFailed     = 6001 // 支付失败
	CodeRefundFailed  = 6002 // 退款失败
	CodeApplyFailed   = 6003 // 进件申请失败
	CodeApplyAuditing = 6004 // 进件审核中

	// 分类相关
	CodeCategoryNotFound = 7001 // 分类不存在
	CodeCategoryHasProd  = 7002 // 分类下存在商品

	// 服务商相关
	CodeProviderConfigError = 8001 // 服务商配置错误
	CodeQiniuUploadFailed   = 8002 // 七牛云上传失败

	// 优惠券相关
	CodeCouponNotFound      = 8501 // 券模板不存在
	CodeCouponUnavailable   = 8502 // 不可领取/已抢光/已停用
	CodeCouponLimitExceeded = 8503 // 超出限领
	CodeCouponNotUsable     = 8504 // 券不可用（过期/已使用/门槛不满足）

	// 服务过程安全相关（阶段三）
	CodeAlertEventNotFound = 8601 // 预警事件不存在
	CodeAlertEventHandled  = 8602 // 预警事件已处理
	CodeAgreementNotFound  = 8611 // 协议不存在
	CodeServiceRecordExist = 8621 // 服务记录已存在

	// 服务评价相关（阶段四）
	CodeReviewNotFound = 8701 // 评价不存在
	CodeReviewExists   = 8702 // 该订单已评价
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（带消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithStatus 错误响应（带HTTP状态码）
func ErrorWithStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// ParamError 参数错误
func ParamError(c *gin.Context, message string) {
	Error(c, CodeParamError, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	ErrorWithStatus(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden 禁止访问
func Forbidden(c *gin.Context, message string) {
	ErrorWithStatus(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, message string) {
	Error(c, CodeServerError, message)
}

// Fail 通用错误响应（带HTTP状态码和错误码）
func Fail(c *gin.Context, httpStatus int, code int, message string) {
	ErrorWithStatus(c, httpStatus, code, message)
}

// InvalidParams 参数错误（简化调用）
func InvalidParams(c *gin.Context, message string) {
	ParamError(c, message)
}
