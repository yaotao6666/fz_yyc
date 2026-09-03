package main

import (
	"log"
	"os"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/handlers/agreement"
	"fz_yyc_api/internal/handlers/callbacks"
	"fz_yyc_api/internal/handlers/health"
	"fz_yyc_api/internal/handlers/merchant"
	rbacHandler "fz_yyc_api/internal/handlers/rbac"
	serviceStaff "fz_yyc_api/internal/handlers/service_staff"
	"fz_yyc_api/internal/handlers/upload"
	"fz_yyc_api/internal/handlers/user"
	wsHandler "fz_yyc_api/internal/handlers/ws"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/tasks"
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

	// 启动优惠券定时任务（过期刷新 + 30天唤回发券）
	go tasks.StartCouponTasks()

	// 启动服务安全定时任务（超时预警 + 录音30天清理）
	go tasks.StartServiceSafetyTasks()

	// 启动订单/业务超时预警定时任务（启动即扫 + 5分钟粒度）
	go tasks.StartOrderTimeoutTasks()

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
					uploadGroup.POST("/sign", uploadHandler.Sign)
				}
		}

		// C端店铺接口
		storeGroup := v1.Group("/store")
		storeGroup.Use(middleware.OptionalJWTAuth())
		{
			storeGroup.GET("/home", user.GetStoreHome)
			storeGroup.GET("/home-recommends", user.GetStoreHomeRecommends)
			storeGroup.GET("/delivery-rules", user.GetStoreDeliveryRules)
			storeGroup.GET("/products", user.GetProducts)
			storeGroup.GET("/products/:product_id", user.GetProductDetail)
			storeGroup.GET("/coupons/available", user.GetAvailableCoupons)
			storeGroup.POST("/visit", user.RecordUserVisit)
			storeGroup.POST("/event", user.RecordBehaviorEvent)
			// 健康宣教独立板块（阶段五 8.3）：公开按分类浏览
			storeGroup.GET("/education/categories", health.StoreListEducationCategories)
			storeGroup.GET("/education/articles", health.StoreListEducationArticles)
			storeGroup.GET("/education/articles/:id", health.UserGetEducationArticle)

			storeAuthedGroup := storeGroup.Group("")
			storeAuthedGroup.Use(middleware.JWTAuth(), middleware.UserAuth())
			{
				storeAuthedGroup.POST("/orders", user.CreateOrder)
				storeAuthedGroup.POST("/coupons/:template_id/receive", user.ReceiveCoupon)
				storeAuthedGroup.GET("/coupons/usable", user.GetOrderUsableCoupons)
				storeAuthedGroup.GET("/my-coupons", user.GetMyCoupons)
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
			userGroup.POST("/orders/:order_id/renew", user.RenewOrder)
			// 服务评价（PRD V2.0 阶段四）
			userGroup.GET("/orders/:order_id/review", user.GetReview)
			userGroup.POST("/orders/:order_id/review", user.SubmitReview)
			userGroup.GET("/addresses", user.GetAddresses)
			userGroup.POST("/addresses", user.CreateAddress)
			userGroup.PUT("/addresses/:id", user.UpdateAddress)
			userGroup.DELETE("/addresses/:id", user.DeleteAddress)
			// 健康服务：居民健康档案 + 健康评估（阶段五 8.2 多档案：列表/新增/按ID更新/删除）
			userGroup.GET("/health-record", health.UserGetHealthRecord)
			userGroup.PUT("/health-record", health.UserUpsertHealthRecord)
			userGroup.GET("/health-records", health.UserListHealthRecords)
			userGroup.POST("/health-records", health.UserCreateHealthRecord)
			userGroup.PUT("/health-records/:id", health.UserUpdateHealthRecord)
			userGroup.DELETE("/health-records/:id", health.UserDeleteHealthRecord)
			userGroup.GET("/assessment-forms", health.UserGetAssessmentForms)
			userGroup.GET("/assessments", health.UserListAssessments)
			userGroup.POST("/assessments", health.UserCreateAssessment)
			// 康复辅具适配建议
			userGroup.GET("/fitting-recommendations", health.UserListFittingRecommendations)
			userGroup.GET("/fitting-recommendations/:id", health.UserGetFittingRecommendation)
			userGroup.POST("/fitting-recommendations/:id/confirm", health.UserConfirmFittingRecommendation)
			// 健康宣教
			userGroup.GET("/health-education", health.UserListEducationArticles)
			userGroup.GET("/health-education/:id", health.UserGetEducationArticle)
			// 协议（用户协议/隐私政策/录音定位授权，服务开始前确认）
			userGroup.GET("/agreements", agreement.GetActiveAgreement)
			userGroup.POST("/agreements/:id/consent", agreement.ConsentAgreement)
		}

		// 商家管理员接口（含 RBAC 权限控制）
		merchantGroup := v1.Group("/merchant")
		merchantGroup.Use(middleware.JWTAuth())
		{
			merchantOnlyGroup := merchantGroup.Group("")
			merchantOnlyGroup.Use(middleware.MerchantAuth())
			{
				// 商家信息
				merchantOnlyGroup.GET("/profile", middleware.RBAC("profile:view"), merchant.GetProfile)
				merchantOnlyGroup.PUT("/profile", middleware.RBAC("profile:update"), merchant.UpdateProfile)
				merchantOnlyGroup.GET("/settings", middleware.RBAC("profile:view"), merchant.GetSettings)
				merchantOnlyGroup.PUT("/settings", middleware.RBAC("profile:update"), merchant.UpdateSettings)
				merchantOnlyGroup.POST("/account/change-password", middleware.RBAC("profile:password"), merchant.ChangePassword)

				merchantOnlyGroup.POST("/status", middleware.RBAC("profile:update"), merchant.UpdateStatus)
				merchantOnlyGroup.GET("/qrcode", middleware.RBAC("profile:view"), merchant.GetQRCode)

				// 支付配置
				merchantOnlyGroup.PUT("/payment-config", middleware.RBAC("profile:payment"), merchant.UpdatePaymentConfig)

				// 员工管理（后台登录账号）
				merchantOnlyGroup.GET("/staff", middleware.RBAC("system:staff:view"), merchant.GetStaffList)
				merchantOnlyGroup.POST("/staff", middleware.RBAC("system:staff:create"), merchant.CreateStaff)
				merchantOnlyGroup.PUT("/staff/:id", middleware.RBAC("system:staff:update"), merchant.UpdateStaff)
				merchantOnlyGroup.DELETE("/staff/:id", middleware.RBAC("system:staff:delete"), merchant.DeleteStaff)
				merchantOnlyGroup.POST("/staff/:id/reset-password", middleware.RBAC("system:staff:reset-password"), merchant.ResetStaffPassword)

				// 服务人员管理（接单小程序账号审核与管理）
				merchantOnlyGroup.GET("/service-staff", middleware.RBAC("staff:view"), merchant.GetServiceStaffList)
				merchantOnlyGroup.POST("/service-staff", middleware.RBAC("staff:create"), merchant.CreateServiceStaff)
				merchantOnlyGroup.PUT("/service-staff/:id/status", middleware.RBAC("staff:update"), merchant.UpdateServiceStaffStatus)
				merchantOnlyGroup.POST("/service-staff/:id/reset-password", middleware.RBAC("staff:reset-password"), merchant.ResetServiceStaffPassword)
				merchantOnlyGroup.DELETE("/service-staff/:id", middleware.RBAC("staff:delete"), merchant.DeleteServiceStaff)
				merchantOnlyGroup.PUT("/service-staff/:id/service-region", middleware.RBAC("staff:update"), merchant.UpdateServiceStaffRegion)

				// 服务人员审核中心（独立权限）
				merchantOnlyGroup.GET("/staff-audits", middleware.RBAC("staffaudit:view"), merchant.ListStaffAudits)
				merchantOnlyGroup.GET("/staff-audits/:id", middleware.RBAC("staffaudit:view"), merchant.StaffAuditDetail)
				merchantOnlyGroup.POST("/staff-audits/:id/approve", middleware.RBAC("staffaudit:approve"), merchant.ApproveStaffAudit)
				merchantOnlyGroup.POST("/staff-audits/:id/reject", middleware.RBAC("staffaudit:reject"), merchant.RejectStaffAudit)

				// 打印机管理（PC 后台）
				merchantOnlyGroup.GET("/printers", middleware.RBAC("printers:view"), merchant.GetPrinters)
				merchantOnlyGroup.GET("/printers/:id", middleware.RBAC("printers:view"), merchant.GetPrinter)
				merchantOnlyGroup.POST("/printers", middleware.RBAC("printers:create"), merchant.CreatePrinter)
				merchantOnlyGroup.PUT("/printers/:id", middleware.RBAC("printers:update"), merchant.UpdatePrinter)
				merchantOnlyGroup.DELETE("/printers/:id", middleware.RBAC("printers:delete"), merchant.DeletePrinter)
				merchantOnlyGroup.POST("/printers/:id/test", middleware.RBAC("printers:update"), merchant.TestPrinter)

				// 订单管理
				merchantOnlyGroup.GET("/orders", middleware.RBACAny("orders:view", "order:goods", "order:service"), merchant.GetOrders)
				merchantOnlyGroup.GET("/orders/dispatchable-staff", middleware.RBAC("order:dispatch"), merchant.GetDispatchableStaffList)
				merchantOnlyGroup.GET("/orders/rental-due", middleware.RBAC("orderrental:view"), merchant.ListRentalDueOrders)
				merchantOnlyGroup.GET("/orders/:order_id", middleware.RBACAny("orders:view", "order:goods", "order:service"), merchant.GetOrderDetail)
				merchantOnlyGroup.POST("/orders/quick-complete", middleware.RBAC("orders:complete"), merchant.QuickCompleteOrder)
				merchantOnlyGroup.POST("/orders/:order_id/complete", middleware.RBAC("orders:complete"), merchant.CompleteOrder)
				merchantOnlyGroup.POST("/orders/:order_id/refund", middleware.RBAC("orders:refund"), merchant.RefundOrder)
				merchantOnlyGroup.POST("/orders/:order_id/return", middleware.RBAC("orders:return"), merchant.ReturnRentalOrder)
				merchantOnlyGroup.POST("/orders/:order_id/dispatch", middleware.RBAC("order:dispatch"), merchant.DispatchOrder)
				merchantOnlyGroup.POST("/orders/:order_id/renew", middleware.RBAC("order:renew"), merchant.RenewOrder)
				merchantOnlyGroup.GET("/orders/statistics", middleware.RBAC("orders:view"), merchant.GetOrderStatistics)
				merchantOnlyGroup.GET("/orders/:order_id/service-record", middleware.RBAC("orders:view"), merchant.GetOrderServiceRecord)

				// 预警中心（阶段三：服务过程安全）
				merchantOnlyGroup.GET("/alert-events", middleware.RBAC("alert-events:view"), merchant.GetAlertEvents)
				merchantOnlyGroup.GET("/alert-events/:id", middleware.RBAC("alert-events:view"), merchant.GetAlertEventDetail)
				merchantOnlyGroup.POST("/alert-events/:id/handle", middleware.RBAC("alert-events:update"), merchant.HandleAlertEvent)

				// 预警设置（阶段五：订单/业务超时预警阈值配置）
				merchantOnlyGroup.GET("/alert-settings", middleware.RBAC("alert-events:view"), merchant.GetAlertSettings)
				merchantOnlyGroup.PUT("/alert-settings", middleware.RBAC("alert-settings:update"), merchant.UpdateAlertSettings)

				// 协议管理（阶段三：服务过程安全）
				merchantOnlyGroup.GET("/agreements", middleware.RBAC("agreements:view"), merchant.GetAgreements)
				merchantOnlyGroup.GET("/agreements/:id", middleware.RBAC("agreements:view"), merchant.GetAgreementDetail)
				merchantOnlyGroup.POST("/agreements", middleware.RBAC("agreements:create"), merchant.CreateAgreement)
				merchantOnlyGroup.PUT("/agreements/:id", middleware.RBAC("agreements:update"), merchant.UpdateAgreement)
				merchantOnlyGroup.POST("/agreements/:id/publish", middleware.RBAC("agreements:update"), merchant.PublishAgreement)

				// 服务评价（阶段四：评价与服务质量分）
				merchantOnlyGroup.GET("/service-reviews", middleware.RBAC("service-reviews:view"), merchant.GetServiceReviews)
				merchantOnlyGroup.POST("/service-reviews/:id/hide", middleware.RBAC("service-reviews:update"), merchant.HideServiceReview)

				// 数据分析
				merchantOnlyGroup.GET("/analytics/overview", middleware.RBAC("analytics:view"), merchant.GetAnalyticsOverview)
				merchantOnlyGroup.GET("/analytics/sales-trend", middleware.RBAC("analytics:view"), merchant.GetSalesTrend)
				merchantOnlyGroup.GET("/analytics/product-ranking", middleware.RBAC("analytics:view"), merchant.GetProductRanking)
				merchantOnlyGroup.GET("/analytics/hourly", middleware.RBAC("analytics:view"), merchant.GetHourlyAnalysis)
				merchantOnlyGroup.GET("/analytics/stock-alert", middleware.RBAC("analytics:view"), merchant.GetStockAlert)
				merchantOnlyGroup.GET("/analytics/customers", middleware.RBAC("analytics:view"), merchant.GetCustomerAnalysis)
				merchantOnlyGroup.GET("/analytics/customer-trend", middleware.RBAC("analytics:view"), merchant.GetCustomerTrend)

				// 商品分类
				merchantOnlyGroup.GET("/categories", middleware.RBAC("categories:view"), merchant.GetCategories)
				merchantOnlyGroup.POST("/categories", middleware.RBAC("categories:create"), merchant.CreateCategory)
				merchantOnlyGroup.PUT("/categories/:category_id", middleware.RBAC("categories:update"), merchant.UpdateCategory)
				merchantOnlyGroup.DELETE("/categories/:category_id", middleware.RBAC("categories:delete"), merchant.DeleteCategory)
				merchantOnlyGroup.POST("/categories/sort", middleware.RBAC("categories:sort"), merchant.SortCategories)

				// 商品管理
				merchantOnlyGroup.GET("/products", middleware.RBAC("products:view"), merchant.GetProducts)
				merchantOnlyGroup.GET("/products/:product_id", middleware.RBAC("products:view"), merchant.GetProduct)
				merchantOnlyGroup.POST("/products", middleware.RBAC("products:create"), merchant.CreateProduct)
				merchantOnlyGroup.PUT("/products/:product_id", middleware.RBAC("products:update"), merchant.UpdateProduct)
				merchantOnlyGroup.POST("/products/:product_id/on-sale", middleware.RBAC("products:status"), merchant.ProductOnSale)
				merchantOnlyGroup.POST("/products/:product_id/off-sale", middleware.RBAC("products:status"), merchant.ProductOffSale)
				merchantOnlyGroup.POST("/products/batch-status", middleware.RBAC("products:status"), merchant.BatchUpdateProductStatus)
				merchantOnlyGroup.DELETE("/products/:product_id", middleware.RBAC("products:delete"), merchant.DeleteProduct)
				merchantOnlyGroup.PUT("/products/:product_id/stock", middleware.RBAC("products:stock"), merchant.UpdateStock)
				merchantOnlyGroup.GET("/products/:product_id/specs", middleware.RBAC("products:specs"), merchant.GetProductSpecs)
				merchantOnlyGroup.PUT("/products/:product_id/specs", middleware.RBAC("products:specs"), merchant.UpdateProductSpecs)
				merchantOnlyGroup.DELETE("/products/:product_id/specs", middleware.RBAC("products:specs"), merchant.DeleteProductSpecs)

				// 优惠券管理（PRD V2.0 阶段二）
				merchantOnlyGroup.GET("/coupon-templates", middleware.RBAC("coupon-templates:view"), merchant.GetCouponTemplates)
				merchantOnlyGroup.GET("/coupon-templates/:id", middleware.RBAC("coupon-templates:view"), merchant.GetCouponTemplate)
				merchantOnlyGroup.POST("/coupon-templates", middleware.RBAC("coupon-templates:create"), merchant.CreateCouponTemplate)
				merchantOnlyGroup.PUT("/coupon-templates/:id", middleware.RBAC("coupon-templates:update"), merchant.UpdateCouponTemplate)
				merchantOnlyGroup.POST("/coupon-templates/:id/status", middleware.RBAC("coupon-templates:update"), merchant.UpdateCouponTemplateStatus)
				merchantOnlyGroup.DELETE("/coupon-templates/:id", middleware.RBAC("coupon-templates:delete"), merchant.DeleteCouponTemplate)
				merchantOnlyGroup.POST("/coupon-templates/:id/grant", middleware.RBAC("coupon-templates:create"), merchant.GrantCoupon)
				merchantOnlyGroup.GET("/user-coupons", middleware.RBAC("user-coupons:view"), merchant.GetUserCoupons)

				// 小程序轮播图配置
				merchantOnlyGroup.GET("/miniprogram-banners", middleware.RBAC("banners:view"), merchant.GetBanners)
				merchantOnlyGroup.GET("/miniprogram-banners/:id", middleware.RBAC("banners:view"), merchant.GetBanner)
				merchantOnlyGroup.POST("/miniprogram-banners", middleware.RBAC("banners:create"), merchant.CreateBanner)
				merchantOnlyGroup.PUT("/miniprogram-banners/:id", middleware.RBAC("banners:update"), merchant.UpdateBanner)
				merchantOnlyGroup.PATCH("/miniprogram-banners/:id/status", middleware.RBAC("banners:status"), merchant.UpdateBannerStatus)
				merchantOnlyGroup.DELETE("/miniprogram-banners/:id", middleware.RBAC("banners:delete"), merchant.DeleteBanner)

				// 小程序首页推荐配置
				merchantOnlyGroup.GET("/home-recommends", middleware.RBAC("home-recommend:view"), merchant.GetHomeRecommends)
				merchantOnlyGroup.POST("/home-recommends", middleware.RBAC("home-recommend:create"), merchant.CreateHomeRecommend)
				merchantOnlyGroup.PUT("/home-recommends/:id", middleware.RBAC("home-recommend:update"), merchant.UpdateHomeRecommend)
				merchantOnlyGroup.PATCH("/home-recommends/:id/status", middleware.RBAC("home-recommend:status"), merchant.UpdateHomeRecommendStatus)
				merchantOnlyGroup.DELETE("/home-recommends/:id", middleware.RBAC("home-recommend:delete"), merchant.DeleteHomeRecommend)

				// RBAC 系统管理
				merchantOnlyGroup.GET("/rbac/permissions", rbacHandler.GetMyMenus)
				merchantOnlyGroup.GET("/rbac/menus", middleware.RBAC("system:menu:view"), rbacHandler.GetMenuTree)
				merchantOnlyGroup.POST("/rbac/menus", middleware.RBAC("system:menu:create"), rbacHandler.CreateMenu)
				merchantOnlyGroup.PUT("/rbac/menus/:id", middleware.RBAC("system:menu:create"), rbacHandler.UpdateMenu)
				merchantOnlyGroup.DELETE("/rbac/menus/:id", middleware.RBAC("system:menu:delete"), rbacHandler.DeleteMenu)
				merchantOnlyGroup.GET("/rbac/roles", middleware.RBAC("system:role:view"), rbacHandler.GetRoleList)
				merchantOnlyGroup.GET("/rbac/roles/all", middleware.RBACAny("system:role:view", "system:staff:view"), rbacHandler.GetAllRoles)
				merchantOnlyGroup.POST("/rbac/roles", middleware.RBAC("system:role:create"), rbacHandler.CreateRole)
				merchantOnlyGroup.PUT("/rbac/roles/:id", middleware.RBAC("system:role:create"), rbacHandler.UpdateRole)
				merchantOnlyGroup.DELETE("/rbac/roles/:id", middleware.RBAC("system:role:delete"), rbacHandler.DeleteRole)
				merchantOnlyGroup.GET("/rbac/roles/:id/menus", middleware.RBAC("system:role:view"), rbacHandler.GetRoleMenus)
				merchantOnlyGroup.PUT("/rbac/roles/:id/menus", middleware.RBAC("system:role:assign"), rbacHandler.AssignRoleMenus)
				merchantOnlyGroup.GET("/rbac/departments", middleware.RBACAny("system:dept:view", "system:staff:view"), rbacHandler.GetDepartmentTree)
				merchantOnlyGroup.POST("/rbac/departments", middleware.RBAC("system:dept:create"), rbacHandler.CreateDepartment)
				merchantOnlyGroup.PUT("/rbac/departments/:id", middleware.RBAC("system:dept:create"), rbacHandler.UpdateDepartment)
				merchantOnlyGroup.DELETE("/rbac/departments/:id", middleware.RBAC("system:dept:delete"), rbacHandler.DeleteDepartment)

				// 通用系统配置（system_configs key-value + JSON + 备注）
				merchantOnlyGroup.GET("/system-configs", middleware.RBAC("systemconfig:view"), merchant.ListSystemConfigs)
				merchantOnlyGroup.PUT("/system-configs", middleware.RBAC("systemconfig:update"), merchant.UpdateSystemConfig)
				merchantOnlyGroup.DELETE("/system-configs/:key", middleware.RBAC("systemconfig:update"), merchant.DeleteSystemConfig)

				// 健康服务：居民健康档案 + 评估量表 + 评估记录
				merchantOnlyGroup.GET("/health-records", middleware.RBAC("health:view"), health.MerchantListHealthRecords)
				merchantOnlyGroup.GET("/health-records/:id", middleware.RBAC("health:view"), health.MerchantGetHealthRecord)
				merchantOnlyGroup.PUT("/health-records/:id", middleware.RBAC("health:update"), health.MerchantUpdateHealthRecord)
				merchantOnlyGroup.GET("/health-records/:id/assessments", middleware.RBAC("health:assessment"), health.MerchantListRecordAssessments)
				merchantOnlyGroup.GET("/assessment-forms", middleware.RBAC("assessment:view"), health.MerchantListAssessmentForms)
				merchantOnlyGroup.POST("/assessment-forms", middleware.RBAC("assessment:create"), health.MerchantCreateAssessmentForm)
				merchantOnlyGroup.PUT("/assessment-forms/:id", middleware.RBAC("assessment:update"), health.MerchantUpdateAssessmentForm)
				merchantOnlyGroup.PATCH("/assessment-forms/:id/status", middleware.RBAC("assessment:update"), health.MerchantSetAssessmentFormStatus)
				merchantOnlyGroup.DELETE("/assessment-forms/:id", middleware.RBAC("assessment:delete"), health.MerchantDeleteAssessmentForm)
				merchantOnlyGroup.GET("/health-assessments", middleware.RBAC("assessment:view"), health.MerchantListHealthAssessments)
				// 康复辅具适配建议
				merchantOnlyGroup.GET("/fitting-recommendations", middleware.RBAC("fitting:view"), health.MerchantListFittingRecommendations)
				merchantOnlyGroup.GET("/fitting-recommendations/:id", middleware.RBAC("fitting:view"), health.MerchantGetFittingRecommendation)
				merchantOnlyGroup.PUT("/fitting-recommendations/:id", middleware.RBAC("fitting:update"), health.MerchantUpdateFittingRecommendation)
				merchantOnlyGroup.DELETE("/fitting-recommendations/:id", middleware.RBAC("fitting:update"), health.MerchantDeleteFittingRecommendation)
				// 健康宣教（独立板块：内容 + 两级分类）
				merchantOnlyGroup.GET("/health-education", middleware.RBAC("education:view"), health.MerchantListEducationArticles)
				merchantOnlyGroup.POST("/health-education", middleware.RBAC("education:create"), health.MerchantCreateEducationArticle)
				merchantOnlyGroup.PUT("/health-education/:id", middleware.RBAC("education:update"), health.MerchantUpdateEducationArticle)
				merchantOnlyGroup.DELETE("/health-education/:id", middleware.RBAC("education:delete"), health.MerchantDeleteEducationArticle)
				merchantOnlyGroup.GET("/education-categories", middleware.RBAC("education-categories:view"), health.MerchantListEducationCategories)
				merchantOnlyGroup.POST("/education-categories", middleware.RBAC("education-categories:create"), health.MerchantCreateEducationCategory)
				merchantOnlyGroup.PUT("/education-categories/:id", middleware.RBAC("education-categories:update"), health.MerchantUpdateEducationCategory)
				merchantOnlyGroup.DELETE("/education-categories/:id", middleware.RBAC("education-categories:delete"), health.MerchantDeleteEducationCategory)

				// 服务商分账
				merchantOnlyGroup.GET("/profit-sharing/receivers", middleware.RBAC("profit:view"), merchant.GetProfitSharingReceivers)
				merchantOnlyGroup.POST("/profit-sharing/receivers", middleware.RBAC("profit:receiver:manage"), merchant.CreateProfitSharingReceiver)
				merchantOnlyGroup.PUT("/profit-sharing/receivers/:id", middleware.RBAC("profit:receiver:manage"), merchant.UpdateProfitSharingReceiver)
				merchantOnlyGroup.DELETE("/profit-sharing/receivers/:id", middleware.RBAC("profit:receiver:manage"), merchant.DeleteProfitSharingReceiver)
				merchantOnlyGroup.POST("/profit-sharing/receivers/:id/sync", middleware.RBAC("profit:receiver:manage"), merchant.SyncProfitSharingReceiver)
				merchantOnlyGroup.GET("/profit-sharing/config", middleware.RBAC("profit:view"), merchant.GetProfitSharingConfig)
				merchantOnlyGroup.PUT("/profit-sharing/config", middleware.RBAC("profit:config"), merchant.UpdateProfitSharingConfig)
				merchantOnlyGroup.GET("/profit-sharing/records", middleware.RBAC("profit:view"), merchant.GetProfitSharingRecords)
				merchantOnlyGroup.GET("/profit-sharing/records/:id", middleware.RBAC("profit:view"), merchant.GetProfitSharingRecordDetail)
				merchantOnlyGroup.POST("/profit-sharing/records/:id/retry", middleware.RBAC("profit:share"), merchant.RetryProfitSharingRecord)
			}
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
			staffAuthedGroup.PUT("/profile", serviceStaff.RequestProfileChange)
			staffAuthedGroup.GET("/audits", serviceStaff.GetMyAuditList)
			staffAuthedGroup.GET("/todo", serviceStaff.GetTodoList)
			staffAuthedGroup.GET("/orders/pending", serviceStaff.PendingOrders)
			staffAuthedGroup.GET("/orders/accepted", serviceStaff.AcceptedOrders)
			staffAuthedGroup.GET("/orders/:id", serviceStaff.OrderDetail)
			staffAuthedGroup.POST("/orders/:id/accept", serviceStaff.AcceptOrder)
			staffAuthedGroup.POST("/orders/:id/give-up", serviceStaff.GiveUpOrder)
			staffAuthedGroup.POST("/orders/:id/check-in", serviceStaff.CheckIn)
			staffAuthedGroup.POST("/orders/:id/check-out", serviceStaff.CheckOut)
			// 服务过程安全（阶段三）：定位上报 / 录音提交 / SOS / 服务区域
			staffAuthedGroup.POST("/orders/:id/location", serviceStaff.ReportLocation)
			staffAuthedGroup.POST("/orders/:id/audio", serviceStaff.SubmitAudio)
			staffAuthedGroup.POST("/sos", serviceStaff.SOS)
			staffAuthedGroup.GET("/my-region", serviceStaff.GetMyRegion)
			// 协议（录音/定位授权等，服务前确认）
			staffAuthedGroup.GET("/agreements", agreement.GetActiveAgreement)
			staffAuthedGroup.POST("/agreements/:id/consent", agreement.ConsentAgreement)
			// 服务评价（阶段四）：我的质量分与近期评价
			staffAuthedGroup.GET("/my-quality-score", serviceStaff.GetMyQualityScore)
			staffAuthedGroup.GET("/statistics", serviceStaff.GetStatistics)
			// 健康服务：客户档案与评估（阶段五 8.2 多档案：列表 + 按档案ID更新健康数值）
			staffAuthedGroup.GET("/assessment-forms", health.StaffGetAssessmentForms)
			staffAuthedGroup.GET("/residents/:user_id/health-record", health.StaffGetResidentHealthRecord)
			staffAuthedGroup.GET("/residents/:user_id/health-records", health.StaffListResidentHealthRecords)
			staffAuthedGroup.PUT("/residents/:user_id/health-records/:record_id/values", health.StaffUpdateResidentHealthValues)
			staffAuthedGroup.GET("/residents/:user_id/assessments", health.StaffListResidentAssessments)
			staffAuthedGroup.POST("/residents/:user_id/assessments", health.StaffCreateAssessment)
			// 康复辅具适配建议
			staffAuthedGroup.GET("/residents/:user_id/fitting-recommendations", health.StaffListResidentFittingRecommendations)
			staffAuthedGroup.POST("/residents/:user_id/fitting-recommendations", health.StaffCreateFittingRecommendation)
			// 健康宣教
			staffAuthedGroup.GET("/health-education", health.StaffListEducationArticles)
		}
	}
}
