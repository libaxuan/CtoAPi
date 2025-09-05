package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatMessage 聊天消息结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest 聊天完成请求结构
type ChatCompletionRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Stream      bool           `json:"stream"`
	Temperature *float64       `json:"temperature,omitempty"`
}

// ModelInfo 模型信息结构
type ModelInfo struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Created   int64  `json:"created"`
	OwnedBy   string `json:"owned_by"`
}

// ModelList 模型列表结构
type ModelList struct {
	Object string     `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// ChatCompletionChoice 聊天完成选择结构
type ChatCompletionChoice struct {
	Message      ChatMessage `json:"message"`
	Index        int         `json:"index"`
	FinishReason string      `json:"finish_reason"`
}

// ChatCompletionResponse 聊天完成响应结构
type ChatCompletionResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []ChatCompletionChoice   `json:"choices"`
	Usage   map[string]int           `json:"usage"`
}

// StreamChoice 流式选择结构
type StreamChoice struct {
	Delta        map[string]interface{} `json:"delta"`
	Index        int                    `json:"index"`
	FinishReason *string                `json:"finish_reason,omitempty"`
}

// StreamResponse 流式响应结构
type StreamResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// TalkAIMessage TalkAI 消息结构
type TalkAIMessage struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	Content string `json:"content"`
}

// TalkAIRequest TalkAI 请求结构
type TalkAIRequest struct {
	Type            string         `json:"type"`
	MessagesHistory []TalkAIMessage `json:"messagesHistory"`
	Settings        map[string]interface{} `json:"settings"`
}

// Config 应用配置结构
type Config struct {
	Port            int      `env:"PORT" envDefault:"9091"`
	APIKeys         []string `env:"API_KEYS" envDefault:""`
	DefaultStream   bool     `env:"DEFAULT_STREAM" envDefault:"false"`
	DefaultModel    string   `env:"DEFAULT_MODEL" envDefault:"claude-opus-4-1-20250805"`
	DefaultTemp     float64  `env:"DEFAULT_TEMPERATURE" envDefault:"0.7"`
	Timeout         int      `env:"TIMEOUT" envDefault:"300"`
	DebugMode       bool     `env:"DEBUG_MODE" envDefault:"false"`
	DashboardEnabled bool     `env:"DASHBOARD_ENABLED" envDefault:"true"`
}

// 请求统计信息
type RequestStats struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	LastRequestTime     time.Time     `json:"last_request_time"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
}

// 实时请求信息
type LiveRequest struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Duration  int64     `json:"duration"`
	UserAgent string    `json:"user_agent"`
}

var (
	config         Config
	validClientKeys = make(map[string]bool)
	modelsMap       = make(map[string]string)
	stats          RequestStats
	liveRequests   = []LiveRequest{}
	statsMutex     sync.Mutex
	requestsMutex  sync.Mutex
)

// loadConfig 从环境变量加载配置
func loadConfig() {
	// 设置默认配置
	config = Config{
		Port:            9091,
		DefaultStream:   false,
		DefaultModel:    "claude-opus-4-1-20250805",
		DefaultTemp:     0.7,
		Timeout:         300,
		DebugMode:       false,
		DashboardEnabled: true,
	}

	// 从环境变量读取配置
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}

	if apiKeys := os.Getenv("API_KEYS"); apiKeys != "" {
		config.APIKeys = strings.Split(apiKeys, ",")
		// 去除每个密钥的空白字符
		for i, key := range config.APIKeys {
			config.APIKeys[i] = strings.TrimSpace(key)
		}
	}

	if defaultStream := os.Getenv("DEFAULT_STREAM"); defaultStream != "" {
		if b, err := strconv.ParseBool(defaultStream); err == nil {
			config.DefaultStream = b
		}
	}

	if defaultModel := os.Getenv("DEFAULT_MODEL"); defaultModel != "" {
		config.DefaultModel = defaultModel
	}

	if defaultTemp := os.Getenv("DEFAULT_TEMPERATURE"); defaultTemp != "" {
		if t, err := strconv.ParseFloat(defaultTemp, 64); err == nil {
			config.DefaultTemp = t
		}
	}

	if timeout := os.Getenv("TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Timeout = t
		}
	}

	if debugMode := os.Getenv("DEBUG_MODE"); debugMode != "" {
		if b, err := strconv.ParseBool(debugMode); err == nil {
			config.DebugMode = b
		}
	}

	if dashboardEnabled := os.Getenv("DASHBOARD_ENABLED"); dashboardEnabled != "" {
		if b, err := strconv.ParseBool(dashboardEnabled); err == nil {
			config.DashboardEnabled = b
		}
	}
}

func init() {
	loadConfig()
	loadClientAPIKeys()
	loadModels()
}

func loadClientAPIKeys() {
	// 从环境变量加载 API 密钥
	if len(config.APIKeys) > 0 {
		for _, key := range config.APIKeys {
			if key != "" {
				validClientKeys[key] = true
			}
		}
		log.Printf("从环境变量加载了 %d 个 API 密钥", len(config.APIKeys))
		return
	}

	// 如果环境变量中没有配置 API 密钥，则生成一个默认密钥
	defaultKey := fmt.Sprintf("sk-talkai-%s", uuid.New().String())
	validClientKeys[defaultKey] = true
	log.Printf("已生成默认 API 密钥: %s", defaultKey)
	log.Printf("要设置您自己的 API 密钥，请使用 API_KEYS 环境变量或 env.local 文件")
}

func loadModels() {
	data, err := os.ReadFile("models.json")
	if err != nil {
		log.Printf("读取 models.json 出错: %v", err)
		return
	}

	if err := json.Unmarshal(data, &modelsMap); err != nil {
		log.Printf("解析 models.json 出错: %v", err)
		return
	}
}

func authenticateClient(c *gin.Context) {
	// 如果没有配置客户端密钥，则跳过认证
	if len(validClientKeys) == 0 {
		c.Next()
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
		c.Abort()
		return
	}

	// 提取 Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		c.Abort()
		return
	}

	token := parts[1]
	if !validClientKeys[token] {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		c.Abort()
		return
	}

	c.Next()
}

// 记录请求统计信息
func recordRequestStats(startTime time.Time, path string, status int) {
	duration := time.Since(startTime)
	
	statsMutex.Lock()
	defer statsMutex.Unlock()
	
	stats.TotalRequests++
	stats.LastRequestTime = time.Now()
	
	if status >= 200 && status < 300 {
		stats.SuccessfulRequests++
	} else {
		stats.FailedRequests++
	}
	
	// 更新平均响应时间
	if stats.TotalRequests > 0 {
		totalDuration := stats.AverageResponseTime*time.Duration(stats.TotalRequests-1) + duration
		stats.AverageResponseTime = totalDuration / time.Duration(stats.TotalRequests)
	} else {
		stats.AverageResponseTime = duration
	}
}

// 添加实时请求信息
func addLiveRequest(method, path string, status int, duration time.Duration, _, userAgent string) {
	requestsMutex.Lock()
	defer requestsMutex.Unlock()
	
	request := LiveRequest{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Method:    method,
		Path:      path,
		Status:    status,
		Duration:  duration.Milliseconds(),
		UserAgent: userAgent,
	}
	
	liveRequests = append(liveRequests, request)
	
	// 只保留最近的100条请求
	if len(liveRequests) > 100 {
		liveRequests = liveRequests[1:]
	}
}

// 获取实时请求数据（用于SSE）
func getLiveRequestsData() []byte {
	requestsMutex.Lock()
	defer requestsMutex.Unlock()
	
	// 确保 liveRequests 不为 nil
	if liveRequests == nil {
		liveRequests = []LiveRequest{}
	}
	
	data, err := json.Marshal(liveRequests)
	if err != nil {
		// 如果序列化失败，返回空数组
		emptyArray := []LiveRequest{}
		data, _ = json.Marshal(emptyArray)
	}
	return data
}

// 获取统计数据（用于SSE）
func getStatsData() []byte {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	
	data, _ := json.Marshal(stats)
	return data
}

func listModels(c *gin.Context) {
	var models []ModelInfo
	for _, modelID := range modelsMap {
		models = append(models, ModelInfo{
			ID:        modelID,
			Object:    "model",
			Created:   time.Now().Unix(),
			OwnedBy:   "talkai",
		})
	}

	c.JSON(http.StatusOK, ModelList{
		Object: "list",
		Data:   models,
	})
}

func chatCompletions(c *gin.Context) {
	startTime := time.Now()
	path := c.Request.URL.Path
	userAgent := c.Request.UserAgent()
	
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		// 记录请求统计
		duration := time.Since(startTime)
		recordRequestStats(startTime, path, http.StatusBadRequest)
		addLiveRequest("POST", path, http.StatusBadRequest, duration, "", userAgent)
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Messages required"})
		// 记录请求统计
		duration := time.Since(startTime)
		recordRequestStats(startTime, path, http.StatusBadRequest)
		addLiveRequest("POST", path, http.StatusBadRequest, duration, "", userAgent)
		return
	}

	// 应用默认配置
	if req.Model == "" {
		req.Model = config.DefaultModel
	}
	
	if req.Temperature == nil {
		req.Temperature = &config.DefaultTemp
	}
	
	// 如果请求中没有指定流模式，则使用环境变量中的默认值
	if c.Request.URL.Query().Get("stream") == "" && !req.Stream {
		req.Stream = config.DefaultStream
	}

	// 处理消息历史
	messagesHistory := []TalkAIMessage{}
	systemPrompt := ""

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else if msg.Role == "user" || msg.Role == "assistant" {
			from := "you"
			if msg.Role == "assistant" {
				from = "assistant"
			}
			messagesHistory = append(messagesHistory, TalkAIMessage{
				ID:      uuid.New().String(),
				From:    from,
				Content: msg.Content,
			})
		}
	}

	// 如果有系统提示且最后一条消息是用户消息，则合并
	if systemPrompt != "" && len(messagesHistory) > 0 && messagesHistory[len(messagesHistory)-1].From == "you" {
		messagesHistory[len(messagesHistory)-1].Content = fmt.Sprintf("%s\n\n%s", systemPrompt, messagesHistory[len(messagesHistory)-1].Content)
	}

	// 构建 TalkAI 请求
	talkAIReq := TalkAIRequest{
		Type:            "chat",
		MessagesHistory: messagesHistory,
		Settings: map[string]interface{}{
			"model":       req.Model,
			"temperature": req.Temperature,
		},
	}

	// 设置默认温度
	if req.Temperature == nil {
		talkAIReq.Settings["temperature"] = 0.7
	}

	// 发送请求到 TalkAI
	resp, err := sendToTalkAI(talkAIReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		// 记录请求统计
		duration := time.Since(startTime)
		recordRequestStats(startTime, path, http.StatusInternalServerError)
		addLiveRequest("POST", path, http.StatusInternalServerError, duration, "", userAgent)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "TalkAI API error"})
		// 记录请求统计
		duration := time.Since(startTime)
		recordRequestStats(startTime, path, resp.StatusCode)
		addLiveRequest("POST", path, resp.StatusCode, duration, "", userAgent)
		return
	}

	if req.Stream {
		handleStreamResponse(c, resp, req.Model)
		// 记录成功请求统计
		duration := time.Since(startTime)
		recordRequestStats(startTime, path, http.StatusOK)
		addLiveRequest("POST", path, http.StatusOK, duration, "", userAgent)
	} else {
		handleNormalResponse(c, resp, req.Model)
		// 记录成功请求统计
		duration := time.Since(startTime)
		recordRequestStats(startTime, path, http.StatusOK)
		addLiveRequest("POST", path, http.StatusOK, duration, "", userAgent)
	}
}

func sendToTalkAI(req TalkAIRequest) (*http.Response, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", "https://claude.talkai.info/chat/send/", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	return client.Do(httpReq)
}

func handleNormalResponse(c *gin.Context, resp *http.Response, model string) {
	// 这里需要解析 TalkAI 的响应并转换为 OpenAI 格式
	// 由于 TalkAI 返回的是流式格式，我们需要聚合所有内容
	content := aggregateStreamContent(resp)

	response := ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", uuid.New().String()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChatCompletionChoice{
			{
				Message: ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				Index:        0,
				FinishReason: "stop",
			},
		},
		Usage: map[string]int{
			"prompt_tokens":     0,
			"completion_tokens": 0,
			"total_tokens":      0,
		},
	}

	c.JSON(http.StatusOK, response)
}

func handleStreamResponse(c *gin.Context, resp *http.Response, model string) {
	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	streamID := fmt.Sprintf("chatcmpl-%s", uuid.New().String())
	createdTime := time.Now().Unix()

	// 发送初始消息
	initialChoice := StreamChoice{
		Delta: map[string]interface{}{
			"role": "assistant",
		},
		Index: 0,
	}

	initialResp := StreamResponse{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: createdTime,
		Model:   model,
		Choices: []StreamChoice{initialChoice},
	}

	c.Stream(func(w io.Writer) bool {
		// 发送初始响应
		jsonData, _ := json.Marshal(initialResp)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		w.(http.Flusher).Flush()

		// 处理流式内容
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				content := strings.TrimSpace(line[5:])
				if content != "" && content != "-1" {
					choice := StreamChoice{
						Delta: map[string]interface{}{
							"content": content,
						},
						Index: 0,
					}

					streamResp := StreamResponse{
						ID:      streamID,
						Object:  "chat.completion.chunk",
						Created: createdTime,
						Model:   model,
						Choices: []StreamChoice{choice},
					}

					jsonData, _ := json.Marshal(streamResp)
					fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
					w.(http.Flusher).Flush()
				}
			}
		}

		// 发送结束消息
		finishReason := "stop"
		finalChoice := StreamChoice{
			Delta:        map[string]interface{}{},
			Index:        0,
			FinishReason: &finishReason,
		}

		finalResp := StreamResponse{
			ID:      streamID,
			Object:  "chat.completion.chunk",
			Created: createdTime,
			Model:   model,
			Choices: []StreamChoice{finalChoice},
		}

		jsonData, _ = json.Marshal(finalResp)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		fmt.Fprintf(w, "data: [DONE]\n\n")
		w.(http.Flusher).Flush()

		return false
	})
}

func aggregateStreamContent(resp *http.Response) string {
	var content strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(line[5:])
			if data != "" && data != "-1" {
				content.WriteString(data)
			}
		}
	}
	
	return content.String()
}

// Dashboard页面处理器
func handleDashboard(c *gin.Context) {
	// 简单的HTML模板
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	   <meta charset="UTF-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1.0">
	   <title>CtoAPi 调用看板</title>
	   <style>
	       body {
	           font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
	           margin: 0;
	           padding: 20px;
	           background-color: #f5f5f5;
	       }
	       .container {
	           max-width: 1200px;
	           margin: 0 auto;
	           background-color: white;
	           border-radius: 8px;
	           box-shadow: 0 2px 10px rgba(0,0,0,0.1);
	           padding: 20px;
	       }
	       h1 {
	           color: #333;
	           text-align: center;
	           margin-bottom: 30px;
	       }
	       .stats-container {
	           display: grid;
	           grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
	           gap: 20px;
	           margin-bottom: 30px;
	       }
	       .stat-card {
	           background-color: #f8f9fa;
	           border-radius: 6px;
	           padding: 15px;
	           text-align: center;
	           box-shadow: 0 1px 3px rgba(0,0,0,0.1);
	       }
	       .stat-value {
	           font-size: 24px;
	           font-weight: bold;
	           color: #007bff;
	       }
	       .stat-label {
	           font-size: 14px;
	           color: #6c757d;
	           margin-top: 5px;
	       }
	       .requests-container {
	           margin-top: 30px;
	       }
	       .requests-table {
	           width: 100%;
	           border-collapse: collapse;
	       }
	       .requests-table th, .requests-table td {
	           padding: 10px;
	           text-align: left;
	           border-bottom: 1px solid #ddd;
	       }
	       .requests-table th {
	           background-color: #f8f9fa;
	       }
	       .status-success {
	           color: #28a745;
	       }
	       .status-error {
	           color: #dc3545;
	       }
	       .refresh-info {
	           text-align: center;
	           margin-top: 20px;
	           color: #6c757d;
	           font-size: 14px;
	       }
	       .pagination-container {
	           display: flex;
	           justify-content: center;
	           align-items: center;
	           margin-top: 20px;
	           gap: 10px;
	       }
	       .pagination-container button {
	           padding: 5px 10px;
	           background-color: #007bff;
	           color: white;
	           border: none;
	           border-radius: 4px;
	           cursor: pointer;
	       }
	       .pagination-container button:disabled {
	           background-color: #cccccc;
	           cursor: not-allowed;
	       }
	       .pagination-container button:hover:not(:disabled) {
	           background-color: #0056b3;
	       }
	       .chart-container {
	           margin-top: 30px;
	           height: 300px;
	           background-color: #f8f9fa;
	           border-radius: 6px;
	           padding: 15px;
	           box-shadow: 0 1px 3px rgba(0,0,0,0.1);
	       }
	   </style>
</head>
<body>
	   <div class="container">
	       <h1>CtoAPi 调用看板</h1>
	       
	       <div class="stats-container">
	           <div class="stat-card">
	               <div class="stat-value" id="total-requests">0</div>
	               <div class="stat-label">总请求数</div>
	           </div>
	           <div class="stat-card">
	               <div class="stat-value" id="successful-requests">0</div>
	               <div class="stat-label">成功请求</div>
	           </div>
	           <div class="stat-card">
	               <div class="stat-value" id="failed-requests">0</div>
	               <div class="stat-label">失败请求</div>
	           </div>
	           <div class="stat-card">
	               <div class="stat-value" id="avg-response-time">0s</div>
	               <div class="stat-label">平均响应时间</div>
	           </div>
	       </div>
	       
	       <div class="chart-container">
	           <h2>请求统计图表</h2>
	           <canvas id="requestsChart"></canvas>
	       </div>
	       
	       <div class="requests-container">
	           <h2>实时请求</h2>
	           <table class="requests-table">
	               <thead>
	                   <tr>
	                       <th>时间</th>
	                       <th>模型</th>
	                       <th>方法</th>
	                       <th>状态</th>
	                       <th>耗时</th>
	                       <th>User Agent</th>
	                   </tr>
	               </thead>
	               <tbody id="requests-tbody">
	                   <!-- 请求记录将通过JavaScript动态添加 -->
	               </tbody>
	           </table>
	           <div class="pagination-container">
	               <button id="prev-page" disabled>上一页</button>
	               <span id="page-info">第 1 页，共 1 页</span>
	               <button id="next-page" disabled>下一页</button>
	           </div>
	       </div>
	       
	       <div class="refresh-info">
	           数据每5秒自动刷新一次
	       </div>
	   </div>

	   <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
	   <script>
	       // 全局变量
	       let allRequests = [];
	       let currentPage = 1;
	       const itemsPerPage = 10;
	       let requestsChart = null;
	       
	       // 更新统计数据
	       function updateStats() {
	           fetch('/dashboard/stats')
	               .then(response => response.json())
	               .then(data => {
	                   document.getElementById('total-requests').textContent = data.total_requests;
	                   document.getElementById('successful-requests').textContent = data.successful_requests;
	                   document.getElementById('failed-requests').textContent = data.failed_requests;
	                   document.getElementById('avg-response-time').textContent = (data.average_response_time / 1000000000).toFixed(2) + 's';
	               })
	               .catch(error => console.error('Error fetching stats:', error));
	       }
	       
	       // 更新请求列表
	       function updateRequests() {
	           fetch('/dashboard/requests')
	               .then(response => response.json())
	               .then(data => {
	                   // 检查数据是否为数组
	                   if (!Array.isArray(data)) {
	                       console.error('返回的数据不是数组:', data);
	                       return;
	                   }
	                   
	                   // 保存所有请求数据
	                   allRequests = data;
	                   
	                   // 按时间倒序排列
	                   allRequests.sort((a, b) => {
	                       const timeA = new Date(a.timestamp);
	                       const timeB = new Date(b.timestamp);
	                       return timeB - timeA;
	                   });
	                   
	                   // 更新表格
	                   updateTable();
	                   
	                   // 更新图表
	                   updateChart();
	                   
	                   // 更新分页信息
	                   updatePagination();
	               })
	               .catch(error => console.error('Error fetching requests:', error));
	       }
	       
	       // 更新表格显示
	       function updateTable() {
	           const tbody = document.getElementById('requests-tbody');
	           tbody.innerHTML = '';
	           
	           // 计算当前页的数据范围
	           const startIndex = (currentPage - 1) * itemsPerPage;
	           const endIndex = startIndex + itemsPerPage;
	           const currentRequests = allRequests.slice(startIndex, endIndex);
	           
	           currentRequests.forEach(request => {
	               const row = document.createElement('tr');
	               
	               // 格式化时间 - 检查时间戳是否有效
	               let timeStr = "Invalid Date";
	               if (request.timestamp) {
	                   try {
	                       const time = new Date(request.timestamp);
	                       if (!isNaN(time.getTime())) {
	                           timeStr = time.toLocaleTimeString();
	                       }
	                   } catch (e) {
	                       console.error("时间格式化错误:", e);
	                   }
	               }
	               
	               // 状态样式
	               const statusClass = request.status >= 200 && request.status < 300 ? 'status-success' : 'status-error';
	               
	               // 截断 User Agent，避免过长
	               let userAgent = request.user_agent || "undefined";
	               if (userAgent.length > 30) {
	                   userAgent = userAgent.substring(0, 30) + "...";
	               }
	               
	               row.innerHTML =
	                  "<td>" + timeStr + "</td>" +
	                  "<td>Claude</td>" +
	                  "<td>" + (request.method || "undefined") + "</td>" +
	                  "<td class=\"" + statusClass + "\">" + (request.status || "undefined") + "</td>" +
	                  "<td>" + ((request.duration / 1000).toFixed(2) || "undefined") + "s</td>" +
	                  "<td title=\"" + (request.user_agent || "") + "\">" + userAgent + "</td>";
	               
	               tbody.appendChild(row);
	           });
	       }
	       
	       // 更新分页信息
	       function updatePagination() {
	           const totalPages = Math.ceil(allRequests.length / itemsPerPage);
	           document.getElementById('page-info').textContent = "第 " + currentPage + " 页，共 " + totalPages + " 页";
	           
	           document.getElementById('prev-page').disabled = currentPage <= 1;
	           document.getElementById('next-page').disabled = currentPage >= totalPages;
	       }
	       
	       // 更新图表
	       function updateChart() {
	           const ctx = document.getElementById('requestsChart').getContext('2d');
	           
	           // 准备图表数据 - 最近20条请求的响应时间
	           const chartData = allRequests.slice(0, 20).reverse();
	           const labels = chartData.map(req => {
	               const time = new Date(req.timestamp);
	               return time.toLocaleTimeString();
	           });
	           const responseTimes = chartData.map(req => req.duration);
	           
	           // 如果图表已存在，先销毁
	           if (requestsChart) {
	               requestsChart.destroy();
	           }
	           
	           // 创建新图表
	           requestsChart = new Chart(ctx, {
	               type: 'line',
	               data: {
	                   labels: labels,
	                   datasets: [{
	                       label: '响应时间 (s)',
	                       data: responseTimes.map(time => time / 1000),
	                       borderColor: '#007bff',
	                       backgroundColor: 'rgba(0, 123, 255, 0.1)',
	                       tension: 0.1,
	                       fill: true
	                   }]
	               },
	               options: {
	                   responsive: true,
	                   maintainAspectRatio: false,
	                   scales: {
	                       y: {
	                           beginAtZero: true,
	                           title: {
	                               display: true,
	                               text: '响应时间 (s)'
	                           }
	                       },
	                       x: {
	                           title: {
	                               display: true,
	                               text: '时间'
	                           }
	                       }
	                   },
	                   plugins: {
	                       title: {
	                           display: true,
	                           text: '最近20条请求的响应时间趋势 (s)'
	                       }
	                   }
	               }
	           });
	       }
	       
	       // 分页按钮事件
	       document.getElementById('prev-page').addEventListener('click', function() {
	           if (currentPage > 1) {
	               currentPage--;
	               updateTable();
	               updatePagination();
	           }
	       });
	       
	       document.getElementById('next-page').addEventListener('click', function() {
	           const totalPages = Math.ceil(allRequests.length / itemsPerPage);
	           if (currentPage < totalPages) {
	               currentPage++;
	               updateTable();
	               updatePagination();
	           }
	       });
	       
	       // 初始加载
	       updateStats();
	       updateRequests();
	       
	       // 定时刷新
	       setInterval(updateStats, 5000);
	       setInterval(updateRequests, 5000);
	   </script>
</body>
</html>`
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, tmpl)
}

// Dashboard统计数据处理器
func handleDashboardStats(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", getStatsData())
}

// Dashboard请求数据处理器
func handleDashboardRequests(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", getLiveRequestsData())
}

// API文档页面处理器
func handleDocs(c *gin.Context) {
	// API文档HTML模板
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	   <meta charset="UTF-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1.0">
	   <title>CtoAPi 文档</title>
	   <style>
	       body {
	           font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
	           margin: 0;
	           padding: 20px;
	           background-color: #f5f5f5;
	           line-height: 1.6;
	       }
	       .container {
	           max-width: 1200px;
	           margin: 0 auto;
	           background-color: white;
	           border-radius: 8px;
	           box-shadow: 0 2px 10px rgba(0,0,0,0.1);
	           padding: 30px;
	       }
	       h1 {
	           color: #333;
	           text-align: center;
	           margin-bottom: 30px;
	           border-bottom: 2px solid #007bff;
	           padding-bottom: 10px;
	       }
	       h2 {
	           color: #007bff;
	           margin-top: 30px;
	           margin-bottom: 15px;
	       }
	       h3 {
	           color: #333;
	           margin-top: 25px;
	           margin-bottom: 10px;
	       }
	       .endpoint {
	           background-color: #f8f9fa;
	           border-radius: 6px;
	           padding: 15px;
	           margin-bottom: 20px;
	           border-left: 4px solid #007bff;
	       }
	       .method {
	           display: inline-block;
	           padding: 4px 8px;
	           border-radius: 4px;
	           color: white;
	           font-weight: bold;
	           margin-right: 10px;
	           font-size: 14px;
	       }
	       .get { background-color: #28a745; }
	       .post { background-color: #007bff; }
	       .path {
	           font-family: monospace;
	           background-color: #e9ecef;
	           padding: 2px 6px;
	           border-radius: 3px;
	           font-size: 16px;
	       }
	       .description {
	           margin: 15px 0;
	       }
	       .parameters {
	           margin: 15px 0;
	       }
	       table {
	           width: 100%;
	           border-collapse: collapse;
	           margin: 15px 0;
	       }
	       th, td {
	           padding: 10px;
	           text-align: left;
	           border-bottom: 1px solid #ddd;
	       }
	       th {
	           background-color: #f8f9fa;
	           font-weight: bold;
	       }
	       .example {
	           background-color: #f8f9fa;
	           border-radius: 6px;
	           padding: 15px;
	           margin: 15px 0;
	           font-family: monospace;
	           white-space: pre-wrap;
	           overflow-x: auto;
	       }
	       .note {
	           background-color: #fff3cd;
	           border-left: 4px solid #ffc107;
	           padding: 10px 15px;
	           margin: 15px 0;
	           border-radius: 0 4px 4px 0;
	       }
	       .response {
	           background-color: #f8f9fa;
	           border-radius: 6px;
	           padding: 15px;
	           margin: 15px 0;
	           font-family: monospace;
	           white-space: pre-wrap;
	           overflow-x: auto;
	       }
	       .tab {
	           overflow: hidden;
	           border: 1px solid #ccc;
	           background-color: #f1f1f1;
	           border-radius: 4px 4px 0 0;
	       }
	       .tab button {
	           background-color: inherit;
	           float: left;
	           border: none;
	           outline: none;
	           cursor: pointer;
	           padding: 14px 16px;
	           transition: 0.3s;
	           font-size: 16px;
	       }
	       .tab button:hover {
	           background-color: #ddd;
	       }
	       .tab button.active {
	           background-color: #ccc;
	       }
	       .tabcontent {
	           display: none;
	           padding: 6px 12px;
	           border: 1px solid #ccc;
	           border-top: none;
	           border-radius: 0 0 4px 4px;
	       }
	       .toc {
	           background-color: #f8f9fa;
	           border-radius: 6px;
	           padding: 15px;
	           margin-bottom: 20px;
	       }
	       .toc ul {
	           padding-left: 20px;
	       }
	       .toc li {
	           margin: 5px 0;
	       }
	       .toc a {
	           color: #007bff;
	           text-decoration: none;
	       }
	       .toc a:hover {
	           text-decoration: underline;
	       }
	   </style>
</head>
<body>
	   <div class="container">
	       <h1>CtoAPi 文档</h1>
	       
	       <div class="toc">
	           <h2>目录</h2>
	           <ul>
	               <li><a href="#overview">概述</a></li>
	               <li><a href="#authentication">身份验证</a></li>
	               <li><a href="#endpoints">API端点</a>
	                   <ul>
	                       <li><a href="#models">获取模型列表</a></li>
	                       <li><a href="#chat-completions">聊天完成</a></li>
	                   </ul>
	               </li>
	               <li><a href="#examples">使用示例</a></li>
	               <li><a href="#error-handling">错误处理</a></li>
	           </ul>
	       </div>
	       
	       <section id="overview">
	           <h2>概述</h2>
	           <p>这是一个为TalkAI Claude模型提供OpenAI兼容API接口的代理服务器。它允许你使用标准的OpenAI API格式与TalkAI的Claude模型进行交互，支持流式和非流式响应。</p>
	           <p><strong>基础URL:</strong> <code>http://localhost:9091/v1</code></p>
	           <div class="note">
	               <strong>注意:</strong> 默认端口为9091，可以通过环境变量PORT进行修改。
	           </div>
	       </section>
	       
	       <section id="authentication">
	           <h2>身份验证</h2>
	           <p>所有API请求都需要在请求头中包含有效的API密钥进行身份验证：</p>
	           <div class="example">
Authorization: Bearer your-api-key</div>
	           <p>API密钥通过环境变量 API_KEYS 配置，多个密钥用逗号分隔。</p>
	       </section>
	       
	       <section id="endpoints">
	           <h2>API端点</h2>
	           
	           <div class="endpoint" id="models">
	               <h3>获取模型列表</h3>
	               <div>
	                   <span class="method get">GET</span>
	                   <span class="path">/v1/models</span>
	               </div>
	               <div class="description">
	                   <p>获取可用模型列表。</p>
	               </div>
	               <div class="parameters">
	                   <h4>请求参数</h4>
	                   <p>无</p>
	               </div>
	               <div class="response">
{
	 "object": "list",
	 "data": [
	   {
	     "id": "claude-opus-4-1-20250805",
	     "object": "model",
	     "created": 1756788845,
	     "owned_by": "talkai"
	   }
	 ]
}</div>
	           </div>
	           
	           <div class="endpoint" id="chat-completions">
	               <h3>聊天完成</h3>
	               <div>
	                   <span class="method post">POST</span>
	                   <span class="path">/v1/chat/completions</span>
	               </div>
	               <div class="description">
	                   <p>基于消息列表生成模型响应。支持流式和非流式两种模式。</p>
	               </div>
	               <div class="parameters">
	                   <h4>请求参数</h4>
	                   <table>
	                       <thead>
	                           <tr>
	                               <th>参数名</th>
	                               <th>类型</th>
	                               <th>必需</th>
	                               <th>说明</th>
	                           </tr>
	                       </thead>
	                       <tbody>
	                           <tr>
	                               <td>model</td>
	                               <td>string</td>
	                               <td>是</td>
	                               <td>要使用的模型ID，例如 "claude-opus-4-1-20250805"</td>
	                           </tr>
	                           <tr>
	                               <td>messages</td>
	                               <td>array</td>
	                               <td>是</td>
	                               <td>消息列表，包含角色和内容</td>
	                           </tr>
	                           <tr>
	                               <td>stream</td>
	                               <td>boolean</td>
	                               <td>否</td>
	                               <td>是否使用流式响应，默认为环境变量 DEFAULT_STREAM 的值</td>
	                           </tr>
	                           <tr>
	                               <td>temperature</td>
	                               <td>number</td>
	                               <td>否</td>
	                               <td>采样温度，控制随机性</td>
	                           </tr>
	                       </tbody>
	                   </table>
	               </div>
	               <div class="parameters">
	                   <h4>消息格式</h4>
	                   <table>
	                       <thead>
	                           <tr>
	                               <th>字段</th>
	                               <th>类型</th>
	                               <th>说明</th>
	                           </tr>
	                       </thead>
	                       <tbody>
	                           <tr>
	                               <td>role</td>
	                               <td>string</td>
	                               <td>消息角色，可选值：system、user、assistant</td>
	                           </tr>
	                           <tr>
	                               <td>content</td>
	                               <td>string</td>
	                               <td>消息内容</td>
	                           </tr>
	                       </tbody>
	                   </table>
	               </div>
	           </div>
	       </section>
	       
	       <section id="examples">
	           <h2>使用示例</h2>
	           
	           <div class="tab">
	               <button class="tablinks active" onclick="openTab(event, 'python-tab')">Python</button>
	               <button class="tablinks" onclick="openTab(event, 'curl-tab')">cURL</button>
	               <button class="tablinks" onclick="openTab(event, 'javascript-tab')">JavaScript</button>
	           </div>
	           
	           <div id="python-tab" class="tabcontent" style="display: block;">
	               <h3>Python示例</h3>
	               <div class="example">
import openai

# 配置客户端
client = openai.OpenAI(
	   api_key="your-api-key",  # 对应 API_KEYS
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
	       print(chunk.choices[0].delta.content, end="")</div>
	           </div>
	           
	           <div id="curl-tab" class="tabcontent">
	               <h3>cURL示例</h3>
	               <div class="example">
# 非流式请求
curl -X POST http://localhost:9091/v1/chat/completions \
	 -H "Content-Type: application/json" \
	 -H "Authorization: Bearer your-api-key" \
	 -d '{
	   "model": "claude-opus-4-1-20250805",
	   "messages": [{"role": "user", "content": "你好"}],
	   "stream": false
	 }'

# 流式请求
curl -X POST http://localhost:9091/v1/chat/completions \
	 -H "Content-Type: application/json" \
	 -H "Authorization: Bearer your-api-key" \
	 -d '{
	   "model": "claude-opus-4-1-20250805",
	   "messages": [{"role": "user", "content": "你好"}],
	   "stream": true
	 }'</div>
	           </div>
	           
	           <div id="javascript-tab" class="tabcontent">
	               <h3>JavaScript示例</h3>
	               <div class="example">
const fetch = require('node-fetch');

async function chatWithClaude(message, stream = false) {
	 const response = await fetch('http://localhost:9091/v1/chat/completions', {
	   method: 'POST',
	   headers: {
	     'Content-Type': 'application/json',
	     'Authorization': 'Bearer your-api-key'
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
chatWithClaude('你好，请介绍一下JavaScript', false);</div>
	           </div>
	       </section>
	       
	       <section id="error-handling">
	           <h2>错误处理</h2>
	           <p>API使用标准HTTP状态码来表示请求的成功或失败：</p>
	           <table>
	               <thead>
	                   <tr>
	                       <th>状态码</th>
	                       <th>说明</th>
	                   </tr>
	               </thead>
	               <tbody>
	                   <tr>
	                       <td>200 OK</td>
	                       <td>请求成功</td>
	                   </tr>
	                   <tr>
	                       <td>400 Bad Request</td>
	                       <td>请求格式错误或参数无效</td>
	                   </tr>
	                   <tr>
	                       <td>401 Unauthorized</td>
	                       <td>API密钥无效或缺失</td>
	                   </tr>
	                   <tr>
	                       <td>500 Internal Server Error</td>
	                       <td>服务器内部错误</td>
	                   </tr>
	               </tbody>
	           </table>
	           <div class="note">
	               <strong>注意:</strong> 在调试模式下，服务器会输出详细的日志信息，可以通过设置环境变量 DEBUG_MODE=true 来启用。
	           </div>
	       </section>
	   </div>

	   <script>
	       function openTab(evt, tabName) {
	           var i, tabcontent, tablinks;
	           tabcontent = document.getElementsByClassName("tabcontent");
	           for (i = 0; i < tabcontent.length; i++) {
	               tabcontent[i].style.display = "none";
	           }
	           tablinks = document.getElementsByClassName("tablinks");
	           for (i = 0; i < tablinks.length; i++) {
	               tablinks[i].className = tablinks[i].className.replace(" active", "");
	           }
	           document.getElementById(tabName).style.display = "block";
	           evt.currentTarget.className += " active";
	       }
	   </script>
</body>
</html>`
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, tmpl)
}

func main() {
	// 设置 Gin 模式
	if config.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// API 路由
	v1 := r.Group("/v1")
	v1.Use(authenticateClient)
	{
		v1.GET("/models", listModels)
		v1.POST("/chat/completions", chatCompletions)
	}

	// Dashboard 路由
	if config.DashboardEnabled {
		r.GET("/dashboard", handleDashboard)
		r.GET("/dashboard/stats", handleDashboardStats)
		r.GET("/dashboard/requests", handleDashboardRequests)
		log.Printf("Dashboard已启用，访问地址: http://localhost:%d/dashboard", config.Port)
	}

	// Docs 路由
	r.GET("/docs", handleDocs)

	// 打印配置信息
	log.Printf("服务器配置:")
	log.Printf("  端口: %d", config.Port)
	log.Printf("  默认流模式: %v", config.DefaultStream)
	log.Printf("  默认模型: %s", config.DefaultModel)
	log.Printf("  默认温度: %.1f", config.DefaultTemp)
	log.Printf("  超时时间: %d 秒", config.Timeout)
	log.Printf("  调试模式: %v", config.DebugMode)
	log.Printf("  Dashboard启用: %v", config.DashboardEnabled)
	if len(config.APIKeys) > 0 {
		log.Printf("  API 密钥: 已配置 %d 个", len(config.APIKeys))
	}

	// 启动服务器
	log.Printf("正在启动服务器，端口: %d", config.Port)
	log.Fatal(r.Run(fmt.Sprintf("0.0.0.0:%d", config.Port)))
}