package main

import (
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 本文件是「接口元数据层」：把散落在 handler.go / service.go 里的接口契约
// 集中成一份机器可消费的清单。
//
// 设计原则（对应 docs/2026-10-07-api-registry-design.md）：
//   - 路由与权限码来自 gin 的真实注册结果，出参/入参字段由反射从结构体实时展开，
//     因此增删字段、改 json tag 都会自动反映到这里，不会漂移。
//   - 唯一需要人工维护的是 apiDecls 里的 Summary（中文语义）与 apiFieldHints（枚举含义），
//     这部分是机器算不出来的业务语义。

// ResetPasswordRequest 重置密码入参。
// 原先是 handler 内的匿名 struct，提取为具名类型后才能被元数据登记与反射。
type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

// AuthKind 标识接口的鉴权方式。
const (
	AuthPublic = "public" // 无需鉴权
	AuthJWT    = "jwt"    // 需登录（权限码见 Perm）
	AuthAppSig = "app"    // 应用签名鉴权（/api/open/**）
)

// ApiDecl 一条接口的静态声明。
type ApiDecl struct {
	Method    string   // GET / POST / PUT / DELETE
	Path      string   // gin 风格路径，含 :id
	Module    string   // 模块英文名
	Perm      string   // 所需权限码，空表示不校验
	Auth      string   // AuthPublic / AuthJWT / AuthAppSig
	Summary   string   // 中文说明（人工维护）
	Req       any      // 入参示例，nil 表示无 body
	Resp      any      // 出参示例，nil 表示无返回体
	RespItem  any      // 当 Resp 为 PageResponse 时，list 元素示例
	RespSlice bool     // true 表示返回的是 Resp 类型的数组
	Query     []string // query 参数名
}

// apiDecls 全量接口声明表。新增路由后请同步在此登记。
func apiDecls() []ApiDecl {
	return []ApiDecl{
		// ===== 系统 =====
		{Method: "GET", Path: "/api/health", Module: "system", Auth: AuthPublic, Summary: "健康检查，无需鉴权"},
		{Method: "GET", Path: "/api/api-registry", Module: "system", Perm: "system:api:view", Auth: AuthJWT, Summary: "全量接口清单（本页数据源，含入参/出参字段级结构）"},
		{Method: "GET", Path: "/api/api-registry/openapi", Module: "system", Perm: "system:api:view", Auth: AuthJWT, Summary: "OpenAPI 3.1 文档，供 AI 与外部工具直接消费"},

		// ===== 认证与账户 =====
		{Method: "POST", Path: "/api/auth/login", Module: "auth", Auth: AuthPublic, Summary: "账号密码登录，返回 accessToken / refreshToken", Req: &LoginRequest{}, Resp: &LoginResponseVO{}},
		{Method: "POST", Path: "/api/auth/refresh", Module: "auth", Auth: AuthPublic, Summary: "用 refreshToken 换发新令牌", Req: &RefreshRequest{}, Resp: &LoginResponseVO{}},
		{Method: "POST", Path: "/api/auth/logout", Module: "auth", Auth: AuthJWT, Summary: "登出（无状态，前端清除 token 即视为登出）"},
		{Method: "GET", Path: "/api/auth/me", Module: "auth", Auth: AuthJWT, Summary: "当前登录用户信息，含 roles / permissions 三段式权限码", Resp: &UserInfoVO{}},
		{Method: "PUT", Path: "/api/auth/password", Module: "auth", Auth: AuthJWT, Summary: "当前用户修改自己的密码", Req: &ChangePasswordRequest{}},
		{Method: "GET", Path: "/api/auth/menus", Module: "auth", Auth: AuthJWT, Summary: "当前用户可见菜单树（仅含本人有权限的菜单）", Resp: &MenuNodeVO{}, RespSlice: true},

		// ===== 用户 =====
		{Method: "GET", Path: "/api/users", Module: "users", Perm: "system:user:list", Auth: AuthJWT, Summary: "用户分页列表", Query: []string{"page", "pageSize", "username", "status"}, Resp: &PageResponse{}, RespItem: &UserVO{}},
		{Method: "GET", Path: "/api/users/:id", Module: "users", Perm: "system:user:list", Auth: AuthJWT, Summary: "用户详情，含角色与岗位", Resp: &UserVO{}},
		{Method: "POST", Path: "/api/users", Module: "users", Perm: "system:user:create", Auth: AuthJWT, Summary: "创建用户", Req: &CreateUserRequest{}, Resp: &UserVO{}},
		{Method: "PUT", Path: "/api/users/:id", Module: "users", Perm: "system:user:edit", Auth: AuthJWT, Summary: "更新用户（不可改 username，也不可改密码）", Req: &UpdateUserRequest{}, Resp: &UserVO{}},
		{Method: "DELETE", Path: "/api/users/:id", Module: "users", Perm: "system:user:delete", Auth: AuthJWT, Summary: "删除用户"},
		{Method: "PUT", Path: "/api/users/:id/roles", Module: "users", Perm: "system:user:assign-role", Auth: AuthJWT, Summary: "分配角色（禁止改自身角色，禁止授予 super_admin）", Req: &AssignRolesRequest{}},
		{Method: "PUT", Path: "/api/users/:id/posts", Module: "users", Perm: "system:user:assign-post", Auth: AuthJWT, Summary: "分配岗位", Req: &AssignPostsRequest{}},
		{Method: "PUT", Path: "/api/users/:id/password", Module: "users", Perm: "system:user:reset-pwd", Auth: AuthJWT, Summary: "管理员重置指定用户密码", Req: &ResetPasswordRequest{}},
		{Method: "GET", Path: "/api/users/:id/roles", Module: "users", Perm: "system:user:list", Auth: AuthJWT, Summary: "查询用户已分配的角色 ID 列表", Resp: new(int64), RespSlice: true},
		{Method: "GET", Path: "/api/users/:id/posts", Module: "users", Perm: "system:user:list", Auth: AuthJWT, Summary: "查询用户已分配的岗位 ID 列表", Resp: new(int64), RespSlice: true},

		// ===== 角色 =====
		{Method: "GET", Path: "/api/roles", Module: "roles", Perm: "system:role:list", Auth: AuthJWT, Summary: "角色分页列表", Query: []string{"page", "pageSize", "keyword"}, Resp: &PageResponse{}, RespItem: &RoleVO{}},
		{Method: "GET", Path: "/api/roles/:id", Module: "roles", Perm: "system:role:list", Auth: AuthJWT, Summary: "角色详情", Resp: &RoleVO{}},
		{Method: "POST", Path: "/api/roles", Module: "roles", Perm: "system:role:create", Auth: AuthJWT, Summary: "创建角色（code 不可为 super_admin）", Req: &CreateRoleRequest{}, Resp: &RoleVO{}},
		{Method: "PUT", Path: "/api/roles/:id", Module: "roles", Perm: "system:role:edit", Auth: AuthJWT, Summary: "更新角色", Req: &UpdateRoleRequest{}, Resp: &RoleVO{}},
		{Method: "DELETE", Path: "/api/roles/:id", Module: "roles", Perm: "system:role:delete", Auth: AuthJWT, Summary: "删除角色"},
		{Method: "PUT", Path: "/api/roles/:id/menus", Module: "roles", Perm: "system:role:assign-menu", Auth: AuthJWT, Summary: "给角色分配菜单权限", Req: &AssignMenusRequest{}},
		{Method: "GET", Path: "/api/roles/:id/menus", Module: "roles", Perm: "system:role:list", Auth: AuthJWT, Summary: "查询角色已分配的菜单 ID 列表", Resp: new(int64), RespSlice: true},

		// ===== 菜单 =====
		{Method: "GET", Path: "/api/menus", Module: "menus", Perm: "system:menu:list", Auth: AuthJWT, Summary: "菜单树（全量，含按钮级）", Resp: &MenuNodeVO{}, RespSlice: true},
		{Method: "POST", Path: "/api/menus", Module: "menus", Perm: "system:menu:create", Auth: AuthJWT, Summary: "创建菜单", Req: &CreateMenuRequest{}, Resp: &MenuNodeVO{}},
		{Method: "PUT", Path: "/api/menus/:id", Module: "menus", Perm: "system:menu:edit", Auth: AuthJWT, Summary: "更新菜单", Req: &UpdateMenuRequest{}, Resp: &MenuNodeVO{}},
		{Method: "DELETE", Path: "/api/menus/:id", Module: "menus", Perm: "system:menu:delete", Auth: AuthJWT, Summary: "删除菜单（有子菜单时拒绝）"},

		// ===== 组织 =====
		{Method: "GET", Path: "/api/orgs", Module: "orgs", Perm: "system:org:list", Auth: AuthJWT, Summary: "组织树（不分页）", Resp: &OrgNodeVO{}, RespSlice: true},
		{Method: "POST", Path: "/api/orgs", Module: "orgs", Perm: "system:org:create", Auth: AuthJWT, Summary: "创建组织", Req: &CreateOrgRequest{}, Resp: &OrgNodeVO{}},
		// 注意：更新沿用 CreateOrgRequest，其 binding:"required" 会让 name 成为必填（已知契约问题）
		{Method: "PUT", Path: "/api/orgs/:id", Module: "orgs", Perm: "system:org:edit", Auth: AuthJWT, Summary: "更新组织（禁止把父级改成自己或其子孙）", Req: &CreateOrgRequest{}, Resp: &OrgNodeVO{}},
		{Method: "DELETE", Path: "/api/orgs/:id", Module: "orgs", Perm: "system:org:delete", Auth: AuthJWT, Summary: "删除组织（有子组织时拒绝）"},

		// ===== 岗位 =====
		{Method: "GET", Path: "/api/posts", Module: "posts", Perm: "system:post:list", Auth: AuthJWT, Summary: "岗位列表（不分页）", Resp: &PostVO{}, RespSlice: true},
		{Method: "POST", Path: "/api/posts", Module: "posts", Perm: "system:post:create", Auth: AuthJWT, Summary: "创建岗位", Req: &CreatePostRequest{}, Resp: &PostVO{}},
		{Method: "PUT", Path: "/api/posts/:id", Module: "posts", Perm: "system:post:edit", Auth: AuthJWT, Summary: "更新岗位", Req: &CreatePostRequest{}, Resp: &PostVO{}},
		{Method: "DELETE", Path: "/api/posts/:id", Module: "posts", Perm: "system:post:delete", Auth: AuthJWT, Summary: "删除岗位"},

		// ===== 租户（后端保留，前端未开放入口） =====
		{Method: "GET", Path: "/api/tenants", Module: "tenants", Perm: "system:tenant:list", Auth: AuthJWT, Summary: "租户列表（不分页）", Resp: &TenantVO{}, RespSlice: true},
		{Method: "POST", Path: "/api/tenants", Module: "tenants", Perm: "system:tenant:create", Auth: AuthJWT, Summary: "创建租户", Req: &CreateTenantRequest{}, Resp: &TenantVO{}},
		{Method: "PUT", Path: "/api/tenants/:id", Module: "tenants", Perm: "system:tenant:edit", Auth: AuthJWT, Summary: "更新租户", Req: &CreateTenantRequest{}, Resp: &TenantVO{}},
		{Method: "DELETE", Path: "/api/tenants/:id", Module: "tenants", Perm: "system:tenant:delete", Auth: AuthJWT, Summary: "删除租户"},

		// ===== 应用 =====
		{Method: "GET", Path: "/api/apps", Module: "apps", Perm: "system:app:list", Auth: AuthJWT, Summary: "应用分页列表", Query: []string{"page", "pageSize"}, Resp: &PageResponse{}, RespItem: &AppVO{}},
		{Method: "GET", Path: "/api/apps/:id", Module: "apps", Perm: "system:app:list", Auth: AuthJWT, Summary: "应用详情", Resp: &AppVO{}},
		{Method: "POST", Path: "/api/apps", Module: "apps", Perm: "system:app:create", Auth: AuthJWT, Summary: "创建应用", Req: &CreateAppRequest{}, Resp: &AppVO{}},
		{Method: "PUT", Path: "/api/apps/:id", Module: "apps", Perm: "system:app:edit", Auth: AuthJWT, Summary: "更新应用", Req: &UpdateAppRequest{}, Resp: &AppVO{}},
		{Method: "DELETE", Path: "/api/apps/:id", Module: "apps", Perm: "system:app:delete", Auth: AuthJWT, Summary: "删除应用"},
		{Method: "GET", Path: "/api/apps/:id/keys", Module: "apps", Perm: "system:app:key-manage", Auth: AuthJWT, Summary: "应用密钥分页列表", Resp: &PageResponse{}, RespItem: &AppKeyVO{}},
		{Method: "POST", Path: "/api/apps/:id/keys", Module: "apps", Perm: "system:app:key-manage", Auth: AuthJWT, Summary: "创建应用密钥（仅此一次返回明文 apiSecret）", Req: &CreateAppKeyRequest{}, Resp: &CreatedAppKeyVO{}},
		{Method: "DELETE", Path: "/api/apps/:id/keys/:keyId", Module: "apps", Perm: "system:app:key-manage", Auth: AuthJWT, Summary: "吊销应用密钥"},

		// ===== 应用绑定 =====
		{Method: "GET", Path: "/api/bind", Module: "bind", Perm: "account:bind:list", Auth: AuthJWT, Summary: "我的绑定列表", Resp: &BindVO{}, RespSlice: true},
		{Method: "POST", Path: "/api/bind/code", Module: "bind", Perm: "account:bind:create", Auth: AuthJWT, Summary: "生成绑定码（返回 qrContent 与 expireAt）", Req: &CreateBindCodeRequest{}, Resp: &BindCodeVO{}},
		{Method: "POST", Path: "/api/bind/confirm", Module: "bind", Perm: "account:bind:create", Auth: AuthJWT, Summary: "确认绑定", Req: &BindConfirmRequest{}, Resp: &BindVO{}},
		{Method: "DELETE", Path: "/api/bind/:id", Module: "bind", Perm: "account:bind:delete", Auth: AuthJWT, Summary: "解除绑定"},
		{Method: "PUT", Path: "/api/bind/:id/default", Module: "bind", Perm: "account:bind:set-default", Auth: AuthJWT, Summary: "设为默认绑定"},

		// ===== 开放接口（应用签名鉴权） =====
		// 签名待签串：apiKey\n ts\n nonce\n METHOD\n path\n rawQuery\n sha256(body)
		{Method: "POST", Path: "/api/open/bind/apply", Module: "open", Auth: AuthAppSig, Summary: "【开放接口】应用侧申请生成绑定码", Req: &OpenBindApplyRequest{}, Resp: &BindCodeVO{}},
		{Method: "POST", Path: "/api/open/bind/claim", Module: "open", Auth: AuthAppSig, Summary: "【开放接口】应用侧凭绑定码完成绑定", Req: &OpenBindClaimRequest{}, Resp: &BindVO{}},
		{Method: "GET", Path: "/api/open/userinfo", Module: "open", Auth: AuthAppSig, Summary: "【开放接口】按 appUserId 查询绑定用户信息", Query: []string{"appUserId"}, Resp: &OpenUserInfoVO{}},
	}
}

// apiFieldHints 字段级业务语义（机器算不出来的部分，人工维护）。
// key 为 json 字段名，命中即写入 Desc，供 AI 与前端展示。
var apiFieldHints = map[string]string{
	"status":      "0=禁用，1=启用",
	"dataScope":   "ALL=全部 / SELF_ONLY=仅本人 / DEPT_ONLY=本部门 / DEPT_AND_BELOW=本部门及以下",
	"menuType":    "M=目录 / C=菜单 / F=按钮",
	"visible":     "0=隐藏，1=显示",
	"direction":   "绑定方向：APP_TO_PLATFORM / PLATFORM_TO_APP",
	"grantType":   "授权类型",
	"roleIds":     "角色 ID 列表，全量覆盖（传空数组表示清空）",
	"postIds":     "岗位 ID 列表，全量覆盖（传空数组表示清空）",
	"menuIds":     "菜单 ID 列表，全量覆盖（传空数组表示清空）",
	"newPassword": "新密码",
	"oldPassword": "原密码",
	"appUserId":   "用户在应用侧的外部 ID",
}

// ApiFieldMeta 字段级元数据。
type ApiFieldMeta struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Required bool           `json:"required"`
	Desc     string         `json:"desc,omitempty"`
	Children []ApiFieldMeta `json:"children,omitempty"`
}

// ApiMeta 对外输出的单条接口元数据。
type ApiMeta struct {
	Method   string         `json:"method"`
	Path     string         `json:"path"`
	Module   string         `json:"module"`
	ModuleCN string         `json:"moduleCn"`
	Perm     string         `json:"perm,omitempty"`
	Auth     string         `json:"auth"`
	Summary  string         `json:"summary"`
	PathVars []string       `json:"pathVars"`
	Query    []string       `json:"query"`
	ReqType  string         `json:"reqType"`
	RespType string         `json:"respType"`
	Request  []ApiFieldMeta `json:"request"`
	Response []ApiFieldMeta `json:"response"`
	RespItem []ApiFieldMeta `json:"respItem,omitempty"`
}

var moduleCN = map[string]string{
	"system":  "系统",
	"auth":    "认证与账户",
	"users":   "用户",
	"roles":   "角色",
	"menus":   "菜单",
	"orgs":    "组织",
	"posts":   "岗位",
	"tenants": "租户",
	"apps":    "应用",
	"bind":    "应用绑定",
	"open":    "开放接口",
}

// pathVarsOf 从 gin 路径中抽出路径参数名（不含冒号）。
func pathVarsOf(path string) []string {
	var out []string
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, ":") {
			out = append(out, strings.TrimPrefix(seg, ":"))
		}
	}
	return out
}

// openapiPath 把 gin 的 :id 转成 OpenAPI 的 {id}。
func openapiPath(path string) string {
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, ":") {
			segs[i] = "{" + strings.TrimPrefix(s, ":") + "}"
		}
	}
	return strings.Join(segs, "/")
}

// jsonTypeOf 把 Go 类型映射为 JSON Schema 的 type。
func jsonTypeOf(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct, reflect.Interface:
		return "object"
	}
	return "string"
}

// isPlainStruct 判断是否为可继续展开的业务结构体（排除 time.Time 等基础类型）。
func isPlainStruct(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	if t == reflect.TypeOf(time.Time{}) {
		return false
	}
	if t.PkgPath() == "time" {
		return false
	}
	return true
}

// fieldsOf 反射展开结构体的字段元数据；depth 控制嵌套展开层数。
func fieldsOf(v any, depth int) []ApiFieldMeta {
	if v == nil {
		return nil
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if !isPlainStruct(t) {
		return nil
	}
	out := make([]ApiFieldMeta, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" {
			name = f.Name
		}
		if name == "-" {
			continue
		}
		fm := ApiFieldMeta{
			Name:     name,
			Type:     jsonTypeOf(f.Type),
			Required: strings.Contains(f.Tag.Get("binding"), "required"),
			Desc:     apiFieldHints[name],
		}
		// 嵌套结构体展开一层，便于 AI 理解 VO 的内部形态
		if depth > 0 {
			et := f.Type
			for et.Kind() == reflect.Ptr || et.Kind() == reflect.Slice {
				et = et.Elem()
			}
			if isPlainStruct(et) {
				fm.Children = fieldsOf(reflect.New(et).Interface(), depth-1)
			}
		}
		out = append(out, fm)
	}
	return out
}

// typeLabel 输出人类可读的类型名。
func typeLabel(v any, slice bool, item any) string {
	if v == nil {
		return "void"
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	n := t.Name()
	if n == "" {
		n = "object"
	}
	if item != nil {
		it := reflect.TypeOf(item)
		for it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		return n + "<" + it.Name() + ">"
	}
	if slice {
		return "[]" + n
	}
	return n
}

// buildApiMeta 生成全量接口元数据。
func buildApiMeta() []ApiMeta {
	decls := apiDecls()
	out := make([]ApiMeta, 0, len(decls))
	for _, d := range decls {
		m := ApiMeta{
			Method:   d.Method,
			Path:     d.Path,
			Module:   d.Module,
			ModuleCN: moduleCN[d.Module],
			Perm:     d.Perm,
			Auth:     d.Auth,
			Summary:  d.Summary,
			PathVars: pathVarsOf(d.Path),
			Query:    d.Query,
			ReqType:  typeLabel(d.Req, false, nil),
			RespType: typeLabel(d.Resp, d.RespSlice, d.RespItem),
			Request:  fieldsOf(d.Req, 2),
			Response: fieldsOf(d.Resp, 2),
		}
		if d.RespItem != nil {
			m.RespItem = fieldsOf(d.RespItem, 1)
		}
		out = append(out, m)
	}
	return out
}

// apiRegistryHandler GET /api/api-registry —— 全量接口清单（登录即可访问）。
func apiRegistryHandler(c *gin.Context) {
	list := buildApiMeta()
	OK(c, gin.H{"total": len(list), "list": list})
}

// openapiDoc 构造 OpenAPI 3.1 文档（AI 训练语料中最常见的契约格式）。
func openapiDoc(list []ApiMeta) map[string]any {
	paths := map[string]any{}
	for _, m := range list {
		p := openapiPath(m.Path)
		op := map[string]any{
			"summary": m.Summary,
			"tags":    []string{m.ModuleCN},
			"responses": map[string]any{
				"200": map[string]any{"description": "success"},
			},
		}
		if m.Auth == AuthJWT {
			op["security"] = []any{map[string]any{"bearerAuth": []string{}}}
		}
		if m.Auth == AuthAppSig {
			op["security"] = []any{map[string]any{"appSignature": []string{}}}
		}

		// 路径参数
		if len(m.PathVars) > 0 {
			params := make([]any, 0, len(m.PathVars))
			for _, pv := range m.PathVars {
				params = append(params, map[string]any{
					"name": pv, "in": "path", "required": true,
					"schema": map[string]any{"type": "integer"},
				})
			}
			op["parameters"] = params
		}
		// query 参数
		for _, q := range m.Query {
			params, _ := op["parameters"].([]any)
			params = append(params, map[string]any{
				"name": q, "in": "query", "required": false,
				"schema": map[string]any{"type": "string"},
			})
			op["parameters"] = params
		}

		// 请求体
		if len(m.Request) > 0 {
			op["requestBody"] = map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{"schema": schemaOf(m.Request)},
				},
			}
		}
		// 响应体
		if len(m.Response) > 0 || len(m.RespItem) > 0 {
			var dataSchema map[string]any
			if len(m.RespItem) > 0 {
				dataSchema = map[string]any{
					"type": "object",
					"properties": map[string]any{
						"list":  map[string]any{"type": "array", "items": schemaOf(m.RespItem)},
						"total": map[string]any{"type": "integer"},
					},
				}
			} else if len(m.Response) > 0 {
				dataSchema = schemaOf(m.Response)
			}
			op["responses"] = map[string]any{
				"200": map[string]any{
					"description": "success",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"code":    map[string]any{"type": "integer", "example": 0},
									"message": map[string]any{"type": "string", "example": "success"},
									"data":    dataSchema,
								},
							},
						},
					},
				},
			}
		}
		entry, ok := paths[p].(map[string]any)
		if !ok {
			entry = map[string]any{}
			paths[p] = entry
		}
		entry[strings.ToLower(m.Method)] = op
	}

	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   "go-auth-admin API",
			"version": "1.0.0",
			"description": "由后端反射自动生成，源为代码本身，不会漂移。" +
				"统一响应信封 {code,message,data}；code=0 表示成功。",
		},
		"paths": paths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]any{
					"type": "http", "scheme": "bearer", "bearerFormat": "JWT",
				},
				"appSignature": map[string]any{
					"type": "apiKey", "in": "header", "name": "X-Api-Key",
					"description": "另需 X-Timestamp / X-Nonce / X-Signature；" +
						"待签串 = apiKey\\n ts\\n nonce\\n METHOD\\n path\\n rawQuery\\n sha256(body)",
				},
			},
		},
	}
}

// schemaOf 把字段列表转成 JSON Schema。
func schemaOf(fields []ApiFieldMeta) map[string]any {
	props := map[string]any{}
	required := make([]string, 0)
	for _, f := range fields {
		s := map[string]any{"type": f.Type}
		if f.Desc != "" {
			s["description"] = f.Desc
		}
		if len(f.Children) > 0 {
			s["properties"] = schemaOf(f.Children)["properties"]
		}
		props[f.Name] = s
		if f.Required {
			required = append(required, f.Name)
		}
	}
	out := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		out["required"] = required
	}
	return out
}

// apiRegistryOpenAPIHandler GET /api/api-registry/openapi —— OpenAPI 3.1 文档。
func apiRegistryOpenAPIHandler(c *gin.Context) {
	c.JSON(http.StatusOK, openapiDoc(buildApiMeta()))
}

// checkApiRegistry 启动时比对 gin 真实路由与声明表，把漏登记/已失效的打印到日志。
//
// 这是防漂移的兜底：声明表是唯一需要人工维护的部分，一旦加路由忘了登记，
// 启动日志会立刻点名，而不是让文档悄悄落后于代码。
func checkApiRegistry(r *gin.Engine) {
	declared := make(map[string]bool, 64)
	for _, d := range apiDecls() {
		declared[d.Method+" "+d.Path] = true
	}

	live := make(map[string]bool, 64)
	var missing, stale []string
	for _, rt := range r.Routes() {
		key := rt.Method + " " + rt.Path
		live[key] = true
		if !declared[key] {
			missing = append(missing, key)
		}
	}
	for k := range declared {
		if !live[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)

	if len(missing) == 0 && len(stale) == 0 {
		log.Printf("[api-registry] 元数据已覆盖全部 %d 个路由", len(live))
		return
	}
	if len(missing) > 0 {
		log.Printf("[api-registry] 警告：%d 个路由未登记元数据，请补进 apiDecls() -> %v", len(missing), missing)
	}
	if len(stale) > 0 {
		log.Printf("[api-registry] 警告：%d 条声明已无对应路由，请从 apiDecls() 移除 -> %v", len(stale), stale)
	}
}
