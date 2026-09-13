package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"cold-chain-trace/backend/blockchain"
	"cold-chain-trace/backend/config"
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/routes"
	"cold-chain-trace/backend/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置（从项目根目录运行时，配置文件路径为 backend/config/config.yaml）
	// config.yaml 因含敏感信息被 .gitignore 忽略；全新 clone 后不存在时回退到 config.example.yaml 模板
	configPath := "backend/config/config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "backend/config/config.example.yaml"
		log.Printf("未找到 %s，回退使用 %s（请复制并填写真实配置）", "backend/config/config.yaml", configPath)
	}
	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// 初始化区块链客户端
	if err := blockchain.InitFISCOClient(); err != nil {
		log.Printf("Warning: Failed to initialize blockchain client: %v", err)
		log.Println("Application will continue without blockchain features")
	}

	// 设置Gin模式
	if config.AppConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 设置路由
	r := routes.SetupRoutes()

	// 链恢复后自动补录待上链商品
	services.StartPendingProductBlockchainReconcile(12 * time.Second)

	// 启动服务器
	port := config.AppConfig.Server.Port
	log.Printf("Server starting on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
