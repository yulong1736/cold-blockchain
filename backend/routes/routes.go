package routes

import (
	"cold-chain-trace/backend/controllers"
	"cold-chain-trace/backend/middleware"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// 如果存在已构建的前端（Vue），则提供静态文件服务。
	// 构建输出目录预期为 ./frontend/dist
	distDir := filepath.Clean(filepath.Join(".", "frontend", "dist"))
	hasFrontend := false
	if stat, err := os.Stat(distDir); err == nil && stat.IsDir() {
		if _, err := os.Stat(filepath.Join(distDir, "index.html")); err == nil {
			hasFrontend = true
		}
	}

	// Vue Router 运行在 history 模式
	// 避免使用 gin.StaticFS("/") 因为会注册一个根路径通配符路由（/*filepath），会与 /api/* 路由产生冲突。
	if hasFrontend {
		r.NoRoute(func(c *gin.Context) {
			if c.Request.Method != http.MethodGet {
				c.Status(http.StatusNotFound)
				return
			}

			p := c.Request.URL.Path
			if strings.HasPrefix(p, "/api/") || p == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
				return
			}

			// 优先尝试从 dist 中提供真实的静态资源文件
			rel := strings.TrimPrefix(filepath.Clean(p), string(filepath.Separator))
			if rel == "." || rel == "" {
				rel = "index.html"
			}

			assetPath := filepath.Join(distDir, rel)
			if strings.HasPrefix(assetPath, distDir+string(filepath.Separator)) {
				if st, err := os.Stat(assetPath); err == nil && !st.IsDir() {
					c.File(assetPath)
					return
				}
			}

			// SPA 回退处理
			c.File(filepath.Join(distDir, "index.html"))
		})
	}

	// 初始化控制器
	authController := controllers.NewAuthController()
	productController := controllers.NewProductController()
	monitorController := controllers.NewMonitorController()
	dashboardController := controllers.NewDashboardController()
	blockchainController := controllers.NewBlockchainController()

	// 公开路由
	api := r.Group("/api")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/forgot-password/send-code", authController.SendPasswordResetCode)
			auth.POST("/forgot-password/reset", authController.ResetPasswordWithCode)
		}

		// 账号管理（需登录）
		api.PUT("/auth/profile", middleware.AuthMiddleware(), authController.UpdateProfile)
		api.PUT("/auth/password", middleware.AuthMiddleware(), authController.ChangePassword)

		// 商品查询（公开，消费者可访问，无需登录）
		products := api.Group("/products")
		{
			products.GET("/trace/:trace_id", productController.GetProductByTraceID)
			products.GET("/history/:trace_id", productController.GetProductHistory)
		}

		// 仪表盘数据（需登录）
		api.GET("/dashboard/stats", middleware.AuthMiddleware(), dashboardController.GetStats)
		api.GET("/dashboard/temperature-trend", middleware.AuthMiddleware(), dashboardController.GetTemperatureTrend)
		api.GET("/dashboard/transport-map", middleware.AuthMiddleware(), dashboardController.GetTransportMap)

		// 告警列表（仅仓储、物流，按角色分流）
		api.GET("/alerts", middleware.AuthMiddleware(), middleware.RoleMiddleware("warehouse", "logistics"), dashboardController.GetAlerts)

		// 监管审计日志（仅监管角色）
		api.GET("/regulator/audit-logs", middleware.AuthMiddleware(), middleware.RoleMiddleware("regulator"), productController.GetAuditLogs)
		// 监管查看全部告警
		api.GET("/regulator/alerts", middleware.AuthMiddleware(), middleware.RoleMiddleware("regulator"), dashboardController.GetAllAlerts)
	}

	// 需要认证的路由
	authenticated := api.Group("")
	authenticated.Use(middleware.AuthMiddleware())
	{
		// 链状态只读（生产商 / 仓储 / 物流轮询上链进度）
		blockchain := authenticated.Group("/blockchain")
		blockchain.Use(middleware.RoleMiddleware("producer", "warehouse", "logistics"))
		{
			blockchain.GET("/status", blockchainController.GetStatus)
		}

		products := authenticated.Group("/products")
		products.Use(middleware.RoleMiddleware("producer"))
		{
			products.GET("", productController.GetMyProducts)
			products.POST("", productController.CreateProduct)
			products.POST("/trace/:trace_id/sync-blockchain", productController.SyncProductBlockchainByTraceID)
			products.POST("/:id/sync-blockchain", productController.SyncProductBlockchain)
			products.PUT("/:id", productController.UpdateProduct)
			products.DELETE("/:id", productController.DeleteProduct)
		}

		// 温控监控（仓储角色）
		temperature := authenticated.Group("/temperature")
		temperature.Use(middleware.RoleMiddleware("warehouse"))
		{
			temperature.GET("", monitorController.ListTemperatureRecords)
			temperature.POST("", monitorController.CreateTemperatureRecord)
			temperature.POST("/:id/sync-blockchain", monitorController.SyncTemperatureBlockchain)
		}

		// 运输管理（物流角色）
		transport := authenticated.Group("/transport")
		transport.Use(middleware.RoleMiddleware("logistics"))
		{
			transport.GET("", monitorController.ListTransportNodes)
			transport.POST("", monitorController.CreateTransportNode)
			transport.GET("/my-products", monitorController.ListMyProducts)
			transport.POST("/:id/sync-blockchain", monitorController.SyncTransportBlockchain)
			transport.PUT("/:id", monitorController.UpdateTransportNodeTimes)
		}

		// 消费者：仅查询自己购买的商品
		consumer := authenticated.Group("/consumer")
		consumer.Use(middleware.RoleMiddleware("consumer"))
		{
			consumer.GET("/products", productController.GetMyPurchasedProducts)
			consumer.GET("/products/trace/:trace_id", productController.GetMyProductByTraceID)
			consumer.GET("/products/history/:trace_id", productController.GetMyProductHistory)
		}
	}

	return r
}
