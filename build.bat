@echo off

if "%1" == "" (
    echo Usage: build.bat windows^|linux^|pi
    exit /b 1
)

if /i "%1" == "windows" (
    set GOOS = windows
    set GOARCH = amd64
    set OUTPUT = build/build-windows.exe
) else if /i "%1"=="linux" (
    set GOOS = linux
    set GOARCH = arm64
    set OUTPUT = build/build-linux
) else if /i "%1" == "pi" (
    set GOOS = linux
    set GOARCH = arm
    set OUTPUT = build/build-pi
) else (
    echo Invalid argument. Use "windows" or "linux".
    exit /b 1
)

echo Building for %GOOS%...
go build -o %OUTPUT% ./src

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b %ERRORLEVEL%
)

echo Build complete: %OUTPUT%
