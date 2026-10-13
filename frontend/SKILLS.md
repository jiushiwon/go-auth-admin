# frontend/SKILLS.md — 前端技能依赖清单

> 本项目（`go-auth-admin/frontend/`）通过 14 个 `file:` 软链接软挂在 `D:/projects/wg-skills` 下的
> vue 技能包。本清单回答"前端 1:1 复刻 java-vue-admin 时，到底在哪些文件里、依赖了哪些技能"。

## 1. 关系类型

| 标记 | 含义 | 例子 |
|---|---|---|
| **生成** | 技能包在 `src/` 下生成真实组件，被项目 import | `vue-button-skill` → `BaseButton` |
| **规范** | 技能只提供 markdown 规范/约束文档，组件由项目自己实现 | `frontend-request-skill`（仅文档） |
| **依赖** | 软链接装在 node_modules，但项目**不直接 import**；通过全局注册 / 软链侧副作用生效 | `vue-card-skill`（main.ts 全局注册后被 vue-login-skill 的 LoginForm.vue 间接使用） |
| **基线** | 行为契约的**事实源**——改它必须同步改后端 + 文档 | 后端 OpenAPI（`http://localhost:8080/api/api-registry`） |

## 2. 前端 14 个 file: 依赖逐项清单

> 标 ✓ = 项目实际用；标 △ = 软链装但**没直接 import**（被其他技能间接消费或被 override）

| 技能 | 用在哪 | 形式 |
|---|---|---|
| `vue-button-skill` | `src/main.ts:36` 全局注册 `BaseButton`；`views/dashboard/index.vue`、`views/account/bind/index.vue` | 依赖（main.ts 全局注册后 LoginForm 内部用 `<base-button>`） |
| `vue-card-skill` | `src/main.ts:32` 全局注册 `BaseCard` | 依赖（同上） |
| `vue-checkbox-skill` | `src/main.ts:40` 全局注册 `BaseCheckbox`；`src/components/CheckboxTree.vue` | 依赖（全局）+ 直接（CheckboxTree） |
| `vue-dialog-skill` | `views/account/bind/index.vue`、`views/system/app/index.vue`、`views/system/post/index.vue` | 生成（直接 import `BaseDialog`） |
| `vue-dropdown-skill` | `src/components/AdminLayout.vue`（user menu dropdown） | 生成（直接 import `BaseDropdown`） |
| `vue-form-skill` | `src/main.ts:33-34` 全局注册 `BaseForm` / `BaseFormItem` | 依赖（同 button/card） |
| `vue-input-skill` | `src/main.ts:35` 全局注册 `BaseInput`；`views/account/bind/index.vue`、`views/login/VueAdminLogin.vue` | 依赖 + 生成 |
| `vue-layout-skill` | `src/components/AdminLayout.vue` △（本地 override 替代）；vite.config.ts alias 不引用 | △ 软链装但**被本地 AdminLayout.vue 整体替代**（README:67-75 解释了原因：原 AppLayout 硬编码 MenuItem 字符串） |
| `vue-login-skill` | `views/login/index.vue`、`views/login/flip/FlipLoginContainer.vue`、`views/login/_test/split/index.vue` | 生成（直接 import `LoginForm`） |
| `vue-select-skill` | `views/account/bind/index.vue`、`views/system/app/index.vue`、`views/system/menu/index.vue` | 生成 |
| `vue-switch-skill` | **无人 import** | △ 软链装但**未使用**（保留供未来表单开关） |
| `vue-table-skill` | `views/account/bind/index.vue`、`views/system/app/index.vue`、`views/system/post/index.vue` | 生成 |
| `vue-tag-skill` | `views/dashboard/index.vue`、`views/system/role/index.vue`、`views/system/user/index.vue` | 生成 |
| `vue-tree-skill` | `views/system/menu/index.vue`、`views/system/org/index.vue` | 生成 |
| `frontend-icon-skill` | `src/components/icons/BaseIcon.vue`（被 alias 接管，详见 vite.config.ts:17-20） | △ 软链装但**被本地 BaseIcon.vue 替代**（同 button-skill/BaseButton.vue 内部引用） |

### 2.1 诚实清单：未使用

| 技能 | 状态 | 决策 |
|---|---|---|
| `vue-switch-skill` | 软链装，**0 文件 import** | 保留在 dependencies；未来用户管理"启用/禁用"开关可加 |
| `vue-layout-skill` | 软链装，被本地 `src/components/AdminLayout.vue` **整体 override** | 不删（README:67-75 解释替代原因） |
| `frontend-icon-skill` | 软链装，被本地 `src/components/icons/BaseIcon.vue` 替代 | 同上（vite.config.ts:8-20 解释 fs/promises 副作用） |

## 3. 改动路由表（"改 X 前先加载哪个技能"）

| 改的文件 | 必读技能 | 必读原因 |
|---|---|---|
| `src/views/login/**` | `vue-login-skill`（templates/LoginForm.vue） | 登录页容器只是壳，**真正控件在技能里**（base-input/base-form/base-button） |
| `src/components/AdminLayout.vue` | `vue-layout-skill`（components/AppLayout.vue） | 改了 AdminLayout 后必须确认 override 理由（README:67-75）是否仍成立；不再成立就回切 |
| `src/components/icons/BaseIcon.vue` | `frontend-icon-skill`（src/index.ts） | 软链的 `src/index.ts` 顶层 `import 'node:fs/promises'`（CLI 工具用），**浏览器端不能 import**；本地 override 18 个 SVG |
| `src/components/AdminSidebar.vue` | `vue-base-skill` 通用规范 | 侧边栏图标一律用 `<base-icon name="...">`；18 个硬编码图标外的名字会被兜底为 "menu"（console 警告） |
| `src/utils/request/*.ts` | `frontend-request-skill`（仅 markdown） | 改 axios 拦截器前先看 skill 规范；401 自动 logout + 不实现 token refresh 队列（README:104 是当前约束） |
| `src/directives/permission.ts` | 后端 `SKILLS.md` + `docs/perms.md` | 改 `v-permission` 判定前先看 39 权限码全集；超管短路规则对齐 `backend/middleware.go` |
| `src/router/*.ts` | `vue-base-skill` 路由守卫 | 守卫靠 `localStorage.vue_admin_token`；不存在的 key 走 logout→dashboard 循环（**不要清 storage**） |
| 新增 `src/views/system/xxx/index.vue` | `vue-table-skill` + `vue-form-skill` + `vue-tree-skill` | CRUD 页面三件套：列表用 `<base-table>`、表单用 `<base-form>`、树用 `<base-tree>` |

## 4. ★ 前端红线速查

| 红线 | 触发场景 | 后果 |
|---|---|---|
| **base-* 必须全局注册** | 在 `main.ts` 之外用 `<base-input>` 等标签 | 浏览器把字面 self-closing 标签当未知 HTML 忽略，**登录页只有标题可见**（main.ts:14-23 解释） |
| **不要清 localStorage** | playwright / 自动化里 `localStorage.clear()` | 路由 guard 死循环（logout→dashboard）；保留 storage 或强制 `goto('/login')`（e2e 报告 P2） |
| **不要用 `import 'frontend-icon-skill'`** | 在 .vue 里直接 import 软链根 | 浏览器报错 "Module 'node:fs/promises' has been externalized"；只 import 具体子路径如 `frontend-icon-skill/components/BaseIcon.vue` |
| **`vue-login-skill` 必须在 optimizeDeps.exclude** | 漏了会让 vite optimizeDeps 只取 JS、`<style>` 被丢弃 | 登录页 frosted 毛玻璃完全失效（vite.config.ts:62-77 解释） |
| **`VITE_API_BASE_URL` 必须留空** | 在 `.env.development` 填 `/api` | URL 变 `/api/api/auth/login` → 401（user-agent 经验；README:78-80 已注明） |
| **token key 是 `vue_admin_token` 不是 `accessToken`** | 自动化用 `localStorage.getItem('accessToken')` | 永远 401（`src/utils/auth.ts:5` 事实源） |

## 5. ★ 跨端契约（与后端必须严格一致）

| 项 | 事实源 |
|---|---|
| 响应信封 `{code:0,message:"success",data}` | 后端 `backend/middleware.go` `OK()` + `backend/common.go` `AppError` |
| 分页字段 `list, total, page, pageSize`（**不是 items**） | 后端 `backend/service.go` `PageResponse` |
| 字段名 camelCase（`tenantId`, `orgId`, `createdAt`） | 后端 `backend/service.go` 所有 VO struct tag |
| 应用签名 7 段 HmacSHA256 | 后端 `backend/middleware.go` `AppSigMiddleware` + `api-contract.md` |
| Token 存 `localStorage['vue_admin_token']` | `src/utils/auth.ts:5`（前端事实源） |
| 39 权限码全集 | `docs/perms.md`（双向参照） |

## 6. 与后端 SKILLS.md 的关系

- 本文件（`frontend/SKILLS.md`）**与** `../SKILLS.md` **互补**，不重复：
  - 后端 SKILLS.md：Go 技能（`go-gin-init-skill` / `go-auth-module-skill`）+ 后端 8 红线 + 改动路由表
  - 本文件：14 个 vue-* 技能 + 前端 6 红线 + 跨端契约

## 7. 端到端验证（`docs/2026-10-13-e2e-result.md`）

5 项 e2e 全绿，验证本表"跨端契约"行的每一项都对得上：
1. health 200 + 完整信封
2. 真实浏览器登录 3 个 URL + 4 个 storage key
3. me 200 + 34 权限码
4. menus 37 节点
5. users 分页 keys=`['list','page','pageSize','total']` 严格对齐 Java
