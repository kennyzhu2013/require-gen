package infrastructure

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"specify-cli/internal/types"
)

// RetryManager 重试管理器
type RetryManager struct {
	mu            sync.RWMutex
	config        *RetryManagerConfig
	errorHandler  *NetworkErrorHandler
	strategies    map[string]*RetryStrategy
	activeRetries map[string]*RetryContext
	stats         *RetryStatistics
}

// RetryManagerConfig 重试管理器配置
type RetryManagerConfig struct {
	// 全局重试配置
	DefaultMaxRetries    int           // 默认最大重试次数
	DefaultBaseDelay     time.Duration // 默认基础延迟
	DefaultMaxDelay      time.Duration // 默认最大延迟
	DefaultMultiplier    float64       // 默认延迟倍数
	
	// 并发控制
	MaxConcurrentRetries int           // 最大并发重试数
	RetryQueueSize       int           // 重试队列大小
	
	// 监控配置
	EnableMetrics        bool          // 启用指标
	MetricsInterval      time.Duration // 指标收集间隔
	
	// 自适应配置
	EnableAdaptive       bool          // 启用自适应重试
	AdaptiveWindow       time.Duration // 自适应窗口
	SuccessRateThreshold float64       // 成功率阈值
}

// RetryContext 重试上下文
type RetryContext struct {
	ID            string                    // 重试ID
	StartTime     time.Time                 // 开始时间
	LastAttempt   time.Time                 // 最后尝试时间
	AttemptCount  int                       // 尝试次数
	MaxRetries    int                       // 最大重试次数
	Strategy      *RetryStrategy            // 重试策略
	Error         *types.NetworkError       // 最后错误
	Host          string                    // 目标主机
	Operation     string                    // 操作类型
	Metadata      map[string]interface{}    // 元数据
	OnSuccess     func()                    // 成功回调
	OnFailure     func(error)               // 失败回调
	OnRetry       func(int, time.Duration)  // 重试回调
}

// RetryStatistics 重试统计
type RetryStatistics struct {
	mu                    sync.RWMutex
	TotalRetries          int64                                 // 总重试次数
	SuccessfulRetries     int64                                 // 成功重试次数
	FailedRetries         int64                                 // 失败重试次数
	RetriesByType         map[types.NetworkErrorType]int64     // 按错误类型分组的重试次数
	RetriesByHost         map[string]int64                     // 按主机分组的重试次数
	RetriesByOperation    map[string]int64                     // 按操作分组的重试次数
	AverageRetryDelay     time.Duration                        // 平均重试延迟
	AverageRetryDuration  time.Duration                        // 平均重试持续时间
	SuccessRate           float64                              // 成功率
	LastSuccessTime       time.Time                            // 最后成功时间
	LastFailureTime       time.Time                            // 最后失败时间
	ActiveRetries         int                                  // 活跃重试数
	QueuedRetries         int                                  // 排队重试数
}

// RetryResult 重试结果
type RetryResult struct {
	Success       bool                    // 是否成功
	AttemptCount  int                     // 尝试次数
	TotalDuration time.Duration           // 总持续时间
	LastError     *types.NetworkError     // 最后错误
	Metadata      map[string]interface{}  // 结果元数据
}

// RetryOperation 重试操作函数类型
type RetryOperation func(ctx context.Context, attempt int) error

// DefaultRetryManagerConfig 默认重试管理器配置
func DefaultRetryManagerConfig() *RetryManagerConfig {
	return &RetryManagerConfig{
		DefaultMaxRetries:    3,
		DefaultBaseDelay:     1 * time.Second,
		DefaultMaxDelay:      30 * time.Second,
		DefaultMultiplier:    2.0,
		MaxConcurrentRetries: 10,
		RetryQueueSize:       100,
		EnableMetrics:        true,
		MetricsInterval:      1 * time.Minute,
		EnableAdaptive:       true,
		AdaptiveWindow:       10 * time.Minute,
		SuccessRateThreshold: 0.8,
	}
}

// NewRetryManager 创建重试管理器
func NewRetryManager(config *RetryManagerConfig, errorHandler *NetworkErrorHandler) *RetryManager {
	if config == nil {
		config = DefaultRetryManagerConfig()
	}
	
	if errorHandler == nil {
		errorHandler = NewNetworkErrorHandler(nil)
	}
	
	manager := &RetryManager{
		config:        config,
		errorHandler:  errorHandler,
		strategies:    make(map[string]*RetryStrategy),
		activeRetries: make(map[string]*RetryContext),
		stats: &RetryStatistics{
			RetriesByType:      make(map[types.NetworkErrorType]int64),
			RetriesByHost:      make(map[string]int64),
			RetriesByOperation: make(map[string]int64),
		},
	}
	
	// 初始化默认策略
	manager.initDefaultStrategies()
	
	// 启动指标收集
	if config.EnableMetrics {
		manager.startMetricsCollection()
	}
	
	return manager
}

// initDefaultStrategies 初始化默认策略
func (rm *RetryManager) initDefaultStrategies() {
	// 默认策略
	rm.strategies["default"] = &RetryStrategy{
		MaxRetries:  rm.config.DefaultMaxRetries,
		BaseDelay:   rm.config.DefaultBaseDelay,
		MaxDelay:    rm.config.DefaultMaxDelay,
		Multiplier:  rm.config.DefaultMultiplier,
		Jitter:      true,
		BackoffType: ExponentialBackoff,
	}
	
	// 快速重试策略（用于临时错误）
	rm.strategies["fast"] = &RetryStrategy{
		MaxRetries:  5,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  1.5,
		Jitter:      true,
		BackoffType: ExponentialBackoff,
	}
	
	// 慢速重试策略（用于严重错误）
	rm.strategies["slow"] = &RetryStrategy{
		MaxRetries:  2,
		BaseDelay:   5 * time.Second,
		MaxDelay:    60 * time.Second,
		Multiplier:  3.0,
		Jitter:      false,
		BackoffType: ExponentialBackoff,
	}
	
	// 线性重试策略
	rm.strategies["linear"] = &RetryStrategy{
		MaxRetries:  4,
		BaseDelay:   2 * time.Second,
		MaxDelay:    20 * time.Second,
		Multiplier:  1.0,
		Jitter:      true,
		BackoffType: LinearBackoff,
	}
}

// ExecuteWithRetry 执行带重试的操作
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, operation RetryOperation, options *RetryOptions) *RetryResult {
	if options == nil {
		options = &RetryOptions{
			StrategyName: "default",
		}
	}
	
	// 创建重试上下文
	retryCtx := rm.createRetryContext(options)
	
	// 注册活跃重试
	rm.registerActiveRetry(retryCtx)
	defer rm.unregisterActiveRetry(retryCtx.ID)
	
	result := &RetryResult{
		Metadata: make(map[string]interface{}),
	}
	
	startTime := time.Now()
	
	for attempt := 0; attempt <= retryCtx.MaxRetries; attempt++ {
		retryCtx.AttemptCount = attempt + 1
		retryCtx.LastAttempt = time.Now()
		
		// 执行操作
		err := operation(ctx, attempt)
		
		if err == nil {
			// 成功
			result.Success = true
			result.AttemptCount = attempt + 1
			result.TotalDuration = time.Since(startTime)
			
			// 记录成功统计
			rm.recordSuccess(retryCtx)
			
			// 调用成功回调
			if retryCtx.OnSuccess != nil {
				retryCtx.OnSuccess()
			}
			
			return result
		}
		
		// 处理错误
		networkErr := rm.errorHandler.HandleError(ctx, err, retryCtx.Host)
		retryCtx.Error = networkErr
		
		// 检查是否应该重试
		if attempt >= retryCtx.MaxRetries || !rm.shouldRetry(networkErr, attempt) {
			// 不再重试
			result.Success = false
			result.AttemptCount = attempt + 1
			result.TotalDuration = time.Since(startTime)
			result.LastError = networkErr
			
			// 记录失败统计
			rm.recordFailure(retryCtx)
			
			// 调用失败回调
			if retryCtx.OnFailure != nil {
				retryCtx.OnFailure(err)
			}
			
			return result
		}
		
		// 计算重试延迟
		delay := rm.calculateRetryDelay(networkErr, attempt, retryCtx.Strategy)
		
		// 调用重试回调
		if retryCtx.OnRetry != nil {
			retryCtx.OnRetry(attempt+1, delay)
		}
		
		// 记录重试统计
		rm.recordRetryAttempt(retryCtx)
		
		// 等待重试延迟
		select {
		case <-ctx.Done():
			result.Success = false
			result.AttemptCount = attempt + 1
			result.TotalDuration = time.Since(startTime)
			result.LastError = &types.NetworkError{
				Type:      types.NetworkErrorTimeout,
				Message:   "Context cancelled during retry",
				Timestamp: time.Now(),
			}
			return result
		case <-time.After(delay):
			// 继续重试
		}
	}
	
	// 不应该到达这里
	result.Success = false
	result.TotalDuration = time.Since(startTime)
	return result
}

// RetryOptions 重试选项
type RetryOptions struct {
	StrategyName  string                    // 策略名称
	MaxRetries    *int                      // 最大重试次数（覆盖策略）
	Host          string                    // 目标主机
	Operation     string                    // 操作类型
	Metadata      map[string]interface{}    // 元数据
	OnSuccess     func()                    // 成功回调
	OnFailure     func(error)               // 失败回调
	OnRetry       func(int, time.Duration)  // 重试回调
}

// createRetryContext 创建重试上下文
func (rm *RetryManager) createRetryContext(options *RetryOptions) *RetryContext {
	strategy := rm.getStrategy(options.StrategyName)
	
	maxRetries := strategy.MaxRetries
	if options.MaxRetries != nil {
		maxRetries = *options.MaxRetries
	}
	
	return &RetryContext{
		ID:          fmt.Sprintf("retry_%d", time.Now().UnixNano()),
		StartTime:   time.Now(),
		MaxRetries:  maxRetries,
		Strategy:    strategy,
		Host:        options.Host,
		Operation:   options.Operation,
		Metadata:    options.Metadata,
		OnSuccess:   options.OnSuccess,
		OnFailure:   options.OnFailure,
		OnRetry:     options.OnRetry,
	}
}

// getStrategy 获取重试策略
func (rm *RetryManager) getStrategy(name string) *RetryStrategy {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	if strategy, exists := rm.strategies[name]; exists {
		return strategy
	}
	
	// 返回默认策略
	return rm.strategies["default"]
}

// shouldRetry 判断是否应该重试
func (rm *RetryManager) shouldRetry(networkErr *types.NetworkError, attempt int) bool {
	if networkErr == nil {
		return false
	}
	
	// 使用错误处理器判断
	return rm.errorHandler.ShouldRetry(networkErr, attempt)
}

// calculateRetryDelay 计算重试延迟
func (rm *RetryManager) calculateRetryDelay(networkErr *types.NetworkError, attempt int, strategy *RetryStrategy) time.Duration {
	if networkErr != nil {
		// 使用错误处理器计算延迟
		return rm.errorHandler.CalculateRetryDelay(networkErr, attempt)
	}
	
	// 使用策略计算延迟
	var delay time.Duration
	
	switch strategy.BackoffType {
	case LinearBackoff:
		delay = strategy.BaseDelay * time.Duration(attempt+1)
	case ExponentialBackoff:
		delay = time.Duration(float64(strategy.BaseDelay) * 
			math.Pow(2, float64(attempt)) * strategy.Multiplier)
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

// registerActiveRetry 注册活跃重试
func (rm *RetryManager) registerActiveRetry(retryCtx *RetryContext) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.activeRetries[retryCtx.ID] = retryCtx
	rm.stats.ActiveRetries++
}

// unregisterActiveRetry 注销活跃重试
func (rm *RetryManager) unregisterActiveRetry(id string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if _, exists := rm.activeRetries[id]; exists {
		delete(rm.activeRetries, id)
		rm.stats.ActiveRetries--
	}
}

// recordSuccess 记录成功
func (rm *RetryManager) recordSuccess(retryCtx *RetryContext) {
	rm.stats.mu.Lock()
	defer rm.stats.mu.Unlock()
	
	rm.stats.SuccessfulRetries++
	rm.stats.LastSuccessTime = time.Now()
	
	if retryCtx.Host != "" {
		rm.stats.RetriesByHost[retryCtx.Host]++
	}
	
	if retryCtx.Operation != "" {
		rm.stats.RetriesByOperation[retryCtx.Operation]++
	}
	
	// 更新成功率
	rm.updateSuccessRate()
}

// recordFailure 记录失败
func (rm *RetryManager) recordFailure(retryCtx *RetryContext) {
	rm.stats.mu.Lock()
	defer rm.stats.mu.Unlock()
	
	rm.stats.FailedRetries++
	rm.stats.LastFailureTime = time.Now()
	
	if retryCtx.Error != nil {
		rm.stats.RetriesByType[retryCtx.Error.Type]++
	}
	
	if retryCtx.Host != "" {
		rm.stats.RetriesByHost[retryCtx.Host]++
	}
	
	if retryCtx.Operation != "" {
		rm.stats.RetriesByOperation[retryCtx.Operation]++
	}
	
	// 更新成功率
	rm.updateSuccessRate()
}

// recordRetryAttempt 记录重试尝试
func (rm *RetryManager) recordRetryAttempt(retryCtx *RetryContext) {
	rm.stats.mu.Lock()
	defer rm.stats.mu.Unlock()
	
	rm.stats.TotalRetries++
}

// updateSuccessRate 更新成功率
func (rm *RetryManager) updateSuccessRate() {
	total := rm.stats.SuccessfulRetries + rm.stats.FailedRetries
	if total > 0 {
		rm.stats.SuccessRate = float64(rm.stats.SuccessfulRetries) / float64(total)
	}
}

// startMetricsCollection 启动指标收集
func (rm *RetryManager) startMetricsCollection() {
	go func() {
		ticker := time.NewTicker(rm.config.MetricsInterval)
		defer ticker.Stop()
		
		for range ticker.C {
			rm.collectMetrics()
		}
	}()
}

// collectMetrics 收集指标
func (rm *RetryManager) collectMetrics() {
	rm.mu.RLock()
	activeCount := len(rm.activeRetries)
	rm.mu.RUnlock()
	
	rm.stats.mu.Lock()
	rm.stats.ActiveRetries = activeCount
	rm.stats.mu.Unlock()
	
	// 这里可以添加更多指标收集逻辑
	// 例如发送到监控系统
}

// GetRetryStatistics 获取重试统计
func (rm *RetryManager) GetRetryStatistics() RetryStatistics {
	rm.stats.mu.RLock()
	defer rm.stats.mu.RUnlock()
	
	return *rm.stats
}

// AddStrategy 添加重试策略
func (rm *RetryManager) AddStrategy(name string, strategy *RetryStrategy) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.strategies[name] = strategy
}

// RemoveStrategy 移除重试策略
func (rm *RetryManager) RemoveStrategy(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// 不能删除默认策略
	if name != "default" {
		delete(rm.strategies, name)
	}
}

// GetActiveRetries 获取活跃重试列表
func (rm *RetryManager) GetActiveRetries() map[string]*RetryContext {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	result := make(map[string]*RetryContext)
	for id, ctx := range rm.activeRetries {
		result[id] = ctx
	}
	
	return result
}

// CancelRetry 取消重试
func (rm *RetryManager) CancelRetry(id string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if _, exists := rm.activeRetries[id]; exists {
		delete(rm.activeRetries, id)
		rm.stats.ActiveRetries--
		return true
	}
	
	return false
}

// Reset 重置重试管理器
func (rm *RetryManager) Reset() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	// 清空活跃重试
	rm.activeRetries = make(map[string]*RetryContext)
	
	// 重置统计
	rm.stats.mu.Lock()
	rm.stats.TotalRetries = 0
	rm.stats.SuccessfulRetries = 0
	rm.stats.FailedRetries = 0
	rm.stats.RetriesByType = make(map[types.NetworkErrorType]int64)
	rm.stats.RetriesByHost = make(map[string]int64)
	rm.stats.RetriesByOperation = make(map[string]int64)
	rm.stats.ActiveRetries = 0
	rm.stats.QueuedRetries = 0
	rm.stats.SuccessRate = 0
	rm.stats.mu.Unlock()
}