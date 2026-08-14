package main

import (
	"log"
	"os"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/handlers/callbacks"
	"fz_yyc_api/internal/handlers/merchant"
	serviceStaff "fz_yyc_api/internal/handlers/service_staff"
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
		wsGroup.Use(middleware.JWTAuth(), middleware.MerchantAuth())
		{
			wsGroup.GET("/merchant", wsHandler.MerchantWS)
		}

		devGroup := v1.Group("/dev")
		{
			devGroup.POST("/order-notify", wsHandler.DevOrderNotify)
			devGroup.POST("/store-visit-notify", wsHandler.DevStoreVisitNotify)
		}

		// 微信支付回调（不需要鉴权）
		callbackGroup := v1.Group("/callback/wechatpay")
		{
			callbackGroup.POST("/pay", callbacks.WechatPayPayCallback)
			callbackGroup.POST("/refund", callbacks.WechatPayRefundCallback)
		}

		// 认证相关
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/merchant/login", merchant.Login)
			authGroup.POST("/merchant/wechat-login", merchant.WechatQuickLogin)
			authGroup.POST("/user/wechat-login", user.WechatLogin)
		}

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

		// C端店铺接口
		storeGroup := v1.Group("/store/:merchant_id")
		storeGroup.Use(middleware.OptionalJWTAuth())
		{
			storeGroup.GET("/home", user.GetStoreHome)
			storeGroup.GET("/products", user.GetProducts)
			storeGroup.GET("/products/:product_id", user.GetProductDetail)
			storeGroup.POST("/visit", user.RecordUserVisit)
			storeGroup.POST("/event", user.RecordBehaviorEvent)

			storeAuthedGroup := storeGroup.Group("")
			storeAuthedGroup.Use(middleware.JWTAuth(), middleware.UserAuth())
			{
				storeAuthedGroup.POST("/orders", user.CreateOrder)
			}
		}

		// C端用户接口（需要登录）
		userGroup := v1.Group("/user")
		userGroup.Use(middleware.JWTAuth(), middleware.UserAuth())
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
			merchantOnlyGroup := merchantGroup.Group("")
			merchantOnlyGroup.Use(middleware.MerchantAuth())
			{
				// 商家信息
				merchantGroup.GET("/profile", merchant.GetProfile)
				merchantGroup.PUT("/profile", merchant.UpdateProfile)
				merchantGroup.GET("/settings", merchant.GetSettings)
				merchantGroup.PUT("/settings", merchant.UpdateSettings)
				merchantGroup.POST("/account/change-password", merchant.ChangePassword)
				merchantGroup.POST("/account/wechat/bind", merchant.BindWechat)
				merchantGroup.DELETE("/account/wechat/bind", merchant.UnbindWechat)
				merchantGroup.POST("/status", merchant.UpdateStatus)
				merchantGroup.GET("/qrcode", merchant.GetQRCode)

				// 支付配置
				merchantOnlyGroup.PUT("/payment-config", merchant.UpdatePaymentConfig)

				// 员工管理
				merchantOnlyGroup.GET("/staff", merchant.GetStaffList)
				merchantOnlyGroup.POST("/staff", merchant.CreateStaff)
				merchantOnlyGroup.PUT("/staff/:id", merchant.UpdateStaff)
				merchantOnlyGroup.DELETE("/staff/:id", merchant.DeleteStaff)
				merchantOnlyGroup.POST("/staff/:id/reset-password", merchant.ResetStaffPassword)

				// 服务人员管理（接单小程序账号审核与管理）
				merchantOnlyGroup.GET("/service-staff", merchant.GetServiceStaffList)
				merchantOnlyGroup.PUT("/service-staff/:id/status", merchant.UpdateServiceStaffStatus)
				merchantOnlyGroup.POST("/service-staff/:id/reset-password", merchant.ResetServiceStaffPassword)
				merchantOnlyGroup.DELETE("/service-staff/:id", merchant.DeleteServiceStaff)

				// 系统公告（商家查看）
				merchantOnlyGroup.GET("/announcements", merchant.GetAnnouncements)
				merchantOnlyGroup.GET("/announcements/:id", merchant.GetAnnouncementDetail)

				// 订单管理
				merchantOnlyGroup.GET("/orders", merchant.GetOrders)
				merchantOnlyGroup.GET("/orders/:order_id", merchant.GetOrderDetail)
				merchantOnlyGroup.POST("/orders/quick-complete", merchant.QuickCompleteOrder)
				merchantOnlyGroup.POST("/orders/:order_id/complete", merchant.CompleteOrder)
				merchantOnlyGroup.POST("/orders/:order_id/refund", merchant.RefundOrder)
				merchantOnlyGroup.POST("/orders/:order_id/return", merchant.ReturnRentalOrder)
				merchantOnlyGroup.GET("/orders/statistics", merchant.GetOrderStatistics)

				// 数据分析
				merchantOnlyGroup.GET("/analytics/overview", merchant.GetAnalyticsOverview)
				merchantOnlyGroup.GET("/analytics/sales-trend", merchant.GetSalesTrend)
				merchantOnlyGroup.GET("/analytics/product-ranking", merchant.GetProductRanking)
				merchantOnlyGroup.GET("/analytics/hourly", merchant.GetHourlyAnalysis)
				merchantOnlyGroup.GET("/analytics/stock-alert", merchant.GetStockAlert)
				merchantOnlyGroup.GET("/analytics/customers", merchant.GetCustomerAnalysis)
				merchantOnlyGroup.GET("/analytics/customer-trend", merchant.GetCustomerTrend)
			}

			merchantProductGroup := merchantGroup.Group("")
			merchantProductGroup.Use(middleware.MerchantAuth())
			{
				// 商品分类
				merchantProductGroup.GET("/categories", merchant.GetCategories)
				merchantProductGroup.POST("/categories", merchant.CreateCategory)
				merchantProductGroup.PUT("/categories/:category_id", merchant.UpdateCategory)
				merchantProductGroup.DELETE("/categories/:category_id", merchant.DeleteCategory)
				merchantProductGroup.POST("/categories/sort", merchant.SortCategories)

				// 商品管理
				merchantProductGroup.GET("/products", merchant.GetProducts)
				merchantProductGroup.GET("/products/:product_id", merchant.GetProduct)
				merchantProductGroup.POST("/products", merchant.CreateProduct)
				merchantProductGroup.PUT("/products/:product_id", merchant.UpdateProduct)
				merchantProductGroup.POST("/products/:product_id/on-sale", merchant.ProductOnSale)
				merchantProductGroup.POST("/products/:product_id/off-sale", merchant.ProductOffSale)
				merchantProductGroup.POST("/products/batch-status", merchant.BatchUpdateProductStatus)
				merchantProductGroup.DELETE("/products/:product_id", merchant.DeleteProduct)
				merchantProductGroup.PUT("/products/:product_id/stock", merchant.UpdateStock)
				merchantProductGroup.GET("/products/:product_id/specs", merchant.GetProductSpecs)
				merchantProductGroup.PUT("/products/:product_id/specs", merchant.UpdateProductSpecs)
			merchantProductGroup.DELETE("/products/:product_id/specs", merchant.DeleteProductSpecs)
		}

		// 服务人员接单小程序接口
		staffPublicGroup := v1.Group("/service-staff")
		{
			staffPublicGroup.POST("/register", serviceStaff.Register)
			staffPublicGroup.POST("/login", serviceStaff.Login)
			staffPublicGroup.POST("/wechat-login", serviceStaff.WechatLogin)
		}

		staffAuthedGroup := v1.Group("/service-staff")
		staffAuthedGroup.Use(middleware.JWTAuth(), middleware.ServiceStaffAuth())
		{
			staffAuthedGroup.GET("/profile", serviceStaff.GetProfile)
			staffAuthedGroup.GET("/todo", serviceStaff.GetTodoList)
			staffAuthedGroup.GET("/orders/pending", serviceStaff.PendingOrders)
			staffAuthedGroup.GET("/orders/accepted", serviceStaff.AcceptedOrders)
			staffAuthedGroup.GET("/orders/:id", serviceStaff.OrderDetail)
			staffAuthedGroup.POST("/orders/:id/accept", serviceStaff.AcceptOrder)
			staffAuthedGroup.POST("/orders/:id/check-in", serviceStaff.CheckIn)
			staffAuthedGroup.POST("/orders/:id/check-out", serviceStaff.CheckOut)
			staffAuthedGroup.GET("/statistics", serviceStaff.GetStatistics)
		}
		}
	}
}
