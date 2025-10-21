# Go版本CLI缺失功能补充实现方案

## 概述

基于对Python版本specify-cli与Go版本require-gen的详细对比分析，本文档提供了Go版本中缺失功能的具体实现方案。这些补充实现将使Go版本的功能完整性达到与Python版本相当的水平，同时保持Go版本的架构优势。

## 1. 缺失功能清单

根据对比分析，Go版本缺失以下关键功能：

### 1.1 Init命令缺失的CLI标志
- `--force`: 强制覆盖现有项目目录
- `--no-git`: 跳过Git仓库初始化
- `--ignore-agent-tools`: 忽略AI助手工具的可用性检查
- `--skip-tls`: 跳过TLS证书验证

### 1.2 配置管理功能
- 动态配置文件生成
- 配置项验证机制
- 配置文件自动加载和保存

## 2. 实现方案详细设计

### 2.1 扩展InitOptions结构体

首先需要扩展 `types.InitOptions` 结构体以支持新的CLI标志：

```go
// internal/types/types.go
type InitOptions struct {
    // 现有字段
    ProjectName  string
    Here         bool
    AIAssistant  string
    ScriptType   string
    GitHubToken  string
    Verbose      bool
    Debug        bool
    
    // 新增字段
    Force           bool   // --force 标志
    NoGit           bool   // --no-git 标志
    IgnoreTools     bool   // --ignore-agent-tools 标志
    SkipTLS         bool   // --skip-tls 标志
}
```

### 2.2 更新Init命令定义

修改 `internal/cli/init.go` 文件，添加新的CLI标志：

```go
// internal/cli/init.go
func init() {
    // 现有标志...
    
    // 新增标志
    initCmd.Flags().BoolVar(&initOpts.Force, "force", false, 
        "Force overwrite existing project directory")
    initCmd.Flags().BoolVar(&initOpts.NoGit, "no-git", false, 
        "Skip Git repository initialization")
    initCmd.Flags().BoolVar(&initOpts.IgnoreTools, "ignore-agent-tools", false, 
        "Ignore AI assistant tool availability checks")
    initCmd.Flags().BoolVar(&initOpts.SkipTLS, "skip-tls", false, 
        "Skip TLS certificate verification")
}
```

### 2.3 实现--force标志功能

#### 2.3.1 目录存在检查和处理

在 `internal/business/init.go` 中的 `createProjectDirectory` 方法中添加强制覆盖逻辑：

```go
// internal/business/init.go
func (h *InitHandler) createProjectDirectory(tracker *ui.StepTracker, opts types.InitOptions) error {
    tracker.SetStepRunning("create_dir", "Creating project directory")

    var targetDir string
    if opts.Here {
        targetDir = "."
    } else {
        targetDir = opts.ProjectName
    }

    // 检查目录是否存在
    if _, err := os.Stat(targetDir); err == nil {
        // 目录存在
        if !opts.Force {
            tracker.SetStepError("create_dir", 
                fmt.Sprintf("Directory '%s' already exists. Use --force to overwrite", targetDir))
            return fmt.Errorf("directory '%s' already exists. Use --force to overwrite", targetDir)
        }
        
        // 使用--force标志，询问用户确认
        if !h.confirmOverwrite(targetDir) {
            tracker.SetStepError("create_dir", "Operation cancelled by user")
            return fmt.Errorf("operation cancelled by user")
        }
        
        // 备份现有目录
        if err := h.backupExistingDirectory(targetDir); err != nil {
            tracker.SetStepError("create_dir", 
                fmt.Sprintf("Failed to backup existing directory: %v", err))
            return fmt.Errorf("failed to backup existing directory: %w", err)
        }
        
        // 清空目录内容
        if err := h.clearDirectory(targetDir); err != nil {
            tracker.SetStepError("create_dir", 
                fmt.Sprintf("Failed to clear directory: %v", err))
            return fmt.Errorf("failed to clear directory: %w", err)
        }
    } else if !os.IsNotExist(err) {
        tracker.SetStepError("create_dir", 
            fmt.Sprintf("Failed to check directory: %v", err))
        return fmt.Errorf("failed to check directory: %w", err)
    }

    // 创建目录（如果不存在）
    if err := os.MkdirAll(targetDir, 0755); err != nil {
        tracker.SetStepError("create_dir", 
            fmt.Sprintf("Failed to create directory: %v", err))
        return fmt.Errorf("failed to create directory: %w", err)
    }

    tracker.SetStepDone("create_dir", 
        fmt.Sprintf("Project directory created: %s", targetDir))
    return nil
}

// 确认覆盖操作
func (h *InitHandler) confirmOverwrite(dir string) bool {
    return h.uiRenderer.ConfirmAction(
        fmt.Sprintf("Directory '%s' already exists. Do you want to overwrite it?", dir))
}

// 备份现有目录
func (h *InitHandler) backupExistingDirectory(dir string) error {
    timestamp := time.Now().Format("20060102_150405")
    backupDir := fmt.Sprintf("%s_backup_%s", dir, timestamp)
    
    return os.Rename(dir, backupDir)
}

// 清空目录内容
func (h *InitHandler) clearDirectory(dir string) error {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return err
    }
    
    for _, entry := range entries {
        path := filepath.Join(dir, entry.Name())
        if err := os.RemoveAll(path); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 2.4 实现--no-git标志功能

#### 2.4.1 条件性Git初始化

修改 `initializeGit` 方法以支持跳过Git初始化：

```go
// internal/business/init.go
func (h *InitHandler) initializeGit(tracker *ui.StepTracker, opts types.InitOptions) error {
    if opts.NoGit {
        tracker.SetStepSkipped("init_git", "Git initialization skipped (--no-git flag)")
        return nil
    }

    tracker.SetStepRunning("init_git", "Initializing Git repository")

    // 检查是否已经是Git仓库
    if h.gitOps.IsGitRepository(".") {
        tracker.SetStepDone("init_git", "Git repository already exists")
        return nil
    }

    // 检查Git是否可用
    if !h.toolChecker.CheckTool("git", tracker) {
        if opts.IgnoreTools {
            tracker.SetStepSkipped("init_git", "Git not available, but ignored due to --ignore-agent-tools")
            return nil
        }
        tracker.SetStepError("init_git", "Git is not available")
        return fmt.Errorf("git is not available")
    }

    // 初始化Git仓库
    if err := h.gitOps.InitRepository("."); err != nil {
        tracker.SetStepError("init_git", fmt.Sprintf("Failed to initialize Git: %v", err))
        return fmt.Errorf("failed to initialize Git repository: %w", err)
    }

    // 创建初始提交
    if err := h.createInitialCommit(); err != nil {
        ui.ShowWarning(fmt.Sprintf("Failed to create initial commit: %v", err))
        // 不返回错误，因为Git初始化已经成功
    }

    tracker.SetStepDone("init_git", "Git repository initialized successfully")
    return nil
}

// 添加步骤跳过状态支持
func (st *StepTracker) SetStepSkipped(id, message string) {
    st.mu.Lock()
    defer st.mu.Unlock()
    
    if step, exists := st.steps[id]; exists {
        step.Status = "skipped"
        step.Message = message
        step.UpdatedAt = time.Now()
    }
}
```

### 2.5 实现--ignore-agent-tools标志功能

#### 2.5.1 条件性工具检查

修改 `checkTools` 方法以支持忽略工具检查：

```go
// internal/business/init.go
func (h *InitHandler) checkTools(tracker *ui.StepTracker, opts types.InitOptions) error {
    if opts.IgnoreTools {
        tracker.SetStepSkipped("check_tools", "Tool checks skipped (--ignore-agent-tools flag)")
        return nil
    }

    tracker.SetStepRunning("check_tools", "Checking required tools")

    // 获取AI助手所需的工具列表
    agentInfo, exists := config.GetAgentInfo(opts.AIAssistant)
    if !exists {
        tracker.SetStepError("check_tools", fmt.Sprintf("Unknown AI assistant: %s", opts.AIAssistant))
        return fmt.Errorf("unknown AI assistant: %s", opts.AIAssistant)
    }

    requiredTools := config.GetRequiredTools(opts.AIAssistant)
    
    // 检查所有必需工具
    missingTools := []string{}
    for _, tool := range requiredTools {
        if !h.toolChecker.CheckTool(tool, tracker) {
            missingTools = append(missingTools, tool)
        }
    }

    if len(missingTools) > 0 {
        errorMsg := fmt.Sprintf("Missing required tools: %s", strings.Join(missingTools, ", "))
        tracker.SetStepError("check_tools", errorMsg)
        
        // 提供安装建议
        h.showToolInstallationSuggestions(missingTools)
        
        return fmt.Errorf("missing required tools: %s", strings.Join(missingTools, ", "))
    }

    tracker.SetStepDone("check_tools", "All required tools are available")
    return nil
}

// 显示工具安装建议
func (h *InitHandler) showToolInstallationSuggestions(missingTools []string) {
    ui.ShowInfo("Installation suggestions:")
    for _, tool := range missingTools {
        suggestion := h.getToolInstallationSuggestion(tool)
        ui.ShowInfo(fmt.Sprintf("  %s: %s", tool, suggestion))
    }
}

// 获取工具安装建议
func (h *InitHandler) getToolInstallationSuggestion(tool string) string {
    suggestions := map[string]string{
        "git":    "Install from https://git-scm.com/downloads",
        "node":   "Install from https://nodejs.org/",
        "python": "Install from https://python.org/downloads/",
        "docker": "Install from https://docker.com/get-started",
        "claude": "Install Claude CLI: npm install -g @anthropic-ai/claude-cli",
    }
    
    if suggestion, exists := suggestions[tool]; exists {
        return suggestion
    }
    return "Please install this tool manually"
}
```

### 2.6 实现--skip-tls标志功能

#### 2.6.1 TLS配置管理

首先扩展网络配置以支持TLS跳过：

```go
// internal/types/types.go
type NetworkConfig struct {
    TLS           *TLSConfig    `json:"tls"`
    ProxyURL      string        `json:"proxy_url"`
    Timeout       time.Duration `json:"timeout"`
    RetryCount    int           `json:"retry_count"`
    RetryWaitTime time.Duration `json:"retry_wait_time"`
    SkipTLS       bool          `json:"skip_tls"`  // 新增字段
}
```

#### 2.6.2 HTTP客户端配置

修改基础设施层的HTTP客户端以支持TLS跳过：

```go
// internal/infrastructure/http_client.go
type HTTPClient struct {
    client *http.Client
    config *types.HTTPClientConfig
}

func NewHTTPClient(skipTLS bool) *HTTPClient {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: skipTLS,
        },
    }
    
    client := &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
    
    return &HTTPClient{
        client: client,
        config: &types.HTTPClientConfig{},
    }
}

// 在模板下载中使用
func (tp *TemplateProvider) Download(opts types.DownloadOptions) (string, error) {
    httpClient := NewHTTPClient(opts.SkipTLS)
    // 使用配置的HTTP客户端进行下载
    // ...
}
```

#### 2.6.3 下载选项扩展

扩展 `DownloadOptions` 以传递TLS配置：

```go
// internal/types/types.go
type DownloadOptions struct {
    AIAssistant  string
    DownloadDir  string
    ScriptType   string
    Verbose      bool
    ShowProgress bool
    GitHubToken  string
    SkipTLS      bool  // 新增字段
}
```

### 2.7 增强配置管理系统

#### 2.7.1 配置文件自动生成

创建配置管理器的增强版本：

```go
// internal/config/manager.go
type EnhancedConfigManager struct {
    configPath string
    config     *types.ProjectConfig
}

func NewEnhancedConfigManager() *EnhancedConfigManager {
    return &EnhancedConfigManager{
        configPath: ".specify/config.json",
    }
}

// 自动生成配置文件
func (cm *EnhancedConfigManager) GenerateConfig(opts types.InitOptions) (*types.ProjectConfig, error) {
    config := &types.ProjectConfig{
        ProjectName: opts.ProjectName,
        Version:     "1.0.0",
        AIAssistant: opts.AIAssistant,
        ScriptType:  opts.ScriptType,
        GitEnabled:  !opts.NoGit,
        Tools:       cm.detectRequiredTools(opts.AIAssistant),
        CustomSettings: map[string]interface{}{
            "force_overwrite":     opts.Force,
            "ignore_tool_checks":  opts.IgnoreTools,
            "skip_tls_verify":     opts.SkipTLS,
        },
        CreatedAt: time.Now().Format(time.RFC3339),
        UpdatedAt: time.Now().Format(time.RFC3339),
    }
    
    return config, nil
}

// 保存配置到文件
func (cm *EnhancedConfigManager) SaveConfig(config *types.ProjectConfig) error {
    // 确保配置目录存在
    configDir := filepath.Dir(cm.configPath)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return fmt.Errorf("failed to create config directory: %w", err)
    }
    
    // 序列化配置
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    // 写入文件
    if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }
    
    return nil
}

// 加载配置文件
func (cm *EnhancedConfigManager) LoadConfig() (*types.ProjectConfig, error) {
    if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("config file not found: %s", cm.configPath)
    }
    
    data, err := os.ReadFile(cm.configPath)
    if err != nil {
        return fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config types.ProjectConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}

// 验证配置
func (cm *EnhancedConfigManager) ValidateConfig(config *types.ProjectConfig) error {
    if config.ProjectName == "" {
        return fmt.Errorf("project name is required")
    }
    
    if config.AIAssistant == "" {
        return fmt.Errorf("AI assistant is required")
    }
    
    // 验证AI助手是否支持
    if _, exists := GetAgentInfo(config.AIAssistant); !exists {
        return fmt.Errorf("unsupported AI assistant: %s", config.AIAssistant)
    }
    
    // 验证脚本类型
    if config.ScriptType != "" {
        if _, exists := GetScriptType(config.ScriptType); !exists {
            return fmt.Errorf("unsupported script type: %s", config.ScriptType)
        }
    }
    
    return nil
}

// 检测所需工具
func (cm *EnhancedConfigManager) detectRequiredTools(aiAssistant string) []string {
    baseTools := []string{"git"}
    
    agentTools := GetRequiredTools(aiAssistant)
    
    // 合并工具列表
    toolSet := make(map[string]bool)
    for _, tool := range baseTools {
        toolSet[tool] = true
    }
    for _, tool := range agentTools {
        toolSet[tool] = true
    }
    
    // 转换为切片
    tools := make([]string, 0, len(toolSet))
    for tool := range toolSet {
        tools = append(tools, tool)
    }
    
    return tools
}
```

#### 2.7.2 集成配置管理到初始化流程

修改 `configureProject` 方法以使用增强的配置管理：

```go
// internal/business/init.go
func (h *InitHandler) configureProject(tracker *ui.StepTracker, opts types.InitOptions) error {
    tracker.SetStepRunning("configure", "Configuring project settings")

    // 创建配置管理器
    configManager := config.NewEnhancedConfigManager()
    
    // 生成配置
    projectConfig, err := configManager.GenerateConfig(opts)
    if err != nil {
        tracker.SetStepError("configure", fmt.Sprintf("Failed to generate config: %v", err))
        return fmt.Errorf("failed to generate config: %w", err)
    }
    
    // 验证配置
    if err := configManager.ValidateConfig(projectConfig); err != nil {
        tracker.SetStepError("configure", fmt.Sprintf("Config validation failed: %v", err))
        return fmt.Errorf("config validation failed: %w", err)
    }
    
    // 保存配置
    if err := configManager.SaveConfig(projectConfig); err != nil {
        tracker.SetStepError("configure", fmt.Sprintf("Failed to save config: %v", err))
        return fmt.Errorf("failed to save config: %w", err)
    }
    
    // 创建其他配置文件
    if err := h.createAdditionalConfigFiles(opts); err != nil {
        ui.ShowWarning(fmt.Sprintf("Failed to create additional config files: %v", err))
        // 不返回错误，因为主配置已经成功
    }

    tracker.SetStepDone("configure", "Project configured successfully")
    return nil
}

// 创建其他配置文件
func (h *InitHandler) createAdditionalConfigFiles(opts types.InitOptions) error {
    // 创建.gitignore文件
    if !opts.NoGit {
        if err := h.createGitignoreFile(); err != nil {
            return fmt.Errorf("failed to create .gitignore: %w", err)
        }
    }
    
    // 创建AI助手特定的配置文件
    if err := h.createAIAssistantConfig(opts.AIAssistant); err != nil {
        return fmt.Errorf("failed to create AI assistant config: %w", err)
    }
    
    return nil
}

// 创建.gitignore文件
func (h *InitHandler) createGitignoreFile() error {
    gitignoreContent := `# Dependencies
node_modules/
*.log

# Build outputs
dist/
build/

# IDE files
.vscode/
.idea/

# OS files
.DS_Store
Thumbs.db

# Temporary files
*.tmp
*.temp

# Specify CLI files
.specify/cache/
.specify/temp/
`
    
    return os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)
}

// 创建AI助手配置
func (h *InitHandler) createAIAssistantConfig(aiAssistant string) error {
    switch aiAssistant {
    case "claude":
        return h.createClaudeConfig()
    case "github-copilot":
        return h.createCopilotConfig()
    case "gemini":
        return h.createGeminiConfig()
    default:
        return nil // 不需要特殊配置
    }
}

// 创建Claude配置
func (h *InitHandler) createClaudeConfig() error {
    claudeConfig := map[string]interface{}{
        "model": "claude-3-sonnet-20240229",
        "max_tokens": 4096,
        "temperature": 0.7,
    }
    
    data, err := json.MarshalIndent(claudeConfig, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(".claude/config.json", data, 0644)
}
```

## 3. UI增强实现

### 3.1 添加确认对话框支持

扩展UI渲染器以支持确认对话框：

```go
// internal/ui/ui.go
func (ui *UIManager) ConfirmAction(message string) bool {
    prompt := &survey.Confirm{
        Message: message,
        Default: false,
    }
    
    var confirmed bool
    err := survey.AskOne(prompt, &confirmed)
    if err != nil {
        return false
    }
    
    return confirmed
}

// 添加警告显示
func (ui *UIManager) ShowWarning(message string) {
    color.Yellow("⚠️  Warning: %s", message)
}

// 添加信息显示
func (ui *UIManager) ShowInfo(message string) {
    color.Cyan("ℹ️  Info: %s", message)
}
```

### 3.2 增强步骤跟踪器

添加跳过状态支持：

```go
// internal/ui/tracker.go
const (
    StepStatusPending   = "pending"
    StepStatusRunning   = "running"
    StepStatusDone      = "done"
    StepStatusError     = "error"
    StepStatusSkipped   = "skipped"  // 新增状态
)

// 显示步骤时处理跳过状态
func (st *StepTracker) displayStep(step *Step) {
    switch step.Status {
    case StepStatusPending:
        color.White("⏳ %s", step.Description)
    case StepStatusRunning:
        color.Yellow("🔄 %s - %s", step.Description, step.Message)
    case StepStatusDone:
        color.Green("✅ %s - %s", step.Description, step.Message)
    case StepStatusError:
        color.Red("❌ %s - %s", step.Description, step.Message)
    case StepStatusSkipped:
        color.Cyan("⏭️  %s - %s", step.Description, step.Message)  // 新增显示
    }
}
```

## 4. 测试实现

### 4.1 单元测试

为新功能添加单元测试：

```go
// internal/business/init_test.go
func TestInitHandler_ForceOverwrite(t *testing.T) {
    // 创建临时目录
    tempDir := t.TempDir()
    projectDir := filepath.Join(tempDir, "test-project")
    
    // 创建现有项目目录
    err := os.MkdirAll(projectDir, 0755)
    require.NoError(t, err)
    
    // 创建一些文件
    testFile := filepath.Join(projectDir, "test.txt")
    err = os.WriteFile(testFile, []byte("existing content"), 0644)
    require.NoError(t, err)
    
    // 测试不使用--force标志
    handler := NewInitHandler()
    opts := types.InitOptions{
        ProjectName: "test-project",
        AIAssistant: "claude",
        ScriptType:  "sh",
        Force:       false,
    }
    
    // 应该失败
    err = handler.Execute(opts)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already exists")
    
    // 测试使用--force标志
    opts.Force = true
    err = handler.Execute(opts)
    assert.NoError(t, err)
    
    // 验证备份目录是否创建
    backupDirs, _ := filepath.Glob(projectDir + "_backup_*")
    assert.NotEmpty(t, backupDirs)
}

func TestInitHandler_NoGit(t *testing.T) {
    tempDir := t.TempDir()
    os.Chdir(tempDir)
    
    handler := NewInitHandler()
    opts := types.InitOptions{
        ProjectName: "test-project",
        AIAssistant: "claude",
        ScriptType:  "sh",
        NoGit:       true,
    }
    
    err := handler.Execute(opts)
    assert.NoError(t, err)
    
    // 验证没有创建.git目录
    gitDir := filepath.Join("test-project", ".git")
    _, err = os.Stat(gitDir)
    assert.True(t, os.IsNotExist(err))
}

func TestInitHandler_IgnoreTools(t *testing.T) {
    handler := NewInitHandler()
    opts := types.InitOptions{
        ProjectName:  "test-project",
        AIAssistant:  "claude",
        ScriptType:   "sh",
        IgnoreTools:  true,
    }
    
    // 即使工具不可用，也应该成功
    err := handler.Execute(opts)
    assert.NoError(t, err)
}
```

### 4.2 集成测试

```go
// test/integration/cli_test.go
func TestCLI_InitWithAllFlags(t *testing.T) {
    tempDir := t.TempDir()
    os.Chdir(tempDir)
    
    // 创建现有项目目录
    projectDir := "test-project"
    err := os.MkdirAll(projectDir, 0755)
    require.NoError(t, err)
    
    // 执行CLI命令
    cmd := exec.Command("specify", "init", "test-project",
        "--ai", "claude",
        "--script", "sh",
        "--force",
        "--no-git",
        "--ignore-agent-tools",
        "--skip-tls")
    
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err, string(output))
    
    // 验证结果
    assert.Contains(t, string(output), "Project initialization completed successfully")
    
    // 验证配置文件
    configPath := filepath.Join(projectDir, ".specify", "config.json")
    assert.FileExists(t, configPath)
    
    // 验证没有Git仓库
    gitDir := filepath.Join(projectDir, ".git")
    _, err = os.Stat(gitDir)
    assert.True(t, os.IsNotExist(err))
}
```

## 5. 文档更新

### 5.1 CLI帮助文档

更新命令帮助信息：

```go
// internal/cli/init.go
var initCmd = &cobra.Command{
    Use:   "init [project-name]",
    Short: "Initialize a new spec-driven project",
    Long: `Initialize a new spec-driven development project with AI assistant integration.

This command creates a new project directory, downloads the appropriate template,
and configures the development environment based on your choices.

Examples:
  # Initialize a new project with Claude
  specify init my-project --ai claude --script sh

  # Initialize in current directory with force overwrite
  specify init --here --force --ai github-copilot

  # Initialize without Git and ignore tool checks
  specify init my-project --no-git --ignore-agent-tools

  # Initialize with TLS verification disabled
  specify init my-project --skip-tls --token YOUR_GITHUB_TOKEN`,
    Args: cobra.MaximumNArgs(1),
    RunE: runInit,
}
```

### 5.2 README更新

更新项目README以包含新功能：

```markdown
## CLI Options

### Init Command

```bash
specify init [project-name] [flags]
```

**Flags:**
- `--ai, -a`: AI assistant type (claude, github-copilot, gemini)
- `--script, -s`: Script type (sh, ps)
- `--here`: Initialize in current directory
- `--force`: Force overwrite existing project directory
- `--no-git`: Skip Git repository initialization
- `--ignore-agent-tools`: Ignore AI assistant tool availability checks
- `--skip-tls`: Skip TLS certificate verification
- `--token`: GitHub access token for template download
- `--verbose, -v`: Enable verbose output
- `--debug`: Enable debug mode

**Examples:**

```bash
# Basic initialization
specify init my-project --ai claude

# Force overwrite existing directory
specify init my-project --ai claude --force

# Initialize without Git
specify init my-project --ai claude --no-git

# Initialize in current directory, ignoring tool checks
specify init --here --ai github-copilot --ignore-agent-tools

# Initialize with TLS verification disabled
specify init my-project --ai claude --skip-tls
```
```

## 6. 实施计划

### 6.1 实施阶段

**阶段1: 核心CLI标志实现**
1. 扩展InitOptions结构体
2. 更新init命令定义
3. 实现--force标志功能
4. 实现--no-git标志功能

**阶段2: 高级功能实现**
1. 实现--ignore-agent-tools标志
2. 实现--skip-tls标志
3. 增强错误处理和用户反馈

**阶段3: 配置管理增强**
1. 实现配置文件自动生成
2. 添加配置验证机制
3. 创建配置管理器

**阶段4: UI和用户体验改进**
1. 添加确认对话框
2. 增强步骤跟踪器
3. 改进错误消息和帮助信息

**阶段5: 测试和文档**
1. 编写单元测试
2. 编写集成测试
3. 更新文档和帮助信息

### 6.2 优先级排序

**高优先级:**
- --force标志（用户强烈需求）
- --no-git标志（企业环境需求）
- 配置管理增强（架构完整性）

**中优先级:**
- --ignore-agent-tools标志（特殊环境需求）
- --skip-tls标志（网络环境需求）

**低优先级:**
- UI增强（用户体验改进）
- 高级错误处理（质量提升）

## 7. 风险评估和缓解

### 7.1 潜在风险

1. **向后兼容性**: 新增CLI标志可能影响现有脚本
2. **配置复杂性**: 过多选项可能增加用户困惑
3. **测试覆盖**: 新功能需要全面测试
4. **文档维护**: 需要同步更新所有相关文档

### 7.2 缓解策略

1. **向后兼容**: 所有新标志都是可选的，默认行为保持不变
2. **用户体验**: 提供清晰的帮助信息和示例
3. **测试策略**: 实施全面的单元测试和集成测试
4. **文档策略**: 建立文档更新检查清单

## 8. 总结

本实现方案提供了完整的解决方案来补充Go版本中缺失的CLI功能，主要包括：

### 8.1 核心改进
- ✅ 添加4个关键CLI标志（--force, --no-git, --ignore-agent-tools, --skip-tls）
- ✅ 增强配置管理系统
- ✅ 改进用户交互体验
- ✅ 完善错误处理机制

### 8.2 架构优势
- 保持Go版本的分层架构优势
- 增强类型安全和错误处理
- 提供更好的测试覆盖
- 改善代码可维护性

### 8.3 用户价值
- 功能完整性达到Python版本水平
- 更好的企业环境适应性
- 更灵活的配置选项
- 更友好的用户体验

通过实施这些改进，Go版本将在保持其性能和架构优势的同时，提供与Python版本相当的功能完整性，为用户提供更好的开发体验。