# go-auth-admin 技能依赖清单

> **本文件是 go-auth-admin 的「技能依赖地图」，是所有后续迭代的第一入口。**
>
> - 改任何代码前，**先按「八、改动路由表」加载对应技能**，再动手。
> - 本项目的核心约束是 **「与 java-vue-admin 的 Java 后端逐字对齐，前端不动」**，
>   因此契约的**唯一事实来源是 Java 侧**，不是 Go 侧。这一点决定了下面很多技能的"关系"是**规范**而非**生成**。
>
> 技能仓库：`D:\projects\wg-skills`（公开仓库 github.com/jiushiwon/wg-skills）
> 契约基线仓库：`D:\projects\java-vue-admin`（Java 原型，本项目是它的 Go 复刻）

---

## 一、为什么要有这份文档

2026-10-07 的审计（`2026-10-07-技能合规审计.md`）结论是：**功能层全部对齐，交付物、文档、中间件被整批漏掉**。

根因不是能力问题，是**流程问题** —— 当时走的是「读 Java 源码 → 手写 Go 等价实现」，把技能当**事后参考**而不是**生成模板**。结果对准 Java 的部分（端点 56/56、VO 字段、权限码 56/56、种子数据）做得极好，而技能自己规定的产物形态、文档、中间件整批漏掉。

这份文档的作用就是**把这个漏洞堵死**：任何一次迭代，先查表确认该加载哪个技能、该遵守哪条红线，而不是凭记忆或"我觉得应该这样"。

---

## 二、总览

```
                    契约基线（唯一事实来源）
        ┌───────────────────────────────────────────┐
        │  springboot-init-skill（Java 原型的行为）  │
        │  java-vue-admin/backend（逐字对齐目标）    │
        └──────────────────┬────────────────────────┘
                           │ 逐字对齐
        ┌──────────────────┴────────────────────────┐
        │                                           │
   ┌────▼─────────────┐                    ┌────────▼──────────┐
   │  后端（Go）       │                    │  前端（Vue3）      │
   │                  │                    │                   │
   │ go-gin-init      │   api-contract.md  │ vue-base-skill    │
   │  （骨架规范）     │◄──────────────────►│  （15 个 npm 包）  │
   │ go-auth-module   │                    │ vue-login-skill   │
   │  （RBAC+通用登录）│                    │ vue-theme-skill   │
   │ shared/* 规范层   │                    │ frontend-request  │
   │ sqlite-guide     │                    │ frontend-icon     │
   └──────────────────┘                    └───────────────────┘
```

**关系类型说明**（本表关键，别看错）：

| 关系 | 含义 | 迭代时该怎么做 |
|------|------|---------------|
| **生成** | 产物由该技能直接产出 | 重新跑技能即可同步升级 |
| **规范** | 产物不是它生成的，但**必须**符合它的要求 | 手改代码前先读它的红线 |
| **依赖** | 以 npm `file:` 形式真实引入 | 改组件前确认是否该改上游技能 |
| **基线** | 它是行为定义方，本项目是复刻方 | 有歧义时以基线为准 |

---

## 三、后端技能依赖

### 3.1 `go-gin-init-skill` —— 骨架规范　【关系：规范】

- 路径：`vibeCoding/backend/go/go-gin-init-skill/`
- 管辖：`common.go`（响应信封/错误码/JWT）、`middleware.go`（中间件链）、`config.go`（.env 加载）、`main.go`（装配）、`restart.sh`

| 强制项 | 现状 | 说明 |
|--------|------|------|
| `api-contract.md` | 由本文配套的补齐动作落地 | 强制交付物 |
| `docs/project-guide.md` | 同上 | 强制交付物 |
| 安全头中间件 | 同上 | **红线 6：强制开启** |
| `restart.sh` / `restart.bat` | 已有（用户风格） | 已落地 |
| `.env` + `.env.example` | `.env.example` 已有 | 红线 9 |
| gofmt | 待执行 | 收尾自检清单 |

**它的红线（不可绕过）**：不生成无契约骨架 / 不硬编码版本 / 不跳过环境探测 /
不强制装系统级数据库 / **不替用户提交 git** / 默认值必须安全（安全头强制开启）/ 注释文档全中文 /
契约必须与 `frontend-request-skill` 对齐 / `.env` 必须被 `godotenv` 真正加载。

### 3.2 `go-auth-module-skill` —— RBAC + 通用登录　【关系：规范】

- 路径：`vibeCoding/backend/go/go-auth-module-skill/`
- 管辖：`model.go`、`repo.go`、`service.go`、`service_app.go`、`handler.go`、`seed.go`

**★ 契约红线**（本 skill 最值钱的部分，全是踩过的坑，迭代时逐条对照）：

| # | 红线 | 为什么 |
|---|------|--------|
| 1 | **两条权限解析路径故意不同，禁止合并** | `me()` = DB 派生（34 个）；鉴权 super_admin 短路 = 常量全集（38 个）。对应 Java 的 `AuthorityLoader` 与 `JwtAuthenticationFilter`，合并即与 Java 不一致 |
| 2 | 目录菜单 `permission` 存 **NULL**，禁止存 `""` | 装载 SQL 是 `IS NOT NULL`，空串会被当成有效权限（曾导致 37 vs 34） |
| 3 | 开放接口 `X-App-Key` 用 **`AppKey.apiKey`** | 不是 `App.appKey`，用错即 401 |
| 4 | 分页字段必须是 **`list`** | 禁止 `items` / `records` |
| 5 | 部分端点返回**裸数组**（前端「勿取 res.list」） | `/api/posts`、`/api/bind`、`/api/menus`、`/api/auth/menus`、`/api/orgs`、`/api/tenants` |
| 6 | 种子权限串必须与 Java **逐字一致** | 38 常量 / 34 菜单权限 |
| 7 | `import` 写全 `github.com/glebarez/sqlite` | 少前缀会让 `go mod tidy` 误删依赖 |
| 8 | **树构建必须自顶向下递归** | 禁止「迭代 + 值拷贝进 roots」，否则子树静默丢失（曾丢 26 个按钮） |

**其他红线**：`apiSecret` 仅创建响应返回一次 / 绑定码一次性消费 /
绑定确认只收 `{ code }`（身份取自登录态）/ 软删除用 `deleted_at` / 密码 `json:"-"` / 中文注释 / 不替用户提交 git。

### 3.3 `backend/shared/` —— 跨语言规范层　【关系：规范】

- 路径：`vibeCoding/backend/shared/`

| 文件 | 约束内容 |
|------|---------|
| `response-envelope-spec.md` | `{ code, message, data }` |
| `error-code-spec.md` | -1001 校验 / -1002 未授权 / -1003 无权限 / -1004 不存在 / -1005 冲突 |
| `jwt-auth-spec.md` | `Authorization: Bearer {token}` |
| `pagination-spec.md` | `page` + `pageSize`，上限 100 |
| `api-contract-template.md` | 契约文档模板 |
| `module-output-spec.md` | 模块接口输出规范 |

> ⚠️ **与模板的已知偏差（有意为之，勿"修正"）**：
> `api-contract-template.md` 用 snake_case（`access_token` / `page_size`），
> 本项目**必须**用 camelCase（`accessToken` / `pageSize`）—— 因为契约基线是 Java，
> 前端未改动，改成 snake_case 会立刻打穿前端。**有冲突时以 Java 基线为准。**

### 3.4 `database/` 技能　【关系：规范】

| 技能 | 状态 | 落地情况 |
|------|------|---------|
| `sqlite-guide-skill` | 使用 | WAL + `busy_timeout=5000` + 连接池（4/2/1h），pragma 在迁移**之前**设置 |
| `database-design-skill` | 规范 | 表名 `wg_sys_*`、软删除 `deleted_at`、关联表命名 |
| `mysql-guide-skill` | 未使用 | 生产可换 MySQL，见 `.env.example` |
| 其余（pgsql/mongo/redis/kafka） | 未使用 | 本项目是同步 HTTP，无消息队列 |

---

## 四、契约基线技能（Java）

### 4.1 `springboot-init-skill`　【关系：基线】

- 路径：`vibeCoding/backend/java/springboot-init-skill/`
- 实际基线仓库：`D:\projects\java-vue-admin\backend\`

**为什么一个 Go 项目要依赖 Java 技能**：本项目的验收标准不是"跑得通"，而是
**"前端无法区分后端是 Go 还是 Java"**。所以 Go 侧的每一处行为都要回溯到 Java 实现：

| 对齐项 | 核对方法 | 结果 |
|--------|---------|------|
| 端点路径 | 机器解析 `@RequestMapping` + `@XxxMapping` vs gin 路由 | 56 = 56 逐字一致 |
| 端点权限码 | `@PreAuthorize` vs `RequirePerm` | 56 = 56 逐一对齐 |
| VO/DTO 字段 | 39 个 Java DTO vs Go VO | 逐字段一致 |
| 权限常量 | `AuthPerms.java` vs `AuthPermsAll` | 38 = 38 逐字一致 |
| 种子数据 | `DataInitializer.java` vs `seed.go` | 用户/角色/组织/岗位/菜单 全部一致 |

> **配对陷阱（已踩两次）**：Java 注解顺序是 `@Operation → @PreAuthorize → @XxxMapping`。
> 按行线性配对会**系统性错位一格**（GET 拿到 create、POST 拿到 edit）。
> 必须按**方法签名切块**再提取，否则审计结论不可信。

---

## 五、前端技能依赖

> **前端当前状态：与 java-vue-admin 逐字节一致，本轮不改。**
> 下面记录的是它**受哪些技能约束**，供后续迭代时查阅。

### 5.1 `vue-base-skill` 及子技能　【关系：依赖（npm `file:`）】

- 路径：`vibeCoding/frontend/vue/vue-base-skill/`
- `package.json` 中 15 个 `file:D:/projects/wg-skills/...` 依赖，**已实际 import**：

| 包 | 引入的组件 | 用在 |
|----|-----------|------|
| `vue-card-skill` | `BaseCard` | 各管理页 / 登录页 |
| `vue-button-skill` | `BaseButton` | 各管理页 |
| `vue-table-skill` | `BaseTable`, `BasePaginated` | 用户/角色/应用页 |
| `vue-form-skill` | `BaseForm`, `BaseFormItem` | 表单弹窗 |
| `vue-input-skill` | `BaseInput` | 表单 |
| `vue-select-skill` | `BaseSelect` | 表单 |
| `vue-dialog-skill` | `BaseDialog` | 弹窗 |
| `vue-dropdown-skill` | `BaseDropdown` | `AdminLayout.vue` |
| `vue-checkbox-skill` | `BaseCheckbox` | `CheckboxTree.vue` |
| `vue-tag-skill` | `BaseTag` | 状态列 |
| `vue-switch-skill` | — | 已装，按需 |
| `vue-tree-skill` | — | 已装（树形独立实现于 `CheckboxTree.vue`） |
| `vue-layout-skill` | — | 已装（布局由 `components/AdminLayout.vue` 承担） |
| `vue-login-skill` | `LoginForm` | `views/login/index.vue` |
| `frontend-icon-skill` | 见 5.5 | `components/icons/BaseIcon.vue` |

> ⚠️ **已知问题（P3-5）**：`package.json` 用 `file:D:/projects/wg-skills/...` **本机绝对路径**，
> 换机器 / CI 无法安装。后续若要正规化，应改为相对路径或发布到 npm。

### 5.2 `vue-login-skill`　【关系：依赖】

- 路径：`vibeCoding/frontend/vue/vue-base-skill/vue-complex-skill/vue-login-skill/`
- 落地：`views/login/index.vue` 引入 `LoginForm`
- 另有手写备份页 `views/login/VueAdminLogin.vue`（用 `vue-form-skill` 自行组装）

### 5.3 `vue-theme-skill`　【关系：生成】

- 路径：`vibeCoding/frontend/vue/vue-theme-skill/`
- 落地：`src/styles/tokens.css` —— **由该技能的 generator 生成，文件头已注明「不要手改」**
- 当前主题：`copper`（HSL 50–950 色阶，命名与 `uniapp-theme-skill` 对齐）
- 生成器副本：`frontend/scripts/generate-tokens.cjs`

**红线**：禁止裸色值 / 禁止裸 px / 禁止 `calc()` 生成主题值 /
**禁止手改 `tokens.css` 命名**（必须走 generator）/ 禁止组件内定义全局变量 / 禁止混用 uniapp 与 vue 命名。

### 5.4 `frontend-request-skill`　【关系：规范】

- 路径：`vibeCoding/frontend/frontend-request-skill/`
- 落地：`src/utils/request.ts`（axios 实例 + 拦截器）

| 约定 | 落地情况 |
|------|---------|
| 响应信封 `{ code, message, data }` | ✅ `code === 0` 返回 `data` |
| 错误码表与后端 `api-contract.md` 对齐 | ✅ `src/utils/error.ts` 的 `ErrorCode` |
| `Authorization: Bearer {token}` | ✅ 请求拦截器注入 |
| 401 → 清 token + 跳登录 | ✅ 同时处理 HTTP 401 与信封 `-1002` |
| 统一 `get/post/put/del` | ✅ |

> ⚠️ **技能间冲突（记录在案，勿"修正"）**：
> `vue-generate-skill` 红线 1 写「❌ 不用 axios（必须 fetch）」，
> 但 `frontend-request-skill` 的通用前端规范明确 **fetch / axios 皆可**。
> 本项目前端沿用 java-vue-admin 原样（axios），因此**以 `frontend-request-skill` 为准**。

### 5.5 `frontend-icon-skill`　【关系：依赖（已改造）】

- 路径：`vibeCoding/frontend/frontend-icon-skill/`
- 落地：`src/components/icons/BaseIcon.vue`

**已做 v2 改造（重要，别改回去）**：原版 `import { generateIcon } from 'frontend-icon-skill'`
会触发该包 `src/index.ts` 顶层的 `node:fs/promises`（CLI 用），浏览器报
`Module 'node:fs/promises' has been externalized for browser compatibility`。
现版本把 SVG 模板**内联进组件**，零外部依赖。图标名统一 kebab-case（lucide 风格），详见 `frontend/ICONS.md`。

### 5.6 `vue-generate-skill`　【关系：规范（未直接执行）】

- 路径：`vibeCoding/frontend/vue/vue-base-skill/vue-generate-skill/`
- 说明：本项目前端是**从 java-vue-admin 整体拷贝**（未改动），并非由该技能生成。
  但它的结构规范（`src/api/`、`src/views/`、`src/store/` 等）仍是后续新增页面的依据。

> 它自身声明 **严格依赖 `frontend-request-skill`**。

### 5.7 `vue-style-skill`　【关系：未落地（技能本身是 stub）】

- 路径：`vibeCoding/frontend/vue/vue-style-skill/`
- **诚实说明**：该技能 `SKILL.md` 目前只有「定位 / 待实现 / 触发关键词」三节，
  「CSS 变量系统 / 主题配置 / 动画库 / 工具类 / 重置样式」**全部是未勾选的 TODO**。
  因此 `src/styles/global.css`、`page.css` 是项目自有的，**不算违反该技能**（技能本身尚无规范可遵守）。
  待该技能补全后再回来对齐。

### 5.8 审计类技能

| 技能 | 用途 | 何时用 |
|------|------|--------|
| `frontend-code-doctor` | Vue3 / TS 反模式扫描 | 提交前 |
| `frontend-style-harmonizer-skill` | 样式一致性（Design Token 命名跨框架对齐） | 改样式后 |

---

## 六、项目级编排技能

| 技能 | 位置 | 状态 |
|------|------|------|
| `vue-admin-skill` | 私有仓库 `git@github.com:jiushiwon/wg-project-kits.git` | 编排器（前端+后端+数据库），java-vue-admin 是它的 dogfooding 输出 |
| `universal-identity-bridge-skill` | 同上（Java 先行，Go/Python 预留） | 统一身份接入编排，本项目是 Go 侧的候选承载方 |

---

## 七、未使用清单（诚实清单）

| 技能 | 为什么不用 |
|------|-----------|
| `go-frame-init-skill` | 面向 Hertz / Fiber / Chi，本项目是 Gin |
| `go-ws-module-skill` | 无 WebSocket 需求 |
| `go-module/go-storage-module-skill` | 无文件存储需求 |
| `java-fast-skill` / `python-fast-skill` / `nodejs-init-skill` / `rust-backend-skill` | 语言不匹配 |
| `fastapi-*`（含 `fastapi-auth-module-skill` 等 13 个） | 语言不匹配 |
| `springboot-module/*`（12 个 Java 业务模块） | Java 侧原型未启用这些模块；本项目是 Go 复刻，不直接消费 |
| `nuxt-generate-skill` / `react-*` / `uniapp-*` | 前端体系不匹配（本项目 Vue3 SPA） |
| `html-frontend-template` / `vue-tui-skill` / `electron-vue-init-skill` | 场景不匹配 |
| `super-deploy-skills`（7 个） | 尚未进入部署阶段 |
| `video/*`、`others/*` | 领域无关 |

---

## 八、★ 改动路由表（改之前先查这里）

| 你要改的东西 | 先加载 | 关键约束 |
|-------------|--------|---------|
| `backend/handler.go`（路由/权限码） | `go-auth-module-skill` + Java 控制器 | 路径与权限码必须与 Java 逐字一致；同步登记 `apireg.go` 的 `apiDecls` |
| `backend/apireg.go` | `go-gin-init-skill` | 路由与权限码取自 gin 真实注册结果，入出参靠反射，不要手写字段清单 |
| `backend/service*.go`（业务逻辑） | `go-auth-module-skill` | 业务规则 13 条（禁删自己/禁删 admin/超管角色不可删…）不得漏 |
| 树结构相关（菜单/组织） | `go-auth-module-skill` 契约红线 8 | **必须递归构建**，禁止迭代+值拷贝 |
| `backend/common.go`（信封/错误码/JWT） | `backend/shared/*` | 与 `frontend-request-skill` 的 `ERROR_CODE_MAP` 对齐 |
| `backend/middleware.go` | `go-gin-init-skill` 红线 6 | 中间件链顺序：安全头 → 日志 → CORS → 应用签名 → JWT → RequirePerm |
| `backend/model.go` / 建表 | `database-design-skill` + `sqlite-guide-skill` | `wg_sys_*` 前缀、`deleted_at` 软删 |
| `backend/seed.go` | `go-auth-module-skill` 红线 6 | 权限串与 Java 逐字一致；目录 permission 存 NULL |
| 后端新增接口后 | `go-auth-module-skill` 强制交付物 | 必须同步 `api-contract.md` |
| `frontend/src/api/*.ts` | `frontend-request-skill` | 按 `api-contract.md` 定义路径/字段/code |
| `frontend/src/styles/tokens.css` | `vue-theme-skill` | **禁止手改**，走 `scripts/generate-tokens.cjs` |
| 新增前端页面 | `vue-generate-skill` + `vue-base-skill` | 组件全部从 skill 引入，禁止手写 Element Plus 已有组件 |
| `frontend/src/components/icons/BaseIcon.vue` | `frontend-icon-skill` | 保持内联 SVG，别改回外部 import |
| 提交前 | `frontend-code-doctor` | 扫 Vue3 / TS 反模式 |

---

## 九、★ 红线速查（一次记全）

**后端**
1. 分页用 `list`，不用 `items` / `records`
2. 目录菜单 `permission` 存 NULL，不存 `""`
3. 开放接口 `X-App-Key` = `AppKey.apiKey`
4. `me()` 与鉴权上下文的权限**不合并**
5. 树构建**递归**，不迭代+值拷贝
6. `apiSecret` 仅创建响应返回一次
7. 绑定确认只收 `{ code }`
8. 密码 `json:"-"`，任何 VO 不得带密码
9. 软删除 `deleted_at`，查询必带 `deleted_at IS NULL`
10. 注释文档全中文；**不替用户提交 git**

**前端**
1. 禁止裸色值 / 裸 px，用 `var(--color-*)` / `var(--space-*)`
2. 禁止手改 `tokens.css`
3. 接口放 `src/api/`，统一走 `utils/request.ts`
4. 禁止在组件里写 401 跳转（统一由请求层处理）
5. 禁止在组件里直接 `localStorage`

**跨端**
1. 契约变更必须前后端同步更新 `api-contract.md`
2. Java 基线与 Go 模板冲突时，**以 Java 基线为准**（camelCase）

---

## 十、同步约定

- wg-skills 升级后，前端 15 个 `file:` 依赖需重新安装（`pnpm install`）。
- 后端 `go-gin-init-skill` / `go-auth-module-skill` 若新增红线，需回到本文第三节更新对应条目。
- `api-contract.md` 是**前后端唯一事实来源**，任何接口变更必须先改它再改代码。
- 本文变更需同步 `README.md` 的「技能依赖」小节。

---

## 相关文档

| 文档 | 内容 |
|------|------|
| `api-contract.md` | 全量接口契约（前后端唯一事实来源） |
| `docs/project-guide.md` | 项目指南（启动/配置/排障/拓展） |
| `docs/perms.md` | 38 个权限码与菜单/按钮映射表 |
| `docs/2026-10-07-audit.md` | 安全与缺陷审计（P0–P3 分级） |
| `docs/2026-10-07-api-registry-design.md` | 接口元数据层设计 |
| `docs/2026-10-07-next-iterations.md` | 后续迭代候选清单（分批排期） |
| `2026-10-07-技能合规审计.md` | 技能合规审计（本文的由来） |
