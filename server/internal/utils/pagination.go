package utils

import (
	"fmt"
	"math"
	"strconv"
	"time"
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

// GenerateOrderNo 生成订单号
func GenerateOrderNo() string {
	now := time.Now()
	dateStr := now.Format("20060102150405")
	random := fmt.Sprintf("%04d", now.UnixNano()%10000)
	return dateStr + random
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
		minDist := rule["min_distance"].(float64)
		maxDist := rule["max_distance"].(float64)
		if distance >= minDist && distance < maxDist {
			return rule["fee"].(float64)
		}
	}

	return baseFee
}
