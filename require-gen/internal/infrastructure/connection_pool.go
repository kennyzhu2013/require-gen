package infrastructure

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"specify-cli/internal/types"
)

// ConnectionPool HTTP连接池
type ConnectionPool struct {
	mu            sync.RWMutex
	pools         map[string]*http.Transport
	config        *ConnectionPoolConfig
	cleanupTicker *time.Ticker
	cleanupDone   chan struct{}
	stats         *ConnectionStats
}

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	// 连接池配置
	MaxIdleConns        int           // 最大空闲连接数
	MaxIdleConnsPerHost int           // 每个主机最大空闲连接数
	MaxConnsPerHost     int           // 每个主机最大连接数
	IdleConnTimeout     time.Duration // 空闲连接超时时间

	// 连接配置
	DialTimeout         time.Duration // 连接超时时间
	KeepAlive           time.Duration // Keep-Alive时间
	TLSHandshakeTimeout time.Duration // TLS握手超时时间

	// 清理配置
	CleanupInterval time.Duration // 清理间隔
	MaxLifetime     time.Duration // 连接最大生存时间

	// 监控配置
	EnableStats   bool          // 启用统计
	StatsInterval time.Duration // 统计间隔
}

// ConnectionStats 连接统计信息
type ConnectionStats struct {
	mu                 sync.RWMutex
	TotalConnections   int64         // 总连接数
	ActiveConnections  int64         // 活跃连接数
	IdleConnections    int64         // 空闲连接数
	ConnectionsCreated int64         // 创建的连接数
	ConnectionsReused  int64         // 复用的连接数
	ConnectionsClosed  int64         // 关闭的连接数
	ConnectionErrors   int64         // 连接错误数
	AverageConnTime    time.Duration // 平均连接时间
	LastCleanup        time.Time     // 最后清理时间
}

// DefaultConnectionPoolConfig 默认连接池配置
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		DialTimeout:         30 * time.Second,
		KeepAlive:           30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		CleanupInterval:     5 * time.Minute,
		MaxLifetime:         30 * time.Minute,
		EnableStats:         true,
		StatsInterval:       1 * time.Minute,
	}
}

// NewConnectionPool 创建连接池
func NewConnectionPool(config *ConnectionPoolConfig) *ConnectionPool {
	if config == nil {
		config = DefaultConnectionPoolConfig()
	}

	pool := &ConnectionPool{
		pools:       make(map[string]*http.Transport),
		config:      config,
		cleanupDone: make(chan struct{}),
		stats:       &ConnectionStats{},
	}

	// 启动清理协程
	pool.startCleanup()

	// 启动统计协程
	if config.EnableStats {
		pool.startStats()
	}

	return pool
}

// GetTransport 获取HTTP传输层
func (cp *ConnectionPool) GetTransport(key string, tlsConfig *types.TLSConfig) *http.Transport {
	cp.mu.RLock()
	transport, exists := cp.pools[key]
	cp.mu.RUnlock()

	if exists {
		cp.updateStats(func(stats *ConnectionStats) {
			stats.ConnectionsReused++
		})
		return transport
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// 双重检查
	if transport, exists := cp.pools[key]; exists {
		cp.updateStats(func(stats *ConnectionStats) {
			stats.ConnectionsReused++
		})
		return transport
	}

	// 创建新的传输层
	transport = cp.createTransport(tlsConfig)
	cp.pools[key] = transport

	cp.updateStats(func(stats *ConnectionStats) {
		stats.ConnectionsCreated++
		stats.TotalConnections++
	})

	return transport
}

// createTransport 创建HTTP传输层
func (cp *ConnectionPool) createTransport(tlsConfig *types.TLSConfig) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   cp.config.DialTimeout,
		KeepAlive: cp.config.KeepAlive,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          cp.config.MaxIdleConns,
		MaxIdleConnsPerHost:   cp.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cp.config.MaxConnsPerHost,
		IdleConnTimeout:       cp.config.IdleConnTimeout,
		TLSHandshakeTimeout:   cp.config.TLSHandshakeTimeout,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 应用TLS配置
	if tlsConfig != nil {
		if tlsConf, err := cp.buildTLSConfig(tlsConfig); err == nil {
			transport.TLSClientConfig = tlsConf
		}
	}

	return transport
}

// buildTLSConfig 构建TLS配置
func (cp *ConnectionPool) buildTLSConfig(tlsConfig *types.TLSConfig) (*tls.Config, error) {
	// 这里复用之前实现的TLS配置构建逻辑
	// 为了简化，直接返回默认TLS配置
	return &tls.Config{}, nil
}

// RemoveTransport 移除传输层
func (cp *ConnectionPool) RemoveTransport(key string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if transport, exists := cp.pools[key]; exists {
		transport.CloseIdleConnections()
		delete(cp.pools, key)

		cp.updateStats(func(stats *ConnectionStats) {
			stats.ConnectionsClosed++
			stats.TotalConnections--
		})
	}
}

// CloseAll 关闭所有连接
func (cp *ConnectionPool) CloseAll() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for key, transport := range cp.pools {
		transport.CloseIdleConnections()
		delete(cp.pools, key)
	}

	// 停止清理协程
	close(cp.cleanupDone)
	if cp.cleanupTicker != nil {
		cp.cleanupTicker.Stop()
	}

	cp.updateStats(func(stats *ConnectionStats) {
		stats.TotalConnections = 0
		stats.ActiveConnections = 0
		stats.IdleConnections = 0
	})
}

// startCleanup 启动清理协程
func (cp *ConnectionPool) startCleanup() {
	// 确保清理间隔不为0，避免panic
	interval := cp.config.CleanupInterval
	if interval <= 0 {
		interval = 5 * time.Minute // 默认5分钟
	}
	
	cp.cleanupTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-cp.cleanupTicker.C:
				cp.cleanup()
			case <-cp.cleanupDone:
				return
			}
		}
	}()
}

// cleanup 清理过期连接
func (cp *ConnectionPool) cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	now := time.Now()

	// 清理空闲连接
	for key, transport := range cp.pools {
		// 这里可以实现更复杂的清理逻辑
		// 例如基于连接的创建时间、使用频率等
		_ = key
		transport.CloseIdleConnections()
	}

	cp.updateStats(func(stats *ConnectionStats) {
		stats.LastCleanup = now
	})
}

// startStats 启动统计协程
func (cp *ConnectionPool) startStats() {
	go func() {
		ticker := time.NewTicker(cp.config.StatsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cp.collectStats()
			case <-cp.cleanupDone:
				return
			}
		}
	}()
}

// collectStats 收集统计信息
func (cp *ConnectionPool) collectStats() {
	cp.mu.RLock()
	poolCount := len(cp.pools)
	cp.mu.RUnlock()

	cp.updateStats(func(stats *ConnectionStats) {
		// 这里可以收集更详细的统计信息
		// 例如每个传输层的连接状态
		_ = poolCount
	})
}

// updateStats 更新统计信息
func (cp *ConnectionPool) updateStats(updater func(*ConnectionStats)) {
	if !cp.config.EnableStats {
		return
	}

	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()

	updater(cp.stats)
}

// GetStats 获取统计信息
func (cp *ConnectionPool) GetStats() ConnectionStats {
	if !cp.config.EnableStats {
		return ConnectionStats{}
	}

	cp.stats.mu.RLock()
	defer cp.stats.mu.RUnlock()

	return *cp.stats
}

// HealthCheck 健康检查
func (cp *ConnectionPool) HealthCheck(ctx context.Context) error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	// 检查连接池状态
	if len(cp.pools) == 0 {
		return nil // 空连接池是正常的
	}

	// 这里可以实现更复杂的健康检查逻辑
	// 例如测试连接的可用性

	return nil
}

// ConnectionPoolManager 连接池管理器
type ConnectionPoolManager struct {
	mu    sync.RWMutex
	pools map[string]*ConnectionPool
}

// NewConnectionPoolManager 创建连接池管理器
func NewConnectionPoolManager() *ConnectionPoolManager {
	return &ConnectionPoolManager{
		pools: make(map[string]*ConnectionPool),
	}
}

// GetPool 获取连接池
func (cpm *ConnectionPoolManager) GetPool(name string, config *ConnectionPoolConfig) *ConnectionPool {
	cpm.mu.RLock()
	pool, exists := cpm.pools[name]
	cpm.mu.RUnlock()

	if exists {
		return pool
	}

	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	// 双重检查
	if pool, exists := cpm.pools[name]; exists {
		return pool
	}

	// 创建新的连接池
	pool = NewConnectionPool(config)
	cpm.pools[name] = pool

	return pool
}

// RemovePool 移除连接池
func (cpm *ConnectionPoolManager) RemovePool(name string) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if pool, exists := cpm.pools[name]; exists {
		pool.CloseAll()
		delete(cpm.pools, name)
	}
}

// CloseAll 关闭所有连接池
func (cpm *ConnectionPoolManager) CloseAll() {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	for name, pool := range cpm.pools {
		pool.CloseAll()
		delete(cpm.pools, name)
	}
}

// GetAllStats 获取所有连接池统计信息
func (cpm *ConnectionPoolManager) GetAllStats() map[string]ConnectionStats {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	stats := make(map[string]ConnectionStats)
	for name, pool := range cpm.pools {
		stats[name] = pool.GetStats()
	}

	return stats
}

// 全局连接池管理器实例
var globalConnectionPoolManager = NewConnectionPoolManager()

// GetGlobalConnectionPool 获取全局连接池
func GetGlobalConnectionPool(name string, config *ConnectionPoolConfig) *ConnectionPool {
	return globalConnectionPoolManager.GetPool(name, config)
}

// CloseGlobalConnectionPools 关闭所有全局连接池
func CloseGlobalConnectionPools() {
	globalConnectionPoolManager.CloseAll()
}
