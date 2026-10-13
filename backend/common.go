package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ===== 统一响应信封 =====
// 字段名与 java ApiResponse / PageResponse 逐字对齐（code/message/data、list/total/page/pageSize）。

type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ApiResponse{Code: 0, Message: "success", Data: data})
}

func Fail(c *gin.Context, e *AppError) {
	c.JSON(httpStatusForCode(e.Code), ApiResponse{Code: e.Code, Message: e.Message})
}

// ===== 业务异常 =====

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string { return e.Message }

func BadRequest(msg string) *AppError   { return &AppError{-1001, msg} }
func Unauthorized(msg string) *AppError { return &AppError{-1002, msg} }
func Forbidden(msg string) *AppError    { return &AppError{-1003, msg} }
func NotFound(msg string) *AppError     { return &AppError{-1004, msg} }
func Conflict(msg string) *AppError     { return &AppError{-1005, msg} }

func httpStatusForCode(code int) int {
	switch code {
	case -1001:
		return http.StatusBadRequest
	case -1002:
		return http.StatusUnauthorized
	case -1003:
		return http.StatusForbidden
	case -1004:
		return http.StatusNotFound
	case -1005:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// ===== JWT =====

type JWTManager struct {
	Secret        []byte
	AccessExpire  time.Duration
	RefreshExpire time.Duration
	Issuer        string
}

type jwtClaims struct {
	UID  int64  `json:"uid"`
	Type string `json:"type"`
	jwt.RegisteredClaims
}

func (m *JWTManager) sign(uid int64, username, ttype string, ttl time.Duration) string {
	now := time.Now()
	claims := jwtClaims{
		UID:  uid,
		Type: ttype,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			Issuer:    m.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString(m.Secret)
	return s
}

func (m *JWTManager) GenerateAccessToken(uid int64, username string) string {
	return m.sign(uid, username, "access", m.AccessExpire)
}

func (m *JWTManager) GenerateRefreshToken(uid int64, username string) string {
	return m.sign(uid, username, "refresh", m.RefreshExpire)
}

func (m *JWTManager) Parse(token string) (*jwtClaims, error) {
	claims := &jwtClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return m.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ===== 上下文键 =====

const (
	ctxUID      = "uid"
	ctxPerms    = "perms"
	ctxIsSuper  = "isSuper"
	ctxAppPrinc = "appPrincipal"
)

// ===== 权限常量（唯一常量源，三段式，与 AuthPerms.java 逐字一致） =====

const RoleSuperAdmin = "super_admin"

var AuthPermsAll = []string{
	"system:user:list", "system:user:create", "system:user:edit", "system:user:delete",
	"system:user:reset-pwd", "system:user:assign-role", "system:user:assign-post",
	"system:role:list", "system:role:create", "system:role:edit", "system:role:delete", "system:role:assign-menu",
	"system:menu:list", "system:menu:create", "system:menu:edit", "system:menu:delete",
	"system:org:list", "system:org:create", "system:org:edit", "system:org:delete",
	"system:post:list", "system:post:create", "system:post:edit", "system:post:delete",
	"system:tenant:list", "system:tenant:create", "system:tenant:edit", "system:tenant:delete",
	"system:app:list", "system:app:create", "system:app:edit", "system:app:delete", "system:app:key-manage",
	"account:bind:list", "account:bind:create", "account:bind:delete", "account:bind:set-default",
	"dashboard:home:view",
	"system:api:view",
}

func permSetOf(all []string) map[string]bool {
	m := make(map[string]bool, len(all))
	for _, p := range all {
		m[p] = true
	}
	return m
}

// ===== 时间格式化 =====
// 所有出参时间字段统一 yyyy-MM-dd HH:mm:ss（对齐 DateTimeFormat）。

func fmtTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func fmtTimeV(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// hasPerm 从 gin 上下文判断当前主体是否持有权限（super_admin 短路为全部）。
func hasPerm(c *gin.Context, perm string) bool {
	if v, ok := c.Get(ctxIsSuper); ok {
		if b, ok := v.(bool); ok && b {
			return true
		}
	}
	if v, ok := c.Get(ctxPerms); ok {
		if set, ok := v.(map[string]bool); ok {
			return set[perm]
		}
	}
	return false
}

// bearerToken 从 Authorization 头剥离前缀提取 token。
func bearerToken(header, prefix string) string {
	if prefix == "" {
		prefix = "Bearer "
	}
	if strings.HasPrefix(header, prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

// atoiDefault 解析查询参数整数，失败回退默认值。
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// pathID 解析路由路径参数 ID。
func pathID(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// strPtr 把字符串转 *string；空串返回 nil，使 GORM 落库为 NULL（与 Java 目录权限存 null 一致）。
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
