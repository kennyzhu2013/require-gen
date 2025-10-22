# Specify CLI 项目源代码深度分析

## 项目概述

Specify CLI 是一个用于初始化规范驱动开发项目的命令行工具，支持多种AI助手集成（如GitHub Copilot、Claude Code、Gemini CLI等）。项目采用分层架构设计，具有良好的模块化和可扩展性。

## 项目整体架构

### 目录结构
```
require-gen/
├── cmd/specify/           # 主程序入口
├── internal/             # 内部包
│   ├── cli/             # CLI层 - 命令行接口
│   ├── business/        # 业务逻辑层
│   ├── infrastructure/  # 基础设施层
│   ├── config/         # 配置管理
│   ├── types/          # 类型定义
│   └── ui/             # 用户界面层
├── examples/           # 示例文件
├── .github/           # GitHub配置
└── .claude/           # Claude配置
```

### 架构分层

1. **表示层 (Presentation Layer)**
   - CLI层：处理命令行交互
   - UI层：用户界面管理

2. **业务逻辑层 (Business Logic Layer)**
   - 初始化处理器 (InitHandler)
   - 下载处理器 (DownloadHandler)

3. **基础设施层 (Infrastructure Layer)**
   - 系统操作、Git操作、模板提供、认证等

4. **数据层 (Data Layer)**
   - 配置管理、类型定义

## 详细分析

### 1. 主程序入口点 (cmd/specify/main.go)

#### 核心函数
- **`main()`**: 程序主入口点
  - 调用 `cli.Execute()` 启动CLI应用
  - 错误处理和程序退出

#### 调用关系
```
main() → cli.Execute()
```

### 2. CLI层 (internal/cli/)

#### 核心文件和函数

**cli.go - CLI核心管理**
- **`Execute()`**: CLI应用主入口点
  - 执行根命令 `rootCmd.Execute()`
  - 全局错误处理
- **`init()`**: 初始化函数
  - 添加全局标志 (verbose, debug)
  - 注册子命令 (initCmd, downloadCmd, checkCmd, versionCmd, configCmd)
- **辅助函数**:
  - `printError()`: 错误输出
  - `printVerbose()`: 详细信息输出
  - `printDebug()`: 调试信息输出

**init.go - 初始化命令**
- **`runInit()`**: 初始化命令执行函数
  - 解析命令行参数
  - 构建 `types.InitOptions`
  - 调用 `business.NewInitHandler().Execute()`
- **`validateInitOptions()`**: 验证初始化选项
  - AI助手验证
  - 脚本类型验证
  - 项目目录验证
- **`initHelpFunc()`**: 自定义帮助函数

**download.go - 下载命令**
- **`runDownload()`**: 下载命令执行函数
  - 解析AI助手参数
  - 构建 `types.DownloadOptions`
  - 调用 `business.NewDownloadHandler().Execute()`

#### 调用关系
```
CLI层调用关系:
Execute() → rootCmd.Execute()
runInit() → business.NewInitHandler().Execute()
runDownload() → business.NewDownloadHandler().Execute()
```

### 3. 业务逻辑层 (internal/business/)

#### init.go - 初始化处理器

**InitHandler 结构体**
```go
type InitHandler struct {
    templateProvider types.TemplateProvider
    authProvider     types.AuthProvider
    uiRenderer       types.UIRenderer
    gitOps          types.GitOperations
    toolChecker     types.ToolChecker
    sysOps          types.SystemOperations
}
```

**核心函数**
- **`NewInitHandler()`**: 构造函数
  - 创建所有依赖组件实例
- **`Execute(opts types.InitOptions)`**: 主执行函数
  - 创建步骤跟踪器
  - 调用 `setupSteps()` 设置步骤
  - 调用 `executeSteps()` 执行步骤
- **`setupSteps()`**: 设置初始化步骤
  - 添加9个核心步骤：验证、选择AI、选择脚本、检查工具、创建目录、下载模板、初始化Git、配置项目、完成设置
- **`executeSteps()`**: 执行所有步骤
  - 按顺序执行每个步骤
  - 错误处理和进度显示

#### download.go - 下载处理器

**DownloadHandler 结构体**
```go
type DownloadHandler struct {
    templateProvider types.TemplateProvider
    authProvider     types.AuthProvider
    uiRenderer       types.UIRenderer
}
```

**核心函数**
- **`NewDownloadHandler()`**: 构造函数
- **`Execute(opts types.DownloadOptions)`**: 主执行函数
  - 创建步骤跟踪器
  - 执行6个核心步骤：验证、准备环境、下载、提取、验证完整性、清理

#### 调用关系
```
业务逻辑层调用关系:
InitHandler.Execute() → setupSteps() → executeSteps()
DownloadHandler.Execute() → 6个步骤的顺序执行
```

### 4. 基础设施层 (internal/infrastructure/)

#### 核心组件

**system.go - 系统操作**
- **`SystemOperations`**: 系统级操作接口实现
- **核心函数**:
  - `NewSystemOperations()`: 构造函数
  - `GetOS()`, `GetArch()`: 系统信息获取
  - `ExecuteCommand()`: 命令执行
  - `CreateDirectory()`, `RemoveDirectory()`: 目录操作
  - `FileExists()`, `ReadFile()`, `WriteFile()`: 文件操作
  - 安全验证函数: `IsPathSafe()`, `IsCommandSafe()`

**template.go - 模板提供器**
- **`TemplateProvider`**: 模板下载和管理
- **核心函数**:
  - `NewTemplateProvider()`: 构造函数
  - `Download()`: 模板下载主函数
  - `getLatestRelease()`: 获取GitHub最新发布
  - `findAsset()`: 查找资产文件
  - `downloadAsset()`: 下载资产
  - `extractAsset()`: 提取资产文件

**git.go - Git操作**
- **`GitOperations`**: Git操作接口实现
- Git仓库初始化、提交等操作

**auth.go - 认证提供器**
- **`AuthProvider`**: 认证和令牌管理
- 支持从CLI参数、环境变量获取令牌

**tools.go - 工具检查器**
- **`ToolChecker`**: AI助手工具可用性检查
- **核心函数**:
  - `NewToolChecker()`: 构造函数
  - `CheckTool()`: 检查单个工具
  - `CheckAllTools()`: 检查所有工具
  - `GetToolVersion()`: 获取工具版本

**ziputil.go / tarutil.go - 压缩文件处理**
- **`ZipProcessor`**: ZIP文件处理接口
- **`TarProcessor`**: TAR文件处理接口
- 文件解压、验证、进度显示等功能

#### 调用关系
```
基础设施层调用关系:
TemplateProvider.Download() → getLatestRelease() → findAsset() → downloadAsset() → extractAsset()
SystemOperations → 各种系统级操作函数
ToolChecker.CheckAllTools() → CheckTool() (循环调用)
```

### 5. 配置管理 (internal/config/)

#### config.go
- **`AgentConfig`**: 全局AI助手配置映射
- 包含GitHub Copilot、Claude Code、Gemini CLI等配置信息
- 每个配置包含：名称、文件夹、安装URL、是否需要CLI工具

### 6. 类型定义 (internal/types/)

#### types.go
**核心数据结构**:
- **网络配置**: `TLSConfig`, `NetworkConfig`, `HTTPClientConfig`
- **进度显示**: `ProgressInfo`, `ProgressDisplay`
- **错误处理**: `NetworkErrorType`, `NetworkError`
- **AI助手**: `AgentInfo`, `AgentOption`
- **选项结构**: `InitOptions`, `DownloadOptions`
- **GitHub集成**: `GitHubRelease`, `Asset`
- **步骤管理**: `Step`, `StepTracker`
- **接口定义**: `SystemOperations`, `TemplateProvider`, `AuthProvider`, `GitOperations`, `ToolChecker`, `UIRenderer`

### 7. UI层 (internal/ui/)

#### ui.go
- **`UIManager`**: 用户界面管理核心组件
- **功能**:
  - 颜色输出控制
  - 交互式提示
  - 界面状态管理
  - 跨平台兼容
  - 主题管理
  - 进度显示

## 函数间调用关系图

### 主要调用链路

1. **初始化流程**:
```
main() → cli.Execute() → runInit() → InitHandler.Execute() → setupSteps() → executeSteps()
```

2. **下载流程**:
```
main() → cli.Execute() → runDownload() → DownloadHandler.Execute() → TemplateProvider.Download()
```

3. **模板下载详细流程**:
```
TemplateProvider.Download() → getLatestRelease() → findAsset() → downloadAsset() → extractAsset()
```

4. **系统操作流程**:
```
各业务函数 → SystemOperations.* → 具体系统调用
```

### 依赖关系

- **CLI层** 依赖 **业务逻辑层**
- **业务逻辑层** 依赖 **基础设施层**
- **所有层** 依赖 **类型定义层** 和 **配置层**
- **UI层** 被各层调用用于用户交互

## 关键设计模式

1. **分层架构模式**: 清晰的层次分离
2. **依赖注入模式**: 通过接口注入依赖
3. **策略模式**: 不同AI助手的配置策略
4. **观察者模式**: 步骤执行的进度观察
5. **工厂模式**: 各种组件的创建函数

## 扩展性分析

1. **新增AI助手**: 只需在config.go中添加配置
2. **新增命令**: 在CLI层添加新的命令文件
3. **新增功能**: 在业务逻辑层添加新的处理器
4. **新增基础设施**: 在infrastructure层添加新的组件

## 总结

该项目采用了良好的分层架构设计，具有以下特点：

1. **模块化程度高**: 各层职责清晰，耦合度低
2. **可扩展性强**: 通过接口和配置实现灵活扩展
3. **错误处理完善**: 各层都有相应的错误处理机制
4. **用户体验好**: 提供进度显示、颜色输出等用户友好功能
5. **跨平台支持**: 通过SystemOperations实现跨平台兼容
6. **安全性考虑**: 包含路径验证、命令安全检查等安全机制

整个项目的函数调用关系清晰，从主入口点开始，通过CLI层解析命令，调用业务逻辑层处理具体业务，再通过基础设施层执行底层操作，形成了完整的调用链路。