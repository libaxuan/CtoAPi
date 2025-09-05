# TalkAI OpenAI API 兼容适配器

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)
[![Release](https://img.shields.io/badge/Release-v1.0.0-blue.svg?style=for-the-badge)](https://github.com/yourusername/CtoAPi/releases)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg?style=for-the-badge&logo=docker)](https://hub.docker.com/r/yourusername/cto-api)

这是一个将 TalkAI 服务转换为 OpenAI API 兼容接口的适配器，使用 Go 语言开发。它允许你使用标准的 OpenAI API 格式与 TalkAI 的 Claude 模型进行交互，支持流式和非流式响应。

## 📋 目录

- [✨ 主要功能](#-主要功能)
- [🚀 快速开始](#-快速开始)
  - [环境要求](#环境要求)
  - [本地部署](#本地部署)
  - [使用一键启动脚本](#1-使用一键启动脚本推荐)
  - [使用 env.local 配置文件](#2-使用-envlocal-配置文件推荐用于本地开发)
  - [直接使用 Go 命令](#3-直接使用-go-命令)
- [⚙️ 环境变量配置](#️-环境变量配置)
  - [快速开始](#-快速开始-1)
  - [环境变量列表](#-环境变量列表)
  - [配置文件](#-配置文件)
  - [获取 TalkAI API 密钥](#-获取-talkai-api-密钥)
  - [使用示例](#-使用示例)
  - [重启服务](#-重启服务)
  - [注意事项](#️-注意事项)
- [支持的模型](#支持的模型)
- [📖 API使用示例](#-api使用示例)
  - [Python示例](#python示例)
  - [curl示例](#curl示例)
  - [JavaScript示例](#javascript示例)
- [API 接口](#api-接口)
  - [获取模型列表](#获取模型列表)
  - [聊天完成（非流式）](#聊天完成非流式)
  - [聊天完成（流式）](#聊天完成流式)
- [API 密钥管理](#api-密钥管理)
- [配置优先级](#配置优先级)
- [示例用法](#示例用法)
  - [开发环境](#开发环境)
  - [生产环境](#生产环境)
  - [测试环境](#测试环境)
- [🔧 故障排除](#-故障排除)
  - [常见问题](#常见问题)
  - [调试模式](#调试模式)
  - [网络问题排查](#网络问题排查)
  - [性能优化](#性能优化)
  - [日志分析](#日志分析)
- [🤝 贡献指南](#-贡献指南)
  - [开发流程](#开发流程)
- [📄 许可证](#-许可证)
- [⚠️ 免责声明](#️-免责声明)
- [📞 联系方式](#-联系方式)

## ✨ 主要功能

- 🔄 **OpenAI API兼容**: 完全兼容OpenAI的API格式，无需修改客户端代码
- 🌊 **流式响应支持**: 支持实时流式输出，提供更好的用户体验
- 🔐 **身份验证**: 支持API密钥验证，确保服务安全
- 🛠️ **灵活配置**: 通过环境变量、配置文件和命令行参数进行灵活配置
- 🐳 **Docker支持**: 提供Docker容器化部署选项
- 🌍 **CORS支持**: 支持跨域请求，便于前端集成
- 📝 **完整接口**: 提供完整的OpenAI API兼容接口
- 🚀 **高性能**: 基于Go语言开发，提供高性能服务
- 🛠️ **一键启动**: 提供便捷的启动脚本，简化部署流程
- 📊 **实时监控仪表板**: 提供Web仪表板，实时显示API转发情况和统计信息
- 📚 **交互式API文档**: 提供详细的API文档，包含请求参数、响应格式和使用示例

## 🚀 快速开始

### 环境要求

- Go 1.19 或更高版本
- TalkAI 的访问令牌

### 本地部署

1. **进入项目目录**
   ```bash
   cd CtoAPi
   ```

2. **配置环境变量**
   ```bash
   cp env.local.example env.local
   # 编辑 env.local 文件，设置你的 API_KEYS
   ```

3. **启动服务**
   ```bash
   # 使用启动脚本（推荐）
   ./start.sh
   
   # 或直接运行
   go run main.go
   ```

4. **测试服务**
    ```bash
    curl http://localhost:9091/v1/models
    ```

5. **访问API文档**
    
    启动服务后，可以通过浏览器访问以下地址查看完整的API文档：
    ```
    http://localhost:9091/docs
    ```
    
    API文档提供了以下功能：
    - 详细的API端点说明
    - 请求参数和响应格式
    - 多种编程语言的使用示例（Python、cURL、JavaScript）
    - 错误处理说明

6. **访问Dashboard**
   
   启动服务后，可以通过浏览器访问以下地址查看实时监控仪表板：
   ```
   http://localhost:9091/dashboard
   ```
   
   Dashboard提供了以下功能：
   - 实时显示API请求统计信息（总请求数、成功请求数、失败请求数、平均响应时间）
   - 显示最近100条请求的详细信息（时间、方法、路径、状态码、耗时、客户端IP）
   - 响应时间趋势图表
   - 数据每5秒自动刷新一次
   
   注意：Dashboard功能可以通过环境变量 `DASHBOARD_ENABLED` 控制开启和关闭，默认为开启状态。

### 1. 使用一键启动脚本（推荐）

#### Linux/macOS

```bash
# 基本启动
./start.sh

# 初始化配置文件（首次使用推荐）
./start.sh --init

# 使用命令行参数
./start.sh --port 8002 --api-keys sk-key1,sk-key2 --stream true

# 使用配置文件
./start.sh --config config.env

# 后台运行
./start.sh --daemon

# 查看服务状态
./start.sh --status

# 停止服务
./start.sh --stop

# 查看帮助
./start.sh --help

# 查看版本
./start.sh --version
```

#### Windows

```cmd
# 基本启动
start.bat

# 初始化配置文件（首次使用推荐）
start.bat --init

# 使用命令行参数
start.bat --port 8002 --api-keys sk-key1,sk-key2 --stream true

# 使用配置文件
start.bat --config config.env

# 后台运行
start.bat --daemon

# 查看服务状态
start.bat --status

# 停止服务
start.bat --stop

# 查看帮助
start.bat --help

# 查看版本
start.bat --version
```

> **注意**: 启动脚本会自动检测并处理端口占用问题，如果9091端口被占用，会自动终止占用进程。

### 2. 使用 env.local 配置文件（推荐用于本地开发）

```bash
# 使用启动脚本自动创建配置文件（推荐）
./start.sh --init

# 或手动复制示例配置文件
cp env.local.example env.local

# 编辑配置文件
nano env.local

# 启动服务器（会自动加载 env.local）
./start.sh
```

### 3. 直接使用 Go 命令

```bash
# 使用环境变量
PORT=8002 API_KEYS=sk-key1,sk-key2 DEFAULT_STREAM=true go run main.go

# 直接运行（使用默认配置）
go run main.go
```

## ⚙️ 环境变量配置

本项目支持通过环境变量进行配置，提供灵活的部署和运行选项。

### 🚀 快速开始

#### 1. 使用启动脚本（推荐）

**Linux/macOS:**
```bash
./start.sh
```

**Windows:**
```cmd
start.bat
```

#### 2. 手动设置环境变量

**Linux/macOS:**
```bash
export PORT="9091"
export API_KEYS="sk-talkai-key1,sk-talkai-key2"
export DEFAULT_MODEL="claude-opus-4-1-20250805"
export DEFAULT_STREAM="true"
go run main.go
```

**Windows:**
```cmd
set PORT=9091
set API_KEYS=sk-talkai-key1,sk-talkai-key2
set DEFAULT_MODEL=claude-opus-4-1-20250805
set DEFAULT_STREAM=true
go run main.go
```

#### 3. Docker运行

```bash
docker run -p 9091:9091 \
  -e API_KEYS=sk-talkai-key1,sk-talkai-key2 \
  -e DEFAULT_MODEL=claude-opus-4-1-20250805 \
  -e PORT=9091 \
  cto-api
```

### 📋 环境变量列表

#### 🔑 必需配置

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `API_KEYS` | TalkAI API密钥列表，多个密钥用逗号分隔 | 从文件读取 | `sk-talkai-key1,sk-talkai-key2` |

#### ⚙️ 可选配置

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `PORT` | 服务器端口 | `9091` | `9000` |
| `DEFAULT_STREAM` | 默认流模式 | `false` | `true` |
| `DEFAULT_MODEL` | 默认模型 | `claude-opus-4-1-20250805` | `claude-3-haiku-20240307` |
| `DEFAULT_TEMPERATURE` | 默认温度 | `0.7` | `0.5` |
| `TIMEOUT` | 请求超时时间（秒） | `300` | `600` |
| `DEBUG_MODE` | 调试模式 | `false` | `true` |
| `DASHBOARD_ENABLED` | Dashboard功能开关 | `true` | `false` |

#### 🔧 高级配置

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `UPSTREAM_URL` | 上游API地址 | TalkAI默认地址 | 自定义URL |

### 📁 配置文件

#### 支持的配置文件（按优先级排序）

1. `env.local` - 本地环境配置（推荐）
2. `env.local.example` - 配置模板

#### 配置文件示例

```bash
# 复制配置文件
cp env.local.example env.local

# 编辑配置文件
nano env.local
```

### 🔐 获取 TalkAI API 密钥

1. 登录 [TalkAI](https://talkai.com)
2. 在账户设置中找到 API 密钥
3. 复制密钥并添加到配置中

### 🎯 使用示例

#### 基本配置

```bash
# env.local
API_KEYS=sk-talkai-key1,sk-talkai-key2
DEFAULT_MODEL=claude-opus-4-1-20250805
PORT=9091
DEBUG_MODE=false
```

#### 生产环境配置

```bash
# env.production
API_KEYS=your_production_keys
DEFAULT_MODEL=claude-opus-4-1-20250805
PORT=9091
DEBUG_MODE=false
DEFAULT_STREAM=true
```

#### 开发环境配置

```bash
# env.development
API_KEYS=your_dev_keys
DEFAULT_MODEL=claude-3-sonnet-4-20250514
PORT=8002
DEBUG_MODE=true
DEFAULT_STREAM=true
```

### 🔄 重启服务

修改环境变量后，需要重启服务使配置生效：

```bash
# 停止当前服务
Ctrl+C

# 重新启动
./start.sh
```

### 📊 Dashboard功能

本项目提供了一个Web仪表板，用于实时监控API转发情况和统计信息。

#### 功能特点

- 实时显示API请求统计信息（总请求数、成功请求数、失败请求数、平均响应时间）
- 显示最近100条请求的详细信息（时间、方法、路径、状态码、耗时、客户端IP）
- 响应时间趋势图表
- 数据每5秒自动刷新一次
- 响应式设计，支持各种设备访问

#### 访问方式

启动服务后，通过浏览器访问以下地址：
```
http://localhost:9091/dashboard
```

#### 配置选项

通过 `DASHBOARD_ENABLED` 环境变量控制Dashboard功能的开启和关闭：

```bash
# 启用Dashboard（默认）
DASHBOARD_ENABLED=true

# 禁用Dashboard
DASHBOARD_ENABLED=false
```

#### 使用场景

- **开发调试**: 实时查看API请求情况，便于调试和问题排查
- **性能监控**: 监控API响应时间和成功率，评估系统性能
- **安全审计**: 查看请求来源和频率，发现异常访问模式

### 🚨 注意事项

1. **密钥安全**: 不要将真实的 TalkAI API 密钥提交到代码仓库
2. **配置文件**: 建议将 `env.local` 添加到 `.gitignore`
3. **权限设置**: 确保启动脚本有执行权限 (`chmod +x start.sh`)
4. **端口冲突**: 确保配置的端口没有被其他服务占用

### 命令行参数

| 参数 | 描述 |
|------|------|
| `-p, --port PORT` | 服务器端口 |
| `-k, --api-keys KEYS` | API 密钥列表 |
| `-s, --stream STREAM` | 默认流模式 |
| `-m, --model MODEL` | 默认模型 |
| `-t, --temperature TEMP` | 默认温度 |
| `-T, --timeout TIMEOUT` | 请求超时时间 |
| `-d, --debug MODE` | 调试模式 |
| `-c, --config FILE` | 配置文件路径 |
| `-i, --init` | 初始化配置文件 |
| `-D, --daemon` | 后台运行模式 |
| `--status` | 查看服务状态 |
| `--stop` | 停止服务 |
| `-v, --version` | 显示版本信息 |
| `-h, --help` | 显示帮助信息 |

### 配置文件

1. 复制示例配置文件：
```bash
cp env.local.example config.env
```

2. 编辑配置文件：
```bash
nano config.env
```

3. 使用配置文件启动：
```bash
./start.sh --config config.env
```

## 支持的模型

- Claude Opus 4.1 最新版 (`claude-opus-4-1-20250805`)
- Claude Opus 4 正式版 (`claude-opus-4-20250514`)
- Claude Sonnet 4 正式版 (`claude-sonnet-4-20250514`)
- Claude 3.7 Sonnet 版 (`claude-3-7-sonnet-20250219`)
- Claude 3.7 Sonnet 最新版 (`claude-3-7-sonnet-latest`)
- Claude 3.5 Haiku 最新版 (`claude-3-5-haiku-latest`)
- Claude 3.5 Haiku 版 (`claude-3-5-haiku-20241022`)
- Claude 3 Haiku 版 (`claude-3-haiku-20240307`)

## 📖 API使用示例

### Python示例

```python
import openai

# 配置客户端
client = openai.OpenAI(
    api_key="YOUR_API_KEY",  # 对应 env.local 中的 API_KEYS
    base_url="http://localhost:9091/v1"
)

# 非流式请求
response = client.chat.completions.create(
    model="claude-opus-4-1-20250805",
    messages=[{"role": "user", "content": "你好，请介绍一下自己"}]
)

print(response.choices[0].message.content)

# 流式请求
response = client.chat.completions.create(
    model="claude-opus-4-1-20250805",
    messages=[{"role": "user", "content": "请写一首关于春天的诗"}],
    stream=True
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

### curl示例

```bash
# 获取模型列表
curl -X GET http://localhost:9091/v1/models \
  -H "Authorization: Bearer YOUR_API_KEY"

# 非流式请求
curl -X POST http://localhost:9091/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "claude-opus-4-1-20250805",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": false
  }'

# 流式请求
curl -X POST http://localhost:9091/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "claude-opus-4-1-20250805",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": true
  }'
```

### JavaScript示例

```javascript
const fetch = require('node-fetch');

async function chatWithClaude(message, stream = false) {
  const response = await fetch('http://localhost:9091/v1/chat/completions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer YOUR_API_KEY'
    },
    body: JSON.stringify({
      model: 'claude-opus-4-1-20250805',
      messages: [{ role: 'user', content: message }],
      stream: stream
    })
  });

  if (stream) {
    // 处理流式响应
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      const chunk = decoder.decode(value);
      const lines = chunk.split('\n');
      
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6);
          if (data === '[DONE]') {
            console.log('\n流式响应完成');
            return;
          }
          
          try {
            const parsed = JSON.parse(data);
            const content = parsed.choices[0]?.delta?.content;
            if (content) {
              process.stdout.write(content);
            }
          } catch (e) {
            // 忽略解析错误
          }
        }
      }
    }
  } else {
    // 处理非流式响应
    const data = await response.json();
    console.log(data.choices[0].message.content);
  }
}

// 使用示例
chatWithClaude('你好，请介绍一下Claude', false);
```

## API 接口

### 获取模型列表

```bash
curl -X GET http://localhost:9091/v1/models \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### 聊天完成（非流式）

```bash
curl -X POST http://localhost:9091/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "claude-opus-4-1-20250805",
    "messages": [
      {"role": "user", "content": "你好，请介绍一下你自己"}
    ],
    "stream": false
  }'
```

### 聊天完成（流式）

```bash
curl -X POST http://localhost:9091/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "claude-opus-4-1-20250805",
    "messages": [
      {"role": "user", "content": "请用三句话描述人工智能"}
    ],
    "stream": true
  }'
```

## API 密钥管理

### 方式一：env.local 文件（推荐用于本地开发）

1. 使用启动脚本自动创建配置文件：
```bash
./start.sh --init
```

2. 或手动复制示例配置文件：
```bash
cp env.local.example env.local
```

3. 编辑 `env.local` 文件：
```
API_KEYS=sk-talkai-key1,sk-talkai-key2
```

4. 启动服务器（会自动加载 env.local）：
```bash
./start.sh
```

### 方式二：环境变量

```bash
export API_KEYS="sk-talkai-key1,sk-talkai-key2"
./start.sh
```

### 方式三：配置文件

在 `env.local.example` 文件中设置：
```
API_KEYS=sk-talkai-key1,sk-talkai-key2
```

使用配置文件启动：
```bash
./start.sh --config config.env
```

## 配置优先级

1. 命令行参数（最高优先级）
2. 环境变量
3. 配置文件（使用 `-c` 参数指定的文件）
4. env.local 文件（自动加载）
5. 默认值（最低优先级）

## 示例用法

### 开发环境

```bash
# 初始化配置
./start.sh --init

# 使用默认配置
./start.sh

# 或以后台模式运行
./start.sh --daemon
```

### 生产环境

```bash
# 使用配置文件
./start.sh --config production.env

# 以后台模式运行
./start.sh --daemon --config production.env

# 查看服务状态
./start.sh --status

# 停止服务
./start.sh --stop
```

### 测试环境

```bash
# 使用命令行参数
./start.sh \
  --port 8002 \
  --api-keys sk-test-key1,sk-test-key2 \
  --stream true \
  --debug true

# 以后台模式运行测试环境
./start.sh \
  --port 8002 \
  --api-keys sk-test-key1,sk-test-key2 \
  --stream true \
  --debug true \
  --daemon
```

## 🔧 故障排除

### 常见问题

1. **连接失败**
   - 检查服务是否正常运行：`curl http://localhost:9091/v1/models`
   - 访问API文档：`http://localhost:9091/docs`
   - 访问Dashboard：`http://localhost:9091/dashboard`
   - 确认端口配置正确
   - 检查防火墙设置

2. **认证失败**
   - 检查 `API_KEYS` 环境变量设置
   - 确认请求头中的 `Authorization` 格式正确
   - 验证 API 密钥是否有效

3. **TalkAI API密钥无效**
   - 检查 `API_KEYS` 环境变量设置
   - 确认密钥未过期
   - 验证密钥是否有足够的权限

4. **模型响应异常**
   - 检查 `DEFAULT_MODEL` 设置是否正确
   - 确认所请求的模型在支持列表中
   - 查看服务日志获取详细信息

5. **端口被占用**: 修改 `PORT` 环境变量或停止占用端口的服务
6. **权限不足**: 确保启动脚本有执行权限
7. **配置未生效**: 重启服务或检查配置文件语法
8. **流式响应问题**: 确认 `DEFAULT_STREAM` 设置正确，检查客户端是否支持流式响应
9. **Dashboard无法访问**: 确认 `DASHBOARD_ENABLED` 设置为 `true`，检查浏览器控制台错误
10. **图表显示异常**: 确认浏览器支持JavaScript，检查网络连接是否正常

### 调试模式

启用调试模式以获取详细日志：

```bash
export DEBUG_MODE=true
go run main.go
```

或使用启动脚本：

```bash
./start.sh --debug true
```

### 网络问题排查

如果遇到网络连接问题，可以尝试：

1. 检查防火墙设置
2. 确认 TalkAI 服务可访问
3. 测试网络连通性

### 性能优化

1. **减少日志输出**: 设置 `DEBUG_MODE=false`
2. **调整超时时间**: 修改 `TIMEOUT` 环境变量
3. **使用反向代理**: 在生产环境中建议使用 Nginx 等反向代理

### 日志分析

服务运行时会产生日志，包含以下信息：
- 请求详情（时间、方法、路径、状态码）
- 响应时间
- 错误信息（如果有）

通过分析日志可以帮助定位问题和优化性能。

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！请确保：

1. 代码符合 Go 的代码风格
2. 提交前运行测试
3. 更新相关文档
4. 遵循项目的代码结构和命名规范

### 开发流程

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/new-feature`
3. 提交更改：`git commit -am 'Add new feature'`
4. 推送分支：`git push origin feature/new-feature`
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## ⚠️ 免责声明

本项目与 TalkAI 官方无关，使用前请确保遵守 TalkAI 的服务条款。开发者不对因使用本项目而产生的任何问题负责。

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 Issue# CtoAPi
