@echo off
REM 获取当前脚本所在目录
set SCRIPT_DIR=%~dp0

REM 设置 Protobuf 编译器的搜索路径为脚本所在目录
set PROTO_PATH=%SCRIPT_DIR%

REM 设置 Go 代码的输出目录
set GO_OUT_DIR=..\test

REM 创建输出目录
mkdir %GO_OUT_DIR% 2>nul

REM 查找所有 .proto 文件并编译
for /r %PROTO_PATH% %%f in (*.proto) do (
    protoc --proto_path=%PROTO_PATH% --go_out=%GO_OUT_DIR% --go_opt=paths=source_relative "%%f"
    if errorlevel 1 (
        echo Protobuf 文件 %%f 编译失败
        exit /b 1
    )
)

echo Protobuf 文件编译成功
exit /b 0