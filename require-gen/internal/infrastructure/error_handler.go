package infrastructure

import (
	"context"
	"math"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"specify-cli/internal/types"
)

// NetworkErrorHandler 网络错误处理器
type NetworkErrorHandler struct {
	mu              sync.RWMutex
	config          *ErrorHandlerConfig
	retryStrategies map[types.NetworkErrorType]RetryStrategy
	errorStats      *ErrorStatistics
	circuitBreaker  *CircuitBreaker
}

// ErrorHandlerConfig 错误处理器配置
type ErrorHandlerConfig struct {
	// 重试配置
	MaxRetries      int           // 最大重试次数
	BaseRetryDelay  time.Duration // 基础重试延迟
	MaxRetryDelay   time.Duration // 最大重试延迟
	RetryMultiplier float64       // 重试延迟倍数

	// 超时配置
	DefaultTimeout    time.Duration // 默认超时时间
	ConnectionTimeout time.Duration // 连接超时时间
	ReadTimeout       time.Duration // 读取超时时间
	WriteTimeout      time.Duration // 写入超时时间

	// 熔断器配置
	EnableCircuitBreaker bool          // 启用熔断器
	FailureThreshold     int           // 失败阈值
	RecoveryTimeout      time.Duration // 恢复超时时间

	// 统计配置
	EnableStats    bool          // 启用统计
	StatsRetention time.Duration // 统计保留时间

	// 错误分类配置
	RetryableErrors    []types.NetworkErrorType // 可重试的错误类型
	NonRetryableErrors []types.NetworkErrorType // 不可重试的错误类型
}

// RetryStrategy 重试策略
type RetryStrategy struct {
	MaxRetries  int           // 最大重试次数
	BaseDelay   time.Duration // 基础延迟
	MaxDelay    time.Duration // 最大延迟
	Multiplier  float64       // 延迟倍数
	Jitter      bool          // 是否添加抖动
	BackoffType BackoffType   // 退避类型
}

// BackoffType 退避类型
type BackoffType int

const (
	LinearBackoff BackoffType = iota
	ExponentialBackoff
	FixedBackoff
)

// ErrorStatistics 错误统计
type ErrorStatistics struct {
	mu                sync.RWMutex
	TotalErrors       int64                            // 总错误数
	ErrorsByType      map[types.NetworkErrorType]int64 // 按类型分组的错误数
	ErrorsByHost      map[string]int64                 // 按主机分组的错误数
	RetryAttempts     int64                            // 重试次数
	SuccessfulRetries int64                            // 成功重试次数
	FailedRetries     int64                            // 失败重试次数
	AverageRetryDelay time.Duration                    // 平均重试延迟
	LastError         *types.NetworkError              // 最后一个错误
	ErrorHistory      []*ErrorRecord                   // 错误历史
	StartTime         time.Time                        // 开始时间
}

// ErrorRecord 错误记录
type ErrorRecord struct {
	Timestamp  time.Time           // 时间戳
	Error      *types.NetworkError // 错误信息
	Host       string              // 主机
	RetryCount int                 // 重试次数
	Duration   time.Duration       // 持续时间
	Recovered  bool                // 是否恢复
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitState
	failureCount    int
	lastFailureTime time.Time
	config          *CircuitBreakerConfig
}

// CircuitState 熔断器状态
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           // 失败阈值
	RecoveryTimeout  time.Duration // 恢复超时时间
	MaxRequests      int           // 半开状态最大请求数
}

// DefaultErrorHandlerConfig 默认错误处理器配置
func DefaultErrorHandlerConfig() *ErrorHandlerConfig {
	return &ErrorHandlerConfig{
		MaxRetries:           3,
		BaseRetryDelay:       1 * time.Second,
		MaxRetryDelay:        30 * time.Second,
		RetryMultiplier:      2.0,
		DefaultTimeout:       30 * time.Second,
		ConnectionTimeout:    10 * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		EnableCircuitBreaker: true,
		FailureThreshold:     5,
		RecoveryTimeout:      60 * time.Second,
		EnableStats:          true,
		StatsRetention:       24 * time.Hour,
		RetryableErrors: []types.NetworkErrorType{
			types.NetworkErrorTimeout,
			types.NetworkErrorConnectionRefused,
			types.NetworkErrorDNSResolution,
			types.NetworkErrorTemporary,
		},
		NonRetryableErrors: []types.NetworkErrorType{
			types.NetworkErrorCertificate,
			types.NetworkErrorAuthentication,
			types.NetworkErrorPermission,
		},
	}
}

// NewNetworkErrorHandler 创建网络错误处理器
func NewNetworkErrorHandler(config *ErrorHandlerConfig) *NetworkErrorHandler {
	if config == nil {
		config = DefaultErrorHandlerConfig()
	}

	handler := &NetworkErrorHandler{
		config:          config,
		retryStrategies: make(map[types.NetworkErrorType]RetryStrategy),
		errorStats: &ErrorStatistics{
			ErrorsByType: make(map[types.NetworkErrorType]int64),
			ErrorsByHost: make(map[string]int64),
			StartTime:    time.Now(),
		},
	}

	// 初始化重试策略
	handler.initRetryStrategies()

	// 初始化熔断器
	if config.EnableCircuitBreaker {
		handler.circuitBreaker = &CircuitBreaker{
			state: CircuitClosed,
			config: &CircuitBreakerConfig{
				FailureThreshold: config.FailureThreshold,
				RecoveryTimeout:  config.RecoveryTimeout,
				MaxRequests:      10,
			},
		}
	}

	return handler
}

// initRetryStrategies 初始化重试策略
func (neh *NetworkErrorHandler) initRetryStrategies() {
	// 为不同错误类型设置不同的重试策略
	strategies := map[types.NetworkErrorType]RetryStrategy{
		types.NetworkErrorTimeout: {
			MaxRetries:  neh.config.MaxRetries,
			BaseDelay:   neh.config.BaseRetryDelay,
			MaxDelay:    neh.config.MaxRetryDelay,
			Multiplier:  neh.config.RetryMultiplier,
			Jitter:      true,
			BackoffType: ExponentialBackoff,
		},
		types.NetworkErrorConnectionRefused: {
			MaxRetries:  neh.config.MaxRetries - 1,
			BaseDelay:   neh.config.BaseRetryDelay * 2,
			MaxDelay:    neh.config.MaxRetryDelay,
			Multiplier:  neh.config.RetryMultiplier,
			Jitter:      true,
			BackoffType: ExponentialBackoff,
		},
		types.NetworkErrorDNSResolution: {
			MaxRetries:  2,
			BaseDelay:   neh.config.BaseRetryDelay,
			MaxDelay:    neh.config.MaxRetryDelay / 2,
			Multiplier:  1.5,
			Jitter:      false,
			BackoffType: LinearBackoff,
		},
		types.NetworkErrorTemporary: {
			MaxRetries:  neh.config.MaxRetries + 1,
			BaseDelay:   neh.config.BaseRetryDelay / 2,
			MaxDelay:    neh.config.MaxRetryDelay,
			Multiplier:  neh.config.RetryMultiplier,
			Jitter:      true,
			BackoffType: ExponentialBackoff,
		},
	}

	neh.retryStrategies = strategies
}

// HandleError 处理网络错误
func (neh *NetworkErrorHandler) HandleError(ctx context.Context, err error, host string) *types.NetworkError {
	// 分类错误
	networkErr := neh.ClassifyError(err, host)

	// 记录错误统计
	neh.recordError(networkErr, host)

	// 检查熔断器状态
	if neh.circuitBreaker != nil && !neh.circuitBreaker.AllowRequest() {
		networkErr.Type = types.NetworkErrorCircuitOpen
		networkErr.Message = "Circuit breaker is open"
		return networkErr
	}

	return networkErr
}

// ClassifyError 分类错误
func (neh *NetworkErrorHandler) ClassifyError(err error, host string) *types.NetworkError {
	if err == nil {
		return nil
	}

	networkErr := &types.NetworkError{
		Type:      types.NetworkErrorUnknown,
		Message:   err.Error(),
		Host:      host,
		Timestamp: time.Now(),
		Retryable: false,
	}

	// 根据错误类型进行分类
	switch e := err.(type) {
	case *net.OpError:
		networkErr = neh.classifyOpError(e, host)
	case *url.Error:
		networkErr = neh.classifyURLError(e, host)
	case net.Error:
		if e.Timeout() {
			networkErr.Type = types.NetworkErrorTimeout
			networkErr.Retryable = true
		} else if e.Temporary() {
			networkErr.Type = types.NetworkErrorTemporary
			networkErr.Retryable = true
		}
	default:
		// 检查HTTP状态码错误
		if httpErr := neh.classifyHTTPError(err); httpErr != nil {
			networkErr = httpErr
		}
	}

	// 设置重试策略
	if strategy, exists := neh.retryStrategies[networkErr.Type]; exists {
		networkErr.Retryable = true
		networkErr.RetryStrategy = &strategy
	}

	return networkErr
}

// classifyOpError 分类操作错误
func (neh *NetworkErrorHandler) classifyOpError(err *net.OpError, host string) *types.NetworkError {
	networkErr := &types.NetworkError{
		Type:      types.NetworkErrorUnknown,
		Message:   err.Error(),
		Host:      host,
		Timestamp: time.Now(),
	}

	if err.Timeout() {
		networkErr.Type = types.NetworkErrorTimeout
		networkErr.Retryable = true
	} else if err.Temporary() {
		networkErr.Type = types.NetworkErrorTemporary
		networkErr.Retryable = true
	} else if strings.Contains(err.Error(), "connection refused") {
		networkErr.Type = types.NetworkErrorConnectionRefused
		networkErr.Retryable = true
	} else if strings.Contains(err.Error(), "no such host") {
		networkErr.Type = types.NetworkErrorDNSResolution
		networkErr.Retryable = true
	}

	return networkErr
}

// classifyURLError 分类URL错误
func (neh *NetworkErrorHandler) classifyURLError(err *url.Error, host string) *types.NetworkError {
	networkErr := &types.NetworkError{
		Type:      types.NetworkErrorUnknown,
		Message:   err.Error(),
		Host:      host,
		Timestamp: time.Now(),
	}

	if err.Timeout() {
		networkErr.Type = types.NetworkErrorTimeout
		networkErr.Retryable = true
	} else if err.Temporary() {
		networkErr.Type = types.NetworkErrorTemporary
		networkErr.Retryable = true
	} else if strings.Contains(err.Error(), "certificate") {
		networkErr.Type = types.NetworkErrorCertificate
		networkErr.Retryable = false
	}

	return networkErr
}

// classifyHTTPError 分类HTTP错误
func (neh *NetworkErrorHandler) classifyHTTPError(err error) *types.NetworkError {
	errStr := err.Error()

	// 检查是否包含HTTP状态码
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		return &types.NetworkError{
			Type:      types.NetworkErrorHTTPServer,
			Message:   errStr,
			Timestamp: time.Now(),
			Retryable: true,
		}
	}

	if strings.Contains(errStr, "401") {
		return &types.NetworkError{
			Type:      types.NetworkErrorAuthentication,
			Message:   errStr,
			Timestamp: time.Now(),
			Retryable: false,
		}
	}

	if strings.Contains(errStr, "403") {
		return &types.NetworkError{
			Type:      types.NetworkErrorPermission,
			Message:   errStr,
			Timestamp: time.Now(),
			Retryable: false,
		}
	}

	return nil
}

// ShouldRetry 判断是否应该重试
func (neh *NetworkErrorHandler) ShouldRetry(networkErr *types.NetworkError, retryCount int) bool {
	if networkErr == nil || !networkErr.Retryable {
		return false
	}

	// 检查重试次数
	strategy, exists := neh.retryStrategies[networkErr.Type]
	if !exists {
		return retryCount < neh.config.MaxRetries
	}

	return retryCount < strategy.MaxRetries
}

// CalculateRetryDelay 计算重试延迟
func (neh *NetworkErrorHandler) CalculateRetryDelay(networkErr *types.NetworkError, retryCount int) time.Duration {
	strategy, exists := neh.retryStrategies[networkErr.Type]
	if !exists {
		return neh.config.BaseRetryDelay
	}

	var delay time.Duration

	switch strategy.BackoffType {
	case LinearBackoff:
		delay = strategy.BaseDelay * time.Duration(retryCount+1)
	case ExponentialBackoff:
		delay = time.Duration(float64(strategy.BaseDelay) *
			math.Pow(2, float64(retryCount)) * strategy.Multiplier)
	case FixedBackoff:
		delay = strategy.BaseDelay
	default:
		delay = strategy.BaseDelay
	}

	// 限制最大延迟
	if delay > strategy.MaxDelay {
		delay = strategy.MaxDelay
	}

	// 添加抖动
	if strategy.Jitter {
		jitter := time.Duration(float64(delay) * 0.1 * (float64(time.Now().UnixNano()%2)*2 - 1))
		delay += jitter
	}

	return delay
}

// recordError 记录错误统计
func (neh *NetworkErrorHandler) recordError(networkErr *types.NetworkError, host string) {
	if !neh.config.EnableStats || networkErr == nil {
		return
	}

	neh.errorStats.mu.Lock()
	defer neh.errorStats.mu.Unlock()

	neh.errorStats.TotalErrors++
	neh.errorStats.ErrorsByType[networkErr.Type]++
	neh.errorStats.ErrorsByHost[host]++
	neh.errorStats.LastError = networkErr

	// 记录错误历史
	record := &ErrorRecord{
		Timestamp:  time.Now(),
		Error:      networkErr,
		Host:       host,
		RetryCount: 0,
	}

	neh.errorStats.ErrorHistory = append(neh.errorStats.ErrorHistory, record)

	// 限制历史记录数量
	if len(neh.errorStats.ErrorHistory) > 1000 {
		neh.errorStats.ErrorHistory = neh.errorStats.ErrorHistory[100:]
	}

	// 更新熔断器
	if neh.circuitBreaker != nil {
		neh.circuitBreaker.RecordFailure()
	}
}

// GetErrorStats 获取错误统计
func (neh *NetworkErrorHandler) GetErrorStats() ErrorStatistics {
	neh.errorStats.mu.RLock()
	defer neh.errorStats.mu.RUnlock()

	return *neh.errorStats
}

// Reset 重置错误处理器
func (neh *NetworkErrorHandler) Reset() {
	neh.errorStats.mu.Lock()
	defer neh.errorStats.mu.Unlock()

	neh.errorStats.TotalErrors = 0
	neh.errorStats.ErrorsByType = make(map[types.NetworkErrorType]int64)
	neh.errorStats.ErrorsByHost = make(map[string]int64)
	neh.errorStats.RetryAttempts = 0
	neh.errorStats.SuccessfulRetries = 0
	neh.errorStats.FailedRetries = 0
	neh.errorStats.ErrorHistory = nil
	neh.errorStats.StartTime = time.Now()

	if neh.circuitBreaker != nil {
		neh.circuitBreaker.Reset()
	}
}

// CircuitBreaker methods

// AllowRequest 检查是否允许请求
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.config.FailureThreshold {
		cb.state = CircuitOpen
	}
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.lastFailureTime = time.Time{}
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}
