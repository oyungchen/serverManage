# ServerManage 设计文档

## 1. 项目概述

**项目名称：** ServerManage
**项目类型：** 本机服务管理工具
**核心功能：** 通过 Web 界面管理本机多个服务的启动、停止、重启，实现开机自启动
**目标用户：** 开发者在 macOS 上管理多个本地服务

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                      Web UI (Vue 3)                     │
│              http://localhost:8080                      │
└─────────────────────┬───────────────────────────────────┘
                      │ HTTP / WebSocket
┌─────────────────────▼───────────────────────────────────┐
│                      Go 后端                            │
│              http://localhost:8081                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│  │  服务管理   │ │  脚本执行   │ │  日志管理   │       │
│  │  (CRUD)     │ │  (启动/停止)│ │  (实时)     │       │
│  └─────────────┘ └─────────────┘ └─────────────┘       │
└─────────────────────────────────────────────────────────┘
```

---

## 3. 技术栈

| 层级 | 技术选型 |
|------|----------|
| 后端 | Go + Gin 框架 |
| 前端 | Vue 3 + Vite + Tailwind CSS |
| 通信 | REST API + WebSocket |
| 数据存储 | JSON 文件 (`~/.serverManage/services.json`) |
| 自启动 | macOS launchd |

---

## 4. 功能设计

### 4.1 服务配置

每个服务只需配置脚本路径：

```yaml
services:
  - name: my-app
    startScript: ./start.sh
    stopScript: ./stop.sh
    restartScript: ./restart.sh
    port: 3000
    autoStart: true
    workDir: /Users/didi/projects/my-app
```

**字段说明：**

| 字段 | 必填 | 说明 |
|------|------|------|
| name | 是 | 服务名称，唯一标识 |
| startScript | 是 | 启动脚本路径 |
| stopScript | 是 | 停止脚本路径 |
| restartScript | 是 | 重启脚本路径 |
| port | 否 | 端口，用于健康检查 |
| autoStart | 否 | 是否开机自启动，默认 false |
| workDir | 否 | 工作目录 |

### 4.2 服务管理

- **添加服务：** 手动填写配置或通过目录扫描自动发现
- **删除服务：** 移除配置，同时移除 launchd 配置
- **编辑服务：** 修改配置
- **服务列表：** 显示所有服务及其状态

### 4.3 进程控制

- **启动：** 执行 startScript，通过 process group 管理
- **停止：** 执行 stopScript
- **重启：** 执行 restartScript
- **状态检测：** 检查进程是否存在，或 HTTP 健康检查

### 4.4 开机自启动

- 生成 `com.servermanage.<服务名>.plist` 到 `~/Library/LaunchAgents/`
- 使用 `launchctl load` 加载服务
- 电脑重启后自动执行 startScript

### 4.5 日志管理

- 捕获脚本的标准输出和标准错误
- 通过 WebSocket 实时推送到前端
- 本地文件存储，按服务名+日期归档

---

## 5. 数据结构

### 5.1 服务配置

`~/.serverManage/services.json`:

```json
{
  "services": [
    {
      "id": "uuid",
      "name": "my-app",
      "startScript": "./start.sh",
      "stopScript": "./stop.sh",
      "restartScript": "./restart.sh",
      "port": 3000,
      "autoStart": true,
      "workDir": "/Users/didi/projects/my-app",
      "status": "running",
      "createdAt": "2026-04-15T10:00:00Z"
    }
  ]
}
```

### 5.2 launchd 配置模板

`~/Library/LaunchAgents/com.servermanage.<name>.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.servermanage.<name></string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>-c</string>
        <string>cd <workDir> && <startScript></string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
```

---

## 6. API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/services | 获取所有服务 |
| POST | /api/services | 添加服务 |
| PUT | /api/services/:id | 更新服务 |
| DELETE | /api/services/:id | 删除服务 |
| POST | /api/services/:id/start | 启动服务 |
| POST | /api/services/:id/stop | 停止服务 |
| POST | /api/services/:id/restart | 重启服务 |
| GET | /api/services/:id/logs | 获取历史日志 |
| GET | /api/ws/logs/:serviceId | WebSocket 日志流 |
| POST | /api/discover | 扫描目录自动发现服务 |

---

## 7. Web 界面设计

### 7.1 布局

```
┌────────────────────────────────────────────────────────┐
│  ServerManage                              [启动全部]  │
├────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────┐ │
│  │ ● my-app        运行中  :3000  [停止] [日志]     │ │
│  │ ○ python-api    已停止  :8000  [启动] [日志]     │ │
│  │ ● go-server     运行中  :8080  [停止] [日志]     │ │
│  └──────────────────────────────────────────────────┘ │
│                                                        │
│  [+ 添加服务]  [🔍 自动发现]  [📁 导入配置]          │
└────────────────────────────────────────────────────────┘
```

### 7.2 添加服务弹窗

```
┌────────────────────────────────────┐
│  添加服务                          │
├────────────────────────────────────┤
│  服务名称: [________________]      │
│  启动脚本: [________________] [📁] │
│  停止脚本: [________________] [📁] │
│  重启脚本: [________________] [📁] │
│  工作目录: [________________] [📁] │
│  端口:     [________________]      │
│  ☑ 开机自启动                    │
│                                    │
│         [取消]  [保存]            │
└────────────────────────────────────┘
```

---

## 8. 目录结构

```
serverManage/
├── cmd/
│   └── server/
│       └── main.go           # 入口文件
├── internal/
│   ├── config/
│   │   └── config.go         # 配置加载
│   ├── handler/
│   │   └── handler.go        # HTTP 处理
│   ├── model/
│   │   └── service.go        # 数据模型
│   ├── service/
│   │   └── manager.go        # 服务管理逻辑
│   ├── launcher/
│   │   └── launchd.go        # launchd 操作
│   └── logger/
│       └── logger.go         # 日志管理
├── web/
│   ├── index.html
│   ├── src/
│   │   ├── main.js
│   │   ├── App.vue
│   │   └── ...
│   ├── package.json
│   └── vite.config.js
├── config.yaml
└── go.mod
```

---

## 9. 验收标准

1. **服务管理：** 可以添加、删除、编辑服务
2. **进程控制：** 可以启动、停止、重启服务
3. **状态显示：** 实时显示服务运行状态
4. **开机自启动：** 勾选后电脑重启自动启动服务
5. **日志查看：** Web 界面实时查看服务日志
6. **自动发现：** 扫描目录可自动识别服务并生成配置