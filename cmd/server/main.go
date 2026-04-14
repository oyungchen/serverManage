package main

import (
	"fmt"
	"log"

	"serverManage/internal/config"
	"serverManage/internal/handler"
	"serverManage/internal/launcher"
	"serverManage/internal/logger"
	"serverManage/internal/service"
	"serverManage/internal/storage"
	"serverManage/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化存储
	store, err := storage.New(cfg.Storage.DataFile)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	// 初始化日志
	logMgr, err := logger.New(cfg.Storage.LogDir)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	// 初始化服务管理器
	manager := service.NewManager(store)

	// 初始化 launchd 管理器
	launchMgr := launcher.New()

	// 初始化 WebSocket hub
	wsHub := websocket.NewHub(logMgr)
	go wsHub.Run()

	// 初始化 HTTP 处理器
	h := handler.New(manager, launchMgr, logMgr, wsHub, cfg.Web.Static)

	// 创建 Gin 引擎
	r := gin.Default()

	// 注册路由
	h.RegisterRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Web UI: http://localhost:%d", cfg.Server.Port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}