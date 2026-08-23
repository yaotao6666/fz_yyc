package main

import (
	"log"
	"os"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/handlers/callbacks"
	"fz_yyc_api/internal/handlers/health"
	"fz_yyc_api/internal/handlers/merchant"
	rbacHandler "fz_yyc_api/internal/handlers/rbac"
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
		storeGroup := v1.Group("/store")
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
			// 健康服务：居民健康档案 + 健康评估
			userGroup.GET("/health-record", health.UserGetHealthRecord)
			userGroup.PUT("/health-record", health.UserUpsertHealthRecord)
			userGroup.GET("/assessment-forms", health.UserGetAssessmentForms)
			userGroup.GET("/assessments", health.UserListAssessments)
			userGroup.POST("/assessments", health.UserCreateAssessment)
			// 康复辅具适配建议
			userGroup.GET("/fitting-recommendations", health.UserListFittingRecommendations)
			userGroup.GET("/fitting-recommendations/:id", health.UserGetFittingRecommendation)
			userGroup.POST("/fitting-recommendations/:id/confirm", health.UserConfirmFittingRecommendation)
			// 居家康养照护计划
			userGroup.GET("/care-plans", health.UserListCarePlans)
			userGroup.GET("/care-plans/:id", health.UserGetCarePlan)
			// 持续健康服务：随访任务 + 健康宣教 + 生命体征监测
			userGroup.GET("/follow-ups", health.UserListFollowUpTasks)
			userGroup.GET("/health-education", health.UserListEducationArticles)
			userGroup.GET("/health-education/:id", health.UserGetEducationArticle)
			userGroup.GET("/monitoring", health.UserListMonitoring)
			userGroup.POST("/monitoring", health.UserCreateMonitoring)
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
				merchantOnlyGroup.POST("/account/wechat/bind", middleware.RBAC("profile:update"), merchant.BindWechat)
				merchantOnlyGroup.DELETE("/account/wechat/bind", middleware.RBAC("profile:update"), merchant.UnbindWechat)
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

				// 服务人员审核中心（独立权限）
				merchantOnlyGroup.GET("/staff-audits", middleware.RBAC("staffaudit:view"), merchant.ListStaffAudits)
				merchantOnlyGroup.GET("/staff-audits/:id", middleware.RBAC("staffaudit:view"), merchant.StaffAuditDetail)
				merchantOnlyGroup.POST("/staff-audits/:id/approve", middleware.RBAC("staffaudit:approve"), merchant.ApproveStaffAudit)
				merchantOnlyGroup.POST("/staff-audits/:id/reject", middleware.RBAC("staffaudit:reject"), merchant.RejectStaffAudit)

				// 系统公告（商家查看）
				merchantOnlyGroup.GET("/announcements", middleware.RBAC("dashboard:view"), merchant.GetAnnouncements)
				merchantOnlyGroup.GET("/announcements/:id", middleware.RBAC("dashboard:view"), merchant.GetAnnouncementDetail)

				// 订单管理
				merchantOnlyGroup.GET("/orders", middleware.RBAC("orders:view"), merchant.GetOrders)
				merchantOnlyGroup.GET("/orders/dispatchable-staff", middleware.RBAC("order:dispatch"), merchant.GetDispatchableStaffList)
				merchantOnlyGroup.GET("/orders/rental-due", middleware.RBAC("orderrental:view"), merchant.ListRentalDueOrders)
				merchantOnlyGroup.GET("/orders/:order_id", middleware.RBAC("orders:view"), merchant.GetOrderDetail)
				merchantOnlyGroup.POST("/orders/quick-complete", middleware.RBAC("orders:complete"), merchant.QuickCompleteOrder)
				merchantOnlyGroup.POST("/orders/:order_id/complete", middleware.RBAC("orders:complete"), merchant.CompleteOrder)
				merchantOnlyGroup.POST("/orders/:order_id/refund", middleware.RBAC("orders:refund"), merchant.RefundOrder)
				merchantOnlyGroup.POST("/orders/:order_id/return", middleware.RBAC("orders:return"), merchant.ReturnRentalOrder)
				merchantOnlyGroup.POST("/orders/:order_id/dispatch", middleware.RBAC("order:dispatch"), merchant.DispatchOrder)
				merchantOnlyGroup.POST("/orders/:order_id/renew", middleware.RBAC("order:renew"), merchant.RenewOrder)
				merchantOnlyGroup.GET("/orders/statistics", middleware.RBAC("orders:view"), merchant.GetOrderStatistics)

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

				// 健康服务：居民健康档案 + 评估量表 + 评估记录
				merchantOnlyGroup.GET("/health-records", middleware.RBAC("health:view"), health.MerchantListHealthRecords)
				merchantOnlyGroup.GET("/health-records/:id", middleware.RBAC("health:view"), health.MerchantGetHealthRecord)
				merchantOnlyGroup.PUT("/health-records/:id", middleware.RBAC("health:update"), health.MerchantUpdateHealthRecord)
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
				// 居家康养照护计划与上门照护记录
				merchantOnlyGroup.GET("/care-plans", middleware.RBAC("care:view"), health.MerchantListCarePlans)
				merchantOnlyGroup.GET("/care-plans/:id", middleware.RBAC("care:view"), health.MerchantGetCarePlan)
				merchantOnlyGroup.POST("/care-plans", middleware.RBAC("care:create"), health.MerchantCreateCarePlan)
				merchantOnlyGroup.PUT("/care-plans/:id", middleware.RBAC("care:update"), health.MerchantUpdateCarePlan)
				merchantOnlyGroup.DELETE("/care-plans/:id", middleware.RBAC("care:delete"), health.MerchantDeleteCarePlan)
				merchantOnlyGroup.GET("/care-visits", middleware.RBAC("care:view"), health.MerchantListCareVisits)
				// 持续健康服务：随访任务 + 生命体征监测 + 健康宣教
				merchantOnlyGroup.GET("/follow-up-tasks", middleware.RBAC("followup:view"), health.MerchantListFollowUpTasks)
				merchantOnlyGroup.GET("/follow-up-tasks/:id", middleware.RBAC("followup:view"), health.MerchantGetFollowUpTask)
				merchantOnlyGroup.POST("/follow-up-tasks", middleware.RBAC("followup:update"), health.MerchantCreateFollowUpTask)
				merchantOnlyGroup.POST("/follow-up-tasks/:id/complete", middleware.RBAC("followup:update"), health.MerchantCompleteFollowUpTask)
				merchantOnlyGroup.GET("/monitoring", middleware.RBAC("monitor:view"), health.MerchantListMonitoring)
				merchantOnlyGroup.GET("/health-education", middleware.RBAC("education:view"), health.MerchantListEducationArticles)
				merchantOnlyGroup.POST("/health-education", middleware.RBAC("education:create"), health.MerchantCreateEducationArticle)
				merchantOnlyGroup.PUT("/health-education/:id", middleware.RBAC("education:create"), health.MerchantUpdateEducationArticle)
				merchantOnlyGroup.DELETE("/health-education/:id", middleware.RBAC("education:create"), health.MerchantDeleteEducationArticle)

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
			staffAuthedGroup.POST("/orders/:id/check-in", serviceStaff.CheckIn)
			staffAuthedGroup.POST("/orders/:id/check-out", serviceStaff.CheckOut)
			staffAuthedGroup.GET("/statistics", serviceStaff.GetStatistics)
			// 健康服务：客户档案与评估
			staffAuthedGroup.GET("/assessment-forms", health.StaffGetAssessmentForms)
			staffAuthedGroup.GET("/residents/:user_id/health-record", health.StaffGetResidentHealthRecord)
			staffAuthedGroup.GET("/residents/:user_id/assessments", health.StaffListResidentAssessments)
			staffAuthedGroup.POST("/residents/:user_id/assessments", health.StaffCreateAssessment)
			// 康复辅具适配建议
			staffAuthedGroup.GET("/residents/:user_id/fitting-recommendations", health.StaffListResidentFittingRecommendations)
			staffAuthedGroup.POST("/residents/:user_id/fitting-recommendations", health.StaffCreateFittingRecommendation)
			// 居家康养照护计划与上门照护记录
			staffAuthedGroup.GET("/care-plans", health.StaffListCarePlans)
			staffAuthedGroup.GET("/care-plans/:id", health.StaffGetCarePlan)
			staffAuthedGroup.POST("/care-visits", health.StaffCreateCareVisit)
			// 持续健康服务：随访任务 + 健康宣教 + 生命体征监测
			staffAuthedGroup.GET("/follow-up-tasks", health.StaffListFollowUpTasks)
			staffAuthedGroup.GET("/follow-up-tasks/:id", health.StaffGetFollowUpTask)
			staffAuthedGroup.POST("/follow-up-tasks/:id/complete", health.StaffCompleteFollowUpTask)
			staffAuthedGroup.POST("/follow-up-tasks/:id/skip", health.StaffSkipFollowUpTask)
			staffAuthedGroup.GET("/health-education", health.StaffListEducationArticles)
			staffAuthedGroup.GET("/residents/:user_id/monitoring", health.StaffListResidentMonitoring)
			staffAuthedGroup.POST("/residents/:user_id/monitoring", health.StaffCreateResidentMonitoring)
		}
	}
}
