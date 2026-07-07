@echo off
:: MCPHub 构建脚本
:: 解决 wails build 不嵌入图标的问题

echo [1/3] 生成版本信息和图标资源...
goversioninfo 2>&1
if errorlevel 1 (
    echo ERROR: goversioninfo failed
    exit /b 1
)

echo [2/3] 编译 Go 应用（含嵌入图标）...
go build -buildvcs=false -tags desktop,wv2runtime.download,production -ldflags "-w -s -H windowsgui" -o build\bin\mcphub.exe 2>&1
if errorlevel 1 (
    echo ERROR: go build failed
    exit /b 1
)

echo [3/3] 清理临时文件...
del /q resource.syso 2>nul
del /q versioninfo.json 2>nul

echo.
echo 构建完成: build\bin\mcphub.exe
dir build\bin\mcphub.exe
