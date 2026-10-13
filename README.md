# go-auth-admin

> **java-vue-admin 的 Go 复刻**：同一套前端、同一套接口契约，后端从 Spring Boot 换成
> **Go + Gin + GORM + SQLite**，目标是**前端零改动直接桥接**。
>
> 已机器核对：端点 56/56、端点权限码 56/56、VO 字段、种子数据、权限常量 —— 与 Java 逐字一致。

---

## 快速开始

```bash
# 方式一：Windows 双击（推荐，进程不会被 IDE 沙箱回收）
start.bat

# 方式二：命令行（必须在自己的终端里跑，不能用 IDE 沙箱）
./restart.sh            # 后台启动双端
./restart.sh dev        # 前台，日志 tee 到屏幕
./restart.sh stop       # 停止
./restart.sh status     # 查看状态
```

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:5174 |
| 后端健康检查 | http://localhost:8080/api/health |

> ⚠️ 前端固定 **5174**（5173 长期被同机另一个项目占用）。
> ⚠️ 在 IDE 的 AI 沙箱里执行 `restart.sh`，脚本一结束进程就被回收 —— 用 `start.bat`。

## 默认账号

| 用户名 | 角色 |
|--------|------|
| `admin` | `super_admin`（全部权限） |
| `user_admin` | `user_admin` |
| `demo` | `common_user`（仅本人数据） |

> ⚠️ **初始密码是随机的，不是 `admin123`**。
> 首次启动（库中无用户时）种子会随机生成初始密码并**打印到终端 / 日志**：
> `[seed] 初始账户密码（admin / user_admin / demo 均为）：xxxxxxxxxxxx`
> 三个账号共用这一个初始密码（P3-18 弱口令修复，见 `docs/2026-10-07-audit.md`）。
> **该密码只出现这一次**，没记下来就删掉 `backend/data/vue-admin.db` 重新播种。

## 技术栈

| 端 | 技术 |
|----|------|
| 后端 | Go 1.26 + Gin + GORM + SQLite（纯 Go 驱动，CGO_ENABLED=0，零环境依赖）+ JWT |
| 前端 | Vue 3.4 + Vite 5 + TypeScript + Pinia + `vue-base-skill` 组件库 |

## 接口文档

启动后（需 `system:api:view` 权限）：

- `GET /api/api-registry` —— 全量接口清单，含字段级入参/出参
- `GET /api/api-registry/openapi` —— OpenAPI 3.1，供 AI 与外部工具直接消费

## 常见问题

| 现象 | 排查 |
|------|------|
| 端口 8080 被占 | `curl /api/health` 返回 HTML 而非 JSON，说明是别的项目占了 |
| 接口全 404 / 返回 HTML | vite proxy 未命中会 SPA fallback 返回 `index.html`（**HTTP 200**）。检查 `vite.config.ts` |
| `database is locked` | 确认 `DB_BUSY_TIMEOUT_MS > 0` 且 `data/` 下有 `-wal` 文件；别把 `DB_MAX_OPEN_CONNS` 调大 |
| 侧边栏菜单少 / 组织树只有根 | 树必须**递归构建**，值拷贝会静默丢子树。验收要断言节点数（菜单 37） |
| 沙箱里服务过会儿就没 | 用 `start.bat` 或自己的终端跑 `restart.sh` |

## 技能依赖

本项目由 wg-skills 驱动，迭代前**必读** [`SKILLS.md`](SKILLS.md) ——
它声明了依赖哪些技能、各自约束什么、改哪个文件前该加载哪个技能。

## 文档索引

| 文档 | 内容 |
|------|------|
| [`SKILLS.md`](SKILLS.md) | **技能依赖清单（迭代第一入口）** |
| [`api-contract.md`](api-contract.md) | 接口契约（前后端唯一事实来源） |
| `docs/project-guide.md` | 项目指南（启动/配置/排障/拓展） |
| `docs/perms.md` | 39 个权限码与菜单/按钮映射表 |
| `docs/2026-10-07-audit.md` | 安全与缺陷审计（P0–P3） |
| `docs/2026-10-07-api-registry-design.md` | 接口元数据层设计 |
| `docs/2026-10-07-next-iterations.md` | 后续迭代候选清单与排期 |
| `2026-10-07-技能合规审计.md` | 技能合规审计 |
