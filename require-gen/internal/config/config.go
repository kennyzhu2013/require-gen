package config

import (
	"os"
	"path/filepath"
	"runtime"

	"specify-cli/internal/types"
)

// AgentConfig AI助手配置映射
var AgentConfig = map[string]types.AgentInfo{
	"github-copilot": {
		Name:        "GitHub Copilot",
		Folder:      "copilot",
		RequiresCLI: false,
	},
	"claude-code": {
		Name:        "Claude Code",
		Folder:      "claude",
		InstallURL:  "https://claude.ai/download",
		RequiresCLI: true,
	},
	"gemini-cli": {
		Name:        "Gemini CLI",
		Folder:      "gemini",
		InstallURL:  "https://ai.google.dev/gemini-api/docs/cli",
		RequiresCLI: true,
	},
	"openai-cli": {
		Name:        "OpenAI CLI",
		Folder:      "openai",
		InstallURL:  "https://platform.openai.com/docs/api-reference",
		RequiresCLI: true,
	},
	"anthropic-cli": {
		Name:        "Anthropic CLI",
		Folder:      "anthropic",
		InstallURL:  "https://docs.anthropic.com/claude/reference/getting-started-with-the-api",
		RequiresCLI: true,
	},
	"azure-openai": {
		Name:        "Azure OpenAI",
		Folder:      "azure-openai",
		InstallURL:  "https://azure.microsoft.com/en-us/products/ai-services/openai-service",
		RequiresCLI: true,
	},
	"huggingface": {
		Name:        "Hugging Face",
		Folder:      "huggingface",
		InstallURL:  "https://huggingface.co/docs/huggingface_hub/guides/cli",
		RequiresCLI: true,
	},
	"ollama": {
		Name:        "Ollama",
		Folder:      "ollama",
		InstallURL:  "https://ollama.ai/download",
		RequiresCLI: true,
	},
	"local-llm": {
		Name:        "Local LLM",
		Folder:      "local-llm",
		RequiresCLI: false,
	},
	"custom": {
		Name:        "Custom Assistant",
		Folder:      "custom",
		RequiresCLI: false,
	},
	"mistral": {
		Name:        "Mistral AI",
		Folder:      "mistral",
		InstallURL:  "https://docs.mistral.ai/",
		RequiresCLI: true,
	},
	"cohere": {
		Name:        "Cohere",
		Folder:      "cohere",
		InstallURL:  "https://docs.cohere.com/docs/the-cohere-platform",
		RequiresCLI: true,
	},
	"perplexity": {
		Name:        "Perplexity AI",
		Folder:      "perplexity",
		InstallURL:  "https://docs.perplexity.ai/",
		RequiresCLI: true,
	},
}

// ScriptTypeChoices 脚本类型选择
var ScriptTypeChoices = map[string]types.ScriptType{
	"sh": {
		Extension:   ".sh",
		Description: "POSIX Shell (bash/zsh)",
	},
	"ps": {
		Extension:   ".ps1",
		Description: "PowerShell",
	},
}

// UI配置
var (
	ClaudeLocalPath = filepath.Join(os.Getenv("HOME"), ".claude", "local", "claude")
	Banner          = `
    ███████╗██████╗ ███████╗ ██████╗    ██╗  ██╗██╗████████╗
    ██╔════╝██╔══██╗██╔════╝██╔════╝    ██║ ██╔╝██║╚══██╔══╝
    ███████╗██████╔╝█████╗  ██║         █████╔╝ ██║   ██║   
    ╚════██║██╔═══╝ ██╔══╝  ██║         ██╔═██╗ ██║   ██║   
    ███████║██║     ███████╗╚██████╗    ██║  ██╗██║   ██║   
    ╚══════╝╚═╝     ╚══════╝ ╚═════╝    ╚═╝  ╚═╝╚═╝   ╚═╝   
    `
	Tagline = "GitHub Spec Kit - Spec-Driven Development Toolkit"
)

// GetDefaultScriptType 根据操作系统获取默认脚本类型
func GetDefaultScriptType() string {
	if runtime.GOOS == "windows" {
		return "ps"
	}
	return "sh"
}

// GetAgentInfo 获取AI助手信息
func GetAgentInfo(assistant string) (types.AgentInfo, bool) {
	info, exists := AgentConfig[assistant]
	return info, exists
}

// GetScriptType 获取脚本类型信息
func GetScriptType(scriptType string) (types.ScriptType, bool) {
	info, exists := ScriptTypeChoices[scriptType]
	return info, exists
}

// GetAllAgents 获取所有AI助手列表
func GetAllAgents() map[string]string {
	agents := make(map[string]string)
	for key, info := range AgentConfig {
		agents[key] = info.Name
	}
	return agents
}

// GetAllScriptTypes 获取所有脚本类型列表
func GetAllScriptTypes() map[string]string {
	scripts := make(map[string]string)
	for key, info := range ScriptTypeChoices {
		scripts[key] = info.Description
	}
	return scripts
}

// GetRequiredTools 根据AI助手获取所需工具列表
func GetRequiredTools(assistant string) []string {
	tools := []string{"git"} // 基础工具

	if info, exists := AgentConfig[assistant]; exists && info.RequiresCLI {
		switch assistant {
		case "claude-code":
			tools = append(tools, "claude")
		case "gemini-cli":
			tools = append(tools, "gemini")
		case "openai-cli":
			tools = append(tools, "openai")
		case "anthropic-cli":
			tools = append(tools, "anthropic")
		case "azure-openai":
			tools = append(tools, "az")
		case "huggingface":
			tools = append(tools, "huggingface-cli")
		case "ollama":
			tools = append(tools, "ollama")
		case "mistral":
			tools = append(tools, "mistral")
		case "cohere":
			tools = append(tools, "cohere")
		case "perplexity":
			tools = append(tools, "perplexity")
		}
	}

	return tools
}