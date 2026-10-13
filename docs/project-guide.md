# go-auth-admin 项目指南

> `go-gin-init-skill` 强制交付物，骨架取自
> `vibeCoding/backend/go/go-gin-init-skill/references/project-guide-template.md`，栈特定段填入本项目实况。

---

## 一、项目定位

`go-auth-admin` 是 **java-vue-admin 的 Go 复刻**：同一套前端、同一套接口契约，
把 Java（Spring Boot）后端换成 Go（Gin + GORM）实现，目标是**前端零改动直接桥接**。

契约的**唯一事实来源是 Java 侧**（`D:\projects\java-vue-admin\backend`），
Go 侧必须逐字对齐：端点 56/56、权限码 56/56、VO 字段、种子数据全部一致（已机器核对）。

**迭代前必读**：`SKILLS.md`（技能依赖清单 + 红线）。

---

## 二、技术栈

| 类别 | 技术 |
|------|------|
| 编程语言 | Go 1.26 |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | SQLite（纯 Go 驱动 `glebarez/sqlite`，CGO_ENABLED=0，零环境依赖） |
| 认证 | JWT（HS256），access 60 分钟 / refresh 7 天 |
| 密码散列 | bcrypt |
| 配置 | godotenv（`.env`，缺省回退代码内默认值） |
| 前端 | Vue 3.4 + Vite 5 + TypeScript + Pinia |
| 前端组件 | `vue-base-skill` 系列 15 个包（`file:` 本地依赖） |
| 前端主题 | `vue-theme-skill` tokens（`scripts/generate-tokens.cjs` 生成） |
| API 文档 | 自研接口注册中心：`/api/api-registry`、`/api/api-registry/openapi` |

> 后端端口固定 **8080**（前端 vite proxy 目标），前端固定 **5174**
> （5173 长期被同机另一个项目占用，见 `restart.sh` 注释）。

---

## 三、项目结构

```
go-auth-admin/
├── backend/
│   ├── main.go          # 装配 + AutoMigrate + Seed + 启动
│   ├── config.go        # 读 .env，默认值与 Java application.yml 对齐
│   ├── model.go         # 全部 GORM 模型（wg_sys_*）+ 软删除
│   ├── common.go        # 响应信封 / 分页 / AppError / JWTManager / 权限常量
│   ├── middleware.go    # CORS / 应用签名 / JWT 鉴权 / RequirePerm / 安全头
│   ├── repo.go          # 角色码 / 权限 / 数据范围 / 菜单 ID / 子孙组织
│   ├── service.go       # 用户/角色/菜单/组织/岗位/租户 + 认证
│   ├── service_app.go   # 应用 / 密钥 / 绑定 / 开放接口
│   ├── handler.go       # RegisterRoutes（路径与 Java 逐字一致）
│   ├── apireg.go        # 接口元数据层（apiDecls 声明表 + OpenAPI 导出）
│   ├── seed.go          # 种子数据
│   ├── .env.example     # 配置模板
│   ├── data/            # SQLite 库（运行时，已 gitignore）
│   └── logs/
├── frontend/
│   ├── src/
│   │   ├── api/         # 接口层，统一走 utils/request.ts
│   │   ├── components/  # AdminLayout / AdminSidebar / CheckboxTree / BaseIcon
│   │   ├── layout/ router/ store/
│   │   ├── styles/      # tokens.css（勿手改）+ global.css + page.css
│   │   ├── types/       # api.d.ts
│   │   ├── utils/       # request.ts / auth.ts / error.ts / validation.ts
│   │   └── views/       # login / dashboard / system/* / account/* / api-docs
│   ├── scripts/generate-tokens.cjs
│   └── vite.config.ts   # proxy /api → localhost:8080
├── docs/                # project-guide / perms / 审计与设计文档
├── api-contract.md      # 接口契约（前后端唯一事实来源）
├── SKILLS.md            # 技能依赖清单
├── restart.sh           # 启停脚本（start/dev/be/fe/stop/status）
├── start.bat            # Windows 双击启动（绕开沙箱进程回收）
└── README.md
```

---

## 四、快速开始

### 环境要求

- Go 1.21+（本项目构建于 1.26）
- Node.js 18+ / pnpm（前端）

### 本地开发

```bash
# 1) 后端配置（可跳过，全部有默认值）
cp backend/.env.example backend/.env

# 2) 后端依赖
cd backend && go mod tidy

# 3) 前端依赖
cd ../frontend && pnpm install

# 4) 一键启动双端
cd .. && ./restart.sh            # 后台
./restart.sh dev                 # 前台（日志 tee 到屏幕，Ctrl+C 停止）
```

> ⚠️ **必须在自己的终端（Git Bash / PowerShell）里跑 `restart.sh`**。
> 在 IDE 的 AI 沙箱里执行，脚本一结束进程就被回收 —— 表现为日志里明明打印了
> "端口就绪"，下一秒端口就没了。这不是脚本坏了，是沙箱限制。
> Windows 用户可直接双击 `start.bat`（进程不属沙箱进程树，不会被回收）。

### 生产构建

```bash
cd backend
CGO_ENABLED=0 go build -mod=mod -o server.exe .
```

---

## 五、启动方式

| 命令 | 说明 |
|------|------|
| `./restart.sh` | 后台启动双端（默认） |
| `./restart.sh dev` | 前台启动，日志 tee 到屏幕 |
| `./restart.sh be` | 只重启后端 |
| `./restart.sh fe` | 只重启前端 |
| `./restart.sh stop` | 停止双端 |
| `./restart.sh status` | 查看状态 |
| `start.bat` | Windows 双击启动（推荐） |

| 服务 | 地址 |
|------|------|
| 后端健康检查 | http://localhost:8080/api/health |
| 前端 | http://localhost:5174 |

---

## 六、接口文档

启动后访问（需登录，权限 `system:api:view`）：

| 入口 | 说明 |
|------|------|
| `GET /api/api-registry` | 全量接口清单，含入参/出参**字段级**结构 |
| `GET /api/api-registry/openapi` | OpenAPI 3.1，供 AI 与外部工具直接消费 |
| 前端「接口文档」页 | `src/views/api-docs` |
| `api-contract.md` | 静态契约文档（机器可读版本的源头） |

> 元数据**不是手写维护**的：路由与权限码来自 gin 的真实注册结果，
> 入出参由反射从结构体实时展开。增删字段、改 json tag 会自动反映。
> 唯一需人工维护的是 `apiDecls` 里的 `Summary` 与 `apiFieldHints`（业务语义）。

---

## 七、配置说明

配置文件：`backend/.env`（模板 `.env.example`）。**所有项都可省略**，省略即回退默认值，默认值与 Java `application.yml` 一致。

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER_PORT` | `8080` | 改端口需同步改前端 `vite.config.ts` 的 proxy |
| `DB_URL` | `./data/vue-admin.db` | SQLite 库路径（是文件路径，不是 DSN） |
| `DB_MAX_OPEN_CONNS` | `4` | SQLite 单写者模型，连接数必须限制 |
| `DB_MAX_IDLE_CONNS` | `2` | 空闲连接数 |
| `DB_BUSY_TIMEOUT_MS` | `5000` | 写锁等待毫秒（0 = 立即报 SQLITE_BUSY） |
| `JWT_SECRET` | 见代码 | **生产必须替换**：`openssl rand -base64 32` |
| `JWT_ACCESS_EXPIRE_MINUTES` | `60` | access token 有效期 |
| `JWT_REFRESH_EXPIRE_DAYS` | `7` | refresh token 有效期 |
| `JWT_ISSUER` | `vue-admin` | 签发者 |
| `JWT_HEADER` | `Authorization` | 改它会导致前端取不到 token |
| `JWT_PREFIX` | `Bearer` | 同上 |
| `CORS_ORIGINS` | `*` | 生产应改为具体 origin |

---

## 八、默认账号

| 用户名 | 角色 | 数据范围 |
|--------|------|---------|
| `admin` | `super_admin` | ALL |
| `user_admin` | `user_admin` | ALL |
| `demo` | `common_user` | SELF_ONLY |

### 初始密码从哪来

**初始密码是随机的**（`seed.go` 用 `randomHex(12)` 生成），首次启动时打印到终端 / 日志：

```
[seed] 初始账户密码（admin / user_admin / demo 均为）：b6b790518033e692c60b1918
```

- 三个账号**共用**这一个初始密码
- 这是 `docs/2026-10-07-audit.md` **P3-18**（默认弱口令 `admin123`）的修复结果 —— `admin123` 已不再有效
- **该密码只打印这一次**，没记下来就删掉 `backend/data/vue-admin.db` 重新播种
- 登录后请立即修改密码

---

## 九、常见问题

### Q: 端口 8080 起不来 / 前端报 404

先确认占用者是不是本项目：

```bash
curl -s http://localhost:8080/api/health    # 应返回 {"code":0,...,"status":"ok"}
```

`netstat` 只能看到端口在监听，看不出是谁。若返回 HTML 而非 JSON，说明被别的项目占了。

### Q: 前端能起但接口全 404 / 返回 HTML

vite proxy 未命中会走 SPA fallback 返回 `index.html`（**HTTP 是 200 不是 404**）。
判据：`curl /api/health` 返回 HTML 即 proxy 没生效。检查 `vite.config.ts` 的 proxy 与前端端口（5174）。

### Q: 并发写报 `database is locked`

检查 `DB_BUSY_TIMEOUT_MS` 是否 > 0、WAL 是否生效（`data/` 下应有 `-wal` / `-shm` 文件）。
SQLite 是单写者模型，不要把 `DB_MAX_OPEN_CONNS` 调大。

### Q: 改了密码 / JWT_SECRET 后要重新登录

正常。改 `JWT_SECRET` 会让已签发 token 全部失效。

### Q: 沙箱里启的服务过一会儿就没了

用 `start.bat` 或在自己的终端里跑 `restart.sh`。详见第四节警告。

### Q: 侧边栏菜单少了 / 组织架构树只有根节点

树的 `children` 为空是**已知踩过的坑**：Go 的 `Children` 是值切片，
用「迭代 + 值拷贝进 roots」会让子树静默丢失。**必须自顶向下递归构建**。
验收要断言节点总数（菜单 37）与三层结构，不能只看 `code:0`。

---

## 十、拓展指南

### 新增一个后端接口

1. 在 `service.go`（或 `service_app.go`）加业务方法
2. 在 `handler.go` 注册路由，受保护端点挂 `RequirePerm("模块:资源:动作")`
3. **在 `apireg.go` 的 `apiDecls()` 登记**（Method/Path/Module/Perm/Auth/Summary/Req/Resp）
4. **同步更新 `api-contract.md`**
5. 若新增权限码：加进 `common.go` 的 `AuthPermsAll`、`seed.go` 的菜单、`docs/perms.md`

### 新增一张表

```go
// model.go
type SysXxx struct {
    ID        int64          `gorm:"primaryKey;column:id" json:"id"`
    Name      string         `gorm:"column:name;size:100" json:"name"`
    CreatedAt time.Time      `gorm:"column:created_at" json:"-"`
    UpdatedAt time.Time      `gorm:"column:updated_at" json:"-"`
    DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (SysXxx) TableName() string { return "wg_sys_xxx" }
```

约定：表名 `wg_` 前缀、软删除 `deleted_at`、密码字段 `json:"-"`、时间字段输出 `yyyy-MM-dd HH:mm:ss`。

### 新增一个前端页面

1. 读 `vue-generate-skill` 与 `vue-base-skill` 规范
2. 组件全部从 skill 包 `import`，**禁止手写 Element Plus 已有的组件**
3. 接口写在 `src/api/*.ts`，走 `utils/request.ts`
4. 样式用 `var(--color-*)` / `var(--space-*)`，**禁止裸色值、裸 px**
5. 侧边栏可见性由后端菜单树驱动，需在后端加菜单与权限

### 改主题

改 `frontend/scripts/generate-tokens.cjs` 后重新生成，**不要手改 `tokens.css`**。

---

## 十一、数据权限四档

角色的 `dataScope`：`ALL` / `DEPT_AND_BELOW` / `DEPT_ONLY` / `SELF_ONLY`，多角色取最宽。
端到端验证：种子场景下分别可见 10 / 9 / 4 / 1 条。详见 `api-contract.md` §6。

---

## 相关文档

| 文档 | 内容 |
|------|------|
| `SKILLS.md` | 技能依赖清单（迭代第一入口） |
| `api-contract.md` | 接口契约（前后端唯一事实来源） |
| `docs/perms.md` | 39 个权限码与菜单/按钮映射表 |
| `docs/2026-10-07-audit.md` | 安全与缺陷审计（P0–P3） |
| `docs/2026-10-07-api-registry-design.md` | 接口元数据层设计 |
| `docs/2026-10-07-next-iterations.md` | 后续迭代候选清单与排期 |
| `2026-10-07-技能合规审计.md` | 技能合规审计 |
