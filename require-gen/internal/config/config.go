package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"specify-cli/internal/types"
)

// AgentConfig AI助手配置映射
//
// AgentConfig 是一个全局配置映射，定义了require-gen框架支持的所有AI助手
// 及其相关配置信息。该映射表提供了统一的AI助手管理和配置接口。
//
// 支持的AI助手平台：
// - GitHub Copilot: GitHub官方AI编程助手
// - Claude Code: Anthropic的Claude AI助手
// - Gemini CLI: Google的Gemini AI助手
// - Cursor: Cursor AI编程助手
// - Qwen Code: 阿里巴巴的Qwen AI助手
// - opencode: opencode AI助手
// - Codex CLI: OpenAI的Codex模型
// - Windsurf: Windsurf AI编程助手
// - Kilo Code: Kilo Code AI助手
// - Auggie CLI: Auggie AI助手
// - CodeBuddy: CodeBuddy AI助手
// - Roo Code: Roo Code AI助手
// - Amazon Q Developer CLI: 亚马逊Q开发者CLI
//
// 配置信息包含：
// - Name: 助手的用户友好显示名称
// - Folder: 助手相关文件的存储目录
// - InstallURL: 官方安装或配置文档链接
// - RequiresCLI: 是否需要安装命令行工具
//
// 使用场景：
// - 用户选择AI助手时的选项列表
// - 项目初始化时的助手配置
// - 工具依赖检查和安装建议
// - 模板下载和配置管理
//
// 扩展性：
// 新的AI助手可以通过添加新的映射条目来支持，
// 无需修改核心业务逻辑代码。
var AgentConfig = map[string]types.AgentInfo{
	"copilot": {
		Name:        "GitHub Copilot",
		Folder:      ".github/",
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
		RequiresCLI: false,
	},
	"kilocode": {
		Name:        "Kilo Code",
		Folder:      ".kilocode/",
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
		RequiresCLI: false,
	},
	"q": {
		Name:        "Amazon Q Developer CLI",
		Folder:      ".amazonq/",
		InstallURL:  "https://aws.amazon.com/developer/learning/q-developer-cli/",
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
	Banner          = `███████╗██████╗ ███████╗ ██████╗██╗███████╗██╗   ██╗
██╔════╝██╔══██╗██╔════╝██╔════╝██║██╔════╝╚██╗ ██╔╝
███████╗██████╔╝█████╗  ██║     ██║█████╗   ╚████╔╝
╚════██║██╔═══╝ ██╔══╝  ██║     ██║██╔══╝    ╚██╔╝
███████║██║     ███████╗╚██████╗██║██║        ██║
╚══════╝╚═╝     ╚══════╝ ╚═════╝╚═╝╚═╝        ╚═╝`
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

// GetAllAgentsOrdered 获取按定义顺序排列的AI助手列表
func GetAllAgentsOrdered() []types.AgentOption {
	// 按照AgentConfig中定义的顺序返回AI助手列表
	// 确保与AgentConfig map中的定义顺序完全一致
	orderedKeys := []string{
		"copilot", "claude", "gemini", "cursor-agent", "qwen", "opencode",
		"codex", "windsurf", "kilocode", "auggie", "codebuddy", "roo", "q",
	}
	
	var agents []types.AgentOption
	for _, key := range orderedKeys {
		if info, exists := AgentConfig[key]; exists {
			agents = append(agents, types.AgentOption{
				Key:  key,
				Name: info.Name,
			})
		}
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
		case "claude":
			tools = append(tools, "claude")
		case "gemini":
			tools = append(tools, "gemini")
		case "qwen":
			tools = append(tools, "qwen")
		case "opencode":
			tools = append(tools, "opencode")
		case "codex":
			tools = append(tools, "codex")
		case "auggie":
			tools = append(tools, "auggie")
		case "codebuddy":
			tools = append(tools, "codebuddy")
		case "q":
			tools = append(tools, "q")
		}
	}

	return tools
}

// FileConfigManager 文件配置管理器
//
// FileConfigManager 实现了ConfigManager接口，提供基于文件的配置管理功能。
// 支持JSON格式的配置文件加载、保存、验证和管理。
//
// 功能特性：
// - JSON格式配置文件支持
// - 配置验证和错误处理
// - 默认配置生成
// - 配置合并和覆盖
// - 自动备份和恢复
//
// 使用示例：
//
//	manager := &FileConfigManager{}
//	config, err := manager.LoadConfig("project.json")
//	if err != nil {
//	    config = manager.GetDefaultConfig()
//	}
type FileConfigManager struct{}

// LoadConfig 从JSON文件加载配置
func (f *FileConfigManager) LoadConfig(filePath string) (*types.ProjectConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config types.ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := f.ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到JSON文件
func (f *FileConfigManager) SaveConfig(config *types.ProjectConfig, filePath string) error {
	// 验证配置
	if err := f.ValidateConfig(config); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 更新时间戳
	config.UpdatedAt = time.Now().Format(time.RFC3339)

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// ValidateConfig 验证配置有效性
func (f *FileConfigManager) ValidateConfig(config *types.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	if config.ProjectName == "" {
		return fmt.Errorf("项目名称不能为空")
	}

	if config.Version == "" {
		return fmt.Errorf("版本号不能为空")
	}

	// 验证AI助手
	if config.AIAssistant != "" {
		if _, exists := AgentConfig[config.AIAssistant]; !exists {
			return fmt.Errorf("不支持的AI助手: %s", config.AIAssistant)
		}
	}

	// 验证脚本类型
	if config.ScriptType != "" {
		if _, exists := ScriptTypeChoices[config.ScriptType]; !exists {
			return fmt.Errorf("不支持的脚本类型: %s", config.ScriptType)
		}
	}

	return nil
}

// GetDefaultConfig 获取默认配置
func (f *FileConfigManager) GetDefaultConfig() *types.ProjectConfig {
	now := time.Now().Format(time.RFC3339)

	return &types.ProjectConfig{
		ProjectName:    "new-project",
		Version:        "1.0.0",
		Description:    "A new require-gen project",
		AIAssistant:    "github-copilot",
		ScriptType:     GetDefaultScriptType(),
		GitEnabled:     true,
		Tools:          []string{"git"},
		CustomSettings: make(map[string]interface{}),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// MergeConfig 合并配置
func (f *FileConfigManager) MergeConfig(base, override *types.ProjectConfig) *types.ProjectConfig {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := *base // 复制基础配置

	// 覆盖非空字段
	if override.ProjectName != "" {
		merged.ProjectName = override.ProjectName
	}
	if override.Version != "" {
		merged.Version = override.Version
	}
	if override.Description != "" {
		merged.Description = override.Description
	}
	if override.AIAssistant != "" {
		merged.AIAssistant = override.AIAssistant
	}
	if override.ScriptType != "" {
		merged.ScriptType = override.ScriptType
	}

	// 合并工具列表
	if len(override.Tools) > 0 {
		toolSet := make(map[string]bool)
		for _, tool := range merged.Tools {
			toolSet[tool] = true
		}
		for _, tool := range override.Tools {
			if !toolSet[tool] {
				merged.Tools = append(merged.Tools, tool)
			}
		}
	}

	// 合并自定义设置
	if override.CustomSettings != nil {
		if merged.CustomSettings == nil {
			merged.CustomSettings = make(map[string]interface{})
		}
		for key, value := range override.CustomSettings {
			merged.CustomSettings[key] = value
		}
	}

	// 更新时间戳
	merged.UpdatedAt = time.Now().Format(time.RFC3339)

	return &merged
}

// GetConfigPath 获取默认配置文件路径
func GetConfigPath(projectDir string) string {
	return filepath.Join(projectDir, ".require-gen", "config.json")
}

// LoadProjectConfig 加载项目配置
func LoadProjectConfig(projectDir string) (*types.ProjectConfig, error) {
	manager := &FileConfigManager{}
	configPath := GetConfigPath(projectDir)

	config, err := manager.LoadConfig(configPath)
	if err != nil {
		// 如果配置文件不存在，返回默认配置
		if os.IsNotExist(err) {
			return manager.GetDefaultConfig(), nil
		}
		return nil, err
	}

	return config, nil
}

// SaveProjectConfig 保存项目配置
func SaveProjectConfig(config *types.ProjectConfig, projectDir string) error {
	manager := &FileConfigManager{}
	configPath := GetConfigPath(projectDir)
	return manager.SaveConfig(config, configPath)
}
