package types

import (
	"sync"
	"time"
)

// AgentInfo 定义AI助手的配置信息
type AgentInfo struct {
	Name        string `json:"name"`
	Folder      string `json:"folder"`
	InstallURL  string `json:"install_url,omitempty"`
	RequiresCLI bool   `json:"requires_cli"`
}

// ScriptType 定义脚本类型配置
type ScriptType struct {
	Extension   string `json:"extension"`
	Description string `json:"description"`
}

// InitOptions 定义初始化命令的选项
type InitOptions struct {
	ProjectName  string
	Here         bool
	AIAssistant  string
	ScriptType   string
	GitHubToken  string
	Verbose      bool
	Debug        bool
}

// DownloadOptions 定义模板下载选项
type DownloadOptions struct {
	AIAssistant  string
	DownloadDir  string
	ScriptType   string
	Verbose      bool
	ShowProgress bool
	GitHubToken  string
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

// CommandResult 命令执行结果
type CommandResult struct {
	Output   string
	Error    error
	ExitCode int
}

// TemplateProvider 模板提供者接口
type TemplateProvider interface {
	Download(opts DownloadOptions) (string, error)
	Validate(path string) error
}

// StepObserver 步骤观察者接口
type StepObserver interface {
	OnStepChanged(step *Step)
}

// AuthProvider 认证提供者接口
type AuthProvider interface {
	GetToken() string
	GetHeaders() map[string]string
}

// GitOperations Git操作接口
type GitOperations interface {
	IsRepo(path string) bool
	InitRepo(path string, quiet bool) (bool, error)
	AddAndCommit(path string, message string) error
}

// ToolChecker 工具检查器接口
type ToolChecker interface {
	CheckTool(tool string, tracker *StepTracker) bool
	CheckAllTools(tools []string, tracker *StepTracker) bool
}

// UIRenderer UI渲染器接口
type UIRenderer interface {
	ShowBanner()
	SelectWithArrows(options map[string]string, prompt, defaultKey string) (string, error)
	GetKey() (string, error)
}