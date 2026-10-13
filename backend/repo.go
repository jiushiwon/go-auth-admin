package main

import (
	"gorm.io/gorm"
)

// Repo 封装权限/数据权限相关的跨表查询（对齐 AuthorityLoader / PermissionEvaluator / DataScopeFilter）。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) FindRoleCodesByUser(uid int64) []string {
	var codes []string
	r.db.Model(&SysRole{}).
		Where("id IN (SELECT role_id FROM wg_sys_user_role WHERE user_id = ?)", uid).
		Pluck("code", &codes)
	return codes
}

func (r *Repo) FindPermissionsByUser(uid int64) []string {
	var ps []string
	r.db.Model(&SysMenu{}).
		Where("id IN (SELECT menu_id FROM wg_sys_role_menu WHERE role_id IN (SELECT role_id FROM wg_sys_user_role WHERE user_id = ?)) AND permission IS NOT NULL AND deleted_at IS NULL", uid).
		Pluck("permission", &ps)
	return ps
}

func (r *Repo) FindDataScopesByUser(uid int64) []string {
	var scopes []string
	r.db.Model(&SysRole{}).
		Where("id IN (SELECT role_id FROM wg_sys_user_role WHERE user_id = ?)", uid).
		Pluck("data_scope", &scopes)
	return scopes
}

func (r *Repo) FindMenuIDsByUser(uid int64) []int64 {
	var ids []int64
	r.db.Model(&SysRoleMenu{}).
		Where("role_id IN (SELECT role_id FROM wg_sys_user_role WHERE user_id = ?)", uid).
		Pluck("menu_id", &ids)
	return ids
}

func (r *Repo) FindRoleMenuCount(roleID int64) int64 {
	var n int64
	r.db.Model(&SysRoleMenu{}).Where("role_id = ?", roleID).Count(&n)
	return n
}

// ResolveDataScope 多角色取最宽档（对齐 PermissionEvaluator.resolveDataScope）。
// 超级管理员恒为 ALL；其余按 width 取最宽。
func (r *Repo) ResolveDataScope(uid int64) string {
	scopes := r.FindDataScopesByUser(uid)
	widest := "SELF_ONLY"
	order := map[string]int{"ALL": 0, "DEPT_AND_BELOW": 1, "DEPT_ONLY": 2, "SELF_ONLY": 3}
	best := 3
	for _, s := range scopes {
		if w, ok := order[s]; ok && w < best {
			best = w
			widest = s
		}
	}
	return widest
}

func (r *Repo) UserOrgID(uid int64) *int64 {
	var u SysUser
	if err := r.db.Select("org_id").First(&u, uid).Error; err != nil {
		return nil
	}
	return u.OrgID
}

// DescendantOrgIDs 自身 + 全部子孙组织（内存递归，对齐 DataScopeFilter.selfAndDescendantIds）。
func (r *Repo) DescendantOrgIDs(root int64) []int64 {
	var orgs []SysOrg
	r.db.Find(&orgs)
	byParent := map[int64][]int64{}
	for _, o := range orgs {
		if o.ParentID != nil {
			byParent[*o.ParentID] = append(byParent[*o.ParentID], o.ID)
		}
	}
	result := []int64{root}
	grew := true
	for grew {
		grew = false
		for _, id := range append([]int64{}, result...) {
			for _, child := range byParent[id] {
				found := false
				for _, x := range result {
					if x == child {
						found = true
						break
					}
				}
				if !found {
					result = append(result, child)
					grew = true
				}
			}
		}
	}
	return result
}
