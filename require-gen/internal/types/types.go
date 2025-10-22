package types

import (
	"sync"
	"time"
)

// TLSConfig 定义SSL/TLS安全配置
type TLSConfig struct {
	InsecureSkipVerify bool   `json:"insecure_skip_verify"` // 是否跳过证书验证
	CertFile          string `json:"cert_file"`            // 客户端证书文件路径
	KeyFile           string `json:"key_file"`             // 客户端私钥文件路径
	CAFile            string `json:"ca_file"`              // CA证书文件路径
	ServerName        string `json:"server_name"`          // 服务器名称验证
}

// NetworkConfig 定义网络通信配置
type NetworkConfig struct {
	TLS           *TLSConfig    `json:"tls"`             // TLS配置
	ProxyURL      string        `json:"proxy_url"`       // 代理服务器URL
	Timeout       time.Duration `json:"timeout"`         // 请求超时时间
	RetryCount    int           `json:"retry_count"`     // 重试次数
	RetryWaitTime time.Duration `json:"retry_wait_time"` // 重试等待时间
}

// HTTPClientConfig 定义HTTP客户端配置
type HTTPClientConfig struct {
	Timeout           time.Duration     `json:"timeout"`             // 请求超时
	RetryCount        int               `json:"retry_count"`         // 重试次数
	RetryWaitTime     time.Duration     `json:"retry_wait_time"`     // 重试等待时间
	MaxRetryWaitTime  time.Duration     `json:"max_retry_wait_time"` // 最大重试等待时间
	FollowRedirects   bool              `json:"follow_redirects"`    // 是否跟随重定向
	MaxRedirects      int               `json:"max_redirects"`       // 最大重定向次数
	UserAgent         string            `json:"user_agent"`          // User-Agent
	Headers           map[string]string `json:"headers"`             // 自定义头部
	Cookies           map[string]string `json:"cookies"`             // Cookies
	KeepAlive         bool              `json:"keep_alive"`          // 是否保持连接
	MaxIdleConns      int               `json:"max_idle_conns"`      // 最大空闲连接数
	MaxConnsPerHost   int               `json:"max_conns_per_host"`  // 每个主机最大连接数
	IdleConnTimeout   time.Duration     `json:"idle_conn_timeout"`   // 空闲连接超时
}

// ProgressInfo 定义下载进度信息
type ProgressInfo struct {
	Downloaded int64         `json:"downloaded"` // 已下载字节数
	Total      int64         `json:"total"`      // 总字节数
	Percentage float64       `json:"percentage"` // 下载百分比
	Speed      float64       `json:"speed"`      // 下载速度 (bytes/sec)
	ETA        time.Duration `json:"eta"`        // 预计剩余时间
	StartTime  time.Time     `json:"start_time"` // 开始时间
	LastUpdate time.Time     `json:"last_update"` // 最后更新时间
}

// ProgressDisplay 定义进度显示接口
type ProgressDisplay interface {
	Start(total int64)
	Update(info *ProgressInfo)
	Finish()
	SetMessage(message string)
}

// NetworkErrorType 定义网络错误类型
type NetworkErrorType int

const (
	NetworkErrorTypeTimeout NetworkErrorType = iota
	NetworkErrorTypeConnection
	NetworkErrorTypeAuthentication
	NetworkErrorTypeNotFound
	NetworkErrorTypeServerError
	NetworkErrorTypeSSL
	NetworkErrorTypeProxy
	NetworkErrorTypeHTTPClient
	NetworkErrorTypeTemporary
	NetworkErrorTypeCircuitOpen
	NetworkErrorTypeRateLimited
	NetworkErrorTypeUnknown
	NetworkErrorConnectionRefused
	NetworkErrorDNSResolution
	NetworkErrorSSLHandshake
	NetworkErrorProxyError
	NetworkErrorTimeout
	NetworkErrorTemporary
	NetworkErrorCertificate
	NetworkErrorAuthentication
	NetworkErrorPermission
	NetworkErrorCircuitOpen
	NetworkErrorUnknown
	NetworkErrorHTTPServer
)

// NetworkError 定义网络错误
type NetworkError struct {
	Type          NetworkErrorType `json:"type"`
	Message       string           `json:"message"`
	Cause         error            `json:"-"`
	URL           string           `json:"url"`
	Status        int              `json:"status"`
	Host          string           `json:"host,omitempty"`
	Timestamp     time.Time        `json:"timestamp"`
	Retryable     bool             `json:"retryable"`
	RetryStrategy interface{}      `json:"retry_strategy,omitempty"`
	StatusCode    int              `json:"status_code,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

func (e *NetworkError) Error() string {
	return e.Message
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// IsRetryable 判断错误是否可重试
func (e *NetworkError) IsRetryable() bool {
	switch e.Type {
	case NetworkErrorTypeTimeout, NetworkErrorTypeConnection, NetworkErrorTypeServerError:
		return true
	default:
		return false
	}
}

// String 返回错误类型的字符串表示
func (t NetworkErrorType) String() string {
	switch t {
	case NetworkErrorTypeTimeout:
		return "TIMEOUT"
	case NetworkErrorTypeConnection:
		return "CONNECTION"
	case NetworkErrorTypeAuthentication:
		return "AUTHENTICATION"
	case NetworkErrorTypeNotFound:
		return "NOT_FOUND"
	case NetworkErrorTypeServerError:
		return "SERVER_ERROR"
	case NetworkErrorTypeSSL:
		return "SSL"
	case NetworkErrorTypeProxy:
		return "PROXY"
	default:
		return "UNKNOWN"
	}
}

// AgentInfo 定义AI助手的配置信息
//
// AgentInfo 结构体封装了AI助手的完整配置信息，用于管理不同AI平台
// 的集成参数和安装要求。该结构体支持多种AI助手的统一配置管理。
//
// 字段说明：
// - Name: AI助手的显示名称，用于用户界面展示
// - Folder: 助手相关文件的存储目录名称
// - InstallURL: 助手安装或配置的官方URL（可选）
// - RequiresCLI: 是否需要安装命令行工具
//
// 支持的AI助手类型：
// - GitHub Copilot: 需要CLI工具和GitHub账户
// - Claude: 需要Anthropic API密钥
// - Gemini: 需要Google AI API密钥
//
// JSON序列化：
// 该结构体支持JSON序列化，用于配置文件的读写操作。
// omitempty标签确保可选字段在为空时不会序列化。
//
// 使用示例：
//   copilot := AgentInfo{
//       Name:        "GitHub Copilot",
//       Folder:      "copilot",
//       InstallURL:  "https://github.com/features/copilot",
//       RequiresCLI: true,
// AgentInfo 定义AI助手信息
type AgentInfo struct {
	Name        string `json:"name"`
	Folder      string `json:"folder"`
	InstallURL  string `json:"install_url,omitempty"`
	RequiresCLI bool   `json:"requires_cli"`
}

// AgentOption 定义AI助手选项
type AgentOption struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ScriptType 定义脚本类型配置
type ScriptType struct {
	Extension   string `json:"extension"`
	Description string `json:"description"`
}

// InitOptions 定义初始化命令的选项
//
// InitOptions 结构体包含了项目初始化过程中所需的所有配置参数。
// 该结构体作为初始化命令的参数载体，支持灵活的项目创建配置。
//
// 字段说明：
// - ProjectName: 项目名称，用于目录创建和配置
// - Here: 是否在当前目录初始化项目（true）或创建新目录（false）
// - AIAssistant: 选择的AI助手类型（如"copilot", "claude", "gemini"）
// - ScriptType: 脚本类型选择（如"shell", "powershell"）
// - GitHubToken: GitHub访问令牌，用于私有仓库和API访问
// - Verbose: 是否启用详细输出模式
// - Debug: 是否启用调试模式，显示详细的诊断信息
//
// 初始化流程：
// 1. 验证项目名称和路径
// 2. 选择或确认AI助手类型
// 3. 选择或确认脚本类型
// 4. 检查必需的开发工具
// 5. 创建项目目录结构
// 6. 下载和配置项目模板
// 7. 初始化Git仓库
// 8. 应用项目特定配置
//
// 使用场景：
// - 命令行参数解析和验证
// - 交互式项目配置收集
// - 批量项目创建脚本
// - CI/CD自动化流程
//
// 使用示例：
//   opts := InitOptions{
//       ProjectName: "my-spec-project",
//       Here:        false,
//       AIAssistant: "copilot",
//       ScriptType:  "shell",
//       Verbose:     true,
//   }
type InitOptions struct {
	ProjectName     string
	Here            bool
	AIAssistant     string
	ScriptType      string
	GitHubToken     string
	Verbose         bool
	Debug           bool
	// 新增的CLI标志
	Force           bool   // --force 标志：强制覆盖现有项目目录
	NoGit           bool   // --no-git 标志：跳过Git仓库初始化
	IgnoreTools     bool   // --ignore-agent-tools 标志：忽略AI助手工具的可用性检查
	SkipTLS         bool   // --skip-tls 标志：跳过TLS证书验证
}

// DownloadOptions 下载选项配置
type DownloadOptions struct {
	AIAssistant     string                 `json:"ai_assistant"`     // AI助手类型
	DownloadDir     string                 `json:"download_dir"`     // 下载目录
	ScriptType      string                 `json:"script_type"`      // 脚本类型
	Verbose         bool                   `json:"verbose"`          // 详细输出
	ShowProgress    bool                   `json:"show_progress"`    // 显示进度
	GitHubToken     string                 `json:"github_token"`     // GitHub令牌
	SkipTLS         bool                   `json:"skip_tls"`         // 跳过TLS证书验证
	NetworkConfig   *NetworkConfig         `json:"network_config"`   // 网络配置
	HTTPConfig      *HTTPClientConfig      `json:"http_config"`      // HTTP客户端配置
	ChunkSize       int64                  `json:"chunk_size"`       // 分块大小
	EnableResume    bool                   `json:"enable_resume"`    // 启用断点续传
	ProgressDisplay ProgressDisplay        `json:"-"`                // 进度显示器（不序列化）
	ProgressCallback func(*ProgressInfo)   `json:"-"`                // 进度回调（不序列化）
	MaxConcurrent   int                    `json:"max_concurrent"`   // 最大并发下载数
	VerifyChecksum  bool                   `json:"verify_checksum"`  // 验证校验和
	Checksum        string                 `json:"checksum"`         // 预期校验和
	ChecksumType    string                 `json:"checksum_type"`    // 校验和类型（md5, sha1, sha256）
}

// GitHubRelease GitHub发布信息
type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset GitHub发布资源
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Step 步骤跟踪器中的单个步骤
type Step struct {
	Key     string
	Label   string
	Status  string
	Detail  string
}

// StepTracker 步骤跟踪器，用于显示进度
type StepTracker struct {
	Title       string
	Steps       map[string]*Step
	StatusOrder map[string]int
	mutex       sync.RWMutex
}

// 步骤状态常量
const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusError   = "error"
	StatusSkipped = "skipped"
)

// SetStepError 设置步骤为错误状态
func (st *StepTracker) SetStepError(key, detail string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if st.Steps == nil {
		st.Steps = make(map[string]*Step)
	}

	if step, exists := st.Steps[key]; exists {
		step.Status = StatusError
		step.Detail = detail
	} else {
		st.Steps[key] = &Step{
			Key:    key,
			Label:  key,
			Status: StatusError,
			Detail: detail,
		}
	}
}

// SetStepRunning 设置步骤为运行状态
func (st *StepTracker) SetStepRunning(key, detail string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if st.Steps == nil {
		st.Steps = make(map[string]*Step)
	}

	if step, exists := st.Steps[key]; exists {
		step.Status = StatusRunning
		step.Detail = detail
	} else {
		st.Steps[key] = &Step{
			Key:    key,
			Label:  key,
			Status: StatusRunning,
			Detail: detail,
		}
	}
}

// SetStepDone 设置步骤为完成状态
func (st *StepTracker) SetStepDone(key, detail string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if st.Steps == nil {
		st.Steps = make(map[string]*Step)
	}

	if step, exists := st.Steps[key]; exists {
		step.Status = StatusDone
		step.Detail = detail
	} else {
		st.Steps[key] = &Step{
			Key:    key,
			Label:  key,
			Status: StatusDone,
			Detail: detail,
		}
	}
}

// Config 应用程序配置
type Config struct {
	Agents      map[string]AgentInfo  `json:"agents"`
	ScriptTypes map[string]ScriptType `json:"script_types"`
	Defaults    DefaultConfig         `json:"defaults"`
}

// DefaultConfig 默认配置
type DefaultConfig struct {
	AIAssistant string `json:"ai_assistant"`
	ScriptType  string `json:"script_type"`
	Timeout     time.Duration `json:"timeout"`
}

// SystemInfo 系统信息
//
// SystemInfo 结构体封装了当前运行环境的完整系统信息，用于系统兼容性
// 检查、工具安装建议和环境诊断等功能。
//
// 字段说明：
// - OS: 操作系统类型（如"windows", "linux", "darwin"）
// - Arch: 系统架构（如"amd64", "arm64", "386"）
// - GoVersion: Go运行时版本信息
// - NumCPU: 逻辑CPU核心数（包括超线程）
// - Compiler: Go编译器类型（通常为"gc"）
//
// 使用场景：
// - 系统兼容性检查和工具安装建议
// - 环境诊断和问题排查
// - 性能优化和资源配置
// - 跨平台功能适配
//
// 使用示例：
//   info := SystemInfo{
//       OS:        "windows",
//       Arch:      "amd64", 
//       GoVersion: "go1.21.0",
//       NumCPU:    8,
//       Compiler:  "gc",
//   }
type SystemInfo struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
	NumCPU    int    `json:"num_cpu"`
	Compiler  string `json:"compiler"`
}

// CommandResult 命令执行结果
//
// CommandResult 结构体封装了系统命令执行的完整结果信息，包括输出、
// 错误状态、执行时间等详细信息。该结构体为命令执行提供统一的结果格式。
//
// 字段说明：
// - Command: 执行的完整命令字符串，用于日志和调试
// - Output: 命令的标准输出内容
// - Error: 错误信息或标准错误输出
// - ExitCode: 命令的退出状态码（0表示成功）
// - Success: 命令是否成功执行的布尔标志
// - WorkingDir: 命令执行时的工作目录
// - Duration: 命令执行耗时
//
// 使用场景：
// - 系统命令执行结果的统一处理
// - 错误诊断和日志记录
// - 性能监控和分析
// - 自动化脚本的状态检查
//
// 使用示例：
//   result, err := sysOps.ExecuteCommand("git", "status")
//   if result.Success {
//       fmt.Printf("命令执行成功，耗时: %v\n", result.Duration)
//       fmt.Println("输出:", result.Output)
//   } else {
//       fmt.Printf("命令失败 (退出码: %d): %s\n", result.ExitCode, result.Error)
//   }
type CommandResult struct {
	Command    string        `json:"command"`
	Output     string        `json:"output"`
	Error      string        `json:"error"`
	ExitCode   int           `json:"exit_code"`
	Success    bool          `json:"success"`
	WorkingDir string        `json:"working_dir"`
	Duration   time.Duration `json:"duration"`
}

// CommandOptions 命令执行选项
//
// CommandOptions 结构体定义了命令执行的各种配置选项，提供了灵活的
// 命令执行控制能力。该结构体支持超时控制、环境配置、输出处理等功能。
//
// 字段说明：
// - WorkingDir: 命令执行的工作目录
// - Timeout: 命令执行超时时间，0表示无超时限制
// - CaptureOutput: 是否捕获命令输出（合并stdout和stderr）
// - Shell: 是否使用Shell模式执行命令
// - Env: 额外的环境变量列表，格式为"KEY=VALUE"
// - Input: 提供给命令的标准输入内容
//
// Shell模式说明：
// - Windows: 使用PowerShell或cmd执行
// - Unix: 使用bash或sh执行
// - 适用于需要Shell特性的复杂命令
//
// 使用场景：
// - 需要特定工作目录的命令执行
// - 长时间运行命令的超时控制
// - 需要特殊环境变量的命令
// - 交互式命令的输入提供
// - Shell脚本和管道命令的执行
//
// 使用示例：
//   options := &CommandOptions{
//       WorkingDir: "/path/to/project",
//       Timeout:    5 * time.Minute,
//       Shell:      true,
//       Env:        []string{"DEBUG=1", "ENV=production"},
//   }
//   result, err := sysOps.ExecuteCommandWithOptions("npm", []string{"install"}, options)
type CommandOptions struct {
	WorkingDir    string        `json:"working_dir"`
	Timeout       time.Duration `json:"timeout"`
	CaptureOutput bool          `json:"capture_output"`
	Shell         bool          `json:"shell"`
	Env           []string      `json:"env"`
	Input         string        `json:"input"`
}

// SystemOperations 系统操作接口
//
// SystemOperations 定义了所有系统级操作的接口，包括：
// - 系统信息获取：操作系统、架构、系统详情
// - 命令执行：同步/异步执行、选项配置、错误处理
// - 文件系统操作：文件/目录的创建、删除、复制、移动
// - 路径管理：路径标准化、连接、分隔符处理
// - 环境变量：设置、获取、列举环境变量
// - 权限检查：文件权限、可执行性检查
// - 临时资源：临时文件和目录管理
// - 跨平台支持：平台检测、Shell类型识别
// - 安全特性：路径验证、命令安全检查
//
// 该接口设计遵循以下原则：
// - 跨平台兼容：自动处理不同操作系统的差异
// - 安全优先：内置安全验证和权限检查
// - 错误透明：提供详细的错误信息和上下文
// - 性能优化：高效的文件操作和命令执行
// - 可扩展性：支持自定义安全策略和验证器
type SystemOperations interface {
	// 系统信息
	GetOS() string
	GetArch() string
	GetSystemInfo() SystemInfo
	
	// 命令执行
	ExecuteCommand(name string, args ...string) (*CommandResult, error)
	ExecuteCommandInDir(dir, name string, args ...string) (*CommandResult, error)
	ExecuteCommandWithOptions(name string, args []string, options *CommandOptions) (*CommandResult, error)
	ExecuteCommandAsync(name string, args []string, options *CommandOptions) (<-chan *CommandResult, error)
	
	// 目录操作
	CreateDirectory(path string) error
	RemoveDirectory(path string) error
	DirectoryExists(path string) bool
	ListDirectory(path string) ([]string, error)
	
	// 文件操作
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	CopyFile(src, dst string) error
	MoveFile(src, dst string) error
	FileExists(path string) bool
	GetFileSize(path string) (int64, error)
	GetFileModTime(path string) (int64, error)
	
	// 路径操作
	GetCurrentDirectory() (string, error)
	ChangeDirectory(path string) error
	GetHomeDirectory() (string, error)
	GetTempDirectory() string
	GetPathSeparator() string
	NormalizePath(path string) string
	JoinPath(paths ...string) string
	
	// 环境变量
	SetEnvironmentVariable(key, value string) error
	GetEnvironmentVariable(key string) string
	GetAllEnvironmentVariables() []string
	
	// 可执行文件
	IsExecutable(path string) bool
	FindExecutable(name string) (string, error)
	CheckPermissions(path, permission string) (bool, error)
	
	// 临时资源
	CreateTempFile(pattern string) (string, error)
	CreateTempDirectory(pattern string) (string, error)
	
	// 跨平台支持
	GetExecutableExtension() string
	GetShellCommand() (string, []string)
	DetectShellType() string
	GetPlatformSpecificPath(pathType string) (string, error)
	IsWindowsSystem() bool
	IsUnixSystem() bool
	IsMacOSSystem() bool
	IsLinuxSystem() bool
	
	// 安全特性
	IsPathSafe(path string) (bool, error)
	IsCommandSafe(command string, args []string) (bool, error)
	
	// ZIP文件处理
	ExtractZipArchive(zipPath, targetDir string, overwrite bool) error
	ValidateZipArchive(zipPath string) error
	ListZipArchiveContents(zipPath string) ([]string, error)
}

// TemplateProvider 模板提供者接口
//
// TemplateProvider 接口定义了项目模板管理的标准方法，提供了
// 模板下载、验证和信息获取的完整功能。
//
// 核心功能：
// - 模板下载和解压
// - 模板结构验证
// - 模板信息获取
// - 模板列表管理
//
// 支持的模板源：
// - GitHub Releases: 官方模板仓库
// - 本地文件: 离线模板包
// - 自定义URL: 第三方模板源
//
// 模板类型：
// - AI助手模板: 针对不同AI助手的配置
// - 脚本类型模板: Shell/PowerShell脚本模板
// - 项目结构模板: 标准项目目录结构
//
// 使用示例：
//   provider := NewTemplateProvider()
//   path, err := provider.Download(opts)
//   err = provider.Validate(path)
//   info, err := provider.GetTemplateInfo(path)
type TemplateProvider interface {
	Download(opts DownloadOptions) (string, error)
	Validate(path string) error
	GetTemplateInfo(path string) (map[string]interface{}, error)
	ListTemplates(token string) ([]string, error)
}

// StepObserver 步骤观察者接口
type StepObserver interface {
	OnStepChanged(step *Step)
}

// AuthProvider 认证提供者接口
//
// AuthProvider 接口定义了认证和授权相关的核心功能，主要用于GitHub API
// 的认证处理。该接口支持多种认证方式和令牌管理功能。
//
// 主要功能：
// - 令牌管理：获取、设置、清除认证令牌
// - 认证头生成：为HTTP请求生成认证头部
// - 认证状态检查：验证当前认证状态
// - 令牌验证：检查令牌的有效性和权限
// - 文件操作：从文件加载和保存令牌
// - 用户信息：获取认证用户的详细信息
//
// 使用场景：
// - GitHub API调用的认证处理
// - 用户身份验证和授权
// - 令牌的安全存储和管理
// - 多种认证方式的统一接口
//
// 实现要求：
// - 支持环境变量和CLI参数的令牌获取
// - 提供安全的令牌存储机制
// - 实现令牌有效性验证
// - 支持不同的认证类型（如Bearer Token）
type AuthProvider interface {
	// GetToken 获取当前认证令牌
	GetToken() string
	
	// GetHeaders 获取HTTP认证头部
	GetHeaders() map[string]string
	
	// GetAuthHeaders 获取GitHub API认证头部（兼容方法）
	GetAuthHeaders() map[string]string
	
	// SetToken 设置认证令牌
	SetToken(token string)
	
	// SetCLIToken 设置CLI参数令牌（最高优先级）
	SetCLIToken(token string)
	
	// IsAuthenticated 检查是否已认证
	IsAuthenticated() bool
	
	// ValidateToken 验证令牌有效性
	ValidateToken() error
	
	// GetAuthType 获取认证类型
	GetAuthType() string
	
	// ClearToken 清除认证令牌
	ClearToken()
	
	// ClearCLIToken 仅清除CLI令牌
	ClearCLIToken()
	
	// LoadFromFile 从文件加载令牌
	LoadFromFile(filePath string) error
	
	// SaveToFile 保存令牌到文件
	SaveToFile(filePath string) error
	
	// GetTokenScopes 获取令牌权限范围
	GetTokenScopes() ([]string, error)
	
	// GetUserInfo 获取认证用户信息
	GetUserInfo() (map[string]interface{}, error)
	
	// IsTokenExpired 检查令牌是否过期
	IsTokenExpired() (bool, error)
	
	// GetTokenSource 获取当前令牌来源的字符串描述
	GetTokenSource() string
}

// GitOperations Git操作接口
//
// GitOperations 接口定义了Git版本控制系统的标准操作方法，提供了
// 完整的Git仓库管理功能，支持项目初始化、版本控制和远程协作。
//
// 核心功能：
// - 仓库初始化和检查
// - 文件添加和提交
// - 分支管理和切换
// - 远程仓库操作
// - 状态查询和历史管理
//
// 支持的操作：
// - 本地仓库：初始化、状态检查、提交管理
// - 分支操作：创建、切换、合并
// - 远程协作：克隆、推送、拉取
// - 状态查询：工作目录状态、提交历史
//
// 使用示例：
//   git := NewGitOperations()
//   git.InitRepo("/path/to/project", false)
//   git.AddAndCommit("/path/to/project", "Initial commit")
//   git.AddRemote("/path/to/project", "origin", "https://github.com/user/repo.git")
type GitOperations interface {
	IsRepo(path string) bool
	InitRepo(path string, quiet bool) (bool, error)
	AddAndCommit(path string, message string) error
	GetStatus(path string) (string, error)
	GetBranch(path string) (string, error)
	CreateBranch(path, branchName string) error
	SwitchBranch(path, branchName string) error
	AddRemote(path, name, url string) error
	Push(path, remote, branch string) error
	Pull(path, remote, branch string) error
	Clone(url, targetPath string) error
	GetCommitHash(path string) (string, error)
	GetRemoteURL(path, remote string) (string, error)
	IsClean(path string) (bool, error)
	HasUncommittedChanges(path string) (bool, error)
}

// ToolChecker 工具检查器接口
//
// ToolChecker 接口定义了开发工具检测和验证的标准方法，提供了
// 全面的工具可用性检查、版本验证和安装建议功能。
//
// 核心功能：
// - 工具可用性检测
// - 版本信息获取
// - 批量工具检查
// - 系统要求验证
// - 安装建议提供
//
// 支持的工具类型：
// - 开发工具：Git, Node.js, Python, Docker
// - AI工具：Claude CLI, OpenAI CLI, Gemini CLI
// - 构建工具：Make, CMake, Gradle, Maven
// - 系统工具：curl, wget, tar, zip
//
// 使用示例：
//   checker := NewToolChecker()
//   tools := []string{"git", "node", "python"}
//   allAvailable := checker.CheckAllTools(tools, tracker)
//   versions := checker.ListAvailableTools(tools)
type ToolChecker interface {
	CheckTool(tool string, tracker *StepTracker) bool
	CheckAllTools(tools []string, tracker *StepTracker) bool
	GetToolVersion(tool string) (string, error)
	ListAvailableTools(tools []string) map[string]string
	CheckSystemRequirements() error
	GetSystemInfo() map[string]string
}

// UIRenderer UI渲染器接口
//
// UIRenderer 定义了用户界面渲染和交互的标准接口。
// 该接口提供了统一的UI操作方法，支持多种UI框架和渲染方式。
//
// 核心功能：
// - 进度显示和状态更新
// - 用户输入和选择处理
// - 消息显示和通知
// - 界面布局和样式管理
//
// 支持的UI类型：
// - 命令行界面（CLI）
// - 终端用户界面（TUI）
// - Web界面
// - 图形用户界面（GUI）
// - 移动应用界面
//
// 使用示例：
//   renderer := &CLIRenderer{}
//   renderer.ShowProgress("正在下载模板...", 50)
//   choice := renderer.SelectOption("选择AI助手", options)
//   renderer.ShowMessage("操作完成", "success")
type UIRenderer interface {
	ShowBanner()
	SelectWithArrows(options map[string]string, prompt, defaultKey string) (string, error)
	SelectWithArrowsOrdered(options []AgentOption, prompt, defaultKey string) (string, error)
	GetKey() (string, error)
	// ShowProgress 显示进度信息
	ShowProgress(message string, percentage int)
	
	// ShowMessage 显示消息
	ShowMessage(message, messageType string)
	
	// SelectOption 显示选择选项
	SelectOption(prompt string, options []string) (string, error)
	
	// ConfirmAction 确认操作
	ConfirmAction(message string) bool
}

// ConfigManager 配置管理器接口
//
// ConfigManager 定义了项目配置管理的标准接口。
// 该接口提供了配置文件的加载、保存、验证和管理功能。
//
// 核心功能：
// - 配置文件加载和解析
// - 配置数据保存和持久化
// - 配置项验证和校验
// - 默认配置管理
// - 配置合并和覆盖
//
// 支持的配置格式：
// - JSON格式配置文件
// - YAML格式配置文件
// - TOML格式配置文件
// - 环境变量配置
// - 命令行参数配置
//
// 使用示例：
//   manager := &FileConfigManager{}
//   config, err := manager.LoadConfig("config.json")
//   if err != nil {
//       config = manager.GetDefaultConfig()
//   }
//   manager.SaveConfig(config, "config.json")
type ConfigManager interface {
	// LoadConfig 从文件加载配置
	LoadConfig(filePath string) (*ProjectConfig, error)
	
	// SaveConfig 保存配置到文件
	SaveConfig(config *ProjectConfig, filePath string) error
	
	// ValidateConfig 验证配置有效性
	ValidateConfig(config *ProjectConfig) error
	
	// GetDefaultConfig 获取默认配置
	GetDefaultConfig() *ProjectConfig
	
	// MergeConfig 合并配置
	MergeConfig(base, override *ProjectConfig) *ProjectConfig
}

// ProjectConfig 项目配置结构体
//
// ProjectConfig 定义了require-gen项目的完整配置信息。
// 该结构体包含了项目初始化、构建、部署等各个阶段的配置参数。
//
// 配置分类：
// - 基础信息：项目名称、版本、描述等
// - AI助手配置：选择的AI助手及其参数
// - 脚本配置：脚本类型和执行参数
// - 工具配置：依赖工具和版本要求
// - 构建配置：构建选项和输出设置
// - 部署配置：部署目标和参数
//
// 字段说明：
// - ProjectName: 项目名称，用于标识项目
// - Version: 项目版本号，遵循语义化版本规范
// - Description: 项目描述信息
// - AIAssistant: 选择的AI助手标识符
// - ScriptType: 脚本类型（sh/ps）
// - GitEnabled: 是否启用Git版本控制
// - Tools: 项目依赖的工具列表
// - CustomSettings: 自定义配置项
//
// 使用示例：
//   config := &ProjectConfig{
//       ProjectName: "my-project",
//       Version: "1.0.0",
//       AIAssistant: "claude-code",
//       ScriptType: "sh",
//       GitEnabled: true,
//   }
type ProjectConfig struct {
	// 基础信息
	ProjectName string `json:"project_name" yaml:"project_name"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
	
	// AI助手配置
	AIAssistant string `json:"ai_assistant" yaml:"ai_assistant"`
	
	// 脚本配置
	ScriptType string `json:"script_type" yaml:"script_type"`
	
	// Git配置
	GitEnabled bool `json:"git_enabled" yaml:"git_enabled"`
	
	// 工具配置
	Tools []string `json:"tools" yaml:"tools"`
	
	// 自定义配置
	CustomSettings map[string]interface{} `json:"custom_settings" yaml:"custom_settings"`
	
	// 创建时间
	CreatedAt string `json:"created_at" yaml:"created_at"`
	
	// 最后更新时间
	UpdatedAt string `json:"updated_at" yaml:"updated_at"`
}