package infrastructure

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"specify-cli/internal/types"
)

// StreamingDownloader 流式下载器
type StreamingDownloader struct {
	client       *http.Client
	chunkSize    int64
	maxRetries   int
	retryWait    time.Duration
	maxRetryWait time.Duration
}

// NewStreamingDownloader 创建流式下载器
func NewStreamingDownloader(client *http.Client, chunkSize int64) *StreamingDownloader {
	if chunkSize <= 0 {
		chunkSize = 1024 * 1024 // 默认1MB
	}

	return &StreamingDownloader{
		client:       client,
		chunkSize:    chunkSize,
		maxRetries:   3,
		retryWait:    2 * time.Second,
		maxRetryWait: 10 * time.Second,
	}
}

// DownloadWithStreaming 流式下载文件
func (sd *StreamingDownloader) DownloadWithStreaming(url, filePath string, opts *types.DownloadOptions) error {
	// 获取文件信息
	size, supportsRange, err := sd.getFileInfo(url)
	if err != nil {
		return CreateNetworkError(fmt.Errorf("failed to get file info: %w", err), url, 0)
	}

	// 检查是否支持断点续传
	var startPos int64 = 0
	if opts.EnableResume && supportsRange {
		if fileInfo, err := os.Stat(filePath); err == nil {
			startPos = fileInfo.Size()
			if startPos >= size {
				// 文件已完整下载
				return nil
			}
		}
	}

	// 创建或打开文件
	var file *os.File
	if startPos > 0 {
		file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		file, err = os.Create(filePath)
	}
	if err != nil {
		return CreateNetworkError(fmt.Errorf("failed to create/open file: %w", err), url, 0)
	}
	defer file.Close()

	// 初始化进度显示
	var progressDisplay types.ProgressDisplay
	if opts.ShowProgress && opts.ProgressDisplay != nil {
		progressDisplay = opts.ProgressDisplay
		progressDisplay.Start(size)
	}

	// 创建校验和计算器
	var hasher hash.Hash
	if opts.VerifyChecksum && opts.Checksum != "" {
		hasher = sd.createHasher(opts.ChecksumType)
	}

	// 流式下载
	err = sd.streamDownload(url, file, startPos, size, opts, progressDisplay, hasher)
	if err != nil {
		return err
	}

	// 验证校验和
	if hasher != nil {
		if err := sd.verifyChecksum(hasher, opts.Checksum); err != nil {
			return err
		}
	}

	// 完成进度显示
	if progressDisplay != nil {
		progressDisplay.Finish()
	}

	return nil
}

// getFileInfo 获取文件信息
func (sd *StreamingDownloader) getFileInfo(url string) (size int64, supportsRange bool, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, false, err
	}

	resp, err := sd.client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("HEAD request failed with status %d", resp.StatusCode)
	}

	// 获取文件大小
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		size, _ = strconv.ParseInt(contentLength, 10, 64)
	}

	// 检查是否支持范围请求
	acceptRanges := resp.Header.Get("Accept-Ranges")
	supportsRange = strings.ToLower(acceptRanges) == "bytes"

	return size, supportsRange, nil
}

// streamDownload 执行流式下载
func (sd *StreamingDownloader) streamDownload(url string, file *os.File, startPos, totalSize int64, opts *types.DownloadOptions, progressDisplay types.ProgressDisplay, hasher hash.Hash) error {
	var downloaded int64 = startPos
	startTime := time.Now()

	for downloaded < totalSize {
		// 计算当前块的范围
		endPos := downloaded + sd.chunkSize - 1
		if endPos >= totalSize {
			endPos = totalSize - 1
		}

		// 下载当前块
		err := ExecuteWithRetry(func() error {
			return sd.downloadChunk(url, file, downloaded, endPos, &downloaded, totalSize, startTime, opts, progressDisplay, hasher)
		}, sd.maxRetries, sd.retryWait, sd.maxRetryWait)

		if err != nil {
			return err
		}
	}

	return nil
}

// downloadChunk 下载单个数据块
func (sd *StreamingDownloader) downloadChunk(url string, file *os.File, startPos, endPos int64, downloaded *int64, totalSize int64, startTime time.Time, opts *types.DownloadOptions, progressDisplay types.ProgressDisplay, hasher hash.Hash) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return CreateNetworkError(err, url, 0)
	}

	// 设置范围请求头
	if startPos > 0 || endPos < totalSize-1 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startPos, endPos))
	}

	resp, err := sd.client.Do(req)
	if err != nil {
		return CreateNetworkError(err, url, 0)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return CreateNetworkError(
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status),
			url,
			resp.StatusCode,
		)
	}

	// 创建进度读取器
	reader := io.Reader(resp.Body)
	if progressDisplay != nil || opts.ProgressCallback != nil {
		reader = &ChunkProgressReader{
			Reader:           resp.Body,
			downloaded:       downloaded,
			totalSize:        totalSize,
			startTime:        startTime,
			progressDisplay:  progressDisplay,
			progressCallback: opts.ProgressCallback,
		}
	}

	// 如果需要校验和，创建多重写入器
	var writer io.Writer = file
	if hasher != nil {
		writer = io.MultiWriter(file, hasher)
	}

	// 复制数据
	_, err = io.Copy(writer, reader)
	if err != nil {
		return CreateNetworkError(fmt.Errorf("failed to copy data: %w", err), url, 0)
	}

	return nil
}

// ChunkProgressReader 分块进度读取器
type ChunkProgressReader struct {
	io.Reader
	downloaded       *int64
	totalSize        int64
	startTime        time.Time
	lastUpdate       time.Time
	progressDisplay  types.ProgressDisplay
	progressCallback func(*types.ProgressInfo)
	mu               sync.Mutex
}

// Read 读取数据并更新进度
func (cpr *ChunkProgressReader) Read(p []byte) (int, error) {
	n, err := cpr.Reader.Read(p)

	cpr.mu.Lock()
	*cpr.downloaded += int64(n)
	now := time.Now()

	// 限制更新频率（每100ms更新一次）
	if now.Sub(cpr.lastUpdate) < 100*time.Millisecond {
		cpr.mu.Unlock()
		return n, err
	}
	cpr.lastUpdate = now

	// 计算进度信息
	info := &types.ProgressInfo{
		Downloaded: *cpr.downloaded,
		Total:      cpr.totalSize,
		Percentage: float64(*cpr.downloaded) / float64(cpr.totalSize) * 100,
		StartTime:  cpr.startTime,
		LastUpdate: now,
	}

	// 计算下载速度
	elapsed := now.Sub(cpr.startTime)
	if elapsed > 0 {
		info.Speed = float64(*cpr.downloaded) / elapsed.Seconds()
	}

	// 计算预计剩余时间
	if info.Speed > 0 && *cpr.downloaded > 0 {
		remaining := cpr.totalSize - *cpr.downloaded
		info.ETA = time.Duration(float64(remaining)/info.Speed) * time.Second
	}

	cpr.mu.Unlock()

	// 更新显示
	if cpr.progressDisplay != nil {
		cpr.progressDisplay.Update(info)
	}

	// 调用回调函数
	if cpr.progressCallback != nil {
		cpr.progressCallback(info)
	}

	return n, err
}

// createHasher 创建校验和计算器
func (sd *StreamingDownloader) createHasher(checksumType string) hash.Hash {
	switch strings.ToLower(checksumType) {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha256", "":
		return sha256.New()
	default:
		return sha256.New() // 默认使用SHA256
	}
}

// verifyChecksum 验证校验和
func (sd *StreamingDownloader) verifyChecksum(hasher hash.Hash, expectedChecksum string) error {
	actualChecksum := fmt.Sprintf("%x", hasher.Sum(nil))
	if strings.ToLower(actualChecksum) != strings.ToLower(expectedChecksum) {
		return fmt.Errorf("checksum verification failed: expected %s, got %s", expectedChecksum, actualChecksum)
	}
	return nil
}

// ConcurrentDownloader 并发下载器
type ConcurrentDownloader struct {
	streamingDownloader *StreamingDownloader
	maxConcurrent       int
}

// NewConcurrentDownloader 创建并发下载器
func NewConcurrentDownloader(client *http.Client, chunkSize int64, maxConcurrent int) *ConcurrentDownloader {
	if maxConcurrent <= 0 {
		maxConcurrent = 4 // 默认4个并发
	}

	return &ConcurrentDownloader{
		streamingDownloader: NewStreamingDownloader(client, chunkSize),
		maxConcurrent:       maxConcurrent,
	}
}

// DownloadConcurrently 并发下载多个文件
func (cd *ConcurrentDownloader) DownloadConcurrently(downloads []DownloadTask) error {
	semaphore := make(chan struct{}, cd.maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, task := range downloads {
		wg.Add(1)
		go func(task DownloadTask) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行下载
			err := cd.streamingDownloader.DownloadWithStreaming(task.URL, task.FilePath, task.Options)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to download %s: %w", task.URL, err))
				mu.Unlock()
			}
		}(task)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("download errors: %v", errors)
	}

	return nil
}

// DownloadTask 下载任务
type DownloadTask struct {
	URL      string
	FilePath string
	Options  *types.DownloadOptions
}
