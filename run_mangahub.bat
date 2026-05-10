@echo off
setlocal enabledelayedexpansion

:: ======================================================
:: MANGAHUB SYSTEM STARTER (Windows)
:: ======================================================

echo.
echo  [MangaHub] Starting system cleanup and initialization...
echo.

:: 1. Stop Docker containers
echo [1/4] Stopping existing Docker containers...
docker-compose down --remove-orphans >nul 2>&1
echo Done.

:: 2. Kill local processes on required ports to avoid "bind: port already in use"
echo [2/4] Checking and clearing ports: 8080, 9090, 50051, 9999...

set PORTS=8080 9090 50051 9999 8888

for %%p in (%PORTS%) do (
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr :%%p ^| findstr LISTENING') do (
        echo  - Found process %%a using port %%p. Killing it...
        taskkill /F /PID %%a >nul 2>&1
    )
)
echo Done.

:: 3. Start Docker Compose
echo [3/4] Building and starting Docker containers (Background mode)...
docker-compose up --build -d
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] Docker Compose failed to start. Please check if Docker Desktop is running.
    pause
    exit /b %ERRORLEVEL%
)
echo Done.

:: 4. Finalizing
echo [4/4] Waiting for services to be ready (5s)...
timeout /t 5 /nobreak >nul

echo.
echo ======================================================
echo   MangaHub is now running!
echo   - Web:  http://localhost:8080
echo   - TCP:  localhost:9090 (CLI)
echo   - gRPC: localhost:50051
echo ======================================================
echo.

:: Launch browser
start http://localhost:8080

echo Press any key to stop the system and exit...
pause >nul

echo.
echo [FINAL] Stopping system...
docker-compose stop
echo Done. Bye!
timeout /t 3 >nul
