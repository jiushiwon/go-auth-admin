package main

import (
	"time"

	"gorm.io/gorm"
)

// ===== 基础 RBAC 实体 =====
// 表名与 java-vue-admin 完全一致（wg_ 前缀），字段用 snake_case 对齐 @Column。

type SysUser struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	Username    string         `gorm:"column:username;size:64;not null;uniqueIndex" json:"username"`
	Password    string         `gorm:"column:password;size:100;not null" json:"-"` // 只写不读
	Nickname    string         `gorm:"column:nickname;size:64" json:"nickname"`
	Email       string         `gorm:"column:email;size:128" json:"email"`
	Phone       string         `gorm:"column:phone;size:20" json:"phone"`
	Avatar      string         `gorm:"column:avatar;size:255" json:"avatar"`
	TenantID    int64          `gorm:"column:tenant_id;not null;default:1" json:"tenantId"`
	OrgID       *int64         `gorm:"column:org_id" json:"orgId"`
	Status      int            `gorm:"column:status;not null;default:1" json:"status"`
	LastLoginAt *time.Time     `gorm:"column:last_login_at" json:"lastLoginAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (SysUser) TableName() string { return "wg_sys_user" }

type SysRole struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	TenantID    int64          `gorm:"column:tenant_id;not null;default:1" json:"tenantId"`
	Name        string         `gorm:"column:name;size:50;not null" json:"name"`
	Code        string         `gorm:"column:code;size:50;not null;uniqueIndex" json:"code"`
	Description string         `gorm:"column:description;size:255" json:"description"`
	DataScope   string         `gorm:"column:data_scope;size:20;not null;default:ALL" json:"dataScope"`
	SortOrder   int            `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	Status      int            `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (SysRole) TableName() string { return "wg_sys_role" }

type SysMenu struct {
	ID         int64          `gorm:"primaryKey" json:"id"`
	TenantID   int64          `gorm:"column:tenant_id;not null;default:1" json:"tenantId"`
	ParentID   *int64         `gorm:"column:parent_id" json:"parentId"`
	Name       string         `gorm:"column:name;size:50;not null" json:"name"`
	Path       string         `gorm:"column:path;size:255" json:"path"`
	Component  string         `gorm:"column:component;size:255" json:"component"`
	MenuType   string         `gorm:"column:menu_type;size:1;not null;default:M" json:"menuType"`
	Icon       string         `gorm:"column:icon;size:50" json:"icon"`
	Permission *string        `gorm:"column:permission;size:100" json:"permission"`
	SortOrder  int            `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	Visible    int            `gorm:"column:visible;not null;default:1" json:"visible"`
	Status     int            `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (SysMenu) TableName() string { return "wg_sys_menu" }

type SysOrg struct {
	ID           int64          `gorm:"primaryKey" json:"id"`
	TenantID     int64          `gorm:"column:tenant_id;not null;default:1" json:"tenantId"`
	ParentID     *int64         `gorm:"column:parent_id" json:"parentId"`
	Name         string         `gorm:"column:name;size:100;not null" json:"name"`
	SortOrder    int            `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	LeaderUserID *int64         `gorm:"column:leader_user_id" json:"leaderUserId"`
	Phone        string         `gorm:"column:phone;size:20" json:"phone"`
	Email        string         `gorm:"column:email;size:128" json:"email"`
	Status       int            `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (SysOrg) TableName() string { return "wg_sys_org" }

type SysPost struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	TenantID  int64          `gorm:"column:tenant_id;not null;default:1" json:"tenantId"`
	Name      string         `gorm:"column:name;size:50;not null" json:"name"`
	Code      string         `gorm:"column:code;size:50;not null;uniqueIndex" json:"code"`
	SortOrder int            `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	Status    int            `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (SysPost) TableName() string { return "wg_sys_post" }

type SysTenant struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"column:name;size:100;not null" json:"name"`
	Code      string         `gorm:"column:code;size:50;not null;uniqueIndex" json:"code"`
	Status    int            `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (SysTenant) TableName() string { return "wg_sys_tenant" }

// ===== 关联表（物理删除，无软删） =====

type SysUserRole struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null;index" json:"userId"`
	RoleID    int64     `gorm:"column:role_id;not null;index" json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SysUserRole) TableName() string { return "wg_sys_user_role" }

type SysUserPost struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null;index" json:"userId"`
	PostID    int64     `gorm:"column:post_id;not null;index" json:"postId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SysUserPost) TableName() string { return "wg_sys_user_post" }

type SysRoleMenu struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	RoleID    int64     `gorm:"column:role_id;not null;index" json:"roleId"`
	MenuID    int64     `gorm:"column:menu_id;not null;index" json:"menuId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SysRoleMenu) TableName() string { return "wg_sys_role_menu" }

// ===== universal-login（应用 / 绑定 / 开放 API） =====

type App struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	AppName     string         `gorm:"column:app_name;size:100;not null" json:"appName"`
	AppKey      string         `gorm:"column:app_key;size:64;not null;uniqueIndex" json:"appKey"`
	Description string         `gorm:"column:description;size:500" json:"description"`
	Logo        string         `gorm:"column:logo;size:500" json:"logo"`
	CallbackURL string         `gorm:"column:callback_url;size:500" json:"callbackUrl"`
	Status      int            `gorm:"column:status;not null;default:1" json:"status"`
	OwnerID     int64          `gorm:"column:owner_id;not null" json:"ownerId"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (App) TableName() string { return "wg_sys_app" }

type AppKey struct {
	ID         int64          `gorm:"primaryKey" json:"id"`
	AppID      int64          `gorm:"column:app_id;not null;index" json:"appId"`
	KeyName    string         `gorm:"column:key_name;size:64" json:"keyName"`
	APIKey     string         `gorm:"column:api_key;size:64;not null;uniqueIndex" json:"apiKey"`
	APISecret  string         `gorm:"column:api_secret;size:128;not null" json:"-"`
	Status     int            `gorm:"column:status;not null;default:1" json:"status"`
	LastUsedAt *time.Time     `gorm:"column:last_used_at" json:"lastUsedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (AppKey) TableName() string { return "wg_sys_app_key" }

type AppBinding struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	SysUserID   int64          `gorm:"column:sys_user_id;not null;index" json:"sysUserId"`
	AppID       int64          `gorm:"column:app_id;not null;index" json:"appId"`
	AppUserID   string         `gorm:"column:app_user_id;size:128" json:"appUserId"`
	AppUserName string         `gorm:"column:app_user_name;size:128" json:"appUserName"`
	BindType    string         `gorm:"column:bind_type;size:20" json:"bindType"`
	IsDefault   int            `gorm:"column:is_default;not null;default:0" json:"isDefault"`
	BindAt      *time.Time     `gorm:"column:bind_at" json:"bindAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (AppBinding) TableName() string { return "wg_sys_app_binding" }

type BindCode struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"column:code;size:64;not null;uniqueIndex" json:"code"`
	Direction   string         `gorm:"column:direction;size:20;not null" json:"direction"`
	AppID       int64          `gorm:"column:app_id;not null;index" json:"appId"`
	SysUserID   *int64         `gorm:"column:sys_user_id" json:"sysUserId"`
	AppUserID   string         `gorm:"column:app_user_id;size:128" json:"appUserId"`
	AppUserName string         `gorm:"column:app_user_name;size:128" json:"appUserName"`
	Status      int            `gorm:"column:status;not null;default:0" json:"status"`
	ExpireAt    time.Time      `gorm:"column:expire_at;not null" json:"expireAt"`
	UsedAt      *time.Time     `gorm:"column:used_at" json:"usedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (BindCode) TableName() string { return "wg_sys_bind_code" }

// BindCode 状态常量（与 java 一致）
const (
	BindDirectionAppInitiated  = "app_initiated"
	BindDirectionUserInitiated = "user_initiated"

	BindStatusPending = 0
	BindStatusUsed    = 1
	BindStatusExpired = 2
)

func (b BindCode) Usable() bool {
	return b.Status == BindStatusPending && b.ExpireAt.After(time.Now())
}

// AppApiNonce 开放接口签名防重放：记录已消费的 (app_key_id, nonce)。
// 与 X-Timestamp 300s 窗口配合，杜绝同签名在窗口内无限重放。
type AppApiNonce struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AppKeyID  int64     `gorm:"column:app_key_id;not null;index" json:"appKeyId"`
	Nonce     string    `gorm:"column:nonce;size:64;not null;uniqueIndex" json:"nonce"`
	CreatedAt time.Time `json:"createdAt"`
}

func (AppApiNonce) TableName() string { return "wg_sys_app_api_nonce" }
