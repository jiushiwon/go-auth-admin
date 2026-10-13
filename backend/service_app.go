package main

import (
	"time"
)

// ===== AppService =====

func (s *Services) AppPage(ownerID int64, page, pageSize int) (*PageResponse, *AppError) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	q := s.db.Model(&App{}).Where("owner_id = ?", ownerID)
	var total int64
	q.Count(&total)
	var apps []App
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&apps)
	list := make([]AppVO, 0, len(apps))
	for _, a := range apps {
		list = append(list, appToVO(&a))
	}
	return &PageResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func appToVO(a *App) AppVO {
	return AppVO{ID: a.ID, AppName: a.AppName, AppKey: a.AppKey, Description: a.Description,
		Logo: a.Logo, CallbackURL: a.CallbackURL, Status: a.Status, CreatedAt: fmtTimeV(a.CreatedAt)}
}

func (s *Services) AppGet(id, ownerID int64) (*AppVO, *AppError) {
	a, e := s.requireOwnedApp(id, ownerID)
	if e != nil {
		return nil, e
	}
	vo := appToVO(a)
	return &vo, nil
}

func (s *Services) AppCreate(req CreateAppRequest, ownerID int64) (*AppVO, *AppError) {
	a := App{AppName: req.AppName, AppKey: s.genUniqueAppKey(), Description: req.Description,
		Logo: req.Logo, CallbackURL: req.CallbackURL, Status: 1, OwnerID: ownerID}
	if err := s.db.Create(&a).Error; err != nil {
		return nil, BadRequest("创建应用失败")
	}
	vo := appToVO(&a)
	return &vo, nil
}

func (s *Services) AppUpdate(id int64, req UpdateAppRequest, ownerID int64) (*AppVO, *AppError) {
	a, e := s.requireOwnedApp(id, ownerID)
	if e != nil {
		return nil, e
	}
	if req.AppName != "" {
		a.AppName = req.AppName
	}
	if req.Description != "" {
		a.Description = req.Description
	}
	if req.Logo != "" {
		a.Logo = req.Logo
	}
	if req.CallbackURL != "" {
		a.CallbackURL = req.CallbackURL
	}
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			return nil, BadRequest("status 只允许 0 或 1")
		}
		a.Status = *req.Status
	}
	s.db.Save(&a)
	vo := appToVO(a)
	return &vo, nil
}

func (s *Services) AppDelete(id, ownerID int64) *AppError {
	a, e := s.requireOwnedApp(id, ownerID)
	if e != nil {
		return e
	}
	now := time.Now()
	s.db.Model(&a).Update("deleted_at", &now)
	s.db.Model(&AppKey{}).Where("app_id = ?", id).Update("deleted_at", &now)
	s.db.Model(&AppBinding{}).Where("app_id = ?", id).Update("deleted_at", &now)
	return nil
}

func (s *Services) AppListKeys(appID, ownerID int64) (*PageResponse, *AppError) {
	if _, e := s.requireOwnedApp(appID, ownerID); e != nil {
		return nil, e
	}
	var keys []AppKey
	s.db.Where("app_id = ?", appID).Order("id DESC").Find(&keys)
	list := make([]AppKeyVO, 0, len(keys))
	for _, k := range keys {
		list = append(list, AppKeyVO{ID: k.ID, KeyName: k.KeyName, APIKey: k.APIKey,
			Status: k.Status, LastUsedAt: fmtTime(k.LastUsedAt), CreatedAt: fmtTimeV(k.CreatedAt)})
	}
	return &PageResponse{List: list, Total: int64(len(keys))}, nil
}

func (s *Services) AppCreateKey(appID int64, req CreateAppKeyRequest, ownerID int64) (*CreatedAppKeyVO, *AppError) {
	if _, e := s.requireOwnedApp(appID, ownerID); e != nil {
		return nil, e
	}
	name := req.KeyName
	if name == "" {
		name = "默认密钥"
	}
	k := AppKey{AppID: appID, KeyName: name, APIKey: s.genUniqueAPIKey(), APISecret: "sk_" + randomHex(32), Status: 1}
	if err := s.db.Create(&k).Error; err != nil {
		return nil, BadRequest("创建密钥失败")
	}
	return &CreatedAppKeyVO{ID: k.ID, KeyName: k.KeyName, APIKey: k.APIKey, APISecret: k.APISecret,
		Status: k.Status, CreatedAt: fmtTimeV(k.CreatedAt)}, nil
}

func (s *Services) AppDeleteKey(appID, keyID, ownerID int64) *AppError {
	if _, e := s.requireOwnedApp(appID, ownerID); e != nil {
		return e
	}
	var k AppKey
	if err := s.db.Where("id = ? AND deleted_at IS NULL", keyID).First(&k).Error; err != nil {
		return NotFound("密钥不存在")
	}
	if k.AppID != appID {
		return Forbidden("密钥不属于该应用")
	}
	now := time.Now()
	s.db.Model(&k).Update("deleted_at", &now)
	return nil
}

func (s *Services) requireOwnedApp(id, ownerID int64) (*App, *AppError) {
	var a App
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&a).Error; err != nil {
		return nil, NotFound("应用不存在")
	}
	if a.OwnerID != ownerID {
		return nil, Forbidden("无权操作该应用")
	}
	return &a, nil
}

func (s *Services) genUniqueAppKey() string {
	for {
		k := "app_" + randomHex(12)
		var cnt int64
		s.db.Model(&App{}).Where("app_key = ?", k).Count(&cnt)
		if cnt == 0 {
			return k
		}
	}
}

func (s *Services) genUniqueAPIKey() string {
	for {
		k := "ak_" + randomHex(12)
		var cnt int64
		s.db.Model(&AppKey{}).Where("api_key = ?", k).Count(&cnt)
		if cnt == 0 {
			return k
		}
	}
}

// ===== BindService =====

func (s *Services) BindListMine(sysUserID int64) []BindVO {
	var bindings []AppBinding
	s.db.Where("sys_user_id = ? AND deleted_at IS NULL", sysUserID).Order("id DESC").Find(&bindings)
	if len(bindings) == 0 {
		return []BindVO{}
	}
	appIDs := map[int64]bool{}
	for _, b := range bindings {
		appIDs[b.AppID] = true
	}
	ids := make([]int64, 0, len(appIDs))
	for id := range appIDs {
		ids = append(ids, id)
	}
	var apps []App
	s.db.Where("id IN ?", ids).Find(&apps)
	appMap := map[int64]App{}
	for _, a := range apps {
		appMap[a.ID] = a
	}
	list := make([]BindVO, 0, len(bindings))
	for _, b := range bindings {
		list = append(list, bindToVO(&b, appMap[b.AppID]))
	}
	return list
}

func (s *Services) BindCreateUserInitiatedCode(sysUserID, appID int64) (*BindCodeVO, *AppError) {
	if e := s.requireEnabledApp(appID); e != nil {
		return nil, e
	}
	var cnt int64
	s.db.Model(&AppBinding{}).Where("sys_user_id = ? AND app_id = ? AND deleted_at IS NULL", sysUserID, appID).Count(&cnt)
	if cnt > 0 {
		return nil, Conflict("该应用已绑定到当前账户")
	}
	bc := s.issueBindCode(BindDirectionUserInitiated, appID, &sysUserID, "", "")
	return s.bindCodeToVO(bc), nil
}

func (s *Services) BindConfirmAppInitiated(sysUserID int64, code string) (BindVO, *AppError) {
	bc, e := s.requireUsableBindCode(code)
	if e != nil {
		return BindVO{}, e
	}
	if bc.Direction != BindDirectionAppInitiated {
		return BindVO{}, BadRequest("绑定码方向不匹配：该码需由应用侧认领")
	}
	now := time.Now()
	res := s.db.Model(&BindCode{}).Where("code = ? AND status = ?", code, BindStatusPending).
		Updates(map[string]interface{}{"status": BindStatusUsed, "used_at": &now})
	if res.RowsAffected != 1 {
		return BindVO{}, Conflict("绑定码已被使用或已失效")
	}
	return s.doBind(sysUserID, bc.AppID, bc.AppUserID, bc.AppUserName, bc.Direction)
}

func (s *Services) BindCancel(sysUserID, bindingID int64) *AppError {
	var b AppBinding
	if err := s.db.Where("id = ? AND sys_user_id = ? AND deleted_at IS NULL", bindingID, sysUserID).First(&b).Error; err != nil {
		return NotFound("绑定不存在")
	}
	now := time.Now()
	s.db.Model(&b).Update("deleted_at", &now)
	return nil
}

func (s *Services) BindSetDefault(sysUserID, bindingID int64) *AppError {
	var b AppBinding
	if err := s.db.Where("id = ? AND sys_user_id = ? AND deleted_at IS NULL", bindingID, sysUserID).First(&b).Error; err != nil {
		return NotFound("绑定不存在")
	}
	s.db.Model(&AppBinding{}).Where("sys_user_id = ?", sysUserID).Update("is_default", 0)
	b.IsDefault = 1
	s.db.Save(&b)
	return nil
}

func (s *Services) BindIssueForApp(appID int64, appUserID, appUserName string) (*BindCodeVO, *AppError) {
	if e := s.requireEnabledApp(appID); e != nil {
		return nil, e
	}
	bc := s.issueBindCode(BindDirectionAppInitiated, appID, nil, appUserID, appUserName)
	return s.bindCodeToVO(bc), nil
}

func (s *Services) BindClaimForApp(appID int64, code, appUserID, appUserName string) (BindVO, *AppError) {
	bc, e := s.requireUsableBindCode(code)
	if e != nil {
		return BindVO{}, e
	}
	if bc.Direction != BindDirectionUserInitiated {
		return BindVO{}, BadRequest("绑定码方向不匹配：该码需由宿主侧确认")
	}
	if bc.AppID != appID {
		return BindVO{}, Forbidden("绑定码不属于当前应用")
	}
	if bc.SysUserID == nil {
		return BindVO{}, Conflict("绑定码缺少宿主账户，无法认领")
	}
	now := time.Now()
	res := s.db.Model(&BindCode{}).Where("code = ? AND status = ?", code, BindStatusPending).
		Updates(map[string]interface{}{"status": BindStatusUsed, "used_at": &now})
	if res.RowsAffected != 1 {
		return BindVO{}, Conflict("绑定码已被使用或已失效")
	}
	return s.doBind(*bc.SysUserID, appID, appUserID, appUserName, bc.Direction)
}

func (s *Services) issueBindCode(direction string, appID int64, sysUserID *int64, appUserID, appUserName string) *BindCode {
	now := time.Now()
	bc := BindCode{
		Code:        randomHex(16),
		Direction:   direction,
		AppID:       appID,
		SysUserID:   sysUserID,
		AppUserID:   appUserID,
		AppUserName: appUserName,
		Status:      BindStatusPending,
		ExpireAt:    now.Add(5 * time.Minute),
		CreatedAt:   now,
	}
	s.db.Create(&bc)
	return &bc
}

func (s *Services) requireUsableBindCode(code string) (*BindCode, *AppError) {
	var bc BindCode
	if err := s.db.Where("code = ?", code).First(&bc).Error; err != nil {
		return nil, NotFound("绑定码无效")
	}
	if !bc.Usable() {
		return nil, Conflict("绑定码已过期或已被使用")
	}
	return &bc, nil
}

func (s *Services) doBind(sysUserID, appID int64, appUserID, appUserName, bindType string) (BindVO, *AppError) {
	if e := s.requireEnabledApp(appID); e != nil {
		return BindVO{}, e
	}
	var cnt int64
	s.db.Model(&AppBinding{}).Where("sys_user_id = ? AND app_id = ? AND deleted_at IS NULL", sysUserID, appID).Count(&cnt)
	if cnt > 0 {
		return BindVO{}, Conflict("该应用已绑定到当前账户")
	}
	now := time.Now()
	var existing int64
	s.db.Model(&AppBinding{}).Where("sys_user_id = ? AND deleted_at IS NULL", sysUserID).Count(&existing)
	isDefault := 0
	if existing == 0 {
		isDefault = 1
	}
	b := AppBinding{SysUserID: sysUserID, AppID: appID, AppUserID: appUserID, AppUserName: appUserName,
		BindType: bindType, IsDefault: isDefault, BindAt: &now, CreatedAt: now}
	s.db.Create(&b)
	var app App
	s.db.First(&app, appID)
	return bindToVO(&b, app), nil
}

func (s *Services) requireEnabledApp(appID int64) *AppError {
	var a App
	if err := s.db.Where("id = ? AND deleted_at IS NULL", appID).First(&a).Error; err != nil {
		return NotFound("应用不存在")
	}
	if a.Status != 1 {
		return Forbidden("应用已被禁用")
	}
	return nil
}

func bindToVO(b *AppBinding, app App) BindVO {
	vo := BindVO{ID: b.ID, AppID: b.AppID, AppUserID: b.AppUserID, AppUserName: b.AppUserName,
		BindType: b.BindType, IsDefault: b.IsDefault == 1, BindAt: fmtTime(b.BindAt)}
	if app.ID != 0 {
		vo.AppKey = app.AppKey
		vo.AppName = app.AppName
		vo.Logo = app.Logo
	}
	return vo
}

func (s *Services) bindCodeToVO(bc *BindCode) *BindCodeVO {
	expire := int64(time.Until(bc.ExpireAt).Seconds())
	if expire < 0 {
		expire = 0
	}
	return &BindCodeVO{
		Code:          bc.Code,
		Direction:     bc.Direction,
		ExpireSeconds: expire,
		QrContent:     bc.Code, // 绑定码本身即可作为扫码内容
		ExpireAt:      fmtTimeV(bc.ExpireAt),
	}
}

// ===== OpenApiService（server-to-server，身份来自 AppSignatureMiddleware 写入的 AppPrincipal） =====

func (s *Services) OpenApplyBind(p *AppPrincipal, req OpenBindApplyRequest) (*BindCodeVO, *AppError) {
	return s.BindIssueForApp(p.AppID, req.AppUserID, req.AppUserName)
}

func (s *Services) OpenClaimBind(p *AppPrincipal, req OpenBindClaimRequest) (BindVO, *AppError) {
	return s.BindClaimForApp(p.AppID, req.Code, req.AppUserID, req.AppUserName)
}

func (s *Services) OpenUserInfo(p *AppPrincipal, appUserID string) (*OpenUserInfoVO, *AppError) {
	vo := &OpenUserInfoVO{AppUserID: appUserID, Bound: false}
	var b AppBinding
	if err := s.db.Where("app_id = ? AND app_user_id = ? AND deleted_at IS NULL", p.AppID, appUserID).First(&b).Error; err != nil {
		return vo, nil
	}
	var u SysUser
	if err := s.db.First(&u, b.SysUserID).Error; err != nil || u.DeletedAt.Valid {
		return vo, nil
	}
	if u.Status != 1 {
		return nil, Forbidden("宿主账户已被禁用")
	}
	vo.Bound = true
	vo.BindingID = b.ID
	vo.UserID = u.ID
	vo.Username = u.Username
	vo.Nickname = u.Nickname
	vo.Avatar = u.Avatar
	vo.Email = u.Email
	vo.Phone = u.Phone
	return vo, nil
}
