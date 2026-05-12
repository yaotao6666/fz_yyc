package middleware

import (
	"strings"
	"time"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT Claims
type Claims struct {
	UserID   uint64 `json:"user_id"`
	UserType string `json:"user_type"` // admin, merchant, user
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成Token
func GenerateToken(userID uint64, userType, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
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

// ParseToken 解析Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}

		// 解析Token
		claims, err := ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存储到Context
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// MerchantAuth 商家认证中间件
func MerchantAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "merchant" {
			response.Forbidden(c, "需要商家权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserAuth 用户认证中间件
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "user" {
			response.Forbidden(c, "需要用户权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserID 获取当前用户ID
func GetUserID(c *gin.Context) uint64 {
	userID, _ := c.Get("user_id")
	return userID.(uint64)
}

// GetUserType 获取当前用户类型
func GetUserType(c *gin.Context) string {
	userType, _ := c.Get("user_type")
	return userType.(string)
}

// GetMerchantID 获取当前商家ID
func GetMerchantID(c *gin.Context) uint64 {
	return GetUserID(c)
}

// GetUsername 获取当前登录用户名
func GetUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	if username == nil {
		return ""
	}
	return username.(string)
}
