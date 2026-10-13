# 接口元数据层（API Registry）设计方案

- 日期：2026-10-07
- 目标：**给 AI 看的接口清单**，让 AI 能直接理解并接上本系统
- 现状基线：56 个端点 / 10 个模块 / 37 个权限码 / 39 个 Request-VO 结构体

---

## 1. 先说结论

**不要用 md 表格手工维护，也不要全用表单录入。** 两者都要求人肉誊抄，而誊抄的内容已经存在于代码里，必然漂移。

推荐 **④ 反射自动提取 + 表存储 + 多格式导出**，核心是把元数据拆成两层：

```
┌─────────────── 机器层（自动生成，永不过期）───────────────┐
│  Gin 路由表 + 结构体 json tag  ──反射──▶  JSON Schema      │
│  产出：method / path / permCode / 入参结构 / 出参结构       │
└──────────────────────────┬──────────────────────────────┘
                           │ upsert（只写机器字段）
                           ▼
                 ┌────────────────────┐
                 │ wg_sys_api_endpoint │ ◀── 人工层（表单维护）
                 └────────────────────┘     中文名/业务语义/AI提示/示例
                           │ export
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
   openapi.json        endpoints.json     endpoints.md
   （给 AI，首选）      （给程序）         （给人粗略浏览）
```

**为什么 AI 消费首选 OpenAPI**：AI 训练语料中 OpenAPI / Swagger 出现频次极高，读到
`paths./api/users/{id}.put.requestBody.schema` 立即理解，无需额外学习成本。
而自定义表格格式需要 AI 先理解你的字段约定，是多出来的推理损耗。

---

## 2. 现状盘点（决定方案可行性的关键事实）

| 项目 | 数量 | 说明 |
|---|---|---|
| 端点 | 56 | `auth` 7 / `users` 10 / `roles` 7 / `menus` 4 / `orgs` 4 / `posts` 4 / `tenants` 4 / `apps` 8 / `bind` 5 / `open` 3 |
| 权限码 | 37 | 已由 `RequirePerm(...)` 标注，`AuthPermsAll` 定义 38 个 |
| Request/VO 结构体 | 39 | **json tag 齐全**，含 `*int64` 可选、嵌套 `[]SimpleRefVO` |
| 统一信封 | — | `{code, message, data}`，分页 `PageResponse{list,total,page,pageSize}` |

> 结论：**入参/出参信息质量足以直接反射提取**，这是本方案成立的前提。

---

## 3. 表结构

### 3.1 主表 `wg_sys_api_endpoint`

字段分三组，`[机]` 表示机器层写入、`[人]` 表示仅表单维护。机器同步**绝不覆盖** `[人]` 字段。

| 字段 | 类型 | 源 | 说明 |
|---|---|---|---|
| `id` | BIGINT PK | 机 | |
| `endpoint_key` | VARCHAR(128) UNIQUE | 机 | `METHOD path`，如 `PUT /api/users/:id`，同步幂等锚点 |
| `module` | VARCHAR(32) | 机 | auth / users / roles / menus / orgs / posts / tenants / apps / bind / open |
| `method` | VARCHAR(8) | 机 | |
| `path` | VARCHAR(255) | 机 | 原始 Gin 路径，含 `:id` |
| `perm_code` | VARCHAR(64) | 机 | 从路由声明提取；`/api/open/**` 为空（走签名鉴权） |
| `is_open` | TINYINT | 机 | 是否开放接口（走 `X-App-Key` 签名而非 JWT） |
| `path_params` | JSON | 机 | `[{name:"id", type:"integer", desc:"目标用户ID"}]` |
| `query_params` | JSON | 机 | `[{name:"page", type:"integer", default:1}]` |
| `request_schema` | JSON | 机 | JSON Schema draft-07，**含 `x-semantic` 扩展位** |
| `response_schema` | JSON | 机 | `data` 字段结构；标量/数组同样支持 |
| `status` | TINYINT | 机 | 1=在用；路由消失自动置 0（**天然下线检测**） |
| `summary` | VARCHAR(128) | 人 | 中文名，如「修改用户」 |
| `description` | TEXT | 人 | 业务语义、副作用、调用前置条件 |
| `ai_hint` | TEXT | 人 | **给 AI 的关键字段**：何时用 / 常见错误 / 不要怎么用 |
| `idempotent` | TINYINT | 人 | 是否幂等（AI 决定是否可重试） |
| `side_effect` | VARCHAR(64) | 人 | 写库 / 改密 / 发通知 / 无 |
| `request_example` | JSON | 人 | |
| `response_example` | JSON | 人 | |
| `synced_at` | DATETIME | 机 | 机器层上次同步时间 |
| `updated_at` | DATETIME | | |

### 3.2 为什么不单独建「字段语义表」

初版不建 `wg_sys_api_field`。字段级语义直接写在 `request_schema` 的 property 里：

```json
{
  "type": "object",
  "properties": {
    "status": {
      "type": "integer",
      "x-semantic": "0=禁用 1=启用；不传表示不修改（指针字段）",
      "x-example": 0
    }
  }
}
```

OpenAPI 允许 `x-` 前缀扩展，AI 读得到；机器同步时**保留人工已填的 `x-semantic`**。
好处：一张表就够了，不用维护两张表的关联一致性。若字段语义膨胀再拆表。

### 3.3 GORM 模型（可直接使用）

```go
type ApiEndpoint struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	EndpointKey     string    `gorm:"column:endpoint_key;size:128;uniqueIndex;not null" json:"endpointKey"`
	Module          string    `gorm:"column:module;size:32;not null;index" json:"module"`
	Method          string    `gorm:"column:method;size:8;not null" json:"method"`
	Path            string    `gorm:"column:path;size:255;not null" json:"path"`
	PermCode        string    `gorm:"column:perm_code;size:64" json:"permCode"`
	IsOpen          int       `gorm:"column:is_open" json:"isOpen"`
	Summary         string    `gorm:"column:summary;size:128" json:"summary"`
	Description     string    `gorm:"column:description;type:text" json:"description"`
	AiHint          string    `gorm:"column:ai_hint;type:text" json:"aiHint"`
	Idempotent      int       `gorm:"column:idempotent" json:"idempotent"`
	SideEffect      string    `gorm:"column:side_effect;size:64" json:"sideEffect"`
	PathParams      string    `gorm:"column:path_params;type:text" json:"pathParams"`
	QueryParams     string    `gorm:"column:query_params;type:text" json:"queryParams"`
	RequestSchema   string    `gorm:"column:request_schema;type:text" json:"requestSchema"`
	ResponseSchema  string    `gorm:"column:response_schema;type:text" json:"responseSchema"`
	RequestExample  string    `gorm:"column:request_example;type:text" json:"requestExample"`
	ResponseExample string    `gorm:"column:response_example;type:text" json:"responseExample"`
	Status          int       `gorm:"column:status;default:1" json:"status"`
	SyncedAt        time.Time `json:"syncedAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (ApiEndpoint) TableName() string { return "wg_sys_api_endpoint" }
```

> JSON 列在 SQLite 下用 `TEXT` 存字符串（与现有项目风格一致），取出后自行 `json.Marshal/Unmarshal`。

---

## 4. 关键设计：集中式路由声明表

**这是整套方案能否零维护的基础。** Go 反射拿不到 handler 闭包内部用了哪个 Request 类型，
所以必须把「路由 ↔ 权限 ↔ 入参 ↔ 出参」显式声明出来。

### 4.1 改造前（现状，`handler.go` 散落闭包）

```go
users.PUT("/:id", RequirePerm("system:user:edit"), func(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, BadRequest("参数错误"))
		return
	}
	resp, e := svc.UserUpdate(pathID(c, "id"), currentUID(c), req)
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, resp)
})
```

### 4.2 改造后（集中式声明，一行搞定路由+权限+文档）

```go
type Endpoint struct {
	Method  string      // GET/POST/PUT/DELETE
	Path    string      // /users/:id
	Module  string      // users
	Perm    string      // system:user:edit；空 = 走签名鉴权（/api/open/**）
	Req     interface{} // 入参样本，反射生成 schema（可为 nil）
	Resp    interface{} // 出参样本（可为 nil）
	Handler gin.HandlerFunc
}

var Endpoints = []Endpoint{
	{
		Method: "PUT", Path: "/users/:id", Module: "users",
		Perm: "system:user:edit",
		Req:  &UpdateUserRequest{},
		Resp: &UserVO{},
		Handler: func(c *gin.Context) {
			var req UpdateUserRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				Fail(c, BadRequest("参数错误"))
				return
			}
			resp, e := svc.UserUpdate(pathID(c, "id"), currentUID(c), req)
			if e != nil {
				Fail(c, e)
				return
			}
			OK(c, resp)
		},
	},
	// ... 其余 55 个
}
```

注册时**三件事一次做完**：

```go
func RegisterEndpoints(r *gin.RouterGroup) {
	for _, ep := range Endpoints {
		chain := []gin.HandlerFunc{AuthOrSig(ep)} // 统一权限/签名入口
		r.Handle(ep.Method, ep.Path, append(chain, ep.Handler)...)
		registry.UpsertMachineFields(ep) // 同步机器层元数据
	}
	registry.MarkMissingAsOffline() // 路由表里没有的旧记录 → status=0
}
```

### 4.3 收益

| 收益 | 说明 |
|---|---|
| 单一事实源 | 路由、权限、文档三者不可能不一致 |
| 零漂移 | 改路由就等于改文档；删接口自动识别下线（`status=0`） |
| 权限可视化 | 37 个权限码与 56 个端点的映射关系自动产出，不再靠人工核对 |
| 便于 review | 56 行声明比 589 行闭包更容易审计权限配置是否恰当 |

### 4.4 代价与风险

- 需要改写 `handler.go`（589 行）→ 约 56 条声明。**handler 函数体本身不用动**，只是搬位置
- 一次性改动，之后永久免维护
- 风险：搬迁过程可能出错 → **缓解**：改完跑 `go build` + 逐个端点 curl 冒烟（已有 `--mount` 之外的健康检查 `/api/health` 兜底）
- 建议在**独立分支**进行，与今日的 P0/P1 安全修复解耦

---

## 5. 反射器：`reflect.OpenAPI` 提取器

零外部依赖，约 200 行，职责：Go 结构体 → JSON Schema draft-07。

```go
func schemaOf(v interface{}) map[string]interface{} {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr { t = t.Elem() }
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		return map[string]interface{}{
			"type": "array",
			"items": schemaOf(reflect.New(t.Elem()).Interface()),
		}
	}
	props := map[string]interface{}{}
	required := []string{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" { continue }
		if strings.Contains(tag, "binding:\"required\"") ||
		   strings.Contains(f.Tag.Get("binding"), "required") {
			required = append(required, name)
		}
		props[name] = fieldType(f.Type)
	}
	s := map[string]interface{}{"type": "object", "properties": props}
	if len(required) > 0 { s["required"] = required }
	return s
}
```

**映射规则**（依据现有结构体实况定制）：

| Go 类型 | JSON Schema |
|---|---|
| `string` / `int64` / `int` / `bool` / `float64` | `string` / `integer` / `boolean` / `number` |
| `*T`（指针，如 `*int64`） | 与 `T` 同类型，**额外标注 `x-optional-semantics: "nil 表示不修改"`** |
| `[]T` / `[]SimpleRefVO` | `array` + 递归 `items` |
| 嵌套具名结构体 | 递归 `object` |
| `binding:"required"` | 加入 `required` 数组 |
| `time.Time` 或 `CreatedAt string` | `string`，`x-semantic` 标注时间格式 |

> 指针识别特别重要：本项目大量用 `*int64`/`*int` 表达「不传 = 不修改」（今日 P1-3 修复引入），
> 这正是 AI 最需要知道却**无法从 OpenAPI 常规字段看出**的语义，必须显式标注。

---

## 6. 导出与消费（三件套）

| 端点 | 格式 | 受众 | 用途 |
|---|---|---|---|
| `GET /api/meta/openapi.json` | OpenAPI 3.1 | **AI（首选）** | 完整 contract，可直接喂给任意 AI |
| `GET /api/meta/endpoints.json` | 扁平 JSON | 程序 / AI 快速检索 | 一次拉全量，token 省 |
| `GET /api/meta/endpoints.md` | Markdown | 人 | 粗略浏览、贴文档 |
| `GET /api/meta/openapi.json?module=users` | 过滤 | AI | 按需缩小上下文 |

**给 AI 的推荐喂法**（写进 `ai_hint` 引导，或直接作为系统提示词一部分）：

```
本系统接口契约见 GET /api/meta/openapi.json。
调用前请先确认：① 该端点需要哪个 perm_code ② 哪些字段是指针（不传=不修改）
③ x-semantic 里标注的枚举含义（如 status 0=禁用 1=启用）。
```

### OpenAPI 输出示例

```json
{
  "openapi": "3.1.0",
  "info": { "title": "go-auth-admin API", "version": "1.0.0" },
  "paths": {
    "/api/users/{id}": {
      "put": {
        "summary": "修改用户",
        "x-perm-code": "system:user:edit",
        "x-side-effect": "写库",
        "x-idempotent": true,
        "parameters": [
          { "name": "id", "in": "path", "required": true,
            "schema": { "type": "integer" } }
        ],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "nickname": { "type": "string" },
                  "orgId":    { "type": "integer",
                                "x-semantic": "nil 表示不修改（指针字段）" },
                  "status":   { "type": "integer",
                                "x-semantic": "0=禁用 1=启用；不传表示不修改" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "OK",
            "content": { "application/json": { "schema": {
              "$ref": "#/components/schemas/UserVO" } } }
          }
        }
      }
    }
  }
}
```

---

## 7. 落地路线图（建议 4 个阶段）

| 阶段 | 内容 | 产出 | 是否需要停服 |
|---|---|---|---|
| **S1** | 反射器 + `OpenApiMeta` 模型 + `ApiEndpoint` 表 + 迁移 | 表建好，反射器可用 | 否（AutoMigrate） |
| **S2** | 路由集中式改造（`Endpoints` 声明表） | 56 端点全部登记，机器字段自动填充 | 滚动重启即可 |
| **S3** | 导出端点（`openapi.json` / `.md`）+ 权限≡端点映射视图 | AI 可消费；管理员可查「哪些端点用了某权限」 | 否 |
| **S4** | Vue 管理端「接口文档」页：列表 + 详情 + **语义补充表单** | 人工层可维护 summary/ai_hint/x-semantic | 否 |

**建议先在独立分支做 S1+S2**，验证 `go build` + 全端点冒烟通过后再合入。

### S1 最小可用验证（不写测试，按项目约定）

```bash
# 1. 编译
cd backend && go build ./...   # 期望 exit 0
go vet ./...                   # 期望 exit 0

# 2. 启动后（本地起服务，非沙箱）
curl -s http://localhost:8080/api/meta/endpoints.json | head -c 800
# 期望：56 条记录，含 method/path/permCode/requestSchema

curl -s http://localhost:8080/api/meta/openapi.json | head -c 800
# 期望：合法 OpenAPI 3.1 JSON；可用任意 OpenAPI 校验器验证
```

---

## 8. 与今日安全修复的关系

- 今日 P0/P1 修复**已完成并验证通过**（`go build`/`go vet` exit 0），本方案不触碰那些改动
- 本方案会**重构 `handler.go` 的路由注册方式** —— 属于结构性改动，与 P0/P1 的业务逻辑修复正交，建议分开提交便于回溯
- P0-4 引入的签名契约变更（`rawQuery` + `sha256(body)`）会自动成为 `/api/open/**` 端点的元数据一部分，
  AI 读到 `is_open=1` 时应意识到走的是 `X-App-Key` 签名而非 Bearer JWT —— 建议写进这些端点的 `ai_hint`

---

## 9. 其他值得迭代的方向（按价值排序）

下方为独立附录，见 `docs/2026-10-07-next-iterations.md`。

---

## 10. 落地进展（2026-10-07 晚，S1+S2 已完成）

用户追问"接口的展示做了吗" → **代码已落地**（此前只有本文档，未写代码）。

### 10.1 重要发现：项目里本来就有一个手写文档页

`frontend/src/views/api-docs/index.vue`（10-04 基线自带，16KB）硬编码了 **48 条**接口表格。
它存在三个问题，恰好实证了本设计文档对"手写方案"的判断：

| 问题 | 实测 |
|---|---|
| 覆盖不全 | 48/56，缺 `GET /api/health`、租户 4 条、开放接口 3 条 |
| 菜单无入口 | DB `wg_sys_menu` 37 条里没有它，侧边栏根本看不到，只能手敲 `/api-docs` |
| 必然漂移 | 手写表格与代码无关联，加接口忘了改它就永久落后 |

该文件按用户"不擅自改名/挪动既有文件"的约定**保留未动**，新建页面取而代之。

### 10.2 已实现

| 层 | 产物 |
|---|---|
| 声明表 | `backend/apireg.go` 的 `apiDecls()`，58 条（56 业务端点 + 2 个元数据端点） |
| 反射 | `fieldsOf()` 从结构体实时展开字段（json tag / 类型 / 必填 / 嵌套一层），**改结构体即改文档** |
| 语义层 | `apiFieldHints` 补机器算不出的枚举含义（`status: 0=禁用,1=启用`、`dataScope`、`menuType` 等） |
| 端点 | `GET /api/api-registry`（扁平 JSON）、`GET /api/api-registry/openapi`（OpenAPI 3.1） |
| 防漂移 | `checkApiRegistry(r)` 启动时用 `gin.Engine.Routes()` 比对声明表，漏登记/已失效直接打日志点名 |
| 权限 | 新权限码 `system:api:view`（已入 `AuthPermsAll`），菜单 id=38 挂「系统管理」下，授权 super_admin |
| 前端 | `views/system/api/index.vue` + `api/apiRegistry.ts` + 路由，按模块分组、可搜索、点击展开入参/出参字段表 |
| 重构 | `PUT /users/:id/password` 的匿名 struct 提取为 `ResetPasswordRequest`（否则无法反射登记） |

启动日志实测：`[api-registry] 元数据已覆盖全部 58 个路由`。

### 10.3 与原设计的差异（有意偏离）

原设计 §4 要求把 `handler.go` 的 589 行重写为集中式路由声明表。实际**没有这么做**，改为
「保留现有闭包写法 + 独立声明表 + 启动时用 `Routes()` 校验」：

- 理由：闭包改声明表是纯结构性重写，风险与收益不成正比；而漂移风险已被启动校验兜住
  （漏登记会点名，不会静默落后）
- 代价：新增路由仍需手工在 `apiDecls()` 登记一行（但忘登记会有日志告警，不是静默失败）
- 人工层（Summary / hints）暂放在 Go 代码里，**没有建 `wg_sys_api_endpoint` 表**
  —— 表单补语义的 UI 属于 S3，收益需先验证 S1/S2 是否被真正用起来

### 10.4 尚未做（S3/S4）

- DB 表 + 管理端表单（人工补 `ai_hint`、业务语义）
- MCP Server 暴露（让 AI 从"读文档"变成"直接调用"）
- TS 类型自动生成（顺带修 P3-15）
- 拆 Update 专用结构体（P3-12，当前 `PUT /orgs|/posts|/tenants` 复用 Create 结构体，
  其 `binding:"required"` 会让 AI 看到的 `required` 字段偏多）
