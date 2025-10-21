package infrastructure

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"specify-cli/internal/types"
)

// TarProcessor 定义TAR文件处理接口
type TarProcessor interface {
	// ExtractTar 解压TAR文件到目标目录
	ExtractTar(tarPath, targetDir string, opts *ExtractOptions) error
	// ExtractWithProgress 带进度显示的TAR解压
	ExtractWithProgress(tarPath, targetDir string, opts *ExtractOptions, 
		progressCallback func(current, total int64)) error
}

// TarProcessorImpl TAR处理器实现
type TarProcessorImpl struct {
	sysOps types.SystemOperations
}

// NewTarProcessor 创建新的TAR处理器
func NewTarProcessor(sysOps types.SystemOperations) TarProcessor {
	return &TarProcessorImpl{
		sysOps: sysOps,
	}
}

// ExtractTar 解压TAR文件到目标目录
func (tp *TarProcessorImpl) ExtractTar(tarPath, targetDir string, opts *ExtractOptions) error {
	if opts == nil {
		opts = &ExtractOptions{}
	}

	// 验证输入参数
	if tarPath == "" {
		return fmt.Errorf("tar file path cannot be empty")
	}

	if targetDir == "" {
		return fmt.Errorf("target directory cannot be empty")
	}

	// 检查TAR文件是否存在
	if !tp.sysOps.FileExists(tarPath) {
		return fmt.Errorf("tar file does not exist: %s", tarPath)
	}

	// 创建目标目录
	if err := tp.sysOps.CreateDirectory(targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 打开TAR文件
	fileData, err := tp.sysOps.ReadFile(tarPath)
	if err != nil {
		return fmt.Errorf("failed to read tar file: %w", err)
	}

	// 创建字节读取器
	reader := bytes.NewReader(fileData)
	// 检查是否为gzip压缩的TAR文件
	var tarReader *tar.Reader
	if strings.HasSuffix(strings.ToLower(tarPath), ".gz") || 
	   strings.HasSuffix(strings.ToLower(tarPath), ".tgz") {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else {
		tarReader = tar.NewReader(reader)
	}

	// 解压所有条目
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 到达文件末尾
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		if err := tp.extractSingleEntry(header, tarReader, targetDir, opts); err != nil {
			return fmt.Errorf("failed to extract entry %s: %w", header.Name, err)
		}
	}

	return nil
}

// ExtractWithProgress 带进度显示的TAR解压
func (tp *TarProcessorImpl) ExtractWithProgress(tarPath, targetDir string, opts *ExtractOptions,
	progressCallback func(current, total int64)) error {
	
	if opts == nil {
		opts = &ExtractOptions{}
	}

	// 验证输入参数
	if tarPath == "" {
		return fmt.Errorf("tar file path cannot be empty")
	}

	if targetDir == "" {
		return fmt.Errorf("target directory cannot be empty")
	}

	// 创建目标目录
	if err := tp.sysOps.CreateDirectory(targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 打开TAR文件
	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	// 获取文件大小用于进度计算
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	totalSize := fileInfo.Size()

	// 检查是否为gzip压缩的tar文件
	var reader io.Reader = file
	if strings.HasSuffix(strings.ToLower(tarPath), ".gz") || 
	   strings.HasSuffix(strings.ToLower(tarPath), ".tgz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// 创建进度读取器
	progressReader := &tarProgressReader{
		Reader:   reader,
		total:    totalSize,
		current:  0,
		callback: progressCallback,
	}

	// 创建TAR读取器
	tarReader := tar.NewReader(progressReader)

	var currentEntry int64 = 0
	var totalEntries int64 = 0
	var skippedFiles []string
	var errorFiles []string

	// 首先计算总条目数（如果需要）
	if opts.OnProgress != nil {
		// 重新打开文件来计算条目数
		countFile, err := os.Open(tarPath)
		if err == nil {
			defer countFile.Close()
			
			var countReader io.Reader = countFile
			if strings.HasSuffix(strings.ToLower(tarPath), ".gz") || 
			   strings.HasSuffix(strings.ToLower(tarPath), ".tgz") {
				if gzReader, err := gzip.NewReader(countFile); err == nil {
					defer gzReader.Close()
					countReader = gzReader
				}
			}
			
			countTarReader := tar.NewReader(countReader)
			for {
				if _, err := countTarReader.Next(); err != nil {
					break
				}
				totalEntries++
			}
		}
	}

	// 解压所有条目
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if opts.OnError != nil {
				if !opts.OnError("", fmt.Errorf("failed to read tar entry: %w", err)) {
					return fmt.Errorf("extraction stopped by user after tar read error: %w", err)
				}
				continue
			}
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// 调用详细进度回调
		if opts.OnProgress != nil {
			opts.OnProgress(currentEntry, totalEntries, header.Name)
		}

		// 解压单个条目
		if err := tp.extractSingleEntry(header, tarReader, targetDir, opts); err != nil {
			if opts.OnError != nil {
				if !opts.OnError(header.Name, err) {
					return fmt.Errorf("extraction stopped by user after error in entry '%s': %w", header.Name, err)
				}
				errorFiles = append(errorFiles, header.Name)
			} else {
				return fmt.Errorf("failed to extract entry '%s': %w", header.Name, err)
			}
		}

		currentEntry++
		
		// 详细模式输出
		if opts.Verbose {
			fmt.Printf("Extracted: %s (%d/%d)\n", header.Name, currentEntry, totalEntries)
		}
	}

	// 输出摘要信息
	if opts.Verbose {
		fmt.Printf("TAR extraction completed: %d entries processed", currentEntry)
		if len(skippedFiles) > 0 {
			fmt.Printf(", %d entries skipped", len(skippedFiles))
		}
		if len(errorFiles) > 0 {
			fmt.Printf(", %d entries had errors", len(errorFiles))
		}
		fmt.Println()
	}

	return nil
}

// extractSingleEntry 解压单个TAR条目
func (tp *TarProcessorImpl) extractSingleEntry(header *tar.Header, reader *tar.Reader, 
	targetDir string, opts *ExtractOptions) error {
	
	// 计算目标路径
	targetPath := tp.calculateTargetPath(header.Name, targetDir, opts)
	
	// 安全检查：防止路径遍历攻击
	if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", header.Name)
	}

	switch header.Typeflag {
	case tar.TypeDir:
		// 创建目录
		return tp.sysOps.CreateDirectory(targetPath)
		
	case tar.TypeReg:
		// 提取常规文件
		return tp.extractRegularFile(header, reader, targetPath, opts)
		
	case tar.TypeSymlink:
		// 创建符号链接（如果支持）
		if opts.PreservePermissions {
			return os.Symlink(header.Linkname, targetPath)
		}
		return nil
		
	default:
		// 跳过不支持的文件类型
		return nil
	}
}

// extractRegularFile 解压常规文件
func (tp *TarProcessorImpl) extractRegularFile(header *tar.Header, reader *tar.Reader, 
	targetPath string, opts *ExtractOptions) error {
	
	// 确保父目录存在
	parentDir := filepath.Dir(targetPath)
	if err := tp.sysOps.CreateDirectory(parentDir); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// 读取文件内容
	fileData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// 写入文件
	if err := tp.sysOps.WriteFile(targetPath, fileData); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// 设置文件权限（如果需要）
	if opts.PreservePermissions {
		if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}
	}

	return nil
}

// handleFileConflict 处理文件冲突（TAR版本）
func (tp *TarProcessorImpl) handleFileConflict(conflictInfo *FileConflictInfo, opts *ExtractOptions) (string, ConflictAction, error) {
	if !opts.SmartMerge {
		// 如果没有启用智能合并，使用传统逻辑
		if opts.OverwriteExisting {
			return conflictInfo.TargetPath, ConflictOverwrite, nil
		}
		return "", ConflictSkip, nil
	}

	// 优先使用回调函数处理冲突
	if opts.OnConflict != nil {
		action := opts.OnConflict(conflictInfo)
		switch action {
		case ConflictSkip:
			return "", ConflictSkip, nil
		case ConflictOverwrite:
			return conflictInfo.TargetPath, ConflictOverwrite, nil
		case ConflictRename:
			newPath := tp.generateUniqueFileName(conflictInfo.TargetPath)
			return newPath, ConflictRename, nil
		default:
			return conflictInfo.TargetPath, ConflictOverwrite, nil
		}
	}

	// 根据配置的冲突解决策略处理
	switch opts.ConflictResolution {
	case "skip":
		return "", ConflictSkip, nil
	case "overwrite":
		return conflictInfo.TargetPath, ConflictOverwrite, nil
	case "rename":
		newPath := tp.generateUniqueFileName(conflictInfo.TargetPath)
		return newPath, ConflictRename, nil
	case "prompt":
		// 这里可以实现交互式提示，暂时返回默认行为
		if opts.Verbose {
			fmt.Printf("File conflict: %s already exists\n", conflictInfo.TargetPath)
		}
		return conflictInfo.TargetPath, ConflictOverwrite, nil
	default:
		// 默认策略：如果源文件更新则覆盖，否则跳过
		if tp.isSourceNewer(conflictInfo) {
			return conflictInfo.TargetPath, ConflictOverwrite, nil
		}
		return "", ConflictSkip, nil
	}
}

// generateUniqueFileName 生成唯一文件名（TAR版本）
func (tp *TarProcessorImpl) generateUniqueFileName(originalPath string) string {
	dir := filepath.Dir(originalPath)
	name := filepath.Base(originalPath)
	ext := filepath.Ext(name)
	nameWithoutExt := strings.TrimSuffix(name, ext)

	counter := 1
	for {
		newName := fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		newPath := filepath.Join(dir, newName)
		if !tp.sysOps.FileExists(newPath) {
			return newPath
		}
		counter++
	}
}

// isSourceNewer 检查源文件是否比目标文件更新（TAR版本）
func (tp *TarProcessorImpl) isSourceNewer(conflictInfo *FileConflictInfo) bool {
	targetModTime, err := tp.sysOps.GetFileModTime(conflictInfo.TargetPath)
	if err != nil {
		return true // 如果无法获取目标文件时间，默认认为源文件更新
	}
	return conflictInfo.ModTime > targetModTime
}

// calculateTargetPath 计算目标路径
func (tp *TarProcessorImpl) calculateTargetPath(entryPath, targetDir string, opts *ExtractOptions) string {
	if opts.FlattenStructure {
		return filepath.Join(targetDir, filepath.Base(entryPath))
	}
	return filepath.Join(targetDir, entryPath)
}

// tarProgressReader 带进度跟踪的读取器
type tarProgressReader struct {
	io.Reader
	total    int64
	current  int64
	callback func(current, total int64)
}

func (pr *tarProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.current += int64(n)
	pr.callback(pr.current, pr.total)
	return n, err
}