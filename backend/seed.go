package main

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// Seed 复刻 DataInitializer：首次启动（无用户时）注入租户/组织/岗位/角色/菜单/用户与关联。
func Seed(db *gorm.DB) {
	var cnt int64
	db.Model(&SysUser{}).Count(&cnt)
	if cnt > 0 {
		return
	}

	now := time.Now()

	// 1. 租户
	tenant := SysTenant{Name: "默认租户", Code: "default", Status: 1}
	db.Create(&tenant)

	// 2. 组织
	rootOrg := saveOrg(db, nil, "总公司", 1, now)
	devOrg := saveOrg(db, &rootOrg.ID, "研发部", 1, now)
	saveOrg(db, &rootOrg.ID, "市场部", 2, now)
	saveOrg(db, &rootOrg.ID, "财务部", 3, now)

	// 3. 岗位
	techLead := savePost(db, "tech_leader", "技术总监", 1, now)
	frontend := savePost(db, "frontend_dev", "前端开发", 2, now)
	savePost(db, "backend_dev", "后端开发", 3, now)
	product := savePost(db, "product_manager", "产品经理", 4, now)

	// 4. 角色
	superRole := saveRole(db, "super_admin", "超级管理员", "ALL", "系统内置最高权限角色", 1, now)
	userAdminRole := saveRole(db, "user_admin", "用户管理员", "ALL", "负责用户与权限管理", 2, now)
	commonRole := saveRole(db, "common_user", "普通用户", "SELF_ONLY", "仅可访问仪表盘", 3, now)

	// 5. 菜单（目录 M / 页面 C / 按钮 F）
	dashboardDir := saveMenu(db, nil, "仪表盘", "", "", "M", "layout-dashboard", "", 1, now)
	systemDir := saveMenu(db, nil, "系统管理", "", "", "M", "settings", "", 2, now)
	accountDir := saveMenu(db, nil, "账户中心", "", "", "M", "link-2", "", 3, now)

	dashboardPage := saveMenu(db, &dashboardDir.ID, "仪表盘", "/dashboard", "Dashboard", "C", "layout-dashboard", "dashboard:home:view", 1, now)
	userPage := saveMenu(db, &systemDir.ID, "用户管理", "/system/user", "system/user/index", "C", "users", "system:user:list", 1, now)
	rolePage := saveMenu(db, &systemDir.ID, "角色管理", "/system/role", "system/role/index", "C", "shield-check", "system:role:list", 2, now)
	menuPage := saveMenu(db, &systemDir.ID, "菜单管理", "/system/menu", "system/menu/index", "C", "settings", "system:menu:list", 3, now)
	orgPage := saveMenu(db, &systemDir.ID, "组织管理", "/system/org", "system/org/index", "C", "building-2", "system:org:list", 4, now)
	postPage := saveMenu(db, &systemDir.ID, "岗位管理", "/system/post", "system/post/index", "C", "briefcase", "system:post:list", 5, now)
	appPage := saveMenu(db, &systemDir.ID, "应用管理", "/system/app", "system/app/index", "C", "plug", "system:app:list", 6, now)
	bindPage := saveMenu(db, &accountDir.ID, "应用绑定", "/account/bind", "account/bind/index", "C", "link-2", "account:bind:list", 1, now)

	// 用户按钮
	saveMenu(db, &userPage.ID, "新增用户", "", "", "F", "", "system:user:create", 1, now)
	saveMenu(db, &userPage.ID, "编辑用户", "", "", "F", "", "system:user:edit", 2, now)
	saveMenu(db, &userPage.ID, "删除用户", "", "", "F", "", "system:user:delete", 3, now)
	saveMenu(db, &userPage.ID, "重置密码", "", "", "F", "", "system:user:reset-pwd", 4, now)
	saveMenu(db, &userPage.ID, "分配角色", "", "", "F", "", "system:user:assign-role", 5, now)
	saveMenu(db, &userPage.ID, "分配岗位", "", "", "F", "", "system:user:assign-post", 6, now)

	// 角色按钮
	saveMenu(db, &rolePage.ID, "新增角色", "", "", "F", "", "system:role:create", 1, now)
	saveMenu(db, &rolePage.ID, "编辑角色", "", "", "F", "", "system:role:edit", 2, now)
	saveMenu(db, &rolePage.ID, "删除角色", "", "", "F", "", "system:role:delete", 3, now)
	saveMenu(db, &rolePage.ID, "分配菜单", "", "", "F", "", "system:role:assign-menu", 4, now)

	// 菜单按钮
	saveMenu(db, &menuPage.ID, "新增菜单", "", "", "F", "", "system:menu:create", 1, now)
	saveMenu(db, &menuPage.ID, "编辑菜单", "", "", "F", "", "system:menu:edit", 2, now)
	saveMenu(db, &menuPage.ID, "删除菜单", "", "", "F", "", "system:menu:delete", 3, now)

	// 组织按钮
	saveMenu(db, &orgPage.ID, "新增组织", "", "", "F", "", "system:org:create", 1, now)
	saveMenu(db, &orgPage.ID, "编辑组织", "", "", "F", "", "system:org:edit", 2, now)
	saveMenu(db, &orgPage.ID, "删除组织", "", "", "F", "", "system:org:delete", 3, now)

	// 岗位按钮
	saveMenu(db, &postPage.ID, "新增岗位", "", "", "F", "", "system:post:create", 1, now)
	saveMenu(db, &postPage.ID, "编辑岗位", "", "", "F", "", "system:post:edit", 2, now)
	saveMenu(db, &postPage.ID, "删除岗位", "", "", "F", "", "system:post:delete", 3, now)

	// 应用按钮
	saveMenu(db, &appPage.ID, "新增应用", "", "", "F", "", "system:app:create", 1, now)
	saveMenu(db, &appPage.ID, "编辑应用", "", "", "F", "", "system:app:edit", 2, now)
	saveMenu(db, &appPage.ID, "删除应用", "", "", "F", "", "system:app:delete", 3, now)
	saveMenu(db, &appPage.ID, "密钥管理", "", "", "F", "", "system:app:key-manage", 4, now)

	// 绑定按钮
	saveMenu(db, &bindPage.ID, "发起绑定", "", "", "F", "", "account:bind:create", 1, now)
	saveMenu(db, &bindPage.ID, "删除绑定", "", "", "F", "", "account:bind:delete", 2, now)
	saveMenu(db, &bindPage.ID, "设为默认", "", "", "F", "", "account:bind:set-default", 3, now)

	// 6. 用户（首次启动随机生成初始密码，打印到日志，请尽快修改）
	initialPwd := randomHex(12)
	hp, _ := hashPassword(initialPwd)
	log.Printf("[seed] 初始账户密码（admin / user_admin / demo 均为）：%s", initialPwd)
	admin := saveUser(db, "admin", hp, "超级管理员", "admin@vue-admin.local", "13800000001", &rootOrg.ID, now)
	userAdmin := saveUser(db, "user_admin", hp, "用户管理员", "user_admin@vue-admin.local", "13800000002", &rootOrg.ID, now)
	demo := saveUser(db, "demo", hp, "演示账号", "demo@vue-admin.local", "13800000003", &devOrg.ID, now)

	// 7. 用户-角色
	saveUserRole(db, admin.ID, superRole.ID)
	saveUserRole(db, userAdmin.ID, userAdminRole.ID)
	saveUserRole(db, demo.ID, commonRole.ID)

	// 8. 用户-岗位
	saveUserPost(db, admin.ID, techLead.ID)
	saveUserPost(db, userAdmin.ID, product.ID)
	saveUserPost(db, demo.ID, frontend.ID)

	// 9. 角色-菜单
	var allMenus []SysMenu
	db.Find(&allMenus)
	for _, m := range allMenus {
		saveRoleMenu(db, superRole.ID, m.ID)
	}

	// user_admin：仪表盘 + 系统管理（不含菜单管理 + 应用管理）+ 账户绑定
	userAdminMenus := []int64{
		dashboardDir.ID, dashboardPage.ID, systemDir.ID, userPage.ID, rolePage.ID, orgPage.ID, postPage.ID,
		accountDir.ID, bindPage.ID,
	}
	for _, id := range []int64{userPage.ID, rolePage.ID, orgPage.ID, postPage.ID, bindPage.ID} {
		_ = id
	}
	// 追加对应按钮（从全量菜单筛：parent 为上述页面 且 非目录/页面）
	pageParents := map[int64]bool{dashboardPage.ID: true, userPage.ID: true, rolePage.ID: true, orgPage.ID: true, postPage.ID: true, bindPage.ID: true}
	for _, m := range allMenus {
		if m.ParentID != nil && pageParents[*m.ParentID] {
			userAdminMenus = append(userAdminMenus, m.ID)
		}
	}
	for _, id := range userAdminMenus {
		saveRoleMenu(db, userAdminRole.ID, id)
	}

	// common_user：仪表盘 + 自己的绑定
	commonMenus := []int64{dashboardDir.ID, dashboardPage.ID, accountDir.ID, bindPage.ID}
	for _, m := range allMenus {
		if m.ParentID != nil && *m.ParentID == bindPage.ID {
			commonMenus = append(commonMenus, m.ID)
		}
	}
	for _, id := range commonMenus {
		saveRoleMenu(db, commonRole.ID, id)
	}

	// 10. demo 应用 + 一对 apiKey（便于开放接口 e2e 与本地联调；密钥仅此处打印一次）
	demoAppKey := seedGenUniqueAppKey(db)
	demoAPIKey := seedGenUniqueAPIKey(db)
	demoAPISecret := "sk_" + randomHex(32)
	demoApp := App{
		AppName:     "演示应用",
		AppKey:      demoAppKey,
		Description: "seed 阶段内置的 demo 应用，供开放接口 e2e / 本地联调使用",
		Status:      1,
		OwnerID:     admin.ID,
		CreatedAt:   now,
	}
	db.Create(&demoApp)
	demoKey := AppKey{
		AppID:     demoApp.ID,
		KeyName:   "默认密钥",
		APIKey:    demoAPIKey,
		APISecret: demoAPISecret,
		Status:    1,
		CreatedAt: now,
	}
	db.Create(&demoKey)
	log.Printf("[seed] demo 应用已创建（owner=admin）  app_id=%d  app_key=%s", demoApp.ID, demoApp.AppKey)
	log.Printf("[seed] demo 应用密钥（APIKey=%s APISecret=%s —— 关闭日志窗口后无法再查）", demoAPIKey, demoAPISecret)
}

// seedGenUniqueAppKey 与 service_app.go 的 genUniqueAppKey 等价；Seed 是自由函数不能调方法。
func seedGenUniqueAppKey(db *gorm.DB) string {
	for {
		k := "app_" + randomHex(12)
		var cnt int64
		db.Model(&App{}).Where("app_key = ?", k).Count(&cnt)
		if cnt == 0 {
			return k
		}
	}
}

func seedGenUniqueAPIKey(db *gorm.DB) string {
	for {
		k := "ak_" + randomHex(12)
		var cnt int64
		db.Model(&AppKey{}).Where("api_key = ?", k).Count(&cnt)
		if cnt == 0 {
			return k
		}
	}
}

func saveOrg(db *gorm.DB, parentID *int64, name string, sort int, now time.Time) SysOrg {
	o := SysOrg{ParentID: parentID, Name: name, SortOrder: sort, Status: 1, TenantID: 1, CreatedAt: now, UpdatedAt: now}
	db.Create(&o)
	return o
}

func savePost(db *gorm.DB, code, name string, sort int, now time.Time) SysPost {
	p := SysPost{Code: code, Name: name, SortOrder: sort, Status: 1, TenantID: 1, CreatedAt: now, UpdatedAt: now}
	db.Create(&p)
	return p
}

func saveRole(db *gorm.DB, code, name, dataScope, desc string, sort int, now time.Time) SysRole {
	r := SysRole{Code: code, Name: name, DataScope: dataScope, Description: desc, SortOrder: sort, Status: 1, TenantID: 1, CreatedAt: now, UpdatedAt: now}
	db.Create(&r)
	return r
}

func saveMenu(db *gorm.DB, parentID *int64, name, path, component, menuType, icon, permission string, sort int, now time.Time) SysMenu {
	m := SysMenu{ParentID: parentID, Name: name, Path: path, Component: component, MenuType: menuType,
		Icon: icon, Permission: strPtr(permission), SortOrder: sort, Visible: 1, Status: 1, TenantID: 1, CreatedAt: now, UpdatedAt: now}
	db.Create(&m)
	return m
}

func saveUser(db *gorm.DB, username, password, nickname, email, phone string, orgID *int64, now time.Time) SysUser {
	u := SysUser{Username: username, Password: password, Nickname: nickname, Email: email, Phone: phone,
		TenantID: 1, OrgID: orgID, Status: 1, CreatedAt: now, UpdatedAt: now}
	db.Create(&u)
	return u
}

func saveUserRole(db *gorm.DB, uid, rid int64) {
	db.Create(&SysUserRole{UserID: uid, RoleID: rid, CreatedAt: time.Now()})
}

func saveUserPost(db *gorm.DB, uid, pid int64) {
	db.Create(&SysUserPost{UserID: uid, PostID: pid, CreatedAt: time.Now()})
}

func saveRoleMenu(db *gorm.DB, rid, mid int64) {
	db.Create(&SysRoleMenu{RoleID: rid, MenuID: mid, CreatedAt: time.Now()})
}
