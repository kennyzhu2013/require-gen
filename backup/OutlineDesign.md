# Specify CLI Go版本概要设计文档

## 1. 项目概述

### 1.1 项目背景
基于现有Python版本的Specify CLI工具，重新设计并实现Go版本，提供AI辅助开发项目的初始化和管理功能。

### 1.2 核心功能需求
- **项目初始化**: 支持多种AI助手的项目模板初始化
- **工具检查**: 验证开发环境中必要工具的安装状态
- **交互式UI**: 提供友好的命令行交互界面
- **GitHub集成**: 从GitHub下载和管理项目模板
- **跨平台支持**: 支持Windows、macOS、Linux操作系统

### 1.3 设计原则
- **模块化设计**: 清晰的职责分离和接口定义
- **接口驱动**: 基于接口的可测试和可扩展架构
- **并发安全**: 充分利用Go的并发特性
- **错误处理**: 明确的错误类型和处理机制
- **可测试性**: 支持单元测试和集成测试

## 2. 系统架构

### 2.1 整体架构
```
┌─────────────────┐
│   CLI Layer     │  ← Cobra命令行框架
├─────────────────┤
│  Service Layer  │  ← 业务逻辑层
├─────────────────┤
│Repository Layer │  ← 数据访问层
├─────────────────┤
│Infrastructure   │  ← 基础设施层
└─────────────────┘
```

### 2.2 依赖注入架构
使用依赖注入容器管理组件生命周期，确保模块间的松耦合。

## 3. 项目结构

```
specify-cli-go/
├── cmd/                          # CLI命令入口
│   ├── root.go                   # 根命令定义
│   ├── init.go                   # init命令实现
│   └── check.go                  # check命令实现
├── internal/
│   ├── cli/                      # CLI相关组件
│   │   ├── commands/             # 命令实现
│   │   │   ├── init_command.go
│   │   │   └── check_command.go
│   │   └── ui/                   # UI组件
│   │       ├── step_tracker.go   # 进度跟踪器
│   │       ├── selector.go       # 交互选择器
│   │       ├── banner.go         # 横幅显示
│   │       └── refresh.go        # 实时刷新机制
│   ├── core/                     # 核心业务逻辑
│   │   ├── services/             # 服务层
│   │   │   ├── init_service.go
│   │   │   ├── check_service.go
│   │   │   └── template_service.go
│   │   ├── downloaders/          # 下载器
│   │   │   ├── github_downloader.go
│   │   │   └── archive_extractor.go
│   │   ├── checkers/             # 检查器
│   │   │   ├── tool_checker.go
│   │   │   ├── git_checker.go
│   │   │   └── special_tool_checker.go
│   │   └── security/             # 安全组件
│   │       └── security_notifier.go
│   ├── infrastructure/           # 基础设施
│   │   ├── config/               # 配置管理
│   │   │   ├── agent_config.go
│   │   │   └── app_config.go
│   │   ├── http/                 # HTTP客户端
│   │   │   └── github_client.go
│   │   ├── filesystem/           # 文件系统操作
│   │   │   ├── file_manager.go
│   │   │   └── permission_manager.go
│   │   └── system/               # 系统集成
│   │       ├── command_executor.go
│   │       └── environment_guide.go
│   ├── models/                   # 数据模型
│   │   ├── agent.go
│   │   ├── project.go
│   │   ├── github_release.go
│   │   └── errors.go
│   └── utils/                    # 工具函数
│       ├── path_utils.go
│       └── string_utils.go
├── pkg/                          # 公共包
│   └── version/
│       └── version.go
├── configs/                      # 配置文件
│   └── agents.yaml
├── scripts/                      # 构建脚本
│   ├── build.sh
│   └── build.ps1
├── tests/                        # 测试文件
│   ├── unit/
│   ├── integration/
│   └── fixtures/
├── docs/                         # 文档
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 4. 核心组件设计

### 4.1 CLI命令组件

#### 4.1.1 根命令 (RootCommand)
```go
type RootCommand struct {
    app     Application
    ui      UI
    config  Config
}

type RootCommandConfig struct {
    Version     string
    Description string
    BannerText  string
}
```

#### 4.1.2 Init命令 (InitCommand)
```go
type InitCommand struct {
    service      InitService
    ui          UI
    validator   InputValidator
}

type InitCommandArgs struct {
    ProjectName        string
    AI                string
    ScriptType        string
    IgnoreAgentTools  bool
    NoGit            bool
    Here             bool
    Force            bool
    SkipTLS          bool
    Debug            bool    // 新增：调试模式
    Verbose          bool    // 新增：详细输出模式
    GitHubToken      string
}

// 新增：调试和详细输出配置
type OutputConfig struct {
    Debug   bool
    Verbose bool
    Logger  Logger
}

type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}

type DefaultLogger struct {
    debug   bool
    verbose bool
    writer  io.Writer
}

func NewLogger(debug, verbose bool) *DefaultLogger {
    return &DefaultLogger{
        debug:   debug,
        verbose: verbose,
        writer:  os.Stderr,
    }
}

func (dl *DefaultLogger) Debug(msg string, args ...interface{}) {
    if dl.debug {
        fmt.Fprintf(dl.writer, "[DEBUG] "+msg+"\n", args...)
    }
}

func (dl *DefaultLogger) Info(msg string, args ...interface{}) {
    if dl.verbose || dl.debug {
        fmt.Fprintf(dl.writer, "[INFO] "+msg+"\n", args...)
    }
}

func (dl *DefaultLogger) Warn(msg string, args ...interface{}) {
    fmt.Fprintf(dl.writer, "[WARN] "+msg+"\n", args...)
}

func (dl *DefaultLogger) Error(msg string, args ...interface{}) {
    fmt.Fprintf(dl.writer, "[ERROR] "+msg+"\n", args...)
}
```

#### 4.1.3 Check命令 (CheckCommand)
```go
type CheckCommand struct {
    service CheckService
    ui      UI
}
```

### 4.2 UI组件 (增强版)

#### 4.2.1 步骤跟踪器 (StepTracker)
```go
type StepTracker struct {
    steps           []Step
    currentStep     int
    refreshCallback func()
    refreshMutex    sync.RWMutex
    live           *live.Live
    renderer       StepRenderer
}

type Step struct {
    ID          string
    Name        string
    Status      StepStatus
    SubSteps    []Step
    Progress    float64
    Message     string
    StartTime   time.Time
    EndTime     *time.Time
}

type StepStatus int

const (
    StepPending StepStatus = iota
    StepInProgress
    StepCompleted
    StepFailed
    StepSkipped
)

// 新增：实时刷新机制
func (st *StepTracker) AttachRefresh(callback func()) {
    st.refreshMutex.Lock()
    defer st.refreshMutex.Unlock()
    st.refreshCallback = callback
}

func (st *StepTracker) triggerRefresh() {
    st.refreshMutex.RLock()
    callback := st.refreshCallback
    st.refreshMutex.RUnlock()
    
    if callback != nil {
        callback()
    }
}
```

#### 4.2.2 交互选择器 (Selector) - 增强版
```go
type Selector struct {
    options     []SelectOption
    selected    int
    keyHandler  KeyHandler
    renderer    SelectRenderer
    live       *live.Live
}

type SelectOption struct {
    Value       string
    Label       string
    Description string
    Disabled    bool
}

// 新增：键盘处理器
type KeyHandler interface {
    HandleKey(key rune) (action KeyAction, handled bool)
}

type KeyAction int

const (
    ActionNone KeyAction = iota
    ActionUp
    ActionDown
    ActionSelect
    ActionCancel
    ActionPageUp
    ActionPageDown
)

// 增强的键盘支持
type EnhancedKeyHandler struct{}

func (ekh *EnhancedKeyHandler) HandleKey(key rune) (KeyAction, bool) {
    switch key {
    case 'k', 'K': // Vim-style up
        return ActionUp, true
    case 'j', 'J': // Vim-style down
        return ActionDown, true
    case 16: // Ctrl+P
        return ActionUp, true
    case 14: // Ctrl+N
        return ActionDown, true
    case 27: // ESC
        return ActionCancel, true
    case 13, 10: // Enter
        return ActionSelect, true
    case 'q', 'Q':
        return ActionCancel, true
    default:
        return ActionNone, false
    }
}
```

#### 4.2.3 横幅显示 (Banner)
```go
type Banner struct {
    text     string
    tagline  string
    style    BannerStyle
    renderer BannerRenderer
}

type BannerStyle struct {
    TextColor    string
    TaglineColor string
    BorderColor  string
    Width        int
}
```

### 4.3 GitHub操作组件 (增强版)

#### 4.3.1 HTTP客户端配置 (新增)
```go
type HTTPClient struct {
    client    *http.Client
    skipTLS   bool
    timeout   time.Duration
    userAgent string
}

// SSL/TLS配置实现
func NewHTTPClient(skipTLS bool, timeout time.Duration) *HTTPClient {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: skipTLS,
        },
        // 使用系统证书池
        TLSClientConfig: &tls.Config{
            RootCAs: loadSystemCertPool(),
        },
    }
    
    // 如果需要跳过TLS验证
    if skipTLS {
        transport.TLSClientConfig.InsecureSkipVerify = true
    }
    
    return &HTTPClient{
        client: &http.Client{
            Transport: transport,
            Timeout:   timeout,
        },
        skipTLS:   skipTLS,
        timeout:   timeout,
        userAgent: "specify-cli-go/1.0",
    }
}

func loadSystemCertPool() *x509.CertPool {
    // 加载系统证书池，类似Python的truststore
    certPool, err := x509.SystemCertPool()
    if err != nil {
        // 如果无法加载系统证书池，创建新的
        certPool = x509.NewCertPool()
    }
    return certPool
}

func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
    req.Header.Set("User-Agent", hc.userAgent)
    return hc.client.Do(req)
}
```

#### 4.3.2 GitHub下载器
```go
type GitHubDownloader struct {
    client       *HTTPClient
    auth         AuthProvider
    progress     ProgressTracker
    streamBuffer int64 // 新增：流式下载缓冲区大小
}

// 新增：流式下载接口
type StreamDownloader interface {
    DownloadStream(url string, dest io.Writer, progress ProgressCallback) error
}

type ProgressCallback func(downloaded, total int64)
```

#### 4.3.2 归档解压器 (增强版)
```go
type ArchiveExtractor struct {
    tempDir     string
    flattener   DirectoryFlattener // 新增：目录扁平化器
    merger      FileMerger         // 新增：文件合并器
}

// 新增：目录扁平化接口
type DirectoryFlattener interface {
    FlattenNestedDirs(srcDir string) error
    ShouldFlatten(dirStructure []string) bool
}

// 新增：文件合并接口
type FileMerger interface {
    MergeToExistingDir(srcDir, destDir string, strategy MergeStrategy) error
}

type MergeStrategy int

const (
    MergeOverwrite MergeStrategy = iota
    MergeSkip
    MergePrompt
)

// 增强的解压逻辑
func (ae *ArchiveExtractor) ExtractWithFlattening(src, dest string, flattenNested bool) error {
    // 1. 标准解压
    if err := ae.extractArchive(src, dest); err != nil {
        return err
    }
    
    // 2. 检查是否需要扁平化
    if flattenNested {
        if err := ae.flattener.FlattenNestedDirs(dest); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 4.4 系统集成组件 (增强版)

#### 4.4.1 特殊工具检查器
```go
type SpecialToolChecker struct {
    claudeLocalPath string
    toolPaths      map[string][]string // 工具名 -> 可能的路径列表
}

// 新增：Claude CLI特殊处理
func (stc *SpecialToolChecker) CheckClaude() (bool, string) {
    // 检查标准路径
    if path, found := stc.checkStandardPath("claude"); found {
        return true, path
    }
    
    // 检查特殊路径：migrate-installer
    migrateInstallerPath := filepath.Join(stc.claudeLocalPath, "migrate-installer")
    if stc.pathExists(migrateInstallerPath) {
        return true, migrateInstallerPath
    }
    
    return false, ""
}

// 新增：工具特殊路径配置
var SpecialToolPaths = map[string][]string{
    "claude": {
        "claude",
        filepath.Join(os.Getenv("CLAUDE_LOCAL_PATH"), "migrate-installer"),
    },
    "cursor": {
        "cursor",
        "/Applications/Cursor.app/Contents/Resources/app/bin/cursor", // macOS
        "C:\\Users\\%USERNAME%\\AppData\\Local\\Programs\\cursor\\cursor.exe", // Windows
    },
}
```

#### 4.4.2 权限管理器 (增强版)
```go
type PermissionManager struct {
    osType OSType
}

type OSType int

const (
    OSWindows OSType = iota
    OSLinux
    OSDarwin
)

// 增强的权限设置
func (pm *PermissionManager) SetExecutablePermissions(dir string, recursive bool) error {
    if pm.osType == OSWindows {
        return nil // Windows不需要设置执行权限
    }
    
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // 只处理.sh文件
        if strings.HasSuffix(path, ".sh") {
            return os.Chmod(path, 0755)
        }
        
        return nil
    })
}
```

#### 4.4.3 安全通知器 (新增)
```go
type SecurityNotifier struct {
    ui UI
}

func (sn *SecurityNotifier) ShowAgentFolderWarning(agentFolder string) {
    warning := fmt.Sprintf(`
⚠️  安全提示：
建议将 %s 文件夹添加到 .gitignore 中，避免意外提交敏感信息。

建议执行：
echo "%s" >> .gitignore
`, agentFolder, agentFolder)
    
    sn.ui.ShowWarning(warning)
}

func (sn *SecurityNotifier) ShowCredentialLeakageWarning() {
    warning := `
🔒 安全提醒：
请确保不要在代码中硬编码API密钥或访问令牌。
建议使用环境变量或配置文件管理敏感信息。
`
    sn.ui.ShowWarning(warning)
}
```

#### 4.4.4 环境设置指导 (新增)
```go
type EnvironmentGuide struct {
    osType OSType
}

func (eg *EnvironmentGuide) GenerateSetupInstructions(agent string, projectPath string) []string {
    instructions := []string{}
    
    switch agent {
    case "claude":
        if eg.osType == OSWindows {
            instructions = append(instructions, 
                fmt.Sprintf("set CODEX_HOME=%s", projectPath),
                "# 或者添加到系统环境变量中",
            )
        } else {
            instructions = append(instructions,
                fmt.Sprintf("export CODEX_HOME=%s", projectPath),
                fmt.Sprintf("echo 'export CODEX_HOME=%s' >> ~/.bashrc", projectPath),
            )
        }
    case "copilot":
        instructions = append(instructions,
            "# GitHub Copilot 已配置完成",
            "# 请确保已在 VS Code 中安装 GitHub Copilot 扩展",
        )
    }
    
    return instructions
}

func (eg *EnvironmentGuide) ShowNextSteps(agent string, isCurrentDir bool) {
    steps := []string{}
    
    if !isCurrentDir {
        steps = append(steps, "cd <project-name>")
    }
    
    switch agent {
    case "claude":
        steps = append(steps,
            "# 启动 Claude Code",
            "code .",
            "# 或使用 Claude CLI",
            "claude chat",
        )
    case "copilot":
        steps = append(steps,
            "# 启动 VS Code",
            "code .",
            "# 开始使用 GitHub Copilot",
        )
    }
    
    // 显示步骤...
}
```

### 4.5 配置管理 (完整版)

#### 4.5.1 Agent配置
```go
type AgentConfig struct {
    Name        string `yaml:"name"`
    Folder      string `yaml:"folder"`
    InstallURL  string `yaml:"install_url"`
    RequiresCLI bool   `yaml:"requires_cli"`
    SpecialPath string `yaml:"special_path,omitempty"`
}

// 完整的Agent配置 (与Python版本保持一致的13个)
var AgentConfigs = map[string]AgentConfig{
    "copilot": {
        Name:        "GitHub Copilot",
        Folder:      ".github/",
        InstallURL:  "",  // IDE-based, no CLI check needed
        RequiresCLI: false,
    },
    "claude": {
        Name:        "Claude Code",
        Folder:      ".claude/",
        InstallURL:  "https://docs.anthropic.com/en/docs/claude-code/setup",
        RequiresCLI: true,
        SpecialPath: "migrate-installer",
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
        InstallURL:  "",  // IDE-based
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
        InstallURL:  "",  // IDE-based
        RequiresCLI: false,
    },
    "kilocode": {
        Name:        "Kilo Code",
        Folder:      ".kilocode/",
        InstallURL:  "",  // IDE-based
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
        InstallURL:  "",  // IDE-based
        RequiresCLI: false,
    },
    "q": {
        Name:        "Amazon Q Developer CLI",
        Folder:      ".amazonq/",
        InstallURL:  "https://aws.amazon.com/developer/learning/q-developer-cli/",
        RequiresCLI: true,
    },
}

type ScriptTypeConfig struct {
    Name      string `yaml:"name"`
    Extension string `yaml:"extension"`
    Shebang   string `yaml:"shebang"`
}

var ScriptTypeChoices = map[string]ScriptTypeConfig{
    "bash": {
        Name:      "Bash Scripts",
        Extension: ".sh",
        Shebang:   "#!/bin/bash",
    },
    "powershell": {
        Name:      "PowerShell Scripts",
        Extension: ".ps1",
        Shebang:   "# PowerShell Script",
    },
}
```

## 5. 核心接口定义

### 5.1 应用程序接口
```go
type Application interface {
    Initialize() error
    Run(args []string) error
    Shutdown() error
}

type InitService interface {
    InitializeProject(args InitCommandArgs) error
    ValidateProjectName(name string) error
    SelectAI(available []string) (string, error)
    SelectScriptType(available []string) (string, error)
}

type CheckService interface {
    CheckAllTools() ([]ToolStatus, error)
    CheckTool(name string) (ToolStatus, error)
    GenerateInstallationTips(missing []string) []string
}
```

### 5.2 UI接口 (增强版)
```go
type UI interface {
    ShowBanner(banner Banner) error
    ShowProgress(tracker *StepTracker) error
    ShowSelection(selector *Selector) (string, error)
    ShowMessage(message string) error
    ShowWarning(warning string) error
    ShowError(err error) error
    Confirm(message string) (bool, error)
}

// 新增：实时刷新接口
type RefreshableUI interface {
    UI
    StartLiveMode() error
    StopLiveMode() error
    Refresh() error
}
```

### 5.3 下载器接口 (增强版)
```go
type Downloader interface {
    Download(url, dest string) error
    DownloadWithProgress(url, dest string, progress ProgressCallback) error
    DownloadStream(url string, dest io.Writer, progress ProgressCallback) error // 新增
}

type ArchiveExtractor interface {
    Extract(src, dest string) error
    ExtractWithFlattening(src, dest string, flattenNested bool) error // 新增
    MergeToExistingDir(src, dest string, strategy MergeStrategy) error // 新增
}
```

### 5.4 检查器接口 (增强版)
```go
type Checker interface {
    Check(name string) (bool, error)
    CheckWithPath(name string) (bool, string, error) // 新增：返回路径信息
}

type SpecialChecker interface {
    Checker
    CheckSpecialPaths(name string, paths []string) (bool, string, error) // 新增
}
```

## 6. 错误处理机制

### 6.1 自定义错误类型
```go
type SpecifyError struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}

type ErrorType int

const (
    ErrorTypeValidation ErrorType = iota
    ErrorTypeNetwork
    ErrorTypeFileSystem
    ErrorTypeGitHub
    ErrorTypeToolCheck
    ErrorTypeUserCancelled
    ErrorTypePermission
    ErrorTypeConfiguration
)

func (e *SpecifyError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func (e *SpecifyError) Unwrap() error {
    return e.Cause
}
```

### 6.2 错误处理策略
```go
type ErrorHandler interface {
    Handle(err error) error
    ShouldRetry(err error) bool
    GetUserFriendlyMessage(err error) string
}

type DefaultErrorHandler struct {
    ui UI
}

func (deh *DefaultErrorHandler) Handle(err error) error {
    var specifyErr *SpecifyError
    if errors.As(err, &specifyErr) {
        switch specifyErr.Type {
        case ErrorTypeUserCancelled:
            deh.ui.ShowMessage("操作已取消")
            return nil
        case ErrorTypeNetwork:
            deh.ui.ShowError(fmt.Errorf("网络错误: %v", specifyErr.Message))
            return specifyErr
        default:
            deh.ui.ShowError(specifyErr)
            return specifyErr
        }
    }
    
    deh.ui.ShowError(err)
    return err
}
```

## 7. 第三方库选择

### 7.1 推荐的Go库
```go
// CLI框架
"github.com/spf13/cobra"     // 命令行框架
"github.com/spf13/viper"     // 配置管理

// UI和交互
"github.com/charmbracelet/lipgloss"  // 样式系统
"github.com/charmbracelet/bubbles"   // UI组件
"github.com/pterm/pterm"             // 终端UI库
"github.com/manifoldco/promptui"     // 交互式提示

// HTTP和网络
"github.com/go-resty/resty/v2"       // HTTP客户端
"github.com/schollz/progressbar/v3"  // 进度条

// 文件和归档
"github.com/mholt/archiver/v4"       // 归档处理
"github.com/otiai10/copy"            // 文件复制

// 系统集成
"github.com/shirou/gopsutil/v3"      // 系统信息
"golang.org/x/sys"                   // 系统调用

// 测试
"github.com/stretchr/testify"        // 测试框架
"github.com/golang/mock"             // Mock生成
```

## 8. 并发和性能优化

### 8.1 并发下载策略
```go
type ConcurrentDownloader struct {
    maxConcurrency int
    semaphore      chan struct{}
    wg             sync.WaitGroup
}

func (cd *ConcurrentDownloader) DownloadMultiple(urls []string, destDir string) error {
    cd.semaphore = make(chan struct{}, cd.maxConcurrency)
    
    for _, url := range urls {
        cd.wg.Add(1)
        go func(u string) {
            defer cd.wg.Done()
            cd.semaphore <- struct{}{} // 获取信号量
            defer func() { <-cd.semaphore }() // 释放信号量
            
            // 执行下载
            cd.downloadSingle(u, destDir)
        }(url)
    }
    
    cd.wg.Wait()
    return nil
}
```

### 8.2 缓存机制
```go
type CacheManager struct {
    cacheDir    string
    ttl         time.Duration
    mutex       sync.RWMutex
    memoryCache map[string]CacheEntry
}

type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}

func (cm *CacheManager) Get(key string) (interface{}, bool) {
    cm.mutex.RLock()
    defer cm.mutex.RUnlock()
    
    entry, exists := cm.memoryCache[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    
    return entry.Data, true
}
```


## 9. 构建和部署

### 9.1 Makefile
```makefile
.PHONY: build test clean install

# 构建配置
BINARY_NAME=specify-cli
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# 构建目标
build:
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd

# 跨平台构建
build-all:
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 ./cmd
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 ./cmd

# 测试
test:
	go test -v -race -coverprofile=coverage.out ./...

# 代码检查
lint:
	golangci-lint run

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out

# 安装
install:
	go install ${LDFLAGS} ./cmd
```

### 9.2 CI/CD配置 (.github/workflows/ci.yml)
```yaml
name: CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: make test
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Build all platforms
      run: make build-all
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: binaries
        path: bin/
```

## 10. 关键功能补充说明

### 10.1 实时UI刷新机制
基于之前分析发现的遗漏，Go版本需要实现类似Python版本的`attach_refresh()`机制：

```go
type LiveRenderer struct {
    live     *live.Live
    callback func()
    mutex    sync.RWMutex
}

func (lr *LiveRenderer) Start() error {
    lr.live = live.New()
    return lr.live.Start()
}

func (lr *LiveRenderer) AttachCallback(callback func()) {
    lr.mutex.Lock()
    defer lr.mutex.Unlock()
    lr.callback = callback
}

func (lr *LiveRenderer) Refresh() {
    lr.mutex.RLock()
    callback := lr.callback
    lr.mutex.RUnlock()
    
    if callback != nil {
        callback()
    }
}
```

### 10.2 复杂ZIP解压逻辑
处理嵌套目录扁平化和文件合并：

```go
func (ae *ArchiveExtractor) handleNestedDirectories(extractPath string) error {
    entries, err := os.ReadDir(extractPath)
    if err != nil {
        return err
    }
    
    // 如果只有一个目录，且该目录包含所有内容，则扁平化
    if len(entries) == 1 && entries[0].IsDir() {
        nestedDir := filepath.Join(extractPath, entries[0].Name())
        return ae.flattenDirectory(nestedDir, extractPath)
    }
    
    return nil
}
```

### 10.3 安全提示和环境指导
确保用户了解安全最佳实践和正确的环境配置。

## 11. 总结

本概要设计文档基于对Python版本的详细分析，补充了之前遗漏的关键功能：

1. **实时UI刷新机制** - 支持动态更新进度显示
2. **增强的键盘交互** - 支持Vim风格和Ctrl快捷键
3. **复杂文件操作** - 处理ZIP解压的边缘情况和文件合并
4. **特殊工具检查** - 处理Claude CLI等工具的特殊路径
5. **安全提示系统** - 提醒用户注意敏感信息保护
6. **完整的Agent配置** - 包含所有13个AI助手的配置
7. **环境设置指导** - 帮助用户正确配置开发环境

该设计充分利用了Go语言的特性，提供了模块化、可测试、高性能的架构，确保与Python版本功能完全对等的同时，还具备更好的性能和维护性。