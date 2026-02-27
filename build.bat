@echo off
REM Build gomon for Windows
REM Run this after installing Go from https://go.dev/dl/

echo [1/2] Fetching dependencies...
go mod tidy
if errorlevel 1 (
    echo ERROR: go mod tidy failed. Is Go installed? https://go.dev/dl/
    pause
    exit /b 1
)

echo [2/2] Building...
go build -o gomon.exe .
if errorlevel 1 (
    echo ERROR: build failed.
    pause
    exit /b 1
)

echo.
echo Build successful! Run:  gomon.exe
pause
