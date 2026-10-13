package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func main() {
	cfg := Load()

	// 确保数据库目录存在
	if dir := filepath.Dir(cfg.DBPath); dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_busy_timeout="+itoa(int64(cfg.DBBusyTimeoutMs))), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}

	// SQLite 并发加固：必须在迁移前设置（迁移本身即写操作）。
	// journal_mode=WAL → 多读并发 + 单写；busy_timeout → 写锁等待而非立即 SQLITE_BUSY。
	db.Exec("PRAGMA journal_mode = WAL;")
	db.Exec("PRAGMA busy_timeout = " + itoa(int64(cfg.DBBusyTimeoutMs)) + ";")
	// 连接池：SQLite 单写者模型，必须限流，否则并发写报 "database is locked"。
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	} else {
		log.Fatalf("获取底层 *sql.DB 失败: %v", err)
	}

	// 自动迁移（对齐 Hibernate ddl-auto=update）
	if err := db.AutoMigrate(
		&SysUser{}, &SysRole{}, &SysMenu{}, &SysOrg{}, &SysPost{}, &SysTenant{},
		&SysUserRole{}, &SysUserPost{}, &SysRoleMenu{},
		&App{}, &AppKey{}, &AppBinding{}, &BindCode{}, &AppApiNonce{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 种子数据
	Seed(db)

	// 依赖装配
	jwtMgr := &JWTManager{
		Secret:        []byte(cfg.JWTSecret),
		AccessExpire:  cfg.AccessExpire(),
		RefreshExpire: cfg.RefreshExpire(),
		Issuer:        cfg.JWTIssuer,
	}
	repo := NewRepo(db)
	svc := NewServices(db, repo, jwtMgr)
	auth := NewAuthMiddleware(jwtMgr, db, cfg.JWTHeader, cfg.JWTPrefix)
	appSig := NewAppSigMiddleware(db, int64(cfg.AppSignatureSkewSeconds))

	r := gin.New()
	// 中间件链顺序对齐 go-gin-init-skill：安全头 →（访问日志，见 docs/2026-10-07-audit.md P2-3，待补）→ CORS → 应用签名 → JWT 鉴权
	r.Use(gin.Recovery())
	r.Use(SecurityHeaders())
	r.Use(CORSMiddleware(cfg.CorsOrigins))
	r.Use(appSig.Handle())
	r.Use(auth.Handle())

	RegisterRoutes(r, svc)

	addr := ":" + itoa(int64(cfg.ServerPort))
	log.Printf("go-auth-admin 启动，监听 %s（DB=%s）", addr, cfg.DBPath)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
