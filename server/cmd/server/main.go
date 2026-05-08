package main

import (
	"log"
	"os"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/handlers/admin"
	"fz_yyc_api/internal/handlers/merchant"
	"fz_yyc_api/internal/handlers/upload"
	"fz_yyc_api/internal/handlers/user"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(&config.Config.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseDB()

	// 初始化七牛云
	if err := qiniu.InitQiniu(); err != nil {
		log.Fatalf("七牛云初始化失败: %v", err)
	}

	// 设置Gin模式
	if config.Config.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin实例
	r := gin.Default()

	// 全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 路由设置
	setupRoutes(r)

	// 启动服务器
	addr := config.Config.App.GetAddr()
	log.Printf("服务启动中，监听地址: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
		os.Exit(1)
	}
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/admin/login", admin.Login)
			authGroup.POST("/merchant/login", merchant.Login)
			authGroup.POST("/user/wechat-login", user.WechatLogin)
		}

		// 文件上传接口
		uploadHandler := upload.NewUploadHandler()
		uploadGroup := v1.Group("/upload")
		uploadGroup.Use(middleware.JWTAuth())
		{
			uploadGroup.GET("/token", uploadHandler.GetToken)
		}

		// C端用户接口（无需特殊权限）
		storeGroup := v1.Group("/store/:merchant_id")
		{
			storeGroup.GET("/home", user.GetStoreHome)
			storeGroup.GET("/products", user.GetProducts)
			storeGroup.GET("/products/:product_id", user.GetProductDetail)
			storeGroup.GET("/delivery-rules", user.GetDeliveryRules)
		}

		// C端用户接口（需要登录）
		userGroup := v1.Group("/user")
		userGroup.Use(middleware.JWTAuth())
		{
			userGroup.POST("/orders", user.CreateOrder)
			userGroup.GET("/orders", user.GetOrders)
			userGroup.GET("/orders/:order_id", user.GetOrderDetail)
			userGroup.POST("/orders/:order_id/cancel", user.CancelOrder)
			userGroup.POST("/orders/:order_id/refund", user.ApplyRefund)
		}

		// 服务商管理员接口
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.JWTAuth())
		{
			// 服务商配置
			adminGroup.GET("/service-provider", admin.GetServiceProvider)
			adminGroup.PUT("/service-provider", admin.UpdateServiceProvider)

			// 商家进件管理
			adminGroup.GET("/merchant-applications", admin.GetMerchantApplications)
			adminGroup.GET("/merchant-applications/:id", admin.GetMerchantApplicationDetail)
			adminGroup.POST("/merchant-applications/:id/submit", admin.SubmitMerchantApplication)
			adminGroup.GET("/merchant-applications/:id/status", admin.GetMerchantApplicationStatus)

			// 商家审核
			adminGroup.GET("/merchants/pending", admin.GetPendingMerchants)
			adminGroup.POST("/merchants/:merchant_id/audit", admin.AuditMerchant)

			// 数据看板
			adminGroup.GET("/dashboard", admin.GetDashboard)

			// 活动管理
			adminGroup.GET("/activities", admin.GetActivities)
			adminGroup.POST("/activities", admin.CreateActivity)
			adminGroup.PUT("/activities/:id", admin.UpdateActivity)
			adminGroup.DELETE("/activities/:id", admin.DeleteActivity)

			// 邀请入驻管理
			adminGroup.GET("/invites/stats", admin.GetInviteStats)
			adminGroup.GET("/invites", admin.GetInviteRecords)
			adminGroup.GET("/invite-rewards", admin.GetInviteRewards)
			adminGroup.PUT("/invite-rewards", admin.UpdateInviteRewards)
		}

		// 商家管理员接口
		merchantGroup := v1.Group("/merchant")
		merchantGroup.Use(middleware.JWTAuth())
		{
			// 商家信息
			merchantGroup.GET("/profile", merchant.GetProfile)
			merchantGroup.PUT("/profile", merchant.UpdateProfile)
			merchantGroup.GET("/settings", merchant.GetSettings)
			merchantGroup.PUT("/settings", merchant.UpdateSettings)
			merchantGroup.PUT("/license", merchant.UpdateLicense)
			merchantGroup.PUT("/bank-account", merchant.UpdateBankAccount)
			merchantGroup.POST("/status", merchant.UpdateStatus)
			merchantGroup.GET("/application/status", merchant.GetApplicationStatus)
			merchantGroup.GET("/qrcode", merchant.GetQRCode)
			merchantGroup.GET("/delivery-settings", merchant.GetDeliverySettings)

			// 邀请入驻
			merchantGroup.POST("/invite/generate", merchant.GenerateInviteCode)
			merchantGroup.GET("/invite/info", merchant.GetInviteInfo)
			merchantGroup.GET("/invite/records", merchant.GetInviteRecords)

			// 商品分类
			merchantGroup.GET("/categories", merchant.GetCategories)
			merchantGroup.POST("/categories", merchant.CreateCategory)
			merchantGroup.PUT("/categories/:category_id", merchant.UpdateCategory)
			merchantGroup.DELETE("/categories/:category_id", merchant.DeleteCategory)
			merchantGroup.POST("/categories/sort", merchant.SortCategories)

			// 商品管理
			merchantGroup.GET("/products", merchant.GetProducts)
			merchantGroup.GET("/products/:product_id", merchant.GetProduct)
			merchantGroup.POST("/products", merchant.CreateProduct)
			merchantGroup.PUT("/products/:product_id", merchant.UpdateProduct)
			merchantGroup.POST("/products/:product_id/on-sale", merchant.ProductOnSale)
			merchantGroup.POST("/products/:product_id/off-sale", merchant.ProductOffSale)
			merchantGroup.POST("/products/batch-status", merchant.BatchUpdateProductStatus)
			merchantGroup.DELETE("/products/:product_id", merchant.DeleteProduct)
			merchantGroup.PUT("/products/:product_id/stock", merchant.UpdateStock)

			// 订单管理
			merchantGroup.GET("/orders", merchant.GetOrders)
			merchantGroup.GET("/orders/:order_id", merchant.GetOrderDetail)
			merchantGroup.POST("/orders/:order_id/complete", merchant.CompleteOrder)
			merchantGroup.POST("/orders/:order_id/refund", merchant.RefundOrder)
			merchantGroup.GET("/orders/statistics", merchant.GetOrderStatistics)

			// 数据分析
			merchantGroup.GET("/analytics/overview", merchant.GetAnalyticsOverview)
			merchantGroup.GET("/analytics/sales-trend", merchant.GetSalesTrend)
			merchantGroup.GET("/analytics/product-ranking", merchant.GetProductRanking)
			merchantGroup.GET("/analytics/hourly", merchant.GetHourlyAnalysis)
			merchantGroup.GET("/analytics/stock-alert", merchant.GetStockAlert)
			merchantGroup.GET("/analytics/customers", merchant.GetCustomerAnalysis)
			merchantGroup.GET("/analytics/customer-trend", merchant.GetCustomerTrend)
		}

		// 微信支付回调
		notifyGroup := v1.Group("/notify")
		{
			notifyGroup.POST("/payment", admin.PaymentNotify)
		}
	}
}
