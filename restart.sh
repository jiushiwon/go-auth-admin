#!/usr/bin/env bash
# go-auth-admin 启停脚本
#
# 用法:
#   ./restart.sh          # 后台启动双端（默认）
#   ./restart.sh dev      # 前台启动（日志 tee 到屏幕，Ctrl+C 停止）
#   ./restart.sh be       # 只重启后端
#   ./restart.sh fe       # 只重启前端
#   ./restart.sh stop     # 停止双端
#   ./restart.sh status   # 查看状态
#   ./restart.sh -h       # 帮助
#
# 注意1：不做任何 git 操作（pull/commit/push 一律不碰）。
#
# 注意2（重要）：必须在**自己的终端**（Git Bash / PowerShell）里跑本脚本。
#   若在 IDE 的 AI 沙箱里执行，脚本一结束它启动的进程就会被回收掉 ——
#   表现是：日志里明明打印了 "VITE ready / 端口就绪"，但下一秒端口就没了。
#   这不是脚本坏了，是沙箱限制。日志在 logs/backend.log、logs/frontend.log。
#
# 注意3：5173 长期被另一个项目（Super Shell）占用，本项目固定用 5174。

# 第一行就要前置 coreutils，否则 IDE 的 bash shim 下 date/awk/tail/dirname 全找不到
export PATH="/c/Users/gang.wang/.workbuddy/binaries/PortableGit/versions/1.2.0/usr/bin:$PATH"
export CODEBUDDY_SAFE_DELETE_ENABLED=0   # 防 vite 依赖预构建被批量删除保护打死
set -u

ROOT="$(cd "$(dirname "$0")" && pwd)"
BE_PORT=8080
# 5173 长期被用户另一个项目（Super Shell）占用，本项目固定 5174
FE_PORT=5174
LOG_DIR="$ROOT/logs"
mkdir -p "$LOG_DIR"

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
die()  { log "!! $*"; exit 1; }

get_port_pids() {
  netstat -ano 2>/dev/null | grep ":$1[[:space:]]" | grep LISTENING \
    | awk '{print $NF}' | sort -u
}

kill_by_port() {
  local pids; pids=$(get_port_pids "$1")
  [ -z "$pids" ] && return 0
  for p in $pids; do
    MSYS_NO_PATHCONV=1 taskkill /F /T /PID "$p" >/dev/null 2>&1
  done
  log "已停止端口 $1 (pid: $(echo $pids | tr '\n' ' '))"
}

wait_port() {
  local i=0
  while [ $i -lt 40 ]; do
    [ -n "$(get_port_pids "$1")" ] && { log "端口 $1 就绪"; return 0; }
    sleep 1; i=$((i+1))
  done
  log "!! 端口 $1 等待超时，看日志: $LOG_DIR/"; return 1
}

check_env() {
  local f="$ROOT/frontend/.env.development"
  [ -f "$f" ] || return 0
  local v; v=$(grep -E '^VITE_API_BASE_URL=' "$f" | head -1 | cut -d= -f2-)
  if [ -n "$v" ]; then
    log "!! frontend/.env.development: VITE_API_BASE_URL='$v' 非空"
    log "   src/api/*.ts 的 URL 已自带 /api 前缀，此项必须留空"
    log "   否则请求变 /api/api/xxx -> 401，登录直接崩"
    return 1
  fi
  return 0
}

start_backend() {
  log "编译后端..."
  (cd "$ROOT/backend" && go build -o server.exe ./...) \
    || die "go build 失败"
  kill_by_port "$BE_PORT"; sleep 2
  (cd "$ROOT/backend" && nohup ./server.exe > "$LOG_DIR/backend.log" 2>&1 \
     & echo $! > "$LOG_DIR/backend.pid")
  wait_port "$BE_PORT" || { tail -20 "$LOG_DIR/backend.log"; return 1; }
  log "后端  http://localhost:$BE_PORT"
}

start_frontend() {
  check_env || die "环境配置有问题，已中止前端启动"
  kill_by_port "$FE_PORT"; sleep 2
  (cd "$ROOT/frontend" \
     && nohup node node_modules/vite/bin/vite.js --port "$FE_PORT" --host 0.0.0.0 \
        > "$LOG_DIR/frontend.log" 2>&1 & echo $! > "$LOG_DIR/frontend.pid")
  wait_port "$FE_PORT" || { tail -20 "$LOG_DIR/frontend.log"; return 1; }
  log "前端  http://localhost:$FE_PORT"
}

do_stop() {
  kill_by_port "$FE_PORT"
  kill_by_port "$BE_PORT"
  rm -f "$LOG_DIR"/*.pid
  log "双端已停止"
}

do_status() {
  local bp fp
  bp=$(get_port_pids "$BE_PORT"); fp=$(get_port_pids "$FE_PORT")
  [ -n "$bp" ] && log "后端 $BE_PORT  运行中 (pid $bp)" || log "后端 $BE_PORT  未运行"
  [ -n "$fp" ] && log "前端 $FE_PORT  运行中 (pid $fp)" || log "前端 $FE_PORT  未运行"
  # 5173 通常是别的项目，提示一下避免混淆
  local op; op=$(get_port_pids 5173)
  [ -n "$op" ] && log "提示: 5173 被占用 (pid $op)，多半是别的项目，本项目用 $FE_PORT"
}

case "${1:-start}" in
  start)
    start_backend && start_frontend && do_status
    ;;
  be) start_backend ;;
  fe) start_frontend ;;
  dev)
    start_backend
    (cd "$ROOT/frontend" && exec node node_modules/vite/bin/vite.js \
       --port "$FE_PORT" --host 0.0.0.0)
    ;;
  stop)   do_stop ;;
  status) do_status ;;
  -h|--help|help) sed -n '2,20p' "$0" ;;
  *) echo "未知参数: $1 (用 -h 看用法)" ; exit 1 ;;
esac
