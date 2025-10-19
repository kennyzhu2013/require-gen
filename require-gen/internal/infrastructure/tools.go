package infrastructure

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

// AgentConfig AI助手配置信息
type AgentConfig struct {
	Name        string // 显示名称
	Folder      string // 配置文件夹
	InstallURL  string // 安装URL
	RequiresCLI bool   // 是否需要CLI工具
}

// AgentConfigs AI助手配置映射
var AgentConfigs = map[string]AgentConfig{
	"copilot": {
		Name:        "GitHub Copilot",
		Folder:      ".github/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"claude": {
		Name:        "Claude Code",
		Folder:      ".claude/",
		InstallURL:  "https://docs.anthropic.com/en/docs/claude-code/setup",
		RequiresCLI: true,
	},
	"gemini": {
		Name:        "Gemini CLI",
		Folder:      ".gemini/",
		InstallURL:  "https://github.com/google-gemini/gemini-cli",
		RequiresCLI: true,
	},
	"cursor-agent": {
		Name:        "Cursor",
		Folder:      ".cursor/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"qwen": {
		Name:        "Qwen Code",
		Folder:      ".qwen/",
		InstallURL:  "https://github.com/QwenLM/qwen-code",
		RequiresCLI: true,
	},
	"opencode": {
		Name:        "opencode",
		Folder:      ".opencode/",
		InstallURL:  "https://opencode.ai",
		RequiresCLI: true,
	},
	"codex": {
		Name:        "Codex CLI",
		Folder:      ".codex/",
		InstallURL:  "https://github.com/openai/codex",
		RequiresCLI: true,
	},
	"windsurf": {
		Name:        "Windsurf",
		Folder:      ".windsurf/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"kilocode": {
		Name:        "Kilo Code",
		Folder:      ".kilocode/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"auggie": {
		Name:        "Auggie CLI",
		Folder:      ".augment/",
		InstallURL:  "https://docs.augmentcode.com/cli/setup-auggie/install-auggie-cli",
		RequiresCLI: true,
	},
	"codebuddy": {
		Name:        "CodeBuddy",
		Folder:      ".codebuddy/",
		InstallURL:  "https://www.codebuddy.ai",
		RequiresCLI: true,
	},
	"roo": {
		Name:        "Roo Code",
		Folder:      ".roo/",
		InstallURL:  "",
		RequiresCLI: false,
	},
	"q": {
		Name:        "Amazon Q Developer CLI",
		Folder:      ".amazonq/",
		InstallURL:  "https://aws.amazon.com/developer/learning/q-developer-cli/",
		RequiresCLI: true,
	},
}

// ToolChecker 工具检查器实现
//
// ToolChecker 是require-gen框架中负责开发工具检测和验证的核心组件。
// 它提供了全面的工具可用性检查、版本验证、安装建议等功能，确保
// 项目初始化过程中所需的外部工具都能正常工作。
//
// 主要功能特性：
// - 工具检测：检查工具是否在系统PATH中可用
// - 版本验证：验证工具版本和基本功能
// - 批量检查：支持多个工具的并行检查
// - 安装建议：提供平台特定的工具安装指导
// - 系统要求：检查系统环境和依赖条件
// - 进度跟踪：与UI组件集成，提供检查进度反馈
//
// 支持的工具类型：
// - 版本控制：Git、SVN等
// - 包管理器：npm、yarn、pip等
// - 构建工具：make、cmake、gradle等
// - 开发环境：Node.js、Python、Java等
// - AI工具：GitHub Copilot CLI、Claude CLI等
//
// 设计原则：
// - 非侵入性：仅检查不修改系统状态
// - 跨平台：支持Windows、Linux、macOS
// - 容错性：单个工具失败不影响整体检查
// - 用户友好：提供清晰的错误信息和解决建议
//
// 使用场景：
// - 项目初始化前的环境验证
// - 开发工具的可用性检查
// - 系统要求的自动化验证
// - 工具安装状态的诊断
type ToolChecker struct{}

// NewToolChecker 创建新的工具检查器实例
func NewToolChecker() types.ToolChecker {
	return &ToolChecker{}
}

// CheckTool 检查单个工具是否可用
func (tc *ToolChecker) CheckTool(tool string, tracker *types.StepTracker) bool {
	// 特殊处理Claude CLI - 检查migrate-installer后的特殊路径
	// 参考: https://github.com/github/spec-kit/issues/123
	// migrate-installer命令会从PATH中移除原始可执行文件
	// 并在~/.claude/local/claude创建别名
	if tool == "claude" {
		if tc.checkClaudeLocalPath() {
			if tracker != nil {
				tracker.SetStepDone("check_tools", "Claude CLI found in local path")
			}
			return true
		}
	}

	// 检查工具是否在PATH中
	_, err := exec.LookPath(tool)
	if err != nil {
		if tracker != nil {
			tracker.SetStepError("check_tools", fmt.Sprintf("Tool '%s' not found in PATH", tool))
		}
		return false
	}

	// 尝试执行工具的版本命令来验证
	if !tc.verifyTool(tool) {
		if tracker != nil {
			tracker.SetStepError("check_tools", fmt.Sprintf("Tool '%s' found but not working properly", tool))
		}
		return false
	}

	return true
}

// checkClaudeLocalPath 检查Claude CLI的本地路径
func (tc *ToolChecker) checkClaudeLocalPath() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	
	claudeLocalPath := filepath.Join(currentUser.HomeDir, ".claude", "local", "claude")
	
	// 在Windows上添加.exe扩展名
	if runtime.GOOS == "windows" {
		claudeLocalPath += ".exe"
	}
	
	if _, err := os.Stat(claudeLocalPath); err == nil {
		return true
	}
	
	return false
}

// CheckAllTools 检查所有必需的工具
func (tc *ToolChecker) CheckAllTools(tools []string, tracker *types.StepTracker) bool {
	allAvailable := true

	for _, tool := range tools {
		if !tc.CheckTool(tool, tracker) {
			allAvailable = false
			ui.ShowError(fmt.Sprintf("Required tool '%s' is not available", tool))

			// 提供安装建议
			if suggestion := tc.getInstallSuggestion(tool); suggestion != "" {
				ui.ShowInfo(fmt.Sprintf("Install suggestion: %s", suggestion))
			}
		} else {
			ui.ShowSuccess(fmt.Sprintf("Tool '%s' is available", tool))
		}
	}

	return allAvailable
}

// verifyTool 验证工具是否正常工作
func (tc *ToolChecker) verifyTool(tool string) bool {
	var cmd *exec.Cmd

	switch tool {
	case "git":
		cmd = exec.Command("git", "--version")
	case "node":
		cmd = exec.Command("node", "--version")
	case "npm":
		cmd = exec.Command("npm", "--version")
	case "python", "python3":
		cmd = exec.Command(tool, "--version")
	case "pip", "pip3":
		cmd = exec.Command(tool, "--version")
	case "docker":
		cmd = exec.Command("docker", "--version")
	case "kubectl":
		cmd = exec.Command("kubectl", "version", "--client")
	case "terraform":
		cmd = exec.Command("terraform", "--version")
	case "aws":
		cmd = exec.Command("aws", "--version")
	case "az":
		cmd = exec.Command("az", "--version")
	case "gcloud":
		cmd = exec.Command("gcloud", "--version")
	case "claude":
		cmd = exec.Command("claude", "--version")
	case "openai":
		cmd = exec.Command("openai", "--version")
	case "anthropic":
		cmd = exec.Command("anthropic", "--version")
	case "gemini":
		cmd = exec.Command("gemini", "--version")
	case "huggingface-cli":
		cmd = exec.Command("huggingface-cli", "--version")
	case "ollama":
		cmd = exec.Command("ollama", "--version")
	case "mistral":
		cmd = exec.Command("mistral", "--version")
	case "cohere":
		cmd = exec.Command("cohere", "--version")
	case "perplexity":
		cmd = exec.Command("perplexity", "--version")
	case "code":
		// Visual Studio Code
		cmd = exec.Command("code", "--version")
	case "code-insiders":
		// Visual Studio Code Insiders
		cmd = exec.Command("code-insiders", "--version")
	default:
		// 对于未知工具，尝试通用的版本命令
		cmd = exec.Command(tool, "--version")
	}

	err := cmd.Run()
	return err == nil
}

// getInstallSuggestion 获取工具安装建议
func (tc *ToolChecker) getInstallSuggestion(tool string) string {
	switch tool {
	case "git":
		if runtime.GOOS == "windows" {
			return "Download from https://git-scm.com/download/win"
		} else if runtime.GOOS == "darwin" {
			return "Install with: brew install git"
		} else {
			return "Install with your package manager: apt install git / yum install git"
		}

	case "node":
		return "Download from https://nodejs.org/ or use nvm"

	case "npm":
		return "Usually comes with Node.js. If missing, reinstall Node.js"

	case "python", "python3":
		return "Download from https://python.org/ or use pyenv"

	case "pip", "pip3":
		return "Usually comes with Python. If missing: python -m ensurepip --upgrade"

	case "docker":
		return "Download from https://docker.com/get-started"

	case "kubectl":
		return "Install from https://kubernetes.io/docs/tasks/tools/"

	case "terraform":
		return "Download from https://terraform.io/downloads"

	case "aws":
		return "Install AWS CLI: pip install awscli"

	case "az":
		return "Install Azure CLI: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"

	case "gcloud":
		return "Install Google Cloud SDK: https://cloud.google.com/sdk/docs/install"

	case "claude":
		return "Install Claude CLI: https://docs.anthropic.com/en/docs/claude-code/setup"

	case "openai":
		return "Install OpenAI CLI: pip install openai"

	case "anthropic":
		return "Install Anthropic CLI: pip install anthropic"

	case "gemini":
		return "Install Gemini CLI: https://github.com/google-gemini/gemini-cli"

	case "huggingface-cli":
		return "Install Hugging Face CLI: pip install huggingface_hub"

	case "ollama":
		return "Install Ollama: https://ollama.ai/download"

	case "mistral":
		return "Install Mistral CLI: pip install mistralai"

	case "cohere":
		return "Install Cohere CLI: pip install cohere"

	case "perplexity":
		return "Install Perplexity CLI: pip install perplexity-ai"

	case "code":
		if runtime.GOOS == "windows" {
			return "Download from https://code.visualstudio.com/download"
		} else if runtime.GOOS == "darwin" {
			return "Download from https://code.visualstudio.com/download or install with: brew install --cask visual-studio-code"
		} else {
			return "Install with snap: snap install code --classic, or download from https://code.visualstudio.com/download"
		}

	case "code-insiders":
		if runtime.GOOS == "windows" {
			return "Download from https://code.visualstudio.com/insiders/"
		} else if runtime.GOOS == "darwin" {
			return "Download from https://code.visualstudio.com/insiders/ or install with: brew install --cask visual-studio-code-insiders"
		} else {
			return "Install with snap: snap install code-insiders --classic, or download from https://code.visualstudio.com/insiders/"
		}

	default:
		return fmt.Sprintf("Please install '%s' according to its official documentation", tool)
	}
}

// GetToolVersion 获取工具版本
func (tc *ToolChecker) GetToolVersion(tool string) (string, error) {
	var cmd *exec.Cmd

	switch tool {
	case "git":
		cmd = exec.Command("git", "--version")
	case "node":
		cmd = exec.Command("node", "--version")
	case "npm":
		cmd = exec.Command("npm", "--version")
	case "python", "python3":
		cmd = exec.Command(tool, "--version")
	case "pip", "pip3":
		cmd = exec.Command(tool, "--version")
	default:
		cmd = exec.Command(tool, "--version")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version for %s: %w", tool, err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// ListAvailableTools 列出系统中可用的工具
func (tc *ToolChecker) ListAvailableTools(tools []string) map[string]string {
	available := make(map[string]string)

	for _, tool := range tools {
		if tc.CheckTool(tool, nil) {
			if version, err := tc.GetToolVersion(tool); err == nil {
				available[tool] = version
			} else {
				available[tool] = "unknown version"
			}
		}
	}

	return available
}

// CheckSystemRequirements 检查系统要求
func (tc *ToolChecker) CheckSystemRequirements() error {
	// 检查操作系统
	supportedOS := []string{"windows", "darwin", "linux"}
	currentOS := runtime.GOOS

	supported := false
	for _, os := range supportedOS {
		if os == currentOS {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported operating system: %s", currentOS)
	}

	// 检查架构
	supportedArch := []string{"amd64", "arm64"}
	currentArch := runtime.GOARCH

	supported = false
	for _, arch := range supportedArch {
		if arch == currentArch {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported architecture: %s", currentArch)
	}

	return nil
}

// GetSystemInfo 获取系统信息
func (tc *ToolChecker) GetSystemInfo() map[string]string {
	return map[string]string{
		"os":           runtime.GOOS,
		"architecture": runtime.GOARCH,
		"go_version":   runtime.Version(),
		"num_cpu":      fmt.Sprintf("%d", runtime.NumCPU()),
	}
}
