package infrastructure

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"specify-cli/internal/types"
)

// SystemOperations 系统操作实现
//
// SystemOperations 提供了跨平台的系统级操作接口，是require-gen框架中处理
// 底层系统交互的核心组件。它封装了文件系统操作、进程管理、环境变量处理
// 等基础功能，为上层业务逻辑提供统一的系统抽象层。
//
// 主要功能特性：
// - 跨平台兼容性：支持Windows、Linux、macOS等主流操作系统
// - 文件系统操作：文件/目录的创建、删除、移动、复制等操作
// - 进程管理：系统命令执行、进程控制、输出捕获
// - 环境管理：环境变量读写、系统信息获取、路径处理
// - 权限检查：可执行文件检测、文件权限验证
// - 临时资源：临时文件和目录的创建管理
//
// 设计原则：
// - 零依赖：仅使用Go标准库，无外部依赖
// - 错误透明：详细的错误信息和上下文
// - 性能优化：高效的文件操作和内存管理
// - 安全考虑：路径验证和权限检查
//
// 使用场景：
// - 项目初始化时的目录结构创建
// - 模板文件的下载和解压
// - 开发工具的安装和配置检查
// - Git仓库的初始化和管理
// - 跨平台脚本的执行
type SystemOperations struct{
	securityConfig *SecurityConfig // 安全配置
}

// NewSystemOperations 创建新的系统操作实例
//
// NewSystemOperations 是SystemOperations的工厂函数，采用简单工厂模式
// 创建并返回一个新的SystemOperations实例。该函数确保实例的正确初始化
// 并为后续的系统操作提供统一的入口点。
//
// 设计特点：
// - 工厂模式：提供统一的实例创建接口
// - 零配置：无需额外参数，开箱即用
// - 线程安全：返回的实例可在多goroutine中安全使用
// - 资源轻量：实例创建成本极低，可按需创建
//
// 返回值：
// - types.SystemOperations: 新创建的系统操作实例，包含所有系统级操作方法
//
// 使用示例：
//   sysOps := NewSystemOperations()
//   info := sysOps.GetSystemInfo()
//   result, err := sysOps.ExecuteCommand("git", "version")
//
// 注意事项：
// - 实例无状态，可重复创建使用
// - 建议在需要时创建，避免全局单例
// - 实例方法均为线程安全的
func NewSystemOperations() types.SystemOperations {
	return &SystemOperations{}
}

// GetOS 获取操作系统信息
//
// GetOS 返回当前运行环境的操作系统标识符，基于Go运行时的GOOS常量。
// 该方法提供跨平台的操作系统检测功能，用于条件逻辑和平台特定的操作。
//
// 支持的操作系统标识符：
// - "windows": Microsoft Windows系统
// - "linux": Linux系统（包括各种发行版）
// - "darwin": macOS系统
// - "freebsd": FreeBSD系统
// - "openbsd": OpenBSD系统
// - "netbsd": NetBSD系统
// - "solaris": Solaris系统
//
// 返回值：
// - string: 操作系统标识符，小写字符串格式
//
// 使用场景：
// - 平台特定的文件路径处理
// - 条件执行不同的系统命令
// - 选择合适的可执行文件扩展名
// - 平台相关的配置选择
//
// 使用示例：
//   os := sysOps.GetOS()
//   if os == "windows" {
//       // Windows特定逻辑
//   } else {
//       // Unix-like系统逻辑
//   }
func (so *SystemOperations) GetOS() string {
	return runtime.GOOS
}

// GetArch 获取系统架构
//
// GetArch 返回当前运行环境的处理器架构标识符，基于Go运行时的GOARCH常量。
// 该方法用于识别目标平台的硬件架构，支持架构相关的决策和优化。
//
// 支持的架构标识符：
// - "amd64": 64位x86架构（Intel/AMD 64位处理器）
// - "386": 32位x86架构（Intel/AMD 32位处理器）
// - "arm": 32位ARM架构（ARM处理器）
// - "arm64": 64位ARM架构（ARM64/AArch64处理器）
// - "mips": MIPS架构
// - "mips64": 64位MIPS架构
// - "ppc64": 64位PowerPC架构
// - "s390x": IBM System z架构
//
// 返回值：
// - string: 处理器架构标识符，小写字符串格式
//
// 使用场景：
// - 选择架构特定的二进制文件
// - 下载对应架构的依赖包
// - 性能优化和内存管理决策
// - 交叉编译目标选择
//
// 使用示例：
//   arch := sysOps.GetArch()
//   binaryName := fmt.Sprintf("tool-%s-%s", sysOps.GetOS(), arch)
//   // 例如: "tool-linux-amd64"
func (so *SystemOperations) GetArch() string {
	return runtime.GOARCH
}

// GetSystemInfo 获取系统信息
//
// GetSystemInfo 收集并返回当前运行环境的完整系统信息，包括操作系统、
// 架构、Go版本、CPU核心数和编译器信息。该方法提供系统环境的全面概览，
// 用于诊断、日志记录和环境兼容性检查。
//
// 返回的系统信息包含：
// - OS: 操作系统标识符（如"windows", "linux", "darwin"）
// - Arch: 处理器架构（如"amd64", "arm64"）
// - GoVersion: Go运行时版本（如"go1.21.0"）
// - NumCPU: 可用的CPU逻辑核心数
// - Compiler: Go编译器标识符（通常为"gc"）
//
// 返回值：
// - types.SystemInfo: 包含完整系统信息的结构体
//
// 使用场景：
// - 系统兼容性检查和验证
// - 性能基准测试和优化决策
// - 错误报告和诊断信息收集
// - 环境配置和调优建议
// - 日志记录和监控数据
//
// 使用示例：
//   info := sysOps.GetSystemInfo()
//   fmt.Printf("运行环境: %s/%s, Go版本: %s, CPU核心: %d\n",
//       info.OS, info.Arch, info.GoVersion, info.NumCPU)
//
// 注意事项：
// - NumCPU返回的是逻辑核心数，包括超线程
// - 信息基于Go运行时，可能与系统实际配置略有差异
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
//
// ExecuteCommand 在当前工作目录中执行指定的系统命令，并捕获其输出和执行状态。
// 该方法提供了跨平台的命令执行能力，支持完整的错误处理和状态报告。
//
// 执行特性：
// - 同步执行：等待命令完成后返回结果
// - 输出捕获：同时捕获标准输出和标准错误
// - 环境继承：继承当前进程的所有环境变量
// - 退出码获取：准确获取命令的退出状态码
// - 错误详情：提供详细的错误信息和上下文
// - 超时控制：防止命令无限期执行
//
// 参数：
// - name: 要执行的命令名称或可执行文件路径
// - args: 命令参数列表，可变参数
//
// 返回值：
// - *types.CommandResult: 命令执行结果，包含输出、状态、错误等信息
// - error: 执行过程中的错误（注意：命令失败不会返回error，而是在Result中标记）
//
// CommandResult结构包含：
// - Command: 完整的命令字符串
// - Output: 合并的标准输出和标准错误
// - Success: 命令是否成功执行（退出码为0）
// - ExitCode: 命令的退出状态码
// - Error: 错误描述信息
//
// 使用场景：
// - Git命令执行和版本控制操作
// - 开发工具的安装和配置检查
// - 系统信息查询和环境检测
// - 构建脚本和自动化任务执行
//
// 使用示例：
//   result, err := sysOps.ExecuteCommand("git", "version")
//   if err != nil {
//       // 处理执行错误
//   }
//   if result.Success {
//       fmt.Println("Git版本:", result.Output)
//   } else {
//       fmt.Printf("命令失败 (退出码: %d): %s\n", result.ExitCode, result.Error)
//   }
//
// 安全考虑：
// - 不进行命令注入防护，调用方需确保参数安全
// - 继承当前进程环境，可能包含敏感信息
// - 建议对用户输入进行验证和清理
func (so *SystemOperations) ExecuteCommand(name string, args ...string) (*types.CommandResult, error) {
	return so.ExecuteCommandWithOptions(name, args, &types.CommandOptions{
		Timeout: 30 * time.Second, // 默认30秒超时
	})
}

// ExecuteCommandInDir 在指定目录执行命令
func (so *SystemOperations) ExecuteCommandInDir(dir, name string, args ...string) (*types.CommandResult, error) {
	return so.ExecuteCommandWithOptions(name, args, &types.CommandOptions{
		WorkingDir: dir,
		Timeout:    30 * time.Second,
	})
}

// ExecuteCommandWithOptions 使用选项执行命令
//
// ExecuteCommandWithOptions 提供了更灵活的命令执行接口，支持多种执行选项和配置。
// 该方法是命令执行功能的核心实现，提供了完整的控制和监控能力。
//
// 支持的选项：
// - WorkingDir: 指定工作目录
// - Timeout: 执行超时时间
// - CaptureOutput: 是否捕获输出
// - Shell: 是否使用Shell模式执行
// - Env: 自定义环境变量
// - Input: 标准输入内容
//
// 参数：
// - name: 命令名称或路径
// - args: 命令参数列表
// - options: 执行选项配置
//
// 返回值：
// - *types.CommandResult: 详细的执行结果
// - error: 执行过程中的系统错误
func (so *SystemOperations) ExecuteCommandWithOptions(name string, args []string, options *types.CommandOptions) (*types.CommandResult, error) {
	if options == nil {
		options = &types.CommandOptions{
			Timeout: 30 * time.Second,
		}
	}

	// 创建上下文用于超时控制
	ctx := context.Background()
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// 构建命令
	var cmd *exec.Cmd
	if options.Shell {
		cmd = so.createShellCommand(ctx, name, args)
	} else {
		cmd = exec.CommandContext(ctx, name, args...)
	}

	// 设置工作目录
	if options.WorkingDir != "" {
		cmd.Dir = options.WorkingDir
	}

	// 设置环境变量
	if len(options.Env) > 0 {
		cmd.Env = append(os.Environ(), options.Env...)
	} else {
		cmd.Env = os.Environ()
	}

	// 设置标准输入
	if options.Input != "" {
		cmd.Stdin = strings.NewReader(options.Input)
	}

	// 准备结果结构
	result := &types.CommandResult{
		Command:    so.formatCommand(name, args),
		WorkingDir: options.WorkingDir,
	}

	// 执行命令并捕获输出
	startTime := time.Now()
	
	if options.CaptureOutput {
		output, err := cmd.CombinedOutput()
		result.Output = string(output)
		result.Duration = time.Since(startTime)
		
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			
			// 处理不同类型的错误
			if ctx.Err() == context.DeadlineExceeded {
				result.Error = fmt.Sprintf("command timed out after %v", options.Timeout)
				result.ExitCode = -1
			} else if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					result.ExitCode = status.ExitStatus()
				}
			}
		} else {
			result.Success = true
			result.ExitCode = 0
		}
	} else {
		// 分离标准输出和标准错误
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
		}

		// 启动命令
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start command: %w", err)
		}

		// 读取输出
		stdoutData, _ := io.ReadAll(stdout)
		stderrData, _ := io.ReadAll(stderr)
		
		result.Output = string(stdoutData)
		result.Error = string(stderrData)
		result.Duration = time.Since(startTime)

		// 等待命令完成
		err = cmd.Wait()
		if err != nil {
			result.Success = false
			if ctx.Err() == context.DeadlineExceeded {
				result.Error = fmt.Sprintf("command timed out after %v", options.Timeout)
				result.ExitCode = -1
			} else if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					result.ExitCode = status.ExitStatus()
				}
			}
		} else {
			result.Success = true
			result.ExitCode = 0
		}
	}

	return result, nil
}

// createShellCommand 创建Shell命令
//
// createShellCommand 根据操作系统创建适当的Shell命令，支持跨平台的Shell执行。
// 在Windows上使用PowerShell或cmd，在Unix系统上使用bash或sh。
//
// 参数：
// - ctx: 上下文对象，用于超时控制
// - name: 命令名称
// - args: 命令参数
//
// 返回值：
// - *exec.Cmd: 配置好的Shell命令对象
func (so *SystemOperations) createShellCommand(ctx context.Context, name string, args []string) *exec.Cmd {
	fullCommand := name
	if len(args) > 0 {
		fullCommand += " " + strings.Join(args, " ")
	}

	switch runtime.GOOS {
	case "windows":
		// 优先使用PowerShell，回退到cmd
		if _, err := exec.LookPath("powershell"); err == nil {
			return exec.CommandContext(ctx, "powershell", "-Command", fullCommand)
		}
		return exec.CommandContext(ctx, "cmd", "/C", fullCommand)
	default:
		// Unix系统：优先使用bash，回退到sh
		if _, err := exec.LookPath("bash"); err == nil {
			return exec.CommandContext(ctx, "bash", "-c", fullCommand)
		}
		return exec.CommandContext(ctx, "sh", "-c", fullCommand)
	}
}

// formatCommand 格式化命令字符串用于显示
func (so *SystemOperations) formatCommand(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return fmt.Sprintf("%s %s", name, strings.Join(args, " "))
}

// ExecuteCommandAsync 异步执行命令
//
// ExecuteCommandAsync 启动一个命令并立即返回，不等待命令完成。
// 返回的通道将在命令完成时接收结果。适用于长时间运行的命令或需要并发执行的场景。
//
// 参数：
// - name: 命令名称
// - args: 命令参数
// - options: 执行选项
//
// 返回值：
// - <-chan *types.CommandResult: 结果通道，命令完成时会发送结果
// - error: 启动命令时的错误
func (so *SystemOperations) ExecuteCommandAsync(name string, args []string, options *types.CommandOptions) (<-chan *types.CommandResult, error) {
	resultChan := make(chan *types.CommandResult, 1)
	
	go func() {
		defer close(resultChan)
		result, _ := so.ExecuteCommandWithOptions(name, args, options)
		resultChan <- result
	}()
	
	return resultChan, nil
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

// GetPathSeparator 获取路径分隔符
//
// GetPathSeparator 返回当前操作系统的路径分隔符。
// Windows系统返回反斜杠(\)，Unix系统返回正斜杠(/)。
//
// 返回值：
// - string: 路径分隔符字符串
func (so *SystemOperations) GetPathSeparator() string {
	return string(filepath.Separator)
}

// NormalizePath 标准化路径
//
// NormalizePath 将路径转换为当前操作系统的标准格式，
// 处理路径分隔符的差异和相对路径的解析。
//
// 参数：
// - path: 需要标准化的路径
//
// 返回值：
// - string: 标准化后的路径
func (so *SystemOperations) NormalizePath(path string) string {
	return filepath.Clean(path)
}

// JoinPath 连接路径
//
// JoinPath 使用当前操作系统的路径分隔符连接多个路径组件，
// 自动处理重复的分隔符和相对路径。
//
// 参数：
// - paths: 要连接的路径组件列表
//
// 返回值：
// - string: 连接后的完整路径
func (so *SystemOperations) JoinPath(paths ...string) string {
	return filepath.Join(paths...)
}

// GetExecutableExtension 获取可执行文件扩展名
//
// GetExecutableExtension 返回当前操作系统的可执行文件扩展名。
// Windows系统返回".exe"，Unix系统返回空字符串。
//
// 返回值：
// - string: 可执行文件扩展名
func (so *SystemOperations) GetExecutableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// GetShellCommand 获取默认Shell命令
//
// GetShellCommand 返回当前操作系统的默认Shell命令和参数。
// 该方法用于跨平台的Shell脚本执行。
//
// 返回值：
// - string: Shell命令名称
// - []string: Shell命令参数
func (so *SystemOperations) GetShellCommand() (string, []string) {
	switch runtime.GOOS {
	case "windows":
		// 优先使用PowerShell
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell", []string{"-NoProfile", "-Command"}
		}
		// 回退到cmd
		return "cmd", []string{"/C"}
	default:
		// Unix系统：优先使用bash
		if _, err := exec.LookPath("bash"); err == nil {
			return "bash", []string{"-c"}
		}
		// 回退到sh
		return "sh", []string{"-c"}
	}
}

// DetectShellType 检测当前Shell类型
//
// DetectShellType 通过环境变量和系统检测确定当前使用的Shell类型。
// 该方法用于Shell特定的功能和脚本生成。
//
// 返回值：
// - string: Shell类型标识符（如"powershell", "bash", "zsh", "cmd"等）
func (so *SystemOperations) DetectShellType() string {
	// 检查SHELL环境变量（Unix系统）
	if shell := os.Getenv("SHELL"); shell != "" {
		shellName := filepath.Base(shell)
		switch shellName {
		case "bash", "zsh", "fish", "tcsh", "csh":
			return shellName
		}
	}

	// 检查Windows特定环境
	if runtime.GOOS == "windows" {
		// 检查PowerShell相关环境变量
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
		// 检查是否在PowerShell Core中运行
		if os.Getenv("PWSH_VERSION") != "" {
			return "pwsh"
		}
		// 默认为cmd
		return "cmd"
	}

	// Unix系统默认检测
	if _, err := exec.LookPath("bash"); err == nil {
		return "bash"
	}
	if _, err := exec.LookPath("zsh"); err == nil {
		return "zsh"
	}
	
	return "sh" // 最后的回退选项
}

// GetPlatformSpecificPath 获取平台特定路径
//
// GetPlatformSpecificPath 根据操作系统返回特定的系统路径，
// 如配置目录、数据目录、缓存目录等。
//
// 参数：
// - pathType: 路径类型（"config", "data", "cache", "temp"）
//
// 返回值：
// - string: 平台特定的路径
// - error: 获取路径时的错误
func (so *SystemOperations) GetPlatformSpecificPath(pathType string) (string, error) {
	switch pathType {
	case "config":
		return so.getConfigDir()
	case "data":
		return so.getDataDir()
	case "cache":
		return so.getCacheDir()
	case "temp":
		return so.GetTempDirectory(), nil
	default:
		return "", fmt.Errorf("unsupported path type: %s", pathType)
	}
}

// getConfigDir 获取配置目录
func (so *SystemOperations) getConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData, nil
		}
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "AppData", "Roaming"), nil
	case "darwin":
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return xdgConfig, nil
		}
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config"), nil
	}
}

// getDataDir 获取数据目录
func (so *SystemOperations) getDataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return localAppData, nil
		}
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "AppData", "Local"), nil
	case "darwin":
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return xdgData, nil
		}
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	}
}

// getCacheDir 获取缓存目录
func (so *SystemOperations) getCacheDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "Temp"), nil
		}
		return so.GetTempDirectory(), nil
	case "darwin":
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Caches"), nil
	default:
		if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
			return xdgCache, nil
		}
		home, err := so.GetHomeDirectory()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".cache"), nil
	}
}

// IsWindowsSystem 检查是否为Windows系统
func (so *SystemOperations) IsWindowsSystem() bool {
	return runtime.GOOS == "windows"
}

// IsUnixSystem 检查是否为Unix系统
func (so *SystemOperations) IsUnixSystem() bool {
	return runtime.GOOS != "windows"
}

// IsMacOSSystem 检查是否为macOS系统
func (so *SystemOperations) IsMacOSSystem() bool {
	return runtime.GOOS == "darwin"
}

// IsLinuxSystem 检查是否为Linux系统
func (so *SystemOperations) IsLinuxSystem() bool {
	return runtime.GOOS == "linux"
}

// SystemError 系统操作错误类型
//
// SystemError 提供了详细的错误信息，包括操作类型、
// 错误代码、错误消息和上下文信息。
type SystemError struct {
	Operation string            // 操作类型
	Code      string            // 错误代码
	Message   string            // 错误消息
	Context   map[string]string // 上下文信息
	Cause     error             // 原始错误
}

// Error 实现error接口
func (se *SystemError) Error() string {
	if se.Context != nil && len(se.Context) > 0 {
		var contextStr strings.Builder
		for k, v := range se.Context {
			contextStr.WriteString(fmt.Sprintf(" %s=%s", k, v))
		}
		return fmt.Sprintf("system operation failed: %s [%s] %s%s", se.Operation, se.Code, se.Message, contextStr.String())
	}
	return fmt.Sprintf("system operation failed: %s [%s] %s", se.Operation, se.Code, se.Message)
}

// Unwrap 返回原始错误
func (se *SystemError) Unwrap() error {
	return se.Cause
}

// NewSystemError 创建系统错误
func NewSystemError(operation, code, message string, cause error) *SystemError {
	return &SystemError{
		Operation: operation,
		Code:      code,
		Message:   message,
		Context:   make(map[string]string),
		Cause:     cause,
	}
}

// WithContext 添加上下文信息
func (se *SystemError) WithContext(key, value string) *SystemError {
	if se.Context == nil {
		se.Context = make(map[string]string)
	}
	se.Context[key] = value
	return se
}

// validatePath 验证路径安全性
//
// validatePath 检查路径是否安全，防止路径遍历攻击
// 和访问受限制的系统目录。
//
// 参数：
// - path: 要验证的路径
// - operation: 操作类型（用于错误报告）
//
// 返回值：
// - error: 验证失败时的错误信息
func (so *SystemOperations) validatePath(path, operation string) error {
	if path == "" {
		return NewSystemError(operation, "EMPTY_PATH", "path cannot be empty", nil)
	}

	// 清理路径
	cleanPath := filepath.Clean(path)
	
	// 检查路径遍历
	if strings.Contains(cleanPath, "..") {
		return NewSystemError(operation, "PATH_TRAVERSAL", 
			"path traversal detected", nil).WithContext("path", path)
	}

	// 检查绝对路径的安全性（仅在Windows上）
	if runtime.GOOS == "windows" && filepath.IsAbs(cleanPath) {
		// 检查是否访问系统关键目录
		systemDirs := []string{
			"C:\\Windows\\System32",
			"C:\\Windows\\SysWOW64",
			"C:\\Program Files\\WindowsApps",
		}
		
		for _, sysDir := range systemDirs {
			if strings.HasPrefix(strings.ToLower(cleanPath), strings.ToLower(sysDir)) {
				return NewSystemError(operation, "RESTRICTED_PATH", 
					"access to system directory is restricted", nil).
					WithContext("path", path).WithContext("restricted_dir", sysDir)
			}
		}
	}

	return nil
}

// validateCommand 验证命令安全性
//
// validateCommand 检查命令是否安全，防止命令注入攻击。
//
// 参数：
// - command: 要验证的命令
// - args: 命令参数
//
// 返回值：
// - error: 验证失败时的错误信息
func (so *SystemOperations) validateCommand(command string, args []string) error {
	if command == "" {
		return NewSystemError("EXECUTE_COMMAND", "EMPTY_COMMAND", 
			"command cannot be empty", nil)
	}

	// 检查危险字符
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "[", "]"}
	for _, char := range dangerousChars {
		if strings.Contains(command, char) {
			return NewSystemError("EXECUTE_COMMAND", "DANGEROUS_COMMAND", 
				"command contains dangerous characters", nil).
				WithContext("command", command).WithContext("dangerous_char", char)
		}
	}

	// 检查参数中的危险内容
	for i, arg := range args {
		for _, char := range dangerousChars {
			if strings.Contains(arg, char) {
				return NewSystemError("EXECUTE_COMMAND", "DANGEROUS_ARGUMENT", 
					"argument contains dangerous characters", nil).
					WithContext("argument", arg).WithContext("index", fmt.Sprintf("%d", i)).
					WithContext("dangerous_char", char)
			}
		}
	}

	return nil
}

// handleFileError 处理文件操作错误
//
// handleFileError 将标准的文件系统错误转换为详细的SystemError。
//
// 参数：
// - operation: 操作类型
// - path: 文件路径
// - err: 原始错误
//
// 返回值：
// - error: 转换后的SystemError
func (so *SystemOperations) handleFileError(operation, path string, err error) error {
	if err == nil {
		return nil
	}

	sysErr := NewSystemError(operation, "FILE_ERROR", err.Error(), err).
		WithContext("path", path)

	// 根据错误类型添加特定的错误代码
	if os.IsNotExist(err) {
		sysErr.Code = "FILE_NOT_FOUND"
		sysErr.Message = "file or directory does not exist"
	} else if os.IsPermission(err) {
		sysErr.Code = "PERMISSION_DENIED"
		sysErr.Message = "permission denied"
	} else if os.IsExist(err) {
		sysErr.Code = "FILE_EXISTS"
		sysErr.Message = "file or directory already exists"
	} else if pathErr, ok := err.(*os.PathError); ok {
		sysErr.Code = "PATH_ERROR"
		sysErr.Message = pathErr.Err.Error()
		sysErr.WithContext("syscall", pathErr.Op)
	}

	return sysErr
}

// handleCommandError 处理命令执行错误
//
// handleCommandError 将命令执行错误转换为详细的SystemError。
//
// 参数：
// - command: 执行的命令
// - err: 原始错误
// - stderr: 标准错误输出
//
// 返回值：
// - error: 转换后的SystemError
func (so *SystemOperations) handleCommandError(command string, err error, stderr string) error {
	if err == nil {
		return nil
	}

	sysErr := NewSystemError("EXECUTE_COMMAND", "COMMAND_ERROR", err.Error(), err).
		WithContext("command", command)

	if stderr != "" {
		sysErr.WithContext("stderr", stderr)
	}

	// 根据错误类型添加特定信息
	if exitErr, ok := err.(*exec.ExitError); ok {
		sysErr.Code = "COMMAND_EXIT_ERROR"
		sysErr.Message = fmt.Sprintf("command exited with code %d", exitErr.ExitCode())
		sysErr.WithContext("exit_code", fmt.Sprintf("%d", exitErr.ExitCode()))
	} else if _, ok := err.(*exec.Error); ok {
		sysErr.Code = "COMMAND_NOT_FOUND"
		sysErr.Message = "command not found or not executable"
	}

	return sysErr
}

// SecurityConfig 安全配置
//
// SecurityConfig 定义了系统操作的安全策略，
// 包括允许的命令、路径限制和权限检查。
type SecurityConfig struct {
	AllowedCommands    []string          // 允许执行的命令白名单
	RestrictedPaths    []string          // 受限制的路径列表
	AllowedExtensions  []string          // 允许的文件扩展名
	MaxFileSize        int64             // 最大文件大小（字节）
	EnablePathSandbox  bool              // 启用路径沙盒
	SandboxRoot        string            // 沙盒根目录
	CustomValidators   map[string]func(string) error // 自定义验证器
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedCommands: []string{
			"git", "npm", "yarn", "go", "python", "pip", "node",
			"docker", "kubectl", "helm", "terraform", "ansible",
			"ls", "dir", "cat", "type", "echo", "pwd", "cd",
		},
		RestrictedPaths: []string{
			"/etc/passwd", "/etc/shadow", "/etc/hosts",
			"C:\\Windows\\System32", "C:\\Windows\\SysWOW64",
			"/System", "/usr/bin/sudo", "/bin/su",
		},
		AllowedExtensions: []string{
			".txt", ".md", ".json", ".yaml", ".yml", ".xml",
			".go", ".py", ".js", ".ts", ".html", ".css",
			".sh", ".bat", ".ps1", ".dockerfile",
		},
		MaxFileSize:       100 * 1024 * 1024, // 100MB
		EnablePathSandbox: false,
		SandboxRoot:       "",
		CustomValidators:  make(map[string]func(string) error),
	}
}

// SetSecurityConfig 设置安全配置
func (so *SystemOperations) SetSecurityConfig(config *SecurityConfig) {
	so.securityConfig = config
}

// GetSecurityConfig 获取当前安全配置
func (so *SystemOperations) GetSecurityConfig() *SecurityConfig {
	if so.securityConfig == nil {
		so.securityConfig = DefaultSecurityConfig()
	}
	return so.securityConfig
}

// validateCommandSecurity 验证命令安全性（增强版）
//
// validateCommandSecurity 使用安全配置验证命令是否被允许执行。
//
// 参数：
// - command: 要验证的命令
// - args: 命令参数
//
// 返回值：
// - error: 验证失败时的错误信息
func (so *SystemOperations) validateCommandSecurity(command string, args []string) error {
	config := so.GetSecurityConfig()
	
	// 基础验证
	if err := so.validateCommand(command, args); err != nil {
		return err
	}

	// 检查命令白名单
	if len(config.AllowedCommands) > 0 {
		allowed := false
		cmdName := filepath.Base(command)
		// 移除可执行文件扩展名
		if ext := filepath.Ext(cmdName); ext == ".exe" || ext == ".bat" || ext == ".cmd" {
			cmdName = strings.TrimSuffix(cmdName, ext)
		}
		
		for _, allowedCmd := range config.AllowedCommands {
			if cmdName == allowedCmd || command == allowedCmd {
				allowed = true
				break
			}
		}
		
		if !allowed {
			return NewSystemError("EXECUTE_COMMAND", "COMMAND_NOT_ALLOWED", 
				"command is not in the allowed list", nil).
				WithContext("command", command).WithContext("base_name", cmdName)
		}
	}

	// 自定义验证器
	if validator, exists := config.CustomValidators["command"]; exists {
		if err := validator(command); err != nil {
			return NewSystemError("EXECUTE_COMMAND", "CUSTOM_VALIDATION_FAILED", 
				err.Error(), err).WithContext("command", command)
		}
	}

	return nil
}

// validatePathSecurity 验证路径安全性（增强版）
//
// validatePathSecurity 使用安全配置验证路径访问权限。
//
// 参数：
// - path: 要验证的路径
// - operation: 操作类型
//
// 返回值：
// - error: 验证失败时的错误信息
func (so *SystemOperations) validatePathSecurity(path, operation string) error {
	config := so.GetSecurityConfig()
	
	// 基础验证
	if err := so.validatePath(path, operation); err != nil {
		return err
	}

	cleanPath := filepath.Clean(path)
	
	// 检查受限制路径
	for _, restrictedPath := range config.RestrictedPaths {
		if strings.HasPrefix(strings.ToLower(cleanPath), strings.ToLower(restrictedPath)) {
			return NewSystemError(operation, "PATH_RESTRICTED", 
				"access to this path is restricted by security policy", nil).
				WithContext("path", path).WithContext("restricted_path", restrictedPath)
		}
	}

	// 沙盒检查
	if config.EnablePathSandbox && config.SandboxRoot != "" {
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return NewSystemError(operation, "PATH_RESOLUTION_FAILED", 
				"failed to resolve absolute path", err).WithContext("path", path)
		}
		
		absSandbox, err := filepath.Abs(config.SandboxRoot)
		if err != nil {
			return NewSystemError(operation, "SANDBOX_RESOLUTION_FAILED", 
				"failed to resolve sandbox root", err).WithContext("sandbox_root", config.SandboxRoot)
		}
		
		if !strings.HasPrefix(absPath, absSandbox) {
			return NewSystemError(operation, "SANDBOX_VIOLATION", 
				"path is outside of sandbox", nil).
				WithContext("path", absPath).WithContext("sandbox_root", absSandbox)
		}
	}

	// 文件扩展名检查（仅对文件操作）
	if operation == "CREATE_FILE" || operation == "WRITE_FILE" {
		ext := strings.ToLower(filepath.Ext(cleanPath))
		if len(config.AllowedExtensions) > 0 && ext != "" {
			allowed := false
			for _, allowedExt := range config.AllowedExtensions {
				if ext == strings.ToLower(allowedExt) {
					allowed = true
					break
				}
			}
			if !allowed {
				return NewSystemError(operation, "EXTENSION_NOT_ALLOWED", 
					"file extension is not allowed", nil).
					WithContext("path", path).WithContext("extension", ext)
			}
		}
	}

	// 自定义验证器
	if validator, exists := config.CustomValidators["path"]; exists {
		if err := validator(path); err != nil {
			return NewSystemError(operation, "CUSTOM_VALIDATION_FAILED", 
				err.Error(), err).WithContext("path", path)
		}
	}

	return nil
}

// validateFileSize 验证文件大小
//
// validateFileSize 检查文件大小是否超过安全配置的限制。
//
// 参数：
// - path: 文件路径
// - size: 文件大小
//
// 返回值：
// - error: 验证失败时的错误信息
func (so *SystemOperations) validateFileSize(path string, size int64) error {
	config := so.GetSecurityConfig()
	
	if config.MaxFileSize > 0 && size > config.MaxFileSize {
		return NewSystemError("FILE_SIZE_CHECK", "FILE_TOO_LARGE", 
			"file size exceeds maximum allowed size", nil).
			WithContext("path", path).
			WithContext("size", fmt.Sprintf("%d", size)).
			WithContext("max_size", fmt.Sprintf("%d", config.MaxFileSize))
	}
	
	return nil
}

// CheckPermissions 检查文件或目录权限
//
// CheckPermissions 检查当前用户对指定路径的访问权限。
//
// 参数：
// - path: 要检查的路径
// - permission: 权限类型（"read", "write", "execute"）
//
// 返回值：
// - bool: 是否有权限
// - error: 检查过程中的错误
func (so *SystemOperations) CheckPermissions(path, permission string) (bool, error) {
	if err := so.validatePathSecurity(path, "CHECK_PERMISSIONS"); err != nil {
		return false, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return false, so.handleFileError("CHECK_PERMISSIONS", path, err)
	}

	mode := info.Mode()
	
	switch permission {
	case "read":
		// 检查读权限
		if runtime.GOOS == "windows" {
			// Windows上简单检查文件是否存在且可读
			file, err := os.Open(path)
			if err != nil {
				return false, nil
			}
			file.Close()
			return true, nil
		} else {
			// Unix系统检查权限位
			return mode&0400 != 0, nil
		}
	case "write":
		// 检查写权限
		if runtime.GOOS == "windows" {
			// Windows上尝试以写模式打开
			if info.IsDir() {
				// 对于目录，尝试创建临时文件
				tempFile := filepath.Join(path, ".temp_permission_check")
				file, err := os.Create(tempFile)
				if err != nil {
					return false, nil
				}
				file.Close()
				os.Remove(tempFile)
				return true, nil
			} else {
				// 对于文件，检查只读属性
				return mode&0200 != 0, nil
			}
		} else {
			// Unix系统检查权限位
			return mode&0200 != 0, nil
		}
	case "execute":
		// 检查执行权限
		if runtime.GOOS == "windows" {
			// Windows上检查文件扩展名
			ext := strings.ToLower(filepath.Ext(path))
			execExts := []string{".exe", ".bat", ".cmd", ".com", ".ps1"}
			for _, execExt := range execExts {
				if ext == execExt {
					return true, nil
				}
			}
			return false, nil
		} else {
			// Unix系统检查权限位
			return mode&0100 != 0, nil
		}
	default:
		return false, NewSystemError("CHECK_PERMISSIONS", "INVALID_PERMISSION", 
			"invalid permission type", nil).WithContext("permission", permission)
	}
}

// IsPathSafe 检查路径是否安全
//
// IsPathSafe 综合检查路径的安全性，包括路径遍历、
// 受限目录访问和沙盒限制。
//
// 参数：
// - path: 要检查的路径
//
// 返回值：
// - bool: 路径是否安全
// - error: 检查过程中的错误
func (so *SystemOperations) IsPathSafe(path string) (bool, error) {
	err := so.validatePathSecurity(path, "PATH_SAFETY_CHECK")
	if err != nil {
		return false, err
	}
	return true, nil
}

// IsCommandSafe 检查命令是否安全
//
// IsCommandSafe 综合检查命令的安全性，包括命令白名单、
// 危险字符和自定义验证。
//
// 参数：
// - command: 要检查的命令
// - args: 命令参数
//
// 返回值：
// - bool: 命令是否安全
// - error: 检查过程中的错误
func (so *SystemOperations) IsCommandSafe(command string, args []string) (bool, error) {
	err := so.validateCommandSecurity(command, args)
	return err == nil, err
}

// ExtractZipArchive 提取ZIP文件到目标目录
//
// ExtractZipArchive 提供了ZIP文件提取的系统级接口，集成了完整的
// ZIP处理功能，包括安全验证、进度跟踪和错误处理。
//
// 参数说明：
//   zipPath - ZIP文件的完整路径
//   targetDir - 目标提取目录
//   overwrite - 是否覆盖已存在的文件
//
// 返回值：
//   error - 提取过程中的错误，nil表示成功
//
// 使用示例：
//   err := sysOps.ExtractZipArchive("archive.zip", "./output", true)
//   if err != nil {
//       log.Printf("ZIP extraction failed: %v", err)
//   }
func (so *SystemOperations) ExtractZipArchive(zipPath, targetDir string, overwrite bool) error {
	// 创建ZIP处理器
	zipProcessor := NewZipProcessor(so)
	
	// 配置提取选项
	opts := &ExtractOptions{
		OverwriteExisting:   overwrite,
		PreservePermissions: true,
		FlattenStructure:    false,
		MaxFileSize:         100 * 1024 * 1024, // 100MB
		AllowedExtensions:   []string{}, // 允许所有扩展名
		SkipHidden:          false,
		Verbose:             false,
	}
	
	// 执行ZIP提取
	return zipProcessor.ExtractZip(zipPath, targetDir, opts)
}

// ValidateZipArchive 验证ZIP文件完整性
//
// ValidateZipArchive 检查ZIP文件的完整性和有效性，确保文件
// 可以正常读取和提取。
//
// 参数说明：
//   zipPath - ZIP文件的完整路径
//
// 返回值：
//   error - 验证过程中的错误，nil表示文件有效
//
// 使用示例：
//   if err := sysOps.ValidateZipArchive("archive.zip"); err != nil {
//       log.Printf("ZIP file is invalid: %v", err)
//   }
func (so *SystemOperations) ValidateZipArchive(zipPath string) error {
	// 创建ZIP处理器
	zipProcessor := NewZipProcessor(so)
	
	// 执行ZIP验证
	return zipProcessor.ValidateZip(zipPath)
}

// ListZipArchiveContents 列出ZIP文件内容
//
// ListZipArchiveContents 读取ZIP文件并返回其中包含的所有文件
// 和目录的列表，提供ZIP文件内容的快速预览功能。
//
// 参数说明：
//   zipPath - ZIP文件的完整路径
//
// 返回值：
//   []string - ZIP文件中的文件和目录列表
//   error - 读取过程中的错误，nil表示成功
//
// 使用示例：
//   contents, err := sysOps.ListZipArchiveContents("archive.zip")
//   if err != nil {
//       log.Printf("Failed to list ZIP contents: %v", err)
//   } else {
//       for _, item := range contents {
//           fmt.Println(item)
//       }
//   }
func (so *SystemOperations) ListZipArchiveContents(zipPath string) ([]string, error) {
	// 创建ZIP处理器
	zipProcessor := NewZipProcessor(so)
	
	// 执行ZIP内容列举
	return zipProcessor.ListZipContents(zipPath)
}