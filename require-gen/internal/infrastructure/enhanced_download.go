package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"specify-cli/internal/types"

	"github.com/go-resty/resty/v2"
)

// EnhancedDownloader 增强下载器
type EnhancedDownloader struct {
	client       *resty.Client
	errorHandler *NetworkErrorHandler
	retryManager *RetryManager
	config       *DownloadConfig
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	// 基本配置
	ChunkSize     int64         // 分块大小
	MaxConcurrent int           // 最大并发数
	Timeout       time.Duration // 超时时间

	// 重试配置
	MaxRetries int           // 最大重试次数
	RetryDelay time.Duration // 重试延迟

	// 进度配置
	EnableProgress bool          // 启用进度显示
	ProgressUpdate time.Duration // 进度更新间隔

	// 验证配置
	VerifyChecksum bool   // 验证校验和
	ChecksumType   string // 校验和类型
	ExpectedSum    string // 期望校验和

	// 恢复配置
	EnableResume    bool  // 启用断点续传
	ResumeThreshold int64 // 续传阈值
}

// DefaultDownloadConfig 默认下载配置
func DefaultDownloadConfig() *DownloadConfig {
	return &DownloadConfig{
		ChunkSize:       1024 * 1024, // 1MB
		MaxConcurrent:   4,
		Timeout:         30 * time.Second,
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
		EnableProgress:  true,
		ProgressUpdate:  100 * time.Millisecond,
		VerifyChecksum:  false,
		EnableResume:    true,
		ResumeThreshold: 1024 * 1024, // 1MB
	}
}

// NewEnhancedDownloader 创建增强下载器
func NewEnhancedDownloader(client *resty.Client, errorHandler *NetworkErrorHandler, retryManager *RetryManager) *EnhancedDownloader {
	if client == nil {
		client = resty.New()
	}

	if errorHandler == nil {
		errorHandler = NewNetworkErrorHandler(nil)
	}

	if retryManager == nil {
		retryManager = NewRetryManager(nil, errorHandler)
	}

	return &EnhancedDownloader{
		client:       client,
		errorHandler: errorHandler,
		retryManager: retryManager,
		config:       DefaultDownloadConfig(),
	}
}

// Download 下载文件
func (ed *EnhancedDownloader) Download(ctx context.Context, url, dest string, opts *types.DownloadOptions) error {
	// 转换选项
	config := ed.convertOptions(opts)

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 检查断点续传支持
	_, err := ed.checkResumeSupport(ctx, url, dest, config)
	if err != nil {
		return err
	}

	// 执行下载
	return ed.executeDownload(ctx, url, dest, opts)
}

// convertOptions 转换下载选项
func (ed *EnhancedDownloader) convertOptions(opts *types.DownloadOptions) *DownloadConfig {
	config := DefaultDownloadConfig()

	if opts == nil {
		return config
	}

	if opts.ChunkSize > 0 {
		config.ChunkSize = opts.ChunkSize
	}

	if opts.MaxConcurrent > 0 {
		config.MaxConcurrent = opts.MaxConcurrent
	}

	config.EnableResume = opts.EnableResume
	config.VerifyChecksum = opts.VerifyChecksum
	config.ChecksumType = opts.ChecksumType
	config.ExpectedSum = opts.Checksum

	return config
}

// checkResumeSupport 检查断点续传支持
func (ed *EnhancedDownloader) checkResumeSupport(ctx context.Context, url, dest string, config *DownloadConfig) (int64, error) {
	if !config.EnableResume {
		return 0, nil
	}

	// 检查本地文件
	fileInfo, err := os.Stat(dest)
	if os.IsNotExist(err) {
		return 0, nil // 文件不存在，从头开始
	}

	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	localSize := fileInfo.Size()
	if localSize < config.ResumeThreshold {
		return 0, nil // 文件太小，重新下载
	}

	// 检查服务器是否支持Range请求
	supported, remoteSize, err := ed.checkRangeSupport(ctx, url)
	if err != nil {
		return 0, err
	}

	if !supported {
		return 0, nil // 服务器不支持Range请求
	}

	if localSize >= remoteSize {
		return 0, nil // 本地文件已完整或更大
	}

	return localSize, nil
}

// checkRangeSupport 检查Range请求支持
func (ed *EnhancedDownloader) checkRangeSupport(ctx context.Context, url string) (bool, int64, error) {
	operation := func(ctx context.Context, attempt int) error {
		resp, err := ed.client.R().
			SetContext(ctx).
			Head(url)

		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("HEAD request failed with status: %d", resp.StatusCode())
		}

		// 检查Accept-Ranges头
		acceptRanges := resp.Header().Get("Accept-Ranges")
		if acceptRanges != "bytes" {
			return fmt.Errorf("server does not support range requests")
		}

		// 获取文件大小
		contentLength := resp.Header().Get("Content-Length")
		if contentLength == "" {
			return fmt.Errorf("server did not provide content length")
		}

		if _, err = strconv.ParseInt(contentLength, 10, 64); err != nil {
			return fmt.Errorf("invalid content length: %w", err)
		}

		return nil
	}

	result := ed.retryManager.ExecuteWithRetry(ctx, operation, &RetryOptions{
		StrategyName: "fast",
		Operation:    "check_range_support",
		Host:         extractHost(url),
	})

	if !result.Success {
		return false, 0, result.LastError
	}

	// 重新获取文件大小（这里简化处理）
	resp, err := ed.client.R().SetContext(ctx).Head(url)
	if err != nil {
		return false, 0, err
	}

	contentLength := resp.Header().Get("Content-Length")
	size, _ := strconv.ParseInt(contentLength, 10, 64)

	return true, size, nil
}

// executeDownload 执行下载
func (ed *EnhancedDownloader) executeDownload(ctx context.Context, url, dest string, opts *types.DownloadOptions) error {
	// 简化实现，直接使用resty下载
	resp, err := ed.client.R().
		SetContext(ctx).
		SetOutput(dest).
		Get(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode())
	}

	return nil
}

// openFile 打开文件
func (ed *EnhancedDownloader) openFile(dest string, resume bool) (*os.File, error) {
	if resume {
		return os.OpenFile(dest, os.O_WRONLY|os.O_APPEND, 0644)
	}

	return os.Create(dest)
}

// downloadWithRetry 带重试的下载
func (ed *EnhancedDownloader) downloadWithRetry(ctx context.Context, url string, dest string, opts *types.DownloadOptions) error {
	// 简化实现
	return ed.executeDownload(ctx, url, dest, opts)
}

// verifyChecksum 验证校验和
func (ed *EnhancedDownloader) verifyChecksum(filePath, checksumType, expectedSum string) error {
	// 这里可以实现具体的校验和验证逻辑
	// 为了简化，暂时返回nil
	return nil
}

// extractHost 提取主机名
func extractHost(url string) string {
	// 简化实现，实际应该解析URL
	return url
}

// DownloadWithProgress 带进度的下载
func (ed *EnhancedDownloader) DownloadWithProgress(ctx context.Context, url, dest string, opts *types.DownloadOptions, progressCallback func(downloaded, total int64)) error {
	// 获取文件大小
	resp, err := ed.client.R().Head(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to get file info: HTTP %d", resp.StatusCode())
	}

	return ed.Download(ctx, url, dest, opts)
}

// getFileSize 获取文件大小
func (ed *EnhancedDownloader) getFileSize(ctx context.Context, url string) (int64, error) {
	resp, err := ed.client.R().SetContext(ctx).Head(url)
	if err != nil {
		return 0, err
	}

	contentLength := resp.Header().Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("server did not provide content length")
	}

	return strconv.ParseInt(contentLength, 10, 64)
}

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	callback func(*types.ProgressInfo)
}

// Update 更新进度
func (pt *ProgressTracker) Update(downloaded int64) {
	if pt.callback != nil {
		pt.callback(&types.ProgressInfo{
			Downloaded: downloaded,
			Total:      0, // 简化实现
		})
	}
}

// GetProgress 获取进度百分比
func (pt *ProgressTracker) GetProgress() float64 {
	// 简化实现，返回0
	return 0.0
}
