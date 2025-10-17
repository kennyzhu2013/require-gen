# Specify CLI Go版本详细设计文档

## 文档信息
- **版本**: 1.0.0
- **创建日期**: 2024年12月
- **基于**: OutlineDesign.md v1.0
- **目标**: 提供Specify CLI Go版本的完整技术实现方案

## 目录

1. [项目概述](#1-项目概述)
2. [系统架构详细设计](#2-系统架构详细设计)
3. [CLI层详细设计](#3-cli层详细设计)
4. [服务层详细设计](#4-服务层详细设计)
5. [UI组件详细设计](#5-ui组件详细设计)
6. [GitHub集成详细设计](#6-github集成详细设计)
7. [系统集成详细设计](#7-系统集成详细设计)
8. [基础设施层详细设计](#8-基础设施层详细设计)
9. [数据模型详细设计](#9-数据模型详细设计)
10. [错误处理详细设计](#10-错误处理详细设计)
11. [并发和性能优化详细设计](#11-并发和性能优化详细设计)
12. [测试策略详细设计](#12-测试策略详细设计)
13. [构建和部署详细设计](#13-构建和部署详细设计)
14. [API接口规范](#14-api接口规范)
15. [配置文件规范](#15-配置文件规范)
16. [实现路线图](#16-实现路线图)

---

## 1. 项目概述

### 1.1 设计目标
本详细设计文档基于概要设计文档，提供Specify CLI Go版本的完整技术实现方案。目标是：

1. **功能完整性**: 确保与Python版本功能完全对等
2. **架构清晰性**: 提供清晰的模块划分和接口定义
3. **实现可行性**: 提供具体的代码实现指导
4. **质量保证**: 确保代码质量和系统稳定性
5. **性能优化**: 充分利用Go语言的性能优势

### 1.2 技术栈选择

#### 1.2.1 核心依赖
```go
// 核心框架
"github.com/spf13/cobra"           // CLI框架
"github.com/spf13/viper"           // 配置管理
"go.uber.org/fx"                   // 依赖注入框架

// UI和交互
"github.com/charmbracelet/lipgloss" // 样式系统
"github.com/charmbracelet/bubbles"  // UI组件
"github.com/pterm/pterm"            // 终端UI库

// HTTP和网络
"github.com/go-resty/resty/v2"      // HTTP客户端
"golang.org/x/net/context"         // 上下文管理

// 文件和归档
"github.com/mholt/archiver/v4"      // 归档处理
"github.com/otiai10/copy"           // 文件复制

// 系统集成
"github.com/shirou/gopsutil/v3"     // 系统信息
"golang.org/x/sys"                  // 系统调用

// 测试
"github.com/stretchr/testify"       // 测试框架
"github.com/golang/mock"            // Mock生成
```

### 1.3 设计原则

1. **接口驱动开发**: 所有组件基于接口设计，便于测试和扩展
2. **依赖注入**: 使用DI容器管理组件生命周期
3. **错误处理**: 明确的错误类型和处理策略
4. **并发安全**: 充分利用Go的并发特性
5. **可测试性**: 每个组件都可独立测试

---

## 2. 系统架构详细设计

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                            │
├─────────────────┬─────────────────┬─────────────────────────┤
│   RootCommand   │   InitCommand   │     CheckCommand        │
├─────────────────┴─────────────────┴─────────────────────────┤
│                     Service Layer                           │
├─────────────────┬─────────────────┬─────────────────────────┤
│  InitService    │  CheckService   │   TemplateService       │
├─────────────────┴─────────────────┴─────────────────────────┤
│                    Component Layer                          │
├─────────────────┬─────────────────┬─────────────────────────┤
│  UI Components  │ GitHub Integration │  System Integration   │
├─────────────────┴─────────────────┴─────────────────────────┤
│                 Infrastructure Layer                        │
├─────────────────┬─────────────────┬─────────────────────────┤
│ Config Manager  │  File Manager   │   Permission Manager    │
└─────────────────┴─────────────────┴─────────────────────────┘
```

### 2.2 依赖注入架构

#### 2.2.1 DI容器设计
```go
// internal/di/container.go
package di

import (
    "go.uber.org/fx"
    "specify-cli-go/internal/cli/commands"
    "specify-cli-go/internal/core/services"
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/infrastructure/config"
)

// Container 依赖注入容器
type Container struct {
    app *fx.App
}

// NewContainer 创建新的DI容器
func NewContainer() *Container {
    app := fx.New(
        // 基础设施层
        fx.Provide(config.NewConfigManager),
        fx.Provide(NewLogger),
        fx.Provide(NewHTTPClient),
        
        // 服务层
        fx.Provide(services.NewInitService),
        fx.Provide(services.NewCheckService),
        fx.Provide(services.NewTemplateService),
        
        // UI组件
        fx.Provide(ui.NewStepTracker),
        fx.Provide(ui.NewSelector),
        fx.Provide(ui.NewBanner),
        
        // CLI命令
        fx.Provide(commands.NewRootCommand),
        fx.Provide(commands.NewInitCommand),
        fx.Provide(commands.NewCheckCommand),
        
        // 应用程序
        fx.Provide(NewApplication),
        
        // 生命周期管理
        fx.Invoke(RegisterHooks),
    )
    
    return &Container{app: app}
}

// Start 启动容器
func (c *Container) Start(ctx context.Context) error {
    return c.app.Start(ctx)
}

// Stop 停止容器
func (c *Container) Stop(ctx context.Context) error {
    return c.app.Stop(ctx)
}
```

### 2.3 模块间通信机制

#### 2.3.1 事件系统设计
```go
// internal/events/event_bus.go
package events

import (
    "context"
    "sync"
)

// EventType 事件类型
type EventType string

const (
    EventProjectInitStarted   EventType = "project.init.started"
    EventProjectInitCompleted EventType = "project.init.completed"
    EventToolCheckStarted     EventType = "tool.check.started"
    EventToolCheckCompleted   EventType = "tool.check.completed"
    EventDownloadStarted      EventType = "download.started"
    EventDownloadProgress     EventType = "download.progress"
    EventDownloadCompleted    EventType = "download.completed"
)

// Event 事件接口
type Event interface {
    Type() EventType
    Data() interface{}
    Timestamp() time.Time
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event Event) error

// EventBus 事件总线
type EventBus struct {
    handlers map[EventType][]EventHandler
    mutex    sync.RWMutex
}

// NewEventBus 创建新的事件总线
func NewEventBus() *EventBus {
    return &EventBus{
        handlers: make(map[EventType][]EventHandler),
    }
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
    eb.mutex.Lock()
    defer eb.mutex.Unlock()
    
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish 发布事件
func (eb *EventBus) Publish(ctx context.Context, event Event) error {
    eb.mutex.RLock()
    handlers := eb.handlers[event.Type()]
    eb.mutex.RUnlock()
    
    for _, handler := range handlers {
        if err := handler(ctx, event); err != nil {
            return err
        }
    }
    
    return nil
}
```

---

## 3. CLI层详细设计

### 3.1 命令结构设计

#### 3.1.1 根命令实现
```go
// internal/cli/commands/root.go
package commands

import (
    "context"
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/infrastructure/config"
    "specify-cli-go/pkg/version"
)

// RootCommand 根命令结构
type RootCommand struct {
    cmd           *cobra.Command
    config        *config.Manager
    banner        *ui.Banner
    globalFlags   *GlobalFlags
    eventBus      *events.EventBus
}

// GlobalFlags 全局标志
type GlobalFlags struct {
    Debug     bool
    Verbose   bool
    NoColor   bool
    ConfigDir string
}

// NewRootCommand 创建根命令
func NewRootCommand(
    config *config.Manager,
    banner *ui.Banner,
    eventBus *events.EventBus,
) *RootCommand {
    rc := &RootCommand{
        config:      config,
        banner:      banner,
        globalFlags: &GlobalFlags{},
        eventBus:    eventBus,
    }
    
    rc.cmd = &cobra.Command{
        Use:     "specify-cli",
        Short:   "AI辅助开发项目初始化工具",
        Long:    rc.getLongDescription(),
        Version: version.Version,
        PersistentPreRunE: rc.persistentPreRun,
        RunE:    rc.run,
    }
    
    rc.setupFlags()
    return rc
}

// setupFlags 设置命令标志
func (rc *RootCommand) setupFlags() {
    flags := rc.cmd.PersistentFlags()
    
    flags.BoolVar(&rc.globalFlags.Debug, "debug", false, "启用调试模式")
    flags.BoolVar(&rc.globalFlags.Verbose, "verbose", false, "启用详细输出")
    flags.BoolVar(&rc.globalFlags.NoColor, "no-color", false, "禁用彩色输出")
    flags.StringVar(&rc.globalFlags.ConfigDir, "config-dir", "", "指定配置目录")
}

// persistentPreRun 预运行钩子
func (rc *RootCommand) persistentPreRun(cmd *cobra.Command, args []string) error {
    // 设置日志级别
    if rc.globalFlags.Debug {
        rc.config.SetLogLevel("debug")
    } else if rc.globalFlags.Verbose {
        rc.config.SetLogLevel("info")
    }
    
    // 设置颜色输出
    if rc.globalFlags.NoColor {
        rc.config.SetColorOutput(false)
    }
    
    // 设置配置目录
    if rc.globalFlags.ConfigDir != "" {
        rc.config.SetConfigDir(rc.globalFlags.ConfigDir)
    }
    
    return nil
}

// run 根命令执行
func (rc *RootCommand) run(cmd *cobra.Command, args []string) error {
    // 显示横幅
    if err := rc.banner.Show(); err != nil {
        return fmt.Errorf("显示横幅失败: %w", err)
    }
    
    // 显示帮助信息
    return cmd.Help()
}

// AddCommand 添加子命令
func (rc *RootCommand) AddCommand(cmds ...*cobra.Command) {
    rc.cmd.AddCommand(cmds...)
}

// Execute 执行命令
func (rc *RootCommand) Execute() error {
    return rc.cmd.Execute()
}

// GetCommand 获取cobra命令
func (rc *RootCommand) GetCommand() *cobra.Command {
    return rc.cmd
}

// getLongDescription 获取长描述
func (rc *RootCommand) getLongDescription() string {
    return `Specify CLI 是一个AI辅助开发项目初始化工具，支持多种AI助手的项目模板初始化。

主要功能：
  • 项目初始化 - 支持多种AI助手的项目模板
  • 工具检查 - 验证开发环境中必要工具的安装状态
  • 交互式UI - 提供友好的命令行交互界面
  • GitHub集成 - 从GitHub下载和管理项目模板
  • 跨平台支持 - 支持Windows、macOS、Linux操作系统

使用示例：
  specify-cli init my-project --ai claude
  specify-cli check
  specify-cli init --help`
}
```

#### 3.1.2 Init命令详细实现
```go
// internal/cli/commands/init.go
package commands

import (
    "context"
    "fmt"
    
    "github.com/spf13/cobra"
    "specify-cli-go/internal/core/services"
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/models"
)

// InitCommand Init命令结构
type InitCommand struct {
    cmd       *cobra.Command
    service   services.InitService
    ui        ui.UI
    validator *InputValidator
    flags     *InitFlags
}

// InitFlags Init命令标志
type InitFlags struct {
    ProjectName       string
    AI               string
    ScriptType       string
    IgnoreAgentTools bool
    NoGit           bool
    Here            bool
    Force           bool
    SkipTLS         bool
    GitHubToken     string
}

// NewInitCommand 创建Init命令
func NewInitCommand(
    service services.InitService,
    ui ui.UI,
    validator *InputValidator,
) *InitCommand {
    ic := &InitCommand{
        service:   service,
        ui:        ui,
        validator: validator,
        flags:     &InitFlags{},
    }
    
    ic.cmd = &cobra.Command{
        Use:   "init [PROJECT_NAME]",
        Short: "初始化AI辅助开发项目",
        Long:  ic.getLongDescription(),
        Args:  cobra.MaximumNArgs(1),
        RunE:  ic.run,
    }
    
    ic.setupFlags()
    return ic
}

// setupFlags 设置命令标志
func (ic *InitCommand) setupFlags() {
    flags := ic.cmd.Flags()
    
    flags.StringVar(&ic.flags.AI, "ai", "", "指定AI助手类型")
    flags.StringVar(&ic.flags.ScriptType, "script-type", "", "指定脚本类型 (bash|powershell)")
    flags.BoolVar(&ic.flags.IgnoreAgentTools, "ignore-agent-tools", false, "跳过Agent工具检查")
    flags.BoolVar(&ic.flags.NoGit, "no-git", false, "不初始化Git仓库")
    flags.BoolVar(&ic.flags.Here, "here", false, "在当前目录初始化")
    flags.BoolVar(&ic.flags.Force, "force", false, "强制覆盖已存在的项目")
    flags.BoolVar(&ic.flags.SkipTLS, "skip-tls", false, "跳过TLS证书验证")
    flags.StringVar(&ic.flags.GitHubToken, "github-token", "", "GitHub访问令牌")
}

// run 执行Init命令
func (ic *InitCommand) run(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    
    // 解析项目名称
    if len(args) > 0 {
        ic.flags.ProjectName = args[0]
    }
    
    // 构建初始化参数
    initArgs, err := ic.buildInitArgs(ctx)
    if err != nil {
        return fmt.Errorf("构建初始化参数失败: %w", err)
    }
    
    // 执行初始化
    if err := ic.service.InitializeProject(ctx, initArgs); err != nil {
        return fmt.Errorf("项目初始化失败: %w", err)
    }
    
    return nil
}

// buildInitArgs 构建初始化参数
func (ic *InitCommand) buildInitArgs(ctx context.Context) (*models.InitArgs, error) {
    args := &models.InitArgs{
        ProjectName:       ic.flags.ProjectName,
        AI:               ic.flags.AI,
        ScriptType:       ic.flags.ScriptType,
        IgnoreAgentTools: ic.flags.IgnoreAgentTools,
        NoGit:           ic.flags.NoGit,
        Here:            ic.flags.Here,
        Force:           ic.flags.Force,
        SkipTLS:         ic.flags.SkipTLS,
        GitHubToken:     ic.flags.GitHubToken,
    }
    
    // 交互式输入缺失的参数
    if err := ic.interactiveInput(ctx, args); err != nil {
        return nil, err
    }
    
    // 验证参数
    if err := ic.validator.ValidateInitArgs(args); err != nil {
        return nil, err
    }
    
    return args, nil
}

// interactiveInput 交互式输入
func (ic *InitCommand) interactiveInput(ctx context.Context, args *models.InitArgs) error {
    // 项目名称输入
    if args.ProjectName == "" {
        projectName, err := ic.ui.PromptInput("请输入项目名称:", "")
        if err != nil {
            return err
        }
        args.ProjectName = projectName
    }
    
    // AI助手选择
    if args.AI == "" {
        aiOptions := ic.service.GetAvailableAIs()
        selectedAI, err := ic.ui.PromptSelect("请选择AI助手:", aiOptions)
        if err != nil {
            return err
        }
        args.AI = selectedAI
    }
    
    // 脚本类型选择
    if args.ScriptType == "" {
        scriptOptions := ic.service.GetAvailableScriptTypes()
        selectedScript, err := ic.ui.PromptSelect("请选择脚本类型:", scriptOptions)
        if err != nil {
            return err
        }
        args.ScriptType = selectedScript
    }
    
    return nil
}

// GetCommand 获取cobra命令
func (ic *InitCommand) GetCommand() *cobra.Command {
    return ic.cmd
}

// getLongDescription 获取长描述
func (ic *InitCommand) getLongDescription() string {
    return `初始化AI辅助开发项目，支持多种AI助手和脚本类型。

支持的AI助手：
  • claude - Claude Code
  • copilot - GitHub Copilot
  • gemini - Gemini CLI
  • cursor-agent - Cursor
  • qwen - Qwen Code
  • 等等...

支持的脚本类型：
  • bash - Bash脚本 (Linux/macOS)
  • powershell - PowerShell脚本 (Windows)

使用示例：
  specify-cli init my-project
  specify-cli init my-project --ai claude --script-type bash
  specify-cli init --here --ai copilot
  specify-cli init my-project --force --no-git`
}
```

### 3.2 参数验证器设计

#### 3.2.1 输入验证器实现
```go
// internal/cli/validation/input_validator.go
package validation

import (
    "fmt"
    "path/filepath"
    "regexp"
    "strings"
    
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/infrastructure/config"
)

// InputValidator 输入验证器
type InputValidator struct {
    config *config.Manager
}

// NewInputValidator 创建输入验证器
func NewInputValidator(config *config.Manager) *InputValidator {
    return &InputValidator{
        config: config,
    }
}

// ValidateInitArgs 验证初始化参数
func (iv *InputValidator) ValidateInitArgs(args *models.InitArgs) error {
    // 验证项目名称
    if err := iv.ValidateProjectName(args.ProjectName); err != nil {
        return fmt.Errorf("项目名称验证失败: %w", err)
    }
    
    // 验证AI助手
    if err := iv.ValidateAI(args.AI); err != nil {
        return fmt.Errorf("AI助手验证失败: %w", err)
    }
    
    // 验证脚本类型
    if err := iv.ValidateScriptType(args.ScriptType); err != nil {
        return fmt.Errorf("脚本类型验证失败: %w", err)
    }
    
    // 验证GitHub令牌
    if args.GitHubToken != "" {
        if err := iv.ValidateGitHubToken(args.GitHubToken); err != nil {
            return fmt.Errorf("GitHub令牌验证失败: %w", err)
        }
    }
    
    return nil
}

// ValidateProjectName 验证项目名称
func (iv *InputValidator) ValidateProjectName(name string) error {
    if name == "" {
        return fmt.Errorf("项目名称不能为空")
    }
    
    // 检查长度
    if len(name) > 100 {
        return fmt.Errorf("项目名称长度不能超过100个字符")
    }
    
    // 检查字符
    validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
    if !validNameRegex.MatchString(name) {
        return fmt.Errorf("项目名称只能包含字母、数字、下划线和连字符")
    }
    
    // 检查保留名称
    reservedNames := []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}
    lowerName := strings.ToLower(name)
    for _, reserved := range reservedNames {
        if lowerName == reserved {
            return fmt.Errorf("项目名称不能使用保留名称: %s", reserved)
        }
    }
    
    return nil
}

// ValidateAI 验证AI助手
func (iv *InputValidator) ValidateAI(ai string) error {
    if ai == "" {
        return fmt.Errorf("AI助手不能为空")
    }
    
    availableAIs := iv.config.GetAvailableAIs()
    for _, available := range availableAIs {
        if available == ai {
            return nil
        }
    }
    
    return fmt.Errorf("不支持的AI助手: %s，支持的AI助手: %s", ai, strings.Join(availableAIs, ", "))
}

// ValidateScriptType 验证脚本类型
func (iv *InputValidator) ValidateScriptType(scriptType string) error {
    if scriptType == "" {
        return fmt.Errorf("脚本类型不能为空")
    }
    
    availableTypes := iv.config.GetAvailableScriptTypes()
    for _, available := range availableTypes {
        if available == scriptType {
            return nil
        }
    }
    
    return fmt.Errorf("不支持的脚本类型: %s，支持的脚本类型: %s", scriptType, strings.Join(availableTypes, ", "))
}

// ValidateGitHubToken 验证GitHub令牌
func (iv *InputValidator) ValidateGitHubToken(token string) error {
    // GitHub个人访问令牌格式验证
    if len(token) < 40 {
        return fmt.Errorf("GitHub令牌长度不足")
    }
    
    // 检查令牌格式 (ghp_开头的新格式或传统格式)
    if strings.HasPrefix(token, "ghp_") {
        if len(token) != 40 {
            return fmt.Errorf("GitHub令牌格式不正确")
        }
    } else {
        // 传统格式验证
        validTokenRegex := regexp.MustCompile(`^[a-f0-9]{40}$`)
        if !validTokenRegex.MatchString(token) {
            return fmt.Errorf("GitHub令牌格式不正确")
        }
    }
    
    return nil
}

// ValidateProjectPath 验证项目路径
func (iv *InputValidator) ValidateProjectPath(path string, force bool) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("无法解析项目路径: %w", err)
    }
    
    // 检查路径是否存在
    if exists, err := iv.pathExists(absPath); err != nil {
        return fmt.Errorf("检查路径失败: %w", err)
    } else if exists && !force {
        return fmt.Errorf("项目路径已存在: %s，使用 --force 强制覆盖", absPath)
    }
    
    return nil
}

// pathExists 检查路径是否存在
func (iv *InputValidator) pathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}
```

---

## 7. 系统集成详细设计

### 7.1 特殊工具检查器设计

#### 7.1.1 特殊工具检查器接口
```go
// internal/core/checkers/special_checker.go
package checkers

import (
    "context"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    
    "specify-cli-go/internal/models"
)

// SpecialChecker 特殊工具检查器接口
type SpecialChecker interface {
    // CheckClaude 检查Claude CLI
    CheckClaude(ctx context.Context) (*models.ToolCheckResult, error)
    
    // CheckCursor 检查Cursor编辑器
    CheckCursor(ctx context.Context) (*models.ToolCheckResult, error)
    
    // CheckSpecialTool 检查特殊工具
    CheckSpecialTool(ctx context.Context, toolName string) (*models.ToolCheckResult, error)
    
    // GetSpecialPaths 获取特殊路径
    GetSpecialPaths(toolName string) []string
}

// SpecialToolChecker 特殊工具检查器实现
type SpecialToolChecker struct {
    osType      string
    toolPaths   map[string][]string
    claudeLocal string
}

// NewSpecialToolChecker 创建特殊工具检查器
func NewSpecialToolChecker() SpecialChecker {
    stc := &SpecialToolChecker{
        osType:      runtime.GOOS,
        claudeLocal: os.Getenv("CLAUDE_LOCAL_PATH"),
    }
    
    stc.initializeSpecialPaths()
    return stc
}

// CheckClaude 检查Claude CLI
func (stc *SpecialToolChecker) CheckClaude(ctx context.Context) (*models.ToolCheckResult, error) {
    result := &models.ToolCheckResult{
        Name:      "claude",
        Installed: false,
    }
    
    // 检查标准路径
    if path, found := stc.checkStandardPath("claude"); found {
        result.Installed = true
        result.Path = path
        result.Message = "Claude CLI 已安装"
        return result, nil
    }
    
    // 检查特殊路径：migrate-installer
    if stc.claudeLocal != "" {
        migrateInstallerPath := filepath.Join(stc.claudeLocal, "migrate-installer")
        if stc.pathExists(migrateInstallerPath) {
            result.Installed = true
            result.Path = migrateInstallerPath
            result.Message = "Claude CLI (migrate-installer) 已安装"
            return result, nil
        }
    }
    
    result.Message = "Claude CLI 未安装或未找到"
    return result, nil
}

// CheckCursor 检查Cursor编辑器
func (stc *SpecialToolChecker) CheckCursor(ctx context.Context) (*models.ToolCheckResult, error) {
    result := &models.ToolCheckResult{
        Name:      "cursor",
        Installed: false,
    }
    
    paths := stc.GetSpecialPaths("cursor")
    for _, path := range paths {
        if stc.pathExists(path) {
            result.Installed = true
            result.Path = path
            result.Message = "Cursor 编辑器已安装"
            return result, nil
        }
    }
    
    result.Message = "Cursor 编辑器未安装或未找到"
    return result, nil
}

// CheckSpecialTool 检查特殊工具
func (stc *SpecialToolChecker) CheckSpecialTool(ctx context.Context, toolName string) (*models.ToolCheckResult, error) {
    switch toolName {
    case "claude":
        return stc.CheckClaude(ctx)
    case "cursor":
        return stc.CheckCursor(ctx)
    default:
        return &models.ToolCheckResult{
            Name:      toolName,
            Installed: false,
            Message:   "不支持的特殊工具",
        }, nil
    }
}

// GetSpecialPaths 获取特殊路径
func (stc *SpecialToolChecker) GetSpecialPaths(toolName string) []string {
    if paths, exists := stc.toolPaths[toolName]; exists {
        return stc.expandPaths(paths)
    }
    return []string{}
}

// initializeSpecialPaths 初始化特殊路径
func (stc *SpecialToolChecker) initializeSpecialPaths() {
    stc.toolPaths = map[string][]string{
        "claude": {
            "claude",
            filepath.Join("${CLAUDE_LOCAL_PATH}", "migrate-installer"),
        },
        "cursor": {
            "cursor",
        },
    }
    
    // 根据操作系统添加特定路径
    switch stc.osType {
    case "windows":
        stc.toolPaths["cursor"] = append(stc.toolPaths["cursor"],
            filepath.Join("${USERPROFILE}", "AppData", "Local", "Programs", "cursor", "cursor.exe"),
            "C:\\Users\\%USERNAME%\\AppData\\Local\\Programs\\cursor\\cursor.exe",
        )
    case "darwin":
        stc.toolPaths["cursor"] = append(stc.toolPaths["cursor"],
            "/Applications/Cursor.app/Contents/Resources/app/bin/cursor",
        )
    case "linux":
        stc.toolPaths["cursor"] = append(stc.toolPaths["cursor"],
            "/usr/local/bin/cursor",
            "/opt/cursor/cursor",
        )
    }
}

// checkStandardPath 检查标准路径
func (stc *SpecialToolChecker) checkStandardPath(toolName string) (string, bool) {
    // 使用which/where命令检查
    var cmd string
    if stc.osType == "windows" {
        cmd = "where"
    } else {
        cmd = "which"
    }
    
    // 这里应该执行命令检查，简化实现
    return "", false
}

// pathExists 检查路径是否存在
func (stc *SpecialToolChecker) pathExists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

// expandPaths 展开路径中的环境变量
func (stc *SpecialToolChecker) expandPaths(paths []string) []string {
    var expanded []string
    for _, path := range paths {
        expandedPath := os.ExpandEnv(path)
        expanded = append(expanded, expandedPath)
    }
    return expanded
}
```

### 7.2 权限管理器设计

#### 7.2.1 权限管理器接口
```go
// internal/core/system/permission_manager.go
package system

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strings"
)

// PermissionManager 权限管理器接口
type PermissionManager interface {
    // SetExecutablePermissions 设置可执行权限
    SetExecutablePermissions(ctx context.Context, dir string, recursive bool) error
    
    // SetFilePermissions 设置文件权限
    SetFilePermissions(ctx context.Context, filePath string, mode os.FileMode) error
    
    // CheckPermissions 检查权限
    CheckPermissions(ctx context.Context, path string) (*PermissionInfo, error)
    
    // FixPermissions 修复权限问题
    FixPermissions(ctx context.Context, dir string) error
}

// PermissionInfo 权限信息
type PermissionInfo struct {
    Path        string      `json:"path"`
    Mode        os.FileMode `json:"mode"`
    Readable    bool        `json:"readable"`
    Writable    bool        `json:"writable"`
    Executable  bool        `json:"executable"`
    Issues      []string    `json:"issues,omitempty"`
}

// PermissionManagerImpl 权限管理器实现
type PermissionManagerImpl struct {
    osType string
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager() PermissionManager {
    return &PermissionManagerImpl{
        osType: runtime.GOOS,
    }
}

// SetExecutablePermissions 设置可执行权限
func (pm *PermissionManagerImpl) SetExecutablePermissions(ctx context.Context, dir string, recursive bool) error {
    if pm.osType == "windows" {
        // Windows不需要设置执行权限
        return nil
    }
    
    if !recursive {
        return pm.setExecutableForPath(dir)
    }
    
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // 检查上下文是否被取消
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        // 只处理脚本文件
        if pm.isScriptFile(path) {
            return pm.setExecutableForPath(path)
        }
        
        return nil
    })
}

// SetFilePermissions 设置文件权限
func (pm *PermissionManagerImpl) SetFilePermissions(ctx context.Context, filePath string, mode os.FileMode) error {
    if pm.osType == "windows" {
        // Windows权限处理简化
        return nil
    }
    
    return os.Chmod(filePath, mode)
}

// CheckPermissions 检查权限
func (pm *PermissionManagerImpl) CheckPermissions(ctx context.Context, path string) (*PermissionInfo, error) {
    info, err := os.Stat(path)
    if err != nil {
        return nil, fmt.Errorf("获取文件信息失败: %w", err)
    }
    
    mode := info.Mode()
    permInfo := &PermissionInfo{
        Path:       path,
        Mode:       mode,
        Readable:   mode&0400 != 0,
        Writable:   mode&0200 != 0,
        Executable: mode&0100 != 0,
    }
    
    // 检查权限问题
    if pm.isScriptFile(path) && !permInfo.Executable {
        permInfo.Issues = append(permInfo.Issues, "脚本文件缺少执行权限")
    }
    
    return permInfo, nil
}

// FixPermissions 修复权限问题
func (pm *PermissionManagerImpl) FixPermissions(ctx context.Context, dir string) error {
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        permInfo, err := pm.CheckPermissions(ctx, path)
        if err != nil {
            return err
        }
        
        // 修复权限问题
        if len(permInfo.Issues) > 0 {
            if pm.isScriptFile(path) {
                return pm.setExecutableForPath(path)
            }
        }
        
        return nil
    })
}

// setExecutableForPath 为路径设置可执行权限
func (pm *PermissionManagerImpl) setExecutableForPath(path string) error {
    if pm.osType == "windows" {
        return nil
    }
    
    return os.Chmod(path, 0755)
}

// isScriptFile 判断是否为脚本文件
func (pm *PermissionManagerImpl) isScriptFile(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    scriptExts := []string{".sh", ".bash", ".zsh", ".fish", ".ps1"}
    
    for _, scriptExt := range scriptExts {
        if ext == scriptExt {
            return true
        }
    }
    
    return false
}
```

### 7.3 安全通知器设计

#### 7.3.1 安全通知器接口
```go
// internal/core/system/security_notifier.go
package system

import (
    "fmt"
    "path/filepath"
    
    "specify-cli-go/internal/cli/ui"
)

// SecurityNotifier 安全通知器接口
type SecurityNotifier interface {
    // ShowAgentFolderWarning 显示Agent文件夹安全警告
    ShowAgentFolderWarning(agentFolder string)
    
    // ShowCredentialLeakageWarning 显示凭据泄露警告
    ShowCredentialLeakageWarning()
    
    // ShowFilePermissionWarning 显示文件权限警告
    ShowFilePermissionWarning(filePath string)
    
    // ShowGitIgnoreRecommendation 显示.gitignore建议
    ShowGitIgnoreRecommendation(patterns []string)
    
    // ShowEnvironmentVariableWarning 显示环境变量警告
    ShowEnvironmentVariableWarning(varNames []string)
}

// SecurityNotifierImpl 安全通知器实现
type SecurityNotifierImpl struct {
    ui ui.UI
}

// NewSecurityNotifier 创建安全通知器
func NewSecurityNotifier(ui ui.UI) SecurityNotifier {
    return &SecurityNotifierImpl{
        ui: ui,
    }
}

// ShowAgentFolderWarning 显示Agent文件夹安全警告
func (sn *SecurityNotifierImpl) ShowAgentFolderWarning(agentFolder string) {
    warning := fmt.Sprintf(`
⚠️  安全提示：
建议将 %s 文件夹添加到 .gitignore 中，避免意外提交敏感信息。

建议执行：
echo "%s" >> .gitignore

或手动编辑 .gitignore 文件添加：
%s
`, agentFolder, agentFolder, agentFolder)
    
    sn.ui.ShowWarning(warning)
}

// ShowCredentialLeakageWarning 显示凭据泄露警告
func (sn *SecurityNotifierImpl) ShowCredentialLeakageWarning() {
    warning := `
🔒 安全提醒：
请确保不要在代码中硬编码API密钥或访问令牌。

建议的安全实践：
1. 使用环境变量存储敏感信息
2. 使用配置文件（并添加到.gitignore）
3. 使用密钥管理服务
4. 定期轮换API密钥

示例环境变量设置：
export OPENAI_API_KEY="your-api-key"
export CLAUDE_API_KEY="your-api-key"
`
    
    sn.ui.ShowWarning(warning)
}

// ShowFilePermissionWarning 显示文件权限警告
func (sn *SecurityNotifierImpl) ShowFilePermissionWarning(filePath string) {
    warning := fmt.Sprintf(`
⚠️  文件权限警告：
文件 %s 可能存在权限问题。

建议检查：
1. 文件是否具有适当的读写权限
2. 脚本文件是否具有执行权限
3. 敏感文件是否权限过于宽松

修复命令（Unix系统）：
chmod 755 %s  # 对于可执行文件
chmod 644 %s  # 对于普通文件
`, filePath, filePath, filePath)
    
    sn.ui.ShowWarning(warning)
}

// ShowGitIgnoreRecommendation 显示.gitignore建议
func (sn *SecurityNotifierImpl) ShowGitIgnoreRecommendation(patterns []string) {
    if len(patterns) == 0 {
        return
    }
    
    warning := `
📝 .gitignore 建议：
为了保护敏感信息，建议将以下模式添加到 .gitignore：

`
    
    for _, pattern := range patterns {
        warning += fmt.Sprintf("  %s\n", pattern)
    }
    
    warning += `
添加方法：
1. 手动编辑 .gitignore 文件
2. 或使用命令：echo "pattern" >> .gitignore
`
    
    sn.ui.ShowWarning(warning)
}

// ShowEnvironmentVariableWarning 显示环境变量警告
func (sn *SecurityNotifierImpl) ShowEnvironmentVariableWarning(varNames []string) {
    if len(varNames) == 0 {
        return
    }
    
    warning := `
🔧 环境变量配置提醒：
以下环境变量可能需要配置：

`
    
    for _, varName := range varNames {
        warning += fmt.Sprintf("  %s\n", varName)
    }
    
    warning += `
配置方法：
Windows: set VARIABLE_NAME=value
Unix:    export VARIABLE_NAME=value

永久配置：
Windows: 通过系统属性 -> 环境变量
Unix:    添加到 ~/.bashrc 或 ~/.zshrc
`
    
    sn.ui.ShowWarning(warning)
}
```

### 7.4 环境指导器设计

#### 7.4.1 环境指导器接口
```go
// internal/core/system/environment_guide.go
package system

import (
    "fmt"
    "path/filepath"
    "runtime"
    "strings"
    
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/models"
)

// EnvironmentGuide 环境指导器接口
type EnvironmentGuide interface {
    // GenerateSetupInstructions 生成设置指令
    GenerateSetupInstructions(agent string, projectPath string) []string
    
    // ShowNextSteps 显示后续步骤
    ShowNextSteps(agent string, isCurrentDir bool)
    
    // GenerateShellConfig 生成Shell配置
    GenerateShellConfig(agent string, projectPath string) *ShellConfig
    
    // ShowIDEInstructions 显示IDE配置指令
    ShowIDEInstructions(agent string, projectPath string)
    
    // ShowEnvironmentSetup 显示环境设置
    ShowEnvironmentSetup(agent string, projectPath string)
}

// ShellConfig Shell配置
type ShellConfig struct {
    Variables   map[string]string `json:"variables"`
    Aliases     map[string]string `json:"aliases"`
    Exports     []string          `json:"exports"`
    ConfigFiles []string          `json:"config_files"`
}

// EnvironmentGuideImpl 环境指导器实现
type EnvironmentGuideImpl struct {
    osType string
    ui     ui.UI
}

// NewEnvironmentGuide 创建环境指导器
func NewEnvironmentGuide(ui ui.UI) EnvironmentGuide {
    return &EnvironmentGuideImpl{
        osType: runtime.GOOS,
        ui:     ui,
    }
}

// GenerateSetupInstructions 生成设置指令
func (eg *EnvironmentGuideImpl) GenerateSetupInstructions(agent string, projectPath string) []string {
    instructions := []string{}
    
    switch strings.ToLower(agent) {
    case "claude":
        if eg.osType == "windows" {
            instructions = append(instructions,
                fmt.Sprintf("set CODEX_HOME=%s", projectPath),
                fmt.Sprintf("set CLAUDE_PROJECT_PATH=%s", projectPath),
                "# 添加到系统环境变量以永久保存",
            )
        } else {
            instructions = append(instructions,
                fmt.Sprintf("export CODEX_HOME=%s", projectPath),
                fmt.Sprintf("export CLAUDE_PROJECT_PATH=%s", projectPath),
                fmt.Sprintf("echo 'export CODEX_HOME=%s' >> ~/.bashrc", projectPath),
                fmt.Sprintf("echo 'export CLAUDE_PROJECT_PATH=%s' >> ~/.bashrc", projectPath),
            )
        }
        
    case "copilot":
        instructions = append(instructions,
            "# GitHub Copilot 已配置完成",
            "# 请确保已在 VS Code 中安装 GitHub Copilot 扩展",
            "# 使用 Ctrl+I 或 Cmd+I 启动 Copilot Chat",
        )
        
    case "cursor":
        instructions = append(instructions,
            "# Cursor 编辑器已配置完成",
            "# 使用 Ctrl+K 或 Cmd+K 启动 AI 助手",
            fmt.Sprintf("# 项目路径: %s", projectPath),
        )
        
    case "gemini":
        if eg.osType == "windows" {
            instructions = append(instructions,
                fmt.Sprintf("set GEMINI_PROJECT_PATH=%s", projectPath),
            )
        } else {
            instructions = append(instructions,
                fmt.Sprintf("export GEMINI_PROJECT_PATH=%s", projectPath),
                fmt.Sprintf("echo 'export GEMINI_PROJECT_PATH=%s' >> ~/.bashrc", projectPath),
            )
        }
        
    default:
        instructions = append(instructions,
            fmt.Sprintf("# %s 项目已初始化", agent),
            fmt.Sprintf("# 项目路径: %s", projectPath),
        )
    }
    
    return instructions
}

// ShowNextSteps 显示后续步骤
func (eg *EnvironmentGuideImpl) ShowNextSteps(agent string, isCurrentDir bool) {
    steps := []string{}
    
    if !isCurrentDir {
        steps = append(steps, "cd <project-name>")
    }
    
    switch strings.ToLower(agent) {
    case "claude":
        steps = append(steps,
            "# 启动 Claude Code",
            "code .",
            "# 或使用 Claude CLI",
            "claude chat",
            "# 或使用 migrate-installer",
            "migrate-installer",
        )
        
    case "copilot":
        steps = append(steps,
            "# 启动 VS Code",
            "code .",
            "# 确保 GitHub Copilot 扩展已安装并登录",
            "# 使用 Ctrl+I 开始对话",
        )
        
    case "cursor":
        steps = append(steps,
            "# 启动 Cursor 编辑器",
            "cursor .",
            "# 或直接打开项目文件夹",
            "# 使用 Ctrl+K 启动 AI 助手",
        )
        
    case "gemini":
        steps = append(steps,
            "# 配置 Gemini API",
            "# 设置 GEMINI_API_KEY 环境变量",
            "# 启动开发环境",
            "code .",
        )
        
    default:
        steps = append(steps,
            "# 启动开发环境",
            "code .",
            fmt.Sprintf("# 开始使用 %s 进行开发", agent),
        )
    }
    
    nextStepsMsg := fmt.Sprintf(`
🚀 后续步骤：

%s

`, strings.Join(steps, "\n"))
    
    eg.ui.ShowInfo(nextStepsMsg)
}

// GenerateShellConfig 生成Shell配置
func (eg *EnvironmentGuideImpl) GenerateShellConfig(agent string, projectPath string) *ShellConfig {
    config := &ShellConfig{
        Variables: make(map[string]string),
        Aliases:   make(map[string]string),
        Exports:   []string{},
    }
    
    switch strings.ToLower(agent) {
    case "claude":
        config.Variables["CODEX_HOME"] = projectPath
        config.Variables["CLAUDE_PROJECT_PATH"] = projectPath
        config.Aliases["claude-project"] = fmt.Sprintf("cd %s", projectPath)
        
    case "gemini":
        config.Variables["GEMINI_PROJECT_PATH"] = projectPath
        config.Aliases["gemini-project"] = fmt.Sprintf("cd %s", projectPath)
        
    default:
        config.Variables[fmt.Sprintf("%s_PROJECT_PATH", strings.ToUpper(agent))] = projectPath
    }
    
    // 根据操作系统设置配置文件
    switch eg.osType {
    case "windows":
        config.ConfigFiles = []string{"PowerShell Profile"}
    default:
        config.ConfigFiles = []string{"~/.bashrc", "~/.zshrc", "~/.profile"}
    }
    
    return config
}

// ShowIDEInstructions 显示IDE配置指令
func (eg *EnvironmentGuideImpl) ShowIDEInstructions(agent string, projectPath string) {
    var instructions string
    
    switch strings.ToLower(agent) {
    case "copilot":
        instructions = `
💡 VS Code + GitHub Copilot 配置：

1. 安装 GitHub Copilot 扩展
2. 登录 GitHub 账户
3. 启用 Copilot 功能
4. 使用快捷键：
   - Ctrl+I (Cmd+I): 启动 Copilot Chat
   - Tab: 接受建议
   - Alt+]: 下一个建议
   - Alt+[: 上一个建议
`
        
    case "cursor":
        instructions = `
💡 Cursor 编辑器配置：

1. 打开项目文件夹
2. 配置 AI 模型（GPT-4, Claude等）
3. 使用快捷键：
   - Ctrl+K (Cmd+K): 启动 AI 助手
   - Ctrl+L (Cmd+L): 与 AI 对话
   - Ctrl+I (Cmd+I): 内联编辑
`
        
    case "claude":
        instructions = `
💡 Claude 开发环境配置：

1. 确保 Claude CLI 已安装
2. 设置项目环境变量
3. 使用命令：
   - claude chat: 启动对话
   - claude code: 代码分析
   - migrate-installer: 特殊安装器
`
        
    default:
        instructions = fmt.Sprintf(`
💡 %s 开发环境：

项目已初始化，请根据 %s 的文档配置开发环境。
项目路径: %s
`, agent, agent, projectPath)
    }
    
    eg.ui.ShowInfo(instructions)
}

// ShowEnvironmentSetup 显示环境设置
func (eg *EnvironmentGuideImpl) ShowEnvironmentSetup(agent string, projectPath string) {
    config := eg.GenerateShellConfig(agent, projectPath)
    
    setupMsg := fmt.Sprintf(`
🔧 环境变量设置：

`)
    
    // 显示环境变量
    for key, value := range config.Variables {
        if eg.osType == "windows" {
            setupMsg += fmt.Sprintf("set %s=%s\n", key, value)
        } else {
            setupMsg += fmt.Sprintf("export %s=%s\n", key, value)
        }
    }
    
    // 显示别名
    if len(config.Aliases) > 0 {
        setupMsg += "\n🔗 便捷别名：\n"
        for alias, command := range config.Aliases {
            setupMsg += fmt.Sprintf("alias %s='%s'\n", alias, command)
        }
    }
    
    // 显示配置文件
    if len(config.ConfigFiles) > 0 {
        setupMsg += "\n📝 配置文件：\n"
        for _, configFile := range config.ConfigFiles {
            setupMsg += fmt.Sprintf("  %s\n", configFile)
        }
    }
    
    eg.ui.ShowInfo(setupMsg)
}
```

### 7.5 系统集成模块

#### 7.5.1 模块定义
```go
// internal/core/system/module.go
package system

import (
    "go.uber.org/fx"
    
    "specify-cli-go/internal/cli/ui"
)

// SystemModule 系统集成模块
var SystemModule = fx.Module("system",
    // 提供特殊工具检查器
    fx.Provide(func() SpecialChecker {
        return NewSpecialToolChecker()
    }),
    
    // 提供权限管理器
    fx.Provide(func() PermissionManager {
        return NewPermissionManager()
    }),
    
    // 提供安全通知器
    fx.Provide(func(ui ui.UI) SecurityNotifier {
        return NewSecurityNotifier(ui)
    }),
    
    // 提供环境指导器
    fx.Provide(func(ui ui.UI) EnvironmentGuide {
        return NewEnvironmentGuide(ui)
    }),
    
    // 生命周期钩子
    fx.Invoke(func(lc fx.Lifecycle, 
        checker SpecialChecker,
        permManager PermissionManager,
        secNotifier SecurityNotifier,
        envGuide EnvironmentGuide) {
        
        lc.Append(fx.Hook{
            OnStart: func(ctx context.Context) error {
                // 初始化系统集成组件
                return nil
            },
            OnStop: func(ctx context.Context) error {
                // 清理资源
                return nil
            },
        })
    }),
)
```

---

## 8. 基础设施层详细设计

### 8.1 配置管理设计

#### 8.1.1 配置管理器接口
```go
// internal/infrastructure/config/manager.go
package config

import (
    "context"
    "fmt"
    "path/filepath"
    "sync"
    
    "github.com/spf13/viper"
    "specify-cli-go/internal/models"
)

// Manager 配置管理器接口
type Manager interface {
    // LoadConfig 加载配置
    LoadConfig(ctx context.Context) error
    
    // GetAgentConfig 获取Agent配置
    GetAgentConfig(agentName string) (*models.AgentConfig, error)
    
    // GetAllAgentConfigs 获取所有Agent配置
    GetAllAgentConfigs() (map[string]*models.AgentConfig, error)
    
    // GetScriptTypeConfig 获取脚本类型配置
    GetScriptTypeConfig(scriptType string) (*models.ScriptTypeConfig, error)
    
    // GetAppConfig 获取应用配置
    GetAppConfig() (*models.AppConfig, error)
    
    // SaveConfig 保存配置
    SaveConfig(ctx context.Context) error
    
    // WatchConfig 监听配置变化
    WatchConfig(ctx context.Context, callback func()) error
    
    // ValidateConfig 验证配置
    ValidateConfig() error
}

// ManagerImpl 配置管理器实现
type ManagerImpl struct {
    viper       *viper.Viper
    configPath  string
    agentConfigs map[string]*models.AgentConfig
    scriptConfigs map[string]*models.ScriptTypeConfig
    appConfig   *models.AppConfig
    mutex       sync.RWMutex
}

// NewManager 创建配置管理器
func NewManager(configPath string) Manager {
    return &ManagerImpl{
        viper:         viper.New(),
        configPath:    configPath,
        agentConfigs:  make(map[string]*models.AgentConfig),
        scriptConfigs: make(map[string]*models.ScriptTypeConfig),
    }
}

// LoadConfig 加载配置
func (m *ManagerImpl) LoadConfig(ctx context.Context) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    // 设置配置文件路径
    m.viper.SetConfigFile(m.configPath)
    m.viper.SetConfigType("yaml")
    
    // 设置默认值
    m.setDefaults()
    
    // 读取配置文件
    if err := m.viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // 配置文件不存在，使用默认配置
            return m.initializeDefaultConfig(ctx)
        }
        return fmt.Errorf("读取配置文件失败: %w", err)
    }
    
    // 解析配置
    return m.parseConfig()
}

// GetAgentConfig 获取Agent配置
func (m *ManagerImpl) GetAgentConfig(agentName string) (*models.AgentConfig, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    config, exists := m.agentConfigs[agentName]
    if !exists {
        return nil, fmt.Errorf("Agent配置不存在: %s", agentName)
    }
    
    return config, nil
}

// GetAllAgentConfigs 获取所有Agent配置
func (m *ManagerImpl) GetAllAgentConfigs() (map[string]*models.AgentConfig, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    // 返回配置的副本
    configs := make(map[string]*models.AgentConfig)
    for name, config := range m.agentConfigs {
        configs[name] = config
    }
    
    return configs, nil
}

// GetScriptTypeConfig 获取脚本类型配置
func (m *ManagerImpl) GetScriptTypeConfig(scriptType string) (*models.ScriptTypeConfig, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    config, exists := m.scriptConfigs[scriptType]
    if !exists {
        return nil, fmt.Errorf("脚本类型配置不存在: %s", scriptType)
    }
    
    return config, nil
}

// GetAppConfig 获取应用配置
func (m *ManagerImpl) GetAppConfig() (*models.AppConfig, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    if m.appConfig == nil {
        return nil, fmt.Errorf("应用配置未初始化")
    }
    
    return m.appConfig, nil
}

// SaveConfig 保存配置
func (m *ManagerImpl) SaveConfig(ctx context.Context) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    return m.viper.WriteConfig()
}

// WatchConfig 监听配置变化
func (m *ManagerImpl) WatchConfig(ctx context.Context, callback func()) error {
    m.viper.WatchConfig()
    m.viper.OnConfigChange(func(e fsnotify.Event) {
        if callback != nil {
            callback()
        }
    })
    
    return nil
}

// ValidateConfig 验证配置
func (m *ManagerImpl) ValidateConfig() error {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    // 验证Agent配置
    for name, config := range m.agentConfigs {
        if err := m.validateAgentConfig(name, config); err != nil {
            return fmt.Errorf("Agent配置验证失败 %s: %w", name, err)
        }
    }
    
    // 验证脚本类型配置
    for name, config := range m.scriptConfigs {
        if err := m.validateScriptConfig(name, config); err != nil {
            return fmt.Errorf("脚本配置验证失败 %s: %w", name, err)
        }
    }
    
    return nil
}

// setDefaults 设置默认值
func (m *ManagerImpl) setDefaults() {
    m.viper.SetDefault("app.name", "Specify CLI")
    m.viper.SetDefault("app.version", "1.0.0")
    m.viper.SetDefault("app.debug", false)
    m.viper.SetDefault("github.timeout", 30)
    m.viper.SetDefault("github.retry_count", 3)
}

// initializeDefaultConfig 初始化默认配置
func (m *ManagerImpl) initializeDefaultConfig(ctx context.Context) error {
    // 初始化Agent配置
    m.initializeAgentConfigs()
    
    // 初始化脚本类型配置
    m.initializeScriptConfigs()
    
    // 初始化应用配置
    m.initializeAppConfig()
    
    // 保存默认配置
    return m.SaveConfig(ctx)
}

// parseConfig 解析配置
func (m *ManagerImpl) parseConfig() error {
    // 解析Agent配置
    agentsConfig := m.viper.GetStringMap("agents")
    for name, configData := range agentsConfig {
        var agentConfig models.AgentConfig
        if err := m.viper.UnmarshalKey(fmt.Sprintf("agents.%s", name), &agentConfig); err != nil {
            return fmt.Errorf("解析Agent配置失败 %s: %w", name, err)
        }
        m.agentConfigs[name] = &agentConfig
    }
    
    // 解析脚本类型配置
    scriptsConfig := m.viper.GetStringMap("scripts")
    for name, configData := range scriptsConfig {
        var scriptConfig models.ScriptTypeConfig
        if err := m.viper.UnmarshalKey(fmt.Sprintf("scripts.%s", name), &scriptConfig); err != nil {
            return fmt.Errorf("解析脚本配置失败 %s: %w", name, err)
        }
        m.scriptConfigs[name] = &scriptConfig
    }
    
    // 解析应用配置
    var appConfig models.AppConfig
    if err := m.viper.UnmarshalKey("app", &appConfig); err != nil {
        return fmt.Errorf("解析应用配置失败: %w", err)
    }
    m.appConfig = &appConfig
    
    return nil
}

// initializeAgentConfigs 初始化Agent配置
func (m *ManagerImpl) initializeAgentConfigs() {
    m.agentConfigs = map[string]*models.AgentConfig{
        "copilot": {
            Name:        "GitHub Copilot",
            Type:        "IDE-based",
            Description: "GitHub Copilot AI助手",
            Requirements: []string{"VS Code", "GitHub Copilot Extension"},
            SetupInstructions: []string{
                "安装VS Code",
                "安装GitHub Copilot扩展",
                "登录GitHub账户",
            },
        },
        "claude": {
            Name:        "Claude",
            Type:        "CLI",
            Description: "Anthropic Claude AI助手",
            Requirements: []string{"Claude CLI"},
            SetupInstructions: []string{
                "安装Claude CLI",
                "配置API密钥",
                "设置项目环境",
            },
        },
        "cursor": {
            Name:        "Cursor",
            Type:        "IDE-based",
            Description: "Cursor AI编辑器",
            Requirements: []string{"Cursor Editor"},
            SetupInstructions: []string{
                "下载并安装Cursor",
                "配置AI模型",
                "设置项目",
            },
        },
        "gemini": {
            Name:        "Google Gemini",
            Type:        "CLI",
            Description: "Google Gemini AI助手",
            Requirements: []string{"Gemini CLI"},
            SetupInstructions: []string{
                "安装Gemini CLI",
                "配置API密钥",
                "设置项目环境",
            },
        },
        // 其他9个Agent配置...
    }
}

// initializeScriptConfigs 初始化脚本类型配置
func (m *ManagerImpl) initializeScriptConfigs() {
    m.scriptConfigs = map[string]*models.ScriptTypeConfig{
        "bash": {
            Name:      "Bash Scripts",
            Extension: ".sh",
            Shebang:   "#!/bin/bash",
            Template:  "#!/bin/bash\n\n# {{.Description}}\n\nset -e\n\necho \"Starting {{.ProjectName}}...\"\n",
        },
        "powershell": {
            Name:      "PowerShell Scripts",
            Extension: ".ps1",
            Shebang:   "# PowerShell Script",
            Template:  "# {{.Description}}\n\nWrite-Host \"Starting {{.ProjectName}}...\"\n",
        },
    }
}

// initializeAppConfig 初始化应用配置
func (m *ManagerImpl) initializeAppConfig() {
    m.appConfig = &models.AppConfig{
        Name:    "Specify CLI",
        Version: "1.0.0",
        Debug:   false,
        GitHub: models.GitHubConfig{
            Timeout:    30,
            RetryCount: 3,
            UserAgent:  "Specify-CLI/1.0.0",
        },
    }
}

// validateAgentConfig 验证Agent配置
func (m *ManagerImpl) validateAgentConfig(name string, config *models.AgentConfig) error {
    if config.Name == "" {
        return fmt.Errorf("Agent名称不能为空")
    }
    
    if config.Type != "CLI" && config.Type != "IDE-based" {
        return fmt.Errorf("无效的Agent类型: %s", config.Type)
    }
    
    return nil
}

// validateScriptConfig 验证脚本配置
func (m *ManagerImpl) validateScriptConfig(name string, config *models.ScriptTypeConfig) error {
    if config.Name == "" {
        return fmt.Errorf("脚本类型名称不能为空")
    }
    
    if config.Extension == "" {
        return fmt.Errorf("脚本扩展名不能为空")
    }
    
    return nil
}
```

### 8.2 文件系统管理设计

#### 8.2.1 文件管理器接口
```go
// internal/infrastructure/filesystem/file_manager.go
package filesystem

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/otiai10/copy"
)

// FileManager 文件管理器接口
type FileManager interface {
    // CreateDirectory 创建目录
    CreateDirectory(ctx context.Context, path string, perm os.FileMode) error
    
    // CreateFile 创建文件
    CreateFile(ctx context.Context, path string, content []byte, perm os.FileMode) error
    
    // CopyFile 复制文件
    CopyFile(ctx context.Context, src, dst string) error
    
    // CopyDirectory 复制目录
    CopyDirectory(ctx context.Context, src, dst string) error
    
    // MoveFile 移动文件
    MoveFile(ctx context.Context, src, dst string) error
    
    // DeleteFile 删除文件
    DeleteFile(ctx context.Context, path string) error
    
    // DeleteDirectory 删除目录
    DeleteDirectory(ctx context.Context, path string) error
    
    // Exists 检查路径是否存在
    Exists(path string) bool
    
    // IsDirectory 检查是否为目录
    IsDirectory(path string) bool
    
    // IsFile 检查是否为文件
    IsFile(path string) bool
    
    // ListDirectory 列出目录内容
    ListDirectory(ctx context.Context, path string) ([]os.FileInfo, error)
    
    // ReadFile 读取文件
    ReadFile(ctx context.Context, path string) ([]byte, error)
    
    // WriteFile 写入文件
    WriteFile(ctx context.Context, path string, content []byte, perm os.FileMode) error
    
    // GetFileInfo 获取文件信息
    GetFileInfo(path string) (os.FileInfo, error)
    
    // WalkDirectory 遍历目录
    WalkDirectory(ctx context.Context, root string, walkFn filepath.WalkFunc) error
}

// FileManagerImpl 文件管理器实现
type FileManagerImpl struct {
    basePath string
}

// NewFileManager 创建文件管理器
func NewFileManager(basePath string) FileManager {
    return &FileManagerImpl{
        basePath: basePath,
    }
}

// CreateDirectory 创建目录
func (fm *FileManagerImpl) CreateDirectory(ctx context.Context, path string, perm os.FileMode) error {
    fullPath := fm.getFullPath(path)
    
    // 检查上下文是否被取消
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    return os.MkdirAll(fullPath, perm)
}

// CreateFile 创建文件
func (fm *FileManagerImpl) CreateFile(ctx context.Context, path string, content []byte, perm os.FileMode) error {
    fullPath := fm.getFullPath(path)
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // 确保父目录存在
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("创建父目录失败: %w", err)
    }
    
    return os.WriteFile(fullPath, content, perm)
}

// CopyFile 复制文件
func (fm *FileManagerImpl) CopyFile(ctx context.Context, src, dst string) error {
    srcPath := fm.getFullPath(src)
    dstPath := fm.getFullPath(dst)
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // 打开源文件
    srcFile, err := os.Open(srcPath)
    if err != nil {
        return fmt.Errorf("打开源文件失败: %w", err)
    }
    defer srcFile.Close()
    
    // 确保目标目录存在
    dstDir := filepath.Dir(dstPath)
    if err := os.MkdirAll(dstDir, 0755); err != nil {
        return fmt.Errorf("创建目标目录失败: %w", err)
    }
    
    // 创建目标文件
    dstFile, err := os.Create(dstPath)
    if err != nil {
        return fmt.Errorf("创建目标文件失败: %w", err)
    }
    defer dstFile.Close()
    
    // 复制内容
    _, err = io.Copy(dstFile, srcFile)
    if err != nil {
        return fmt.Errorf("复制文件内容失败: %w", err)
    }
    
    // 复制权限
    srcInfo, err := srcFile.Stat()
    if err != nil {
        return fmt.Errorf("获取源文件信息失败: %w", err)
    }
    
    return os.Chmod(dstPath, srcInfo.Mode())
}

// CopyDirectory 复制目录
func (fm *FileManagerImpl) CopyDirectory(ctx context.Context, src, dst string) error {
    srcPath := fm.getFullPath(src)
    dstPath := fm.getFullPath(dst)
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // 使用第三方库复制目录
    opt := copy.Options{
        OnSymlink: func(src string) copy.SymlinkAction {
            return copy.Deep
        },
        OnDirExists: func(src, dest string) copy.DirExistsAction {
            return copy.Merge
        },
    }
    
    return copy.Copy(srcPath, dstPath, opt)
}

// MoveFile 移动文件
func (fm *FileManagerImpl) MoveFile(ctx context.Context, src, dst string) error {
    srcPath := fm.getFullPath(src)
    dstPath := fm.getFullPath(dst)
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // 确保目标目录存在
    dstDir := filepath.Dir(dstPath)
    if err := os.MkdirAll(dstDir, 0755); err != nil {
        return fmt.Errorf("创建目标目录失败: %w", err)
    }
    
    return os.Rename(srcPath, dstPath)
}

// DeleteFile 删除文件
func (fm *FileManagerImpl) DeleteFile(ctx context.Context, path string) error {
    fullPath := fm.getFullPath(path)
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    return os.Remove(fullPath)
}

// DeleteDirectory 删除目录
func (fm *FileManagerImpl) DeleteDirectory(ctx context.Context, path string) error {
    fullPath := fm.getFullPath(path)
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    return os.RemoveAll(fullPath)
}

// Exists 检查路径是否存在
func (fm *FileManagerImpl) Exists(path string) bool {
    fullPath := fm.getFullPath(path)
    _, err := os.Stat(fullPath)
    return !os.IsNotExist(err)
}

// IsDirectory 检查是否为目录
func (fm *FileManagerImpl) IsDirectory(path string) bool {
    fullPath := fm.getFullPath(path)
    info, err := os.Stat(fullPath)
    if err != nil {
        return false
    }
    return info.IsDir()
}

// IsFile 检查是否为文件
func (fm *FileManagerImpl) IsFile(path string) bool {
    fullPath := fm.getFullPath(path)
    info, err := os.Stat(fullPath)
    if err != nil {
        return false
    }
    return !info.IsDir()
}

// ListDirectory 列出目录内容
func (fm *FileManagerImpl) ListDirectory(ctx context.Context, path string) ([]os.FileInfo, error) {
    fullPath := fm.getFullPath(path)
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    entries, err := os.ReadDir(fullPath)
    if err != nil {
        return nil, fmt.Errorf("读取目录失败: %w", err)
    }
    
    var fileInfos []os.FileInfo
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue
        }
        fileInfos = append(fileInfos, info)
    }
    
    return fileInfos, nil
}

// ReadFile 读取文件
func (fm *FileManagerImpl) ReadFile(ctx context.Context, path string) ([]byte, error) {
    fullPath := fm.getFullPath(path)
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    return os.ReadFile(fullPath)
}

// WriteFile 写入文件
func (fm *FileManagerImpl) WriteFile(ctx context.Context, path string, content []byte, perm os.FileMode) error {
    fullPath := fm.getFullPath(path)
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // 确保父目录存在
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("创建父目录失败: %w", err)
    }
    
    return os.WriteFile(fullPath, content, perm)
}

// GetFileInfo 获取文件信息
func (fm *FileManagerImpl) GetFileInfo(path string) (os.FileInfo, error) {
    fullPath := fm.getFullPath(path)
    return os.Stat(fullPath)
}

// WalkDirectory 遍历目录
func (fm *FileManagerImpl) WalkDirectory(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
    fullPath := fm.getFullPath(root)
    
    return filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        // 转换为相对路径
        relPath, err := filepath.Rel(fm.basePath, path)
        if err != nil {
            relPath = path
        }
        
        return walkFn(relPath, info, err)
    })
}

// getFullPath 获取完整路径
func (fm *FileManagerImpl) getFullPath(path string) string {
    if filepath.IsAbs(path) {
        return path
    }
    
    if fm.basePath == "" {
        return path
    }
    
    return filepath.Join(fm.basePath, path)
}
```

### 8.3 错误处理设计

#### 8.3.1 错误处理器接口
```go
// internal/infrastructure/errors/handler.go
package errors

import (
    "context"
    "fmt"
    "log"
    "runtime"
    "strings"
    
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/models"
)

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
    // HandleError 处理错误
    HandleError(ctx context.Context, err error) error
    
    // HandlePanic 处理panic
    HandlePanic(ctx context.Context, recovered interface{}) error
    
    // LogError 记录错误
    LogError(ctx context.Context, err error)
    
    // FormatError 格式化错误
    FormatError(err error) string
    
    // IsRetryableError 判断是否可重试错误
    IsRetryableError(err error) bool
    
    // WrapError 包装错误
    WrapError(err error, message string) error
}

// ErrorHandlerImpl 错误处理器实现
type ErrorHandlerImpl struct {
    ui     ui.UI
    logger *log.Logger
    debug  bool
}

// NewErrorHandler 创建错误处理器
func NewErrorHandler(ui ui.UI, logger *log.Logger, debug bool) ErrorHandler {
    return &ErrorHandlerImpl{
        ui:     ui,
        logger: logger,
        debug:  debug,
    }
}

// HandleError 处理错误
func (eh *ErrorHandlerImpl) HandleError(ctx context.Context, err error) error {
    if err == nil {
        return nil
    }
    
    // 记录错误
    eh.LogError(ctx, err)
    
    // 根据错误类型进行不同处理
    switch e := err.(type) {
    case *models.ServiceError:
        return eh.handleServiceError(ctx, e)
    case *models.ValidationError:
        return eh.handleValidationError(ctx, e)
    case *models.NetworkError:
        return eh.handleNetworkError(ctx, e)
    default:
        return eh.handleGenericError(ctx, err)
    }
}

// HandlePanic 处理panic
func (eh *ErrorHandlerImpl) HandlePanic(ctx context.Context, recovered interface{}) error {
    // 获取调用栈
    buf := make([]byte, 4096)
    n := runtime.Stack(buf, false)
    stack := string(buf[:n])
    
    panicErr := fmt.Errorf("panic recovered: %v\nStack trace:\n%s", recovered, stack)
    
    // 记录panic
    eh.LogError(ctx, panicErr)
    
    // 显示用户友好的错误信息
    eh.ui.ShowError("程序遇到了意外错误，请联系技术支持")
    
    return panicErr
}

// LogError 记录错误
func (eh *ErrorHandlerImpl) LogError(ctx context.Context, err error) {
    if eh.logger == nil {
        return
    }
    
    // 获取调用信息
    _, file, line, ok := runtime.Caller(2)
    if ok {
        eh.logger.Printf("ERROR [%s:%d] %v", file, line, err)
    } else {
        eh.logger.Printf("ERROR %v", err)
    }
    
    // 在调试模式下显示详细信息
    if eh.debug {
        eh.ui.ShowError(fmt.Sprintf("DEBUG: %v", err))
    }
}

// FormatError 格式化错误
func (eh *ErrorHandlerImpl) FormatError(err error) string {
    if err == nil {
        return ""
    }
    
    switch e := err.(type) {
    case *models.ServiceError:
        return eh.formatServiceError(e)
    case *models.ValidationError:
        return eh.formatValidationError(e)
    case *models.NetworkError:
        return eh.formatNetworkError(e)
    default:
        return err.Error()
    }
}

// IsRetryableError 判断是否可重试错误
func (eh *ErrorHandlerImpl) IsRetryableError(err error) bool {
    switch e := err.(type) {
    case *models.NetworkError:
        return e.Retryable
    case *models.ServiceError:
        return e.Type == models.ErrorTypeNetwork || e.Type == models.ErrorTypeTimeout
    default:
        // 检查常见的可重试错误
        errMsg := strings.ToLower(err.Error())
        retryableKeywords := []string{
            "timeout",
            "connection refused",
            "connection reset",
            "temporary failure",
            "service unavailable",
        }
        
        for _, keyword := range retryableKeywords {
            if strings.Contains(errMsg, keyword) {
                return true
            }
        }
        
        return false
    }
}

// WrapError 包装错误
func (eh *ErrorHandlerImpl) WrapError(err error, message string) error {
    if err == nil {
        return nil
    }
    
    return fmt.Errorf("%s: %w", message, err)
}

// handleServiceError 处理服务错误
func (eh *ErrorHandlerImpl) handleServiceError(ctx context.Context, err *models.ServiceError) error {
    message := eh.formatServiceError(err)
    
    switch err.Type {
    case models.ErrorTypeValidation:
        eh.ui.ShowWarning(message)
    case models.ErrorTypeNetwork:
        eh.ui.ShowError(message)
        if eh.IsRetryableError(err) {
            eh.ui.ShowInfo("此错误可能是临时的，请稍后重试")
        }
    case models.ErrorTypePermission:
        eh.ui.ShowError(message)
        eh.ui.ShowInfo("请检查文件权限或以管理员身份运行")
    default:
        eh.ui.ShowError(message)
    }
    
    return err
}

// handleValidationError 处理验证错误
func (eh *ErrorHandlerImpl) handleValidationError(ctx context.Context, err *models.ValidationError) error {
    message := eh.formatValidationError(err)
    eh.ui.ShowWarning(message)
    
    // 显示修复建议
    if len(err.Suggestions) > 0 {
        eh.ui.ShowInfo("建议:")
        for _, suggestion := range err.Suggestions {
            eh.ui.ShowInfo(fmt.Sprintf("  - %s", suggestion))
        }
    }
    
    return err
}

// handleNetworkError 处理网络错误
func (eh *ErrorHandlerImpl) handleNetworkError(ctx context.Context, err *models.NetworkError) error {
    message := eh.formatNetworkError(err)
    eh.ui.ShowError(message)
    
    if err.Retryable {
        eh.ui.ShowInfo("网络错误，建议检查网络连接后重试")
    }
    
    return err
}

// handleGenericError 处理通用错误
func (eh *ErrorHandlerImpl) handleGenericError(ctx context.Context, err error) error {
    eh.ui.ShowError(err.Error())
    return err
}

// formatServiceError 格式化服务错误
func (eh *ErrorHandlerImpl) formatServiceError(err *models.ServiceError) string {
    message := err.Message
    if err.Cause != nil {
        message += fmt.Sprintf(" (原因: %v)", err.Cause)
    }
    return message
}

// formatValidationError 格式化验证错误
func (eh *ErrorHandlerImpl) formatValidationError(err *models.ValidationError) string {
    message := fmt.Sprintf("验证失败: %s", err.Message)
    if err.Field != "" {
        message = fmt.Sprintf("字段 '%s' %s", err.Field, err.Message)
    }
    return message
}

// formatNetworkError 格式化网络错误
func (eh *ErrorHandlerImpl) formatNetworkError(err *models.NetworkError) string {
    message := fmt.Sprintf("网络错误: %s", err.Message)
    if err.URL != "" {
        message += fmt.Sprintf(" (URL: %s)", err.URL)
    }
    if err.StatusCode > 0 {
        message += fmt.Sprintf(" (状态码: %d)", err.StatusCode)
    }
    return message
}
```

### 8.4 基础设施模块

#### 8.4.1 模块定义
```go
// internal/infrastructure/module.go
package infrastructure

import (
    "log"
    "os"
    
    "go.uber.org/fx"
    
    "specify-cli-go/internal/infrastructure/config"
    "specify-cli-go/internal/infrastructure/filesystem"
    "specify-cli-go/internal/infrastructure/errors"
    "specify-cli-go/internal/cli/ui"
)

// InfrastructureModule 基础设施模块
var InfrastructureModule = fx.Module("infrastructure",
    // 提供配置管理器
    fx.Provide(func() config.Manager {
        configPath := os.Getenv("SPECIFY_CONFIG_PATH")
        if configPath == "" {
            configPath = "configs/agents.yaml"
        }
        return config.NewManager(configPath)
    }),
    
    // 提供文件管理器
    fx.Provide(func() filesystem.FileManager {
        basePath := os.Getenv("SPECIFY_BASE_PATH")
        if basePath == "" {
            basePath, _ = os.Getwd()
        }
        return filesystem.NewFileManager(basePath)
    }),
    
    // 提供错误处理器
    fx.Provide(func(ui ui.UI) errors.ErrorHandler {
        logger := log.New(os.Stderr, "[SPECIFY] ", log.LstdFlags|log.Lshortfile)
        debug := os.Getenv("SPECIFY_DEBUG") == "true"
        return errors.NewErrorHandler(ui, logger, debug)
    }),
    
    // 生命周期钩子
    fx.Invoke(func(lc fx.Lifecycle,
        configManager config.Manager,
        fileManager filesystem.FileManager,
        errorHandler errors.ErrorHandler) {
        
        lc.Append(fx.Hook{
            OnStart: func(ctx context.Context) error {
                // 初始化配置
                return configManager.LoadConfig(ctx)
            },
            OnStop: func(ctx context.Context) error {
                // 保存配置
                return configManager.SaveConfig(ctx)
            },
        })
    }),
)
 ```

---

## 9. 测试策略详细设计

### 9.1 测试架构设计

#### 9.1.1 测试分层架构
```
测试金字塔架构:
┌─────────────────────────────────────┐
│           E2E Tests (少量)           │  ← 端到端测试
├─────────────────────────────────────┤
│        Integration Tests (中等)      │  ← 集成测试
├─────────────────────────────────────┤
│         Unit Tests (大量)           │  ← 单元测试
└─────────────────────────────────────┘
```

#### 9.1.2 测试目录结构
```
tests/
├── unit/                    # 单元测试
│   ├── cli/                # CLI层测试
│   ├── services/           # 服务层测试
│   ├── components/         # 组件层测试
│   └── infrastructure/     # 基础设施层测试
├── integration/            # 集成测试
│   ├── github/            # GitHub集成测试
│   ├── system/            # 系统集成测试
│   └── config/            # 配置集成测试
├── e2e/                   # 端到端测试
│   ├── init_command/      # 初始化命令测试
│   ├── check_command/     # 检查命令测试
│   └── scenarios/         # 场景测试
├── fixtures/              # 测试数据
│   ├── configs/          # 配置文件
│   ├── repositories/     # 模拟仓库
│   └── responses/        # 模拟响应
├── mocks/                 # Mock对象
│   ├── github/           # GitHub API Mock
│   ├── filesystem/       # 文件系统Mock
│   └── ui/               # UI Mock
└── testutils/            # 测试工具
    ├── helpers.go        # 测试辅助函数
    ├── assertions.go     # 自定义断言
    └── setup.go          # 测试环境设置
```

### 9.2 单元测试设计

#### 9.2.1 测试框架和工具
```go
// 测试依赖
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    "github.com/golang/mock/gomock"
)
```

#### 9.2.2 CLI层单元测试
```go
// tests/unit/cli/commands/init_command_test.go
package commands_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    
    "specify-cli-go/internal/cli/commands"
    "specify-cli-go/internal/services"
    "specify-cli-go/tests/mocks"
)

type InitCommandTestSuite struct {
    suite.Suite
    mockInitService *mocks.MockInitService
    mockUI          *mocks.MockUI
    command         *commands.InitCommand
}

func (suite *InitCommandTestSuite) SetupTest() {
    suite.mockInitService = mocks.NewMockInitService(suite.T())
    suite.mockUI = mocks.NewMockUI(suite.T())
    
    suite.command = commands.NewInitCommand(
        suite.mockInitService,
        suite.mockUI,
    )
}

func (suite *InitCommandTestSuite) TestExecute_Success() {
    // Arrange
    ctx := context.Background()
    args := &commands.InitArgs{
        ProjectName: "test-project",
        AgentType:   "claude",
        OutputDir:   "./output",
        Force:       false,
    }
    
    suite.mockInitService.On("Initialize", ctx, mock.MatchedBy(func(req *services.InitRequest) bool {
        return req.ProjectName == "test-project" &&
               req.AgentType == "claude" &&
               req.OutputDir == "./output"
    })).Return(&services.InitResponse{
        ProjectPath: "./output/test-project",
        FilesCreated: []string{"setup.sh", "README.md"},
    }, nil)
    
    suite.mockUI.On("ShowSuccess", mock.AnythingOfType("string")).Return()
    
    // Act
    err := suite.command.Execute(ctx, args)
    
    // Assert
    assert.NoError(suite.T(), err)
    suite.mockInitService.AssertExpectations(suite.T())
    suite.mockUI.AssertExpectations(suite.T())
}

func (suite *InitCommandTestSuite) TestExecute_ValidationError() {
    // Arrange
    ctx := context.Background()
    args := &commands.InitArgs{
        ProjectName: "", // 空项目名
        AgentType:   "claude",
        OutputDir:   "./output",
    }
    
    suite.mockUI.On("ShowError", mock.AnythingOfType("string")).Return()
    
    // Act
    err := suite.command.Execute(ctx, args)
    
    // Assert
    assert.Error(suite.T(), err)
    assert.Contains(suite.T(), err.Error(), "项目名称不能为空")
    suite.mockUI.AssertExpectations(suite.T())
}

func (suite *InitCommandTestSuite) TestExecute_ServiceError() {
    // Arrange
    ctx := context.Background()
    args := &commands.InitArgs{
        ProjectName: "test-project",
        AgentType:   "claude",
        OutputDir:   "./output",
    }
    
    expectedError := errors.New("初始化失败")
    suite.mockInitService.On("Initialize", ctx, mock.Anything).Return(nil, expectedError)
    suite.mockUI.On("ShowError", mock.AnythingOfType("string")).Return()
    
    // Act
    err := suite.command.Execute(ctx, args)
    
    // Assert
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), expectedError, err)
    suite.mockInitService.AssertExpectations(suite.T())
    suite.mockUI.AssertExpectations(suite.T())
}

func TestInitCommandTestSuite(t *testing.T) {
    suite.Run(t, new(InitCommandTestSuite))
}
```

#### 9.2.3 服务层单元测试
```go
// tests/unit/services/init_service_test.go
package services_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    
    "specify-cli-go/internal/services"
    "specify-cli-go/tests/mocks"
)

type InitServiceTestSuite struct {
    suite.Suite
    mockDownloader    *mocks.MockDownloader
    mockExtractor     *mocks.MockArchiveExtractor
    mockFileManager   *mocks.MockFileManager
    mockConfigManager *mocks.MockConfigManager
    service           services.InitService
}

func (suite *InitServiceTestSuite) SetupTest() {
    suite.mockDownloader = mocks.NewMockDownloader(suite.T())
    suite.mockExtractor = mocks.NewMockArchiveExtractor(suite.T())
    suite.mockFileManager = mocks.NewMockFileManager(suite.T())
    suite.mockConfigManager = mocks.NewMockConfigManager(suite.T())
    
    suite.service = services.NewInitService(
        suite.mockDownloader,
        suite.mockExtractor,
        suite.mockFileManager,
        suite.mockConfigManager,
    )
}

func (suite *InitServiceTestSuite) TestInitialize_Success() {
    // Arrange
    ctx := context.Background()
    req := &services.InitRequest{
        ProjectName: "test-project",
        AgentType:   "claude",
        OutputDir:   "./output",
    }
    
    // Mock配置获取
    agentConfig := &models.AgentConfig{
        Name: "Claude",
        Type: "CLI",
        Requirements: []string{"Claude CLI"},
    }
    suite.mockConfigManager.On("GetAgentConfig", "claude").Return(agentConfig, nil)
    
    // Mock下载
    suite.mockDownloader.On("Download", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
    
    // Mock解压
    suite.mockExtractor.On("Extract", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
    
    // Mock文件操作
    suite.mockFileManager.On("CreateDirectory", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("os.FileMode")).Return(nil)
    suite.mockFileManager.On("CreateFile", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]byte"), mock.AnythingOfType("os.FileMode")).Return(nil)
    
    // Act
    response, err := suite.service.Initialize(ctx, req)
    
    // Assert
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), response)
    assert.Equal(suite.T(), "./output/test-project", response.ProjectPath)
    assert.NotEmpty(suite.T(), response.FilesCreated)
    
    suite.mockConfigManager.AssertExpectations(suite.T())
    suite.mockDownloader.AssertExpectations(suite.T())
    suite.mockExtractor.AssertExpectations(suite.T())
    suite.mockFileManager.AssertExpectations(suite.T())
}

func (suite *InitServiceTestSuite) TestInitialize_InvalidAgentType() {
    // Arrange
    ctx := context.Background()
    req := &services.InitRequest{
        ProjectName: "test-project",
        AgentType:   "invalid-agent",
        OutputDir:   "./output",
    }
    
    suite.mockConfigManager.On("GetAgentConfig", "invalid-agent").Return(nil, errors.New("Agent配置不存在"))
    
    // Act
    response, err := suite.service.Initialize(ctx, req)
    
    // Assert
    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), response)
    assert.Contains(suite.T(), err.Error(), "Agent配置不存在")
    
    suite.mockConfigManager.AssertExpectations(suite.T())
}

func TestInitServiceTestSuite(t *testing.T) {
    suite.Run(t, new(InitServiceTestSuite))
}
```

#### 9.2.4 组件层单元测试
```go
// tests/unit/components/github/downloader_test.go
package github_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    
    "specify-cli-go/internal/components/github"
)

type GitHubDownloaderTestSuite struct {
    suite.Suite
    downloader github.Downloader
    server     *httptest.Server
}

func (suite *GitHubDownloaderTestSuite) SetupTest() {
    suite.downloader = github.NewGitHubDownloader(&github.DownloadOptions{
        Timeout:    30,
        RetryCount: 3,
        UserAgent:  "Specify-CLI-Test/1.0.0",
    })
}

func (suite *InitServiceTestSuite) SetupSuite() {
    // 创建测试HTTP服务器
    suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/repos/test/repo/zipball/main":
            w.Header().Set("Content-Type", "application/zip")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("fake zip content"))
        case "/repos/test/repo":
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{"name": "repo", "full_name": "test/repo"}`))
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
}

func (suite *GitHubDownloaderTestSuite) TearDownSuite() {
    suite.server.Close()
}

func (suite *GitHubDownloaderTestSuite) TestDownload_Success() {
    // Arrange
    ctx := context.Background()
    url := suite.server.URL + "/repos/test/repo/zipball/main"
    outputPath := "./test-output.zip"
    
    // Act
    err := suite.downloader.Download(ctx, url, outputPath)
    
    // Assert
    assert.NoError(suite.T(), err)
    
    // 验证文件是否创建
    // 这里可以添加文件存在性检查
}

func (suite *GitHubDownloaderTestSuite) TestDownload_InvalidURL() {
    // Arrange
    ctx := context.Background()
    url := "invalid-url"
    outputPath := "./test-output.zip"
    
    // Act
    err := suite.downloader.Download(ctx, url, outputPath)
    
    // Assert
    assert.Error(suite.T(), err)
    assert.Contains(suite.T(), err.Error(), "无效的URL")
}

func TestGitHubDownloaderTestSuite(t *testing.T) {
    suite.Run(t, new(GitHubDownloaderTestSuite))
}
```

### 9.3 集成测试设计

#### 9.3.1 GitHub集成测试
```go
// tests/integration/github/github_integration_test.go
package github_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    
    "specify-cli-go/internal/components/github"
    "specify-cli-go/tests/testutils"
)

type GitHubIntegrationTestSuite struct {
    suite.Suite
    tempDir    string
    downloader github.Downloader
    extractor  github.ArchiveExtractor
}

func (suite *GitHubIntegrationTestSuite) SetupSuite() {
    // 创建临时目录
    var err error
    suite.tempDir, err = os.MkdirTemp("", "github-integration-test")
    assert.NoError(suite.T(), err)
    
    // 初始化组件
    suite.downloader = github.NewGitHubDownloader(&github.DownloadOptions{
        Timeout:    60,
        RetryCount: 3,
        UserAgent:  "Specify-CLI-Integration-Test/1.0.0",
    })
    
    suite.extractor = github.NewArchiveExtractor(&github.ExtractOptions{
        FlattenSingleDir: true,
        OverwriteExisting: true,
    })
}

func (suite *GitHubIntegrationTestSuite) TearDownSuite() {
    os.RemoveAll(suite.tempDir)
}

func (suite *GitHubIntegrationTestSuite) TestDownloadAndExtract_RealRepository() {
    // 跳过集成测试（除非设置了环境变量）
    if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
        suite.T().Skip("跳过集成测试，设置 RUN_INTEGRATION_TESTS=true 来运行")
    }
    
    // Arrange
    ctx := context.Background()
    repoURL := "https://github.com/octocat/Hello-World/archive/refs/heads/master.zip"
    downloadPath := filepath.Join(suite.tempDir, "repo.zip")
    extractPath := filepath.Join(suite.tempDir, "extracted")
    
    // Act - 下载
    err := suite.downloader.Download(ctx, repoURL, downloadPath)
    assert.NoError(suite.T(), err)
    
    // 验证下载的文件存在
    assert.FileExists(suite.T(), downloadPath)
    
    // Act - 解压
    err = suite.extractor.Extract(ctx, downloadPath, extractPath)
    assert.NoError(suite.T(), err)
    
    // Assert - 验证解压结果
    assert.DirExists(suite.T(), extractPath)
    
    // 检查是否包含预期的文件
    readmePath := filepath.Join(extractPath, "README")
    assert.FileExists(suite.T(), readmePath)
}

func TestGitHubIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(GitHubIntegrationTestSuite))
}
```

#### 9.3.2 系统集成测试
```go
// tests/integration/system/system_integration_test.go
package system_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    
    "specify-cli-go/internal/components/system"
    "specify-cli-go/tests/testutils"
)

type SystemIntegrationTestSuite struct {
    suite.Suite
    tempDir           string
    permissionManager system.PermissionManager
    specialChecker    system.SpecialChecker
}

func (suite *SystemIntegrationTestSuite) SetupSuite() {
    var err error
    suite.tempDir, err = os.MkdirTemp("", "system-integration-test")
    assert.NoError(suite.T(), err)
    
    suite.permissionManager = system.NewPermissionManager()
    suite.specialChecker = system.NewSpecialToolChecker()
}

func (suite *SystemIntegrationTestSuite) TearDownSuite() {
    os.RemoveAll(suite.tempDir)
}

func (suite *SystemIntegrationTestSuite) TestPermissionManager_SetExecutablePermission() {
    // Arrange
    ctx := context.Background()
    scriptPath := filepath.Join(suite.tempDir, "test-script.sh")
    
    // 创建测试脚本
    err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'test'"), 0644)
    assert.NoError(suite.T(), err)
    
    // Act
    err = suite.permissionManager.SetExecutablePermission(ctx, scriptPath)
    
    // Assert
    assert.NoError(suite.T(), err)
    
    // 验证权限
    isExecutable, err := suite.permissionManager.IsExecutable(scriptPath)
    assert.NoError(suite.T(), err)
    assert.True(suite.T(), isExecutable)
}

func (suite *SystemIntegrationTestSuite) TestSpecialChecker_CheckInstalledTools() {
    // Arrange
    ctx := context.Background()
    
    // Act
    results, err := suite.specialChecker.CheckAll(ctx)
    
    // Assert
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), results)
    
    // 验证结果结构
    for toolName, result := range results {
        assert.NotEmpty(suite.T(), toolName)
        assert.NotNil(suite.T(), result)
        // 结果应该包含是否安装的信息
        assert.Contains(suite.T(), []bool{true, false}, result.Installed)
    }
}

func TestSystemIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(SystemIntegrationTestSuite))
}
```

### 9.4 端到端测试设计

#### 9.4.1 初始化命令E2E测试
```go
// tests/e2e/init_command/init_e2e_test.go
package init_command_test

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    
    "specify-cli-go/tests/testutils"
)

type InitCommandE2ETestSuite struct {
    suite.Suite
    tempDir    string
    binaryPath string
}

func (suite *InitCommandE2ETestSuite) SetupSuite() {
    // 跳过E2E测试（除非设置了环境变量）
    if os.Getenv("RUN_E2E_TESTS") != "true" {
        suite.T().Skip("跳过E2E测试，设置 RUN_E2E_TESTS=true 来运行")
    }
    
    var err error
    suite.tempDir, err = os.MkdirTemp("", "init-e2e-test")
    assert.NoError(suite.T(), err)
    
    // 构建二进制文件
    suite.binaryPath = filepath.Join(suite.tempDir, "specify-cli")
    if testutils.IsWindows() {
        suite.binaryPath += ".exe"
    }
    
    err = suite.buildBinary()
    assert.NoError(suite.T(), err)
}

func (suite *InitCommandE2ETestSuite) TearDownSuite() {
    os.RemoveAll(suite.tempDir)
}

func (suite *InitCommandE2ETestSuite) buildBinary() error {
    cmd := exec.Command("go", "build", "-o", suite.binaryPath, "../../cmd/specify-cli")
    cmd.Dir = suite.tempDir
    return cmd.Run()
}

func (suite *InitCommandE2ETestSuite) TestInitCommand_Success() {
    // Arrange
    projectName := "test-e2e-project"
    outputDir := filepath.Join(suite.tempDir, "output")
    
    // Act
    cmd := exec.Command(suite.binaryPath, "init",
        "--name", projectName,
        "--agent", "claude",
        "--output", outputDir,
    )
    
    output, err := cmd.CombinedOutput()
    
    // Assert
    assert.NoError(suite.T(), err, "命令执行失败: %s", string(output))
    
    // 验证项目目录是否创建
    projectPath := filepath.Join(outputDir, projectName)
    assert.DirExists(suite.T(), projectPath)
    
    // 验证关键文件是否创建
    expectedFiles := []string{
        "setup.sh",
        "README.md",
        ".gitignore",
    }
    
    for _, file := range expectedFiles {
        filePath := filepath.Join(projectPath, file)
        assert.FileExists(suite.T(), filePath, "文件 %s 应该存在", file)
    }
    
    // 验证输出内容
    outputStr := string(output)
    assert.Contains(suite.T(), outputStr, "初始化成功")
    assert.Contains(suite.T(), outputStr, projectName)
}

func (suite *InitCommandE2ETestSuite) TestInitCommand_InvalidArguments() {
    // Act - 缺少必需参数
    cmd := exec.Command(suite.binaryPath, "init")
    output, err := cmd.CombinedOutput()
    
    // Assert
    assert.Error(suite.T(), err)
    
    outputStr := string(output)
    assert.Contains(suite.T(), outputStr, "项目名称不能为空")
}

func (suite *InitCommandE2ETestSuite) TestInitCommand_Help() {
    // Act
    cmd := exec.Command(suite.binaryPath, "init", "--help")
    output, err := cmd.CombinedOutput()
    
    // Assert
    assert.NoError(suite.T(), err)
    
    outputStr := string(output)
    assert.Contains(suite.T(), outputStr, "初始化AI Agent项目")
    assert.Contains(suite.T(), outputStr, "--name")
    assert.Contains(suite.T(), outputStr, "--agent")
    assert.Contains(suite.T(), outputStr, "--output")
}

func TestInitCommandE2ETestSuite(t *testing.T) {
    suite.Run(t, new(InitCommandE2ETestSuite))
}
```

### 9.5 性能测试设计

#### 9.5.1 基准测试
```go
// tests/benchmark/download_benchmark_test.go
package benchmark_test

import (
    "context"
    "testing"
    
    "specify-cli-go/internal/components/github"
)

func BenchmarkGitHubDownloader_Download(b *testing.B) {
    downloader := github.NewGitHubDownloader(&github.DownloadOptions{
        Timeout:    30,
        RetryCount: 1,
        UserAgent:  "Specify-CLI-Benchmark/1.0.0",
    })
    
    ctx := context.Background()
    url := "https://github.com/octocat/Hello-World/archive/refs/heads/master.zip"
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        outputPath := fmt.Sprintf("./benchmark-output-%d.zip", i)
        err := downloader.Download(ctx, url, outputPath)
        if err != nil {
            b.Fatalf("下载失败: %v", err)
        }
        
        // 清理文件
        os.Remove(outputPath)
    }
}

func BenchmarkArchiveExtractor_Extract(b *testing.B) {
    // 准备测试数据
    testZipPath := "./fixtures/test-archive.zip"
    
    extractor := github.NewArchiveExtractor(&github.ExtractOptions{
        FlattenSingleDir:  true,
        OverwriteExisting: true,
    })
    
    ctx := context.Background()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        outputPath := fmt.Sprintf("./benchmark-extract-%d", i)
        err := extractor.Extract(ctx, testZipPath, outputPath)
        if err != nil {
            b.Fatalf("解压失败: %v", err)
        }
        
        // 清理目录
        os.RemoveAll(outputPath)
    }
}
```

### 9.6 测试工具和辅助函数

#### 9.6.1 测试辅助函数
```go
// tests/testutils/helpers.go
package testutils

import (
    "os"
    "runtime"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

// IsWindows 检查是否为Windows系统
func IsWindows() bool {
    return runtime.GOOS == "windows"
}

// CreateTempDir 创建临时目录
func CreateTempDir(t *testing.T, prefix string) string {
    tempDir, err := os.MkdirTemp("", prefix)
    assert.NoError(t, err)
    
    t.Cleanup(func() {
        os.RemoveAll(tempDir)
    })
    
    return tempDir
}

// CreateTempFile 创建临时文件
func CreateTempFile(t *testing.T, dir, pattern string, content []byte) string {
    file, err := os.CreateTemp(dir, pattern)
    assert.NoError(t, err)
    
    if content != nil {
        _, err = file.Write(content)
        assert.NoError(t, err)
    }
    
    err = file.Close()
    assert.NoError(t, err)
    
    t.Cleanup(func() {
        os.Remove(file.Name())
    })
    
    return file.Name()
}

// AssertFileExists 断言文件存在
func AssertFileExists(t *testing.T, path string) {
    _, err := os.Stat(path)
    assert.NoError(t, err, "文件应该存在: %s", path)
}

// AssertDirExists 断言目录存在
func AssertDirExists(t *testing.T, path string) {
    info, err := os.Stat(path)
    assert.NoError(t, err, "目录应该存在: %s", path)
    assert.True(t, info.IsDir(), "路径应该是目录: %s", path)
}

// AssertFileContent 断言文件内容
func AssertFileContent(t *testing.T, path string, expectedContent string) {
    content, err := os.ReadFile(path)
    assert.NoError(t, err)
    assert.Equal(t, expectedContent, string(content))
}
```

#### 9.6.2 自定义断言
```go
// tests/testutils/assertions.go
package testutils

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "specify-cli-go/internal/models"
)

// AssertAgentConfig 断言Agent配置
func AssertAgentConfig(t *testing.T, expected, actual *models.AgentConfig) {
    assert.Equal(t, expected.Name, actual.Name)
    assert.Equal(t, expected.Type, actual.Type)
    assert.Equal(t, expected.Description, actual.Description)
    assert.ElementsMatch(t, expected.Requirements, actual.Requirements)
    assert.ElementsMatch(t, expected.SetupInstructions, actual.SetupInstructions)
}

// AssertInitResponse 断言初始化响应
func AssertInitResponse(t *testing.T, response *services.InitResponse) {
    assert.NotNil(t, response)
    assert.NotEmpty(t, response.ProjectPath)
    assert.NotEmpty(t, response.FilesCreated)
    
    // 验证项目路径存在
    AssertDirExists(t, response.ProjectPath)
    
    // 验证创建的文件存在
    for _, file := range response.FilesCreated {
        AssertFileExists(t, filepath.Join(response.ProjectPath, file))
    }
}
```

### 9.7 测试配置和CI/CD集成

#### 9.7.1 测试配置文件
```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test -v -race -coverprofile=coverage.out ./tests/unit/...
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run integration tests
      run: RUN_INTEGRATION_TESTS=true go test -v ./tests/integration/...
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  e2e-tests:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: go mod download
    
    - name: Build binary
      run: go build -o specify-cli ./cmd/specify-cli
    
    - name: Run E2E tests
      run: RUN_E2E_TESTS=true go test -v ./tests/e2e/...
```

#### 9.7.2 Makefile测试目标
```makefile
# Makefile
.PHONY: test test-unit test-integration test-e2e test-benchmark test-coverage

# 运行所有测试
test: test-unit test-integration

# 单元测试
test-unit:
	go test -v -race ./tests/unit/...

# 集成测试
test-integration:
	RUN_INTEGRATION_TESTS=true go test -v ./tests/integration/...

# 端到端测试
test-e2e:
	RUN_E2E_TESTS=true go test -v ./tests/e2e/...

# 基准测试
test-benchmark:
	go test -bench=. -benchmem ./tests/benchmark/...

# 测试覆盖率
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

# 清理测试文件
test-clean:
	rm -f coverage.out coverage.html
	find . -name "*.test" -delete
	find . -name "*-test-*" -type d -exec rm -rf {} +
```

---

## 10. 构建和部署详细设计

### 10.1 构建系统设计

#### 10.1.1 Makefile构建配置
```makefile
# Makefile
.PHONY: build clean test install lint fmt vet deps help

# 变量定义
BINARY_NAME=specify-cli
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# 默认目标
all: clean fmt vet test build

# 构建二进制文件
build:
	@echo "构建 ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/specify-cli

# 跨平台构建
build-all: clean
	@echo "构建所有平台版本..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd/specify-cli
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 ./cmd/specify-cli
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 ./cmd/specify-cli
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd/specify-cli

# 安装到本地
install:
	@echo "安装 ${BINARY_NAME}..."
	go install ${LDFLAGS} ./cmd/specify-cli

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	go clean

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
vet:
	@echo "运行 go vet..."
	go vet ./...

# 代码规范检查
lint:
	@echo "运行 golangci-lint..."
	golangci-lint run

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 运行测试
test:
	@echo "运行测试..."
	go test -v -race ./...

# 显示帮助
help:
	@echo "可用的make目标:"
	@echo "  build      - 构建二进制文件"
	@echo "  build-all  - 构建所有平台版本"
	@echo "  install    - 安装到本地"
	@echo "  clean      - 清理构建文件"
	@echo "  fmt        - 格式化代码"
	@echo "  vet        - 运行go vet"
	@echo "  lint       - 运行代码规范检查"
	@echo "  deps       - 下载依赖"
	@echo "  test       - 运行测试"
	@echo "  help       - 显示此帮助信息"
```

#### 10.1.2 Go模块配置
```go
// go.mod
module specify-cli-go

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    go.uber.org/fx v1.20.1
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.17.1
    github.com/pterm/pterm v0.12.79
    github.com/manifoldco/promptui v0.9.0
    github.com/go-resty/resty/v2 v2.11.0
    github.com/schollz/progressbar/v3 v3.14.1
    github.com/mholt/archiver/v4 v4.0.0-alpha.8
    github.com/otiai10/copy v1.14.0
    github.com/shirou/gopsutil/v3 v3.23.12
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    golang.org/x/sys v0.16.0
)

require (
    // 间接依赖...
)
```

### 10.2 CI/CD流水线设计

#### 10.2.1 GitHub Actions配置
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'

jobs:
  # 代码质量检查
  quality:
    name: Code Quality
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Run go fmt
      run: |
        if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
          gofmt -s -l .
          exit 1
        fi

    - name: Run go vet
      run: go vet ./...

    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest

  # 测试
  test:
    name: Test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Upload coverage to Codecov
      if: matrix.os == 'ubuntu-latest'
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  # 构建
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [quality, test]
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Build for multiple platforms
      run: make build-all

    - name: Upload build artifacts
      uses: actions/upload-artifact@v3
      with:
        name: binaries
        path: bin/

  # 发布
  release:
    name: Release
    runs-on: ubuntu-latest
    needs: [build]
    if: startsWith(github.ref, 'refs/tags/v')
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Download build artifacts
      uses: actions/download-artifact@v3
      with:
        name: binaries
        path: bin/

    - name: Create Release
      uses: goreleaser/goreleaser-action@v5
      with:
        distribution: goreleaser
        version: latest
        args: release --clean
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### 10.2.2 GoReleaser配置
```yaml
# .goreleaser.yml
project_name: specify-cli

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X main.Version={{.Version}}
      - -X main.BuildTime={{.Date}}
    main: ./cmd/specify-cli

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'

release:
  github:
    owner: your-org
    name: specify-cli-go
  draft: false
  prerelease: auto
  name_template: "{{.ProjectName}} v{{.Version}}"
```

### 10.3 部署策略设计

#### 10.3.1 包管理器集成
```bash
# 安装脚本 - install.sh
#!/bin/bash

set -e

# 检测操作系统和架构
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

# 获取最新版本
LATEST_VERSION=$(curl -s https://api.github.com/repos/your-org/specify-cli-go/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo "无法获取最新版本信息"
    exit 1
fi

# 下载URL
DOWNLOAD_URL="https://github.com/your-org/specify-cli-go/releases/download/${LATEST_VERSION}/specify-cli_${OS}_${ARCH}.tar.gz"

# 创建临时目录
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# 下载并解压
echo "下载 Specify CLI ${LATEST_VERSION}..."
curl -L "$DOWNLOAD_URL" | tar xz

# 安装到系统路径
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo "需要管理员权限安装到 $INSTALL_DIR"
    sudo mv specify-cli "$INSTALL_DIR/"
else
    mv specify-cli "$INSTALL_DIR/"
fi

# 验证安装
if command -v specify-cli >/dev/null 2>&1; then
    echo "Specify CLI 安装成功!"
    specify-cli --version
else
    echo "安装失败"
    exit 1
fi

# 清理
rm -rf "$TMP_DIR"
```

#### 10.3.2 Docker支持
```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o specify-cli ./cmd/specify-cli

FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /root/

COPY --from=builder /app/specify-cli .

ENTRYPOINT ["./specify-cli"]
```

---

## 11. API设计详细说明

### 11.1 内部API接口设计

#### 11.1.1 服务接口定义
```go
// internal/api/interfaces.go
package api

import (
    "context"
    "specify-cli-go/internal/models"
)

// InitAPI 初始化API接口
type InitAPI interface {
    // Initialize 初始化项目
    Initialize(ctx context.Context, req *InitRequest) (*InitResponse, error)
    
    // ValidateRequest 验证请求参数
    ValidateRequest(req *InitRequest) error
    
    // GetSupportedAgents 获取支持的Agent列表
    GetSupportedAgents(ctx context.Context) ([]*models.AgentConfig, error)
}

// CheckAPI 检查API接口
type CheckAPI interface {
    // CheckEnvironment 检查环境
    CheckEnvironment(ctx context.Context, req *CheckRequest) (*CheckResponse, error)
    
    // CheckSpecificTool 检查特定工具
    CheckSpecificTool(ctx context.Context, toolName string) (*ToolCheckResult, error)
    
    // GetCheckHistory 获取检查历史
    GetCheckHistory(ctx context.Context) ([]*CheckRecord, error)
}

// ConfigAPI 配置API接口
type ConfigAPI interface {
    // GetConfig 获取配置
    GetConfig(ctx context.Context, key string) (interface{}, error)
    
    // SetConfig 设置配置
    SetConfig(ctx context.Context, key string, value interface{}) error
    
    // ListConfigs 列出所有配置
    ListConfigs(ctx context.Context) (map[string]interface{}, error)
    
    // ResetConfig 重置配置
    ResetConfig(ctx context.Context, key string) error
}
```

#### 11.1.2 请求响应模型
```go
// internal/api/models.go
package api

import (
    "time"
    "specify-cli-go/internal/models"
)

// InitRequest 初始化请求
type InitRequest struct {
    ProjectName   string            `json:"project_name" validate:"required,min=1,max=100"`
    AgentType     string            `json:"agent_type" validate:"required,oneof=claude copilot cursor gemini"`
    OutputDir     string            `json:"output_dir" validate:"required"`
    Force         bool              `json:"force"`
    Options       map[string]string `json:"options,omitempty"`
    TemplateURL   string            `json:"template_url,omitempty"`
}

// InitResponse 初始化响应
type InitResponse struct {
    Success      bool              `json:"success"`
    ProjectPath  string            `json:"project_path"`
    FilesCreated []string          `json:"files_created"`
    Message      string            `json:"message"`
    Warnings     []string          `json:"warnings,omitempty"`
    NextSteps    []string          `json:"next_steps,omitempty"`
    Duration     time.Duration     `json:"duration"`
}

// CheckRequest 检查请求
type CheckRequest struct {
    AgentType    string   `json:"agent_type,omitempty"`
    ToolNames    []string `json:"tool_names,omitempty"`
    ProjectPath  string   `json:"project_path,omitempty"`
    Detailed     bool     `json:"detailed"`
}

// CheckResponse 检查响应
type CheckResponse struct {
    Success       bool                        `json:"success"`
    OverallStatus string                      `json:"overall_status"` // "healthy", "warning", "error"
    Results       map[string]*ToolCheckResult `json:"results"`
    Summary       *CheckSummary               `json:"summary"`
    Recommendations []string                  `json:"recommendations,omitempty"`
    Duration      time.Duration               `json:"duration"`
}

// ToolCheckResult 工具检查结果
type ToolCheckResult struct {
    ToolName    string            `json:"tool_name"`
    Installed   bool              `json:"installed"`
    Version     string            `json:"version,omitempty"`
    Path        string            `json:"path,omitempty"`
    Status      string            `json:"status"` // "ok", "warning", "error", "not_found"
    Message     string            `json:"message,omitempty"`
    Details     map[string]string `json:"details,omitempty"`
    CheckedAt   time.Time         `json:"checked_at"`
}

// CheckSummary 检查摘要
type CheckSummary struct {
    TotalChecked int `json:"total_checked"`
    Passed       int `json:"passed"`
    Failed       int `json:"failed"`
    Warnings     int `json:"warnings"`
}

// CheckRecord 检查记录
type CheckRecord struct {
    ID        string           `json:"id"`
    Timestamp time.Time        `json:"timestamp"`
    AgentType string           `json:"agent_type"`
    Results   *CheckResponse   `json:"results"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Error   string            `json:"error"`
    Code    string            `json:"code,omitempty"`
    Details map[string]string `json:"details,omitempty"`
}
```

### 11.2 REST API设计（可选扩展）

#### 11.2.1 HTTP服务器设计
```go
// internal/server/server.go
package server

import (
    "context"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
    
    "specify-cli-go/internal/api"
    "specify-cli-go/internal/server/handlers"
    "specify-cli-go/internal/server/middleware"
)

// Server HTTP服务器
type Server struct {
    engine   *gin.Engine
    initAPI  api.InitAPI
    checkAPI api.CheckAPI
    configAPI api.ConfigAPI
}

// ServerParams 服务器依赖参数
type ServerParams struct {
    fx.In
    
    InitAPI   api.InitAPI
    CheckAPI  api.CheckAPI
    ConfigAPI api.ConfigAPI
}

// NewServer 创建新的HTTP服务器
func NewServer(params ServerParams) *Server {
    gin.SetMode(gin.ReleaseMode)
    
    engine := gin.New()
    engine.Use(gin.Recovery())
    engine.Use(middleware.Logger())
    engine.Use(middleware.CORS())
    
    server := &Server{
        engine:    engine,
        initAPI:   params.InitAPI,
        checkAPI:  params.CheckAPI,
        configAPI: params.ConfigAPI,
    }
    
    server.setupRoutes()
    return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
    api := s.engine.Group("/api/v1")
    
    // 初始化相关路由
    init := api.Group("/init")
    {
        init.POST("/", handlers.NewInitHandler(s.initAPI).Initialize)
        init.GET("/agents", handlers.NewInitHandler(s.initAPI).GetSupportedAgents)
    }
    
    // 检查相关路由
    check := api.Group("/check")
    {
        check.POST("/", handlers.NewCheckHandler(s.checkAPI).CheckEnvironment)
        check.GET("/tool/:name", handlers.NewCheckHandler(s.checkAPI).CheckSpecificTool)
        check.GET("/history", handlers.NewCheckHandler(s.checkAPI).GetCheckHistory)
    }
    
    // 配置相关路由
    config := api.Group("/config")
    {
        config.GET("/", handlers.NewConfigHandler(s.configAPI).ListConfigs)
        config.GET("/:key", handlers.NewConfigHandler(s.configAPI).GetConfig)
        config.PUT("/:key", handlers.NewConfigHandler(s.configAPI).SetConfig)
        config.DELETE("/:key", handlers.NewConfigHandler(s.configAPI).ResetConfig)
    }
    
    // 健康检查
    s.engine.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
            "timestamp": time.Now(),
        })
    })
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context, addr string) error {
    srv := &http.Server{
        Addr:    addr,
        Handler: s.engine,
    }
    
    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        srv.Shutdown(shutdownCtx)
    }()
    
    return srv.ListenAndServe()
}
```

#### 11.2.2 API处理器实现
```go
// internal/server/handlers/init_handler.go
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "specify-cli-go/internal/api"
)

// InitHandler 初始化处理器
type InitHandler struct {
    initAPI api.InitAPI
}

// NewInitHandler 创建初始化处理器
func NewInitHandler(initAPI api.InitAPI) *InitHandler {
    return &InitHandler{
        initAPI: initAPI,
    }
}

// Initialize 处理初始化请求
func (h *InitHandler) Initialize(c *gin.Context) {
    var req api.InitRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, api.ErrorResponse{
            Error: "无效的请求参数",
            Code:  "INVALID_REQUEST",
            Details: map[string]string{
                "validation_error": err.Error(),
            },
        })
        return
    }
    
    // 验证请求
    if err := h.initAPI.ValidateRequest(&req); err != nil {
        c.JSON(http.StatusBadRequest, api.ErrorResponse{
            Error: "请求验证失败",
            Code:  "VALIDATION_FAILED",
            Details: map[string]string{
                "validation_error": err.Error(),
            },
        })
        return
    }
    
    // 执行初始化
    resp, err := h.initAPI.Initialize(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, api.ErrorResponse{
            Error: "初始化失败",
            Code:  "INIT_FAILED",
            Details: map[string]string{
                "error": err.Error(),
            },
        })
        return
    }
    
    c.JSON(http.StatusOK, resp)
}

// GetSupportedAgents 获取支持的Agent列表
func (h *InitHandler) GetSupportedAgents(c *gin.Context) {
    agents, err := h.initAPI.GetSupportedAgents(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, api.ErrorResponse{
            Error: "获取Agent列表失败",
            Code:  "GET_AGENTS_FAILED",
            Details: map[string]string{
                "error": err.Error(),
            },
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "agents": agents,
    })
}
```

---

## 12. 配置管理详细设计

### 12.1 配置文件结构设计

#### 12.1.1 主配置文件
```yaml
# config/config.yaml
app:
  name: "Specify CLI"
  version: "1.0.0"
  debug: false
  log_level: "info"
  
# 默认设置
defaults:
  output_dir: "./specify-projects"
  force_overwrite: false
  auto_install_deps: true
  check_updates: true
  
# 网络设置
network:
  timeout: 30s
  retry_count: 3
  user_agent: "Specify-CLI/1.0.0"
  proxy: ""
  
# GitHub设置
github:
  api_url: "https://api.github.com"
  download_timeout: 300s
  max_file_size: "100MB"
  
# UI设置
ui:
  theme: "auto" # auto, light, dark
  progress_style: "bar" # bar, spinner, dots
  color_output: true
  interactive: true
  
# 安全设置
security:
  verify_checksums: true
  allow_insecure: false
  trusted_domains:
    - "github.com"
    - "raw.githubusercontent.com"
```

#### 12.1.2 Agent配置文件
```yaml
# config/agents.yaml
agents:
  claude:
    name: "Claude"
    type: "CLI"
    description: "Anthropic Claude AI Assistant"
    requirements:
      - "Claude CLI"
    setup_instructions:
      - "安装Claude CLI工具"
      - "配置API密钥"
      - "验证连接"
    repository:
      url: "https://github.com/anthropics/claude-cli"
      branch: "main"
      path: "templates/specify"
    environment:
      variables:
        CLAUDE_API_KEY: "your-api-key"
      aliases:
        claude-init: "claude init --template specify"
    
  copilot:
    name: "GitHub Copilot"
    type: "IDE"
    description: "GitHub Copilot AI Assistant"
    requirements:
      - "GitHub Copilot Extension"
      - "VS Code or compatible IDE"
    setup_instructions:
      - "安装GitHub Copilot扩展"
      - "登录GitHub账户"
      - "启用Copilot功能"
    repository:
      url: "https://github.com/github/copilot-templates"
      branch: "main"
      path: "specify"
    environment:
      variables:
        GITHUB_TOKEN: "your-github-token"
      
  cursor:
    name: "Cursor"
    type: "IDE"
    description: "Cursor AI Code Editor"
    requirements:
      - "Cursor Editor"
    setup_instructions:
      - "下载并安装Cursor编辑器"
      - "配置AI功能"
      - "设置项目模板"
    repository:
      url: "https://github.com/cursor-ai/templates"
      branch: "main"
      path: "specify"
    environment:
      variables:
        CURSOR_API_KEY: "your-cursor-key"
        
  gemini:
    name: "Google Gemini"
    type: "API"
    description: "Google Gemini AI"
    requirements:
      - "Google Cloud Account"
      - "Gemini API Access"
    setup_instructions:
      - "创建Google Cloud项目"
      - "启用Gemini API"
      - "获取API密钥"
    repository:
      url: "https://github.com/google/gemini-templates"
      branch: "main"
      path: "specify"
    environment:
      variables:
        GEMINI_API_KEY: "your-gemini-key"
```

#### 12.1.3 脚本类型配置
```yaml
# config/scripts.yaml
script_types:
  bash:
    name: "Bash Script"
    extension: ".sh"
    shebang: "#!/bin/bash"
    platforms:
      - "linux"
      - "darwin"
    template: |
      #!/bin/bash
      set -e
      
      # Specify项目设置脚本
      # 项目名称: {{.ProjectName}}
      # Agent类型: {{.AgentType}}
      
      echo "设置{{.AgentType}}开发环境..."
      
      # 检查依赖
      {{range .Requirements}}
      if ! command -v {{.}} &> /dev/null; then
          echo "错误: {{.}} 未安装"
          exit 1
      fi
      {{end}}
      
      # 设置环境变量
      {{range $key, $value := .Environment.Variables}}
      export {{$key}}="{{$value}}"
      {{end}}
      
      echo "环境设置完成!"
      
  powershell:
    name: "PowerShell Script"
    extension: ".ps1"
    shebang: ""
    platforms:
      - "windows"
    template: |
      # Specify项目设置脚本
      # 项目名称: {{.ProjectName}}
      # Agent类型: {{.AgentType}}
      
      Write-Host "设置{{.AgentType}}开发环境..." -ForegroundColor Green
      
      # 检查依赖
      {{range .Requirements}}
      if (!(Get-Command "{{.}}" -ErrorAction SilentlyContinue)) {
          Write-Error "错误: {{.}} 未安装"
          exit 1
      }
      {{end}}
      
      # 设置环境变量
      {{range $key, $value := .Environment.Variables}}
      $env:{{$key}} = "{{$value}}"
      {{end}}
      
      Write-Host "环境设置完成!" -ForegroundColor Green
```

### 12.2 配置管理实现

#### 12.2.1 配置加载器
```go
// internal/config/loader.go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/spf13/viper"
    "specify-cli-go/internal/models"
)

// Loader 配置加载器
type Loader struct {
    configDir string
    viper     *viper.Viper
}

// NewLoader 创建配置加载器
func NewLoader(configDir string) *Loader {
    v := viper.New()
    v.SetConfigType("yaml")
    v.AutomaticEnv()
    v.SetEnvPrefix("SPECIFY")
    
    return &Loader{
        configDir: configDir,
        viper:     v,
    }
}

// LoadAppConfig 加载应用配置
func (l *Loader) LoadAppConfig() (*models.AppConfig, error) {
    configPath := filepath.Join(l.configDir, "config.yaml")
    
    // 设置默认值
    l.setAppDefaults()
    
    // 加载配置文件
    if _, err := os.Stat(configPath); err == nil {
        l.viper.SetConfigFile(configPath)
        if err := l.viper.ReadInConfig(); err != nil {
            return nil, fmt.Errorf("读取配置文件失败: %w", err)
        }
    }
    
    var config models.AppConfig
    if err := l.viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }
    
    return &config, nil
}

// LoadAgentConfigs 加载Agent配置
func (l *Loader) LoadAgentConfigs() (map[string]*models.AgentConfig, error) {
    agentsPath := filepath.Join(l.configDir, "agents.yaml")
    
    agentViper := viper.New()
    agentViper.SetConfigFile(agentsPath)
    agentViper.SetConfigType("yaml")
    
    if err := agentViper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("读取Agent配置失败: %w", err)
    }
    
    var agentsConfig struct {
        Agents map[string]*models.AgentConfig `yaml:"agents"`
    }
    
    if err := agentViper.Unmarshal(&agentsConfig); err != nil {
        return nil, fmt.Errorf("解析Agent配置失败: %w", err)
    }
    
    return agentsConfig.Agents, nil
}

// LoadScriptConfigs 加载脚本配置
func (l *Loader) LoadScriptConfigs() (map[string]*models.ScriptTypeConfig, error) {
    scriptsPath := filepath.Join(l.configDir, "scripts.yaml")
    
    scriptViper := viper.New()
    scriptViper.SetConfigFile(scriptsPath)
    scriptViper.SetConfigType("yaml")
    
    if err := scriptViper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("读取脚本配置失败: %w", err)
    }
    
    var scriptsConfig struct {
        ScriptTypes map[string]*models.ScriptTypeConfig `yaml:"script_types"`
    }
    
    if err := scriptViper.Unmarshal(&scriptsConfig); err != nil {
        return nil, fmt.Errorf("解析脚本配置失败: %w", err)
    }
    
    return scriptsConfig.ScriptTypes, nil
}

// setAppDefaults 设置应用默认配置
func (l *Loader) setAppDefaults() {
    // 应用设置
    l.viper.SetDefault("app.name", "Specify CLI")
    l.viper.SetDefault("app.version", "1.0.0")
    l.viper.SetDefault("app.debug", false)
    l.viper.SetDefault("app.log_level", "info")
    
    // 默认设置
    l.viper.SetDefault("defaults.output_dir", "./specify-projects")
    l.viper.SetDefault("defaults.force_overwrite", false)
    l.viper.SetDefault("defaults.auto_install_deps", true)
    l.viper.SetDefault("defaults.check_updates", true)
    
    // 网络设置
    l.viper.SetDefault("network.timeout", "30s")
    l.viper.SetDefault("network.retry_count", 3)
    l.viper.SetDefault("network.user_agent", "Specify-CLI/1.0.0")
    
    // GitHub设置
    l.viper.SetDefault("github.api_url", "https://api.github.com")
    l.viper.SetDefault("github.download_timeout", "300s")
    l.viper.SetDefault("github.max_file_size", "100MB")
    
    // UI设置
    l.viper.SetDefault("ui.theme", "auto")
    l.viper.SetDefault("ui.progress_style", "bar")
    l.viper.SetDefault("ui.color_output", true)
    l.viper.SetDefault("ui.interactive", true)
    
    // 安全设置
    l.viper.SetDefault("security.verify_checksums", true)
    l.viper.SetDefault("security.allow_insecure", false)
    l.viper.SetDefault("security.trusted_domains", []string{
        "github.com",
        "raw.githubusercontent.com",
    })
}
```

---

## 13. 项目路线图和扩展计划

### 13.1 版本规划

#### 13.1.1 v1.0.0 - 核心功能（当前版本）
**发布时间**: 2024年Q1

**核心功能**:
- ✅ 基础CLI框架（Cobra + Viper）
- ✅ 初始化命令（init）
- ✅ 环境检查命令（check）
- ✅ 支持4种主要AI Agent（Claude, Copilot, Cursor, Gemini）
- ✅ GitHub仓库下载和解压
- ✅ 基础UI组件（进度条、选择器、横幅）
- ✅ 配置管理系统
- ✅ 错误处理和日志记录
- ✅ 跨平台支持（Windows, macOS, Linux）

**技术特性**:
- 依赖注入框架（Fx）
- 模块化架构设计
- 完整的测试覆盖
- CI/CD流水线
- 多平台构建和发布

#### 13.1.2 v1.1.0 - 增强功能
**发布时间**: 2024年Q2

**新增功能**:
- 🔄 模板管理系统
  - 自定义模板支持
  - 模板版本管理
  - 模板市场集成
- 🔄 插件系统
  - 插件架构设计
  - 第三方插件支持
  - 插件管理命令
- 🔄 配置向导
  - 交互式配置设置
  - 配置验证和建议
  - 配置导入/导出

**改进项目**:
- 性能优化
- 更好的错误提示
- 增强的日志记录
- UI/UX改进

#### 13.1.3 v1.2.0 - 高级功能
**发布时间**: 2024年Q3

**新增功能**:
- 🔄 项目管理
  - 项目列表和状态跟踪
  - 项目更新和同步
  - 批量操作支持
- 🔄 团队协作
  - 团队配置共享
  - 项目模板共享
  - 协作工作流
- 🔄 集成开发环境
  - IDE插件支持
  - 编辑器集成
  - 实时同步功能

#### 13.1.4 v2.0.0 - 重大更新
**发布时间**: 2024年Q4

**重大功能**:
- 🔄 Web界面
  - 基于Web的管理界面
  - 可视化项目管理
  - 实时监控面板
- 🔄 云服务集成
  - 云端配置同步
  - 远程模板仓库
  - 使用分析和统计
- 🔄 AI助手集成
  - 智能项目建议
  - 自动化配置优化
  - 问题诊断和修复

### 13.2 技术路线图

#### 13.2.1 架构演进
```
当前架构 (v1.0)          目标架构 (v2.0)
┌─────────────────┐      ┌─────────────────┐
│   CLI Client    │      │   Web Client    │
├─────────────────┤      ├─────────────────┤
│  Service Layer  │ ---> │  API Gateway    │
├─────────────────┤      ├─────────────────┤
│ Component Layer │      │ Microservices   │
├─────────────────┤      ├─────────────────┤
│Infrastructure   │      │  Cloud Services │
└─────────────────┘      └─────────────────┘
```

#### 13.2.2 技术栈演进
**当前技术栈**:
- Go 1.21+
- Cobra (CLI)
- Viper (配置)
- Fx (依赖注入)
- Lipgloss/Bubbles (UI)

**未来技术栈**:
- Go 1.22+ (后端服务)
- React/TypeScript (Web前端)
- gRPC (服务通信)
- PostgreSQL (数据存储)
- Redis (缓存)
- Docker/Kubernetes (容器化)
- AWS/GCP (云服务)

### 13.3 扩展计划

#### 13.3.1 新Agent支持
```yaml
# 计划支持的新Agent
planned_agents:
  openai:
    name: "OpenAI GPT"
    type: "API"
    priority: "high"
    target_version: "v1.1.0"
    
  azure_openai:
    name: "Azure OpenAI"
    type: "API"
    priority: "medium"
    target_version: "v1.2.0"
    
  huggingface:
    name: "Hugging Face"
    type: "API"
    priority: "medium"
    target_version: "v1.2.0"
    
  local_llm:
    name: "Local LLM"
    type: "Local"
    priority: "low"
    target_version: "v2.0.0"
```

#### 13.3.2 平台扩展
- **移动端支持**: iOS/Android应用
- **浏览器扩展**: Chrome/Firefox插件
- **IDE集成**: VS Code, IntelliJ, Vim插件
- **CI/CD集成**: GitHub Actions, GitLab CI, Jenkins插件

#### 13.3.3 企业功能
- **单点登录(SSO)**: SAML, OAuth2支持
- **权限管理**: 基于角色的访问控制
- **审计日志**: 操作记录和合规性
- **私有部署**: 本地化部署方案

### 13.4 社区建设

#### 13.4.1 开源社区
- **GitHub仓库**: 开源代码和文档
- **贡献指南**: 开发者参与指南
- **问题跟踪**: Bug报告和功能请求
- **讨论论坛**: 社区交流平台

#### 13.4.2 文档和教程
- **用户手册**: 完整的使用指南
- **开发者文档**: API和架构文档
- **视频教程**: 操作演示和最佳实践
- **博客文章**: 技术分享和案例研究

#### 13.4.3 生态系统
- **模板市场**: 社区贡献的项目模板
- **插件商店**: 第三方插件和扩展
- **集成伙伴**: 与其他工具的集成
- **认证计划**: 专业用户认证

---

## 14. 总结

### 14.1 设计文档概述

本详细设计文档全面描述了Specify CLI Go版本的系统架构、技术实现和发展规划。文档涵盖了以下主要方面：

1. **系统架构**: 采用分层架构设计，包括CLI层、服务层、组件层和基础设施层
2. **技术栈**: 基于Go语言，使用现代化的开源库和框架
3. **核心功能**: 实现AI Agent项目的初始化、环境检查和配置管理
4. **质量保证**: 完整的测试策略和CI/CD流水线
5. **扩展性**: 模块化设计支持未来功能扩展

### 14.2 关键设计原则

#### 14.2.1 技术原则
- **接口驱动**: 通过接口定义组件边界，提高可测试性和可维护性
- **依赖注入**: 使用Fx框架实现依赖管理，降低组件耦合度
- **错误处理**: 统一的错误处理机制，提供清晰的错误信息
- **并发安全**: 所有组件都考虑了并发访问的安全性
- **可测试性**: 每个组件都有对应的单元测试和集成测试

#### 14.2.2 用户体验原则
- **简单易用**: 提供直观的命令行界面和交互体验
- **快速响应**: 优化性能，减少用户等待时间
- **清晰反馈**: 提供详细的进度信息和操作结果
- **错误友好**: 当出现错误时，提供有用的诊断信息和解决建议

### 14.3 实施建议

#### 14.3.1 开发阶段
1. **第一阶段**: 实现核心CLI框架和基础组件
2. **第二阶段**: 开发初始化和检查功能
3. **第三阶段**: 完善UI组件和用户体验
4. **第四阶段**: 添加高级功能和优化性能

#### 14.3.2 质量控制
- 每个功能都要有对应的测试用例
- 代码审查和静态分析
- 持续集成和自动化测试
- 性能监控和优化

#### 14.3.3 文档维护
- 保持设计文档与代码同步
- 及时更新API文档
- 维护用户手册和教程
- 记录重要的设计决策和变更

### 14.4 风险评估

#### 14.4.1 技术风险
- **依赖管理**: 第三方库的版本兼容性问题
- **跨平台兼容**: 不同操作系统的行为差异
- **性能瓶颈**: 大文件下载和解压的性能问题
- **安全漏洞**: 网络请求和文件操作的安全风险

#### 14.4.2 缓解措施
- 定期更新依赖库，进行安全扫描
- 在多个平台上进行充分测试
- 实施性能监控和优化
- 遵循安全最佳实践，进行安全审计

### 14.5 成功指标

#### 14.5.1 技术指标
- **代码覆盖率**: 目标90%以上
- **构建时间**: 小于5分钟
- **启动时间**: 小于1秒
- **内存使用**: 小于100MB

#### 14.5.2 用户指标
- **安装成功率**: 目标95%以上
- **用户满意度**: 目标4.5/5.0以上
- **问题解决时间**: 平均小于24小时
- **社区参与度**: 活跃贡献者数量

通过遵循本设计文档的指导，Specify CLI Go版本将成为一个高质量、易用且可扩展的AI Agent项目管理工具，为开发者提供优秀的使用体验。

---

**文档版本**: v1.0.0  
**最后更新**: 2024年1月  
**维护者**: Specify CLI开发团队
```go
// internal/core/services/check_service.go
package services

import (
    "context"
    "fmt"
    "os/exec"
    "runtime"
    "strings"
    
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/infrastructure/config"
    "specify-cli-go/internal/core/checkers"
    "specify-cli-go/internal/cli/ui"
)

// CheckServiceImpl 检查服务实现
type CheckServiceImpl struct {
    config   *config.Manager
    checker  checkers.Checker
    ui       ui.UI
    toolDefs map[string]*models.ToolDefinition
}

// NewCheckService 创建检查服务
func NewCheckService(
    config *config.Manager,
    checker checkers.Checker,
    ui ui.UI,
) CheckService {
    cs := &CheckServiceImpl{
        config:  config,
        checker: checker,
        ui:      ui,
    }
    
    cs.initializeToolDefinitions()
    return cs
}

// CheckAllTools 检查所有工具
func (cs *CheckServiceImpl) CheckAllTools(ctx context.Context) ([]*models.ToolCheckResult, error) {
    var results []*models.ToolCheckResult
    
    // 获取所有工具定义
    for name, toolDef := range cs.toolDefs {
        result, err := cs.checkToolWithDefinition(ctx, name, toolDef)
        if err != nil {
            result = &models.ToolCheckResult{
                Name:      name,
                Installed: false,
                Error:     err,
                Message:   fmt.Sprintf("检查失败: %v", err),
            }
        }
        results = append(results, result)
    }
    
    return results, nil
}

// CheckTool 检查单个工具
func (cs *CheckServiceImpl) CheckTool(ctx context.Context, name string) (*models.ToolCheckResult, error) {
    toolDef, exists := cs.toolDefs[name]
    if !exists {
        return nil, fmt.Errorf("未知工具: %s", name)
    }
    
    return cs.checkToolWithDefinition(ctx, name, toolDef)
}

// checkToolWithDefinition 使用工具定义检查工具
func (cs *CheckServiceImpl) checkToolWithDefinition(ctx context.Context, name string, toolDef *models.ToolDefinition) (*models.ToolCheckResult, error) {
    result := &models.ToolCheckResult{
        Name: name,
    }
    
    // 检查工具是否安装
    installed, version, path, err := cs.checkToolInstallation(toolDef)
    result.Installed = installed
    result.Version = version
    result.Path = path
    
    if err != nil {
        result.Error = err
        result.Message = fmt.Sprintf("检查失败: %v", err)
        return result, nil
    }
    
    if !installed {
        result.Message = "未安装"
        return result, nil
    }
    
    // 检查版本要求
    if toolDef.MinVersion != "" {
        compatible, err := cs.checkVersionCompatibility(version, toolDef.MinVersion)
        if err != nil {
            result.Message = fmt.Sprintf("版本检查失败: %v", err)
        } else if !compatible {
            result.Message = fmt.Sprintf("版本过低，需要 >= %s", toolDef.MinVersion)
        } else {
            result.Message = "已安装且版本兼容"
        }
    } else {
        result.Message = "已安装"
    }
    
    return result, nil
}

// checkToolInstallation 检查工具安装状态
func (cs *CheckServiceImpl) checkToolInstallation(toolDef *models.ToolDefinition) (bool, string, string, error) {
    // 尝试执行版本命令
    cmd := exec.Command(toolDef.Command, toolDef.VersionArgs...)
    output, err := cmd.Output()
    if err != nil {
        return false, "", "", nil // 工具未安装
    }
    
    // 解析版本信息
    version := cs.parseVersion(string(output), toolDef.VersionRegex)
    
    // 获取工具路径
    pathCmd := exec.Command("which", toolDef.Command)
    if runtime.GOOS == "windows" {
        pathCmd = exec.Command("where", toolDef.Command)
    }
    
    pathOutput, err := pathCmd.Output()
    path := ""
    if err == nil {
        path = strings.TrimSpace(string(pathOutput))
    }
    
    return true, version, path, nil
}

// parseVersion 解析版本信息
func (cs *CheckServiceImpl) parseVersion(output, regex string) string {
    if regex == "" {
        // 默认版本解析逻辑
        lines := strings.Split(output, "\n")
        for _, line := range lines {
            line = strings.TrimSpace(line)
            if line != "" {
                // 简单的版本提取
                fields := strings.Fields(line)
                for _, field := range fields {
                    if cs.isVersionString(field) {
                        return field
                    }
                }
            }
        }
        return "unknown"
    }
    
    // 使用正则表达式解析
    re := regexp.MustCompile(regex)
    matches := re.FindStringSubmatch(output)
    if len(matches) > 1 {
        return matches[1]
    }
    
    return "unknown"
}

// isVersionString 判断是否为版本字符串
func (cs *CheckServiceImpl) isVersionString(s string) bool {
    // 简单的版本字符串判断
    return regexp.MustCompile(`^\d+\.\d+`).MatchString(s)
}

// checkVersionCompatibility 检查版本兼容性
func (cs *CheckServiceImpl) checkVersionCompatibility(current, required string) (bool, error) {
    // 简化的版本比较逻辑
    currentParts := strings.Split(current, ".")
    requiredParts := strings.Split(required, ".")
    
    for i := 0; i < len(requiredParts) && i < len(currentParts); i++ {
        currentNum, err1 := strconv.Atoi(currentParts[i])
        requiredNum, err2 := strconv.Atoi(requiredParts[i])
        
        if err1 != nil || err2 != nil {
            return false, fmt.Errorf("版本格式错误")
        }
        
        if currentNum > requiredNum {
            return true, nil
        } else if currentNum < requiredNum {
            return false, nil
        }
    }
    
    return true, nil
}

// GenerateInstallationTips 生成安装建议
func (cs *CheckServiceImpl) GenerateInstallationTips(missingTools []string) []string {
    var tips []string
    
    for _, tool := range missingTools {
        if toolDef, exists := cs.toolDefs[tool]; exists {
            tip := cs.generateToolInstallationTip(tool, toolDef)
            tips = append(tips, tip)
        }
    }
    
    return tips
}

// generateToolInstallationTip 生成工具安装建议
func (cs *CheckServiceImpl) generateToolInstallationTip(name string, toolDef *models.ToolDefinition) string {
    switch runtime.GOOS {
    case "windows":
        if toolDef.InstallCommands.Windows != "" {
            return fmt.Sprintf("%s: %s", name, toolDef.InstallCommands.Windows)
        }
    case "darwin":
        if toolDef.InstallCommands.MacOS != "" {
            return fmt.Sprintf("%s: %s", name, toolDef.InstallCommands.MacOS)
        }
    case "linux":
        if toolDef.InstallCommands.Linux != "" {
            return fmt.Sprintf("%s: %s", name, toolDef.InstallCommands.Linux)
        }
    }
    
    return fmt.Sprintf("%s: 请访问官方网站获取安装指南", name)
}

// FixTool 尝试修复工具
func (cs *CheckServiceImpl) FixTool(ctx context.Context, name string) error {
    toolDef, exists := cs.toolDefs[name]
    if !exists {
        return fmt.Errorf("未知工具: %s", name)
    }
    
    // 获取当前平台的安装命令
    var installCmd string
    switch runtime.GOOS {
    case "windows":
        installCmd = toolDef.InstallCommands.Windows
    case "darwin":
        installCmd = toolDef.InstallCommands.MacOS
    case "linux":
        installCmd = toolDef.InstallCommands.Linux
    }
    
    if installCmd == "" {
        return fmt.Errorf("当前平台不支持自动安装")
    }
    
    // 执行安装命令
    cs.ui.ShowInfo(fmt.Sprintf("正在安装 %s...", name))
    
    parts := strings.Fields(installCmd)
    cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("安装失败: %v\n输出: %s", err, string(output))
    }
    
    cs.ui.ShowSuccess(fmt.Sprintf("%s 安装成功", name))
    return nil
}

// initializeToolDefinitions 初始化工具定义
func (cs *CheckServiceImpl) initializeToolDefinitions() {
    cs.toolDefs = map[string]*models.ToolDefinition{
        "git": {
            Command:     "git",
            VersionArgs: []string{"--version"},
            VersionRegex: `git version (\d+\.\d+\.\d+)`,
            MinVersion:  "2.0.0",
            InstallCommands: models.PlatformCommands{
                Windows: "winget install Git.Git",
                MacOS:   "brew install git",
                Linux:   "sudo apt-get install git",
            },
        },
        "node": {
            Command:     "node",
            VersionArgs: []string{"--version"},
            VersionRegex: `v(\d+\.\d+\.\d+)`,
            MinVersion:  "14.0.0",
            InstallCommands: models.PlatformCommands{
                Windows: "winget install OpenJS.NodeJS",
                MacOS:   "brew install node",
                Linux:   "sudo apt-get install nodejs npm",
            },
        },
        "python": {
            Command:     "python",
            VersionArgs: []string{"--version"},
            VersionRegex: `Python (\d+\.\d+\.\d+)`,
            MinVersion:  "3.8.0",
            InstallCommands: models.PlatformCommands{
                Windows: "winget install Python.Python.3",
                MacOS:   "brew install python",
                Linux:   "sudo apt-get install python3 python3-pip",
            },
        },
        "go": {
            Command:     "go",
            VersionArgs: []string{"version"},
            VersionRegex: `go(\d+\.\d+\.\d+)`,
            MinVersion:  "1.19.0",
            InstallCommands: models.PlatformCommands{
                Windows: "winget install GoLang.Go",
                MacOS:   "brew install go",
                Linux:   "sudo apt-get install golang-go",
            },
        },
        "docker": {
            Command:     "docker",
            VersionArgs: []string{"--version"},
            VersionRegex: `Docker version (\d+\.\d+\.\d+)`,
            MinVersion:  "20.0.0",
            InstallCommands: models.PlatformCommands{
                Windows: "winget install Docker.DockerDesktop",
                MacOS:   "brew install --cask docker",
                Linux:   "sudo apt-get install docker.io",
            },
        },
    }
}
```

### 4.3 模板服务设计

#### 4.3.1 模板服务实现
```go
// internal/core/services/template_service.go
package services

import (
    "context"
    "fmt"
    "path/filepath"
    
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/infrastructure/config"
    "specify-cli-go/internal/core/downloaders"
    "specify-cli-go/internal/core/extractors"
)

// TemplateServiceImpl 模板服务实现
type TemplateServiceImpl struct {
    config     *config.Manager
    downloader downloaders.Downloader
    extractor  extractors.ArchiveExtractor
    templates  map[string]*models.TemplateInfo
}

// NewTemplateService 创建模板服务
func NewTemplateService(
    config *config.Manager,
    downloader downloaders.Downloader,
    extractor extractors.ArchiveExtractor,
) TemplateService {
    ts := &TemplateServiceImpl{
        config:     config,
        downloader: downloader,
        extractor:  extractor,
    }
    
    ts.initializeTemplates()
    return ts
}

// DownloadTemplate 下载模板
func (ts *TemplateServiceImpl) DownloadTemplate(ctx context.Context, ai string, dest string) error {
    templateInfo, err := ts.GetTemplateInfo(ai)
    if err != nil {
        return fmt.Errorf("获取模板信息失败: %w", err)
    }
    
    // 下载模板文件
    tempFile := filepath.Join(dest, "template.zip")
    if err := ts.downloader.Download(templateInfo.DownloadURL, tempFile); err != nil {
        return fmt.Errorf("下载模板失败: %w", err)
    }
    
    // 解压模板
    if err := ts.extractor.Extract(tempFile, dest); err != nil {
        return fmt.Errorf("解压模板失败: %w", err)
    }
    
    return nil
}

// GetTemplateInfo 获取模板信息
func (ts *TemplateServiceImpl) GetTemplateInfo(ai string) (*models.TemplateInfo, error) {
    templateInfo, exists := ts.templates[ai]
    if !exists {
        return nil, fmt.Errorf("不支持的AI助手: %s", ai)
    }
    
    return templateInfo, nil
}

// ListTemplates 列出所有模板
func (ts *TemplateServiceImpl) ListTemplates() ([]*models.TemplateInfo, error) {
    var templates []*models.TemplateInfo
    for _, template := range ts.templates {
        templates = append(templates, template)
    }
    return templates, nil
}

// initializeTemplates 初始化模板定义
func (ts *TemplateServiceImpl) initializeTemplates() {
    ts.templates = map[string]*models.TemplateInfo{
        "claude": {
            Name:        "Claude Code",
            Description: "Claude AI助手项目模板",
            Version:     "1.0.0",
            DownloadURL: "https://github.com/specify-ai/claude-template/archive/main.zip",
            Author:      "Specify AI",
            Tags:        []string{"claude", "ai", "assistant"},
            Requirements: []string{"python>=3.8", "anthropic"},
        },
        "copilot": {
            Name:        "GitHub Copilot",
            Description: "GitHub Copilot项目模板",
            Version:     "1.0.0",
            DownloadURL: "https://github.com/specify-ai/copilot-template/archive/main.zip",
            Author:      "Specify AI",
            Tags:        []string{"copilot", "github", "ai"},
            Requirements: []string{"node>=14", "@github/copilot"},
        },
        "gemini": {
            Name:        "Gemini CLI",
            Description: "Google Gemini AI项目模板",
            Version:     "1.0.0",
            DownloadURL: "https://github.com/specify-ai/gemini-template/archive/main.zip",
            Author:      "Specify AI",
            Tags:        []string{"gemini", "google", "ai"},
            Requirements: []string{"python>=3.8", "google-generativeai"},
        },
        "cursor-agent": {
            Name:        "Cursor Agent",
            Description: "Cursor编辑器AI助手模板",
            Version:     "1.0.0",
            DownloadURL: "https://github.com/specify-ai/cursor-template/archive/main.zip",
            Author:      "Specify AI",
            Tags:        []string{"cursor", "editor", "ai"},
            Requirements: []string{"node>=14", "cursor-api"},
        },
        "qwen": {
            Name:        "Qwen Code",
            Description: "阿里云通义千问代码助手模板",
            Version:     "1.0.0",
            DownloadURL: "https://github.com/specify-ai/qwen-template/archive/main.zip",
            Author:      "Specify AI",
            Tags:        []string{"qwen", "alibaba", "ai"},
            Requirements: []string{"python>=3.8", "dashscope"},
        },
    }
}
```

### 4.4 服务层集成和依赖管理

#### 4.4.1 服务模块定义
```go
// internal/core/services/module.go
package services

import (
    "go.uber.org/fx"
)

// ServicesModule 服务层模块
var ServicesModule = fx.Module("services",
    // 服务接口实现
    fx.Provide(
        fx.Annotate(
            NewInitService,
            fx.As(new(InitService)),
        ),
    ),
    fx.Provide(
        fx.Annotate(
            NewCheckService,
            fx.As(new(CheckService)),
        ),
    ),
    fx.Provide(
        fx.Annotate(
            NewTemplateService,
            fx.As(new(TemplateService)),
        ),
    ),
    
    // 服务配置
    fx.Invoke(RegisterServiceHooks),
)

// RegisterServiceHooks 注册服务钩子
func RegisterServiceHooks(
    lifecycle fx.Lifecycle,
    initService InitService,
    checkService CheckService,
    templateService TemplateService,
) {
    lifecycle.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            // 服务启动时的初始化逻辑
            return nil
        },
        OnStop: func(ctx context.Context) error {
            // 服务停止时的清理逻辑
            return nil
        },
    })
}
```

#### 4.4.2 服务层错误处理
```go
// internal/core/services/errors.go
package services

import (
    "fmt"
)

// ServiceError 服务层错误
type ServiceError struct {
    Type    ErrorType
    Message string
    Cause   error
}

// ErrorType 错误类型
type ErrorType string

const (
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypeNotFound      ErrorType = "not_found"
    ErrorTypeNetwork       ErrorType = "network"
    ErrorTypeFileSystem    ErrorType = "filesystem"
    ErrorTypePermission    ErrorType = "permission"
    ErrorTypeConfiguration ErrorType = "configuration"
)

// Error 实现error接口
func (se *ServiceError) Error() string {
    if se.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", se.Type, se.Message, se.Cause)
    }
    return fmt.Sprintf("[%s] %s", se.Type, se.Message)
}

// Unwrap 解包错误
func (se *ServiceError) Unwrap() error {
    return se.Cause
}

// NewValidationError 创建验证错误
func NewValidationError(message string, cause error) *ServiceError {
    return &ServiceError{
        Type:    ErrorTypeValidation,
        Message: message,
        Cause:   cause,
    }
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(message string, cause error) *ServiceError {
    return &ServiceError{
        Type:    ErrorTypeNotFound,
        Message: message,
        Cause:   cause,
    }
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, cause error) *ServiceError {
    return &ServiceError{
        Type:    ErrorTypeNetwork,
        Message: message,
        Cause:   cause,
    }
}

// NewFileSystemError 创建文件系统错误
func NewFileSystemError(message string, cause error) *ServiceError {
    return &ServiceError{
        Type:    ErrorTypeFileSystem,
        Message: message,
        Cause:   cause,
    }
}

// NewPermissionError 创建权限错误
func NewPermissionError(message string, cause error) *ServiceError {
    return &ServiceError{
        Type:    ErrorTypePermission,
        Message: message,
        Cause:   cause,
    }
}

// NewConfigurationError 创建配置错误
func NewConfigurationError(message string, cause error) *ServiceError {
    return &ServiceError{
        Type:    ErrorTypeConfiguration,
        Message: message,
        Cause:   cause,
    }
}
```

---

## 5. UI组件详细设计

### 5.1 步骤跟踪器设计

#### 5.1.1 步骤跟踪器接口
```go
// internal/cli/ui/interfaces.go
package ui

import (
    "specify-cli-go/internal/models"
)

// StepTracker 步骤跟踪器接口
type StepTracker interface {
    // SetSteps 设置步骤列表
    SetSteps(steps []*models.Step)
    
    // StartStep 开始执行步骤
    StartStep(stepID string)
    
    // UpdateStepProgress 更新步骤进度
    UpdateStepProgress(stepID string, progress float64)
    
    // CompleteStep 完成步骤
    CompleteStep(stepID string)
    
    // FailStep 步骤失败
    FailStep(stepID string, err error)
    
    // GetCurrentStep 获取当前步骤
    GetCurrentStep() *models.Step
    
    // GetProgress 获取总体进度
    GetProgress() float64
    
    // Render 渲染步骤跟踪器
    Render() error
}

// Selector 选择器接口
type Selector interface {
    // Select 显示选择列表
    Select(prompt string, options []string) (string, error)
    
    // MultiSelect 多选
    MultiSelect(prompt string, options []string) ([]string, error)
    
    // Confirm 确认对话框
    Confirm(prompt string, defaultValue bool) (bool, error)
}

// Banner 横幅接口
type Banner interface {
    // Show 显示横幅
    Show() error
    
    // ShowWithMessage 显示带消息的横幅
    ShowWithMessage(message string) error
}

// UI 统一UI接口
type UI interface {
    StepTracker
    Selector
    Banner
    
    // PromptInput 输入提示
    PromptInput(prompt, defaultValue string) (string, error)
    
    // PromptSelect 选择提示
    PromptSelect(prompt string, options []string) (string, error)
    
    // ShowInfo 显示信息
    ShowInfo(message string)
    
    // ShowSuccess 显示成功信息
    ShowSuccess(message string)
    
    // ShowWarning 显示警告信息
    ShowWarning(message string)
    
    // ShowError 显示错误信息
    ShowError(message string)
    
    // NewTable 创建表格
    NewTable() Table
}

// Table 表格接口
type Table interface {
    // SetHeader 设置表头
    SetHeader(headers []string)
    
    // Append 添加行
    Append(row []string)
    
    // Render 渲染表格
    Render()
}
```

### 5.2 步骤跟踪器实现

#### 5.2.1 步骤跟踪器核心实现
```go
// internal/cli/ui/step_tracker.go
package ui

import (
    "fmt"
    "strings"
    "sync"
    "time"
    
    "github.com/charmbracelet/lipgloss"
    "github.com/pterm/pterm"
    "specify-cli-go/internal/models"
)

// StepTrackerImpl 步骤跟踪器实现
type StepTrackerImpl struct {
    steps       []*models.Step
    currentStep *models.Step
    mutex       sync.RWMutex
    style       *StepTrackerStyle
    liveMode    bool
    spinner     *pterm.SpinnerPrinter
}

// StepTrackerStyle 步骤跟踪器样式
type StepTrackerStyle struct {
    PendingStyle   lipgloss.Style
    RunningStyle   lipgloss.Style
    CompletedStyle lipgloss.Style
    FailedStyle    lipgloss.Style
    ProgressStyle  lipgloss.Style
}

// NewStepTracker 创建步骤跟踪器
func NewStepTracker() StepTracker {
    return &StepTrackerImpl{
        steps: make([]*models.Step, 0),
        style: newStepTrackerStyle(),
    }
}

// newStepTrackerStyle 创建步骤跟踪器样式
func newStepTrackerStyle() *StepTrackerStyle {
    return &StepTrackerStyle{
        PendingStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).
            PaddingLeft(2),
        RunningStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("33")).
            Bold(true).
            PaddingLeft(2),
        CompletedStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("42")).
            Bold(true).
            PaddingLeft(2),
        FailedStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Bold(true).
            PaddingLeft(2),
        ProgressStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("33")).
            Bold(true),
    }
}

// SetSteps 设置步骤列表
func (st *StepTrackerImpl) SetSteps(steps []*models.Step) {
    st.mutex.Lock()
    defer st.mutex.Unlock()
    
    st.steps = steps
    for _, step := range st.steps {
        step.Status = models.StepStatusPending
        step.Progress = 0.0
    }
}

// StartStep 开始执行步骤
func (st *StepTrackerImpl) StartStep(stepID string) {
    st.mutex.Lock()
    defer st.mutex.Unlock()
    
    for _, step := range st.steps {
        if step.ID == stepID {
            step.Status = models.StepStatusRunning
            step.StartTime = time.Now()
            st.currentStep = step
            
            // 启动实时模式
            if st.liveMode {
                st.startSpinner(step.Name)
            }
            break
        }
    }
    
    st.render()
}

// UpdateStepProgress 更新步骤进度
func (st *StepTrackerImpl) UpdateStepProgress(stepID string, progress float64) {
    st.mutex.Lock()
    defer st.mutex.Unlock()
    
    for _, step := range st.steps {
        if step.ID == stepID {
            step.Progress = progress
            break
        }
    }
    
    if st.liveMode {
        st.render()
    }
}

// CompleteStep 完成步骤
func (st *StepTrackerImpl) CompleteStep(stepID string) {
    st.mutex.Lock()
    defer st.mutex.Unlock()
    
    for _, step := range st.steps {
        if step.ID == stepID {
            step.Status = models.StepStatusCompleted
            step.Progress = 1.0
            step.EndTime = time.Now()
            
            if st.currentStep == step {
                st.currentStep = nil
            }
            
            // 停止旋转器
            if st.spinner != nil {
                st.spinner.Success(fmt.Sprintf("✓ %s", step.Name))
                st.spinner = nil
            }
            break
        }
    }
    
    st.render()
}

// FailStep 步骤失败
func (st *StepTrackerImpl) FailStep(stepID string, err error) {
    st.mutex.Lock()
    defer st.mutex.Unlock()
    
    for _, step := range st.steps {
        if step.ID == stepID {
            step.Status = models.StepStatusFailed
            step.Error = err
            step.EndTime = time.Now()
            
            if st.currentStep == step {
                st.currentStep = nil
            }
            
            // 停止旋转器并显示错误
            if st.spinner != nil {
                st.spinner.Fail(fmt.Sprintf("✗ %s: %v", step.Name, err))
                st.spinner = nil
            }
            break
        }
    }
    
    st.render()
}

// GetCurrentStep 获取当前步骤
func (st *StepTrackerImpl) GetCurrentStep() *models.Step {
    st.mutex.RLock()
    defer st.mutex.RUnlock()
    
    return st.currentStep
}

// GetProgress 获取总体进度
func (st *StepTrackerImpl) GetProgress() float64 {
    st.mutex.RLock()
    defer st.mutex.RUnlock()
    
    if len(st.steps) == 0 {
        return 0.0
    }
    
    var totalProgress float64
    for _, step := range st.steps {
        totalProgress += step.Progress
    }
    
    return totalProgress / float64(len(st.steps))
}

// Render 渲染步骤跟踪器
func (st *StepTrackerImpl) Render() error {
    st.mutex.RLock()
    defer st.mutex.RUnlock()
    
    return st.render()
}

// render 内部渲染方法
func (st *StepTrackerImpl) render() error {
    if len(st.steps) == 0 {
        return nil
    }
    
    var output strings.Builder
    
    // 渲染总体进度
    progress := st.GetProgress()
    progressBar := st.renderProgressBar(progress)
    output.WriteString(st.style.ProgressStyle.Render(progressBar))
    output.WriteString("\n\n")
    
    // 渲染每个步骤
    for i, step := range st.steps {
        stepLine := st.renderStep(step, i+1)
        output.WriteString(stepLine)
        output.WriteString("\n")
    }
    
    fmt.Print(output.String())
    return nil
}

// renderStep 渲染单个步骤
func (st *StepTrackerImpl) renderStep(step *models.Step, index int) string {
    var icon, status string
    var style lipgloss.Style
    
    switch step.Status {
    case models.StepStatusPending:
        icon = "○"
        status = "待执行"
        style = st.style.PendingStyle
    case models.StepStatusRunning:
        icon = "●"
        status = "执行中"
        style = st.style.RunningStyle
    case models.StepStatusCompleted:
        icon = "✓"
        status = "已完成"
        style = st.style.CompletedStyle
    case models.StepStatusFailed:
        icon = "✗"
        status = "失败"
        style = st.style.FailedStyle
    }
    
    stepText := fmt.Sprintf("%s %d. %s", icon, index, step.Name)
    
    // 添加进度信息
    if step.Status == models.StepStatusRunning && step.Progress > 0 {
        progressPercent := int(step.Progress * 100)
        stepText += fmt.Sprintf(" (%d%%)", progressPercent)
    }
    
    // 添加耗时信息
    if step.Status == models.StepStatusCompleted || step.Status == models.StepStatusFailed {
        if !step.StartTime.IsZero() && !step.EndTime.IsZero() {
            duration := step.EndTime.Sub(step.StartTime)
            stepText += fmt.Sprintf(" (%v)", duration.Round(time.Millisecond))
        }
    }
    
    // 添加错误信息
    if step.Status == models.StepStatusFailed && step.Error != nil {
        stepText += fmt.Sprintf("\n    错误: %v", step.Error)
    }
    
    return style.Render(stepText)
}

// renderProgressBar 渲染进度条
func (st *StepTrackerImpl) renderProgressBar(progress float64) string {
    const barWidth = 40
    filledWidth := int(progress * barWidth)
    
    var bar strings.Builder
    bar.WriteString("进度: [")
    
    for i := 0; i < barWidth; i++ {
        if i < filledWidth {
            bar.WriteString("█")
        } else {
            bar.WriteString("░")
        }
    }
    
    bar.WriteString(fmt.Sprintf("] %.1f%%", progress*100))
    return bar.String()
}

// startSpinner 启动旋转器
func (st *StepTrackerImpl) startSpinner(message string) {
    st.spinner, _ = pterm.DefaultSpinner.Start(message)
}

// StartLiveMode 启动实时模式
func (st *StepTrackerImpl) StartLiveMode() {
    st.mutex.Lock()
    defer st.mutex.Unlock()
    
    st.liveMode = true
}

// StopLiveMode 停止实时模式
func (st *StepTrackerImpl) StopLiveMode() {
    st.mutex.Lock()
    defer st.mutex.Unlock()
    
    st.liveMode = false
    if st.spinner != nil {
        st.spinner.Stop()
        st.spinner = nil
    }
}
```

### 5.3 选择器实现

#### 5.3.1 选择器核心实现
```go
// internal/cli/ui/selector.go
package ui

import (
    "fmt"
    "strings"
    
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// SelectorImpl 选择器实现
type SelectorImpl struct {
    style *SelectorStyle
}

// SelectorStyle 选择器样式
type SelectorStyle struct {
    SelectedStyle   lipgloss.Style
    UnselectedStyle lipgloss.Style
    PromptStyle     lipgloss.Style
    HelpStyle       lipgloss.Style
}

// NewSelector 创建选择器
func NewSelector() Selector {
    return &SelectorImpl{
        style: newSelectorStyle(),
    }
}

// newSelectorStyle 创建选择器样式
func newSelectorStyle() *SelectorStyle {
    return &SelectorStyle{
        SelectedStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("42")).
            Bold(true).
            PaddingLeft(2),
        UnselectedStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).
            PaddingLeft(2),
        PromptStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("33")).
            Bold(true),
        HelpStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).
            Italic(true),
    }
}

// Select 显示选择列表
func (s *SelectorImpl) Select(prompt string, options []string) (string, error) {
    if len(options) == 0 {
        return "", fmt.Errorf("选项列表不能为空")
    }
    
    // 创建列表项
    items := make([]list.Item, len(options))
    for i, option := range options {
        items[i] = listItem{title: option, desc: ""}
    }
    
    // 创建列表模型
    l := list.New(items, itemDelegate{}, 80, 14)
    l.Title = prompt
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(true)
    l.Styles.Title = s.style.PromptStyle
    l.Styles.PaginationStyle = s.style.HelpStyle
    l.Styles.HelpStyle = s.style.HelpStyle
    
    // 创建程序模型
    m := selectModel{
        list:   l,
        choice: "",
        quit:   false,
    }
    
    // 运行程序
    p := tea.NewProgram(m)
    finalModel, err := p.Run()
    if err != nil {
        return "", fmt.Errorf("选择器运行失败: %w", err)
    }
    
    result := finalModel.(selectModel)
    if result.quit {
        return "", fmt.Errorf("用户取消选择")
    }
    
    return result.choice, nil
}

// MultiSelect 多选
func (s *SelectorImpl) MultiSelect(prompt string, options []string) ([]string, error) {
    if len(options) == 0 {
        return nil, fmt.Errorf("选项列表不能为空")
    }
    
    // 创建多选项
    items := make([]multiSelectItem, len(options))
    for i, option := range options {
        items[i] = multiSelectItem{
            title:    option,
            selected: false,
        }
    }
    
    // 创建程序模型
    m := multiSelectModel{
        items:    items,
        cursor:   0,
        selected: make(map[int]bool),
        quit:     false,
        prompt:   prompt,
        style:    s.style,
    }
    
    // 运行程序
    p := tea.NewProgram(m)
    finalModel, err := p.Run()
    if err != nil {
        return nil, fmt.Errorf("多选器运行失败: %w", err)
    }
    
    result := finalModel.(multiSelectModel)
    if result.quit {
        return nil, fmt.Errorf("用户取消选择")
    }
    
    // 收集选中的项目
    var selected []string
    for i, item := range result.items {
        if result.selected[i] {
            selected = append(selected, item.title)
        }
    }
    
    return selected, nil
}

// Confirm 确认对话框
func (s *SelectorImpl) Confirm(prompt string, defaultValue bool) (bool, error) {
    defaultText := "N"
    if defaultValue {
        defaultText = "Y"
    }
    
    fullPrompt := fmt.Sprintf("%s [y/N]", prompt)
    if defaultValue {
        fullPrompt = fmt.Sprintf("%s [Y/n]", prompt)
    }
    
    // 创建文本输入模型
    ti := textinput.New()
    ti.Placeholder = defaultText
    ti.Focus()
    ti.CharLimit = 1
    ti.Width = 20
    
    // 创建程序模型
    m := confirmModel{
        textInput:    ti,
        prompt:       fullPrompt,
        defaultValue: defaultValue,
        quit:         false,
        result:       defaultValue,
        style:        s.style,
    }
    
    // 运行程序
    p := tea.NewProgram(m)
    finalModel, err := p.Run()
    if err != nil {
        return false, fmt.Errorf("确认对话框运行失败: %w", err)
    }
    
    result := finalModel.(confirmModel)
    if result.quit {
        return false, fmt.Errorf("用户取消确认")
    }
    
    return result.result, nil
}
```

### 5.4 横幅实现

#### 5.4.1 横幅核心实现
```go
// internal/cli/ui/banner.go
package ui

import (
    "fmt"
    "os"
    "strings"
    "time"
    
    "github.com/charmbracelet/lipgloss"
    "golang.org/x/term"
)

// BannerImpl 横幅实现
type BannerImpl struct {
    style  *BannerStyle
    config *BannerConfig
}

// BannerStyle 横幅样式
type BannerStyle struct {
    TitleStyle       lipgloss.Style
    SubtitleStyle    lipgloss.Style
    VersionStyle     lipgloss.Style
    BorderStyle      lipgloss.Style
    MessageStyle     lipgloss.Style
    TimestampStyle   lipgloss.Style
}

// BannerConfig 横幅配置
type BannerConfig struct {
    Title       string
    Subtitle    string
    Version     string
    Width       int
    ShowTime    bool
    ShowBorder  bool
    ColorOutput bool
}

// NewBanner 创建横幅
func NewBanner() Banner {
    return &BannerImpl{
        style:  newBannerStyle(),
        config: newBannerConfig(),
    }
}

// Show 显示横幅
func (b *BannerImpl) Show() error {
    banner := b.renderBanner("")
    fmt.Print(banner)
    return nil
}

// ShowWithMessage 显示带消息的横幅
func (b *BannerImpl) ShowWithMessage(message string) error {
    banner := b.renderBanner(message)
    fmt.Print(banner)
    return nil
}
```

### 5.5 统一UI实现

#### 5.5.1 UI组合实现
```go
// internal/cli/ui/ui.go
package ui

import (
    "fmt"
    "os"
    
    "github.com/olekukonko/tablewriter"
    "github.com/pterm/pterm"
)

// UIImpl 统一UI实现
type UIImpl struct {
    StepTracker
    Selector
    Banner
    colorOutput bool
}

// NewUI 创建统一UI
func NewUI() UI {
    return &UIImpl{
        StepTracker: NewStepTracker(),
        Selector:    NewSelector(),
        Banner:      NewBanner(),
        colorOutput: true,
    }
}

// PromptInput 输入提示
func (ui *UIImpl) PromptInput(prompt, defaultValue string) (string, error) {
    var promptText string
    if defaultValue != "" {
        promptText = fmt.Sprintf("%s [%s]", prompt, defaultValue)
    } else {
        promptText = prompt
    }
    
    result, err := pterm.DefaultInteractiveTextInput.Show(promptText)
    if err != nil {
        return "", fmt.Errorf("输入提示失败: %w", err)
    }
    
    if result == "" && defaultValue != "" {
        return defaultValue, nil
    }
    
    return result, nil
}

// PromptSelect 选择提示
func (ui *UIImpl) PromptSelect(prompt string, options []string) (string, error) {
    return ui.Select(prompt, options)
}

// ShowInfo 显示信息
func (ui *UIImpl) ShowInfo(message string) {
    if ui.colorOutput {
        pterm.Info.Println(message)
    } else {
        fmt.Printf("INFO: %s\n", message)
    }
}

// ShowSuccess 显示成功信息
func (ui *UIImpl) ShowSuccess(message string) {
    if ui.colorOutput {
        pterm.Success.Println(message)
    } else {
        fmt.Printf("SUCCESS: %s\n", message)
    }
}

// ShowWarning 显示警告信息
func (ui *UIImpl) ShowWarning(message string) {
    if ui.colorOutput {
        pterm.Warning.Println(message)
    } else {
        fmt.Printf("WARNING: %s\n", message)
    }
}

// ShowError 显示错误信息
func (ui *UIImpl) ShowError(message string) {
    if ui.colorOutput {
        pterm.Error.Println(message)
    } else {
        fmt.Printf("ERROR: %s\n", message)
    }
}

// NewTable 创建表格
func (ui *UIImpl) NewTable() Table {
    return NewTable()
}

// SetColorOutput 设置颜色输出
func (ui *UIImpl) SetColorOutput(enabled bool) {
    ui.colorOutput = enabled
    if banner, ok := ui.Banner.(*BannerImpl); ok {
        banner.SetColorOutput(enabled)
    }
}
```

---

## 6. GitHub集成详细设计

### 6.1 下载器设计

#### 6.1.1 下载器接口定义
```go
// internal/core/github/interfaces.go
package github

import (
    "context"
    "io"
    "time"
)

// Downloader 下载器接口
type Downloader interface {
    // Download 下载文件
    Download(url, dest string) error
    
    // DownloadWithContext 带上下文的下载
    DownloadWithContext(ctx context.Context, url, dest string) error
    
    // DownloadStream 流式下载
    DownloadStream(ctx context.Context, url string, writer io.Writer, progressCallback ProgressCallback) error
    
    // SetTimeout 设置超时时间
    SetTimeout(timeout time.Duration)
    
    // SetRetryCount 设置重试次数
    SetRetryCount(count int)
    
    // SetUserAgent 设置用户代理
    SetUserAgent(userAgent string)
    
    // SetHeaders 设置请求头
    SetHeaders(headers map[string]string)
}

// ProgressCallback 进度回调函数
type ProgressCallback func(downloaded, total int64, percentage float64)

// DownloadOptions 下载选项
type DownloadOptions struct {
    Timeout     time.Duration
    RetryCount  int
    UserAgent   string
    Headers     map[string]string
    SkipTLS     bool
    BufferSize  int
}
```

#### 6.1.2 下载器实现
```go
// internal/core/github/downloader.go
package github

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
    
    "github.com/go-resty/resty/v2"
)

// DownloaderImpl 下载器实现
type DownloaderImpl struct {
    client  *resty.Client
    options *DownloadOptions
}

// NewDownloader 创建下载器
func NewDownloader(options *DownloadOptions) Downloader {
    if options == nil {
        options = &DownloadOptions{
            Timeout:    30 * time.Second,
            RetryCount: 3,
            UserAgent:  "Specify-CLI/1.0.0",
            BufferSize: 32 * 1024, // 32KB
        }
    }
    
    client := resty.New()
    client.SetTimeout(options.Timeout)
    client.SetRetryCount(options.RetryCount)
    client.SetUserAgent(options.UserAgent)
    
    if options.Headers != nil {
        client.SetHeaders(options.Headers)
    }
    
    if options.SkipTLS {
        client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
    }
    
    return &DownloaderImpl{
        client:  client,
        options: options,
    }
}

// Download 下载文件
func (d *DownloaderImpl) Download(url, dest string) error {
    return d.DownloadWithContext(context.Background(), url, dest)
}

// DownloadWithContext 带上下文的下载
func (d *DownloaderImpl) DownloadWithContext(ctx context.Context, url, dest string) error {
    // 确保目标目录存在
    if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
        return fmt.Errorf("创建目标目录失败: %w", err)
    }
    
    // 创建目标文件
    file, err := os.Create(dest)
    if err != nil {
        return fmt.Errorf("创建目标文件失败: %w", err)
    }
    defer file.Close()
    
    // 流式下载
    return d.DownloadStream(ctx, url, file, nil)
}

// DownloadStream 流式下载
func (d *DownloaderImpl) DownloadStream(ctx context.Context, url string, writer io.Writer, progressCallback ProgressCallback) error {
    req := d.client.R().SetContext(ctx)
    
    // 发送HEAD请求获取文件大小
    headResp, err := req.Head(url)
    if err != nil {
        return fmt.Errorf("获取文件信息失败: %w", err)
    }
    
    contentLength := headResp.Header().Get("Content-Length")
    var totalSize int64
    if contentLength != "" {
        if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
            totalSize = size
        }
    }
    
    // 发送GET请求下载文件
    resp, err := req.SetDoNotParseResponse(true).Get(url)
    if err != nil {
        return fmt.Errorf("下载请求失败: %w", err)
    }
    defer resp.RawBody().Close()
    
    if resp.StatusCode() != http.StatusOK {
        return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode())
    }
    
    // 创建进度跟踪器
    var downloaded int64
    buffer := make([]byte, d.options.BufferSize)
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            n, err := resp.RawBody().Read(buffer)
            if n > 0 {
                if _, writeErr := writer.Write(buffer[:n]); writeErr != nil {
                    return fmt.Errorf("写入文件失败: %w", writeErr)
                }
                
                downloaded += int64(n)
                
                // 调用进度回调
                if progressCallback != nil && totalSize > 0 {
                    percentage := float64(downloaded) / float64(totalSize) * 100
                    progressCallback(downloaded, totalSize, percentage)
                }
            }
            
            if err == io.EOF {
                return nil
            }
            if err != nil {
                return fmt.Errorf("读取响应失败: %w", err)
            }
        }
    }
}

// SetTimeout 设置超时时间
func (d *DownloaderImpl) SetTimeout(timeout time.Duration) {
    d.options.Timeout = timeout
    d.client.SetTimeout(timeout)
}

// SetRetryCount 设置重试次数
func (d *DownloaderImpl) SetRetryCount(count int) {
    d.options.RetryCount = count
    d.client.SetRetryCount(count)
}

// SetUserAgent 设置用户代理
func (d *DownloaderImpl) SetUserAgent(userAgent string) {
    d.options.UserAgent = userAgent
    d.client.SetUserAgent(userAgent)
}

// SetHeaders 设置请求头
func (d *DownloaderImpl) SetHeaders(headers map[string]string) {
    d.options.Headers = headers
    d.client.SetHeaders(headers)
}
```

### 6.2 归档解压器设计

#### 6.2.1 归档解压器接口
```go
// internal/core/github/interfaces.go
package github

import (
    "context"
    "io"
)

// ArchiveExtractor 归档解压器接口
type ArchiveExtractor interface {
    // Extract 解压归档文件
    Extract(archivePath, destDir string) error
    
    // ExtractWithContext 带上下文的解压
    ExtractWithContext(ctx context.Context, archivePath, destDir string) error
    
    // ExtractWithFlattening 解压并扁平化目录结构
    ExtractWithFlattening(archivePath, destDir string, skipLevels int) error
    
    // MergeToExistingDir 合并到现有目录
    MergeToExistingDir(archivePath, destDir string, overwrite bool) error
    
    // ExtractFromReader 从Reader解压
    ExtractFromReader(reader io.Reader, destDir string, archiveType ArchiveType) error
    
    // ListContents 列出归档内容
    ListContents(archivePath string) ([]ArchiveEntry, error)
    
    // ValidateArchive 验证归档文件
    ValidateArchive(archivePath string) error
}

// ArchiveType 归档类型
type ArchiveType int

const (
    ArchiveTypeZip ArchiveType = iota
    ArchiveTypeTarGz
    ArchiveTypeTar
    ArchiveTypeRar
)

// ArchiveEntry 归档条目
type ArchiveEntry struct {
    Name     string
    Size     int64
    IsDir    bool
    ModTime  time.Time
    Mode     os.FileMode
}

// ExtractOptions 解压选项
type ExtractOptions struct {
    SkipLevels    int
    Overwrite     bool
    PreservePerms bool
    FilterFunc    func(entry ArchiveEntry) bool
    ProgressCallback func(extracted, total int, currentFile string)
}
```

#### 6.2.2 归档解压器实现
```go
// internal/core/github/extractor.go
package github

import (
    "archive/tar"
    "archive/zip"
    "compress/gzip"
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/mholt/archiver/v4"
)

// ArchiveExtractorImpl 归档解压器实现
type ArchiveExtractorImpl struct {
    options *ExtractOptions
}

// NewArchiveExtractor 创建归档解压器
func NewArchiveExtractor(options *ExtractOptions) ArchiveExtractor {
    if options == nil {
        options = &ExtractOptions{
            SkipLevels:    0,
            Overwrite:     false,
            PreservePerms: true,
        }
    }
    
    return &ArchiveExtractorImpl{
        options: options,
    }
}

// Extract 解压归档文件
func (ae *ArchiveExtractorImpl) Extract(archivePath, destDir string) error {
    return ae.ExtractWithContext(context.Background(), archivePath, destDir)
}

// ExtractWithContext 带上下文的解压
func (ae *ArchiveExtractorImpl) ExtractWithContext(ctx context.Context, archivePath, destDir string) error {
    // 确保目标目录存在
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return fmt.Errorf("创建目标目录失败: %w", err)
    }
    
    // 检测归档类型
    archiveType, err := ae.detectArchiveType(archivePath)
    if err != nil {
        return fmt.Errorf("检测归档类型失败: %w", err)
    }
    
    // 打开归档文件
    file, err := os.Open(archivePath)
    if err != nil {
        return fmt.Errorf("打开归档文件失败: %w", err)
    }
    defer file.Close()
    
    return ae.ExtractFromReader(file, destDir, archiveType)
}

// ExtractWithFlattening 解压并扁平化目录结构
func (ae *ArchiveExtractorImpl) ExtractWithFlattening(archivePath, destDir string, skipLevels int) error {
    tempOptions := *ae.options
    tempOptions.SkipLevels = skipLevels
    
    tempExtractor := &ArchiveExtractorImpl{options: &tempOptions}
    return tempExtractor.Extract(archivePath, destDir)
}

// MergeToExistingDir 合并到现有目录
func (ae *ArchiveExtractorImpl) MergeToExistingDir(archivePath, destDir string, overwrite bool) error {
    tempOptions := *ae.options
    tempOptions.Overwrite = overwrite
    
    tempExtractor := &ArchiveExtractorImpl{options: &tempOptions}
    return tempExtractor.Extract(archivePath, destDir)
}

// ExtractFromReader 从Reader解压
func (ae *ArchiveExtractorImpl) ExtractFromReader(reader io.Reader, destDir string, archiveType ArchiveType) error {
    switch archiveType {
    case ArchiveTypeZip:
        return ae.extractZipFromReader(reader, destDir)
    case ArchiveTypeTarGz:
        return ae.extractTarGzFromReader(reader, destDir)
    case ArchiveTypeTar:
        return ae.extractTarFromReader(reader, destDir)
    default:
        return fmt.Errorf("不支持的归档类型: %v", archiveType)
    }
}

// extractZipFromReader 从Reader解压ZIP文件
func (ae *ArchiveExtractorImpl) extractZipFromReader(reader io.Reader, destDir string) error {
    // 由于zip.Reader需要ReaderAt，我们需要先将内容读取到内存或临时文件
    tempFile, err := os.CreateTemp("", "extract_*.zip")
    if err != nil {
        return fmt.Errorf("创建临时文件失败: %w", err)
    }
    defer os.Remove(tempFile.Name())
    defer tempFile.Close()
    
    // 复制内容到临时文件
    if _, err := io.Copy(tempFile, reader); err != nil {
        return fmt.Errorf("复制到临时文件失败: %w", err)
    }
    
    // 获取文件大小
    stat, err := tempFile.Stat()
    if err != nil {
        return fmt.Errorf("获取文件信息失败: %w", err)
    }
    
    // 打开ZIP读取器
    zipReader, err := zip.NewReader(tempFile, stat.Size())
    if err != nil {
        return fmt.Errorf("创建ZIP读取器失败: %w", err)
    }
    
    // 解压文件
    for i, file := range zipReader.File {
        if ae.options.ProgressCallback != nil {
            ae.options.ProgressCallback(i, len(zipReader.File), file.Name)
        }
        
        if err := ae.extractZipFile(file, destDir); err != nil {
            return fmt.Errorf("解压文件 %s 失败: %w", file.Name, err)
        }
    }
    
    return nil
}

// extractZipFile 解压单个ZIP文件条目
func (ae *ArchiveExtractorImpl) extractZipFile(file *zip.File, destDir string) error {
    // 应用跳过级别
    relativePath := ae.applySkipLevels(file.Name)
    if relativePath == "" {
        return nil // 跳过此文件
    }
    
    // 构建目标路径
    destPath := filepath.Join(destDir, relativePath)
    
    // 安全检查：防止路径遍历攻击
    if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
        return fmt.Errorf("不安全的路径: %s", file.Name)
    }
    
    // 应用过滤器
    if ae.options.FilterFunc != nil {
        entry := ArchiveEntry{
            Name:    file.Name,
            Size:    int64(file.UncompressedSize64),
            IsDir:   file.FileInfo().IsDir(),
            ModTime: file.FileInfo().ModTime(),
            Mode:    file.FileInfo().Mode(),
        }
        if !ae.options.FilterFunc(entry) {
            return nil // 跳过此文件
        }
    }
    
    // 检查是否覆盖现有文件
    if !ae.options.Overwrite {
        if _, err := os.Stat(destPath); err == nil {
            return nil // 文件已存在，跳过
        }
    }
    
    // 创建目录
    if file.FileInfo().IsDir() {
        return os.MkdirAll(destPath, file.FileInfo().Mode())
    }
    
    // 确保父目录存在
    if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
        return fmt.Errorf("创建父目录失败: %w", err)
    }
    
    // 打开ZIP文件条目
    rc, err := file.Open()
    if err != nil {
        return fmt.Errorf("打开ZIP文件条目失败: %w", err)
    }
    defer rc.Close()
    
    // 创建目标文件
    destFile, err := os.Create(destPath)
    if err != nil {
        return fmt.Errorf("创建目标文件失败: %w", err)
    }
    defer destFile.Close()
    
    // 复制内容
    if _, err := io.Copy(destFile, rc); err != nil {
        return fmt.Errorf("复制文件内容失败: %w", err)
    }
    
    // 设置文件权限
    if ae.options.PreservePerms {
        if err := os.Chmod(destPath, file.FileInfo().Mode()); err != nil {
            return fmt.Errorf("设置文件权限失败: %w", err)
        }
    }
    
    return nil
}

// extractTarGzFromReader 从Reader解压TAR.GZ文件
func (ae *ArchiveExtractorImpl) extractTarGzFromReader(reader io.Reader, destDir string) error {
    gzReader, err := gzip.NewReader(reader)
    if err != nil {
        return fmt.Errorf("创建GZIP读取器失败: %w", err)
    }
    defer gzReader.Close()
    
    return ae.extractTarFromReader(gzReader, destDir)
}

// extractTarFromReader 从Reader解压TAR文件
func (ae *ArchiveExtractorImpl) extractTarFromReader(reader io.Reader, destDir string) error {
    tarReader := tar.NewReader(reader)
    
    fileCount := 0
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("读取TAR头失败: %w", err)
        }
        
        if ae.options.ProgressCallback != nil {
            ae.options.ProgressCallback(fileCount, -1, header.Name)
        }
        
        if err := ae.extractTarFile(tarReader, header, destDir); err != nil {
            return fmt.Errorf("解压文件 %s 失败: %w", header.Name, err)
        }
        
        fileCount++
    }
    
    return nil
}

// extractTarFile 解压单个TAR文件条目
func (ae *ArchiveExtractorImpl) extractTarFile(tarReader *tar.Reader, header *tar.Header, destDir string) error {
    // 应用跳过级别
    relativePath := ae.applySkipLevels(header.Name)
    if relativePath == "" {
        return nil // 跳过此文件
    }
    
    // 构建目标路径
    destPath := filepath.Join(destDir, relativePath)
    
    // 安全检查：防止路径遍历攻击
    if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
        return fmt.Errorf("不安全的路径: %s", header.Name)
    }
    
    // 应用过滤器
    if ae.options.FilterFunc != nil {
        entry := ArchiveEntry{
            Name:    header.Name,
            Size:    header.Size,
            IsDir:   header.FileInfo().IsDir(),
            ModTime: header.FileInfo().ModTime(),
            Mode:    header.FileInfo().Mode(),
        }
        if !ae.options.FilterFunc(entry) {
            return nil // 跳过此文件
        }
    }
    
    // 检查是否覆盖现有文件
    if !ae.options.Overwrite {
        if _, err := os.Stat(destPath); err == nil {
            return nil // 文件已存在，跳过
        }
    }
    
    // 根据文件类型处理
    switch header.Typeflag {
    case tar.TypeDir:
        return os.MkdirAll(destPath, header.FileInfo().Mode())
        
    case tar.TypeReg:
        // 确保父目录存在
        if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
            return fmt.Errorf("创建父目录失败: %w", err)
        }
        
        // 创建目标文件
        destFile, err := os.Create(destPath)
        if err != nil {
            return fmt.Errorf("创建目标文件失败: %w", err)
        }
        defer destFile.Close()
        
        // 复制内容
        if _, err := io.Copy(destFile, tarReader); err != nil {
            return fmt.Errorf("复制文件内容失败: %w", err)
        }
        
        // 设置文件权限
        if ae.options.PreservePerms {
            if err := os.Chmod(destPath, header.FileInfo().Mode()); err != nil {
                return fmt.Errorf("设置文件权限失败: %w", err)
            }
        }
        
    case tar.TypeSymlink:
        // 创建符号链接
        if err := os.Symlink(header.Linkname, destPath); err != nil {
            return fmt.Errorf("创建符号链接失败: %w", err)
        }
        
    default:
        // 忽略其他类型的文件
        return nil
    }
    
    return nil
}

// applySkipLevels 应用跳过级别
func (ae *ArchiveExtractorImpl) applySkipLevels(path string) string {
    if ae.options.SkipLevels <= 0 {
        return path
    }
    
    parts := strings.Split(path, "/")
    if len(parts) <= ae.options.SkipLevels {
        return "" // 跳过此文件
    }
    
    return strings.Join(parts[ae.options.SkipLevels:], "/")
}

// detectArchiveType 检测归档类型
func (ae *ArchiveExtractorImpl) detectArchiveType(archivePath string) (ArchiveType, error) {
    ext := strings.ToLower(filepath.Ext(archivePath))
    
    switch ext {
    case ".zip":
        return ArchiveTypeZip, nil
    case ".gz":
        if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
            return ArchiveTypeTarGz, nil
        }
        return ArchiveTypeTarGz, nil
    case ".tar":
        return ArchiveTypeTar, nil
    case ".rar":
        return ArchiveTypeRar, nil
    default:
        return ArchiveTypeZip, fmt.Errorf("不支持的归档类型: %s", ext)
    }
}

// ListContents 列出归档内容
func (ae *ArchiveExtractorImpl) ListContents(archivePath string) ([]ArchiveEntry, error) {
    archiveType, err := ae.detectArchiveType(archivePath)
    if err != nil {
        return nil, err
    }
    
    file, err := os.Open(archivePath)
    if err != nil {
        return nil, fmt.Errorf("打开归档文件失败: %w", err)
    }
    defer file.Close()
    
    switch archiveType {
    case ArchiveTypeZip:
        return ae.listZipContents(file)
    case ArchiveTypeTarGz:
        return ae.listTarGzContents(file)
    case ArchiveTypeTar:
        return ae.listTarContents(file)
    default:
        return nil, fmt.Errorf("不支持的归档类型: %v", archiveType)
    }
}

// listZipContents 列出ZIP文件内容
func (ae *ArchiveExtractorImpl) listZipContents(file *os.File) ([]ArchiveEntry, error) {
    stat, err := file.Stat()
    if err != nil {
        return nil, err
    }
    
    zipReader, err := zip.NewReader(file, stat.Size())
    if err != nil {
        return nil, err
    }
    
    var entries []ArchiveEntry
    for _, f := range zipReader.File {
        entries = append(entries, ArchiveEntry{
            Name:    f.Name,
            Size:    int64(f.UncompressedSize64),
            IsDir:   f.FileInfo().IsDir(),
            ModTime: f.FileInfo().ModTime(),
            Mode:    f.FileInfo().Mode(),
        })
    }
    
    return entries, nil
}

// listTarGzContents 列出TAR.GZ文件内容
func (ae *ArchiveExtractorImpl) listTarGzContents(file *os.File) ([]ArchiveEntry, error) {
    gzReader, err := gzip.NewReader(file)
    if err != nil {
        return nil, err
    }
    defer gzReader.Close()
    
    return ae.listTarContents(gzReader)
}

// listTarContents 列出TAR文件内容
func (ae *ArchiveExtractorImpl) listTarContents(reader io.Reader) ([]ArchiveEntry, error) {
    tarReader := tar.NewReader(reader)
    
    var entries []ArchiveEntry
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        
        entries = append(entries, ArchiveEntry{
            Name:    header.Name,
            Size:    header.Size,
            IsDir:   header.FileInfo().IsDir(),
            ModTime: header.FileInfo().ModTime(),
            Mode:    header.FileInfo().Mode(),
        })
    }
    
    return entries, nil
}

// ValidateArchive 验证归档文件
func (ae *ArchiveExtractorImpl) ValidateArchive(archivePath string) error {
    _, err := ae.ListContents(archivePath)
    return err
}
```

### 6.3 GitHub客户端设计

#### 6.3.1 GitHub客户端接口
```go
// internal/core/github/interfaces.go
package github

import (
    "context"
    "time"
)

// GitHubClient GitHub客户端接口
type GitHubClient interface {
    // GetRepository 获取仓库信息
    GetRepository(ctx context.Context, owner, repo string) (*Repository, error)
    
    // GetLatestRelease 获取最新发布版本
    GetLatestRelease(ctx context.Context, owner, repo string) (*Release, error)
    
    // GetReleases 获取发布版本列表
    GetReleases(ctx context.Context, owner, repo string, page, perPage int) ([]*Release, error)
    
    // DownloadArchive 下载仓库归档
    DownloadArchive(ctx context.Context, owner, repo, ref, format, dest string) error
    
    // GetContents 获取文件内容
    GetContents(ctx context.Context, owner, repo, path, ref string) (*Content, error)
    
    // SearchRepositories 搜索仓库
    SearchRepositories(ctx context.Context, query string, options *SearchOptions) (*SearchResult, error)
    
    // GetRateLimit 获取API限制信息
    GetRateLimit(ctx context.Context) (*RateLimit, error)
    
    // SetToken 设置访问令牌
    SetToken(token string)
}

// Repository 仓库信息
type Repository struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    FullName    string    `json:"full_name"`
    Owner       *User     `json:"owner"`
    Description string    `json:"description"`
    Private     bool      `json:"private"`
    Fork        bool      `json:"fork"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    PushedAt    time.Time `json:"pushed_at"`
    Size        int       `json:"size"`
    Language    string    `json:"language"`
    ForksCount  int       `json:"forks_count"`
    StarsCount  int       `json:"stargazers_count"`
    DefaultBranch string  `json:"default_branch"`
    CloneURL    string    `json:"clone_url"`
    SSHURL      string    `json:"ssh_url"`
    HTMLURL     string    `json:"html_url"`
}

// User 用户信息
type User struct {
    ID        int64  `json:"id"`
    Login     string `json:"login"`
    AvatarURL string `json:"avatar_url"`
    HTMLURL   string `json:"html_url"`
    Type      string `json:"type"`
}

// Release 发布版本信息
type Release struct {
    ID          int64     `json:"id"`
    TagName     string    `json:"tag_name"`
    Name        string    `json:"name"`
    Body        string    `json:"body"`
    Draft       bool      `json:"draft"`
    Prerelease  bool      `json:"prerelease"`
    CreatedAt   time.Time `json:"created_at"`
    PublishedAt time.Time `json:"published_at"`
    Author      *User     `json:"author"`
    Assets      []*Asset  `json:"assets"`
    TarballURL  string    `json:"tarball_url"`
    ZipballURL  string    `json:"zipball_url"`
}

// Asset 发布资产
type Asset struct {
    ID                 int64     `json:"id"`
    Name               string    `json:"name"`
    Label              string    `json:"label"`
    ContentType        string    `json:"content_type"`
    Size               int       `json:"size"`
    DownloadCount      int       `json:"download_count"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
    BrowserDownloadURL string    `json:"browser_download_url"`
}

// Content 文件内容
type Content struct {
    Name        string `json:"name"`
    Path        string `json:"path"`
    SHA         string `json:"sha"`
    Size        int    `json:"size"`
    Type        string `json:"type"`
    Content     string `json:"content"`
    Encoding    string `json:"encoding"`
    DownloadURL string `json:"download_url"`
    HTMLURL     string `json:"html_url"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
    Sort      string // stars, forks, updated
    Order     string // asc, desc
    Page      int
    PerPage   int
    Language  string
    Topic     string
    User      string
    Org       string
    Created   string
    Updated   string
    Pushed    string
    Size      string
    Fork      *bool
    Archived  *bool
    Mirror    *bool
}

// SearchResult 搜索结果
type SearchResult struct {
    TotalCount        int           `json:"total_count"`
    IncompleteResults bool          `json:"incomplete_results"`
    Items             []*Repository `json:"items"`
}

// RateLimit API限制信息
type RateLimit struct {
    Limit     int       `json:"limit"`
    Remaining int       `json:"remaining"`
    Reset     time.Time `json:"reset"`
    Used      int       `json:"used"`
}
```

#### 6.3.2 GitHub客户端实现
```go
// internal/core/github/client.go
package github

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"
    
    "github.com/go-resty/resty/v2"
)

// GitHubClientImpl GitHub客户端实现
type GitHubClientImpl struct {
    client  *resty.Client
    baseURL string
    token   string
}

// NewGitHubClient 创建GitHub客户端
func NewGitHubClient(token string) GitHubClient {
    client := resty.New()
    client.SetBaseURL("https://api.github.com")
    client.SetTimeout(30 * time.Second)
    client.SetRetryCount(3)
    client.SetUserAgent("Specify-CLI/1.0.0")
    
    // 设置通用头部
    client.SetHeaders(map[string]string{
        "Accept":               "application/vnd.github.v3+json",
        "X-GitHub-Api-Version": "2022-11-28",
    })
    
    githubClient := &GitHubClientImpl{
        client:  client,
        baseURL: "https://api.github.com",
    }
    
    if token != "" {
        githubClient.SetToken(token)
    }
    
    return githubClient
}

// SetToken 设置访问令牌
func (gc *GitHubClientImpl) SetToken(token string) {
    gc.token = token
    if token != "" {
        gc.client.SetAuthToken(token)
    }
}

// GetRepository 获取仓库信息
func (gc *GitHubClientImpl) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
    var repository Repository
    
    resp, err := gc.client.R().
        SetContext(ctx).
        SetResult(&repository).
        Get(fmt.Sprintf("/repos/%s/%s", owner, repo))
    
    if err != nil {
        return nil, fmt.Errorf("获取仓库信息失败: %w", err)
    }
    
    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("获取仓库信息失败，状态码: %d", resp.StatusCode())
    }
    
    return &repository, nil
}

// GetLatestRelease 获取最新发布版本
func (gc *GitHubClientImpl) GetLatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
    var release Release
    
    resp, err := gc.client.R().
        SetContext(ctx).
        SetResult(&release).
        Get(fmt.Sprintf("/repos/%s/%s/releases/latest", owner, repo))
    
    if err != nil {
        return nil, fmt.Errorf("获取最新发布版本失败: %w", err)
    }
    
    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("获取最新发布版本失败，状态码: %d", resp.StatusCode())
    }
    
    return &release, nil
}

// GetReleases 获取发布版本列表
func (gc *GitHubClientImpl) GetReleases(ctx context.Context, owner, repo string, page, perPage int) ([]*Release, error) {
    var releases []*Release
    
    if page <= 0 {
        page = 1
    }
    if perPage <= 0 || perPage > 100 {
        perPage = 30
    }
    
    resp, err := gc.client.R().
        SetContext(ctx).
        SetQueryParams(map[string]string{
            "page":     strconv.Itoa(page),
            "per_page": strconv.Itoa(perPage),
        }).
        SetResult(&releases).
        Get(fmt.Sprintf("/repos/%s/%s/releases", owner, repo))
    
    if err != nil {
        return nil, fmt.Errorf("获取发布版本列表失败: %w", err)
    }
    
    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("获取发布版本列表失败，状态码: %d", resp.StatusCode())
    }
    
    return releases, nil
}

// DownloadArchive 下载仓库归档
func (gc *GitHubClientImpl) DownloadArchive(ctx context.Context, owner, repo, ref, format, dest string) error {
    if format != "zipball" && format != "tarball" {
        format = "zipball"
    }
    
    if ref == "" {
        ref = "main"
    }
    
    // 创建下载器
    downloader := NewDownloader(&DownloadOptions{
        Timeout:   60 * time.Second,
        UserAgent: "Specify-CLI/1.0.0",
        Headers: map[string]string{
            "Accept":               "application/vnd.github.v3+json",
            "X-GitHub-Api-Version": "2022-11-28",
        },
    })
    
    // 如果有token，添加认证头
    if gc.token != "" {
        downloader.SetHeaders(map[string]string{
            "Authorization": "token " + gc.token,
        })
    }
    
    // 构建下载URL
    downloadURL := fmt.Sprintf("%s/repos/%s/%s/%s/%s", gc.baseURL, owner, repo, format, ref)
    
    return downloader.DownloadWithContext(ctx, downloadURL, dest)
}

// GetContents 获取文件内容
func (gc *GitHubClientImpl) GetContents(ctx context.Context, owner, repo, path, ref string) (*Content, error) {
    var content Content
    
    queryParams := make(map[string]string)
    if ref != "" {
        queryParams["ref"] = ref
    }
    
    resp, err := gc.client.R().
        SetContext(ctx).
        SetQueryParams(queryParams).
        SetResult(&content).
        Get(fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path))
    
    if err != nil {
        return nil, fmt.Errorf("获取文件内容失败: %w", err)
    }
    
    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("获取文件内容失败，状态码: %d", resp.StatusCode())
    }
    
    // 解码Base64内容
    if content.Encoding == "base64" && content.Content != "" {
        decoded, err := base64.StdEncoding.DecodeString(content.Content)
        if err != nil {
            return nil, fmt.Errorf("解码文件内容失败: %w", err)
        }
        content.Content = string(decoded)
    }
    
    return &content, nil
}

// SearchRepositories 搜索仓库
func (gc *GitHubClientImpl) SearchRepositories(ctx context.Context, query string, options *SearchOptions) (*SearchResult, error) {
    var result SearchResult
    
    if options == nil {
        options = &SearchOptions{
            Sort:    "stars",
            Order:   "desc",
            Page:    1,
            PerPage: 30,
        }
    }
    
    // 构建查询参数
    queryParams := map[string]string{
        "q": query,
    }
    
    if options.Sort != "" {
        queryParams["sort"] = options.Sort
    }
    if options.Order != "" {
        queryParams["order"] = options.Order
    }
    if options.Page > 0 {
        queryParams["page"] = strconv.Itoa(options.Page)
    }
    if options.PerPage > 0 {
        queryParams["per_page"] = strconv.Itoa(options.PerPage)
    }
    
    // 添加搜索限定符
    searchQuery := query
    if options.Language != "" {
        searchQuery += " language:" + options.Language
    }
    if options.Topic != "" {
        searchQuery += " topic:" + options.Topic
    }
    if options.User != "" {
        searchQuery += " user:" + options.User
    }
    if options.Org != "" {
        searchQuery += " org:" + options.Org
    }
    if options.Fork != nil {
        searchQuery += " fork:" + strconv.FormatBool(*options.Fork)
    }
    if options.Archived != nil {
        searchQuery += " archived:" + strconv.FormatBool(*options.Archived)
    }
    if options.Mirror != nil {
        searchQuery += " mirror:" + strconv.FormatBool(*options.Mirror)
    }
    
    queryParams["q"] = searchQuery
    
    resp, err := gc.client.R().
        SetContext(ctx).
        SetQueryParams(queryParams).
        SetResult(&result).
        Get("/search/repositories")
    
    if err != nil {
        return nil, fmt.Errorf("搜索仓库失败: %w", err)
    }
    
    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("搜索仓库失败，状态码: %d", resp.StatusCode())
    }
    
    return &result, nil
}

// GetRateLimit 获取API限制信息
func (gc *GitHubClientImpl) GetRateLimit(ctx context.Context) (*RateLimit, error) {
    var rateLimitResponse struct {
        Rate *RateLimit `json:"rate"`
    }
    
    resp, err := gc.client.R().
        SetContext(ctx).
        SetResult(&rateLimitResponse).
        Get("/rate_limit")
    
    if err != nil {
        return nil, fmt.Errorf("获取API限制信息失败: %w", err)
    }
    
    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("获取API限制信息失败，状态码: %d", resp.StatusCode())
    }
    
    return rateLimitResponse.Rate, nil
}
```

### 6.4 GitHub集成模块

#### 6.4.1 模块定义
```go
// internal/core/github/module.go
package github

import (
    "go.uber.org/fx"
)

// GitHubModule GitHub集成模块
var GitHubModule = fx.Module("github",
    // 提供下载器
    fx.Provide(func() Downloader {
        return NewDownloader(&DownloadOptions{
            Timeout:    30 * time.Second,
            RetryCount: 3,
            UserAgent:  "Specify-CLI/1.0.0",
            BufferSize: 32 * 1024,
        })
    }),
    
    // 提供归档解压器
    fx.Provide(func() ArchiveExtractor {
        return NewArchiveExtractor(&ExtractOptions{
            SkipLevels:    0,
            Overwrite:     false,
            PreservePerms: true,
        })
    }),
    
    // 提供GitHub客户端
    fx.Provide(func() GitHubClient {
        return NewGitHubClient("")
    }),
    
    // 生命周期钩子
    fx.Invoke(func(lc fx.Lifecycle, client GitHubClient) {
        lc.Append(fx.Hook{
            OnStart: func(ctx context.Context) error {
                // 初始化GitHub客户端
                return nil
            },
            OnStop: func(ctx context.Context) error {
                // 清理资源
                return nil
            },
        })
    }),
)
```