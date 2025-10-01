@echo off
setlocal enabledelayedexpansion

if "%~1"=="" (
    echo Usage: build.bat windows^|linux amd64^|arm64^|arm
    exit /b 1
)

if "%~2"=="" (
    echo Usage: build.bat windows^|linux amd64^|arm64^|arm
    exit /b 1
)

if /i "%~2"=="amd64" (
    set GOARCH=amd64
) else if /i "%~2"=="arm64" (
    set GOARCH=arm64
) else if /i "%~2"=="arm" (
    set GOARCH=arm
) else (
    echo Invalid architecture. Use amd64, arm64 or arm.
    exit /b 1
)

if /i "%~1"=="windows" (
    set GOOS=windows
) else if /i "%~1"=="linux" (
    set GOOS=linux
) else (
    echo Invalid OS. Use windows or linux.
    exit /b 1
)

set OUTPUT=build-%GOOS%-%GOARCH%
if /i "%GOOS%"=="windows" (
    set OUTPUT=%OUTPUT%.exe
)

echo Building %OUTPUT% for %GOOS%-%GOARCH%...

set GOOS=%GOOS%
set GOARCH=%GOARCH%

go build -ldflags="-s -w" -trimpath -o .\src\%OUTPUT%

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b %ERRORLEVEL%
)

echo Build complete: %OUTPUT%
