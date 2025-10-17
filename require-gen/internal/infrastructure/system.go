package infrastructure

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"specify-cli/internal/types"
)

// SystemOperations 系统操作实现
type SystemOperations struct{}

// NewSystemOperations 创建新的系统操作实例
func NewSystemOperations() *SystemOperations {
	return &SystemOperations{}
}

// GetOS 获取操作系统信息
func (so *SystemOperations) GetOS() string {
	return runtime.GOOS
}

// GetArch 获取系统架构
func (so *SystemOperations) GetArch() string {
	return runtime.GOARCH
}

// GetSystemInfo 获取系统信息
func (so *SystemOperations) GetSystemInfo() types.SystemInfo {
	return types.SystemInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		Compiler:     runtime.Compiler,
	}
}

// ExecuteCommand 执行系统命令
func (so *SystemOperations) ExecuteCommand(name string, args ...string) (*types.CommandResult, error) {
	cmd := exec.Command(name, args...)
	
	// 设置环境变量
	cmd.Env = os.Environ()
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	
	result := &types.CommandResult{
		Command:    fmt.Sprintf("%s %s", name, strings.Join(args, " ")),
		Output:     string(output),
		Success:    err == nil,
	}
	
	if err != nil {
		result.Error = err.Error()
		
		// 尝试获取退出码
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
	}
	
	return result, nil
}

// ExecuteCommandInDir 在指定目录执行命令
func (so *SystemOperations) ExecuteCommandInDir(dir, name string, args ...string) (*types.CommandResult, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
	
	result := &types.CommandResult{
		Command:    fmt.Sprintf("%s %s", name, strings.Join(args, " ")),
		Output:     string(output),
		Success:    err == nil,
		WorkingDir: dir,
	}
	
	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
	}
	
	return result, nil
}

// CreateDirectory 创建目录
func (so *SystemOperations) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// RemoveDirectory 删除目录
func (so *SystemOperations) RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}

// CopyFile 复制文件
func (so *SystemOperations) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// 复制文件内容
	_, err = sourceFile.WriteTo(destFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// 复制文件权限
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// MoveFile 移动文件
func (so *SystemOperations) MoveFile(src, dst string) error {
	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	return os.Rename(src, dst)
}

// FileExists 检查文件是否存在
func (so *SystemOperations) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirectoryExists 检查目录是否存在
func (so *SystemOperations) DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetCurrentDirectory 获取当前工作目录
func (so *SystemOperations) GetCurrentDirectory() (string, error) {
	return os.Getwd()
}

// ChangeDirectory 改变工作目录
func (so *SystemOperations) ChangeDirectory(path string) error {
	return os.Chdir(path)
}

// GetHomeDirectory 获取用户主目录
func (so *SystemOperations) GetHomeDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}

// GetTempDirectory 获取临时目录
func (so *SystemOperations) GetTempDirectory() string {
	return os.TempDir()
}

// SetEnvironmentVariable 设置环境变量
func (so *SystemOperations) SetEnvironmentVariable(key, value string) error {
	return os.Setenv(key, value)
}

// GetEnvironmentVariable 获取环境变量
func (so *SystemOperations) GetEnvironmentVariable(key string) string {
	return os.Getenv(key)
}

// GetAllEnvironmentVariables 获取所有环境变量
func (so *SystemOperations) GetAllEnvironmentVariables() []string {
	return os.Environ()
}

// IsExecutable 检查文件是否可执行
func (so *SystemOperations) IsExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// 在Windows上，检查文件扩展名
	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(path))
		executableExts := []string{".exe", ".bat", ".cmd", ".com", ".ps1"}
		for _, execExt := range executableExts {
			if ext == execExt {
				return true
			}
		}
		return false
	}

	// 在Unix系统上，检查执行权限
	mode := info.Mode()
	return mode&0111 != 0
}

// FindExecutable 查找可执行文件
func (so *SystemOperations) FindExecutable(name string) (string, error) {
	return exec.LookPath(name)
}

// GetFileSize 获取文件大小
func (so *SystemOperations) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}

// GetFileModTime 获取文件修改时间
func (so *SystemOperations) GetFileModTime(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.ModTime().Unix(), nil
}

// ListDirectory 列出目录内容
func (so *SystemOperations) ListDirectory(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	return files, nil
}

// CreateTempFile 创建临时文件
func (so *SystemOperations) CreateTempFile(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	return file.Name(), nil
}

// CreateTempDirectory 创建临时目录
func (so *SystemOperations) CreateTempDirectory(pattern string) (string, error) {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	return dir, nil
}