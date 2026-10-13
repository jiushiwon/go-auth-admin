package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ===== 安全头（go-gin-init-skill 红线 6：强制开启） =====

// SecurityHeaders 注入安全响应头。
// 实现取自 go-gin-init-skill/references/middleware-guide.md「安全头中间件」，
// 逐字保持一致，不要自行增删头（增删会让前端行为与技能预期不符）。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		// XSS 防护
		c.Header("X-XSS-Protection", "1; mode=block")
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// 内容安全策略：技能中默认注释掉（会拦截内联脚本与 Vite 开发服务器），
		// 本项目前端为 Vue SPA，同样不启用，待前端改造后再评估。
		// c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// ===== CORS（对齐 WebConfig + SecurityConfig 的 corsConfigurationSource） =====

func CORSMiddleware(allowed string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowed == "" || allowed == "*" {
			// 通配不允许带凭证（规范禁止 * + Allow-Credentials）
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			ok := false
			for _, o := range strings.Split(allowed, ",") {
				if strings.TrimSpace(o) == origin {
					ok = true
					break
				}
			}
			if ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Vary", "Origin")
			}
			// 不在白名单则不加 Allow-Origin，浏览器会拒绝跨域
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Max-Age", "3600")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ===== 宿主 JWT 鉴权 + 权限装载（对齐 JwtAuthenticationFilter） =====

type AuthMiddleware struct {
	jwt  *JWTManager
	db   *gorm.DB
	head string
	pref string
}

func NewAuthMiddleware(jwt *JWTManager, db *gorm.DB, head, pref string) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, db: db, head: head, pref: pref}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开放接口由 AppSignatureMiddleware 处理；登录/刷新/健康检查放行。
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/open/") ||
			p == "/api/auth/login" || p == "/api/auth/refresh" || p == "/api/health" {
			c.Next()
			return
		}
		token := bearerToken(c.GetHeader(m.head), m.pref)
		if token == "" {
			Fail(c, Unauthorized("未登录"))
			c.Abort()
			return
		}
		claims, err := m.jwt.Parse(token)
		if err != nil {
			Fail(c, Unauthorized("令牌无效或已过期"))
			c.Abort()
			return
		}
		if claims.Type != "access" {
			Fail(c, Unauthorized("令牌类型错误，需要 accessToken"))
			c.Abort()
			return
		}
		uid := claims.UID

		var roles []string
		m.db.Model(&SysRole{}).
			Where("id IN (SELECT role_id FROM wg_sys_user_role WHERE user_id = ?)", uid).
			Pluck("code", &roles)

		isSuper := false
		for _, r := range roles {
			if r == RoleSuperAdmin {
				isSuper = true
				break
			}
		}
		perms := map[string]bool{}
		if isSuper {
			perms = permSetOf(AuthPermsAll)
		} else {
			var ps []string
			m.db.Model(&SysMenu{}).
				Where("id IN (SELECT menu_id FROM wg_sys_role_menu WHERE role_id IN (SELECT role_id FROM wg_sys_user_role WHERE user_id = ?)) AND permission IS NOT NULL AND deleted_at IS NULL", uid).
				Pluck("permission", &ps)
			perms = permSetOf(ps)
		}

		c.Set(ctxUID, uid)
		c.Set(ctxPerms, perms)
		c.Set(ctxIsSuper, isSuper)
		c.Next()
	}
}

// RequirePerm 方法级鉴权（对齐 @PreAuthorize("hasAuthority('...')")）。
func RequirePerm(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasPerm(c, perm) {
			c.Next()
			return
		}
		Fail(c, Forbidden("无权限："+perm))
		c.Abort()
	}
}

// ===== 应用级签名鉴权（对齐 AppSignatureAuthFilter，保护 /api/open/**） =====

// appSignatureSkewSeconds 为兜底默认值；实际值由 Config.AppSignatureSkewSeconds 注入。
const appSignatureSkewSeconds = 300

type AppPrincipal struct {
	AppID   int64  `json:"appId"`
	OwnerID int64  `json:"ownerId"`
	APIKey  string `json:"apiKey"`
}

type AppSigMiddleware struct {
	db *gorm.DB
	// skewSeconds 签名时间容忍窗口（秒），来自 Config，避免硬编码无法按环境收紧。
	skewSeconds int64
}

func NewAppSigMiddleware(db *gorm.DB, skewSeconds int64) *AppSigMiddleware {
	if skewSeconds <= 0 {
		skewSeconds = appSignatureSkewSeconds
	}
	return &AppSigMiddleware{db: db, skewSeconds: skewSeconds}
}

func (m *AppSigMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/open/") {
			c.Next()
			return
		}
		apiKey := c.GetHeader("X-App-Key")
		ts := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		sig := c.GetHeader("X-Signature")
		if apiKey == "" || ts == "" || nonce == "" || sig == "" {
			Fail(c, Unauthorized("缺少应用签名头（X-App-Key / X-Timestamp / X-Nonce / X-Signature）"))
			c.Abort()
			return
		}
		tsNum, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			Fail(c, Unauthorized("X-Timestamp 非法"))
			c.Abort()
			return
		}
		skew := time.Now().Unix() - tsNum
		if skew < 0 {
			skew = -skew
		}
		if skew > m.skewSeconds {
			Fail(c, Unauthorized("请求已过期，时间戳偏差 "+strconv.FormatInt(skew, 10)+" 秒"))
			c.Abort()
			return
		}

		var ak AppKey
		if err := m.db.Where("api_key = ? AND status = 1", apiKey).First(&ak).Error; err != nil {
			Fail(c, Unauthorized("应用密钥无效或已禁用"))
			c.Abort()
			return
		}
		var app App
		if err := m.db.Where("id = ?", ak.AppID).First(&app).Error; err != nil {
			Fail(c, Unauthorized("应用不存在"))
			c.Abort()
			return
		}
		if app.Status != 1 {
			Fail(c, Forbidden("应用已被禁用"))
			c.Abort()
			return
		}

		// 读取请求体后立即回填：body 需参与签名，同时下游 handler 还要读到原始 body。
		var body []byte
		if c.Request.Body != nil {
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				Fail(c, BadRequest("读取请求体失败"))
				c.Abort()
				return
			}
			body = b
			c.Request.Body = io.NopCloser(bytes.NewReader(b))
		}

		// P0-4：签名必须覆盖 query 与 body。
		// 仅签 Path 时，/api/open/userinfo?appUserId=X 的签名换成 appUserId=Y 仍验签通过，
		// POST 请求体也可被整体替换，等同于"签名形同虚设"。
		payload := strings.Join([]string{apiKey, ts, nonce, c.Request.Method,
			c.Request.URL.Path, c.Request.URL.RawQuery, sha256Hex(body)}, "\n")
		expected := hmacHex(ak.APISecret, payload)
		if !constantTimeHexEq(expected, sig) {
			Fail(c, Unauthorized("签名校验失败"))
			c.Abort()
			return
		}

		// P0-4 防重放：签名校验通过后登记 nonce，10 分钟内复用直接拒绝
		if e := m.consumeNonce(ak.ID, nonce); e != nil {
			Fail(c, e)
			c.Abort()
			return
		}

		// 更新 last_used_at（非关键，失败不影响主流程）
		now := time.Now()
		m.db.Model(&ak).Update("last_used_at", &now)

		c.Set(ctxAppPrinc, &AppPrincipal{AppID: app.ID, OwnerID: app.OwnerID, APIKey: apiKey})
		c.Next()
	}
}

func hmacHex(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func constantTimeHexEq(a, b string) bool {
	ab, err1 := hex.DecodeString(a)
	bb, err2 := hex.DecodeString(strings.ToLower(b))
	if err1 != nil || err2 != nil {
		return false
	}
	return subtle.ConstantTimeCompare(ab, bb) == 1
}

func (m *AppSigMiddleware) consumeNonce(appKeyID int64, nonce string) *AppError {
	// 清理窗口取「签名时间窗」：表不会无限增长，且窗口内的 nonce 不会被提前清除
	if err := m.db.Where("created_at < ?", time.Now().Add(-time.Duration(m.skewSeconds)*time.Second)).Delete(&AppApiNonce{}); err != nil {
		// 清理失败不阻断主流程
	}
	var cnt int64
	m.db.Model(&AppApiNonce{}).Where("nonce = ?", nonce).Count(&cnt)
	if cnt > 0 {
		return Unauthorized("签名已使用（疑似重放攻击）")
	}
	m.db.Create(&AppApiNonce{AppKeyID: appKeyID, Nonce: nonce, CreatedAt: time.Now()})
	return nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
