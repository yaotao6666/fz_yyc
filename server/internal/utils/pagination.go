package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"fz_yyc_api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	orderNoRandomDigits    = 6
	orderNoMerchantMaxSize = 12
)

// Pagination 分页参数
type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// GetPagination 获取分页参数
func GetPagination(c *gin.Context) *Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// GetOffset 获取偏移量
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetTotalPages 获取总页数
func (p *Pagination) GetTotalPages() int {
	if p.Total == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.Total) / float64(p.PageSize)))
}

// IsLastPage 是否最后一页
func (p *Pagination) IsLastPage() bool {
	return p.Page >= p.GetTotalPages()
}

func generateFixedDigitsRandom(length int) string {
	if length <= 0 {
		return ""
	}

	upperBound := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	randomNumber, err := rand.Int(rand.Reader, upperBound)
	if err != nil {
		return fmt.Sprintf("%0*d", length, time.Now().UnixNano()%int64(upperBound.Int64()))
	}
	return fmt.Sprintf("%0*d", length, randomNumber.Int64())
}

// GenerateOrderNo 生成订单号
func GenerateOrderNo(merchantID uint64) string {
	now := time.Now()
	dateStr := now.Format("20060102150405")
	merchantIDStr := strconv.FormatUint(merchantID, 10)
	if len(merchantIDStr) > orderNoMerchantMaxSize {
		merchantIDStr = merchantIDStr[len(merchantIDStr)-orderNoMerchantMaxSize:]
	}
	random := generateFixedDigitsRandom(orderNoRandomDigits)
	return dateStr + merchantIDStr + random
}

// GenerateVerifyCode 生成核销码（6位数字）
func GenerateVerifyCode() string {
	now := time.Now()
	return fmt.Sprintf("%06d", now.UnixNano()%1000000)
}

// GenerateRefundNo 生成退款单号
func GenerateRefundNo() string {
	now := time.Now()
	dateStr := now.Format("20060102150405")
	random := fmt.Sprintf("%04d", now.UnixNano()%10000)
	return "RF" + dateStr + random
}

// GenerateRandomNumber 生成指定范围的随机数
func GenerateRandomNumber(min, max int) int {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + int(time.Now().UnixNano()%int64(max-min))
}

// CalculateDeliveryFee 计算配送费
func CalculateDeliveryFee(totalAmount, baseFee, freeDeliveryAmount float64, distance float64, rules []map[string]interface{}) float64 {
	if totalAmount >= freeDeliveryAmount && freeDeliveryAmount > 0 {
		return 0
	}

	for _, rule := range rules {
		minDist, okMin := rule["min_distance"].(float64)
		maxDist, okMax := rule["max_distance"].(float64)
		fee, okFee := rule["fee"].(float64)
		if !okMin || !okMax || !okFee {
			continue
		}
		if distance >= minDist && distance <= maxDist {
			return fee
		}
	}

	return baseFee
}

type TokenClaims struct {
	UserID   uint64 `json:"user_id"`
	StaffID  uint64 `json:"staff_id,omitempty"` // 商家后台员工ID（merchant类型token）
	UserType string `json:"user_type"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, staffID uint64, userType, username string) (string, error) {
	claims := &TokenClaims{
		UserID:   userID,
		StaffID:  staffID,
		UserType: userType,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Config.JWT.Expire) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Config.JWT.Secret))
}

func GetUserID(c *gin.Context) uint64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	if id, ok := userID.(uint64); ok {
		return id
	}
	return 0
}
