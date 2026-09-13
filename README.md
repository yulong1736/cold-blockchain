# 基于区块链的跨境冷链物流溯源与监管系统

## 项目简介

本项目是基于 FISCO BCOS 区块链的跨境冷链物流溯源与监管系统，采用前后端分离架构，实现商品信息链上存证、多角色多租户隔离、温控与运输管理、告警分流及监管审计。

## 技术栈

### 后端
- **语言**: Go 1.25+（见根目录 `go.mod`）
- **框架**: Gin
- **数据库**: PostgreSQL（GORM）
- **区块链**: FISCO BCOS 3.x（通过 **go-ethereum** `ethclient` 连接节点 **Web3 JSON-RPC**；配置了 `contract_address` 时调用 `TraceEvidence` 合约，否则回退为 calldata 轻量存证，见 `backend/blockchain/fisco_client.go`）
- **认证**: JWT（含 user_id, username, role, company_name）

### 前端
- **框架**: Vue 2.x
- **UI**: Element UI
- **状态管理**: Vuex
- **路由**: Vue Router（History 模式）
- **HTTP**: Axios
- **图表**: ECharts

## 项目结构（当前）

```
final-design/
├── main.go                    # 应用入口（在项目根目录执行 go run main.go）
├── go.mod / go.sum            # Go 模块依赖
├── backend/                   # 后端逻辑
│   ├── config/
│   │   ├── config.example.yaml  # 配置模板（已提交；复制为 config.yaml 填写真实值）
│   │   └── config.go            # 配置加载、GetDSN()
│   ├── models/
│   │   ├── user.go            # 用户（含 role, company_name）
│   │   └── product.go         # Product, ProductHistory, TemperatureRecord, TransportNode
│   ├── database/
│   │   ├── database.go        # 连接与 AutoMigrate
│   │   └── migrations/        # 可选 SQL 初始化脚本
│   ├── controllers/           # 注册、登录、商品 CRUD、温控、运输、仪表盘、审计、公共响应/上下文工具
│   ├── services/              # 业务与多租户逻辑（含 SQL IN 子句辅助）
│   ├── blockchain/
│   │   └── fisco_client.go    # FISCO 3.x Web3 RPC（go-ethereum ethclient）
│   ├── middleware/
│   │   ├── auth.go            # JWT 解析、RoleMiddleware
│   │   └── cors.go            # CORS 中间件
│   ├── routes/
│   │   └── routes.go          # API 注册、静态资源与 SPA 回退（frontend/dist）
│   └── utils/
│       ├── jwt.go             # Token 生成/解析（含 company_name）
│       └── trace.go           # 追溯码工具
└── frontend/                  # Vue 2 + Element UI
    ├── src/
    │   ├── views/             # Login, Register, Dashboard, Products, Trace,
    │   │                     # Temperature, Transport, Alert, Regulator, RegulatorAlert, Account
    │   ├── components/        # TraceDetailDrawer 等
    │   ├── router/            # 路由与鉴权（meta.roles）
    │   ├── store/             # Vuex（token、user）
    │   ├── api/               # Axios 封装
    │   ├── mixins/            # 复用逻辑（如告警页轮询/汇总）
    │   └── utils/             # dateFormat, traceSummary, http
    └── dist/                  # 构建后由后端从根目录 frontend/dist 提供
```

详细目录与模块说明见 **PROJECT_STRUCTURE.md**。

## 功能模块概览

| 模块         | 说明 |
|--------------|------|
| 用户与鉴权   | 注册（含角色、company_name）、登录、JWT、改用户名/密码、按角色路由与数据隔离。 |
| 商品管理     | 生产商创建/编辑/删除商品，写库+链上存证，操作历史与审计日志（修改字段键值对）。 |
| 追溯查询     | 按追溯码查商品详情与操作历史（含区块链哈希变化）；**公开 API**，访客与消费者均可使用。 |
| 温控监控     | 仓储上报温湿度与地点，超限自动标为异常；按追溯码汇总与明细。 |
| 运输管理     | 物流按本公司（company_name = transport_company）添加/查看运输节点；追溯码下拉仅本公司商品。 |
| 仪表盘       | 统计（商品数、链上记录、异常数、当日操作等）、温度/湿度趋势、物流轨迹图；按角色/公司过滤。 |
| 告警         | 仓储/物流：告警页按仓库或物流公司分流；监管：全局告警页查看全部。 |
| 监管审计     | 监管查看全部商品操作审计（操作人、操作、追溯码、修改字段、具体变化、时间），可轮询刷新。 |

## 用户角色与数据可见范围

- **producer**：仅自己创建/拥有的商品；仪表盘仅自家数据。
- **warehouse**：温控记录为自己所录；告警仅自己产生的异常。
- **logistics**：运输节点与仪表盘/告警仅限 `transport_company = 当前用户 company_name` 的商品。
- **regulator**：全部商品、全部审计日志、全部告警。
- **consumer**：**未登录**可用公开追溯接口/`/trace`；**登录后**仪表盘按**本人关联/收货的商品**追溯码过滤（与 `dashboard_service` 中 consumer 逻辑一致），可访问追溯查询、账号设置等，无商品/温控/运输/告警管理。

## API 摘要

- `POST /api/auth/register`、`POST /api/auth/login`
- `POST /api/auth/forgot-password/send-code`、`POST /api/auth/forgot-password/reset`（**公开**，找回密码）
- `PUT /api/auth/profile`、`PUT /api/auth/password`（需登录）
- `GET /api/products/trace/:trace_id`、`GET /api/products/history/:trace_id`（**公开**，无需登录）
- `GET/POST /api/products`、`PUT /api/products/:id`、`DELETE /api/products/:id`（生产商）
- `GET /api/dashboard/stats`、`/api/dashboard/temperature-trend`、`/api/dashboard/transport-map`（按角色过滤）
- `GET /api/alerts`（仓储/物流，按角色分流）
- `GET /api/regulator/audit-logs`、`GET /api/regulator/alerts`（监管）
- `GET /api/blockchain/status`（生产商/仓储/物流，轮询链连接与上链进度）
- `GET/POST /api/temperature`、`POST /api/temperature/:id/sync-blockchain`（仓储）
- `GET/POST /api/transport`、`GET /api/transport/my-products`、`POST /api/transport/:id/sync-blockchain`、`PUT /api/transport/:id`（物流）
- `GET /api/consumer/products`、`GET /api/consumer/products/trace/:trace_id`、`GET /api/consumer/products/history/:trace_id`（消费者，仅本人名下商品）

详见 **QUICKSTART.md**、**PROJECT_STRUCTURE.md**。

## 安装与运行

- 环境：Go 1.25+（见 `go.mod`）、Node 14+、PostgreSQL 12+
- 数据库名：`cold_chain_db`（与 `backend/config/config.yaml` 一致）
- 配置：首次使用请复制 `backend/config/config.example.yaml` 为 `backend/config/config.yaml` 并填写真实值；`config.yaml` 含敏感信息已被 `.gitignore` 忽略，缺失时后端自动回退到示例模板
- 后端：在**项目根目录**执行 `go run main.go` 或 `go run .`；或 `go build -o cold-chain-trace.exe .` 后运行
- 前端生产：`cd frontend && npm install && npm run build`，后端从根目录下的 `frontend/dist` 提供页面
- Docker Compose：`docker-compose up -d` 后默认前端 `8081`、后端 `8082`、数据库 `5432`

### 启动前关键一致性（与 QUICKSTART 对齐）

- `backend/config/config.yaml` 的 `server.port` 默认是 `8082`
- `frontend/nginx.conf` 的 `proxy_pass` 需与后端容器端口一致（默认 `backend:8082`）
- 若走 Compose，`backend` 服务应保持 `8082:8082` 映射，避免与前端代理目标不一致
- 若不使用区块链，可先留空合约地址；链路初始化失败不会阻塞基础业务启动

更多见 **QUICKSTART.md**（含 Docker Compose：前端 **8081**、后端默认 **8082**，须与 `backend/config/config.yaml` 中 `server.port` 及 `frontend/nginx.conf` 的 upstream 一致）。

## 文档索引

- **QUICKSTART.md**：快速启动、配置、角色与多租户、常见问题。
- **QUICKSTART.md**：快速启动、配置、角色与多租户、安装部署、常见问题。
- **PROJECT_STRUCTURE.md**：目录与模块说明。
- **REGRESSION_CHECKLIST.md**（本地文档，不入库）：人工回归测试清单（角色、主业务链、接口冒烟）。
- **PROGRESS_REPORT.md**（若有）：已实现功能、实现要点、遇到的问题与可改进项。

## 许可证

MIT License
