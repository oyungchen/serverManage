# ServerManage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个 Go 后端 + Vue 3 前端的本机服务管理工具，支持 Web 界面管理多个服务的启动/停止/重启，实现开机自启动

**Architecture:** 后端使用 Gin 框架提供 REST API 和 WebSocket，前端使用 Vue 3 + Vite，数据存储在 JSON 文件，使用 macOS launchd 实现开机自启动

**Tech Stack:** Go + Gin + Vue 3 + Vite + Tailwind CSS + WebSocket

---

## 文件结构

```
serverManage/
├── cmd/
│   └── server/
│       └── main.go              # 入口，初始化并启动服务
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载，管理配置结构
│   ├── model/
│   │   └── service.go           # 数据模型，服务结构定义
│   ├── storage/
│   │   └── storage.go           # JSON 文件存储
│   ├── service/
│   │   └── manager.go           # 服务管理核心逻辑
│   ├── launcher/
│   │   └── launchd.go           # launchd 操作
│   ├── handler/
│   │   └── handler.go           # HTTP 处理函数
│   ├── logger/
│   │   └── logger.go            # 日志管理
│   └── websocket/
│       └── hub.go               # WebSocket  hub
├── web/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── src/
│   │   ├── main.js
│   │   ├── App.vue
│   │   ├── api/
│   │   │   └── index.js         # API 调用
│   │   ├── components/
│   │   │   ├── ServiceList.vue  # 服务列表
│   │   │   ├── ServiceItem.vue  # 单个服务项
│   │   │   └── ServiceModal.vue # 添加/编辑弹窗
│   │   └── views/
│   │       └── Home.vue         # 主页面
├── config.yaml                  # 配置文件
├── go.mod
└── go.sum
```

---

## Task 1: 项目初始化

**Files:**
- Create: `go.mod`
- Create: `config.yaml`

- [ ] **Step 1: 创建 go.mod**

```bash
cd /Users/didi/warhouse/serverManage
go mod init serverManage
```

- [ ] **Step 2: 创建 config.yaml**

```yaml
server:
  host: "0.0.0.0"
  port: 8081

web:
  static: "./web/dist"

storage:
  dataFile: "~/.serverManage/services.json"
  logDir: "~/.serverManage/logs"

discover:
  scanDirs:
    - "/Users/didi/projects"
    - "/Users/didi/go/src"
  excludeDirs:
    - node_modules
    - vendor
    - .git
```

- [ ] **Step 3: Commit**

```bash
git add go.mod config.yaml
git commit -m "chore: init project structure"
```

---

## Task 2: 数据模型定义

**Files:**
- Create: `internal/model/service.go`

- [ ] **Step 1: 创建数据模型**

```go
package model

import (
	"time"
)

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusError   ServiceStatus = "error"
)

type Service struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	StartScript   string        `json:"startScript"`
	StopScript    string        `json:"stopScript"`
	RestartScript string        `json:"restartScript"`
	WorkDir       string        `json:"workDir"`
	Port          int           `json:"port"`
	AutoStart     bool          `json:"autoStart"`
	Status        ServiceStatus `json:"status"`
	PID           int           `json:"pid,omitempty"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type ServiceList struct {
	Services []Service `json:"services"`
}

type ServiceRequest struct {
	Name          string `json:"name"`
	StartScript   string `json:"startScript"`
	StopScript    string `json:"stopScript"`
	RestartScript string `json:"restartScript"`
	WorkDir       string `json:"workDir"`
	Port          int    `json:"port"`
	AutoStart     bool   `json:"autoStart"`
}

type DiscoverRequest struct {
	Dirs []string `json:"dirs"`
}

type DiscoverResult struct {
	Services []ServiceRequest `json:"services"`
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/model/service.go
git commit -m "feat: add service data model"
```

---

## Task 3: 配置加载

**Files:**
- Create: `internal/config/config.go`

- [ ] **Step 1: 创建配置加载**

```go
package config

import (
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Web      WebConfig      `yaml:"web"`
	Storage  StorageConfig  `yaml:"storage"`
	Discover DiscoverConfig `yaml:"discover"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type WebConfig struct {
	Static string `yaml:"static"`
}

type StorageConfig struct {
	DataFile string `yaml:"dataFile"`
	LogDir   string `yaml:"logDir"`
}

type DiscoverConfig struct {
	ScanDirs   []string `yaml:"scanDirs"`
	ExcludeDirs []string `yaml:"excludeDirs"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 展开 ~ 为用户目录
	homeDir := getHomeDir()
	cfg.Storage.DataFile = expandHome(cfg.Storage.DataFile, homeDir)
	cfg.Storage.LogDir = expandHome(cfg.Storage.LogDir, homeDir)

	return &cfg, nil
}

func getHomeDir() string {
	user, _ := user.Current()
	if user.HomeDir != "" {
		return user.HomeDir
	}
	return os.Getenv("HOME")
}

func expandHome(path, home string) string {
	if len(path) > 0 && path[0] == '~' {
		return filepath.Join(home, path[1:])
	}
	return path
}
```

- [ ] **Step 2: 安装依赖**

```bash
go get gopkg.in/yaml.v3
```

- [ ] **Step 3: Commit**

```bash
git add internal/config/config.go go.mod go.sum
git commit -m "feat: add config loading"
```

---

## Task 4: JSON 存储

**Files:**
- Create: `internal/storage/storage.go`

- [ ] **Step 1: 创建存储逻辑**

```go
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"serverManage/internal/model"
)

type Storage struct {
	dataFile string
}

func New(dataFile string) (*Storage, error) {
	// 确保目录存在
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 如果文件不存在，创建空文件
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		empty := model.ServiceList{Services: []model.Service{}}
		data, _ := json.MarshalIndent(empty, "", "  ")
		os.WriteFile(dataFile, data, 0644)
	}

	return &Storage{dataFile: dataFile}, nil
}

func (s *Storage) Load() (*model.ServiceList, error) {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return nil, err
	}

	var list model.ServiceList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return &list, nil
}

func (s *Storage) Save(list *model.ServiceList) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/storage/storage.go
git commit -m "feat: add JSON storage"
```

---

## Task 5: 服务管理核心

**Files:**
- Create: `internal/service/manager.go`

- [ ] **Step 1: 创建服务管理器**

```go
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"serverManage/internal/model"
	"serverManage/internal/storage"

	"github.com/google/uuid"
)

type Manager struct {
	store   *storage.Storage
	process map[string]*exec.Cmd
	mu      sync.RWMutex
}

func NewManager(store *storage.Storage) *Manager {
	return &Manager{
		store:   store,
		process: make(map[string]*exec.Cmd),
	}
}

func (m *Manager) GetAll() ([]model.Service, error) {
	list, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	// 更新每个服务的实际状态
	for i := range list.Services {
		m.updateStatus(&list.Services[i])
	}

	return list.Services, nil
}

func (m *Manager) Get(id string) (*model.Service, error) {
	list, err :=.m.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range list.Services {
		if list.Services[i].ID == id {
			m.updateStatus(&list.Services[i])
			return &list.Services[i], nil
		}
	}

	return nil, fmt.Errorf("service not found: %s", id)
}

func (m *Manager) Create(req model.ServiceRequest) (*model.Service, error) {
	list, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	// 检查名称是否重复
	for _, s := range list.Services {
		if s.Name == req.Name {
			return nil, fmt.Errorf("service already exists: %s", req.Name)
		}
	}

	svc := model.Service{
		ID:            uuid.New().String(),
		Name:          req.Name,
		StartScript:   req.StartScript,
		StopScript:    req.StopScript,
		RestartScript: req.RestartScript,
		WorkDir:       req.WorkDir,
		Port:          req.Port,
		AutoStart:     req.AutoStart,
		Status:        model.StatusStopped,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	list.Services = append(list.Services, svc)
	if err := m.store.Save(list); err != nil {
		return nil, err
	}

	return &svc, nil
}

func (m *Manager) Update(id string, req model.ServiceRequest) (*model.Service, error) {
	list, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services[i].Name = req.Name
			list.Services[i].StartScript = req.StartScript
			list.Services[i].StopScript = req.StopScript
			list.Services[i].RestartScript = req.RestartScript
			list.Services[i].WorkDir = req.WorkDir
			list.Services[i].Port = req.Port
			list.Services[i].AutoStart = req.AutoStart
			list.Services[i].UpdatedAt = time.Now()

			if err := m.store.Save(list); err != nil {
				return nil, err
			}

			return &list.Services[i], nil
		}
	}

	return nil, fmt.Errorf("service not found: %s", id)
}

func (m *Manager) Delete(id string) error {
	list, err := m.store.Load()
	if err != nil {
		return err
	}

	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services = append(list.Services[:i], list.Services[i+1:]...)
			return m.store.Save(list)
		}
	}

	return fmt.Errorf("service not found: %s", id)
}

func (m *Manager) Start(id string) error {
	svc, err := m.Get(id)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在运行
	if m.isRunning(svc.Name) {
		return fmt.Errorf("service already running: %s", svc.Name)
	}

	cmd := exec.Command("bash", "-c", svc.StartScript)
	if svc.WorkDir != "" {
		cmd.Dir = svc.WorkDir
	}

	// 设置进程组
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	m.process[svc.Name] = cmd

	// 更新状态
	list, _ := m.store.Load()
	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services[i].Status = model.StatusRunning
			list.Services[i].PID = cmd.Process.Pid
			list.Services[i].UpdatedAt = time.Now()
			m.store.Save(list)
			break
		}
	}

	return nil
}

func (m *Manager) Stop(id string) error {
	svc, err := m.Get(id)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning(svc.Name) {
		return fmt.Errorf("service not running: %s", svc.Name)
	}

	// 执行停止脚本
	cmd := exec.Command("bash", "-c", svc.StopScript)
	if svc.WorkDir != "" {
		cmd.Dir = svc.WorkDir
	}
	cmd.Run()

	// 杀死进程组
	if proc, ok := m.process[svc.Name]; ok && proc.Process != nil {
		pgid, _ := syscall.Getpgid(proc.Process.Pid)
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	delete(m.process, svc.Name)

	// 更新状态
	list, _ := m.store.Load()
	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services[i].Status = model.StatusStopped
			list.Services[i].PID = 0
			list.Services[i].UpdatedAt = time.Now()
			m.store.Save(list)
			break
		}
	}

	return nil
}

func (m *Manager) Restart(id string) error {
	if err := m.Stop(id); err != nil {
		// 忽略停止错误，继续启动
	}
	return m.Start(id)
}

func (m *Manager) isRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if cmd, ok := m.process[name]; ok {
		if cmd.Process != nil {
			err := cmd.Process.Signal(syscall.Signal(0))
			return err == nil
		}
	}
	return false
}

func (m *Manager) updateStatus(svc *model.Service) {
	if m.isRunning(svc.Name) {
		svc.Status = model.StatusRunning
	} else {
		svc.Status = model.StatusStopped
	}
}
```

- [ ] **Step 2: 添加 syscall 导入**

在文件顶部添加：
```go
import (
	"syscall"
	// ... other imports
)
```

- [ ] **Step 3: 安装依赖**

```bash
go get github.com/google/uuid
```

- [ ] **Step 4: Commit**

```bash
git add internal/service/manager.go go.mod go.sum
git commit -m "feat: add service manager core"
```

---

## Task 6: launchd 集成

**Files:**
- Create: `internal/launcher/launchd.go`

- [ ] **Step 1: 创建 launchd 管理**

```go
package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const launchAgentsDir = "Library/LaunchAgents"

type Launcher struct {
	homeDir string
}

func New() *Launcher {
	home := os.Getenv("HOME")
	return &Launcher{homeDir: home}
}

func (l *Launcher) GetPlistPath(serviceName string) string {
	return filepath.Join(l.homeDir, launchAgentsDir, fmt.Sprintf("com.servermanage.%s.plist", serviceName))
}

func (l *Launcher) EnableAutoStart(serviceName, workDir, startScript string) error {
	// 确保目录存在
	dir := filepath.Join(l.homeDir, launchAgentsDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	plistPath := l.GetPlistPath(serviceName)

	tmpl := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.servermanage.{{.Name}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>-c</string>
        <string>cd {{.WorkDir}} && {{.StartScript}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
    <key>StandardOutPath</key>
    <string>{{.LogDir}}/{{.Name}}.out.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogDir}}/{{.Name}}.err.log</string>
</dict>
</plist>`

	data := struct {
		Name      string
		WorkDir   string
		StartScript string
		LogDir    string
	}{
		Name:        serviceName,
		WorkDir:     workDir,
		StartScript: startScript,
		LogDir:      filepath.Join(l.homeDir, ".serverManage", "logs"),
	}

	t, err := template.New("plist").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(plistPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return err
	}

	return nil
}

func (l *Launcher) DisableAutoStart(serviceName string) error {
	plistPath := l.GetPlistPath(serviceName)

	// 先 unload
	exec.Command("launchctl", " unload", plistPath).Run()

	// 删除文件
	if _, err := os.Stat(plistPath); err == nil {
		return os.Remove(plistPath)
	}

	return nil
}

func (l *Launcher) IsAutoStartEnabled(serviceName string) bool {
	_, err := os.Stat(l.GetPlistPath(serviceName))
	return err == nil
}

func (l *Launcher) GetAllManagedServices() ([]string, error) {
	dir := filepath.Join(l.homeDir, launchAgentsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var services []string
	prefix := "com.servermanage."
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".plist") {
			services = append(services, strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".plist"))
		}
	}

	return services, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/launcher/launchd.go
git commit -m "feat: add launchd integration"
```

---

## Task 7: 日志管理

**Files:**
- Create: `internal/logger/logger.go`

- [ ] **Step 1: 创建日志管理**

```go
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogEntry struct {
	ServiceName string    `json:"serviceName"`
	Level       LogLevel  `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

type Logger struct {
	logDir  string
	mu      sync.RWMutex
	writers map[string]*os.File
}

func New(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	return &Logger{
		logDir:  logDir,
		writers: make(map[string]*os.File),
	}, nil
}

func (l *Logger) GetLogFile(serviceName string) string {
	date := time.Now().Format("2006-01-02")
	return filepath.Join(l.logDir, fmt.Sprintf("%s_%s.log", serviceName, date))
}

func (l *Logger) Write(serviceName, message string, level LogLevel) {
	entry := LogEntry{
		ServiceName: serviceName,
		Level:       level,
		Message:     message,
		Timestamp:   time.Now(),
	}

	data, _ := json.Marshal(entry)
	line := string(data) + "\n"

	// 写入文件
	logFile := l.GetLogFile(serviceName)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(line)
		f.Close()
	}
}

func (l *Logger) Read(serviceName string, lines int) ([]string, error) {
	logFile := l.GetLogFile(serviceName)

	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, err
	}

	allLines := strings.Split(string(data), "\n")
	if lines > 0 && len(allLines) > lines {
		allLines = allLines[len(allLines)-lines:]
	}

	var result []string
	for _, line := range allLines {
		if line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

func (l *Logger) ReadSince(serviceName string, since time.Time) ([]string, error) {
	logFile := l.GetLogFile(serviceName)

	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, err
	}

	var result []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Timestamp.After(since) {
			result = append(result, line)
		}
	}

	return result, nil
}
```

- [ ] **Step 2: 添加 strings 导入**

在 import 中添加 `"strings"`

- [ ] **Step 3: Commit**

```bash
git add internal/logger/logger.go
git commit -m "feat: add logger"
```

---

## Task 8: WebSocket Hub

**Files:**
- Create: `internal/websocket/hub.go`

- [ ] **Step 1: 创建 WebSocket hub**

```go
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"serverManage/internal/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	id       string
	service  string
	socket   *websocket.Conn
	send     chan []byte
	hub      *Hub
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	log        *logger.Logger
	mu         sync.RWMutex
}

func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		log:        log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastLog(serviceName, message string, level logger.LogLevel) {
	entry := logger.LogEntry{
		ServiceName: serviceName,
		Level:       level,
		Message:     message,
		Timestamp:   time.Now(),
	}

	data, _ := json.Marshal(entry)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.service == "" || client.service == serviceName {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, serviceName string) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		id:      r.RemoteAddr,
		service: serviceName,
		socket:  socket,
		send:    make(chan []byte, 256),
		hub:     h,
	}

	h.register <- client

	go client.write()
	go client.read()
}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			break
		}

		// 可以处理客户端消息，如订阅/取消订阅
		var msg map[string]string
		if json.Unmarshal(message, &msg) == nil {
			if action, ok := msg["action"]; ok {
				if action == "subscribe" {
					c.service = msg["service"]
				}
			}
		}
	}
}

func (c *Client) write() {
	defer c.socket.Close()

	for {
		message, ok := <-c.send
		if !ok {
			c.socket.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.socket.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
```

- [ ] **Step 2: 安装依赖**

```bash
go get github.com/gorilla/websocket
```

- [ ] **Step 3: Commit**

```bash
git add internal/websocket/hub.go go.mod go.sum
git commit -m "feat: add WebSocket hub"
```

---

## Task 9: HTTP 处理器

**Files:**
- Create: `internal/handler/handler.go`

- [ ] **Step 1: 创建 HTTP 处理器**

```go
package handler

import (
	"net/http"
	"path/filepath"
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

	svc, err := h.manager.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var lineCount int
	strings.Scan(lines, &lineCount)

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
```

- [ ] **Step 2: Commit**

```bash
git add internal/handler/handler.go
git commit -m "feat: add HTTP handler"
```

---

## Task 10: 主入口

**Files:**
- Create: `cmd/server/main.go`

- [ ] **Step 1: 创建主入口**

```go
package main

import (
	"fmt"
	"log"
	"os"

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
```

- [ ] **Step 2: 修复编译错误**

需要在 internal/handler/handler.go 中添加 strings 包的导入

- [ ] **Step 3: 编译测试**

```bash
go build -o server ./cmd/server
```

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: add main entry point"
```

---

## Task 11: 前端初始化

**Files:**
- Create: `web/package.json`
- Create: `web/vite.config.js`
- Create: `web/tailwind.config.js`
- Create: `web/postcss.config.js`
- Create: `web/index.html`

- [ ] **Step 1: 创建 package.json**

```json
{
  "name": "servermanage-web",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "vite": "^5.0.0",
    "tailwindcss": "^3.4.0",
    "postcss": "^8.4.0",
    "autoprefixer": "^10.4.0"
  }
}
```

- [ ] **Step 2: 创建 vite.config.js**

```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 8080,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:8081',
        ws: true
      }
    }
  }
})
```

- [ ] **Step 3: 创建 tailwind.config.js**

```javascript
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {}
  },
  plugins: []
}
```

- [ ] **Step 4: 创建 postcss.config.js**

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {}
  }
}
```

- [ ] **Step 5: 创建 index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>ServerManage</title>
  <link href="/src/style.css" rel="stylesheet">
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>
```

- [ ] **Step 6: Commit**

```bash
git add web/package.json web/vite.config.js web/tailwind.config.js web/postcss.config.js web/index.html
git commit -m "feat: init frontend project"
```

---

## Task 12: 前端核心代码

**Files:**
- Create: `web/src/main.js`
- Create: `web/src/style.css`
- Create: `web/src/App.vue`
- Create: `web/src/api/index.js`
- Create: `web/src/components/ServiceList.vue`
- Create: `web/src/components/ServiceItem.vue`
- Create: `web/src/components/ServiceModal.vue`

- [ ] **Step 1: 创建 main.js**

```javascript
import { createApp } from 'vue'
import App from './App.vue'
import './style.css'

createApp(App).mount('#app')
```

- [ ] **Step 2: 创建 style.css**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

body {
  background-color: #f5f5f5;
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}
```

- [ ] **Step 3: 创建 api/index.js**

```javascript
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

export const getServices = () => api.get('/services')
export const createService = (data) => api.post('/services', data)
export const updateService = (id, data) => api.put(`/services/${id}`, data)
export const deleteService = (id) => api.delete(`/services/${id}`)
export const startService = (id) => api.post(`/services/${id}/start`)
export const stopService = (id) => api.post(`/services/${id}/stop`)
export const restartService = (id) => api.post(`/services/${id}/restart`)
export const getLogs = (id, lines = 100) => api.get(`/services/${id}/logs?lines=${lines}`)
export const discover = (dirs) => api.post('/discover', { dirs })

export default api
```

- [ ] **Step 4: 创建 App.vue**

```vue
<template>
  <div class="min-h-screen bg-gray-100">
    <header class="bg-white shadow">
      <div class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8 flex justify-between items-center">
        <h1 class="text-3xl font-bold text-gray-900">ServerManage</h1>
        <button
          @click="startAll"
          :disabled="loading"
          class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded disabled:opacity-50"
        >
          启动全部
        </button>
      </div>
    </header>

    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <ServiceList
        :services="services"
        @refresh="loadServices"
        @start="handleStart"
        @stop="handleStop"
        @restart="handleRestart"
        @edit="editService"
        @delete="handleDelete"
      />

      <div class="mt-4 flex gap-4">
        <button
          @click="showModal = true"
          class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
        >
          + 添加服务
        </button>
      </div>

      <ServiceModal
        v-if="showModal"
        :service="editingService"
        @close="closeModal"
        @save="saveService"
      />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getServices, createService, updateService, deleteService, startService, stopService, restartService } from './api'
import ServiceList from './components/ServiceList.vue'
import ServiceModal from './components/ServiceModal.vue'

const services = ref([])
const showModal = ref(false)
const editingService = ref(null)
const loading = ref(false)

const loadServices = async () => {
  try {
    const res = await getServices()
    services.value = res.data
  } catch (e) {
    console.error('Failed to load services:', e)
  }
}

const handleStart = async (id) => {
  await startService(id)
  await loadServices()
}

const handleStop = async (id) => {
  await stopService(id)
  await loadServices()
}

const handleRestart = async (id) => {
  await restartService(id)
  await loadServices()
}

const startAll = async () => {
  loading.value = true
  for (const svc of services.value) {
    if (svc.status === 'stopped') {
      await startService(svc.id)
    }
  }
  await loadServices()
  loading.value = false
}

const editService = (service) => {
  editingService.value = service
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingService.value = null
}

const saveService = async (data) => {
  if (editingService.value) {
    await updateService(editingService.value.id, data)
  } else {
    await createService(data)
  }
  await loadServices()
  closeModal()
}

const handleDelete = async (id) => {
  if (confirm('确定要删除这个服务吗？')) {
    await deleteService(id)
    await loadServices()
  }
}

onMounted(() => {
  loadServices()
})
</script>
```

- [ ] **Step 5: 创建 ServiceList.vue**

```vue
<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-md">
    <ul class="divide-y divide-gray-200">
      <li v-for="service in services" :key="service.id" class="px-4 py-4 sm:px-6">
        <ServiceItem
          :service="service"
          @start="$emit('start', service.id)"
          @stop="$emit('stop', service.id)"
          @restart="$emit('restart', service.id)"
          @edit="$emit('edit', service)"
          @delete="$emit('delete', service.id)"
        />
      </li>
      <li v-if="services.length === 0" class="px-4 py-8 text-center text-gray-500">
        暂无服务，请添加服务
      </li>
    </ul>
  </div>
</template>

<script setup>
import ServiceItem from './ServiceItem.vue'

defineProps({
  services: {
    type: Array,
    default: () => []
  }
})

defineEmits(['refresh', 'start', 'stop', 'restart', 'edit', 'delete'])
</script>
```

- [ ] **Step 6: 创建 ServiceItem.vue**

```vue
<template>
  <div class="flex items-center justify-between">
    <div class="flex items-center">
      <span
        :class="[
          'inline-flex items-center justify-center h-3 w-3 rounded-full mr-3',
          service.status === 'running' ? 'bg-green-500' : 'bg-gray-400'
        ]"
      ></span>
      <div>
        <p class="text-lg font-medium text-gray-900">{{ service.name }}</p>
        <p class="text-sm text-gray-500">
          {{ service.status === 'running' ? '运行中' : '已停止' }}
          <span v-if="service.port"> :{{ service.port }}</span>
        </p>
      </div>
    </div>

    <div class="flex items-center gap-2">
      <button
        v-if="service.status === 'running'"
        @click="$emit('stop')"
        class="bg-red-500 hover:bg-red-600 text-white px-3 py-1 rounded text-sm"
      >
        停止
      </button>
      <button
        v-else
        @click="$emit('start')"
        class="bg-green-500 hover:bg-green-600 text-white px-3 py-1 rounded text-sm"
      >
        启动
      </button>

      <button
        @click="$emit('restart')"
        :disabled="service.status !== 'running'"
        class="bg-yellow-500 hover:bg-yellow-600 text-white px-3 py-1 rounded text-sm disabled:opacity-50"
      >
        重启
      </button>

      <button
        @click="$emit('edit')"
        class="bg-gray-500 hover:bg-gray-600 text-white px-3 py-1 rounded text-sm"
      >
        编辑
      </button>

      <button
        @click="$emit('delete')"
        class="bg-gray-300 hover:bg-gray-400 text-gray-700 px-3 py-1 rounded text-sm"
      >
        删除
      </button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  service: {
    type: Object,
    required: true
  }
})

defineEmits(['start', 'stop', 'restart', 'edit', 'delete'])
</script>
```

- [ ] **Step 7: 创建 ServiceModal.vue**

```vue
<template>
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg shadow-xl p-6 w-full max-w-md">
      <h2 class="text-xl font-bold mb-4">{{ service ? '编辑服务' : '添加服务' }}</h2>

      <form @submit.prevent="handleSubmit">
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">服务名称</label>
            <input
              v-model="form.name"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">启动脚本</label>
            <input
              v-model="form.startScript"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">停止脚本</label>
            <input
              v-model="form.stopScript"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">重启脚本</label>
            <input
              v-model="form.restartScript"
              type="text"
              required
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">工作目录</label>
            <input
              v-model="form.workDir"
              type="text"
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700">端口</label>
            <input
              v-model.number="form.port"
              type="number"
              class="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>

          <div class="flex items-center">
            <input
              v-model="form.autoStart"
              type="checkbox"
              id="autoStart"
              class="h-4 w-4 text-blue-600 border-gray-300 rounded"
            />
            <label for="autoStart" class="ml-2 block text-sm text-gray-900">
              开机自启动
            </label>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button
            type="button"
            @click="$emit('close')"
            class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
          >
            取消
          </button>
          <button
            type="submit"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            保存
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  service: Object
})

const emit = defineEmits(['close', 'save'])

const form = ref({
  name: '',
  startScript: '',
  stopScript: '',
  restartScript: '',
  workDir: '',
  port: 0,
  autoStart: false
})

watch(() => props.service, (newVal) => {
  if (newVal) {
    form.value = { ...newVal }
  }
}, { immediate: true })

const handleSubmit = () => {
  emit('save', { ...form.value })
}
</script>
```

- [ ] **Step 8: 安装前端依赖并构建**

```bash
cd web
npm install
npm run build
```

- [ ] **Step 9: Commit**

```bash
git add web/src web/package-lock.json
git commit -m "feat: add frontend code"
```

---

## Task 13: 测试与运行

- [ ] **Step 1: 编译后端**

```bash
go build -o server ./cmd/server
```

- [ ] **Step 2: 运行服务**

```bash
./server
```

- [ ] **Step 3: 访问 Web 界面**

打开浏览器访问 http://localhost:8081

- [ ] **Step 4: 测试功能**

1. 添加服务，填写脚本路径
2. 测试启动/停止/重启
3. 测试开机自启动

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "chore: complete project and test"
```

---

## 验收标准检查

| 功能 | 状态 |
|------|------|
| 服务管理（添加/删除/编辑） | Task 2-5, 9 |
| 进程控制（启动/停止/重启） | Task 5 |
| 状态显示 | Task 5 |
| 开机自启动 | Task 6 |
| 日志查看 | Task 7, 8 |
| Web 界面 | Task 11-12 |