package main

import (
	"log"
	"os"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/handlers/admin"
	"fz_yyc_api/internal/handlers/merchant"
	"fz_yyc_api/internal/handlers/sp"
	"fz_yyc_api/internal/handlers/upload"
	"fz_yyc_api/internal/handlers/user"
	wsHandler "fz_yyc_api/internal/handlers/ws"
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
		wsGroup := v1.Group("/ws")
		wsGroup.Use(middleware.JWTAuth())
		{
			wsGroup.GET("/merchant", wsHandler.MerchantWS)
		}

		devGroup := v1.Group("/dev")
		{
			devGroup.POST("/order-notify", wsHandler.DevOrderNotify)
			devGroup.POST("/store-visit-notify", wsHandler.DevStoreVisitNotify)
		}

		// 认证相关
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/admin/login", admin.Login)
			authGroup.POST("/merchant/login", merchant.Login)
			authGroup.POST("/merchant/wechat-login", merchant.WechatQuickLogin)
			authGroup.POST("/user/wechat-login", user.WechatLogin)
		}

		v1.POST("/merchant/register", merchant.Register)
		v1.POST("/user/auth/wechat-login", user.WechatLogin)
		v1.GET("/user/merchant/status", user.GetMerchantStatus)

		// 文件上传接口
		uploadHandler := upload.NewUploadHandler()
		uploadPublicGroup := v1.Group("/upload")
		{
			uploadPublicGroup.POST("/callback", uploadHandler.Callback)

			uploadGroup := uploadPublicGroup.Group("")
			uploadGroup.Use(middleware.JWTAuth())
			{
				uploadGroup.GET("/token", uploadHandler.GetToken)
			}
		}

		// C端店铺公开接口（无需登录即可访问）
		storeGroup := v1.Group("/store/:merchant_id")
		{
			storeGroup.GET("/home", user.GetStoreHome)
			storeGroup.GET("/products", user.GetProducts)
			storeGroup.GET("/products/:product_id", user.GetProductDetail)
			storeGroup.GET("/delivery-rules", user.GetDeliveryRules)
			storeGroup.POST("/orders", user.CreateOrder)
			storeGroup.POST("/visit", user.RecordUserVisit)
			storeGroup.POST("/event", user.RecordBehaviorEvent)
		}

		// C端用户接口（需要登录）
		userGroup := v1.Group("/user")
		userGroup.Use(middleware.JWTAuth())
		{
			userGroup.GET("/orders", user.GetOrders)
			userGroup.GET("/orders/:order_id", user.GetOrderDetail)
			userGroup.POST("/orders/:order_id/cancel", user.CancelOrder)
			userGroup.POST("/orders/:order_id/refund", user.ApplyRefund)
			userGroup.GET("/addresses", user.GetAddresses)
			userGroup.POST("/addresses", user.CreateAddress)
			userGroup.PUT("/addresses/:id", user.UpdateAddress)
			userGroup.DELETE("/addresses/:id", user.DeleteAddress)
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
			merchantGroup.POST("/account/change-password", merchant.ChangePassword)
			merchantGroup.POST("/account/wechat/bind", merchant.BindWechat)
			merchantGroup.DELETE("/account/wechat/bind", merchant.UnbindWechat)
			merchantGroup.PUT("/license", merchant.UpdateLicense)
			merchantGroup.PUT("/bank-account", merchant.UpdateBankAccount)
			merchantGroup.POST("/status", merchant.UpdateStatus)
			merchantGroup.GET("/application/status", merchant.GetApplicationStatus)
			merchantGroup.GET("/qrcode", merchant.GetQRCode)
			merchantGroup.GET("/delivery-settings", merchant.GetDeliverySettings)
			merchantGroup.PUT("/delivery-settings", merchant.UpdateDeliverySettings)
			merchantGroup.GET("/subscriptions", merchant.GetSubscriptions)
			merchantGroup.PUT("/subscriptions", merchant.UpdateSubscriptions)

			// 员工管理
			merchantGroup.GET("/staff", merchant.GetStaffList)
			merchantGroup.POST("/staff", merchant.CreateStaff)
			merchantGroup.PUT("/staff/:id", merchant.UpdateStaff)
			merchantGroup.DELETE("/staff/:id", merchant.DeleteStaff)
			merchantGroup.POST("/staff/:id/reset-password", merchant.ResetStaffPassword)

			// 系统公告（商家查看）
			merchantGroup.GET("/announcements", merchant.GetAnnouncements)
			merchantGroup.GET("/announcements/:id", merchant.GetAnnouncementDetail)

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
			merchantGroup.GET("/products/:product_id/specs", merchant.GetProductSpecs)
			merchantGroup.PUT("/products/:product_id/specs", merchant.UpdateProductSpecs)
			merchantGroup.DELETE("/products/:product_id/specs", merchant.DeleteProductSpecs)

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

			merchantGroup.GET("/printers", merchant.GetPrinters)
			merchantGroup.POST("/printers", merchant.CreatePrinter)
			merchantGroup.PUT("/printers/:printer_id", merchant.UpdatePrinter)
			merchantGroup.DELETE("/printers/:printer_id", merchant.DeletePrinter)
			merchantGroup.POST("/printers/:printer_id/test", merchant.TestPrinter)
			merchantGroup.GET("/print-logs", merchant.GetPrintLogs)
		}

		// 服务商小程序接口
		spPublicGroup := v1.Group("/sp")
		{
			spPublicGroup.POST("/auth/login", sp.Login)

			spGroup := spPublicGroup.Group("")
			spGroup.Use(middleware.JWTAuth())
			{
				spGroup.POST("/auth/logout", sp.Logout)
				spGroup.GET("/dashboard", sp.GetDashboard)
				spGroup.GET("/merchants/pending", sp.GetPendingMerchants)
				spGroup.GET("/merchants/:merchant_id", sp.GetMerchantDetail)
				spGroup.POST("/merchants/:merchant_id/approve", sp.AuditMerchant)
				spGroup.GET("/merchants/analytics/distribution", sp.GetMerchantDistribution)
				spGroup.GET("/merchants/list", sp.GetMerchantList)
				spGroup.GET("/merchants/audit-records", sp.GetAuditRecords)
				spGroup.GET("/merchants/:merchant_id/fee", sp.GetMerchantFee)
				spGroup.GET("/merchants/:merchant_id/rate", sp.GetMerchantRate)
				spGroup.POST("/merchants/:merchant_id/rate", sp.SetMerchantRate)
				spGroup.GET("/merchants/:merchant_id/qrcode", sp.GetMerchantQRCode)
				spGroup.GET("/orders/analytics", sp.GetOrderAnalytics)
				spGroup.GET("/amount/analytics", sp.GetAmountAnalytics)
				spGroup.GET("/amount/top-merchants", sp.GetTopMerchants)
				spGroup.GET("/orders/refunds", sp.GetRefunds)
				spGroup.GET("/settings", sp.GetSettings)
				spGroup.PUT("/settings", sp.UpdateSettings)
				spGroup.POST("/account/change-password", sp.ChangePassword)
				spGroup.GET("/announcements", sp.GetAnnouncements)
				spGroup.GET("/announcements/:id", sp.GetAnnouncementDetail)
				spGroup.POST("/announcements", sp.CreateAnnouncement)
				spGroup.PUT("/announcements/:id", sp.UpdateAnnouncement)
				spGroup.DELETE("/announcements/:id", sp.DeleteAnnouncement)

				// 商家进件管理
				spGroup.GET("/merchant-applications", sp.GetMerchantApplications)
				spGroup.GET("/merchant-applications/:id", sp.GetMerchantApplicationDetail)
				spGroup.POST("/merchant-applications/:id/submit", sp.SubmitMerchantApplication)
				spGroup.GET("/merchant-applications/:id/status", sp.GetMerchantApplicationStatus)

				// 活动管理
				spGroup.GET("/activities", sp.GetActivities)
				spGroup.POST("/activities", sp.CreateActivity)
				spGroup.PUT("/activities/:id", sp.UpdateActivity)
				spGroup.DELETE("/activities/:id", sp.DeleteActivity)

				// 服务号配置
				spGroup.GET("/wechat-config", sp.GetWechatConfig)
				spGroup.PUT("/wechat-config", sp.UpdateWechatConfig)
			}
		}

		// 微信支付回调
		notifyGroup := v1.Group("/notify")
		{
			notifyGroup.POST("/payment", admin.PaymentNotify)
		}

		// 微信支付回调（PRD口径）
		callbackGroup := v1.Group("/callback")
		{
			callbackGroup.POST("/wechat", admin.PaymentNotify)
		}
	}
}
