@echo off
setlocal EnableDelayedExpansion

REM Configuration
set APP_NAME=ccproxy
set BUILD_DIR=bin
set VERSION=1.0.0
if not "%VERSION%"=="" set VERSION=%VERSION%

REM Get git commit hash
for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set COMMIT_HASH=%%i
if "%COMMIT_HASH%"=="" set COMMIT_HASH=unknown

REM Get build time
for /f "tokens=1-4 delims=/ " %%i in ('date /t') do set BUILD_DATE=%%i-%%j-%%k
for /f "tokens=1-2 delims=: " %%i in ('time /t') do set BUILD_TIME=%%i:%%j
set BUILD_TIMESTAMP=%BUILD_DATE%T%BUILD_TIME%Z

REM Build flags
set LDFLAGS=-X main.version=%VERSION% -X main.commit=%COMMIT_HASH% -X main.buildTime=%BUILD_TIMESTAMP% -s -w

echo Building %APP_NAME% v%VERSION%
echo Commit: %COMMIT_HASH%
echo Build Time: %BUILD_TIMESTAMP%
echo.

REM Create build directory
if not exist %BUILD_DIR% mkdir %BUILD_DIR%

REM Build proxy command
echo Building proxy command...
echo.

REM Build for Windows AMD64
echo Building proxy for windows/amd64...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-windows-amd64.exe .\cmd\proxy
if %errorlevel% neq 0 (
    echo Failed to build proxy for windows/amd64
    exit /b 1
)
echo Successfully built %APP_NAME%-windows-amd64.exe

REM Build for Linux AMD64
echo Building proxy for linux/amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-linux-amd64 .\cmd\proxy
if %errorlevel% neq 0 (
    echo Failed to build proxy for linux/amd64
    exit /b 1
)
echo Successfully built %APP_NAME%-linux-amd64

REM Build for Linux ARM64
echo Building proxy for linux/arm64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-linux-arm64 .\cmd\proxy
if %errorlevel% neq 0 (
    echo Failed to build proxy for linux/arm64
    exit /b 1
)
echo Successfully built %APP_NAME%-linux-arm64

REM Build for macOS AMD64
echo Building proxy for darwin/amd64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-darwin-amd64 .\cmd\proxy
if %errorlevel% neq 0 (
    echo Failed to build proxy for darwin/amd64
    exit /b 1
)
echo Successfully built %APP_NAME%-darwin-amd64

REM Build for macOS ARM64 (Apple Silicon)
echo Building proxy for darwin/arm64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-darwin-arm64 .\cmd\proxy
if %errorlevel% neq 0 (
    echo Failed to build proxy for darwin/arm64
    exit /b 1
)
echo Successfully built %APP_NAME%-darwin-arm64
echo.

REM Build setup command
echo Building setup command...
echo.

REM Build setup for Windows AMD64
echo Building setup for windows/amd64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-setup-windows-amd64.exe .\cmd\setup
if %errorlevel% neq 0 (
    echo Failed to build setup for windows/amd64
    exit /b 1
)
echo Successfully built %APP_NAME%-setup-windows-amd64.exe

REM Build setup for Linux AMD64
echo Building setup for linux/amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-setup-linux-amd64 .\cmd\setup
if %errorlevel% neq 0 (
    echo Failed to build setup for linux/amd64
    exit /b 1
)
echo Successfully built %APP_NAME%-setup-linux-amd64

REM Build setup for Linux ARM64
echo Building setup for linux/arm64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-setup-linux-arm64 .\cmd\setup
if %errorlevel% neq 0 (
    echo Failed to build setup for linux/arm64
    exit /b 1
)
echo Successfully built %APP_NAME%-setup-linux-arm64

REM Build setup for macOS AMD64
echo Building setup for darwin/amd64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-setup-darwin-amd64 .\cmd\setup
if %errorlevel% neq 0 (
    echo Failed to build setup for darwin/amd64
    exit /b 1
)
echo Successfully built %APP_NAME%-setup-darwin-amd64

REM Build setup for macOS ARM64 (Apple Silicon)
echo Building setup for darwin/arm64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-setup-darwin-arm64 .\cmd\setup
if %errorlevel% neq 0 (
    echo Failed to build setup for darwin/arm64
    exit /b 1
)
echo Successfully built %APP_NAME%-setup-darwin-arm64
echo.

echo All builds completed successfully!
echo Binaries available in %BUILD_DIR%\ directory
echo.

REM List built binaries
echo Built binaries:
dir /b %BUILD_DIR%\%APP_NAME%-*

echo.
echo Build process completed!