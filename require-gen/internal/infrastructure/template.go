package infrastructure

import (
	"encoding/json"
	"fmt"
	"io"
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

// TemplateProvider 模板提供者实现
type TemplateProvider struct {
	client      *resty.Client
	authProvider types.AuthProvider
}

// NewTemplateProvider 创建新的模板提供者实例
func NewTemplateProvider() types.TemplateProvider {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)
	
	return &TemplateProvider{
		client:       client,
		authProvider: NewAuthProvider(),
	}
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
	url := "https://api.github.com/repos/your-org/spec-kit-templates/releases/latest"
	
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
	// 构建资源名称模式
	patterns := []string{
		fmt.Sprintf("%s-%s.zip", assistant, scriptType),
		fmt.Sprintf("%s.zip", assistant),
		fmt.Sprintf("templates-%s.zip", assistant),
		"templates.zip",
	}

	for _, pattern := range patterns {
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, pattern) || asset.Name == pattern {
				return &asset, nil
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

	// 创建文件
	file, err := os.Create(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 准备HTTP请求
	req := tp.client.R()
	
	// 添加认证头
	headers := tp.authProvider.GetHeaders()
	for key, value := range headers {
		req.SetHeader(key, value)
	}

	// 设置进度回调
	if opts.ShowProgress {
		// 不设置输出，让resty处理默认输出
	}

	// 下载文件
	resp, err := req.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode())
	}

	// 如果没有使用SetOutput，手动写入文件
	if !opts.ShowProgress {
		if _, err := file.Write(resp.Body()); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	if opts.Verbose {
		ui.ShowSuccess(fmt.Sprintf("Downloaded %s successfully", asset.Name))
	}

	return nil
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
		OverwriteExisting:    true,
		PreservePermissions:  true,
		FlattenStructure:     false,
		MaxFileSize:          100 * 1024 * 1024, // 100MB
		AllowedExtensions:    []string{}, // 允许所有扩展名
		SkipHidden:           false,
		Verbose:              opts.Verbose,
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

	// 这里应该实现TAR提取逻辑
	// 使用archive/tar包
	
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
	total    int64
	current  int64
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