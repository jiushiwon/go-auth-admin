package main

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

// Config 聚合所有运行时配置（对齐 java-vue-admin application.yml 的环境变量）。
type Config struct {
	ServerPort int

	DBPath string

	// SQLite 并发相关（Go 侧专有配置，Java 无对应项；默认值保守，不改变 API 契约）
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBBusyTimeoutMs int

	JWTSecret            string
	JWTAccessExpireMin   int
	JWTRefreshExpireDays int
	JWTIssuer            string
	JWTHeader            string
	JWTPrefix            string

	CorsOrigins string

	// AppSignatureSkewSeconds：开放接口 /api/open/** 的签名时间容忍窗口（秒）。
	// 窗口越小越抗重放，但要求调用方时钟更准；过期 nonce 清理窗口与其联动。
	AppSignatureSkewSeconds int
}

// Load 读取 .env（若存在）并回退到与 application.yml 相同的默认值。
func Load() Config {
	_ = godotenv.Load() // 缺失 .env 不致命

	cfg := Config{
		ServerPort: getInt("SERVER_PORT", 8080),
		DBPath:     getStr("DB_URL", "./data/vue-admin.db"),
		// SQLite 是单写者模型：连接数必须限制，否则并发写直接 SQLITE_BUSY。
		// WAL 模式下可多读并发 + 单写，busy_timeout 让写锁等待而非立即报错。
		DBMaxOpenConns:       getInt("DB_MAX_OPEN_CONNS", 4),
		DBMaxIdleConns:       getInt("DB_MAX_IDLE_CONNS", 2),
		DBBusyTimeoutMs:      getInt("DB_BUSY_TIMEOUT_MS", 5000),
		JWTSecret:            getStr("JWT_SECRET", "Y2hhbmdlLW1lLXBsZWFzZS11c2Utb3BlbnNzbC1yYW5kLWJhc2U2NC0zMg=="),
		JWTAccessExpireMin:   getInt("JWT_ACCESS_EXPIRE_MINUTES", 60),
		JWTRefreshExpireDays: getInt("JWT_REFRESH_EXPIRE_DAYS", 7),
		JWTIssuer:            getStr("JWT_ISSUER", "vue-admin"),
		JWTHeader:            getStr("JWT_HEADER", "Authorization"),
		JWTPrefix:            getStr("JWT_PREFIX", "Bearer "),
		CorsOrigins:          getStr("CORS_ORIGINS", "*"),
		// 对齐既有 Java 版 appSignatureSkewSeconds=300；收紧请同时确认调用方时钟同步。
		AppSignatureSkewSeconds: getInt("APP_SIGNATURE_SKEW_SECONDS", 300),
	}
	return cfg
}

func getStr(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}

func getInt(k string, def int) int {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// AccessExpire / RefreshExpire 暴露为时长，供 JWTManager 使用。
func (c Config) AccessExpire() time.Duration {
	return time.Duration(c.JWTAccessExpireMin) * time.Minute
}

func (c Config) RefreshExpire() time.Duration {
	return time.Duration(c.JWTRefreshExpireDays) * 24 * time.Hour
}
