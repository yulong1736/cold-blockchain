# 项目文件目录结构

项目根目录即包含 `main.go`、`backend/`、`frontend/` 的目录（可能命名为 `final-design/` 或 `cold-chain-trace/`）。后端须在根目录执行，以正确加载 `backend/config/config.yaml`（缺失时回退 `config.example.yaml`）与 `frontend/dist`。

```
<项目根目录>/
│
├── main.go                          # 应用入口：加载配置、初始化 DB、区块链、路由并启动服务（在根目录执行 go run main.go）
├── go.mod / go.sum                  # Go 模块依赖（根目录）
├── README.md                        # 项目说明
├── QUICKSTART.md                    # 快速启动与安装部署
├── PROJECT_STRUCTURE.md             # 本文件
├── docker-compose.yml               # 容器编排（postgres/backend/frontend，默认 5432/8082/8081）
├── Dockerfile.backend               # 后端多阶段镜像构建（Go builder -> Alpine runtime）
├── PROGRESS_REPORT.md               # （可选）进度汇报；仓库中未必包含此文件
│
├── backend/                         # 后端逻辑（config、models、controllers 等）
│   ├── config/
│   │   ├── config.example.yaml      # 配置模板（已提交；复制为 config.yaml 填写真实值）
│   │   └── config.go                # 配置加载、GetDSN()
│   │
│   ├── models/
│   │   ├── user.go                  # User（含 Role, CompanyName）、UserRole 常量
│   │   └── product.go               # Product, ProductHistory, TemperatureRecord, TransportNode
│   │
│   ├── database/
│   │   ├── database.go             # InitDB、AutoMigrate、CloseDB
│   │   └── migrations/
│   │       └── 001_init.sql        # （可选）手工初始化表结构，见 QUICKSTART.md
│   │
│   ├── controllers/
│   │   ├── auth_controller.go       # 注册、登录、忘记密码验证码、UpdateProfile、ChangePassword
│   │   ├── product_controller.go    # 商品 CRUD、公开/消费者追溯、GetAuditLogs
│   │   ├── monitor_controller.go    # 温控、运输节点、上链同步、ListMyProducts
│   │   ├── dashboard_controller.go  # GetStats、GetTemperatureTrend、GetTransportMap、GetAlerts、GetAllAlerts
│   │   ├── blockchain_controller.go # GetStatus（链状态只读，供前端轮询）
│   │   └── common.go                # 控制器公共能力：请求用户上下文、统一错误响应、limit 解析
│   │
│   ├── services/
│   │   ├── user_service.go          # 用户 CRUD、校验密码、改用户名/密码
│   │   ├── product_service.go       # 商品/历史/温控/运输/审计；物流按 company_name 隔离；审计键值对
│   │   ├── dashboard_service.go     # 仪表盘统计与趋势、物流图、告警（按 role/company_name/warehouse_id 分流）
│   │   └── sql_helpers.go           # PostgreSQL IN 子句占位符与参数辅助
│   │
│   ├── blockchain/
│   │   └── fisco_client.go          # FISCO 3.x Web3 RPC（go-ethereum ethclient）、存证与查询、IsConnected
│   │
│   ├── middleware/
│   │   ├── auth.go                  # AuthMiddleware（设 user_id, username, role, company_name）、RoleMiddleware
│   │   └── cors.go                  # CORS 中间件（OPTIONS 预检与跨域响应头）
│   │
│   ├── routes/
│   │   └── routes.go                # 所有 API 注册；静态资源与 SPA 回退（frontend/dist）
│   │
│   └── utils/
│       ├── jwt.go                   # Claims（含 CompanyName）、GenerateToken、ParseToken
│       └── trace.go                 # 追溯码等工具
│
└── frontend/                        # 前端（Vue 2 + Element UI）
    ├── package.json
    ├── vue.config.js                # 开发时代理 /api -> 后端
    ├── public/index.html
    └── src/
        ├── main.js                  # Vue 入口、全局 filter formatDateTimeTZ
        ├── App.vue                  # 布局、侧栏（按角色显示仪表盘/商品/追溯/温控/运输/告警/监管审计/全局告警/账号）
        ├── views/
        │   ├── Login.vue
        │   ├── Register.vue         # 含角色下拉（producer/warehouse/logistics/consumer/regulator）、company_name
        │   ├── Dashboard.vue        # 统计、温度湿度趋势、物流轨迹；追溯码下拉过滤图表
        │   ├── Products.vue         # 生产商商品列表、创建/编辑/删除；编辑时仅提交变更字段
        │   ├── Trace.vue            # 追溯查询、操作历史（格式化块状展示）
        │   ├── Temperature.vue      # 温控记录、按追溯码汇总、TraceDetailDrawer
        │   ├── Transport.vue        # 运输节点、本公司追溯码下拉、TraceDetailDrawer
        │   ├── Alert.vue            # 仓储/物流告警（按角色分流）、轮询
        │   ├── Regulator.vue        # 监管审计日志、轮询
        │   ├── RegulatorAlert.vue   # 监管全局告警、轮询
        │   └── Account.vue          # 改用户名、改密码、退出
        ├── components/
        │   └── TraceDetailDrawer.vue # 按追溯码展示明细的抽屉
        ├── router/index.js          # 路由与 meta.roles、beforeEach 鉴权
        ├── store/index.js           # token、user、getters（含 companyName）
        ├── api/index.js             # Axios 实例、Authorization 注入
        ├── mixins/
        │   └── traceAlertMixin.js   # 告警页通用轮询/汇总/明细逻辑
        └── utils/
            ├── dateFormat.js        # formatDateTimeTZ、formatDateTimeNoTZ
            ├── traceSummary.js      # buildTraceSummary（按 trace_id 汇总）
            └── http.js              # unwrapData/getErrorMessage 通用 HTTP 辅助
```

## 核心模块说明

### 后端（backend/）

- **config**：服务端口、数据库连接、JWT 密钥与过期时间、区块链节点等。配置文件路径为 `backend/config/config.yaml`（从项目根目录启动时）；该文件含敏感信息已被 `.gitignore` 忽略，仓库仅提交 `config.example.yaml` 模板，缺失 `config.yaml` 时后端自动回退到示例模板。
- **models**：与表对应的结构体；User 含 CompanyName，Product 含 TransportCompany，ProductHistory 含 ChangedFields/ChangeDetails。
- **database**：GORM 连接 PostgreSQL，AutoMigrate 创建/更新表。
- **controllers**：解析请求、从 context 取 user_id/role/company_name，调用 service，返回 JSON。
- **services**：业务与多租户逻辑（生产商按 producer_id、物流按 company_name、仓储按 warehouse_id；消费者按关联商品追溯码）。
- **blockchain**：FISCO 存证与查询；仪表盘用 IsConnected() 显示链状态。
- **middleware**：JWT 解析后写入 context；RoleMiddleware 限制接口角色。
- **routes**：除 API 外，对 GET 请求先尝试 frontend/dist 静态文件，再回退到 index.html（SPA）。

### 前端（frontend/）

- **views**：各业务页；Dashboard 支持按追溯码筛选图表；Transport/Alert 等按角色或公司隔离。
- **components**：TraceDetailDrawer 复用于温控、运输、告警等按追溯码查看明细。
- **router**：以 meta.roles + beforeEach 控制访问；监管有 /regulator、/regulator-alerts。
- **store**：登录后存 token 与 user（含 company_name），请求头带 Authorization。
- **utils**：统一时间展示、按 trace_id 汇总列表。

## 数据流向

- 请求：浏览器 → Vue Router → 页面 → api/index.js（带 Token）→ Gin → Middleware → Controller → Service → DB/区块链 → JSON → 前端渲染。
- 多租户：Controller 从 context 取 role、company_name、user_id，Service 内按角色决定 trace_id 集合或 warehouse_id，再查库。

## 启动与部署

- **开发**：在项目根目录执行 `go run main.go` 或 `go run .` 启动后端；可选 `cd frontend && npm run serve` 启动前端开发服务器（代理到后端）。
- **生产**：在项目根目录执行 `go run main.go`；需先在 `frontend` 下执行 `npm run build`，后端从根目录下的 `frontend/dist` 提供页面，默认端口见 `backend/config/config.yaml`（如 8082）。
- **Compose**：`docker-compose up -d` 启动三容器；前端 `8081` 通过 nginx 反代 `/api` 到 `backend:8082`，后端端口需与 `config.yaml` 一致。
