# go-auth-admin 权限常量表

> `go-auth-module-skill` 强制交付物。权限码是**三段式** `模块:资源:动作`，
> 后端常量在 `backend/common.go` 的 `AuthPermsAll`，种子菜单在 `backend/seed.go`，
> 端点挂载在 `backend/handler.go`（`RequirePerm`），元数据登记在 `backend/apireg.go`。

---

## 一、数量口径（别记错）

| 口径 | 数量 | 说明 |
|------|------|------|
| **权限常量总数** | **39** | `AuthPermsAll`，鉴权时 super_admin 短路装载的全集 |
| 种子菜单总数 | 37 | 3 目录 + 8 页面 + 26 按钮 |
| **菜单派生权限** | **34** | `me()` 返回的数量（`FindPermissionsByUser` DB 派生） |
| 无菜单权限 | 5 | 4 个 `system:tenant:*` + 1 个 `system:api:view` |

> ⚠️ **技能里写的是 38，实际是 39** —— 2026-10-07 新增「接口注册中心」时加了
> `system:api:view`，技能未同步。本文以代码为准，技能已同步修正（见 `SKILLS.md` 第十节）。
>
> 39 = 34（菜单派生）+ 5（无菜单）。

---

## 二、用户管理（`system:user:*`，7 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `system:user:list` | 页面 | 用户管理 | `/system/user` | `GET /api/users`、`GET /api/users/:id`、`GET /api/users/:id/roles`、`GET /api/users/:id/posts` |
| `system:user:create` | 按钮 | 新增用户 | — | `POST /api/users` |
| `system:user:edit` | 按钮 | 编辑用户 | — | `PUT /api/users/:id` |
| `system:user:delete` | 按钮 | 删除用户 | — | `DELETE /api/users/:id` |
| `system:user:reset-pwd` | 按钮 | 重置密码 | — | `PUT /api/users/:id/password` |
| `system:user:assign-role` | 按钮 | 分配角色 | — | `PUT /api/users/:id/roles` |
| `system:user:assign-post` | 按钮 | 分配岗位 | — | `PUT /api/users/:id/posts` |

## 三、角色管理（`system:role:*`，5 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `system:role:list` | 页面 | 角色管理 | `/system/role` | `GET /api/roles`、`GET /api/roles/:id`、`GET /api/roles/:id/menus` |
| `system:role:create` | 按钮 | 新增角色 | — | `POST /api/roles` |
| `system:role:edit` | 按钮 | 编辑角色 | — | `PUT /api/roles/:id` |
| `system:role:delete` | 按钮 | 删除角色 | — | `DELETE /api/roles/:id` |
| `system:role:assign-menu` | 按钮 | 分配菜单 | — | `PUT /api/roles/:id/menus` |

## 四、菜单管理（`system:menu:*`，4 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `system:menu:list` | 页面 | 菜单管理 | `/system/menu` | `GET /api/menus` |
| `system:menu:create` | 按钮 | 新增菜单 | — | `POST /api/menus` |
| `system:menu:edit` | 按钮 | 编辑菜单 | — | `PUT /api/menus/:id` |
| `system:menu:delete` | 按钮 | 删除菜单 | — | `DELETE /api/menus/:id` |

## 五、组织管理（`system:org:*`，4 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `system:org:list` | 页面 | 组织管理 | `/system/org` | `GET /api/orgs` |
| `system:org:create` | 按钮 | 新增组织 | — | `POST /api/orgs` |
| `system:org:edit` | 按钮 | 编辑组织 | — | `PUT /api/orgs/:id` |
| `system:org:delete` | 按钮 | 删除组织 | — | `DELETE /api/orgs/:id` |

## 六、岗位管理（`system:post:*`，4 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `system:post:list` | 页面 | 岗位管理 | `/system/post` | `GET /api/posts` |
| `system:post:create` | 按钮 | 新增岗位 | — | `POST /api/posts` |
| `system:post:edit` | 按钮 | 编辑岗位 | — | `PUT /api/posts/:id` |
| `system:post:delete` | 按钮 | 删除岗位 | — | `DELETE /api/posts/:id` |

## 七、应用管理（`system:app:*`，5 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `system:app:list` | 页面 | 应用管理 | `/system/app` | `GET /api/apps`、`GET /api/apps/:id` |
| `system:app:create` | 按钮 | 新增应用 | — | `POST /api/apps` |
| `system:app:edit` | 按钮 | 编辑应用 | — | `PUT /api/apps/:id` |
| `system:app:delete` | 按钮 | 删除应用 | — | `DELETE /api/apps/:id` |
| `system:app:key-manage` | 按钮 | 密钥管理 | — | `GET/POST /api/apps/:id/keys`、`DELETE /api/apps/:id/keys/:keyId` |

## 八、账户中心（`account:bind:*`，4 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `account:bind:list` | 页面 | 应用绑定 | `/account/bind` | `GET /api/bind` |
| `account:bind:create` | 按钮 | 发起绑定 | — | `POST /api/bind/code`、`POST /api/bind/confirm` |
| `account:bind:delete` | 按钮 | 删除绑定 | — | `DELETE /api/bind/:id` |
| `account:bind:set-default` | 按钮 | 设为默认 | — | `PUT /api/bind/:id/default` |

## 九、仪表盘（`dashboard:*`，1 个）

| 权限码 | 类型 | 菜单名 | 路由 | 端点 |
|--------|------|--------|------|------|
| `dashboard:home:view` | 页面 | 仪表盘 | `/dashboard` | 无后端端点（纯前端路由） |

## 十、★ 无菜单权限（5 个）

这些权限**没有对应种子菜单**，因此无法在「角色分配菜单」弹窗里勾选。
只有 `super_admin` 能通过**常量短路**拿到，普通角色永远无法被授予。

| 权限码 | 端点 | 状态 |
|--------|------|------|
| `system:tenant:list` | `GET /api/tenants` | 后端完整实现，前端无页面（P3-1 已知缺口） |
| `system:tenant:create` | `POST /api/tenants` | 同上 |
| `system:tenant:edit` | `PUT /api/tenants/:id` | 同上 |
| `system:tenant:delete` | `DELETE /api/tenants/:id` | 同上 |
| `system:api:view` | `GET /api/api-registry`、`GET /api/api-registry/openapi` | 前端有「接口文档」页，但菜单未纳管 |

> 若要让普通角色可用，需在 `seed.go` 补建对应菜单并绑定角色。属待办，见 `docs/2026-10-07-next-iterations.md`。

---

## 十一、★ 权限双路径（禁止合并）

| 场景 | 走哪条 | 数量 | 对应 Java |
|------|--------|------|-----------|
| `GET /api/auth/me` 返回的 `permissions` | **DB 派生**（`FindPermissionsByUser`） | **34** | `AuthorityLoader.permissions()` |
| 鉴权中间件 `RequirePerm` 判定 | DB 派生 ∪ **常量全集**（super_admin 短路） | **39** | `JwtAuthenticationFilter` 反射装载 `AuthPerms` |

**原因**：Java 侧就是这两条路径，本项目必须逐字对齐。
**不要**把 `me()` 也短路成 39 —— 否则前端拿到的权限数组与 Java 不一致。

---

## 十二、种子角色绑定

| 角色码 | 名称 | 数据范围 | 菜单绑定 |
|--------|------|---------|---------|
| `super_admin` | 超级管理员 | `ALL` | **全部 37 个菜单** |
| `user_admin` | 用户管理员 | `ALL` | 系统管理相关（用户/角色/菜单/组织/岗位） |
| `common_user` | 普通用户 | `SELF_ONLY` | 仪表盘 + 应用绑定 |

种子账户：`admin`（super_admin）、`user_admin`（user_admin）、`demo`（common_user）。

> ⚠️ **初始密码由种子随机生成并打印到日志**（`seed.go` 的 `randomHex(12)`），
> 三个账号共用。**`admin123` 已不再有效**（P3-18 弱口令修复，见 `docs/2026-10-07-audit.md`）。
> 详见 `docs/project-guide.md` 第八节。

---

## 相关文档

| 文档 | 内容 |
|------|------|
| `api-contract.md` | 全量接口契约 |
| `SKILLS.md` | 技能依赖清单与红线 |
| `docs/2026-10-07-audit.md` | P3-1 租户权限缺口等审计项 |
