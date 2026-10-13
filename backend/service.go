package main

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ===== 请求 / 响应 VO =====

type LoginResponseVO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type UserInfoVO struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Avatar      string   `json:"avatar"`
	TenantID    int64    `json:"tenantId"`
	OrgID       *int64   `json:"orgId"`
	Status      int      `json:"status"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// 注意：与 Java UserInfoResponse 逐字对齐，此处不含 orgName（Java 不返回该字段）。

type MenuNodeVO struct {
	ID         int64        `json:"id"`
	ParentID   *int64       `json:"parentId"`
	Name       string       `json:"name"`
	Path       string       `json:"path"`
	Component  string       `json:"component"`
	MenuType   string       `json:"menuType"`
	Icon       string       `json:"icon"`
	Permission *string      `json:"permission"`
	SortOrder  int          `json:"sortOrder"`
	Visible    int          `json:"visible"`
	Status     int          `json:"status"`
	Children   []MenuNodeVO `json:"children"`
}

type OrgNodeVO struct {
	ID           int64       `json:"id"`
	ParentID     *int64      `json:"parentId"`
	Name         string      `json:"name"`
	SortOrder    int         `json:"sortOrder"`
	LeaderUserID *int64      `json:"leaderUserId"`
	Phone        string      `json:"phone"`
	Email        string      `json:"email"`
	Status       int         `json:"status"`
	Children     []OrgNodeVO `json:"children"`
}

type SimpleRefVO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type UserVO struct {
	ID        int64         `json:"id"`
	Username  string        `json:"username"`
	Nickname  string        `json:"nickname"`
	Email     string        `json:"email"`
	Phone     string        `json:"phone"`
	Avatar    string        `json:"avatar"`
	TenantID  int64         `json:"tenantId"`
	OrgID     *int64        `json:"orgId"`
	OrgName   string        `json:"orgName"`
	Status    int           `json:"status"`
	CreatedAt string        `json:"createdAt"`
	UpdatedAt string        `json:"updatedAt"`
	Roles     []SimpleRefVO `json:"roles"`
	Posts     []SimpleRefVO `json:"posts"`
}

type RoleVO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	DataScope   string `json:"dataScope"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status"`
	MenuCount   int64  `json:"menuCount"`
	CreatedAt   string `json:"createdAt"`
}

type PostVO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	TenantID  int64  `json:"tenantId"`
	SortOrder int    `json:"sortOrder"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type TenantVO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type AppVO struct {
	ID          int64  `json:"id"`
	AppName     string `json:"appName"`
	AppKey      string `json:"appKey"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	CallbackURL string `json:"callbackUrl"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type AppKeyVO struct {
	ID         int64  `json:"id"`
	KeyName    string `json:"keyName"`
	APIKey     string `json:"apiKey"`
	Status     int    `json:"status"`
	LastUsedAt string `json:"lastUsedAt"`
	CreatedAt  string `json:"createdAt"`
}

type CreatedAppKeyVO struct {
	ID        int64  `json:"id"`
	KeyName   string `json:"keyName"`
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type BindVO struct {
	ID          int64  `json:"id"`
	AppID       int64  `json:"appId"`
	AppUserID   string `json:"appUserId"`
	AppUserName string `json:"appUserName"`
	BindType    string `json:"bindType"`
	IsDefault   bool   `json:"isDefault"`
	BindAt      string `json:"bindAt"`
	AppKey      string `json:"appKey"`
	AppName     string `json:"appName"`
	Logo        string `json:"logo"`
}

type BindCodeVO struct {
	Code          string `json:"code"`
	Direction     string `json:"direction"`
	ExpireSeconds int64  `json:"expireSeconds"`
	QrContent     string `json:"qrContent"`
	ExpireAt      string `json:"expireAt"`
}

type OpenUserInfoVO struct {
	AppUserID string `json:"appUserId"`
	Bound     bool   `json:"bound"`
	BindingID int64  `json:"bindingId,omitempty"`
	UserID    int64  `json:"userId,omitempty"`
	Username  string `json:"username,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// 请求结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Nickname string  `json:"nickname"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Avatar   string  `json:"avatar"`
	TenantID int64   `json:"tenantId"`
	OrgID    *int64  `json:"orgId"`
	Status   int     `json:"status"`
	RoleIDs  []int64 `json:"roleIds"`
	PostIDs  []int64 `json:"postIds"`
}
type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	OrgID    *int64 `json:"orgId"`
	// Status 用指针：前端不传 → nil（不改）；传 0（禁用）/1（启用）均生效，
	// 修复原先 `if Status != 0` 把合法禁用值静默吞掉的问题。
	// 注意：密码字段已移除——改密码只能走专用 reset-pwd 接口，避免 user:edit 越权改密。
	Status *int `json:"status"`
}
type AssignRolesRequest struct {
	RoleIDs []int64 `json:"roleIds"`
}
type AssignPostsRequest struct {
	PostIDs []int64 `json:"postIds"`
}
type AssignMenusRequest struct {
	MenuIDs []int64 `json:"menuIds"`
}
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Description string  `json:"description"`
	DataScope   string  `json:"dataScope"`
	SortOrder   int     `json:"sortOrder"`
	Status      int     `json:"status"`
	MenuIDs     []int64 `json:"menuIds"`
}
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataScope   string `json:"dataScope"`
	SortOrder   int    `json:"sortOrder"`
	Status      *int   `json:"status"`
}
type CreateMenuRequest struct {
	ParentID   *int64 `json:"parentId"`
	Name       string `json:"name" binding:"required"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	MenuType   string `json:"menuType"`
	Icon       string `json:"icon"`
	Permission string `json:"permission"`
	SortOrder  int    `json:"sortOrder"`
	Visible    int    `json:"visible"`
	Status     int    `json:"status"`
	TenantID   int64  `json:"tenantId"`
}
type UpdateMenuRequest struct {
	ParentID   *int64 `json:"parentId"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	MenuType   string `json:"menuType"`
	Icon       string `json:"icon"`
	Permission string `json:"permission"`
	SortOrder  int    `json:"sortOrder"`
	Visible    int    `json:"visible"`
	Status     *int   `json:"status"`
}
type CreateOrgRequest struct {
	ParentID     *int64 `json:"parentId"`
	Name         string `json:"name" binding:"required"`
	SortOrder    int    `json:"sortOrder"`
	LeaderUserID *int64 `json:"leaderUserId"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Status       *int   `json:"status"`
	TenantID     int64  `json:"tenantId"`
}
type CreatePostRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
	TenantID  int64  `json:"tenantId"`
}
type CreateTenantRequest struct {
	Name   string `json:"name" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Status *int   `json:"status"`
}
type CreateAppRequest struct {
	AppName     string `json:"appName"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	CallbackURL string `json:"callbackUrl"`
}
type UpdateAppRequest struct {
	AppName     string `json:"appName"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	CallbackURL string `json:"callbackUrl"`
	Status      *int   `json:"status"`
}
type CreateAppKeyRequest struct {
	KeyName string `json:"keyName"`
}
type CreateBindCodeRequest struct {
	AppID int64 `json:"appId" binding:"required"`
}
type BindConfirmRequest struct {
	Code string `json:"code" binding:"required"`
}
type OpenBindApplyRequest struct {
	AppUserID   string `json:"appUserId"`
	AppUserName string `json:"appUserName"`
}
type OpenBindClaimRequest struct {
	Code        string `json:"code" binding:"required"`
	AppUserID   string `json:"appUserId"`
	AppUserName string `json:"appUserName"`
}

// ===== Services 聚合依赖 =====

type Services struct {
	db   *gorm.DB
	repo *Repo
	jwt  *JWTManager
}

func NewServices(db *gorm.DB, repo *Repo, jwt *JWTManager) *Services {
	return &Services{db: db, repo: repo, jwt: jwt}
}

func hashPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func randomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// ===== AuthService =====

func (s *Services) Login(req LoginRequest) (*LoginResponseVO, *AppError) {
	var u SysUser
	if err := s.db.Where("username = ?", req.Username).First(&u).Error; err != nil {
		return nil, Unauthorized("用户名或密码错误")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return nil, Unauthorized("用户名或密码错误")
	}
	if u.Status == 0 {
		return nil, Forbidden("账号已禁用")
	}
	now := time.Now()
	s.db.Model(&u).Update("last_login_at", &now)
	return &LoginResponseVO{
		AccessToken:  s.jwt.GenerateAccessToken(u.ID, u.Username),
		RefreshToken: s.jwt.GenerateRefreshToken(u.ID, u.Username),
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwt.AccessExpire.Seconds()),
	}, nil
}

func (s *Services) Refresh(req RefreshRequest) (*LoginResponseVO, *AppError) {
	claims, err := s.jwt.Parse(req.RefreshToken)
	if err != nil {
		return nil, Unauthorized("刷新令牌无效或已过期")
	}
	if claims.Type != "refresh" {
		return nil, Unauthorized("令牌类型错误，需要 refreshToken")
	}
	var u SysUser
	if err := s.db.First(&u, claims.UID).Error; err != nil {
		return nil, Unauthorized("用户不存在")
	}
	if u.Status == 0 {
		return nil, Forbidden("账号已禁用")
	}
	return &LoginResponseVO{
		AccessToken:  s.jwt.GenerateAccessToken(u.ID, u.Username),
		RefreshToken: s.jwt.GenerateRefreshToken(u.ID, u.Username),
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwt.AccessExpire.Seconds()),
	}, nil
}

func (s *Services) Me(uid int64) (*UserInfoVO, *AppError) {
	var u SysUser
	if err := s.db.First(&u, uid).Error; err != nil {
		return nil, Unauthorized("用户不存在")
	}
	roles := s.repo.FindRoleCodesByUser(uid)
	perms := s.repo.FindPermissionsByUser(uid)
	return &UserInfoVO{
		ID: u.ID, Username: u.Username, Nickname: u.Nickname, Email: u.Email,
		Phone: u.Phone, Avatar: u.Avatar, TenantID: u.TenantID, OrgID: u.OrgID,
		Status: u.Status, Roles: roles, Permissions: perms,
	}, nil
}

func (s *Services) ChangePassword(uid int64, req ChangePasswordRequest) *AppError {
	var u SysUser
	if err := s.db.First(&u, uid).Error; err != nil {
		return Unauthorized("用户不存在")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.OldPassword)) != nil {
		return BadRequest("旧密码错误")
	}
	hp, err := hashPassword(req.NewPassword)
	if err != nil {
		return BadRequest("密码加密失败")
	}
	s.db.Model(&u).Update("password", hp)
	return nil
}

func (s *Services) GetMenus(uid int64) ([]MenuNodeVO, *AppError) {
	ids := s.repo.FindMenuIDsByUser(uid)
	if len(ids) == 0 {
		return []MenuNodeVO{}, nil
	}
	var menus []SysMenu
	s.db.Where("id IN ?", ids).Find(&menus)
	return buildMenuTree(menus), nil
}

func buildMenuTree(menus []SysMenu) []MenuNodeVO {
	// 自顶向下递归建树。
	// 不能用「迭代 + 先挂子再收根」的写法：Children 是值切片，挂进父节点的是
	// 当时的快照副本，之后往该子节点上挂的孙节点不会同步回副本 —— 三层结构
	// （目录→页面→按钮）下孙节点会静默丢失。递归保证子树完整后才拷进父节点。
	ids := make(map[int64]bool, len(menus))
	for i := range menus {
		ids[menus[i].ID] = true
	}
	byParent := map[int64][]SysMenu{}
	roots := make([]SysMenu, 0)
	for i := range menus {
		m := menus[i]
		switch {
		case m.ParentID == nil:
			roots = append(roots, m)
		case !ids[*m.ParentID]:
			roots = append(roots, m) // 父节点不存在：对齐 Java，降级为根节点
		default:
			byParent[*m.ParentID] = append(byParent[*m.ParentID], m)
		}
	}
	var build func([]SysMenu) []MenuNodeVO
	build = func(ms []SysMenu) []MenuNodeVO {
		out := make([]MenuNodeVO, 0, len(ms))
		for i := range ms {
			n := menuToNode(&ms[i])
			n.Children = build(byParent[ms[i].ID])
			out = append(out, n)
		}
		return out
	}
	res := build(roots)
	sortMenuTree(res)
	return res
}

func menuToNode(m *SysMenu) MenuNodeVO {
	return MenuNodeVO{
		ID: m.ID, ParentID: m.ParentID, Name: m.Name, Path: m.Path, Component: m.Component,
		MenuType: m.MenuType, Icon: m.Icon, Permission: m.Permission, SortOrder: m.SortOrder,
		Visible: m.Visible, Status: m.Status, Children: []MenuNodeVO{},
	}
}

func sortMenuTree(nodes []MenuNodeVO) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[j].SortOrder < nodes[i].SortOrder {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
	for i := range nodes {
		sortMenuTree(nodes[i].Children)
	}
}

// ===== UserService =====

// enforceUserScope 校验 targetUID 是否在 currentUID 的数据权限范围内。
// 对齐 java DataScopeFilter：SELF_ONLY 仅自己；DEPT_ONLY/DEPT_AND_BELOW 限本部门及子孙；ALL 不限制。
func (s *Services) enforceUserScope(currentUID, targetUID int64) *AppError {
	if currentUID == targetUID {
		return nil
	}
	switch s.repo.ResolveDataScope(currentUID) {
	case "ALL":
		return nil
	case "SELF_ONLY":
		return Forbidden("无权访问该用户（数据权限限制）")
	case "DEPT_ONLY", "DEPT_AND_BELOW":
		cur := s.repo.UserOrgID(currentUID)
		tgt := s.repo.UserOrgID(targetUID)
		if cur == nil || tgt == nil {
			return Forbidden("无权访问该用户（数据权限限制）")
		}
		for _, o := range s.repo.DescendantOrgIDs(*cur) {
			if o == *tgt {
				return nil
			}
		}
		return Forbidden("无权访问该用户（数据权限限制）")
	}
	return Forbidden("无权访问该用户（数据权限限制）")
}

func (s *Services) UserPage(page, pageSize int, username string, status *int, currentUID int64) (*PageResponse, *AppError) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	q := s.db.Model(&SysUser{})
	// 数据权限
	switch s.repo.ResolveDataScope(currentUID) {
	case "ALL":
		// 无额外限制
	case "SELF_ONLY":
		q = q.Where("id = ?", currentUID)
	case "DEPT_ONLY", "DEPT_AND_BELOW":
		orgID := s.repo.UserOrgID(currentUID)
		if orgID == nil {
			q = q.Where("id = ?", currentUID)
		} else {
			orgIDs := []int64{*orgID}
			if s.repo.ResolveDataScope(currentUID) == "DEPT_AND_BELOW" {
				orgIDs = s.repo.DescendantOrgIDs(*orgID)
			}
			q = q.Where("org_id IN ?", orgIDs)
		}
	}
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}

	var total int64
	q.Count(&total)
	var users []SysUser
	q.Order("id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	// org 名称映射
	orgIDs := map[int64]bool{}
	userIDs := make([]int64, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
		if u.OrgID != nil {
			orgIDs[*u.OrgID] = true
		}
	}
	orgNames := map[int64]string{}
	if len(orgIDs) > 0 {
		var orgs []SysOrg
		ids := make([]int64, 0, len(orgIDs))
		for id := range orgIDs {
			ids = append(ids, id)
		}
		s.db.Where("id IN ?", ids).Find(&orgs)
		for _, o := range orgs {
			orgNames[o.ID] = o.Name
		}
	}

	// 角色映射 userID → []SimpleRefVO（一次查，避免 N+1）
	rolesByUser := map[int64][]SimpleRefVO{}
	postsByUser := map[int64][]SimpleRefVO{}
	if len(userIDs) > 0 {
		// 角色：JOIN sys_role 取 code/name
		type roleRow struct {
			UserID int64  `gorm:"column:user_id"`
			RoleID int64  `gorm:"column:role_id"`
			Code   string `gorm:"column:code"`
			Name   string `gorm:"column:name"`
		}
		var rrows []roleRow
		s.db.Table("wg_sys_user_role ur").
			Select("ur.user_id AS user_id, r.id AS role_id, r.code AS code, r.name AS name").
			Joins("JOIN wg_sys_role r ON r.id = ur.role_id AND r.deleted_at IS NULL").
			Where("ur.user_id IN ?", userIDs).
			Find(&rrows)
		for _, r := range rrows {
			rolesByUser[r.UserID] = append(rolesByUser[r.UserID], SimpleRefVO{ID: r.RoleID, Code: r.Code, Name: r.Name})
		}
		// 岗位：JOIN sys_post 取 code/name
		type postRow struct {
			UserID int64  `gorm:"column:user_id"`
			PostID int64  `gorm:"column:post_id"`
			Code   string `gorm:"column:code"`
			Name   string `gorm:"column:name"`
		}
		var prows []postRow
		s.db.Table("wg_sys_user_post up").
			Select("up.user_id AS user_id, p.id AS post_id, p.code AS code, p.name AS name").
			Joins("JOIN wg_sys_post p ON p.id = up.post_id AND p.deleted_at IS NULL").
			Where("up.user_id IN ?", userIDs).
			Find(&prows)
		for _, p := range prows {
			postsByUser[p.UserID] = append(postsByUser[p.UserID], SimpleRefVO{ID: p.PostID, Code: p.Code, Name: p.Name})
		}
	}

	list := make([]UserVO, 0, len(users))
	for _, u := range users {
		var orgName string
		if u.OrgID != nil {
			orgName = orgNames[*u.OrgID]
		}
		list = append(list, userToVO(&u, orgName, rolesByUser[u.ID], postsByUser[u.ID]))
	}
	return &PageResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func userToVO(u *SysUser, orgName string, roles, posts []SimpleRefVO) UserVO {
	return UserVO{
		ID: u.ID, Username: u.Username, Nickname: u.Nickname, Email: u.Email, Phone: u.Phone,
		Avatar: u.Avatar, TenantID: u.TenantID, OrgID: u.OrgID, OrgName: orgName, Status: u.Status,
		CreatedAt: fmtTimeV(u.CreatedAt), UpdatedAt: fmtTimeV(u.UpdatedAt), Roles: roles, Posts: posts,
	}
}

func (s *Services) UserGetVO(id, currentUID int64) (*UserVO, *AppError) {
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return nil, e
	}
	return s.userGetVO(id, currentUID)
}

func (s *Services) userGetVO(id, currentUID int64) (*UserVO, *AppError) {
	var u SysUser
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, NotFound("用户不存在")
	}
	roleIDs, e := s.UserRoleIDs(id, currentUID)
	if e != nil {
		return nil, e
	}
	postIDs, e := s.UserPostIDs(id, currentUID)
	if e != nil {
		return nil, e
	}
	roles := s.roleRefs(roleIDs)
	posts := s.postRefs(postIDs)
	var orgName string
	if u.OrgID != nil {
		var o SysOrg
		if err := s.db.First(&o, *u.OrgID).Error; err == nil {
			orgName = o.Name
		}
	}
	vo := userToVO(&u, orgName, roles, posts)
	return &vo, nil
}

func (s *Services) UserCreate(req CreateUserRequest, currentUID int64) (*UserVO, *AppError) {
	var cnt int64
	s.db.Model(&SysUser{}).Where("username = ?", req.Username).Count(&cnt)
	if cnt > 0 {
		return nil, Conflict("用户名已存在")
	}
	hp, err := hashPassword(req.Password)
	if err != nil {
		return nil, BadRequest("密码加密失败")
	}
	u := SysUser{
		Username: req.Username, Password: hp, Nickname: req.Nickname, Email: req.Email,
		Phone: req.Phone, Avatar: req.Avatar, TenantID: req.TenantID, OrgID: req.OrgID, Status: req.Status,
	}
	if u.TenantID == 0 {
		u.TenantID = 1
	}
	if u.Status == 0 {
		u.Status = 1
	}
	if err := s.db.Create(&u).Error; err != nil {
		return nil, BadRequest("创建用户失败")
	}
	s.assignRolesInternal(u.ID, req.RoleIDs)
	s.assignPostsInternal(u.ID, req.PostIDs)
	return s.UserGetVO(u.ID, currentUID)
}

func (s *Services) UserUpdate(id, currentUID int64, req UpdateUserRequest) (*UserVO, *AppError) {
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return nil, e
	}
	var u SysUser
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, NotFound("用户不存在")
	}
	if req.Nickname != "" {
		u.Nickname = req.Nickname
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}
	if req.Avatar != "" {
		u.Avatar = req.Avatar
	}
	if req.OrgID != nil {
		u.OrgID = req.OrgID
	}
	if req.Status != nil {
		u.Status = *req.Status
	}
	s.db.Save(&u)
	return s.userGetVO(id, currentUID)
}

func (s *Services) UserDelete(id, currentUID int64) *AppError {
	if id == currentUID {
		return BadRequest("不能删除当前登录用户")
	}
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return e
	}
	var u SysUser
	if err := s.db.First(&u, id).Error; err != nil {
		return NotFound("用户不存在")
	}
	if u.Username == "admin" {
		return BadRequest("内置管理员不可删除")
	}
	now := time.Now()
	s.db.Model(&u).Update("deleted_at", &now)
	s.db.Where("user_id = ?", id).Delete(&SysUserRole{})
	s.db.Where("user_id = ?", id).Delete(&SysUserPost{})
	return nil
}

func (s *Services) UserAssignRoles(id, currentUID int64, req AssignRolesRequest) *AppError {
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return e
	}
	if id == currentUID {
		return BadRequest("不能修改自身角色")
	}
	// 禁止把超级管理员角色授予任何用户（仅种子内置，避免自提升提权）
	for _, rid := range req.RoleIDs {
		var r SysRole
		if err := s.db.First(&r, rid).Error; err != nil {
			return NotFound("角色不存在: " + itoa(rid))
		}
		if r.Code == RoleSuperAdmin {
			return Forbidden("禁止授予超级管理员角色")
		}
	}
	if err := s.db.First(&SysUser{}, id).Error; err != nil {
		return NotFound("用户不存在")
	}
	s.db.Where("user_id = ?", id).Delete(&SysUserRole{})
	s.assignRolesInternal(id, req.RoleIDs)
	return nil
}

func (s *Services) UserAssignPosts(id, currentUID int64, req AssignPostsRequest) *AppError {
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return e
	}
	if err := s.db.First(&SysUser{}, id).Error; err != nil {
		return NotFound("用户不存在")
	}
	s.db.Where("user_id = ?", id).Delete(&SysUserPost{})
	s.assignPostsInternal(id, req.PostIDs)
	return nil
}

func (s *Services) UserResetPassword(id, currentUID int64, newPassword string) *AppError {
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return e
	}
	var u SysUser
	if err := s.db.First(&u, id).Error; err != nil {
		return NotFound("用户不存在")
	}
	hp, err := hashPassword(newPassword)
	if err != nil {
		return BadRequest("密码加密失败")
	}
	s.db.Model(&u).Update("password", hp)
	return nil
}

func (s *Services) UserRoleIDs(id, currentUID int64) ([]int64, *AppError) {
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return nil, e
	}
	var ids []int64
	s.db.Model(&SysUserRole{}).Where("user_id = ?", id).Pluck("role_id", &ids)
	return ids, nil
}

func (s *Services) UserPostIDs(id, currentUID int64) ([]int64, *AppError) {
	if e := s.enforceUserScope(currentUID, id); e != nil {
		return nil, e
	}
	var ids []int64
	s.db.Model(&SysUserPost{}).Where("user_id = ?", id).Pluck("post_id", &ids)
	return ids, nil
}

func (s *Services) assignRolesInternal(uid int64, roleIDs []int64) {
	for _, rid := range roleIDs {
		if rid == 0 {
			continue
		}
		s.db.Create(&SysUserRole{UserID: uid, RoleID: rid})
	}
}

func (s *Services) assignPostsInternal(uid int64, postIDs []int64) {
	for _, pid := range postIDs {
		if pid == 0 {
			continue
		}
		s.db.Create(&SysUserPost{UserID: uid, PostID: pid})
	}
}

func (s *Services) roleRefs(ids []int64) []SimpleRefVO {
	if len(ids) == 0 {
		return []SimpleRefVO{}
	}
	var roles []SysRole
	s.db.Where("id IN ?", ids).Find(&roles)
	refs := make([]SimpleRefVO, 0, len(roles))
	for _, r := range roles {
		refs = append(refs, SimpleRefVO{ID: r.ID, Name: r.Name, Code: r.Code})
	}
	return refs
}

func (s *Services) postRefs(ids []int64) []SimpleRefVO {
	if len(ids) == 0 {
		return []SimpleRefVO{}
	}
	var posts []SysPost
	s.db.Where("id IN ?", ids).Find(&posts)
	refs := make([]SimpleRefVO, 0, len(posts))
	for _, p := range posts {
		refs = append(refs, SimpleRefVO{ID: p.ID, Name: p.Name, Code: p.Code})
	}
	return refs
}

// ===== RoleService =====

func (s *Services) RolePage(page, pageSize int, keyword string) (*PageResponse, *AppError) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := s.db.Model(&SysRole{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	var total int64
	q.Count(&total)
	var roles []SysRole
	q.Order("sort_order ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles)
	list := make([]RoleVO, 0, len(roles))
	for _, r := range roles {
		list = append(list, roleToVO(&r, s.repo.FindRoleMenuCount(r.ID)))
	}
	return &PageResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func roleToVO(r *SysRole, menuCount int64) RoleVO {
	return RoleVO{
		ID: r.ID, Name: r.Name, Code: r.Code, Description: r.Description, DataScope: r.DataScope,
		SortOrder: r.SortOrder, Status: r.Status, MenuCount: menuCount, CreatedAt: fmtTimeV(r.CreatedAt),
	}
}

func (s *Services) RoleGetVO(id int64) (*RoleVO, *AppError) {
	var r SysRole
	if err := s.db.First(&r, id).Error; err != nil {
		return nil, NotFound("角色不存在")
	}
	vo := roleToVO(&r, s.repo.FindRoleMenuCount(id))
	return &vo, nil
}

func (s *Services) RoleCreate(req CreateRoleRequest) (*RoleVO, *AppError) {
	var cnt int64
	s.db.Model(&SysRole{}).Where("code = ?", req.Code).Count(&cnt)
	if cnt > 0 {
		return nil, Conflict("角色编码已存在")
	}
	if req.Code == RoleSuperAdmin {
		return nil, Forbidden("禁止创建超级管理员角色")
	}
	scope := req.DataScope
	if scope == "" {
		scope = "ALL"
	}
	r := SysRole{Name: req.Name, Code: req.Code, Description: req.Description, DataScope: scope,
		SortOrder: req.SortOrder, Status: req.Status}
	if r.Status == 0 {
		r.Status = 1
	}
	s.db.Create(&r)
	s.assignMenusInternal(r.ID, req.MenuIDs)
	return s.RoleGetVO(r.ID)
}

func (s *Services) RoleUpdate(id int64, req UpdateRoleRequest) (*RoleVO, *AppError) {
	var r SysRole
	if err := s.db.First(&r, id).Error; err != nil {
		return nil, NotFound("角色不存在")
	}
	if req.Name != "" {
		r.Name = req.Name
	}
	if req.Description != "" {
		r.Description = req.Description
	}
	if req.DataScope != "" {
		r.DataScope = req.DataScope
	}
	if req.SortOrder != 0 {
		r.SortOrder = req.SortOrder
	}
	if req.Status != nil {
		r.Status = *req.Status
	}
	s.db.Save(&r)
	return s.RoleGetVO(id)
}

func (s *Services) RoleDelete(id int64) *AppError {
	var r SysRole
	if err := s.db.First(&r, id).Error; err != nil {
		return NotFound("角色不存在")
	}
	if r.Code == RoleSuperAdmin {
		return BadRequest("超级管理员角色不可删除")
	}
	var bound int64
	s.db.Model(&SysUserRole{}).Where("role_id = ?", id).Count(&bound)
	if bound > 0 {
		return Conflict("该角色已分配给 " + itoa(bound) + " 个用户，不可删除")
	}
	s.db.Where("role_id = ?", id).Delete(&SysRoleMenu{})
	s.db.Delete(&r)
	return nil
}

func (s *Services) RoleMenuIDs(id int64) []int64 {
	var ids []int64
	s.db.Model(&SysRoleMenu{}).Where("role_id = ?", id).Pluck("menu_id", &ids)
	return ids
}

func (s *Services) RoleAssignMenus(id int64, req AssignMenusRequest) *AppError {
	if err := s.db.First(&SysRole{}, id).Error; err != nil {
		return NotFound("角色不存在")
	}
	s.db.Where("role_id = ?", id).Delete(&SysRoleMenu{})
	s.assignMenusInternal(id, req.MenuIDs)
	return nil
}

func (s *Services) assignMenusInternal(roleID int64, menuIDs []int64) {
	for _, mid := range menuIDs {
		if mid == 0 {
			continue
		}
		s.db.Create(&SysRoleMenu{RoleID: roleID, MenuID: mid})
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

// ===== MenuService =====

var permPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*:[a-z][a-z0-9_-]+$`)

func (s *Services) MenuTree() []MenuNodeVO {
	var menus []SysMenu
	s.db.Order("sort_order ASC, id ASC").Find(&menus)
	return buildMenuTree(menus)
}

func (s *Services) MenuCreate(req CreateMenuRequest) (*MenuNodeVO, *AppError) {
	if err := s.validateMenu(req.MenuType, req.Path, req.Component, req.Permission, 0); err != nil {
		return nil, err
	}
	mt := req.MenuType
	if mt == "" {
		mt = "M"
	}
	visible, status := req.Visible, req.Status
	if visible == 0 {
		visible = 1
	}
	if status == 0 {
		status = 1
	}
	tenantID := req.TenantID
	if tenantID == 0 {
		tenantID = 1
	}
	m := SysMenu{ParentID: req.ParentID, Name: req.Name, Path: req.Path, Component: req.Component,
		MenuType: mt, Icon: req.Icon, Permission: strPtr(req.Permission), SortOrder: req.SortOrder,
		Visible: visible, Status: status, TenantID: tenantID}
	if err := s.db.Create(&m).Error; err != nil {
		return nil, BadRequest("创建菜单失败")
	}
	if e := s.assertMenuIntegrity(); e != nil {
		return nil, e
	}
	n := menuToNode(&m)
	return &n, nil
}

func (s *Services) MenuUpdate(id int64, req UpdateMenuRequest) (*MenuNodeVO, *AppError) {
	var m SysMenu
	if err := s.db.First(&m, id).Error; err != nil {
		return nil, NotFound("菜单不存在")
	}
	mt := req.MenuType
	if mt == "" {
		mt = m.MenuType
	}
	path := req.Path
	if path == "" {
		path = m.Path
	}
	perm := req.Permission
	if perm == "" {
		if m.Permission != nil {
			perm = *m.Permission
		}
	}
	if err := s.validateMenu(mt, path, m.Component, perm, id); err != nil {
		return nil, err
	}
	if req.ParentID != nil {
		m.ParentID = req.ParentID
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Component != "" {
		m.Component = req.Component
	}
	if req.Icon != "" {
		m.Icon = req.Icon
	}
	if req.SortOrder != 0 {
		m.SortOrder = req.SortOrder
	}
	if req.Visible != 0 {
		m.Visible = req.Visible
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	m.MenuType = mt
	m.Path = path
	m.Permission = strPtr(perm)
	s.db.Save(&m)
	if e := s.assertMenuIntegrity(); e != nil {
		return nil, e
	}
	n := menuToNode(&m)
	return &n, nil
}

func (s *Services) MenuDelete(id int64) *AppError {
	var m SysMenu
	if err := s.db.First(&m, id).Error; err != nil {
		return NotFound("菜单不存在")
	}
	var cnt int64
	s.db.Model(&SysMenu{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&cnt)
	if cnt > 0 {
		return Conflict("存在子菜单，不可删除")
	}
	now := time.Now()
	s.db.Model(&m).Update("deleted_at", &now)
	return nil
}

func (s *Services) validateMenu(menuType, path, component, permission string, selfID int64) *AppError {
	mt := strings.ToUpper(menuType)
	if mt == "" {
		mt = "M"
	}
	if mt != "M" && mt != "C" && mt != "F" && mt != "B" {
		return BadRequest("非法 menu_type: " + menuType + "（仅允许 M=目录 / C=菜单 / F=按钮）")
	}
	if mt == "C" {
		if path == "" || component == "" {
			return BadRequest("菜单类（menuType=C）必须填写 path 与 component")
		}
		var cnt int64
		if selfID == 0 {
			s.db.Model(&SysMenu{}).Where("path = ? AND deleted_at IS NULL", path).Count(&cnt)
		} else {
			s.db.Model(&SysMenu{}).Where("path = ? AND deleted_at IS NULL AND id <> ?", path, selfID).Count(&cnt)
		}
		if cnt > 0 {
			return Conflict("菜单路径已被占用：" + path)
		}
	} else if mt == "F" || mt == "B" {
		if permission == "" {
			return BadRequest("按钮类（menuType=F）必须填写权限码 permission")
		}
		if !permPattern.MatchString(permission) {
			return BadRequest("权限码必须三段式 模块:资源:动作（小写字母/数字/下划线）—— 当前值：" + permission)
		}
	}
	return nil
}

func (s *Services) assertMenuIntegrity() *AppError {
	var all []SysMenu
	s.db.Find(&all)
	ids := map[int64]bool{}
	for _, m := range all {
		ids[m.ID] = true
	}
	for _, m := range all {
		mt := strings.ToUpper(m.MenuType)
		if mt != "M" && mt != "C" && mt != "F" && mt != "B" {
			return Conflict("非法 menu_type: " + m.MenuType + "（只允许 M/C/F），id=" + itoa(m.ID))
		}
		if m.ParentID == nil {
			continue
		}
		if *m.ParentID == m.ID {
			return Conflict("菜单不能把自己作为父节点: id=" + itoa(m.ID))
		}
		if !ids[*m.ParentID] {
			return Conflict("菜单树存在孤儿节点: id=" + itoa(m.ID) + " parentId=" + itoa(*m.ParentID) + " 不存在")
		}
	}
	return nil
}

// ===== OrgService =====

func (s *Services) OrgTree() []OrgNodeVO {
	var orgs []SysOrg
	s.db.Order("sort_order ASC, id ASC").Find(&orgs)
	// 同 buildMenuTree：自顶向下递归建树，避免值切片快照导致子树丢失。
	ids := make(map[int64]bool, len(orgs))
	for i := range orgs {
		ids[orgs[i].ID] = true
	}
	byParent := map[int64][]SysOrg{}
	roots := make([]SysOrg, 0)
	for i := range orgs {
		o := orgs[i]
		switch {
		case o.ParentID == nil:
			roots = append(roots, o)
		case !ids[*o.ParentID]:
			roots = append(roots, o) // 父节点不存在：对齐 Java，降级为根节点
		default:
			byParent[*o.ParentID] = append(byParent[*o.ParentID], o)
		}
	}
	var build func([]SysOrg) []OrgNodeVO
	build = func(os []SysOrg) []OrgNodeVO {
		out := make([]OrgNodeVO, 0, len(os))
		for i := range os {
			o := os[i]
			out = append(out, OrgNodeVO{ID: o.ID, ParentID: o.ParentID, Name: o.Name, SortOrder: o.SortOrder,
				LeaderUserID: o.LeaderUserID, Phone: o.Phone, Email: o.Email, Status: o.Status,
				Children: build(byParent[o.ID])})
		}
		return out
	}
	res := build(roots)
	sortOrgTree(res)
	return res
}

func sortOrgTree(nodes []OrgNodeVO) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[j].SortOrder < nodes[i].SortOrder {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
	for i := range nodes {
		sortOrgTree(nodes[i].Children)
	}
}

func (s *Services) OrgCreate(req CreateOrgRequest) (*OrgNodeVO, *AppError) {
	tenantID := req.TenantID
	if tenantID == 0 {
		tenantID = 1
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	o := SysOrg{ParentID: req.ParentID, Name: req.Name, SortOrder: req.SortOrder,
		LeaderUserID: req.LeaderUserID, Phone: req.Phone, Email: req.Email, Status: status, TenantID: tenantID}
	if err := s.db.Create(&o).Error; err != nil {
		return nil, BadRequest("创建组织失败")
	}
	n := OrgNodeVO{ID: o.ID, ParentID: o.ParentID, Name: o.Name, SortOrder: o.SortOrder,
		LeaderUserID: o.LeaderUserID, Phone: o.Phone, Email: o.Email, Status: o.Status, Children: []OrgNodeVO{}}
	return &n, nil
}

func (s *Services) OrgUpdate(id int64, req CreateOrgRequest) (*OrgNodeVO, *AppError) {
	var o SysOrg
	if err := s.db.First(&o, id).Error; err != nil {
		return nil, NotFound("组织不存在")
	}
	if req.ParentID != nil {
		pid := *req.ParentID
		if pid == id {
			return nil, BadRequest("不能将组织挂到自身下")
		}
		// 禁止挂到自己的子孙下，否则组织树成环、两个组织在树中静默消失
		for _, d := range s.repo.DescendantOrgIDs(id) {
			if d == pid {
				return nil, BadRequest("不能将组织挂到其子孙组织下")
			}
		}
		o.ParentID = req.ParentID
	}
	if req.Name != "" {
		o.Name = req.Name
	}
	if req.SortOrder != 0 {
		o.SortOrder = req.SortOrder
	}
	if req.LeaderUserID != nil {
		o.LeaderUserID = req.LeaderUserID
	}
	if req.Phone != "" {
		o.Phone = req.Phone
	}
	if req.Email != "" {
		o.Email = req.Email
	}
	if req.Status != nil {
		o.Status = *req.Status
	}
	s.db.Save(&o)
	n := OrgNodeVO{ID: o.ID, ParentID: o.ParentID, Name: o.Name, SortOrder: o.SortOrder,
		LeaderUserID: o.LeaderUserID, Phone: o.Phone, Email: o.Email, Status: o.Status, Children: []OrgNodeVO{}}
	return &n, nil
}

func (s *Services) OrgDelete(id int64) *AppError {
	var o SysOrg
	if err := s.db.First(&o, id).Error; err != nil {
		return NotFound("组织不存在")
	}
	var cnt int64
	s.db.Model(&SysOrg{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&cnt)
	if cnt > 0 {
		return Conflict("存在子部门，不可删除")
	}
	now := time.Now()
	s.db.Model(&o).Update("deleted_at", &now)
	return nil
}

// ===== PostService =====

func (s *Services) PostList() []PostVO {
	var posts []SysPost
	s.db.Find(&posts)
	list := make([]PostVO, 0, len(posts))
	for _, p := range posts {
		list = append(list, postToVO(&p))
	}
	return list
}

func postToVO(p *SysPost) PostVO {
	return PostVO{ID: p.ID, Name: p.Name, Code: p.Code, TenantID: p.TenantID,
		SortOrder: p.SortOrder, Status: p.Status, CreatedAt: fmtTimeV(p.CreatedAt)}
}

func (s *Services) PostCreate(req CreatePostRequest) (*PostVO, *AppError) {
	var cnt int64
	s.db.Model(&SysPost{}).Where("code = ?", req.Code).Count(&cnt)
	if cnt > 0 {
		return nil, Conflict("岗位编码已存在")
	}
	tenantID := req.TenantID
	if tenantID == 0 {
		tenantID = 1
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	p := SysPost{Name: req.Name, Code: req.Code, SortOrder: req.SortOrder, Status: status, TenantID: tenantID}
	s.db.Create(&p)
	vo := postToVO(&p)
	return &vo, nil
}

func (s *Services) PostUpdate(id int64, req CreatePostRequest) (*PostVO, *AppError) {
	var p SysPost
	if err := s.db.First(&p, id).Error; err != nil {
		return nil, NotFound("岗位不存在")
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Code != "" {
		p.Code = req.Code
	}
	if req.SortOrder != 0 {
		p.SortOrder = req.SortOrder
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	if req.TenantID != 0 {
		p.TenantID = req.TenantID
	}
	s.db.Save(&p)
	vo := postToVO(&p)
	return &vo, nil
}

func (s *Services) PostDelete(id int64) *AppError {
	var p SysPost
	if err := s.db.First(&p, id).Error; err != nil {
		return NotFound("岗位不存在")
	}
	now := time.Now()
	s.db.Model(&p).Update("deleted_at", &now)
	return nil
}

// ===== TenantService =====

func (s *Services) TenantList() []TenantVO {
	var ts []SysTenant
	s.db.Find(&ts)
	list := make([]TenantVO, 0, len(ts))
	for _, t := range ts {
		list = append(list, TenantVO{ID: t.ID, Name: t.Name, Code: t.Code, Status: t.Status,
			CreatedAt: fmtTimeV(t.CreatedAt), UpdatedAt: fmtTimeV(t.UpdatedAt)})
	}
	return list
}

func (s *Services) TenantCreate(req CreateTenantRequest) (*TenantVO, *AppError) {
	var cnt int64
	s.db.Model(&SysTenant{}).Where("code = ?", req.Code).Count(&cnt)
	if cnt > 0 {
		return nil, Conflict("租户编码已存在")
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	t := SysTenant{Name: req.Name, Code: req.Code, Status: status}
	s.db.Create(&t)
	return &TenantVO{ID: t.ID, Name: t.Name, Code: t.Code, Status: t.Status,
		CreatedAt: fmtTimeV(t.CreatedAt), UpdatedAt: fmtTimeV(t.UpdatedAt)}, nil
}

func (s *Services) TenantUpdate(id int64, req CreateTenantRequest) (*TenantVO, *AppError) {
	var t SysTenant
	if err := s.db.First(&t, id).Error; err != nil {
		return nil, NotFound("租户不存在")
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Code != "" {
		t.Code = req.Code
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	s.db.Save(&t)
	return &TenantVO{ID: t.ID, Name: t.Name, Code: t.Code, Status: t.Status,
		CreatedAt: fmtTimeV(t.CreatedAt), UpdatedAt: fmtTimeV(t.UpdatedAt)}, nil
}

func (s *Services) TenantDelete(id int64) *AppError {
	var t SysTenant
	if err := s.db.First(&t, id).Error; err != nil {
		return NotFound("租户不存在")
	}
	now := time.Now()
	s.db.Model(&t).Update("deleted_at", &now)
	return nil
}
