package infrastructure

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"specify-cli/internal/config"
	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

// TemplateProvider 模板提供者
type TemplateProvider struct {
	client        *resty.Client
	authProvider  types.AuthProvider
	networkConfig *types.NetworkConfig
	httpConfig    *types.HTTPClientConfig
	clientManager *HTTPClientManager
	errorHandler  *NetworkErrorHandler
	retryManager  *RetryManager
}

// NewTemplateProvider 创建新的模板提供者实例
func NewTemplateProvider() types.TemplateProvider {
	return NewTemplateProviderWithConfig(nil, nil)
}

// NewTemplateProviderWithConfig 创建带配置的模板提供者实例
func NewTemplateProviderWithConfig(networkConfig *types.NetworkConfig, httpConfig *types.HTTPClientConfig) types.TemplateProvider {
	// 设置默认配置
	if httpConfig == nil {
		httpConfig = DefaultHTTPClientConfig()
	}

	// 创建HTTP客户端管理器
	clientManager := NewHTTPClientManager(httpConfig)

	// 创建错误处理器
	errorHandler := NewNetworkErrorHandler(nil)

	// 创建重试管理器
	retryManager := NewRetryManager(nil, errorHandler)

	// 获取默认客户端
	client := clientManager.GetDefaultClient()

	tp := &TemplateProvider{
		client:        client,
		authProvider:  NewAuthProvider(),
		networkConfig: networkConfig,
		httpConfig:    httpConfig,
		clientManager: clientManager,
		errorHandler:  errorHandler,
		retryManager:  retryManager,
	}

	// 应用配置
	tp.applyHTTPConfig()
	tp.applyNetworkConfig()

	return tp
}

// applyHTTPConfig 应用HTTP客户端配置
func (tp *TemplateProvider) applyHTTPConfig() {
	if tp.httpConfig == nil {
		return
	}

	// 基本配置
	tp.client.SetTimeout(tp.httpConfig.Timeout)
	tp.client.SetRetryCount(tp.httpConfig.RetryCount)
	tp.client.SetRetryWaitTime(tp.httpConfig.RetryWaitTime)
	tp.client.SetRetryMaxWaitTime(tp.httpConfig.MaxRetryWaitTime)

	// 重定向配置
	if !tp.httpConfig.FollowRedirects {
		tp.client.SetRedirectPolicy(resty.NoRedirectPolicy())
	} else if tp.httpConfig.MaxRedirects > 0 {
		tp.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(tp.httpConfig.MaxRedirects))
	}

	// User-Agent
	if tp.httpConfig.UserAgent != "" {
		tp.client.SetHeader("User-Agent", tp.httpConfig.UserAgent)
	}

	// 自定义头部
	if tp.httpConfig.Headers != nil {
		tp.client.SetHeaders(tp.httpConfig.Headers)
	}

	// Cookies
	if tp.httpConfig.Cookies != nil {
		for name, value := range tp.httpConfig.Cookies {
			tp.client.SetCookie(&http.Cookie{
				Name:  name,
				Value: value,
			})
		}
	}

	// 传输层配置
	transport := &http.Transport{
		MaxIdleConns:        tp.httpConfig.MaxIdleConns,
		MaxIdleConnsPerHost: tp.httpConfig.MaxConnsPerHost,
		IdleConnTimeout:     tp.httpConfig.IdleConnTimeout,
	}

	tp.client.SetTransport(transport)
}

// applyNetworkConfig 应用网络配置
func (tp *TemplateProvider) applyNetworkConfig() {
	if tp.networkConfig == nil {
		return
	}

	// 代理配置
	if tp.networkConfig.ProxyURL != "" {
		tp.client.SetProxy(tp.networkConfig.ProxyURL)
	}

	// 超时配置
	if tp.networkConfig.Timeout > 0 {
		tp.client.SetTimeout(tp.networkConfig.Timeout)
	}

	// 重试配置
	if tp.networkConfig.RetryCount > 0 {
		tp.client.SetRetryCount(tp.networkConfig.RetryCount)
	}
	if tp.networkConfig.RetryWaitTime > 0 {
		tp.client.SetRetryWaitTime(tp.networkConfig.RetryWaitTime)
	}

	// TLS配置
	if tp.networkConfig.TLS != nil {
		tlsConfig, err := tp.buildTLSConfig(tp.networkConfig.TLS)
		if err == nil {
			tp.client.SetTLSClientConfig(tlsConfig)
		}
	}
}

// buildTLSConfig 构建TLS配置
func (tp *TemplateProvider) buildTLSConfig(tlsConf *types.TLSConfig) (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: tlsConf.InsecureSkipVerify,
		ServerName:         tlsConf.ServerName,
	}

	// 加载客户端证书
	if tlsConf.CertFile != "" && tlsConf.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsConf.CertFile, tlsConf.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// 加载CA证书
	if tlsConf.CAFile != "" {
		caCert, err := ioutil.ReadFile(tlsConf.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		config.RootCAs = caCertPool
	}

	return config, nil
}

// Download 下载模板
func (tp *TemplateProvider) Download(opts types.DownloadOptions) (string, error) {
	// 获取AI助手信息
	agentInfo, exists := config.GetAgentInfo(opts.AIAssistant)
	if !exists {
		return "", fmt.Errorf("unknown AI assistant: %s", opts.AIAssistant)
	}

	// 构建目标路径
	targetDir := filepath.Join(opts.DownloadDir, agentInfo.Folder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	// 获取最新发布信息
	release, err := tp.getLatestRelease(opts.GitHubToken)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}

	// 查找合适的资源
	asset, err := tp.findAsset(release, opts.AIAssistant, opts.ScriptType)
	if err != nil {
		return "", fmt.Errorf("failed to find suitable asset: %w", err)
	}

	// 下载资源
	downloadPath := filepath.Join(targetDir, asset.Name)
	if err := tp.downloadAsset(asset, downloadPath, opts); err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}

	// 提取文件（如果是压缩包）
	if err := tp.extractAsset(downloadPath, targetDir, opts); err != nil {
		return "", fmt.Errorf("failed to extract asset: %w", err)
	}

	return targetDir, nil
}

// Validate 验证模板
func (tp *TemplateProvider) Validate(path string) error {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("template path does not exist: %s", path)
	}

	// 检查必需的文件
	requiredFiles := []string{
		"README.md",
		"spec-template.md",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(path, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", file)
		}
	}

	// 验证模板结构
	if err := tp.validateTemplateStructure(path); err != nil {
		return fmt.Errorf("invalid template structure: %w", err)
	}

	return nil
}

// getLatestRelease 获取最新发布信息
func (tp *TemplateProvider) getLatestRelease(token string) (*types.GitHubRelease, error) {
	url := "https://api.github.com/repos/github/spec-kit/releases/latest"

	req := tp.client.R()

	// 添加认证头
	if token != "" {
		req.SetHeader("Authorization", fmt.Sprintf("token %s", token))
	}

	// 设置User-Agent
	req.SetHeader("User-Agent", "Specify-CLI/1.0.0")

	resp, err := req.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var release types.GitHubRelease
	if err := json.Unmarshal(resp.Body(), &release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

// findAsset 查找合适的资源
func (tp *TemplateProvider) findAsset(release *types.GitHubRelease, assistant, scriptType string) (*types.Asset, error) {
	// 构建资源名称模式，匹配实际的GitHub发布资产格式
	patterns := []string{
		fmt.Sprintf("spec-kit-template-%s-", assistant), // 匹配 spec-kit-template-copilot-v0.0.20.zip
		fmt.Sprintf("template-%s-", assistant),          // 匹配 template-copilot-v0.0.20.zip
		fmt.Sprintf("%s-%s.zip", assistant, scriptType), // 匹配 copilot-ps.zip
		fmt.Sprintf("%s.zip", assistant),                // 匹配 copilot.zip
		fmt.Sprintf("templates-%s.zip", assistant),      // 匹配 templates-copilot.zip
		"templates.zip",                                 // 匹配 templates.zip
	}

	for _, pattern := range patterns {
		for _, asset := range release.Assets {
			// 对于前两个模式，使用包含匹配；对于其他模式，使用精确匹配或包含匹配
			if strings.Contains(asset.Name, pattern) {
				// 确保找到的是zip文件
				if strings.HasSuffix(asset.Name, ".zip") {
					return &asset, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no suitable asset found for %s with script type %s", assistant, scriptType)
}

// downloadAsset 下载资源
func (tp *TemplateProvider) downloadAsset(asset *types.Asset, downloadPath string, opts types.DownloadOptions) error {
	if opts.Verbose {
		ui.ShowInfo(fmt.Sprintf("Downloading %s (%d bytes)", asset.Name, asset.Size))
	}

	// 使用增强的下载方法
	return tp.downloadWithEnhancedProgress(asset.BrowserDownloadURL, downloadPath, asset.Size, opts)
}

// downloadWithEnhancedProgress 增强的带进度下载方法
func (tp *TemplateProvider) downloadWithEnhancedProgress(url, filePath string, size int64, opts types.DownloadOptions) error {
	// 如果配置了流式下载，使用流式下载器
	if opts.ChunkSize > 0 || opts.EnableResume || opts.MaxConcurrent > 1 {
		// 创建HTTP客户端
		client := &http.Client{
			Timeout: tp.getTimeout(),
		}

		// 应用TLS配置
		if tp.networkConfig != nil && tp.networkConfig.TLS != nil {
			tlsConfig, err := tp.buildTLSConfig(tp.networkConfig.TLS)
			if err == nil {
				transport := &http.Transport{
					TLSClientConfig: tlsConfig,
				}
				client.Transport = transport
			}
		}

		// 使用流式下载器
		downloader := &StreamingDownloader{
			client:       client,
			chunkSize:    1024 * 1024, // 1MB
			maxRetries:   3,
			retryWait:    2 * time.Second,
			maxRetryWait: 10 * time.Second,
		}
		return downloader.DownloadWithStreaming(url, filePath, &opts)
	}

	// 使用增强的下载方法，集成错误处理
	return tp.downloadWithErrorHandling(url, filePath, size, opts)
}

// downloadWithErrorHandling 使用错误处理的下载
func (tp *TemplateProvider) downloadWithErrorHandling(url, dest string, size int64, opts types.DownloadOptions) error {
	// 创建增强下载器
	downloader := NewEnhancedDownloader(tp.client, tp.errorHandler, tp.retryManager)
	
	// 执行下载
	return downloader.Download(context.Background(), url, dest, &opts)
}

// getTimeout 获取超时时间
func (tp *TemplateProvider) getTimeout() time.Duration {
	if tp.networkConfig != nil && tp.networkConfig.Timeout > 0 {
		return tp.networkConfig.Timeout
	}
	if tp.httpConfig != nil && tp.httpConfig.Timeout > 0 {
		return tp.httpConfig.Timeout
	}
	return 30 * time.Second // 默认值
}

// getRetryCount 获取重试次数
func (tp *TemplateProvider) getRetryCount() int {
	if tp.networkConfig != nil && tp.networkConfig.RetryCount > 0 {
		return tp.networkConfig.RetryCount
	}
	if tp.httpConfig != nil && tp.httpConfig.RetryCount > 0 {
		return tp.httpConfig.RetryCount
	}
	return 3 // 默认值
}

// getRetryWaitTime 获取重试等待时间
func (tp *TemplateProvider) getRetryWaitTime() time.Duration {
	if tp.networkConfig != nil && tp.networkConfig.RetryWaitTime > 0 {
		return tp.networkConfig.RetryWaitTime
	}
	if tp.httpConfig != nil && tp.httpConfig.RetryWaitTime > 0 {
		return tp.httpConfig.RetryWaitTime
	}
	return 2 * time.Second // 默认值
}

// getMaxRetryWaitTime 获取最大重试等待时间
func (tp *TemplateProvider) getMaxRetryWaitTime() time.Duration {
	if tp.httpConfig != nil && tp.httpConfig.MaxRetryWaitTime > 0 {
		return tp.httpConfig.MaxRetryWaitTime
	}
	return 10 * time.Second // 默认值
}

// extractAsset 提取资源
func (tp *TemplateProvider) extractAsset(assetPath, targetDir string, opts types.DownloadOptions) error {
	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(assetPath))

	switch ext {
	case ".zip":
		return tp.extractZip(assetPath, targetDir, opts)
	case ".tar", ".gz":
		return tp.extractTar(assetPath, targetDir, opts)
	default:
		// 不是压缩文件，直接返回
		return nil
	}
}

// extractZip 提取ZIP文件
func (tp *TemplateProvider) extractZip(zipPath, targetDir string, opts types.DownloadOptions) error {
	if opts.Verbose {
		ui.ShowInfo(fmt.Sprintf("Extracting %s", filepath.Base(zipPath)))
	}

	// 创建系统操作实例
	sysOps := NewSystemOperations()

	// 创建ZIP处理器
	zipProcessor := NewZipProcessor(sysOps)

	// 配置提取选项
	extractOpts := &ExtractOptions{
		OverwriteExisting:   true,
		PreservePermissions: true,
		FlattenStructure:    false,
		MaxFileSize:         100 * 1024 * 1024, // 100MB
		AllowedExtensions:   []string{},        // 允许所有扩展名
		SkipHidden:          false,
		Verbose:             opts.Verbose,
	}

	// 执行ZIP提取
	if opts.Verbose {
		// 使用带进度的提取
		progressCallback := func(current, total int64) {
			if total > 0 {
				percentage := float64(current) / float64(total) * 100
				ui.ShowInfo(fmt.Sprintf("Extracting... %.1f%% (%d/%d bytes)",
					percentage, current, total))
			}
		}

		err := zipProcessor.ExtractWithProgress(zipPath, targetDir, extractOpts, progressCallback)
		if err != nil {
			return fmt.Errorf("failed to extract ZIP file: %w", err)
		}
	} else {
		// 使用普通提取
		err := zipProcessor.ExtractZip(zipPath, targetDir, extractOpts)
		if err != nil {
			return fmt.Errorf("failed to extract ZIP file: %w", err)
		}
	}

	if opts.Verbose {
		ui.ShowSuccess("Extraction completed")
	}

	return nil
}

// extractTar 提取TAR文件
func (tp *TemplateProvider) extractTar(tarPath, targetDir string, opts types.DownloadOptions) error {
	if opts.Verbose {
		ui.ShowInfo(fmt.Sprintf("Extracting %s", filepath.Base(tarPath)))
	}

	// 创建TAR处理器
	sysOps := NewSystemOperations()
	tarProcessor := NewTarProcessor(sysOps)

	// 配置解压选项
	extractOpts := &ExtractOptions{
		OverwriteExisting:   true,
		PreservePermissions: true,
		FlattenStructure:    false,
		SkipHidden:          true,
		MaxFileSize:         100 * 1024 * 1024, // 100MB 限制
	}

	// 进度回调函数
	progressCallback := func(current, total int64) {
		if opts.Verbose && total > 0 {
			percentage := float64(current) / float64(total) * 100
			fmt.Printf("\rExtracting... %.1f%%", percentage)
		}
	}

	var err error
	if opts.Verbose {
		// 带进度显示的解压
		err = tarProcessor.ExtractWithProgress(tarPath, targetDir, extractOpts, progressCallback)
		if opts.Verbose && err == nil {
			fmt.Println() // 换行
		}
	} else {
		// 静默解压
		err = tarProcessor.ExtractTar(tarPath, targetDir, extractOpts)
	}

	if err != nil {
		return fmt.Errorf("failed to extract TAR file: %w", err)
	}

	if opts.Verbose {
		ui.ShowSuccess("Extraction completed")
	}

	return nil
}

// validateTemplateStructure 验证模板结构
func (tp *TemplateProvider) validateTemplateStructure(path string) error {
	// 检查目录结构
	expectedDirs := []string{
		"templates",
		"scripts",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(path, dir)
		if stat, err := os.Stat(dirPath); err != nil || !stat.IsDir() {
			return fmt.Errorf("expected directory not found: %s", dir)
		}
	}

	return nil
}

// GetTemplateInfo 获取模板信息
func (tp *TemplateProvider) GetTemplateInfo(path string) (map[string]interface{}, error) {
	infoFile := filepath.Join(path, "template-info.json")

	if _, err := os.Stat(infoFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("template info file not found")
	}

	data, err := os.ReadFile(infoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read template info: %w", err)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse template info: %w", err)
	}

	return info, nil
}

// ListTemplates 列出可用模板
func (tp *TemplateProvider) ListTemplates(token string) ([]string, error) {
	release, err := tp.getLatestRelease(token)
	if err != nil {
		return nil, err
	}

	var templates []string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			// 从文件名提取模板名称
			name := strings.TrimSuffix(asset.Name, ".zip")
			templates = append(templates, name)
		}
	}

	return templates, nil
}

// downloadWithProgress 带进度的下载
func (tp *TemplateProvider) downloadWithProgress(url, filePath string, size int64) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建进度读取器
	reader := &progressReader{
		Reader: resp.Body,
		total:  size,
	}

	_, err = io.Copy(file, reader)
	return err
}

// progressReader 进度读取器
type progressReader struct {
	io.Reader
	total   int64
	current int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.current += int64(n)

	// 计算进度百分比
	if pr.total > 0 {
		percent := float64(pr.current) / float64(pr.total) * 100
		fmt.Printf("\rDownloading... %.1f%%", percent)
	}

	return n, err
}
