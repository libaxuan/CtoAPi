#!/bin/bash

# TalkAI OpenAI API 适配器启动脚本
# 使用方法: ./start.sh [选项]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 版本信息
SCRIPT_VERSION="1.1.0"

# 默认配置
DEFAULT_PORT=9091
DEFAULT_STREAM=false
DEFAULT_MODEL="claude-opus-4-1-20250805"
DEFAULT_TEMPERATURE=0.7
DEFAULT_TIMEOUT=300
DEFAULT_DEBUG_MODE=false
LOG_FILE="talkai.log"
PID_FILE="talkai.pid"

# 配置变量
PORT=""
API_KEYS=""
STREAM=""
MODEL=""
TEMPERATURE=""
TIMEOUT=""
DEBUG_MODE=""
CONFIG_FILE=""
HELP=false
INIT_CONFIG=false
DAEMON_MODE=false
STATUS=false
STOP=false
VERSION=false

# 打印帮助信息
print_help() {
    echo -e "${BLUE}TalkAI OpenAI API 适配器启动脚本 v${SCRIPT_VERSION}${NC}"
    echo ""
    echo "使用方法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -p, --port PORT          服务器端口 (默认: $DEFAULT_PORT)"
    echo "  -k, --api-keys KEYS     API 密钥列表，多个密钥用逗号分隔"
    echo "  -s, --stream STREAM      默认流模式 (true/false, 默认: $DEFAULT_STREAM)"
    echo "  -m, --model MODEL        默认模型 (默认: $DEFAULT_MODEL)"
    echo "  -t, --temperature TEMP   默认温度 (默认: $DEFAULT_TEMPERATURE)"
    echo "  -T, --timeout TIMEOUT    请求超时时间，单位秒 (默认: $DEFAULT_TIMEOUT)"
    echo "  -d, --debug MODE         调试模式 (true/false, 默认: $DEFAULT_DEBUG_MODE)"
    echo "  -c, --config FILE        配置文件路径"
    echo "  -i, --init               初始化配置文件"
    echo "  -D, --daemon             后台运行模式"
    echo "  --status                 查看服务状态"
    echo "  --stop                   停止服务"
    echo "  -v, --version            显示版本信息"
    echo "  -h, --help               显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --port 8002 --api-keys sk-key1,sk-key2 --stream true"
    echo "  $0 --config config.env"
    echo "  $0 --init                 # 初始化配置"
    echo "  $0 --daemon               # 后台运行"
    echo "  $0 --status               # 查看状态"
    echo "  $0 --stop                 # 停止服务"
    echo ""
    echo "配置文件:"
    echo "  如果存在 env.local 文件，会自动加载其中的配置"
    echo "  也可以使用 -c 参数指定其他配置文件"
    echo ""
    echo "环境变量:"
    echo "  也可以通过环境变量设置配置，环境变量优先级高于配置文件"
    echo "  PORT, API_KEYS, DEFAULT_STREAM, DEFAULT_MODEL, DEFAULT_TEMPERATURE, TIMEOUT, DEBUG_MODE"
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--port)
                PORT="$2"
                shift 2
                ;;
            -k|--api-keys)
                API_KEYS="$2"
                shift 2
                ;;
            -s|--stream)
                STREAM="$2"
                shift 2
                ;;
            -m|--model)
                MODEL="$2"
                shift 2
                ;;
            -t|--temperature)
                TEMPERATURE="$2"
                shift 2
                ;;
            -T|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -d|--debug)
                DEBUG_MODE="$2"
                shift 2
                ;;
            -c|--config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            -i|--init)
                INIT_CONFIG=true
                shift
                ;;
            -D|--daemon)
                DAEMON_MODE=true
                shift
                ;;
            --status)
                STATUS=true
                shift
                ;;
            --stop)
                STOP=true
                shift
                ;;
            -v|--version)
                VERSION=true
                shift
                ;;
            -h|--help)
                HELP=true
                shift
                ;;
            *)
                echo -e "${RED}错误: 未知参数 $1${NC}"
                print_help
                exit 1
                ;;
        esac
    done
}

# 从配置文件读取配置
load_config_file() {
    # 如果指定了配置文件，则使用指定的配置文件
    if [[ -n "$CONFIG_FILE" && -f "$CONFIG_FILE" ]]; then
        echo -e "${YELLOW}从配置文件加载配置: $CONFIG_FILE${NC}"
        source "$CONFIG_FILE"
    # 如果没有指定配置文件，但存在 env.local 文件，则使用 env.local
    elif [[ -z "$CONFIG_FILE" && -f "env.local" ]]; then
        echo -e "${YELLOW}从 env.local 文件加载配置${NC}"
        source "env.local"
    fi
}

# 初始化配置文件
init_config() {
    echo -e "${CYAN}初始化配置文件...${NC}"
    
    if [[ -f "env.local" ]]; then
        echo -e "${YELLOW}env.local 文件已存在，是否覆盖？(y/N)${NC}"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            echo -e "${GREEN}保留现有配置文件${NC}"
            return 0
        fi
    fi
    
    cp env.local.example env.local
    echo -e "${GREEN}配置文件已创建: env.local${NC}"
    echo -e "${YELLOW}请编辑 env.local 文件以配置您的设置${NC}"
    
    # 询问是否立即编辑
    echo -e "${YELLOW}是否立即编辑配置文件？(Y/n)${NC}"
    read -r edit_response
    if [[ -z "$edit_response" || "$edit_response" =~ ^[Yy]$ ]]; then
        ${EDITOR:-nano} env.local
    fi
}

# 检查服务状态
check_status() {
    if [[ -f "$PID_FILE" ]]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${GREEN}服务正在运行 (PID: $PID)${NC}"
            return 0
        else
            echo -e "${YELLOW}PID 文件存在但服务未运行，清理中...${NC}"
            rm -f "$PID_FILE"
            return 1
        fi
    else
        echo -e "${YELLOW}服务未运行${NC}"
        return 1
    fi
}

# 停止服务
stop_service() {
    if [[ -f "$PID_FILE" ]]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${YELLOW}正在停止服务 (PID: $PID)...${NC}"
            kill "$PID"
            sleep 2
            
            if ps -p "$PID" > /dev/null 2>&1; then
                echo -e "${YELLOW}强制停止服务...${NC}"
                kill -9 "$PID"
            fi
            
            rm -f "$PID_FILE"
            echo -e "${GREEN}服务已停止${NC}"
        else
            echo -e "${YELLOW}服务未运行，清理 PID 文件...${NC}"
            rm -f "$PID_FILE"
        fi
    else
        echo -e "${YELLOW}服务未运行${NC}"
    fi
}

# 记录日志
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# 设置环境变量
set_env_vars() {
    echo -e "${YELLOW}设置环境变量...${NC}"
    
    # 命令行参数优先级最高
    [[ -n "$PORT" ]] && export PORT="$PORT"
    [[ -n "$API_KEYS" ]] && export API_KEYS="$API_KEYS"
    [[ -n "$STREAM" ]] && export DEFAULT_STREAM="$STREAM"
    [[ -n "$MODEL" ]] && export DEFAULT_MODEL="$MODEL"
    [[ -n "$TEMPERATURE" ]] && export DEFAULT_TEMPERATURE="$TEMPERATURE"
    [[ -n "$TIMEOUT" ]] && export TIMEOUT="$TIMEOUT"
    [[ -n "$DEBUG_MODE" ]] && export DEBUG_MODE="$DEBUG_MODE"
    
    # 如果没有设置，则使用默认值
    [[ -z "$PORT" ]] && export PORT="$DEFAULT_PORT"
    [[ -z "$DEFAULT_STREAM" ]] && export DEFAULT_STREAM="$DEFAULT_STREAM"
    [[ -z "$DEFAULT_MODEL" ]] && export DEFAULT_MODEL="$DEFAULT_MODEL"
    [[ -z "$DEFAULT_TEMPERATURE" ]] && export DEFAULT_TEMPERATURE="$DEFAULT_TEMPERATURE"
    [[ -z "$TIMEOUT" ]] && export TIMEOUT="$DEFAULT_TIMEOUT"
    [[ -z "$DEBUG_MODE" ]] && export DEBUG_MODE="$DEFAULT_DEBUG_MODE"
    
    # 记录配置到日志
    log "INFO" "配置: PORT=$PORT, MODEL=$DEFAULT_MODEL, STREAM=$DEFAULT_STREAM, DEBUG=$DEBUG_MODE"
    
    echo -e "${GREEN}环境变量设置完成${NC}"
}

# 打印配置信息
print_config() {
    echo -e "${GREEN}当前配置:${NC}"
    echo "  端口: $PORT"
    if [[ -n "$API_KEYS" ]]; then
        echo "  API 密钥: 已配置 ($(echo "$API_KEYS" | tr ',' '\n' | wc -l | tr -d ' ') 个)"
    else
        echo "  API 密钥: 从文件读取"
    fi
    echo "  默认流模式: $DEFAULT_STREAM"
    echo "  默认模型: $DEFAULT_MODEL"
    echo "  默认温度: $DEFAULT_TEMPERATURE"
    echo "  超时时间: $TIMEOUT 秒"
    echo "  调试模式: $DEBUG_MODE"
    echo ""
}

# 检查依赖
check_dependencies() {
    # 检查 Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}错误: 未找到 Go 命令${NC}"
        echo -e "${YELLOW}正在尝试安装 Go...${NC}"
        
        # 检测操作系统
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux
            if command -v apt-get &> /dev/null; then
                sudo apt-get update && sudo apt-get install -y golang-go
            elif command -v yum &> /dev/null; then
                sudo yum install -y golang
            elif command -v pacman &> /dev/null; then
                sudo pacman -S go
            else
                echo -e "${RED}无法自动安装 Go，请手动安装${NC}"
                exit 1
            fi
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            if command -v brew &> /dev/null; then
                brew install go
            else
                echo -e "${RED}请先安装 Homebrew 或手动安装 Go${NC}"
                exit 1
            fi
        else
            echo -e "${RED}不支持的操作系统，请手动安装 Go${NC}"
            exit 1
        fi
        
        # 再次检查
        if ! command -v go &> /dev/null; then
            echo -e "${RED}Go 安装失败，请手动安装${NC}"
            exit 1
        fi
        
        echo -e "${GREEN}Go 安装成功${NC}"
    fi
    
    # 检查 main.go
    if [[ ! -f "main.go" ]]; then
        echo -e "${RED}错误: 未找到 main.go 文件${NC}"
        exit 1
    fi
    
    # 检查 env.local.example
    if [[ ! -f "env.local.example" ]]; then
        echo -e "${RED}错误: 未找到 env.local.example 文件${NC}"
        exit 1
    fi
}

# 检查并处理端口占用
check_and_kill_port() {
    local port="$1"
    
    # 检查端口是否被占用
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${YELLOW}端口 $port 已被占用，尝试终止占用进程...${NC}"
        
        # 获取占用端口的进程ID
        local pid=$(lsof -Pi :$port -sTCP:LISTEN -t)
        
        if [[ -n "$pid" ]]; then
            # 终止进程
            kill -9 "$pid"
            echo -e "${GREEN}已终止占用端口 $port 的进程 (PID: $pid)${NC}"
            
            # 等待进程完全退出
            sleep 2
            
            # 再次检查端口是否仍被占用
            if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
                echo -e "${RED}无法释放端口 $port，请手动处理${NC}"
                return 1
            else
                echo -e "${GREEN}端口 $port 已释放${NC}"
                return 0
            fi
        else
            echo -e "${RED}无法获取占用端口 $port 的进程ID${NC}"
            return 1
        fi
    else
        echo -e "${GREEN}端口 $port 可用${NC}"
        return 0
    fi
}

# 启动服务
start_service() {
    # 检查并处理端口占用
    if ! check_and_kill_port "$PORT"; then
        echo -e "${RED}无法启动服务，端口 $port 被占用且无法自动释放${NC}"
        exit 1
    fi
    
    if [[ "$DAEMON_MODE" == true ]]; then
        echo -e "${YELLOW}以后台模式启动服务...${NC}"
        nohup go run main.go > "$LOG_FILE" 2>&1 &
        echo $! > "$PID_FILE"
        echo -e "${GREEN}服务已启动 (PID: $!)${NC}"
        echo -e "${CYAN}日志文件: $LOG_FILE${NC}"
        echo -e "${CYAN}服务启动后，您可以通过以下地址访问：${NC}"
        echo -e "${BLUE}  API 文档: http://localhost:$PORT/docs${NC}"
        echo -e "${BLUE}  监控面板: http://localhost:$PORT/dashboard${NC}"
        sleep 2
        check_status
    else
        echo -e "${GREEN}启动 TalkAI OpenAI API 适配器...${NC}"
        echo ""
        echo -e "${CYAN}服务启动后，您可以通过以下地址访问：${NC}"
        echo -e "${BLUE}  API 文档: http://localhost:$PORT/docs${NC}"
        echo -e "${BLUE}  监控面板: http://localhost:$PORT/dashboard${NC}"
        echo ""
        exec go run main.go
    fi
}

# 主函数
main() {
    parse_args "$@"
    
    if [[ "$HELP" == true ]]; then
        print_help
        exit 0
    fi
    
    if [[ "$VERSION" == true ]]; then
        echo -e "${CYAN}TalkAI OpenAI API 适配器启动脚本 v${SCRIPT_VERSION}${NC}"
        exit 0
    fi
    
    if [[ "$INIT_CONFIG" == true ]]; then
        init_config
        exit 0
    fi
    
    if [[ "$STATUS" == true ]]; then
        check_status
        exit 0
    fi
    
    if [[ "$STOP" == true ]]; then
        stop_service
        exit 0
    fi
    
    # 检查是否已经在运行
    if check_status; then
        echo -e "${YELLOW}服务已经在运行${NC}"
        echo -e "${YELLOW}使用 --stop 参数停止服务${NC}"
        exit 1
    fi
    
    check_dependencies
    load_config_file
    set_env_vars
    print_config
    
    start_service
}

# 运行主函数
main "$@"