package main

import (
	"github.com/gin-gonic/gin"
)

// currentUID 从上下文取当前登录用户 ID。
func currentUID(c *gin.Context) int64 {
	if v, ok := c.Get(ctxUID); ok {
		if uid, ok := v.(int64); ok {
			return uid
		}
	}
	return 0
}

func currentPrincipal(c *gin.Context) *AppPrincipal {
	if v, ok := c.Get(ctxAppPrinc); ok {
		if p, ok := v.(*AppPrincipal); ok {
			return p
		}
	}
	return nil
}

// RegisterRoutes 装配全部路由（路径与 java-vue-admin 逐字一致）。
func RegisterRoutes(r *gin.Engine, svc *Services) {
	api := r.Group("/api")

	// 健康检查（公共）
	api.GET("/health", func(c *gin.Context) { OK(c, gin.H{"status": "ok"}) })

	// ===== 接口元数据（供 AI 与前端文档页消费，需登录） =====
	// 全量接口清单：GET /api/api-registry
	// OpenAPI 3.1：GET /api/api-registry/openapi
	api.GET("/api-registry", RequirePerm("system:api:view"), apiRegistryHandler)
	api.GET("/api-registry/openapi", RequirePerm("system:api:view"), apiRegistryOpenAPIHandler)

	// ===== 认证（login/refresh 公共，其余需登录） =====
	api.POST("/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.Login(req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	api.POST("/auth/refresh", func(c *gin.Context) {
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.Refresh(req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	api.POST("/auth/logout", func(c *gin.Context) { OK(c, nil) })
	api.GET("/auth/me", func(c *gin.Context) {
		resp, e := svc.Me(currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	api.PUT("/auth/password", func(c *gin.Context) {
		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		if e := svc.ChangePassword(currentUID(c), req); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	api.GET("/auth/menus", func(c *gin.Context) {
		resp, e := svc.GetMenus(currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})

	// ===== 用户 =====
	users := api.Group("/users")
	users.GET("", RequirePerm("system:user:list"), func(c *gin.Context) {
		page := atoiDefault(c.Query("page"), 1)
		size := atoiDefault(c.Query("pageSize"), 10)
		username := c.Query("username")
		var status *int
		if c.Query("status") != "" {
			v := atoiDefault(c.Query("status"), 0)
			status = &v
		}
		resp, e := svc.UserPage(page, size, username, status, currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	users.GET("/:id", RequirePerm("system:user:list"), func(c *gin.Context) {
		id := pathID(c, "id")
		resp, e := svc.UserGetVO(id, currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	users.POST("", RequirePerm("system:user:create"), func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.UserCreate(req, currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	users.PUT("/:id", RequirePerm("system:user:edit"), func(c *gin.Context) {
		id := pathID(c, "id")
		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.UserUpdate(id, currentUID(c), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	users.DELETE("/:id", RequirePerm("system:user:delete"), func(c *gin.Context) {
		if e := svc.UserDelete(pathID(c, "id"), currentUID(c)); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	users.PUT("/:id/roles", RequirePerm("system:user:assign-role"), func(c *gin.Context) {
		var req AssignRolesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		if e := svc.UserAssignRoles(pathID(c, "id"), currentUID(c), req); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	users.PUT("/:id/posts", RequirePerm("system:user:assign-post"), func(c *gin.Context) {
		var req AssignPostsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		if e := svc.UserAssignPosts(pathID(c, "id"), currentUID(c), req); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	users.PUT("/:id/password", RequirePerm("system:user:reset-pwd"), func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		if e := svc.UserResetPassword(pathID(c, "id"), currentUID(c), req.NewPassword); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	users.GET("/:id/roles", RequirePerm("system:user:list"), func(c *gin.Context) {
		ids, e := svc.UserRoleIDs(pathID(c, "id"), currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, ids)
	})
	users.GET("/:id/posts", RequirePerm("system:user:list"), func(c *gin.Context) {
		ids, e := svc.UserPostIDs(pathID(c, "id"), currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, ids)
	})

	// ===== 角色 =====
	roles := api.Group("/roles")
	roles.GET("", RequirePerm("system:role:list"), func(c *gin.Context) {
		page := atoiDefault(c.Query("page"), 1)
		size := atoiDefault(c.Query("pageSize"), 10)
		resp, e := svc.RolePage(page, size, c.Query("keyword"))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	roles.GET("/:id", RequirePerm("system:role:list"), func(c *gin.Context) {
		resp, e := svc.RoleGetVO(pathID(c, "id"))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	roles.POST("", RequirePerm("system:role:create"), func(c *gin.Context) {
		var req CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.RoleCreate(req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	roles.PUT("/:id", RequirePerm("system:role:edit"), func(c *gin.Context) {
		var req UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.RoleUpdate(pathID(c, "id"), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	roles.DELETE("/:id", RequirePerm("system:role:delete"), func(c *gin.Context) {
		if e := svc.RoleDelete(pathID(c, "id")); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	roles.PUT("/:id/menus", RequirePerm("system:role:assign-menu"), func(c *gin.Context) {
		var req AssignMenusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		if e := svc.RoleAssignMenus(pathID(c, "id"), req); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	roles.GET("/:id/menus", RequirePerm("system:role:list"), func(c *gin.Context) {
		OK(c, svc.RoleMenuIDs(pathID(c, "id")))
	})

	// ===== 菜单 =====
	menus := api.Group("/menus")
	menus.GET("", RequirePerm("system:menu:list"), func(c *gin.Context) {
		OK(c, svc.MenuTree())
	})
	menus.POST("", RequirePerm("system:menu:create"), func(c *gin.Context) {
		var req CreateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.MenuCreate(req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	menus.PUT("/:id", RequirePerm("system:menu:edit"), func(c *gin.Context) {
		var req UpdateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.MenuUpdate(pathID(c, "id"), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	menus.DELETE("/:id", RequirePerm("system:menu:delete"), func(c *gin.Context) {
		if e := svc.MenuDelete(pathID(c, "id")); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})

	// ===== 组织 =====
	orgs := api.Group("/orgs")
	orgs.GET("", RequirePerm("system:org:list"), func(c *gin.Context) { OK(c, svc.OrgTree()) })
	orgs.POST("", RequirePerm("system:org:create"), func(c *gin.Context) {
		var req CreateOrgRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.OrgCreate(req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	orgs.PUT("/:id", RequirePerm("system:org:edit"), func(c *gin.Context) {
		var req CreateOrgRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.OrgUpdate(pathID(c, "id"), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	orgs.DELETE("/:id", RequirePerm("system:org:delete"), func(c *gin.Context) {
		if e := svc.OrgDelete(pathID(c, "id")); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})

	// ===== 岗位 =====
	posts := api.Group("/posts")
	posts.GET("", RequirePerm("system:post:list"), func(c *gin.Context) { OK(c, svc.PostList()) })
	posts.POST("", RequirePerm("system:post:create"), func(c *gin.Context) {
		var req CreatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.PostCreate(req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	posts.PUT("/:id", RequirePerm("system:post:edit"), func(c *gin.Context) {
		var req CreatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.PostUpdate(pathID(c, "id"), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	posts.DELETE("/:id", RequirePerm("system:post:delete"), func(c *gin.Context) {
		if e := svc.PostDelete(pathID(c, "id")); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})

	// ===== 租户 =====
	// 【租户：功能保留但标记为「未启用」】
	// 后端 CRUD 与 4 个 system:tenant:* 权限保留（对齐 Java 版），
	// 但刻意不接入前端：seed 不播种租户菜单，故非超管角色无从勾选获得租户权限。
	// 结论：租户能力当前仅超管（短路放行）可见，勿在前台/文档对外暴露，勿新增前端页面。
	tenants := api.Group("/tenants")
	tenants.GET("", RequirePerm("system:tenant:list"), func(c *gin.Context) { OK(c, svc.TenantList()) })
	tenants.POST("", RequirePerm("system:tenant:create"), func(c *gin.Context) {
		var req CreateTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.TenantCreate(req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	tenants.PUT("/:id", RequirePerm("system:tenant:edit"), func(c *gin.Context) {
		var req CreateTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.TenantUpdate(pathID(c, "id"), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	tenants.DELETE("/:id", RequirePerm("system:tenant:delete"), func(c *gin.Context) {
		if e := svc.TenantDelete(pathID(c, "id")); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})

	// ===== 应用 =====
	apps := api.Group("/apps")
	apps.GET("", RequirePerm("system:app:list"), func(c *gin.Context) {
		page := atoiDefault(c.Query("page"), 1)
		size := atoiDefault(c.Query("pageSize"), 10)
		resp, e := svc.AppPage(currentUID(c), page, size)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	apps.GET("/:id", RequirePerm("system:app:list"), func(c *gin.Context) {
		resp, e := svc.AppGet(pathID(c, "id"), currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	apps.POST("", RequirePerm("system:app:create"), func(c *gin.Context) {
		var req CreateAppRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.AppCreate(req, currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	apps.PUT("/:id", RequirePerm("system:app:edit"), func(c *gin.Context) {
		var req UpdateAppRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.AppUpdate(pathID(c, "id"), req, currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	apps.DELETE("/:id", RequirePerm("system:app:delete"), func(c *gin.Context) {
		if e := svc.AppDelete(pathID(c, "id"), currentUID(c)); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	apps.GET("/:id/keys", RequirePerm("system:app:key-manage"), func(c *gin.Context) {
		resp, e := svc.AppListKeys(pathID(c, "id"), currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	apps.POST("/:id/keys", RequirePerm("system:app:key-manage"), func(c *gin.Context) {
		var req CreateAppKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.AppCreateKey(pathID(c, "id"), req, currentUID(c))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	apps.DELETE("/:id/keys/:keyId", RequirePerm("system:app:key-manage"), func(c *gin.Context) {
		if e := svc.AppDeleteKey(pathID(c, "id"), pathID(c, "keyId"), currentUID(c)); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})

	// ===== 绑定（宿主侧） =====
	bind := api.Group("/bind")
	bind.GET("", RequirePerm("account:bind:list"), func(c *gin.Context) {
		OK(c, svc.BindListMine(currentUID(c)))
	})
	bind.POST("/code", RequirePerm("account:bind:create"), func(c *gin.Context) {
		var req CreateBindCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.BindCreateUserInitiatedCode(currentUID(c), req.AppID)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	bind.POST("/confirm", RequirePerm("account:bind:create"), func(c *gin.Context) {
		var req BindConfirmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.BindConfirmAppInitiated(currentUID(c), req.Code)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	bind.DELETE("/:id", RequirePerm("account:bind:delete"), func(c *gin.Context) {
		if e := svc.BindCancel(currentUID(c), pathID(c, "id")); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})
	bind.PUT("/:id/default", RequirePerm("account:bind:set-default"), func(c *gin.Context) {
		if e := svc.BindSetDefault(currentUID(c), pathID(c, "id")); e != nil {
			Fail(c, e)
			return
		}
		OK(c, nil)
	})

	// ===== 开放接口（应用签名鉴权，AppSignatureMiddleware 已写入 AppPrincipal） =====
	open := api.Group("/open")
	open.POST("/bind/apply", func(c *gin.Context) {
		var req OpenBindApplyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.OpenApplyBind(currentPrincipal(c), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	open.POST("/bind/claim", func(c *gin.Context) {
		var req OpenBindClaimRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, BadRequest("参数错误"))
			return
		}
		resp, e := svc.OpenClaimBind(currentPrincipal(c), req)
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})
	open.GET("/userinfo", func(c *gin.Context) {
		resp, e := svc.OpenUserInfo(currentPrincipal(c), c.Query("appUserId"))
		if e != nil {
			Fail(c, e)
			return
		}
		OK(c, resp)
	})

	// 路由装配完成后校验接口元数据是否漏登记（只打日志，不影响启动）
	checkApiRegistry(r)
}
