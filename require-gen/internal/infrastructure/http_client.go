package infrastructure

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"specify-cli/internal/types"
)

// HTTPClientManager 管理HTTP客户端
type HTTPClientManager struct {
	mu            sync.RWMutex
	clients       map[string]*resty.Client
	defaultClient *resty.Client
	pool          *HTTPClientPool
	connPool      *ConnectionPoolManager
	errorHandler  *NetworkErrorHandler
	retryManager  *RetryManager
}

// NewHTTPClientManager 创建HTTP客户端管理器
func NewHTTPClientManager(defaultConfig *types.HTTPClientConfig) *HTTPClientManager {
	if defaultConfig == nil {
		defaultConfig = DefaultHTTPClientConfig()
	}
	
	errorHandler := NewNetworkErrorHandler(DefaultErrorHandlerConfig())
	
	manager := &HTTPClientManager{
		clients:  make(map[string]*resty.Client),
		pool:     NewHTTPClientPool(nil, 100),
		connPool: NewConnectionPoolManager(),
		errorHandler: errorHandler,
		retryManager: NewRetryManager(DefaultRetryManagerConfig(), errorHandler),
	}
	
	// 创建默认客户端
	manager.defaultClient = manager.createClient(defaultConfig)
	
	return manager
}

// DefaultHTTPClientConfig 默认HTTP客户端配置
func DefaultHTTPClientConfig() *types.HTTPClientConfig {
	return &types.HTTPClientConfig{
		Timeout:            30 * time.Second,
		RetryCount:         3,
		RetryWaitTime:      1 * time.Second,
		MaxRetryWaitTime:   30 * time.Second,
		FollowRedirects:    true,
		MaxRedirects:       10,
		UserAgent:          "specify-cli/1.0",
		Headers:            make(map[string]string),
		Cookies:            make(map[string]string),
		KeepAlive:          true,
		MaxIdleConns:       100,
		MaxConnsPerHost:    10,
		IdleConnTimeout:    90 * time.Second,
	}
}

// GetDefaultClient 获取默认客户端
func (hcm *HTTPClientManager) GetDefaultClient() *resty.Client {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()
	return hcm.defaultClient
}

// GetClient 获取命名客户端
func (hcm *HTTPClientManager) GetClient(name string) *resty.Client {
	hcm.mu.RLock()
	client, exists := hcm.clients[name]
	hcm.mu.RUnlock()
	
	if exists {
		return client
	}
	
	// 如果不存在，返回默认客户端
	return hcm.GetDefaultClient()
}

// CreateClient 创建带配置的客户端
func (hcm *HTTPClientManager) CreateClient(config *types.HTTPClientConfig) *resty.Client {
	return hcm.createClient(config)
}

// CreateClientWithConfig 使用配置创建客户端
func (hcm *HTTPClientManager) CreateClientWithConfig(config *types.HTTPClientConfig, networkConfig *types.NetworkConfig) *resty.Client {
	client := hcm.createClient(config)
	
	// 应用网络配置
	if networkConfig != nil {
		hcm.applyNetworkConfig(client, networkConfig)
	}
	
	return client
}

// createClient 创建resty客户端
func (hcm *HTTPClientManager) createClient(config *types.HTTPClientConfig) *resty.Client {
	client := resty.New()
	
	// 基本配置
	if config.Timeout > 0 {
		client.SetTimeout(config.Timeout)
	}
	
	// 重试配置
	if config.RetryCount > 0 {
		client.SetRetryCount(config.RetryCount)
		client.SetRetryWaitTime(config.RetryWaitTime)
		client.SetRetryMaxWaitTime(config.MaxRetryWaitTime)
		client.AddRetryCondition(hcm.createRetryCondition())
	}
	
	// 重定向配置
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(config.MaxRedirects))
	
	// User-Agent
	if config.UserAgent != "" {
		client.SetHeader("User-Agent", config.UserAgent)
	}
	
	// 自定义头部
	if len(config.Headers) > 0 {
		client.SetHeaders(config.Headers)
	}
	
	// Cookies
	if len(config.Cookies) > 0 {
		for name, value := range config.Cookies {
			client.SetCookies([]*http.Cookie{
				{
					Name:  name,
					Value: value,
				},
			})
		}
	}
	
	// 压缩
	client.SetDisableWarn(true)
	
	// 创建传输层
	transport := hcm.createTransport(config)
	client.SetTransport(transport)
	
	// 设置中间件
	hcm.setupMiddleware(client)
	
	return client
}

// createTransport 创建HTTP传输层
func (hcm *HTTPClientManager) createTransport(config *types.HTTPClientConfig) *http.Transport {
	// 从连接池管理器获取传输层
	if hcm.connPool != nil {
		// 获取连接池
		pool := hcm.connPool.GetPool("default", &ConnectionPoolConfig{
			MaxIdleConns:        config.MaxIdleConns,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			IdleConnTimeout:     config.IdleConnTimeout,
		})
		if pool != nil {
			return pool.GetTransport("default", &types.TLSConfig{})
		}
	}
	
	// 创建默认传输层
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   !config.KeepAlive,
		DisableCompression:  false,
		TLSHandshakeTimeout: 10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
	
	return transport
}

// buildTLSConfig 构建TLS配置
func (hcm *HTTPClientManager) buildTLSConfig(tlsConf *types.TLSConfig) *tls.Config {
	config := &tls.Config{
		InsecureSkipVerify: tlsConf.InsecureSkipVerify,
		ServerName:         tlsConf.ServerName,
	}
	
	// 加载客户端证书
	if tlsConf.CertFile != "" && tlsConf.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsConf.CertFile, tlsConf.KeyFile)
		if err == nil {
			config.Certificates = []tls.Certificate{cert}
		}
	}
	
	// 加载CA证书
	if tlsConf.CAFile != "" {
		// 这里可以添加CA证书加载逻辑
	}
	
	return config
}

// getTLSVersion 获取TLS版本
func (hcm *HTTPClientManager) getTLSVersion(version string) uint16 {
	switch strings.ToUpper(version) {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12 // 默认版本
	}
}

// getCipherSuites 获取密码套件
func (hcm *HTTPClientManager) getCipherSuites(suites []string) []uint16 {
	var cipherSuites []uint16
	
	suiteMap := map[string]uint16{
		"TLS_RSA_WITH_AES_128_CBC_SHA":                tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		"TLS_RSA_WITH_AES_256_CBC_SHA":                tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		"TLS_RSA_WITH_AES_128_GCM_SHA256":             tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_RSA_WITH_AES_256_GCM_SHA384":             tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":          tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":          tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":       tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":       tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":     tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":     tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"TLS_CHACHA20_POLY1305_SHA256":                tls.TLS_CHACHA20_POLY1305_SHA256,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256": tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	}
	
	for _, suite := range suites {
		if id, exists := suiteMap[suite]; exists {
			cipherSuites = append(cipherSuites, id)
		}
	}
	
	return cipherSuites
}

// applyNetworkConfig 应用网络配置
func (hcm *HTTPClientManager) applyNetworkConfig(client *resty.Client, config *types.NetworkConfig) {
	// 设置代理
	if config.ProxyURL != "" {
		client.SetProxy(config.ProxyURL)
	}
}

// createRetryCondition 创建重试条件
func (hcm *HTTPClientManager) createRetryCondition() resty.RetryConditionFunc {
	return func(r *resty.Response, err error) bool {
		// 网络错误重试
		if err != nil {
			return true
		}
		
		// HTTP状态码重试条件
		statusCode := r.StatusCode()
		
		// 5xx服务器错误重试
		if statusCode >= 500 && statusCode < 600 {
			return true
		}
		
		// 特定4xx错误重试
		retryableStatusCodes := []int{408, 429, 502, 503, 504}
		for _, code := range retryableStatusCodes {
			if statusCode == code {
				return true
			}
		}
		
		return false
	}
}

// setupMiddleware 设置中间件
func (hcm *HTTPClientManager) setupMiddleware(client *resty.Client) {
	// 请求中间件
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		// 添加请求ID
		req.SetHeader("X-Request-ID", generateRequestID())
		return nil
	})
	
	// 响应中间件
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		// 记录响应日志
		fmt.Printf("Response: %d %s\n", resp.StatusCode(), resp.Status())
		return nil
	})
	
	// 错误中间件
	client.OnError(func(req *resty.Request, err error) {
		// 记录错误日志
		fmt.Printf("Error: %s %s - %v\n", req.Method, req.URL, err)
	})
}

// HTTPClientPool HTTP客户端池
type HTTPClientPool struct {
	manager   *HTTPClientManager
	pool      map[string]*resty.Client
	maxSize   int
	cleanupInterval time.Duration
}

// NewHTTPClientPool 创建HTTP客户端池
func NewHTTPClientPool(manager *HTTPClientManager, maxSize int) *HTTPClientPool {
	pool := &HTTPClientPool{
		manager:         manager,
		pool:            make(map[string]*resty.Client),
		maxSize:         maxSize,
		cleanupInterval: 5 * time.Minute,
	}
	
	// 启动清理协程
	go pool.startCleanup()
	
	return pool
}

// GetClient 从池中获取客户端
func (hcp *HTTPClientPool) GetClient(key string, config *types.HTTPClientConfig) *resty.Client {
	if client, exists := hcp.pool[key]; exists {
		return client
	}
	
	// 检查池大小限制
	if len(hcp.pool) >= hcp.maxSize {
		hcp.evictOldest()
	}
	
	client := hcp.manager.CreateClientWithConfig(config, nil)
	hcp.pool[key] = client
	
	return client
}

// evictOldest 驱逐最旧的客户端
func (hcp *HTTPClientPool) evictOldest() {
	// 简化实现，删除第一个找到的客户端
	for key := range hcp.pool {
		delete(hcp.pool, key)
		break
	}
}

// startCleanup 启动清理协程
func (hcp *HTTPClientPool) startCleanup() {
	ticker := time.NewTicker(hcp.cleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		// 执行清理逻辑
		hcp.cleanup()
	}
}

// cleanup 清理过期的客户端
func (hcp *HTTPClientPool) cleanup() {
	// 这里可以实现更复杂的清理逻辑
	// 例如基于最后使用时间、连接状态等
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// HTTPClientBuilder HTTP客户端构建器
type HTTPClientBuilder struct {
	config        *types.HTTPClientConfig
	networkConfig *types.NetworkConfig
	manager       *HTTPClientManager
}

// NewHTTPClientBuilder 创建HTTP客户端构建器
func NewHTTPClientBuilder() *HTTPClientBuilder {
	return &HTTPClientBuilder{
		config:        DefaultHTTPClientConfig(),
		networkConfig: &types.NetworkConfig{},
		manager:       NewHTTPClientManager(nil),
	}
}

// WithTimeout 设置超时时间
func (hcb *HTTPClientBuilder) WithTimeout(timeout time.Duration) *HTTPClientBuilder {
	hcb.config.Timeout = timeout
	return hcb
}

// WithRetries 设置重试配置
func (hcb *HTTPClientBuilder) WithRetries(maxRetries int, waitTime, maxWaitTime time.Duration) *HTTPClientBuilder {
	hcb.config.RetryCount = maxRetries
	hcb.config.RetryWaitTime = waitTime
	hcb.config.MaxRetryWaitTime = maxWaitTime
	return hcb
}

// WithProxy 设置代理
func (hcb *HTTPClientBuilder) WithProxy(proxyURL string) *HTTPClientBuilder {
	hcb.networkConfig.ProxyURL = proxyURL
	return hcb
}

// WithTLS 设置TLS配置
func (hcb *HTTPClientBuilder) WithTLS(tlsConfig *types.TLSConfig) *HTTPClientBuilder {
	// TLS配置通过NetworkConfig设置
	if hcb.networkConfig.TLS == nil {
		hcb.networkConfig.TLS = tlsConfig
	}
	return hcb
}

// WithHeaders 设置头部
func (hcb *HTTPClientBuilder) WithHeaders(headers map[string]string) *HTTPClientBuilder {
	if hcb.config.Headers == nil {
		hcb.config.Headers = make(map[string]string)
	}
	for k, v := range headers {
		hcb.config.Headers[k] = v
	}
	return hcb
}

// WithUserAgent 设置User-Agent
func (hcb *HTTPClientBuilder) WithUserAgent(userAgent string) *HTTPClientBuilder {
	hcb.config.UserAgent = userAgent
	return hcb
}

// Build 构建HTTP客户端
func (hcb *HTTPClientBuilder) Build() *resty.Client {
	return hcb.manager.CreateClientWithConfig(hcb.config, hcb.networkConfig)
}

// BuildWithContext 使用上下文构建HTTP客户端
func (hcb *HTTPClientBuilder) BuildWithContext(ctx context.Context) *resty.Client {
	client := hcb.Build()
	
	// 为所有请求设置上下文
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.SetContext(ctx)
		return nil
	})
	
	return client
}