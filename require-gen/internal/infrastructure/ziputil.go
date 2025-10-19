package infrastructure

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"specify-cli/internal/types"
)

// ZipProcessor ZIP文件处理器接口
type ZipProcessor interface {
	ExtractZip(zipPath, targetDir string, opts *ExtractOptions) error
	ListZipContents(zipPath string) ([]string, error)
	ValidateZip(zipPath string) error
	ExtractWithProgress(zipPath, targetDir string, opts *ExtractOptions,
		progressCallback func(current, total int64)) error
}

// ExtractOptions 提取选项配置
type ExtractOptions struct {
	OverwriteExisting   bool     // 是否覆盖现有文件
	FlattenStructure    bool     // 是否扁平化目录结构
	MergeDirectories    bool     // 是否合并目录
	PreservePermissions bool     // 是否保持文件权限
	SkipHidden          bool     // 是否跳过隐藏文件
	MaxFileSize         int64    // 最大文件大小限制（字节）
	AllowedExtensions   []string // 允许的文件扩展名
	Verbose             bool     // 详细输出模式
	TempDir             string   // 临时目录路径
}

// ZipProcessorImpl ZIP处理器实现
type ZipProcessorImpl struct {
	sysOps types.SystemOperations // 系统操作接口
	mu     sync.RWMutex           // 读写锁，保证线程安全
}

// NewZipProcessor 创建ZIP处理器实例
func NewZipProcessor(sysOps types.SystemOperations) ZipProcessor {
	return &ZipProcessorImpl{
		sysOps: sysOps,
	}
}

// ZipError ZIP操作错误类型
type ZipError struct {
	Operation string // 操作类型
	Path      string // 文件路径
	Cause     error  // 原因错误
}

// Error 实现error接口
func (ze *ZipError) Error() string {
	return fmt.Sprintf("zip %s failed for %s: %v", ze.Operation, ze.Path, ze.Cause)
}

// Unwrap 实现errors.Unwrap接口
func (ze *ZipError) Unwrap() error {
	return ze.Cause
}

// ListZipContents 列出ZIP文件内容
func (zp *ZipProcessorImpl) ListZipContents(zipPath string) ([]string, error) {
	zp.mu.RLock()
	defer zp.mu.RUnlock()

	// 验证ZIP文件路径
	if zipPath == "" {
		return nil, &ZipError{
			Operation: "list",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip file path cannot be empty"),
		}
	}

	// 检查文件是否存在
	if !zp.sysOps.FileExists(zipPath) {
		return nil, &ZipError{
			Operation: "list",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip file does not exist"),
		}
	}

	// 打开ZIP文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, &ZipError{
			Operation: "list",
			Path:      zipPath,
			Cause:     fmt.Errorf("failed to open zip file: %w", err),
		}
	}
	defer reader.Close()

	// 提取文件列表
	var contents []string
	for _, file := range reader.File {
		// 标准化路径分隔符
		normalizedPath := filepath.ToSlash(file.Name)
		contents = append(contents, normalizedPath)
	}

	return contents, nil
}

// validateZipInternal 内部ZIP文件验证方法，不使用锁
func (zp *ZipProcessorImpl) validateZipInternal(zipPath string) error {
	// 验证ZIP文件路径
	if zipPath == "" {
		return &ZipError{
			Operation: "validate",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip file path cannot be empty"),
		}
	}

	// 检查文件是否存在
	if !zp.sysOps.FileExists(zipPath) {
		return &ZipError{
			Operation: "validate",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip file does not exist"),
		}
	}

	// 尝试打开ZIP文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return &ZipError{
			Operation: "validate",
			Path:      zipPath,
			Cause:     fmt.Errorf("invalid zip file format: %w", err),
		}
	}
	defer reader.Close()

	// 验证ZIP文件结构
	if len(reader.File) == 0 {
		return &ZipError{
			Operation: "validate",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip file is empty"),
		}
	}

	// 验证每个文件条目
	for _, file := range reader.File {
		// 检查文件名有效性
		if file.Name == "" {
			return &ZipError{
				Operation: "validate",
				Path:      zipPath,
				Cause:     fmt.Errorf("zip contains entry with empty name"),
			}
		}

		// 检查路径安全性（防止路径遍历攻击）
		if err := zp.validateZipEntryPath(file.Name); err != nil {
			return &ZipError{
				Operation: "validate",
				Path:      zipPath,
				Cause:     fmt.Errorf("unsafe entry path '%s': %w", file.Name, err),
			}
		}
	}

	return nil
}

// ValidateZip 验证ZIP文件完整性
func (zp *ZipProcessorImpl) ValidateZip(zipPath string) error {
	zp.mu.RLock()
	defer zp.mu.RUnlock()

	return zp.validateZipInternal(zipPath)
}

// validateZipEntryPath 验证ZIP条目路径安全性
func (zp *ZipProcessorImpl) validateZipEntryPath(entryPath string) error {
	// 检查空路径
	if entryPath == "" {
		return fmt.Errorf("empty path")
	}

	// 标准化路径
	cleanPath := filepath.Clean(entryPath)

	// 检查路径遍历
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains '..' traversal")
	}

	// 检查绝对路径
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute path not allowed")
	}

	// 检查路径长度
	if len(cleanPath) > 260 { // Windows路径长度限制
		return fmt.Errorf("path too long (%d characters)", len(cleanPath))
	}

	// 检查危险字符（Windows）
	if strings.ContainsAny(cleanPath, "<>:\"|?*") {
		return fmt.Errorf("path contains dangerous characters")
	}

	// 检查控制字符
	for _, r := range cleanPath {
		if r < 32 {
			return fmt.Errorf("path contains control characters")
		}
	}

	return nil
}

// ExtractZip 提取ZIP文件到目标目录
func (zp *ZipProcessorImpl) ExtractZip(zipPath, targetDir string, opts *ExtractOptions) error {
	zp.mu.Lock()
	defer zp.mu.Unlock()

	// 参数验证
	if zipPath == "" {
		return &ZipError{
			Operation: "extract",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip file path cannot be empty"),
		}
	}

	if targetDir == "" {
		return &ZipError{
			Operation: "extract",
			Path:      zipPath,
			Cause:     fmt.Errorf("target directory cannot be empty"),
		}
	}

	// 设置默认选项
	if opts == nil {
		opts = &ExtractOptions{}
	}

	// 验证ZIP文件 - 直接调用内部方法避免死锁
	if err := zp.validateZipInternal(zipPath); err != nil {
		return &ZipError{
			Operation: "extract",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip validation failed: %w", err),
		}
	}

	// 创建目标目录
	if err := zp.sysOps.CreateDirectory(targetDir); err != nil {
		return &ZipError{
			Operation: "extract",
			Path:      zipPath,
			Cause:     fmt.Errorf("failed to create target directory: %w", err),
		}
	}

	// 打开ZIP文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return &ZipError{
			Operation: "extract",
			Path:      zipPath,
			Cause:     fmt.Errorf("failed to open zip file: %w", err),
		}
	}
	defer reader.Close()

	// 提取所有文件
	for _, file := range reader.File {
		if err := zp.extractSingleFile(file, targetDir, opts); err != nil {
			return &ZipError{
				Operation: "extract",
				Path:      zipPath,
				Cause:     fmt.Errorf("failed to extract file '%s': %w", file.Name, err),
			}
		}
	}

	return nil
}

// ExtractWithProgress 带进度的提取操作
func (zp *ZipProcessorImpl) ExtractWithProgress(zipPath, targetDir string, opts *ExtractOptions,
	progressCallback func(current, total int64)) error {
	zp.mu.Lock()
	defer zp.mu.Unlock()

	// 参数验证
	if zipPath == "" {
		return &ZipError{
			Operation: "extract_progress",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip file path cannot be empty"),
		}
	}

	if targetDir == "" {
		return &ZipError{
			Operation: "extract_progress",
			Path:      zipPath,
			Cause:     fmt.Errorf("target directory cannot be empty"),
		}
	}

	// 设置默认选项
	if opts == nil {
		opts = &ExtractOptions{}
	}

	// 验证ZIP文件 - 直接调用内部方法避免死锁
	if err := zp.validateZipInternal(zipPath); err != nil {
		return &ZipError{
			Operation: "extract_progress",
			Path:      zipPath,
			Cause:     fmt.Errorf("zip validation failed: %w", err),
		}
	}

	// 创建目标目录
	if err := zp.sysOps.CreateDirectory(targetDir); err != nil {
		return &ZipError{
			Operation: "extract_progress",
			Path:      zipPath,
			Cause:     fmt.Errorf("failed to create target directory: %w", err),
		}
	}

	// 打开ZIP文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return &ZipError{
			Operation: "extract_progress",
			Path:      zipPath,
			Cause:     fmt.Errorf("failed to open zip file: %w", err),
		}
	}
	defer reader.Close()

	// 计算总文件数
	totalFiles := int64(len(reader.File))
	var currentFile int64 = 0

	// 提取所有文件并报告进度
	for _, file := range reader.File {
		if err := zp.extractSingleFile(file, targetDir, opts); err != nil {
			return &ZipError{
				Operation: "extract_progress",
				Path:      zipPath,
				Cause:     fmt.Errorf("failed to extract file '%s': %w", file.Name, err),
			}
		}

		// 更新进度
		currentFile++
		if progressCallback != nil {
			progressCallback(currentFile, totalFiles)
		}
	}

	return nil
}

// extractSingleFile 提取单个文件
func (zp *ZipProcessorImpl) extractSingleFile(file *zip.File, targetDir string, opts *ExtractOptions) error {
	// 验证文件路径安全性
	if err := zp.validateZipEntryPath(file.Name); err != nil {
		return fmt.Errorf("unsafe file path: %w", err)
	}

	// 处理隐藏文件
	if opts.SkipHidden && zp.isHiddenFile(file.Name) {
		return nil // 跳过隐藏文件
	}

	// 处理扩展名过滤
	if len(opts.AllowedExtensions) > 0 && !zp.isAllowedExtension(file.Name, opts.AllowedExtensions) {
		return nil // 跳过不允许的文件类型
	}

	// 检查文件大小限制
	if opts.MaxFileSize > 0 && file.UncompressedSize64 > uint64(opts.MaxFileSize) {
		return fmt.Errorf("file size (%d bytes) exceeds limit (%d bytes)",
			file.UncompressedSize64, opts.MaxFileSize)
	}

	// 计算目标路径
	targetPath := zp.calculateTargetPath(file.Name, targetDir, opts)

	// 创建目标文件的父目录
	targetParent := filepath.Dir(targetPath)
	if err := zp.sysOps.CreateDirectory(targetParent); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// 处理目录
	if file.FileInfo().IsDir() {
		return zp.sysOps.CreateDirectory(targetPath)
	}

	// 检查文件是否已存在
	if zp.sysOps.FileExists(targetPath) && !opts.OverwriteExisting {
		if opts.Verbose {
			fmt.Printf("Skipping existing file: %s\n", targetPath)
		}
		return nil
	}

	// 提取文件内容
	return zp.extractFileContentWithProgress(file, targetPath, opts)
}

// isHiddenFile 检查文件是否为隐藏文件
func (zp *ZipProcessorImpl) isHiddenFile(filename string) bool {
	base := filepath.Base(filename)
	return strings.HasPrefix(base, ".")
}

// isAllowedExtension 检查文件扩展名是否被允许
func (zp *ZipProcessorImpl) isAllowedExtension(filename string, allowedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExts {
		if strings.ToLower(allowed) == ext {
			return true
		}
	}
	return false
}

// calculateTargetPath 计算目标文件路径
func (zp *ZipProcessorImpl) calculateTargetPath(entryPath, targetDir string, opts *ExtractOptions) string {
	if opts.FlattenStructure {
		// 扁平化结构，只保留文件名
		return filepath.Join(targetDir, filepath.Base(entryPath))
	}
	// 保持原有目录结构
	return filepath.Join(targetDir, entryPath)
}

// extractFileContentWithProgress 带进度跟踪的文件内容提取
func (zp *ZipProcessorImpl) extractFileContentWithProgress(file *zip.File, targetPath string, opts *ExtractOptions) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer writer.Close()

	// 使用缓冲区进行流式复制
	buffer := make([]byte, 32*1024) // 32KB缓冲区
	var copied int64

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if _, writeErr := writer.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			copied += int64(n)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// 设置文件权限
	if opts.PreservePermissions {
		if err := os.Chmod(targetPath, file.Mode()); err != nil {
			// 忽略权限设置错误，只在详细模式下输出
			if opts.Verbose {
				fmt.Printf("Warning: Failed to set file permissions for %s: %v\n", targetPath, err)
			}
		}
	}

	return nil
}
