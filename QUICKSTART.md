# 快速启动指南

## 系统要求

- 操作系统: Linux / macOS / Windows
- Go 版本: 1.25 或更高
- Node.js 版本: 14.x 或更高
- PostgreSQL 版本: 12 或更高

## ✅ 当前状态

项目已编译成功，代码结构完整。**可以直接启动**，需满足以下前置条件。

---

## 🚀 方式一：Docker Compose 部署（推荐）

`docker-compose.yml` 包含三个服务：`postgres`（5432）、`backend`（宿主机映射 8082）、`frontend`（8081，内置 nginx 将 `/api` 代理到 backend）。

1. **配置应用**

   首次使用请先复制 `backend/config/config.example.yaml` 为 `backend/config/config.yaml` 并填写真实配置。Compose 会通过 `volumes: ./backend/config:/app/backend/config` 挂载该目录；`docker-compose.yml` 中的 `DB_*` 环境变量会覆盖配置文件里对应的数据库字段。若未创建 `config.yaml`，后端会自动回退到 `config.example.yaml`。

   **重要**：容器内后端监听端口由 `config.yaml` 的 `server.port` 决定，须与 `frontend/nginx.conf` 中 `proxy_pass http://backend:<端口>` 一致（默认均为 **8082**）。

2. **启动服务**

   ```bash
   docker-compose up -d
   ```

3. **访问应用**

   - 前端：http://localhost:8081
   - 后端 API：http://localhost:8082

---

## 🛠 方式二：手动部署

以下命令均在**项目根目录**（包含 `main.go` 与 `backend/`、`frontend/` 的目录）执行。

### 1. PostgreSQL 数据库

**检查数据库是否运行：**
```bash
# Windows PowerShell
Get-Service -Name postgresql*

# 或尝试连接
psql -U postgres -h localhost -p 5432
```

**如果未安装或未运行：**

方式一 —— Docker（推荐）：
```bash
docker run -d --name postgres-coldchain \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=cold_chain_db \
  -p 5432:5432 \
  postgres:14
```

方式二 —— 手动安装 PostgreSQL 后创建库：
```sql
CREATE DATABASE cold_chain_db;
```

**数据库表结构**：启动后端时 GORM 会执行 AutoMigrate 自动建表（见 `backend/database/database.go`）。若环境禁止自动迁移，可手工执行：
```bash
psql -U postgres -d cold_chain_db -f backend/database/migrations/001_init.sql
```
（与 AutoMigrate 同时使用可能重复，择一即可。）

### 2. 配置文件

编辑 `backend/config/config.yaml`：
```yaml
server:
  port: 8082        # 服务端口
  mode: debug

database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: cold_chain_db
  sslmode: disable
  max_connections: 100

jwt:
  secret: your-secret-key-change-in-production
  expires_hours: 24
```

### 3. 后端

```bash
# 安装依赖
go mod download

# 运行（项目根目录）
go run main.go
```

或编译后运行：
```bash
go build -o cold-chain-trace.exe .
.\cold-chain-trace.exe
```

### 4. 前端

```bash
cd frontend
npm install

# 开发模式
npm run serve          # → http://localhost:8081（vue.config.js 将 /api 代理到 8082）

# 生产构建（供后端提供 SPA）
npm run build          # 产物在 dist/，后端自动从该目录提供页面
```

> `src/api/index.js` 默认使用 `baseURL: process.env.VUE_APP_API_BASE_URL || '/api'`。本地 `npm run serve` 时无需额外配置。需要指定完整 URL 时创建 `frontend/.env`：`VUE_APP_API_BASE_URL=http://localhost:8082/api`

### 5. FISCO BCOS 区块链节点（可选）

后端通过 **Web3 JSON-RPC** 连接 FISCO BCOS 3.x 节点（使用 go-ethereum `ethclient`）。连接失败不会阻止应用启动，仅显示警告。

#### 5.1 搭建节点

使用 WeBASE 快速搭建（参考 [WeBASE 文档](https://fisco-bcos-doc.readthedocs.io/zh-cn/latest/docs/operation_and_maintenance/webase.html)）：
```bash
git clone https://github.com/WeBankFinTech/WeBASE-Deploy.git
cd WeBASE-Deploy
python3 deploy.py installAll
# WeBASE 默认地址: http://localhost:5000，默认账号 admin / Abcd1234
```

或使用 FISCO BCOS 控制台：
```bash
cd ~/fisco && curl -#LO https://gitee.com/FISCO-BCOS/console/raw/master/tools/download_console.sh
chmod +x download_console.sh && bash download_console.sh
cp -n console/conf/config-example.toml console/conf/config.toml
cp -r nodes/127.0.0.1/sdk/* console/conf
cd ~/fisco/console && chmod +x start.sh && bash start.sh
```

#### 5.2 部署智能合约

通过控制台部署：
```bash
[group0]: /> deploy TraceEvidence
transaction hash: 0x4b75694984fcd41d75e856a90d940ceab7aed25de992e1bde3e2a4405fbedcbd
contract address: 0x713ab8e2926bce2cf29efb0dd9a82f02bcfc96e6
```

或通过 WeBASE 图形界面：进入"合约管理" → "合约部署"，上传编译后的合约文件。

#### 5.3 提取私钥

```bash
openssl ec -in <控制台目录>/account/ecdsa/<地址>.pem -noout -text 2>/dev/null \
  | sed -n '/^[[:space:]]*priv:/,/^[[:space:]]*pub:/p' \
  | grep -oE '[0-9a-f]{2}' \
  | tr -d '\n'
echo
```

#### 5.4 配置链连接

编辑 `backend/config/config.yaml`：
```yaml
blockchain:
  fisco_bcos:
    node_url: "http://127.0.0.1:8545"   # FISCO BCOS RPC 地址
    group_id: 0
    chain_id: 1
    private_key: "<16进制私钥>"
    contract_address: "0x..."           # 部署合约后填入；留空则使用轻量存证模式
```

若节点在远程，建立 SSH 端口转发：
```bash
ssh -N -p 2222 -L 8545:127.0.0.1:8545 ubuntu@<远程主机>
```

#### 5.5 验证链连接

```bash
Invoke-RestMethod -Uri "http://127.0.0.1:8545" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

## ✅ 启动成功标志

```
Database connected successfully
Server starting on port 8082
```

若看到区块链警告（可忽略）：
```
Warning: Failed to initialize blockchain client: ...
Application will continue without blockchain features
```

---

## 🌐 访问应用

| 场景 | 地址 | 说明 |
|------|------|------|
| Docker Compose | http://localhost:8081 | 前端容器 |
| Docker Compose | http://localhost:8082 | 后端直连 |
| 本地 `npm run serve` | http://localhost:8081 | vue.config.js 代理 /api 到 8082 |
| 后端提供 SPA | http://localhost:8082 | 需先 `cd frontend && npm run build` |

---

## 🔐 鉴权与角色（多租户）

| 角色 | 说明 | 仪表盘 / 告警 / 运输 |
|------|------|----------------------|
| **生产商（producer）** | 仅看本账号下商品 | 仪表盘：商品总数、温度趋势、物流轨迹均按**自己拥有的追溯码**过滤；无告警菜单 |
| **仓储（warehouse）** | 记录温控数据 | 仪表盘：全局统计；告警页：仅看**本账号记录的**异常温控告警 |
| **物流（logistics）** | 按运输公司隔离 | 仪表盘与告警均按 **company_name = 商品 transport_company** 过滤；运输管理仅能查看/添加**本公司**商品的节点；追溯码下拉仅列出本公司商品 |
| **监管（regulator）** | 全量审计 | 仪表盘：**全部**商品与物流数据；左侧有「监管审计」「全局告警」，可查看全部审计日志与全部温控告警 |
| **消费者（consumer）** | 追溯 + 登录后看板 | **未登录**可用公开追溯 API；**登录后**仪表盘按**本人关联/收货商品**过滤，可访问追溯查询、账号设置；无商品管理/温控/运输/告警菜单 |

- 物流账号的 **company_name**（注册时填写）须与商品的 **transport_company** 完全一致，才能正确隔离。
- 修改 JWT 相关逻辑后，用户需**重新登录**以获取新 Token。

---

## 📝 首次启动后的操作

**注册用户：**
```bash
curl -X POST http://localhost:8082/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456","email":"admin@example.com","role":"producer","company_name":"测试公司"}'
```

**登录获取 Token：**
```bash
curl -X POST http://localhost:8082/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

> 物流账号注册时 `role=logistics`，`company_name` 填公司名（如「顺丰快递」），与商品「运输公司」字段一致方可看到对应商品与告警。

---

## 🔍 验证安装

```bash
# 检查后端服务（返回 JSON 即表示服务在监听）
curl -s "http://localhost:8082/api/products/trace/__ping_check__"

# 检查数据库
psql -U postgres -d cold_chain_db -c "SELECT COUNT(*) FROM users;"

# 检查前端
# 浏览器访问 http://localhost:8081（npm run serve）或 http://localhost:8082（后端提供 dist）
```

---

## ⚠️ 常见问题

**数据库连接失败**
- 检查 PostgreSQL 服务是否运行、`config.yaml` 中数据库连接信息是否正确、数据库 `cold_chain_db` 是否已创建

**端口被占用**
- 修改 `config.yaml` 中 `server.port`，或关闭占用该端口的程序。Compose 场景下还需同步修改 `frontend/nginx.conf` 的 `proxy_pass`

**访问 8082 显示 404**
- 确认已执行 `cd frontend && npm run build`，且 `frontend/dist/index.html` 存在

**登录/仪表盘报错 pq 语法错误**
- 多为 GORM 在 PostgreSQL 下对 `IN (?)` 与 slice 的占位符问题，当前已在 `dashboard_service` 中通过 `inClauseStrings` / `inClauseUints` 辅助函数规避，需确保运行的是最新编译的后端

**区块链连接失败**
- 检查 FISCO BCOS 节点是否运行、node_url 和端口是否正确、网络连接和 SSH 隧道是否正常

**合约调用失败**
- 检查合约是否已部署、`config.yaml` 中 `contract_address` 是否正确、私钥是否有权限

**前端无法连接后端**
- 检查后端是否启动、CORS 配置、API 地址配置

---

## 🏭 生产环境建议

1. 使用 HTTPS 并配置 SSL 证书
2. 定期备份 PostgreSQL 数据
3. 配置日志轮转和监控
4. 配置服务监控和告警
5. 安全加固：修改默认密码、配置防火墙规则、使用环境变量管理敏感信息
6. Compose 场景下，后端端口需同时与 `config.yaml` 的 `server.port` 和 `frontend/nginx.conf` 的 `proxy_pass` 保持一致

---

## 📚 更多信息

- README.md — 项目说明
- PROJECT_STRUCTURE.md — 项目结构说明
