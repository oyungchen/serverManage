package handler

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"serverManage/internal/launcher"
	"serverManage/internal/logger"
	"serverManage/internal/model"
	"serverManage/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	manager *service.Manager
	launch  *launcher.Launcher
	log     *logger.Logger
	wsHub   interface {
		HandleWS(w http.ResponseWriter, r *http.Request, serviceName string)
	}
	staticDir string
}

func New(manager *service.Manager, launch *launcher.Launcher, log *logger.Logger, wsHub interface {
	HandleWS(w http.ResponseWriter, r *http.Request, serviceName string)
}, staticDir string) *Handler {
	return &Handler{
		manager:   manager,
		launch:    launch,
		log:       log,
		wsHub:     wsHub,
		staticDir: staticDir,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/services", h.GetServices)
		api.POST("/services", h.CreateService)
		api.PUT("/services/:id", h.UpdateService)
		api.DELETE("/services/:id", h.DeleteService)
		api.POST("/services/:id/start", h.StartService)
		api.POST("/services/:id/stop", h.StopService)
		api.POST("/services/:id/restart", h.RestartService)
		api.GET("/services/:id/logs", h.GetLogs)
		api.POST("/discover", h.Discover)
	}

	r.GET("/ws/logs", h.HandleWebSocket)
	r.GET("/ws/logs/:serviceId", h.HandleWebSocket)

	// 静态文件
	r.Static("/static", h.staticDir)
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(h.staticDir, "index.html"))
	})
}

func (h *Handler) GetServices(c *gin.Context) {
	services, err := h.manager.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services)
}

func (h *Handler) CreateService(c *gin.Context) {
	var req model.ServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc, err := h.manager.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果启用自启动
	if svc.AutoStart {
		h.launch.EnableAutoStart(svc.Name, svc.WorkDir, svc.StartScript)
	}

	c.JSON(http.StatusCreated, svc)
}

func (h *Handler) UpdateService(c *gin.Context) {
	id := c.Param("id")

	var req model.ServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc, err := h.manager.Update(id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 更新 launchd 配置
	h.launch.DisableAutoStart(svc.Name)
	if svc.AutoStart {
		h.launch.EnableAutoStart(svc.Name, svc.WorkDir, svc.StartScript)
	}

	c.JSON(http.StatusOK, svc)
}

func (h *Handler) DeleteService(c *gin.Context) {
	id := c.Param("id")

	svc, err := h.manager.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 停止服务
	h.manager.Stop(id)

	// 移除 launchd 配置
	h.launch.DisableAutoStart(svc.Name)

	if err := h.manager.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) StartService(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Start(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Write("service", "Service started", logger.LogLevelInfo)

	c.JSON(http.StatusOK, gin.H{"message": "started"})
}

func (h *Handler) StopService(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Stop(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Write("service", "Service stopped", logger.LogLevelInfo)

	c.JSON(http.StatusOK, gin.H{"message": "stopped"})
}

func (h *Handler) RestartService(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Restart(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Write("service", "Service restarted", logger.LogLevelInfo)

	c.JSON(http.StatusOK, gin.H{"message": "restarted"})
}

func (h *Handler) GetLogs(c *gin.Context) {
	id := c.Param("id")
	lines := c.DefaultQuery("lines", "100")

	lineCount, _ := strconv.Atoi(lines)

	svc, err := h.manager.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logs, err := h.log.Read(svc.Name, lineCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *Handler) Discover(c *gin.Context) {
	var req model.DiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现目录扫描自动发现服务
	// 扫描 package.json、requirements.txt 等文件

	c.JSON(http.StatusOK, model.DiscoverResult{Services: []model.ServiceRequest{}})
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	serviceId := c.Param("serviceId")
	h.wsHub.HandleWS(c.Writer, c.Request, serviceId)
}