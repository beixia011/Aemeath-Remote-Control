# Aemeath-Remote-Control V0.1.0

本项目仍处于开发阶段，可能会存在未知明BUG

本项目是一个可运行的 WebRTC 远程控制系统，目标是通过浏览器直接查看并控制被控端桌面，包含三部分：

1. Go 服务端：鉴权、HTTP API、WebSocket 信令、设备管理、会话管理、审计日志、安全中间件。
2. Python 被控端 Agent：屏幕采集、WebRTC 推流、键鼠输入注入、心跳上报。
3. Web 前端控制台：设备选择、会话发起、桌面控制、移动端手势与输入、传输模式可视化、双语切换。

## 1. 项目能力概览

1. 基于 WebRTC 的远程桌面视频传输与控制。
2. 基于角色的登录鉴权（`admin` / `viewer` / `agent`）。
3. 设备上线注册、在线状态维护、会话生命周期维护。
4. 可配置传输模式：默认 `P2P`，可切换 `TURN` 中继（TURN协议仍在开发，可能存在BUG）。
5. 前端展示“配置模式 + 实际链路”（P2P 直连 / TURN 中继 / 检测中）。
6. P2P 失败监测与告警链路（前端状态提示 + 上报服务端日志）。
7. 移动端独立优化：单指移动、双指滚动、长按右键、拖拽模式、软键盘桥接输入。
8. 安全加固：Origin 校验、安全响应头、登录限流、弱密码/弱密钥启动告警。
9. 审计能力：登录、连接、设备注册、会话起止、客户端告警等事件记录。
10. 多语言支持：前端英文/中文切换。

## 2. 总体架构与数据流

### 2.1 架构关系

```text
Viewer Browser
  |  HTTP (login/runtime/devices/sessions/audit)
  |  WebSocket (auth + signal + control + warning)
  v
Go Signal Server
  |  WebSocket (auth + register_device + signal + heartbeat)
  v
Python Agent
  |  屏幕采集 + WebRTC 视频轨
  |  控制事件注入（鼠标/键盘/滚轮/文本）
  v
被控主机桌面
```

### 2.2 关键流程

1. 登录：前端/Agent 调用 `POST /api/login` 获取 JWT。
2. 建立 WS：首条消息必须是 `{"type":"auth","token":"..."}`。
3. 设备注册：Agent 发送 `register_device`，服务端更新在线设备列表。
4. 发起会话：Viewer 发送 `start_session`，Hub 生成 `session_id`。
5. WebRTC 协商：Viewer 与 Agent 通过服务端中转 `offer/answer/candidate`。
6. 控制输入：Viewer 发送 `control`，Agent 注入到系统输入设备。
7. 会话结束：任一端发送 `session_end` 或断连触发自动回收。
8. 监控告警：P2P 失败时前端发送 `client_warning`，服务端审计并打印 warning。

## 3. 模块说明（按子系统）

### 3.1 服务端模块（Go）

1. `cmd/signal` 启动模块
   负责加载环境变量、初始化鉴权/Hub/审计/HTTP 服务、打印启动告警、处理优雅退出。
2. `internal/config` 配置模块
   负责解析 `WEBRTC_MODE`、STUN/TURN、`ALLOWED_ORIGINS`、登录限流参数，并返回结构化配置与警告。
3. `internal/auth` 鉴权模块
   负责从环境变量构建用户表、签发 JWT、验证 JWT、输出角色信息。
4. `internal/httpapi` 接口与网关模块
   负责 REST API、WebSocket 接入、安全头、Origin 校验、登录限流、请求日志与权限校验。
5. `internal/hub` 会话中心模块
   负责客户端注册、设备在线表、会话表、信令转发、控制转发、状态日志快照、事件广播。
6. `internal/model` 数据模型模块
   定义角色、设备/显示器、会话状态、WebRTC 配置等跨模块结构。
7. `internal/audit` 审计模块
   负责事件落盘（JSON 行日志）和内存窗口缓存（用于 API 查询）。

### 3.2 被控端模块（Python Agent）

1. `AgentConfig`
   运行配置结构，承载服务地址、账号、设备标识、心跳周期。
2. `ScreenTrack`
   WebRTC 视频轨实现，基于 `mss + numpy + av` 抓屏并按 FPS/质量缩放。
3. `ControlInjector`
   控制注入器，处理鼠标移动/按键、滚轮、键盘按键与文本输入事件。
4. `RemoteAgent`
   连接管理主流程，负责登录、注册设备、心跳、会话建立、信令处理、会话清理与重连。
5. `load_config/load_env_file`
   环境变量加载与默认值组装，支持 `ENV_FILE` 指定 `.env` 文件路径。

### 3.3 前端模块（Web）

1. 视图层（`index.html` + `style.css`）
   提供登录视图、设备列表、会话参数面板、视频舞台、状态栏、移动端按钮与语言切换控件。
2. 状态与国际化（`app.js`）
   管理全局状态、`I18N` 字典、语言持久化与文案刷新。
3. 会话与信令（`app.js`）
   负责登录、拉取运行配置、建立 WebSocket、处理 Hub 消息、发起/结束会话。
4. WebRTC 连接管理（`app.js`）
   负责 `RTCPeerConnection` 生命周期、ICE candidate 交换、链路类型探测。
5. 桌面控制模块（`app.js`）
   PC 端鼠标、键盘事件采样与控制消息发送。
6. 移动端控制模块（`app.js`）
   单指移动点击、长按右键、双指滚动、拖拽模式、软键盘桥接输入。

## 4. 文件说明（逐文件）

说明：以下清单基于当前工作区中可见的业务文件/配置文件/运行产物；`缓存目录`不逐项展开。

| 文件路径 | 类型 | 作用 |
| --- | --- | --- |
| `README.md` | 文档 | 项目文档 |
| `server/cmd/signal/main.go` | 服务端入口 | 进程启动、配置加载、服务初始化、HTTP Server 生命周期管理。 |
| `server/internal/audit/audit.go` | 模块源码 | 审计日志结构、落盘和查询逻辑。 |
| `server/internal/auth/auth.go` | 模块源码 | 用户表、登录签发 JWT、Token 校验。 |
| `server/internal/config/config.go` | 模块源码 | 环境变量解析、WebRTC/Security 配置构建与参数兜底。 |
| `server/internal/hub/hub.go` | 模块源码 | 连接中心：设备/会话状态管理、消息路由、状态日志快照。 |
| `server/internal/httpapi/server.go` | 模块源码 | HTTP/WS 路由、安全中间件、鉴权、限流、读写泵。 |
| `server/internal/model/model.go` | 模块源码 | 共享数据模型定义。 |
| `server/web/index.html` | 前端页面 | 控制台 DOM 结构、控件容器。 |
| `server/web/style.css` | 前端样式 | 控制台布局、主题、移动端适配样式。 |
| `server/web/app.js` | 前端逻辑 | 登录、设备管理、WS 信令、WebRTC、PC/移动端控制、i18n。 |
| `agent/requirements.txt` | 依赖声明 | Agent Python 依赖版本。 |
| `agent/agent.py` | Agent 主程序 | 抓屏、信令、会话、输入注入、重连主逻辑。 |


## 5. 配置项说明

### 5.1 服务端配置（`server/.env`）

| 变量 | 说明 | 默认/示例 |
| --- | --- | --- |
| `ENV_FILE` | 指定 env 文件路径 | `.env` |
| `HTTP_ADDR` | HTTP 监听地址 | `:8080` |
| `STATIC_DIR` | 前端静态目录 | `./web` |
| `AUDIT_LOG_PATH` | 审计日志文件路径 | `./audit.log` |
| `JWT_SECRET` | JWT 密钥 | `change-me-please`（应替换） |
| `ALLOWED_ORIGINS` | 允许的浏览器 Origin，逗号分隔，`*` 代表全部 | `*` |
| `LOGIN_MAX_ATTEMPTS` | 登录窗口内最大失败次数 | `10` |
| `LOGIN_WINDOW_SEC` | 登录限流时间窗口（秒） | `60` |
| `WEBRTC_MODE` | 传输模式：`p2p` 或 `turn` | `p2p` |
| `STUN_URLS` | STUN 服务器列表，逗号分隔 | `stun:stun.l.google.com:19302` |
| `TURN_URLS` | TURN 服务器列表，逗号分隔 | 需自行配置 |
| `TURN_USERNAME` | TURN 用户名 | 需自行配置 |
| `TURN_CREDENTIAL` | TURN 密码 | 需自行配置 |
| `ADMIN_USER` / `ADMIN_PASS` | 管理员账号 | `admin` / `admin123` |
| `VIEWER_USER` / `VIEWER_PASS` | 控制端账号 | `viewer` / `viewer123` |
| `AGENT_USER` / `AGENT_PASS` | 被控端账号 | `agent` / `agent123` |

### 5.2 Agent 配置（`agent/.env` 或 `agent/.env.example`）

| 变量 | 说明 | 默认/示例 |
| --- | --- | --- |
| `ENV_FILE` | 指定 env 文件路径 | `.env` |
| `SERVER_HTTP` | 服务端 HTTP 地址 | `http://127.0.0.1:8080` |
| `SERVER_WS` | 服务端 WS 地址，不配则由 `SERVER_HTTP` 推导 | `ws://127.0.0.1:8080/ws` |
| `AGENT_USER` / `AGENT_PASS` | Agent 登录账号 | `agent` / `agent123` |
| `DEVICE_ID` | 设备唯一 ID | 默认主机名+随机后缀 |
| `DEVICE_NAME` | 设备显示名 | 默认系统节点名 |

## 6. API 与 WebSocket 协议

### 6.1 HTTP API

1. `POST /api/login`：登录并返回 Token/角色。
2. `GET /api/runtime`：获取当前 WebRTC 配置（鉴权后）。
3. `GET /api/devices`：获取在线设备列表（viewer/admin）。
4. `GET /api/sessions`：获取会话列表（viewer/admin）。
5. `GET /api/audit`：获取审计事件（admin）。
6. `GET /healthz`：健康检查。

### 6.2 WebSocket 消息类型（核心）

1. 客户端上行：`auth`、`register_device`、`heartbeat`、`start_session`、`signal`、`control`、`session_end`、`client_warning`、`list_devices`。
2. 服务端下行：`ws_ready`、`devices_updated`、`devices`、`device_registered`、`session_created`、`session_offer_request`、`signal`、`session_ended`、`error`。

## 7. 本地运行

### 7.1 启动服务端

```bash
cd server
go mod tidy
go run ./cmd/signal
```

### 7.2 启动 Agent

```bash
cd agent
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt
python agent.py
```

### 7.3 启动前端

浏览器打开 `http://127.0.0.1:8080`，使用 `viewer` 或 `admin` 登录，选择设备后连接。

## 8. 生产与安全建议

1. 公网部署必须替换默认账号密码和 `JWT_SECRET`。
2. `ALLOWED_ORIGINS` 不建议使用 `*`，应改成明确域名列表。
3. 默认优先 `P2P`；复杂网络建议配置 `TURN` 并可切换 `WEBRTC_MODE=turn`。
4. 结合反向代理（Nginx/Caddy）启用 HTTPS/WSS。
5. 定期轮转审计日志并保护 `audit.log` 访问权限。
