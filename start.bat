@echo off
REM go-auth-admin 一键启动 —— 双击即可
REM
REM 为什么需要它：在 IDE 的 AI 沙箱里启动的服务会被回收，十几分钟到一天不等，
REM 完全不可预测。这个 bat 由 Windows 资源管理器启动，进程不属于沙箱进程树，不会被回收。
REM
REM 停止服务：关掉标题为 gaa-backend / gaa-frontend 的两个最小化窗口。
REM 命令行用户也可以用 restart.sh（start/stop/status/be/fe/dev 子命令）。

cd /d "%~dp0"
title go-auth-admin

echo ==========================================
echo   go-auth-admin 一键启动
echo ==========================================
echo.

if not exist "%~dp0backend\server.exe" goto need_build
goto start_be

:need_build
echo [0/2] 未找到 server.exe，先编译后端 ...
pushd "%~dp0backend"
go build -o server.exe ./...
popd
if errorlevel 1 goto build_fail

:start_be
netstat -ano | findstr ":8080 " | findstr LISTENING >nul
if not errorlevel 1 goto be_ok
echo [1/2] 启动后端 端口 8080 ...
start "gaa-backend" /min /d "%~dp0backend" server.exe
goto be_wait

:be_ok
echo [1/2] 后端已在运行，跳过

:be_wait
echo [2/2] 启动前端 端口 5174 ...
start "gaa-frontend" /min /d "%~dp0frontend" node node_modules/vite/bin/vite.js --port 5174 --host 0.0.0.0

echo.
echo 等待服务就绪 ...
ping -n 8 127.0.0.1 >nul
echo.
echo   后端   http://localhost:8080/api/health
echo   前端   http://localhost:5174
echo.
echo   登录账号  admin
echo   初始密码  随机生成，见 gaa-backend 窗口日志（[seed] 初始账户密码...）
echo             不是 admin123；没记下就删 backend\data\vue-admin.db 重新播种
echo   注意     前端用 5174；5173 通常是别的项目
echo.
goto end

:build_fail
echo.
echo 后端编译失败，请确认已安装 Go 并在 PATH 中。

:end
echo 按任意键关闭本窗口（服务会在后台继续运行）...
pause >nul
