# go-auth-admin 接口契约

> 本文档是 `go-gin-init-skill` 与 `go-auth-module-skill` 共同要求的**强制交付物**，
> 骨架取自 `vibeCoding/backend/go/go-gin-init-skill/references/api-contract-template.md`，
> 内容取自 `backend/apireg.go` 的 `apiDecls()` 全量声明表（**事实来源**）。
>
> **⚠️ 与模板的已知偏差（有意为之，勿"修正"）**：模板用 snake_case（`access_token` / `page_size`），
> 本项目一律 **camelCase**。原因是契约基线是 Java 版 java-vue-admin，前端未改动，
> 改成 snake_case 会立刻打穿前端。**有冲突时以 Java 基线为准。**
>
> 运行时可消费的机器版本：`GET /api/api-registry`（字段级元数据）、`GET /api/api-registry/openapi`（OpenAPI 3.1）。

---

## 1. 基础信息

| 项 | 值 |
|----|-----|
| Base URL（dev） | `http://localhost:8080/api` |
| API Prefix | `/api` |
| 鉴权方式 | `Authorization: Bearer {accessToken}` |
| 开放接口鉴权 | 应用签名，见 §4.11 |
| Content-Type | `application/json` |
| 字符编码 | UTF-8 |
| 时间格式 | `yyyy-MM-dd HH:mm:ss`（**不是** ISO 8601，与 Java 基线一致） |
| 分页默认值 | `page=1`，`pageSize=10`；上限 `pageSize <= 100` |
| 契约基线 | `D:\projects\java-vue-admin\backend`（逐字对齐目标） |

---

## 2. 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | integer | 是 | 业务状态码，`0` 成功，负数失败 |
| message | string | 是 | 人类可读描述 |
| data | any | 是 | 业务数据，空时为 `null` |

- HTTP 状态码：业务失败时同时返回对应 HTTP 状态（400/401/403/404/409），前端按信封 `code` 判断。
- 响应中禁止返回堆栈、SQL、内部错误信息。

### 2.1 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": { "list": [], "total": 0, "page": 1, "pageSize": 10 }
}
```

> **分页字段是 `list`，不是 `items` / `records`** —— `go-auth-module-skill` 契约红线 4。

### 2.2 裸数组响应

以下端点 `data` 直接是数组，**不是**分页对象（前端「勿取 res.list」）：

| 端点 | 说明 |
|------|------|
| `GET /api/auth/menus` | 当前用户菜单树 |
| `GET /api/menus` | 全量菜单树 |
| `GET /api/orgs` | 组织树 |
| `GET /api/posts` | 岗位列表 |
| `GET /api/tenants` | 租户列表 |
| `GET /api/bind` | 我的绑定列表 |
| `GET /api/roles/:id/menus`、`GET /api/users/:id/roles`、`GET /api/users/:id/posts` | ID 列表 |

---

## 3. 全局错误码

> 遵循 `vibeCoding/backend/shared/error-code-spec.md`（单一事实来源），
> 与 Java 侧 `BusinessException`（-1001~-1005）逐字对齐。

| 错误码 | HTTP | 含义 | 触发场景 |
|--------|------|------|----------|
| 0 | 200 | 成功 | 业务正常 |
| -1001 | 400 | 参数校验失败 | 必填缺失、格式错误、业务规则拒绝 |
| -1002 | 401 | 未授权 | Token 缺失/无效/过期、用户名或密码错误 |
| -1003 | 403 | 禁止访问 | 权限码不足 |
| -1004 | 404 | 资源不存在 | 用户/角色/菜单等未找到 |
| -1005 | 409 | 资源冲突 | 用户名已存在、旧密码错误、编码重复 |
| -2000 | 500 | 系统异常 | 未捕获异常 |

前端 `src/utils/error.ts` 的 `ErrorCode` 与本表一一对应。

---

## 4. 接口清单

共 60 个端点。权限码为空表示不校验（仅需登录或完全公开）。

### 4.1 系统

| 方法 | 路径 | 权限码 | 鉴权 | 说明 |
|------|------|--------|------|------|
| GET | `/api/health` | — | 公开 | 健康检查，返回 `{"status":"ok"}` |
| GET | `/api/api-registry` | `system:api:view` | JWT | 全量接口清单（含入参/出参字段级结构） |
| GET | `/api/api-registry/openapi` | `system:api:view` | JWT | OpenAPI 3.1 文档，供 AI 与外部工具消费 |

### 4.2 认证与账户

| 方法 | 路径 | 权限码 | 鉴权 | 说明 |
|------|------|--------|------|------|
| POST | `/api/auth/login` | — | 公开 | 账号密码登录，返回 `accessToken` / `refreshToken` |
| POST | `/api/auth/refresh` | — | 公开 | 用 `refreshToken` 换发新令牌 |
| POST | `/api/auth/logout` | — | JWT | 登出（无状态，前端清 token 即视为登出） |
| GET | `/api/auth/me` | — | JWT | 当前用户信息，含 `roles` / `permissions` |
| PUT | `/api/auth/password` | — | JWT | 当前用户修改自己的密码 |
| GET | `/api/auth/menus` | — | JWT | 当前用户可见菜单树（裸数组） |

**`POST /api/auth/login` 请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**响应 data：`LoginResponseVO`**

| 字段 | 类型 | 说明 |
|------|------|------|
| accessToken | string | 访问令牌，有效期 60 分钟 |
| refreshToken | string | 刷新令牌，有效期 7 天 |
| tokenType | string | 固定 `Bearer` |
| expiresIn | int64 | 过期秒数（3600） |

**`POST /api/auth/refresh` 请求体**：`{ "refreshToken": "string" }`，响应同 login。

**`PUT /api/auth/password` 请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| oldPassword | string | 是 | 原密码 |
| newPassword | string | 是 | 新密码 |

**`GET /api/auth/me` 响应 data：`UserInfoVO`**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 用户 ID |
| username | string | 用户名 |
| nickname | string | 昵称 |
| email | string | 邮箱 |
| phone | string | 手机号 |
| avatar | string | 头像 |
| tenantId | int64 | 租户 ID |
| orgId | int64 / null | 组织 ID |
| status | int | 0=禁用，1=启用 |
| roles | string[] | 角色码（如 `super_admin`） |
| permissions | string[] | 权限码，DB 派生（超管为 34 个） |

> ⚠️ `me()` 的权限是 **DB 派生**（34 个），与鉴权时 super_admin 短路装载的**常量全集**（38 个）
> **故意不同**，禁止合并 —— `go-auth-module-skill` 契约红线 1。

**可能错误码**：`-1002`（用户名或密码错误）、`-1005`（旧密码错误）

### 4.3 用户

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/users` | `system:user:list` | 用户分页列表 |
| GET | `/api/users/:id` | `system:user:list` | 用户详情（含角色与岗位） |
| POST | `/api/users` | `system:user:create` | 创建用户 |
| PUT | `/api/users/:id` | `system:user:edit` | 更新用户（不可改 username，不可改密码） |
| DELETE | `/api/users/:id` | `system:user:delete` | 删除用户（软删） |
| PUT | `/api/users/:id/roles` | `system:user:assign-role` | 分配角色 |
| PUT | `/api/users/:id/posts` | `system:user:assign-post` | 分配岗位 |
| PUT | `/api/users/:id/password` | `system:user:reset-pwd` | 管理员重置指定用户密码 |
| GET | `/api/users/:id/roles` | `system:user:list` | 已分配角色 ID 列表（裸数组） |
| GET | `/api/users/:id/posts` | `system:user:list` | 已分配岗位 ID 列表（裸数组） |

**`GET /api/users` 查询参数**：`page`、`pageSize`、`username`、`status`

**`POST /api/users` 请求体：`CreateUserRequest`**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，唯一 |
| password | string | 是 | 密码 |
| nickname | string | 否 | 昵称 |
| email | string | 否 | 邮箱 |
| phone | string | 否 | 手机号 |
| avatar | string | 否 | 头像 |
| tenantId | int64 | 否 | 租户 ID |
| orgId | int64 / null | 否 | 组织 ID |
| status | int | 否 | 0=禁用，1=启用 |
| roleIds | int64[] | 否 | 角色 ID 列表（全量覆盖，空数组=清空） |
| postIds | int64[] | 否 | 岗位 ID 列表（全量覆盖，空数组=清空） |

**`PUT /api/users/:id` 请求体**：`nickname`、`email`、`phone`、`avatar`、`orgId`、`status`
> ⚠️ **不接受 `password`** —— 改密码必须走 `system:user:reset-pwd` 端点，否则绕过权限校验。

**响应 data：`UserVO`**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 用户 ID |
| username | string | 用户名 |
| nickname / email / phone / avatar | string | 基本信息 |
| tenantId | int64 | 租户 ID |
| orgId | int64 / null | 组织 ID |
| orgName | string | 组织名称 |
| status | int | 0=禁用，1=启用 |
| createdAt / updatedAt | string | `yyyy-MM-dd HH:mm:ss` |
| roles | SimpleRefVO[] | 角色引用 |
| posts | SimpleRefVO[] | 岗位引用 |

**可能错误码**：`-1005`（用户名已存在）、`-1004`（用户不存在）、`-1001`（不能删除当前登录用户 / 内置管理员不可删除）

### 4.4 角色

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/roles` | `system:role:list` | 角色分页列表 |
| GET | `/api/roles/:id` | `system:role:list` | 角色详情 |
| POST | `/api/roles` | `system:role:create` | 创建角色（`code` 不可为 `super_admin`） |
| PUT | `/api/roles/:id` | `system:role:edit` | 更新角色 |
| DELETE | `/api/roles/:id` | `system:role:delete` | 删除角色 |
| PUT | `/api/roles/:id/menus` | `system:role:assign-menu` | 给角色分配菜单权限 |
| GET | `/api/roles/:id/menus` | `system:role:list` | 已分配菜单 ID 列表（裸数组） |

**`GET /api/roles` 查询参数**：`page`、`pageSize`、`keyword`

**`POST /api/roles` 请求体：`CreateRoleRequest`**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 角色名 |
| code | string | 角色码，**不可为 `super_admin`** |
| description | string | 描述 |
| dataScope | string | `ALL` / `SELF_ONLY` / `DEPT_ONLY` / `DEPT_AND_BELOW` |
| sortOrder | int | 排序 |
| status | int | 0=禁用，1=启用 |
| menuIds | int64[] | 菜单 ID 列表（全量覆盖） |

**响应 data：`RoleVO`**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 角色 ID |
| name / code / description | string | 基本信息 |
| dataScope | string | 四档数据范围 |
| sortOrder | int | 排序 |
| status | int | 0=禁用，1=启用 |
| menuCount | int64 | 已绑定菜单数 |
| createdAt | string | 创建时间 |

**可能错误码**：`-1005`（角色编码已存在 / 角色已分配给 N 个用户不可删除）、`-1001`（超级管理员角色不可删除）

### 4.5 菜单

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/menus` | `system:menu:list` | 菜单树（全量，含按钮级，裸数组） |
| POST | `/api/menus` | `system:menu:create` | 创建菜单 |
| PUT | `/api/menus/:id` | `system:menu:edit` | 更新菜单 |
| DELETE | `/api/menus/:id` | `system:menu:delete` | 删除菜单（有子菜单时拒绝） |

**请求体字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| parentId | int64 / null | 父菜单 ID |
| name | string | 菜单名 |
| path | string | 路由路径 |
| component | string | 组件路径 |
| menuType | string | `M`=目录 / `C`=菜单 / `F`=按钮 |
| icon | string | 图标名（kebab-case） |
| permission | string | 权限码，**目录留空（存 NULL）** |
| sortOrder | int | 排序 |
| visible | int | 0=隐藏，1=显示 |
| status | int | 0=禁用，1=启用 |

**响应 data：`MenuNodeVO`**（含递归 `children`）

> ⚠️ 目录的 `permission` 序列化为 `null`，**不是** `""`。
> 装载 SQL 用 `IS NOT NULL`，空串会被当成有效权限 —— `go-auth-module-skill` 契约红线 2。

### 4.6 组织

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/orgs` | `system:org:list` | 组织树（裸数组） |
| POST | `/api/orgs` | `system:org:create` | 创建组织 |
| PUT | `/api/orgs/:id` | `system:org:edit` | 更新组织（禁止把父级改成自己或子孙） |
| DELETE | `/api/orgs/:id` | `system:org:delete` | 删除组织（有子组织时拒绝） |

**请求体：`CreateOrgRequest`**：`parentId`、`name`、`sortOrder`、`leaderUserId`、`phone`、`email`、`status`、`tenantId`

**响应 data：`OrgNodeVO`**（含递归 `children`）：`id`、`parentId`、`name`、`sortOrder`、`leaderUserId`、`phone`、`email`、`status`、`children`

### 4.7 岗位

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/posts` | `system:post:list` | 岗位列表（裸数组，不分页） |
| POST | `/api/posts` | `system:post:create` | 创建岗位 |
| PUT | `/api/posts/:id` | `system:post:edit` | 更新岗位 |
| DELETE | `/api/posts/:id` | `system:post:delete` | 删除岗位 |

**请求体**：`name`、`code`、`sortOrder`、`status`、`tenantId`
**响应 data：`PostVO`**：`id`、`name`、`code`、`tenantId`、`sortOrder`、`status`、`createdAt`

### 4.8 租户（后端保留，前端暂无入口）

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/tenants` | `system:tenant:list` | 租户列表（裸数组） |
| POST | `/api/tenants` | `system:tenant:create` | 创建租户 |
| PUT | `/api/tenants/:id` | `system:tenant:edit` | 更新租户 |
| DELETE | `/api/tenants/:id` | `system:tenant:delete` | 删除租户 |

> 注：种子数据**未创建租户菜单**，因此非超管角色无法被授予这组权限（无菜单可勾选）。属已知缺口，见 `docs/2026-10-07-audit.md` P3-1。

### 4.9 应用

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/apps` | `system:app:list` | 应用分页列表 |
| GET | `/api/apps/:id` | `system:app:list` | 应用详情 |
| POST | `/api/apps` | `system:app:create` | 创建应用 |
| PUT | `/api/apps/:id` | `system:app:edit` | 更新应用 |
| DELETE | `/api/apps/:id` | `system:app:delete` | 删除应用 |
| GET | `/api/apps/:id/keys` | `system:app:key-manage` | 应用密钥分页列表 |
| POST | `/api/apps/:id/keys` | `system:app:key-manage` | 创建密钥（**仅此一次返回明文 apiSecret**） |
| DELETE | `/api/apps/:id/keys/:keyId` | `system:app:key-manage` | 吊销密钥 |

**`POST /api/apps` 请求体**：`appName`、`description`、`logo`、`callbackUrl`
**响应 data：`AppVO`**：`id`、`appName`、`appKey`、`description`、`logo`、`callbackUrl`、`status`、`createdAt`

**`POST /api/apps/:id/keys` 请求体**：`{ "keyName": "string" }`
**响应 data：`CreatedAppKeyVO`**：`id`、`keyName`、`apiKey`、`apiSecret`、`status`、`createdAt`

> ⚠️ **`apiSecret` 仅此一次返回**，列表/详情永不返回（`AppKeyVO` 无该字段）—— `go-auth-module-skill` 红线 5。

### 4.10 应用绑定

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/api/bind` | `account:bind:list` | 我的绑定列表（裸数组） |
| POST | `/api/bind/code` | `account:bind:create` | 生成绑定码（返回 `qrContent` 与 `expireAt`） |
| POST | `/api/bind/confirm` | `account:bind:create` | 确认绑定 |
| DELETE | `/api/bind/:id` | `account:bind:delete` | 解除绑定 |
| PUT | `/api/bind/:id/default` | `account:bind:set-default` | 设为默认绑定 |

**`POST /api/bind/code` 请求体**：`{ "appId": 0 }`
**响应 data：`BindCodeVO`**：`code`、`direction`、`expireSeconds`、`qrContent`、`expireAt`

**`POST /api/bind/confirm` 请求体**：`{ "code": "string" }`
> ⚠️ **只接收 `code`**，身份一律取自登录态，禁止接收用户名/密码（旧版越权漏洞）—— `go-auth-module-skill` 红线 7。

**响应 data：`BindVO`**：`id`、`appId`、`appUserId`、`appUserName`、`bindType`、`isDefault`、`bindAt`、`appKey`、`appName`、`logo`

> 绑定码 5 分钟有效，**一次性消费**。

### 4.11 开放接口（应用签名鉴权）

**签名方案**（`AuthAppSig`）：

**待签串是 7 段**（`\n` 分隔，顺序不可变）：

```
payload = apiKey \n ts \n nonce \n METHOD \n path \n rawQuery \n sha256Hex(body)
签名    = HmacSHA256(apiSecret, payload).hex()
```

| 段 | 取值 | 说明 |
|----|------|------|
| apiKey | `AppKey.apiKey` | 与 `X-App-Key` 头一致 |
| ts | `X-Timestamp` | 秒级时间戳 |
| nonce | `X-Nonce` | 随机串 |
| METHOD | 大写，如 `GET` | — |
| path | 如 `/api/open/userinfo` | **不含** query |
| rawQuery | 如 `appUserId=demo` | **不含**前导 `?`；无 query 时为空串 |
| sha256Hex(body) | 请求体十六进制摘要 | 无 body 时为 `sha256Hex("")` = `e3b0c442...b855` |

> ⚠️ **query 与 body 必须参与签名**（P0-4 修复）。只签 path 时，
> 把 `?appUserId=X` 换成 `?appUserId=Y` 仍会验签通过，POST 请求体也可被整体替换 —— 签名形同虚设。

| 请求头 | 说明 |
|--------|------|
| `X-App-Key` | **`AppKey.apiKey`（`ak_` 前缀），不是 `App.appKey`（`app_` 前缀）** |
| `X-Timestamp` | 秒级时间戳，与服务时间偏差 ≤ 300 秒 |
| `X-Nonce` | 随机串，防重放 |
| `X-Signature` | 上述 HMAC 签名（常量时间比较） |

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/open/bind/apply` | 应用侧申请生成绑定码 |
| POST | `/api/open/bind/claim` | 应用侧凭绑定码完成绑定 |
| GET | `/api/open/userinfo` | 按 `appUserId` 查询绑定用户信息（`?appUserId=xxx`） |

**错误响应**：签名校验失败返回 `-1002` / HTTP 401。

---

## 5. 公共数据模型

### 5.1 `PageResponse<T>`

| 字段 | 类型 | 说明 |
|------|------|------|
| list | T[] | 数据列表 |
| total | int64 | 总数 |
| page | int | 当前页码 |
| pageSize | int | 每页条数 |

### 5.2 `SimpleRefVO`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | ID |
| name | string | 名称 |
| code | string | 编码 |

### 5.3 `OpenUserInfoVO`

| 字段 | 类型 | 说明 |
|------|------|------|
| appUserId | string | 用户在应用侧的外部 ID |
| bound | bool | 是否已绑定 |
| bindingId | int64 | 绑定记录 ID（未绑定时省略） |
| userId / username / nickname / avatar / email / phone | — | 绑定用户信息（未绑定时省略） |

---

## 6. 数据权限四档

角色的 `dataScope` 决定列表可见范围，多角色取**最宽**一档：

| 档位 | 含义 | SQL 约束 |
|------|------|---------|
| `ALL` | 全部 | 不加过滤 |
| `DEPT_AND_BELOW` | 本部门及下级 | `org_id IN (本部门 + 全部子孙)` |
| `DEPT_ONLY` | 本部门 | `org_id = 当前部门` |
| `SELF_ONLY` | 仅本人 | `id = 当前用户` |

> 已端到端验证：种子场景下四档分别可见 10 / 9 / 4 / 1 条。
> 验证方法见 `SKILLS.md` 与技能 `references/seed.md`。

---

## 7. 契约变更约定

1. 后端新增/修改接口 → **先改本文档**，再改代码，最后同步 `apireg.go` 的 `apiDecls`。
2. 前端新增接口 → 按本文档定义路径/字段/code，写在 `src/api/*.ts`。
3. 权限码变更 → 同步 `docs/perms.md` 与种子数据。
4. 冲突仲裁：Java 基线 > Go 模板 > 本文档其余约定。

---

## 相关文档

| 文档 | 内容 |
|------|------|
| `SKILLS.md` | 技能依赖清单（迭代第一入口） |
| `docs/perms.md` | 38 个权限码与菜单/按钮映射表 |
| `docs/project-guide.md` | 项目指南（启动/配置/排障/拓展） |
