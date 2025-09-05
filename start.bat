
@echo off
REM TalkAI OpenAI API 适配器启动脚本 (Windows版本)
REM 使用方法: start.bat [选项]

setlocal enabledelayedexpansion

REM 版本信息
set "SCRIPT_VERSION=1.1.0"

REM 默认配置
set "DEFAULT_PORT=9091"
set "DEFAULT_STREAM=false"
set "DEFAULT_MODEL=claude-opus-4-1-20250805"
set "DEFAULT_TEMPERATURE=0.7"
set "DEFAULT_TIMEOUT=300"
set "DEFAULT_DEBUG_MODE=false"
set "LOG_FILE=talkai.log"
set "PID_FILE=talkai.pid"

REM 配置变量
set "PORT="
set "API_KEYS="
set "STREAM="
set "MODEL="
set "TEMPERATURE="
set "TIMEOUT="
set "DEBUG_MODE="
set "CONFIG_FILE="
set "HELP=false"
set "INIT_CONFIG=false"
set "DAEMON_MODE=false"
set "STATUS=false"
set "STOP=false"
set "VERSION=false"

REM 解析命令行参数
:parse_args
if "%~1"=="" goto end_parse_args
if "%~1"=="-p" goto set_port
if "%~1"=="--port" goto set_port
if "%~1"=="-k" goto set_api_keys
if "%~1"=="--api-keys" goto set_api_keys
if "%~1"=="-s" goto set_stream
if "%~1"=="--stream" goto set_stream
if "%~1"=="-m" goto set_model
if "%~1"=="--model" goto set_model
if "%~1"=="-t" goto set_temperature
if "%~1"=="--temperature" goto set_temperature
if "%~1"=="-T" goto set_timeout
if "%~1"=="--timeout" goto set_timeout
if "%~1"=="-d" goto set_debug
if "%~1"=="--debug" goto set_debug
if "%~1"=="-c" goto set_config
if "%~1"=="--config" goto set_config
if "%~1"=="-i" goto set_init
if "%~1"=="--init" goto set_init
if "%~1"=="-D" goto set_daemon
if "%~1"=="--daemon" goto set_daemon
if "%~1"=="--status" goto set_status
if "%~1"=="--stop" goto set_stop
if "%~1"=="-v" goto set_version
if "%~1"=="--version" goto set_version
if "%~1"=="-h" goto set_help
if "%~1"=="--help" goto set_help

echo [错误] 未知参数 %~1
call :print_help
exit /b 1

:set_port
set "PORT=%~2"
shift
shift
goto parse_args

:set_api_keys
set "API_KEYS=%~2"
shift
shift
goto parse_args

:set_stream
set "STREAM=%~2"
shift
shift
goto parse_args

:set_model
set "MODEL=%~2"
shift
shift
goto parse_args

:set_temperature
set "TEMPERATURE=%~2"
shift
shift
goto parse_args

:set_timeout
set "TIMEOUT=%~2"
shift
shift
goto parse_args

:set_debug
set "DEBUG_MODE=%~2"
shift
shift
goto parse_args

:set_config
set "CONFIG_FILE=%~2"
shift
shift
goto parse_args

:set_init
set "INIT_CONFIG=true"
shift
goto parse_args

:set_daemon
set "DAEMON_MODE=true"
shift
goto parse_args

:set_status
set "STATUS=true"
shift
goto parse_args

:set_stop
set "STOP=true"
shift
goto parse_args

:set_version
set "VERSION=true"
shift
goto parse_args

:set_help
set "HELP=true"
shift
goto parse_args

:end_parse_args

REM 打印帮助信息
:print_help
echo TalkAI OpenAI API 适配器启动脚本 v%SCRIPT_VERSION%
echo.
echo 使用方法: %~nx0 [选项]
echo.
echo 选项:
echo   -p, --port PORT          服务器端口 (默认: %DEFAULT_PORT%)
echo   -k, --api-keys KEYS     API 密钥列表，多个密钥用逗号分隔
echo   -s, --stream STREAM      默认流模式 (true/false, 默认: %DEFAULT_STREAM%)
echo   -m, --model MODEL        默认模型 (默认: %DEFAULT_MODEL%)
echo   -t, --temperature TEMP   默认温度 (默认: %DEFAULT_TEMPERATURE%)
echo   -T, --timeout TIMEOUT    请求超时时间，单位秒 (默认: %DEFAULT_TIMEOUT%)
echo   -d, --debug MODE         调试模式 (true/false, 默认: %DEFAULT_DEBUG_MODE%)
echo   -c, --config FILE        配置文件路径
echo   -i, --init               初始化配置文件
echo   -D, --daemon             后台运行模式
echo   --status                 查看服务状态
echo   --stop                   停止服务
echo   -v, --version            显示版本信息
echo   -h, --help               显示帮助信息
echo.
echo 示例:
echo   %~nx0 --port 8002 --api-keys sk-key1,sk-key2 --stream true
echo   %~nx0 --config config.env
echo   %~nx0 --init                 # 初始化配置
echo   %~nx0 --daemon               # 后台运行
echo   %~nx0 --status               # 查看状态
echo   %~nx0 --stop                 # 停止服务
echo.
echo 配置文件:
echo   如果存在 env.local 文件，会自动加载其中的配置
echo   也可以使用 -c 参数指定其他配置文件
echo.
echo 环境变量:
echo   也可以通过环境变量设置配置，环境变量优先级高于配置文件
echo   PORT, API_KEYS, DEFAULT_STREAM, DEFAULT_MODEL, DEFAULT_TEMPERATURE, TIMEOUT, DEBUG_MODE
exit /b 0

REM 从配置文件读取配置
:load_config_file
if defined CONFIG_FILE (
    if exist "%CONFIG_FILE%" (
        echo [从配置文件加载配置: %CONFIG_FILE%]
        for /f "usebackq tokens=*" %%a in ("%CONFIG_FILE%") do (
            for /f "tokens=1,2 delims==" %%b in ("%%a") do (
                set "%%b=%%c"
            )
        )
    )
) else (
    if exist "env.local" (
        echo [从 env.local 文件加载配置]
        for /f "usebackq tokens=*" %%a in ("env.local") do (
            for /f "tokens=1,2 delims==" %%b in ("%%a") do (
                set "%%b=%%c"
            )
        )
    )
)
exit /b 0

REM 初始化配置文件
:init_config
echo [初始化配置文件...]

if exist "env.local" (
    set /p "response=env.local 文件已存在，是否覆盖？(y/N) "
    if /i not "!response!"=="y" (
        echo [保留现有配置文件]
        exit /b 0
    )
)

copy "env.local.example" "env.local" >nul
echo [配置文件已创建: env.local]
echo [请编辑 env.local 文件以配置您的设置]

set /p "edit_response=是否立即编辑配置文件？(Y/n) "
if /i "!edit_response!"=="" (
    set "edit_response=y"
)
if /i "!edit_response!"=="y" (
    notepad "env.local"
)
exit /b 0

REM 检查服务状态
:check_status
if exist "%PID_FILE%" (
    set /p "PID=<%PID_FILE%"
    tasklist /FI "PID eq !PID!" 2>nul | find "!PID!" >nul
    if !errorlevel! equ 0 (
        echo [服务正在运行 (PID: !PID!)]
        exit /b 0
    ) else (
        echo [PID 文件存在但服务未运行，清理中...]
        del "%PID_FILE%" >nul 2>&1
        exit /b 1
    )
) else (
    echo [服务未运行]
    exit /b 1
)

REM 停止服务
:stop_service
if exist "%PID_FILE%" (
    set /p "PID=<%PID_FILE%"
    tasklist /FI "PID eq !PID!" 2>nul | find "!PID!" >nul
    if !errorlevel! equ 0 (
        echo [正在停止服务 (PID: !PID!)...]
        taskkill /F /PID !PID! >nul 2>&1
        del "%PID_FILE%" >nul 2>&1
        echo [服务已停止]
    ) else (
        echo [服务未运行，清理 PID 文件...]
        del "%PID_FILE%" >nul 2>&1
    )
) else (
    echo [服务未运行]
)
exit /b 0

REM 记录日志
:log
set "level=%~1"
shift
set "message=%*"
for /f "tokens=1-3 delims=/ " %%a in ("%date%") do set "date=%%c-%%a-%%b"
for /f "tokens=1-3 delims=:." %%a in ("%time%") do set "time=%%a:%%b:%%c"
echo [%date% %time%] [%level%] %message% >> "%LOG_FILE%"
exit /b 0

REM 设置环境变量
:set_env_vars
echo [设置环境变量...]

REM 命令行参数优先级最高
if defined PORT set "PORT=%PORT%"
if defined API_KEYS set "API_KEYS=%API_KEYS%"
if defined STREAM set "DEFAULT_STREAM=%STREAM%"
if defined MODEL set "DEFAULT_MODEL=%MODEL%"
if defined TEMPERATURE set "DEFAULT_TEMPERATURE=%TEMPERATURE%"
if defined TIMEOUT set "TIMEOUT=%TIMEOUT%"
if defined DEBUG_MODE set "DEBUG_MODE=%DEBUG_MODE%"

REM 如果没有设置，则使用默认值
if not defined PORT set "PORT=%DEFAULT_PORT%"
if not defined DEFAULT_STREAM set "DEFAULT_STREAM=%DEFAULT_STREAM%"
if not defined DEFAULT_MODEL set "DEFAULT_MODEL=%DEFAULT_MODEL%"
if not defined DEFAULT_TEMPERATURE set "DEFAULT_TEMPERATURE=%DEFAULT_TEMPERATURE%"
if not defined TIMEOUT set "TIMEOUT=%DEFAULT_TIMEOUT%"
if not defined DEBUG_MODE set "DEBUG_MODE=%DEFAULT_DEBUG_MODE%"

REM 记录配置到日志
call :log INFO "配置: PORT=%PORT%, MODEL=%DEFAULT_MODEL%, STREAM=%DEFAULT_STREAM%, DEBUG=%DEBUG_MODE%"

echo [环境变量设置完成]
exit /b 0

REM 打印配置信息
:print_config
echo [当前配置:]
echo   端口: %PORT%
if defined API_KEYS (
    REM 计算API密钥数量
    set "count=0"
    for %%a in (%API_KEYS:,= %) do (
        set /a "count+=1"
    )
    echo   API 密钥: 已配置 (!count! 个)
) else (
    echo   API 密钥: 从文件读取
)
echo   默认流模式: %DEFAULT_STREAM%
echo   默认模型: %DEFAULT_MODEL%
echo   默认温度: %DEFAULT_TEMPERATURE%
echo   超时时间: %TIMEOUT% 秒
echo   调试模式: %DEBUG_MODE%
echo.
exit /b 0

REM 检查依赖
:check_dependencies
REM 检查 Go
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误: 未找到 Go 命令]
    echo [正在尝试安装 Go...]
    echo [请手动安装 Go: https://golang.org/dl/]
    exit /b 1
)

REM 检查 main.go
if not exist "main.go" (
    echo [错误: 未找到 main.go 文件]
    exit /b 1
)

REM 检查 env.local.example
if not exist "env.local.example" (
    echo [错误: 未找到 env.local.example 文件]
    exit /b 1
)
exit /b 0

REM 检查并处理端口占用
:check_and_kill_port
set "port=%~1"

REM 检查端口是否被占用
netstat -ano | findstr ":%port%" | findstr "LISTENING" >nul
if %errorlevel% equ 0 (
    echo [端口 %port% 已被占用，尝试终止占用进程...]
    
    REM 获取占用端口的进程ID
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%port%" ^| findstr "LISTENING"') do (
        set "pid=%%a"
    )
    
    if defined pid (
        REM 终止进程
        taskkill /F /PID !pid! >nul 2>&1
        echo [已终止占用端口 %port% 的进程 (PID: !pid!)]
        
        REM 等待进程完全退出
                timeout /t 2 /nobreak >nul
                
                REM 再次检查端口是否仍被占用
                netstat -ano | findstr ":%port%" | findstr "LISTENING" >nul
                if %errorlevel% equ 0 (
                    echo [无法释放端口 %port%，请手动处理]
                    exit /b 1
                ) else (
                    echo [端口 %port% 已释放]
                    exit /b 0
                )
            ) else (
                echo [无法获取占用端口 %port% 的进程ID]
                exit /b 1
            )
        ) else (
            echo [端口 %port% 可用]
            exit /b 0
        )
        exit /b 0
        
        REM 启动服务
        :start_service
        REM 检查并处理端口占用
        call :check_and_kill_port %PORT%
        if %errorlevel% neq 0 (
            echo [无法启动服务，端口 %PORT% 被占用且无法自动释放]
            exit /b 1
        )
        
        if "%DAEMON_MODE%"=="true" (
            echo [以后台模式启动服务...]
            start /B go run main.go > "%LOG_FILE%" 2>&1
            echo !PID_FILE! > "%PID_FILE%"
            echo [服务已启动]
            echo [日志文件: %LOG_FILE%]
            timeout /t 2 /nobreak >nul
            call :check_status
        ) else (
            echo [启动 TalkAI OpenAI API 适配器...]
            echo.
            go run main.go
        )
        exit /b 0
        
        REM 主函数
        :main
        if "%HELP%"=="true" (
            call :print_help
            exit /b 0
        )
        
        if "%VERSION%"=="true" (
            echo TalkAI OpenAI API 适配器启动脚本 v%SCRIPT_VERSION%
            exit /b 0
        )
        
        if "%INIT_CONFIG%"=="true" (
            call :init_config
            exit /b 0
        )
        
        if "%STATUS%"=="true" (
            call :check_status
            exit /b 0
        )
        
        if "%STOP%"=="true" (
            call :stop_service
            exit /b 0
        )
        
        REM 检查是否已经在运行
        call :check_status
        if %errorlevel% equ 0 (
            echo [服务已经在运行]
            echo [使用 --stop 参数停止服务]
            exit /b 1
        )
        
        call :check_dependencies
        call :load_config_file
        call :set_env_vars
        call :print_config
        
        call :start_service
        exit /b 0
        
        REM 运行主函数
        call :main