# Specify CLI Go版本单元测试用例文档

## 文档信息
- **版本**: 1.0.0
- **创建日期**: 2024年12月
- **基于**: DetailedDesign.md v1.0
- **目标**: 提供Specify CLI Go版本的完整单元测试用例规范

## 目录

1. [测试概述](#1-测试概述)
2. [测试架构设计](#2-测试架构设计)
3. [CLI层测试用例](#3-cli层测试用例)
4. [服务层测试用例](#4-服务层测试用例)
5. [UI组件测试用例](#5-ui组件测试用例)
6. [GitHub集成测试用例](#6-github集成测试用例)
7. [系统集成测试用例](#7-系统集成测试用例)
8. [基础设施层测试用例](#8-基础设施层测试用例)
9. [数据模型测试用例](#9-数据模型测试用例)
10. [错误处理测试用例](#10-错误处理测试用例)
11. [并发安全测试用例](#11-并发安全测试用例)
12. [Mock对象设计](#12-mock对象设计)
13. [测试工具和框架](#13-测试工具和框架)
14. [测试覆盖率要求](#14-测试覆盖率要求)
15. [测试执行策略](#15-测试执行策略)

---

## 1. 测试概述

### 1.1 测试目标
本文档定义了Specify CLI Go版本的完整单元测试用例，确保：

1. **功能正确性**: 验证所有功能按设计要求正确实现
2. **边界条件**: 测试各种边界条件和异常情况
3. **接口契约**: 验证所有接口的输入输出契约
4. **并发安全**: 确保并发场景下的数据一致性
5. **错误处理**: 验证错误处理机制的完整性

### 1.2 测试原则

1. **单一职责**: 每个测试用例只验证一个功能点
2. **独立性**: 测试用例之间相互独立，不依赖执行顺序
3. **可重复性**: 测试结果可重复，不受环境影响
4. **快速执行**: 单元测试应快速执行，提供即时反馈
5. **清晰命名**: 测试用例名称清晰表达测试意图

### 1.3 测试框架选择

```go
// 主要测试依赖
"github.com/stretchr/testify/assert"    // 断言库
"github.com/stretchr/testify/mock"      // Mock框架
"github.com/stretchr/testify/suite"     // 测试套件
"github.com/golang/mock/gomock"         // Mock生成器
"github.com/onsi/ginkgo/v2"            // BDD测试框架
"github.com/onsi/gomega"               // 匹配器库
```

---

## 2. 测试架构设计

### 2.1 测试目录结构

```
specify-cli-go/
├── internal/
│   ├── cli/
│   │   ├── commands/
│   │   │   ├── root_test.go
│   │   │   ├── init_test.go
│   │   │   └── check_test.go
│   │   ├── ui/
│   │   │   ├── step_tracker_test.go
│   │   │   ├── selector_test.go
│   │   │   ├── banner_test.go
│   │   │   └── ui_impl_test.go
│   │   └── validation/
│   │       └── input_validator_test.go
│   ├── core/
│   │   ├── services/
│   │   │   ├── init_service_test.go
│   │   │   ├── check_service_test.go
│   │   │   └── template_service_test.go
│   │   ├── downloaders/
│   │   │   ├── github_downloader_test.go
│   │   │   └── archive_extractor_test.go
│   │   └── checkers/
│   │       ├── tool_checker_test.go
│   │       └── special_tool_checker_test.go
│   ├── infrastructure/
│   │   ├── config/
│   │   │   └── config_manager_test.go
│   │   ├── file/
│   │   │   └── file_manager_test.go
│   │   └── permission/
│   │       └── permission_manager_test.go
│   └── models/
│       ├── agent_test.go
│       ├── step_test.go
│       └── tool_test.go
├── mocks/                              # 生成的Mock文件
│   ├── mock_services.go
│   ├── mock_ui.go
│   └── mock_infrastructure.go
└── testdata/                          # 测试数据
    ├── templates/
    ├── configs/
    └── fixtures/
```

### 2.2 测试基础设施

#### 2.2.1 测试基类设计
```go
// internal/testutil/base_test.go
package testutil

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/suite"
    "github.com/golang/mock/gomock"
)

// BaseTestSuite 基础测试套件
type BaseTestSuite struct {
    suite.Suite
    ctx        context.Context
    cancel     context.CancelFunc
    ctrl       *gomock.Controller
    timeout    time.Duration
}

// SetupSuite 套件级别的设置
func (s *BaseTestSuite) SetupSuite() {
    s.timeout = 30 * time.Second
}

// SetupTest 测试级别的设置
func (s *BaseTestSuite) SetupTest() {
    s.ctx, s.cancel = context.WithTimeout(context.Background(), s.timeout)
    s.ctrl = gomock.NewController(s.T())
}

// TearDownTest 测试级别的清理
func (s *BaseTestSuite) TearDownTest() {
    if s.cancel != nil {
        s.cancel()
    }
    if s.ctrl != nil {
        s.ctrl.Finish()
    }
}

// AssertNoError 断言无错误
func (s *BaseTestSuite) AssertNoError(err error, msgAndArgs ...interface{}) {
    s.Require().NoError(err, msgAndArgs...)
}

// AssertError 断言有错误
func (s *BaseTestSuite) AssertError(err error, msgAndArgs ...interface{}) {
    s.Require().Error(err, msgAndArgs...)
}
```

---

## 3. CLI层测试用例

### 3.1 根命令测试用例

#### 3.1.1 RootCommand测试
```go
// internal/cli/commands/root_test.go
package commands_test

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/cli/commands"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type RootCommandTestSuite struct {
    testutil.BaseTestSuite
    mockConfig   *mocks.MockConfigManager
    mockBanner   *mocks.MockBanner
    mockEventBus *mocks.MockEventBus
    rootCmd      *commands.RootCommand
}

func (s *RootCommandTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockConfig = mocks.NewMockConfigManager(s.ctrl)
    s.mockBanner = mocks.NewMockBanner(s.ctrl)
    s.mockEventBus = mocks.NewMockEventBus(s.ctrl)
    
    s.rootCmd = commands.NewRootCommand(
        s.mockConfig,
        s.mockBanner,
        s.mockEventBus,
    )
}

// TestRootCommand_NewRootCommand 测试根命令创建
func (s *RootCommandTestSuite) TestRootCommand_NewRootCommand() {
    // Given
    // 已在SetupTest中创建
    
    // When & Then
    s.NotNil(s.rootCmd)
    s.Equal("specify-cli", s.rootCmd.GetCommand().Use)
    s.Contains(s.rootCmd.GetCommand().Short, "AI辅助开发项目初始化工具")
}

// TestRootCommand_GlobalFlags 测试全局标志
func (s *RootCommandTestSuite) TestRootCommand_GlobalFlags() {
    // Given
    cmd := s.rootCmd.GetCommand()
    
    // When
    flags := cmd.PersistentFlags()
    
    // Then
    s.True(flags.HasFlag("debug"))
    s.True(flags.HasFlag("verbose"))
    s.True(flags.HasFlag("no-color"))
    s.True(flags.HasFlag("config-dir"))
}

// TestRootCommand_PersistentPreRun_DebugMode 测试调试模式设置
func (s *RootCommandTestSuite) TestRootCommand_PersistentPreRun_DebugMode() {
    // Given
    cmd := s.rootCmd.GetCommand()
    cmd.SetArgs([]string{"--debug"})
    
    s.mockConfig.EXPECT().SetLogLevel("debug").Times(1)
    
    // When
    err := cmd.Execute()
    
    // Then
    s.AssertNoError(err)
}

// TestRootCommand_PersistentPreRun_VerboseMode 测试详细模式设置
func (s *RootCommandTestSuite) TestRootCommand_PersistentPreRun_VerboseMode() {
    // Given
    cmd := s.rootCmd.GetCommand()
    cmd.SetArgs([]string{"--verbose"})
    
    s.mockConfig.EXPECT().SetLogLevel("info").Times(1)
    
    // When
    err := cmd.Execute()
    
    // Then
    s.AssertNoError(err)
}

// TestRootCommand_PersistentPreRun_NoColor 测试禁用颜色设置
func (s *RootCommandTestSuite) TestRootCommand_PersistentPreRun_NoColor() {
    // Given
    cmd := s.rootCmd.GetCommand()
    cmd.SetArgs([]string{"--no-color"})
    
    s.mockConfig.EXPECT().SetColorOutput(false).Times(1)
    
    // When
    err := cmd.Execute()
    
    // Then
    s.AssertNoError(err)
}

// TestRootCommand_Run_ShowsBanner 测试显示横幅
func (s *RootCommandTestSuite) TestRootCommand_Run_ShowsBanner() {
    // Given
    s.mockBanner.EXPECT().Show().Return(nil).Times(1)
    
    // When
    err := s.rootCmd.GetCommand().Execute()
    
    // Then
    s.AssertNoError(err)
}

// TestRootCommand_Run_BannerError 测试横幅显示错误
func (s *RootCommandTestSuite) TestRootCommand_Run_BannerError() {
    // Given
    expectedErr := errors.New("banner display failed")
    s.mockBanner.EXPECT().Show().Return(expectedErr).Times(1)
    
    // When
    err := s.rootCmd.GetCommand().Execute()
    
    // Then
    s.AssertError(err)
    s.Contains(err.Error(), "显示横幅失败")
}

func TestRootCommandTestSuite(t *testing.T) {
    suite.Run(t, new(RootCommandTestSuite))
}
```

### 3.2 Init命令测试用例

#### 3.2.1 InitCommand测试
```go
// internal/cli/commands/init_test.go
package commands_test

import (
    "errors"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/cli/commands"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type InitCommandTestSuite struct {
    testutil.BaseTestSuite
    mockService   *mocks.MockInitService
    mockUI        *mocks.MockUI
    mockValidator *mocks.MockInputValidator
    initCmd       *commands.InitCommand
}

func (s *InitCommandTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockService = mocks.NewMockInitService(s.ctrl)
    s.mockUI = mocks.NewMockUI(s.ctrl)
    s.mockValidator = mocks.NewMockInputValidator(s.ctrl)
    
    s.initCmd = commands.NewInitCommand(
        s.mockService,
        s.mockUI,
        s.mockValidator,
    )
}

// TestInitCommand_NewInitCommand 测试Init命令创建
func (s *InitCommandTestSuite) TestInitCommand_NewInitCommand() {
    // Given
    // 已在SetupTest中创建
    
    // When & Then
    s.NotNil(s.initCmd)
    s.Equal("init [PROJECT_NAME]", s.initCmd.GetCommand().Use)
    s.Contains(s.initCmd.GetCommand().Short, "初始化AI辅助开发项目")
}

// TestInitCommand_Flags 测试命令标志
func (s *InitCommandTestSuite) TestInitCommand_Flags() {
    // Given
    cmd := s.initCmd.GetCommand()
    
    // When
    flags := cmd.Flags()
    
    // Then
    s.True(flags.HasFlag("ai"))
    s.True(flags.HasFlag("script-type"))
    s.True(flags.HasFlag("ignore-agent-tools"))
    s.True(flags.HasFlag("no-git"))
    s.True(flags.HasFlag("here"))
    s.True(flags.HasFlag("force"))
    s.True(flags.HasFlag("skip-tls"))
    s.True(flags.HasFlag("github-token"))
}

// TestInitCommand_Run_WithProjectName 测试带项目名称的执行
func (s *InitCommandTestSuite) TestInitCommand_Run_WithProjectName() {
    // Given
    projectName := "test-project"
    cmd := s.initCmd.GetCommand()
    cmd.SetArgs([]string{projectName})
    
    expectedArgs := &models.InitArgs{
        ProjectName: projectName,
        AI:         "claude",
        ScriptType: "bash",
    }
    
    s.mockValidator.EXPECT().ValidateInitArgs(gomock.Any()).Return(nil).Times(1)
    s.mockService.EXPECT().InitializeProject(s.ctx, expectedArgs).Return(nil).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertNoError(err)
}

// TestInitCommand_Run_InteractiveInput 测试交互式输入
func (s *InitCommandTestSuite) TestInitCommand_Run_InteractiveInput() {
    // Given
    cmd := s.initCmd.GetCommand()
    
    s.mockUI.EXPECT().PromptInput("请输入项目名称:", "").Return("my-project", nil).Times(1)
    s.mockService.EXPECT().GetAvailableAIs().Return([]string{"claude", "copilot"}).Times(1)
    s.mockUI.EXPECT().PromptSelect("请选择AI助手:", []string{"claude", "copilot"}).Return("claude", nil).Times(1)
    s.mockService.EXPECT().GetAvailableScriptTypes().Return([]string{"bash", "powershell"}).Times(1)
    s.mockUI.EXPECT().PromptSelect("请选择脚本类型:", []string{"bash", "powershell"}).Return("bash", nil).Times(1)
    
    s.mockValidator.EXPECT().ValidateInitArgs(gomock.Any()).Return(nil).Times(1)
    s.mockService.EXPECT().InitializeProject(s.ctx, gomock.Any()).Return(nil).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertNoError(err)
}

// TestInitCommand_Run_ValidationError 测试验证错误
func (s *InitCommandTestSuite) TestInitCommand_Run_ValidationError() {
    // Given
    cmd := s.initCmd.GetCommand()
    cmd.SetArgs([]string{"invalid-project"})
    
    expectedErr := errors.New("项目名称验证失败")
    s.mockValidator.EXPECT().ValidateInitArgs(gomock.Any()).Return(expectedErr).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertError(err)
    s.Contains(err.Error(), "构建初始化参数失败")
}

// TestInitCommand_Run_ServiceError 测试服务错误
func (s *InitCommandTestSuite) TestInitCommand_Run_ServiceError() {
    // Given
    cmd := s.initCmd.GetCommand()
    cmd.SetArgs([]string{"test-project"})
    
    s.mockValidator.EXPECT().ValidateInitArgs(gomock.Any()).Return(nil).Times(1)
    
    expectedErr := errors.New("初始化失败")
    s.mockService.EXPECT().InitializeProject(s.ctx, gomock.Any()).Return(expectedErr).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertError(err)
    s.Contains(err.Error(), "项目初始化失败")
}

// TestInitCommand_BuildInitArgs_AllFlags 测试所有标志的参数构建
func (s *InitCommandTestSuite) TestInitCommand_BuildInitArgs_AllFlags() {
    // Given
    cmd := s.initCmd.GetCommand()
    cmd.SetArgs([]string{
        "test-project",
        "--ai", "claude",
        "--script-type", "bash",
        "--ignore-agent-tools",
        "--no-git",
        "--here",
        "--force",
        "--skip-tls",
        "--github-token", "token123",
    })
    
    s.mockValidator.EXPECT().ValidateInitArgs(gomock.Any()).DoAndReturn(
        func(args *models.InitArgs) error {
            s.Equal("test-project", args.ProjectName)
            s.Equal("claude", args.AI)
            s.Equal("bash", args.ScriptType)
            s.True(args.IgnoreAgentTools)
            s.True(args.NoGit)
            s.True(args.Here)
            s.True(args.Force)
            s.True(args.SkipTLS)
            s.Equal("token123", args.GitHubToken)
            return nil
        },
    ).Times(1)
    
    s.mockService.EXPECT().InitializeProject(s.ctx, gomock.Any()).Return(nil).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertNoError(err)
}

func TestInitCommandTestSuite(t *testing.T) {
    suite.Run(t, new(InitCommandTestSuite))
}
```

### 3.3 Check命令测试用例

#### 3.3.1 CheckCommand测试
```go
// internal/cli/commands/check_test.go
package commands_test

import (
    "errors"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/cli/commands"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type CheckCommandTestSuite struct {
    testutil.BaseTestSuite
    mockService *mocks.MockCheckService
    mockUI      *mocks.MockUI
    checkCmd    *commands.CheckCommand
}

func (s *CheckCommandTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockService = mocks.NewMockCheckService(s.ctrl)
    s.mockUI = mocks.NewMockUI(s.ctrl)
    
    s.checkCmd = commands.NewCheckCommand(
        s.mockService,
        s.mockUI,
    )
}

// TestCheckCommand_NewCheckCommand 测试Check命令创建
func (s *CheckCommandTestSuite) TestCheckCommand_NewCheckCommand() {
    // Given
    // 已在SetupTest中创建
    
    // When & Then
    s.NotNil(s.checkCmd)
    s.Equal("check", s.checkCmd.GetCommand().Use)
    s.Contains(s.checkCmd.GetCommand().Short, "检查开发环境工具")
}

// TestCheckCommand_Run_AllToolsInstalled 测试所有工具已安装
func (s *CheckCommandTestSuite) TestCheckCommand_Run_AllToolsInstalled() {
    // Given
    cmd := s.checkCmd.GetCommand()
    
    toolStatuses := []models.ToolStatus{
        {Name: "git", Installed: true, Version: "2.40.0"},
        {Name: "node", Installed: true, Version: "18.17.0"},
        {Name: "python", Installed: true, Version: "3.11.0"},
    }
    
    s.mockService.EXPECT().CheckAllTools(s.ctx).Return(toolStatuses, nil).Times(1)
    s.mockUI.EXPECT().ShowSuccess("所有工具检查完成").Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertNoError(err)
}

// TestCheckCommand_Run_SomeToolsMissing 测试部分工具缺失
func (s *CheckCommandTestSuite) TestCheckCommand_Run_SomeToolsMissing() {
    // Given
    cmd := s.checkCmd.GetCommand()
    
    toolStatuses := []models.ToolStatus{
        {Name: "git", Installed: true, Version: "2.40.0"},
        {Name: "node", Installed: false},
        {Name: "python", Installed: false},
    }
    
    installTips := []string{
        "安装Node.js: https://nodejs.org/",
        "安装Python: https://python.org/",
    }
    
    s.mockService.EXPECT().CheckAllTools(s.ctx).Return(toolStatuses, nil).Times(1)
    s.mockService.EXPECT().GenerateInstallationTips([]string{"node", "python"}).Return(installTips).Times(1)
    s.mockUI.EXPECT().ShowWarning(gomock.Any()).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertNoError(err)
}

// TestCheckCommand_Run_CheckError 测试检查错误
func (s *CheckCommandTestSuite) TestCheckCommand_Run_CheckError() {
    // Given
    cmd := s.checkCmd.GetCommand()
    
    expectedErr := errors.New("工具检查失败")
    s.mockService.EXPECT().CheckAllTools(s.ctx).Return(nil, expectedErr).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertError(err)
    s.Contains(err.Error(), "工具检查失败")
}

// TestCheckCommand_Run_SpecificTool 测试特定工具检查
func (s *CheckCommandTestSuite) TestCheckCommand_Run_SpecificTool() {
    // Given
    cmd := s.checkCmd.GetCommand()
    cmd.SetArgs([]string{"git"})
    
    toolStatus := models.ToolStatus{
        Name:      "git",
        Installed: true,
        Version:   "2.40.0",
        Path:      "/usr/bin/git",
    }
    
    s.mockService.EXPECT().CheckTool(s.ctx, "git").Return(toolStatus, nil).Times(1)
    s.mockUI.EXPECT().ShowInfo(gomock.Any()).Times(1)
    
    // When
    err := cmd.ExecuteContext(s.ctx)
    
    // Then
    s.AssertNoError(err)
}

func TestCheckCommandTestSuite(t *testing.T) {
    suite.Run(t, new(CheckCommandTestSuite))
}
```

### 3.4 输入验证器测试用例

#### 3.4.1 InputValidator测试
```go
// internal/cli/validation/input_validator_test.go
package validation_test

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/cli/validation"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type InputValidatorTestSuite struct {
    testutil.BaseTestSuite
    mockConfig *mocks.MockConfigManager
    validator  *validation.InputValidator
}

func (s *InputValidatorTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockConfig = mocks.NewMockConfigManager(s.ctrl)
    s.validator = validation.NewInputValidator(s.mockConfig)
}

// TestInputValidator_ValidateProjectName_Valid 测试有效项目名称
func (s *InputValidatorTestSuite) TestInputValidator_ValidateProjectName_Valid() {
    testCases := []string{
        "my-project",
        "MyProject",
        "my_project",
        "project123",
        "a",
        "project-with-long-name-but-under-100-chars",
    }
    
    for _, projectName := range testCases {
        s.Run(projectName, func() {
            // When
            err := s.validator.ValidateProjectName(projectName)
            
            // Then
            s.AssertNoError(err)
        })
    }
}

// TestInputValidator_ValidateProjectName_Invalid 测试无效项目名称
func (s *InputValidatorTestSuite) TestInputValidator_ValidateProjectName_Invalid() {
    testCases := []struct {
        name        string
        projectName string
        expectedErr string
    }{
        {"empty", "", "项目名称不能为空"},
        {"too_long", string(make([]byte, 101)), "项目名称长度不能超过100个字符"},
        {"invalid_chars", "project@name", "项目名称只能包含字母、数字、下划线和连字符"},
        {"spaces", "my project", "项目名称只能包含字母、数字、下划线和连字符"},
        {"reserved_name", "con", "项目名称不能使用保留名称"},
        {"reserved_name_case", "CON", "项目名称不能使用保留名称"},
    }
    
    for _, tc := range testCases {
        s.Run(tc.name, func() {
            // When
            err := s.validator.ValidateProjectName(tc.projectName)
            
            // Then
            s.AssertError(err)
            s.Contains(err.Error(), tc.expectedErr)
        })
    }
}

// TestInputValidator_ValidateAI_Valid 测试有效AI助手
func (s *InputValidatorTestSuite) TestInputValidator_ValidateAI_Valid() {
    // Given
    availableAIs := []string{"claude", "copilot", "gemini"}
    s.mockConfig.EXPECT().GetAvailableAIs().Return(availableAIs).Times(3)
    
    for _, ai := range availableAIs {
        s.Run(ai, func() {
            // When
            err := s.validator.ValidateAI(ai)
            
            // Then
            s.AssertNoError(err)
        })
    }
}

// TestInputValidator_ValidateAI_Invalid 测试无效AI助手
func (s *InputValidatorTestSuite) TestInputValidator_ValidateAI_Invalid() {
    testCases := []struct {
        name        string
        ai          string
        expectedErr string
    }{
        {"empty", "", "AI助手不能为空"},
        {"unsupported", "unknown-ai", "不支持的AI助手"},
    }
    
    availableAIs := []string{"claude", "copilot"}
    s.mockConfig.EXPECT().GetAvailableAIs().Return(availableAIs).AnyTimes()
    
    for _, tc := range testCases {
        s.Run(tc.name, func() {
            // When
            err := s.validator.ValidateAI(tc.ai)
            
            // Then
            s.AssertError(err)
            s.Contains(err.Error(), tc.expectedErr)
        })
    }
}

// TestInputValidator_ValidateScriptType_Valid 测试有效脚本类型
func (s *InputValidatorTestSuite) TestInputValidator_ValidateScriptType_Valid() {
    // Given
    availableTypes := []string{"bash", "powershell"}
    s.mockConfig.EXPECT().GetAvailableScriptTypes().Return(availableTypes).Times(2)
    
    for _, scriptType := range availableTypes {
        s.Run(scriptType, func() {
            // When
            err := s.validator.ValidateScriptType(scriptType)
            
            // Then
            s.AssertNoError(err)
        })
    }
}

// TestInputValidator_ValidateGitHubToken_Valid 测试有效GitHub令牌
func (s *InputValidatorTestSuite) TestInputValidator_ValidateGitHubToken_Valid() {
    testCases := []string{
        "ghp_1234567890abcdef1234567890abcdef12345678",
        "github_pat_11ABCDEFG0123456789_abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ",
    }
    
    for _, token := range testCases {
        s.Run(token[:10]+"...", func() {
            // When
            err := s.validator.ValidateGitHubToken(token)
            
            // Then
            s.AssertNoError(err)
        })
    }
}

// TestInputValidator_ValidateInitArgs_Complete 测试完整初始化参数验证
func (s *InputValidatorTestSuite) TestInputValidator_ValidateInitArgs_Complete() {
    // Given
    args := &models.InitArgs{
        ProjectName: "test-project",
        AI:         "claude",
        ScriptType: "bash",
    }
    
    availableAIs := []string{"claude", "copilot"}
    availableTypes := []string{"bash", "powershell"}
    
    s.mockConfig.EXPECT().GetAvailableAIs().Return(availableAIs).Times(1)
    s.mockConfig.EXPECT().GetAvailableScriptTypes().Return(availableTypes).Times(1)
    
    // When
    err := s.validator.ValidateInitArgs(args)
    
    // Then
    s.AssertNoError(err)
}

func TestInputValidatorTestSuite(t *testing.T) {
    suite.Run(t, new(InputValidatorTestSuite))
}
```

---

## 4. 服务层测试用例

### 4.1 InitService测试用例

#### 4.1.1 InitService核心功能测试
```go
// internal/core/services/init_service_test.go
package services_test

import (
    "errors"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/core/services"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type InitServiceTestSuite struct {
    testutil.BaseTestSuite
    mockTemplateService *mocks.MockTemplateService
    mockFileManager     *mocks.MockFileManager
    mockPermManager     *mocks.MockPermissionManager
    mockUI              *mocks.MockUI
    mockEventBus        *mocks.MockEventBus
    service             services.InitService
}

func (s *InitServiceTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockTemplateService = mocks.NewMockTemplateService(s.ctrl)
    s.mockFileManager = mocks.NewMockFileManager(s.ctrl)
    s.mockPermManager = mocks.NewMockPermissionManager(s.ctrl)
    s.mockUI = mocks.NewMockUI(s.ctrl)
    s.mockEventBus = mocks.NewMockEventBus(s.ctrl)
    
    s.service = services.NewInitService(
        s.mockTemplateService,
        s.mockFileManager,
        s.mockPermManager,
        s.mockUI,
        s.mockEventBus,
    )
}

// TestInitService_InitializeProject_Success 测试项目初始化成功
func (s *InitServiceTestSuite) TestInitService_InitializeProject_Success() {
    // Given
    args := &models.InitArgs{
        ProjectName: "test-project",
        AI:         "claude",
        ScriptType: "bash",
        Here:       false,
        NoGit:      false,
    }
    
    steps := []*models.Step{
        {ID: "create_directory", Name: "创建项目目录", Status: models.StepStatusPending},
        {ID: "download_template", Name: "下载项目模板", Status: models.StepStatusPending},
        {ID: "extract_template", Name: "解压模板文件", Status: models.StepStatusPending},
        {ID: "setup_permissions", Name: "设置文件权限", Status: models.StepStatusPending},
        {ID: "init_git", Name: "初始化Git仓库", Status: models.StepStatusPending},
    }
    
    // 设置期望调用
    s.mockUI.EXPECT().SetSteps(steps).Times(1)
    
    // 创建目录
    s.mockUI.EXPECT().StartStep("create_directory").Times(1)
    s.mockFileManager.EXPECT().CreateDirectory("test-project").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("create_directory").Times(1)
    
    // 下载模板
    s.mockUI.EXPECT().StartStep("download_template").Times(1)
    s.mockTemplateService.EXPECT().DownloadTemplate(s.ctx, "claude").Return("template.zip", nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("download_template").Times(1)
    
    // 解压模板
    s.mockUI.EXPECT().StartStep("extract_template").Times(1)
    s.mockTemplateService.EXPECT().ExtractTemplate(s.ctx, "template.zip", "test-project").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("extract_template").Times(1)
    
    // 设置权限
    s.mockUI.EXPECT().StartStep("setup_permissions").Times(1)
    s.mockPermManager.EXPECT().SetExecutablePermissions("test-project/.specify/scripts").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("setup_permissions").Times(1)
    
    // 初始化Git
    s.mockUI.EXPECT().StartStep("init_git").Times(1)
    s.mockFileManager.EXPECT().InitializeGitRepository("test-project").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("init_git").Times(1)
    
    // 发布事件
    s.mockEventBus.EXPECT().Publish(s.ctx, gomock.Any()).AnyTimes()
    
    // When
    err := s.service.InitializeProject(s.ctx, args)
    
    // Then
    s.AssertNoError(err)
}

// TestInitService_InitializeProject_DirectoryCreationFailed 测试目录创建失败
func (s *InitServiceTestSuite) TestInitService_InitializeProject_DirectoryCreationFailed() {
    // Given
    args := &models.InitArgs{
        ProjectName: "test-project",
        AI:         "claude",
        ScriptType: "bash",
    }
    
    s.mockUI.EXPECT().SetSteps(gomock.Any()).Times(1)
    s.mockUI.EXPECT().StartStep("create_directory").Times(1)
    
    expectedErr := errors.New("权限不足")
    s.mockFileManager.EXPECT().CreateDirectory("test-project").Return(expectedErr).Times(1)
    s.mockUI.EXPECT().FailStep("create_directory", expectedErr).Times(1)
    
    // When
    err := s.service.InitializeProject(s.ctx, args)
    
    // Then
    s.AssertError(err)
    s.Contains(err.Error(), "创建项目目录失败")
}

// TestInitService_InitializeProject_HereFlag 测试在当前目录初始化
func (s *InitServiceTestSuite) TestInitService_InitializeProject_HereFlag() {
    // Given
    args := &models.InitArgs{
        ProjectName: "test-project",
        AI:         "claude",
        ScriptType: "bash",
        Here:       true,
    }
    
    s.mockUI.EXPECT().SetSteps(gomock.Any()).Times(1)
    
    // 跳过目录创建步骤
    s.mockUI.EXPECT().StartStep("download_template").Times(1)
    s.mockTemplateService.EXPECT().DownloadTemplate(s.ctx, "claude").Return("template.zip", nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("download_template").Times(1)
    
    s.mockUI.EXPECT().StartStep("extract_template").Times(1)
    s.mockTemplateService.EXPECT().ExtractTemplate(s.ctx, "template.zip", ".").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("extract_template").Times(1)
    
    // 其他步骤...
    s.mockUI.EXPECT().StartStep(gomock.Any()).AnyTimes()
    s.mockPermManager.EXPECT().SetExecutablePermissions(gomock.Any()).Return(nil).AnyTimes()
    s.mockFileManager.EXPECT().InitializeGitRepository(gomock.Any()).Return(nil).AnyTimes()
    s.mockUI.EXPECT().CompleteStep(gomock.Any()).AnyTimes()
    s.mockEventBus.EXPECT().Publish(s.ctx, gomock.Any()).AnyTimes()
    
    // When
    err := s.service.InitializeProject(s.ctx, args)
    
    // Then
    s.AssertNoError(err)
}

// TestInitService_InitializeProject_NoGitFlag 测试跳过Git初始化
func (s *InitServiceTestSuite) TestInitService_InitializeProject_NoGitFlag() {
    // Given
    args := &models.InitArgs{
        ProjectName: "test-project",
        AI:         "claude",
        ScriptType: "bash",
        NoGit:      true,
    }
    
    s.mockUI.EXPECT().SetSteps(gomock.Any()).Times(1)
    
    // 执行除Git初始化外的所有步骤
    s.mockUI.EXPECT().StartStep("create_directory").Times(1)
    s.mockFileManager.EXPECT().CreateDirectory("test-project").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("create_directory").Times(1)
    
    s.mockUI.EXPECT().StartStep("download_template").Times(1)
    s.mockTemplateService.EXPECT().DownloadTemplate(s.ctx, "claude").Return("template.zip", nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("download_template").Times(1)
    
    s.mockUI.EXPECT().StartStep("extract_template").Times(1)
    s.mockTemplateService.EXPECT().ExtractTemplate(s.ctx, "template.zip", "test-project").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("extract_template").Times(1)
    
    s.mockUI.EXPECT().StartStep("setup_permissions").Times(1)
    s.mockPermManager.EXPECT().SetExecutablePermissions("test-project/.specify/scripts").Return(nil).Times(1)
    s.mockUI.EXPECT().CompleteStep("setup_permissions").Times(1)
    
    // 不应该调用Git初始化
    s.mockFileManager.EXPECT().InitializeGitRepository(gomock.Any()).Times(0)
    
    s.mockEventBus.EXPECT().Publish(s.ctx, gomock.Any()).AnyTimes()
    
    // When
    err := s.service.InitializeProject(s.ctx, args)
    
    // Then
    s.AssertNoError(err)
}

// TestInitService_GetAvailableAIs 测试获取可用AI助手
func (s *InitServiceTestSuite) TestInitService_GetAvailableAIs() {
    // When
    ais := s.service.GetAvailableAIs()
    
    // Then
    s.NotEmpty(ais)
    s.Contains(ais, "claude")
    s.Contains(ais, "copilot")
    s.Contains(ais, "gemini")
}

// TestInitService_GetAvailableScriptTypes 测试获取可用脚本类型
func (s *InitServiceTestSuite) TestInitService_GetAvailableScriptTypes() {
    // When
    types := s.service.GetAvailableScriptTypes()
    
    // Then
    s.NotEmpty(types)
    s.Contains(types, "bash")
    s.Contains(types, "powershell")
}

func TestInitServiceTestSuite(t *testing.T) {
    suite.Run(t, new(InitServiceTestSuite))
}
```

### 4.2 CheckService测试用例

#### 4.2.1 CheckService核心功能测试
```go
// internal/core/services/check_service_test.go
package services_test

import (
    "errors"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/core/services"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type CheckServiceTestSuite struct {
    testutil.BaseTestSuite
    mockToolChecker        *mocks.MockToolChecker
    mockSpecialToolChecker *mocks.MockSpecialToolChecker
    mockConfig             *mocks.MockConfigManager
    service                services.CheckService
}

func (s *CheckServiceTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockToolChecker = mocks.NewMockToolChecker(s.ctrl)
    s.mockSpecialToolChecker = mocks.NewMockSpecialToolChecker(s.ctrl)
    s.mockConfig = mocks.NewMockConfigManager(s.ctrl)
    
    s.service = services.NewCheckService(
        s.mockToolChecker,
        s.mockSpecialToolChecker,
        s.mockConfig,
    )
}

// TestCheckService_CheckAllTools_Success 测试检查所有工具成功
func (s *CheckServiceTestSuite) TestCheckService_CheckAllTools_Success() {
    // Given
    requiredTools := []string{"git", "node", "python"}
    s.mockConfig.EXPECT().GetRequiredTools().Return(requiredTools).Times(1)
    
    expectedStatuses := []models.ToolStatus{
        {Name: "git", Installed: true, Version: "2.40.0", Path: "/usr/bin/git"},
        {Name: "node", Installed: true, Version: "18.17.0", Path: "/usr/bin/node"},
        {Name: "python", Installed: true, Version: "3.11.0", Path: "/usr/bin/python3"},
    }
    
    for i, tool := range requiredTools {
        s.mockToolChecker.EXPECT().CheckTool(s.ctx, tool).Return(expectedStatuses[i], nil).Times(1)
    }
    
    // When
    statuses, err := s.service.CheckAllTools(s.ctx)
    
    // Then
    s.AssertNoError(err)
    s.Len(statuses, 3)
    s.Equal(expectedStatuses, statuses)
}

// TestCheckService_CheckAllTools_SomeToolsMissing 测试部分工具缺失
func (s *CheckServiceTestSuite) TestCheckService_CheckAllTools_SomeToolsMissing() {
    // Given
    requiredTools := []string{"git", "node", "python"}
    s.mockConfig.EXPECT().GetRequiredTools().Return(requiredTools).Times(1)
    
    expectedStatuses := []models.ToolStatus{
        {Name: "git", Installed: true, Version: "2.40.0", Path: "/usr/bin/git"},
        {Name: "node", Installed: false},
        {Name: "python", Installed: false},
    }
    
    for i, tool := range requiredTools {
        s.mockToolChecker.EXPECT().CheckTool(s.ctx, tool).Return(expectedStatuses[i], nil).Times(1)
    }
    
    // When
    statuses, err := s.service.CheckAllTools(s.ctx)
    
    // Then
    s.AssertNoError(err)
    s.Len(statuses, 3)
    
    installedCount := 0
    for _, status := range statuses {
        if status.Installed {
            installedCount++
        }
    }
    s.Equal(1, installedCount)
}

// TestCheckService_CheckTool_Success 测试检查单个工具成功
func (s *CheckServiceTestSuite) TestCheckService_CheckTool_Success() {
    // Given
    toolName := "git"
    expectedStatus := models.ToolStatus{
        Name:      "git",
        Installed: true,
        Version:   "2.40.0",
        Path:      "/usr/bin/git",
    }
    
    s.mockToolChecker.EXPECT().CheckTool(s.ctx, toolName).Return(expectedStatus, nil).Times(1)
    
    // When
    status, err := s.service.CheckTool(s.ctx, toolName)
    
    // Then
    s.AssertNoError(err)
    s.Equal(expectedStatus, status)
}

// TestCheckService_CheckTool_NotInstalled 测试工具未安装
func (s *CheckServiceTestSuite) TestCheckService_CheckTool_NotInstalled() {
    // Given
    toolName := "nonexistent-tool"
    expectedStatus := models.ToolStatus{
        Name:      "nonexistent-tool",
        Installed: false,
    }
    
    s.mockToolChecker.EXPECT().CheckTool(s.ctx, toolName).Return(expectedStatus, nil).Times(1)
    
    // When
    status, err := s.service.CheckTool(s.ctx, toolName)
    
    // Then
    s.AssertNoError(err)
    s.Equal(expectedStatus, status)
    s.False(status.Installed)
}

// TestCheckService_CheckTool_Error 测试工具检查错误
func (s *CheckServiceTestSuite) TestCheckService_CheckTool_Error() {
    // Given
    toolName := "git"
    expectedErr := errors.New("检查工具失败")
    
    s.mockToolChecker.EXPECT().CheckTool(s.ctx, toolName).Return(models.ToolStatus{}, expectedErr).Times(1)
    
    // When
    _, err := s.service.CheckTool(s.ctx, toolName)
    
    // Then
    s.AssertError(err)
    s.Equal(expectedErr, err)
}

// TestCheckService_CheckSpecialTools 测试检查特殊工具
func (s *CheckServiceTestSuite) TestCheckService_CheckSpecialTools() {
    // Given
    aiType := "claude"
    expectedStatus := models.ToolStatus{
        Name:      "claude-cli",
        Installed: true,
        Version:   "1.0.0",
        Path:      "/usr/local/bin/claude",
    }
    
    s.mockSpecialToolChecker.EXPECT().CheckSpecialTool(s.ctx, aiType).Return(expectedStatus, nil).Times(1)
    
    // When
    status, err := s.service.CheckSpecialTool(s.ctx, aiType)
    
    // Then
    s.AssertNoError(err)
    s.Equal(expectedStatus, status)
}

// TestCheckService_GenerateInstallationTips 测试生成安装提示
func (s *CheckServiceTestSuite) TestCheckService_GenerateInstallationTips() {
    // Given
    missingTools := []string{"node", "python", "docker"}
    
    expectedTips := []string{
        "安装Node.js: https://nodejs.org/",
        "安装Python: https://python.org/",
        "安装Docker: https://docker.com/",
    }
    
    s.mockConfig.EXPECT().GetInstallationTips(missingTools).Return(expectedTips).Times(1)
    
    // When
    tips := s.service.GenerateInstallationTips(missingTools)
    
    // Then
    s.Equal(expectedTips, tips)
    s.Len(tips, 3)
}

// TestCheckService_GenerateInstallationTips_Empty 测试空的缺失工具列表
func (s *CheckServiceTestSuite) TestCheckService_GenerateInstallationTips_Empty() {
    // Given
    missingTools := []string{}
    
    s.mockConfig.EXPECT().GetInstallationTips(missingTools).Return([]string{}).Times(1)
    
    // When
    tips := s.service.GenerateInstallationTips(missingTools)
    
    // Then
    s.Empty(tips)
}

func TestCheckServiceTestSuite(t *testing.T) {
    suite.Run(t, new(CheckServiceTestSuite))
}
```

### 4.3 TemplateService测试用例

#### 4.3.1 TemplateService核心功能测试
```go
// internal/core/services/template_service_test.go
package services_test

import (
    "errors"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/core/services"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type TemplateServiceTestSuite struct {
    testutil.BaseTestSuite
    mockDownloader *mocks.MockGitHubDownloader
    mockExtractor  *mocks.MockArchiveExtractor
    mockConfig     *mocks.MockConfigManager
    service        services.TemplateService
}

func (s *TemplateServiceTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockDownloader = mocks.NewMockGitHubDownloader(s.ctrl)
    s.mockExtractor = mocks.NewMockArchiveExtractor(s.ctrl)
    s.mockConfig = mocks.NewMockConfigManager(s.ctrl)
    
    s.service = services.NewTemplateService(
        s.mockDownloader,
        s.mockExtractor,
        s.mockConfig,
    )
}

// TestTemplateService_DownloadTemplate_Success 测试下载模板成功
func (s *TemplateServiceTestSuite) TestTemplateService_DownloadTemplate_Success() {
    // Given
    aiType := "claude"
    expectedURL := "https://github.com/specify-ai/templates/archive/refs/heads/main.zip"
    expectedFilePath := "/tmp/template-claude.zip"
    
    s.mockConfig.EXPECT().GetTemplateURL(aiType).Return(expectedURL).Times(1)
    s.mockDownloader.EXPECT().Download(s.ctx, expectedURL).Return(expectedFilePath, nil).Times(1)
    
    // When
    filePath, err := s.service.DownloadTemplate(s.ctx, aiType)
    
    // Then
    s.AssertNoError(err)
    s.Equal(expectedFilePath, filePath)
}

// TestTemplateService_DownloadTemplate_InvalidAI 测试无效AI类型
func (s *TemplateServiceTestSuite) TestTemplateService_DownloadTemplate_InvalidAI() {
    // Given
    aiType := "invalid-ai"
    
    s.mockConfig.EXPECT().GetTemplateURL(aiType).Return("").Times(1)
    
    // When
    _, err := s.service.DownloadTemplate(s.ctx, aiType)
    
    // Then
    s.AssertError(err)
    s.Contains(err.Error(), "不支持的AI类型")
}

// TestTemplateService_DownloadTemplate_DownloadError 测试下载错误
func (s *TemplateServiceTestSuite) TestTemplateService_DownloadTemplate_DownloadError() {
    // Given
    aiType := "claude"
    expectedURL := "https://github.com/specify-ai/templates/archive/refs/heads/main.zip"
    expectedErr := errors.New("网络连接失败")
    
    s.mockConfig.EXPECT().GetTemplateURL(aiType).Return(expectedURL).Times(1)
    s.mockDownloader.EXPECT().Download(s.ctx, expectedURL).Return("", expectedErr).Times(1)
    
    // When
    _, err := s.service.DownloadTemplate(s.ctx, aiType)
    
    // Then
    s.AssertError(err)
    s.Equal(expectedErr, err)
}

// TestTemplateService_ExtractTemplate_Success 测试解压模板成功
func (s *TemplateServiceTestSuite) TestTemplateService_ExtractTemplate_Success() {
    // Given
    archivePath := "/tmp/template.zip"
    destPath := "/path/to/project"
    
    s.mockExtractor.EXPECT().Extract(s.ctx, archivePath, destPath).Return(nil).Times(1)
    
    // When
    err := s.service.ExtractTemplate(s.ctx, archivePath, destPath)
    
    // Then
    s.AssertNoError(err)
}

// TestTemplateService_ExtractTemplate_Error 测试解压错误
func (s *TemplateServiceTestSuite) TestTemplateService_ExtractTemplate_Error() {
    // Given
    archivePath := "/tmp/template.zip"
    destPath := "/path/to/project"
    expectedErr := errors.New("解压失败")
    
    s.mockExtractor.EXPECT().Extract(s.ctx, archivePath, destPath).Return(expectedErr).Times(1)
    
    // When
    err := s.service.ExtractTemplate(s.ctx, archivePath, destPath)
    
    // Then
    s.AssertError(err)
    s.Equal(expectedErr, err)
}

// TestTemplateService_GetTemplateInfo 测试获取模板信息
func (s *TemplateServiceTestSuite) TestTemplateService_GetTemplateInfo() {
    // Given
    aiType := "claude"
    expectedInfo := &models.TemplateInfo{
        Name:        "Claude Code Template",
        Description: "Claude AI助手项目模板",
        Version:     "1.0.0",
        Author:      "Specify AI",
    }
    
    s.mockConfig.EXPECT().GetTemplateInfo(aiType).Return(expectedInfo).Times(1)
    
    // When
    info := s.service.GetTemplateInfo(aiType)
    
    // Then
    s.Equal(expectedInfo, info)
}

func TestTemplateServiceTestSuite(t *testing.T) {
    suite.Run(t, new(TemplateServiceTestSuite))
}
```

---

## 5. UI组件测试用例

### 5.1 StepTracker测试用例

#### 5.1.1 StepTracker核心功能测试
```go
// internal/cli/ui/step_tracker_test.go
package ui_test

import (
    "errors"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
)

type StepTrackerTestSuite struct {
    testutil.BaseTestSuite
    tracker ui.StepTracker
}

func (s *StepTrackerTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    s.tracker = ui.NewStepTracker()
}

// TestStepTracker_SetSteps 测试设置步骤
func (s *StepTrackerTestSuite) TestStepTracker_SetSteps() {
    // Given
    steps := []*models.Step{
        {ID: "step1", Name: "步骤1", Status: models.StepStatusPending},
        {ID: "step2", Name: "步骤2", Status: models.StepStatusPending},
        {ID: "step3", Name: "步骤3", Status: models.StepStatusPending},
    }
    
    // When
    s.tracker.SetSteps(steps)
    
    // Then
    s.Equal(0.0, s.tracker.GetProgress())
    s.Nil(s.tracker.GetCurrentStep())
}

// TestStepTracker_StartStep 测试开始步骤
func (s *StepTrackerTestSuite) TestStepTracker_StartStep() {
    // Given
    steps := []*models.Step{
        {ID: "step1", Name: "步骤1", Status: models.StepStatusPending},
        {ID: "step2", Name: "步骤2", Status: models.StepStatusPending},
    }
    s.tracker.SetSteps(steps)
    
    // When
    s.tracker.StartStep("step1")
    
    // Then
    currentStep := s.tracker.GetCurrentStep()
    s.NotNil(currentStep)
    s.Equal("step1", currentStep.ID)
    s.Equal(models.StepStatusRunning, currentStep.Status)
}

// TestStepTracker_CompleteStep 测试完成步骤
func (s *StepTrackerTestSuite) TestStepTracker_CompleteStep() {
    // Given
    steps := []*models.Step{
        {ID: "step1", Name: "步骤1", Status: models.StepStatusPending},
        {ID: "step2", Name: "步骤2", Status: models.StepStatusPending},
    }
    s.tracker.SetSteps(steps)
    s.tracker.StartStep("step1")
    
    // When
    s.tracker.CompleteStep("step1")
    
    // Then
    s.Equal(50.0, s.tracker.GetProgress()) // 1/2 = 50%
}

// TestStepTracker_FailStep 测试步骤失败
func (s *StepTrackerTestSuite) TestStepTracker_FailStep() {
    // Given
    steps := []*models.Step{
        {ID: "step1", Name: "步骤1", Status: models.StepStatusPending},
        {ID: "step2", Name: "步骤2", Status: models.StepStatusPending},
    }
    s.tracker.SetSteps(steps)
    s.tracker.StartStep("step1")
    
    expectedErr := errors.New("步骤执行失败")
    
    // When
    s.tracker.FailStep("step1", expectedErr)
    
    // Then
    currentStep := s.tracker.GetCurrentStep()
    s.NotNil(currentStep)
    s.Equal(models.StepStatusFailed, currentStep.Status)
    s.Equal(expectedErr, currentStep.Error)
}

// TestStepTracker_UpdateProgress 测试更新进度
func (s *StepTrackerTestSuite) TestStepTracker_UpdateProgress() {
    // Given
    steps := []*models.Step{
        {ID: "step1", Name: "步骤1", Status: models.StepStatusPending},
    }
    s.tracker.SetSteps(steps)
    s.tracker.StartStep("step1")
    
    // When
    s.tracker.UpdateProgress("step1", 75.0)
    
    // Then
    currentStep := s.tracker.GetCurrentStep()
    s.NotNil(currentStep)
    s.Equal(75.0, currentStep.Progress)
}

// TestStepTracker_Render 测试渲染输出
func (s *StepTrackerTestSuite) TestStepTracker_Render() {
    // Given
    steps := []*models.Step{
        {ID: "step1", Name: "步骤1", Status: models.StepStatusCompleted},
        {ID: "step2", Name: "步骤2", Status: models.StepStatusRunning, Progress: 50.0},
        {ID: "step3", Name: "步骤3", Status: models.StepStatusPending},
    }
    s.tracker.SetSteps(steps)
    
    // When
    output := s.tracker.Render()
    
    // Then
    s.NotEmpty(output)
    s.Contains(output, "步骤1")
    s.Contains(output, "步骤2")
    s.Contains(output, "步骤3")
}

// TestStepTracker_ConcurrentAccess 测试并发访问
func (s *StepTrackerTestSuite) TestStepTracker_ConcurrentAccess() {
    // Given
    steps := []*models.Step{
        {ID: "step1", Name: "步骤1", Status: models.StepStatusPending},
        {ID: "step2", Name: "步骤2", Status: models.StepStatusPending},
    }
    s.tracker.SetSteps(steps)
    
    // When - 并发操作
    go func() {
        s.tracker.StartStep("step1")
        s.tracker.UpdateProgress("step1", 50.0)
        s.tracker.CompleteStep("step1")
    }()
    
    go func() {
        s.tracker.StartStep("step2")
        s.tracker.UpdateProgress("step2", 25.0)
        s.tracker.CompleteStep("step2")
    }()
    
    // Then - 等待并验证最终状态
    time.Sleep(100 * time.Millisecond)
    s.Equal(100.0, s.tracker.GetProgress())
}

func TestStepTrackerTestSuite(t *testing.T) {
    suite.Run(t, new(StepTrackerTestSuite))
}
```

### 5.2 Selector测试用例

#### 5.2.1 Selector核心功能测试
```go
// internal/cli/ui/selector_test.go
package ui_test

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/testutil"
)

type SelectorTestSuite struct {
    testutil.BaseTestSuite
    selector ui.Selector
}

func (s *SelectorTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    s.selector = ui.NewSelector()
}

// TestSelector_SingleSelect 测试单选
func (s *SelectorTestSuite) TestSelector_SingleSelect() {
    // Given
    options := []string{"选项1", "选项2", "选项3"}
    prompt := "请选择一个选项:"
    
    // 模拟用户选择第二个选项
    // 注意：实际测试中需要模拟用户输入
    
    // When
    selected, err := s.selector.SingleSelect(prompt, options)
    
    // Then
    s.AssertNoError(err)
    s.Contains(options, selected)
}

// TestSelector_MultiSelect 测试多选
func (s *SelectorTestSuite) TestSelector_MultiSelect() {
    // Given
    options := []string{"选项1", "选项2", "选项3", "选项4"}
    prompt := "请选择多个选项:"
    
    // When
    selected, err := s.selector.MultiSelect(prompt, options)
    
    // Then
    s.AssertNoError(err)
    s.NotEmpty(selected)
    for _, item := range selected {
        s.Contains(options, item)
    }
}

// TestSelector_Confirm 测试确认对话框
func (s *SelectorTestSuite) TestSelector_Confirm() {
    // Given
    message := "确定要继续吗?"
    
    // When
    confirmed, err := s.selector.Confirm(message)
    
    // Then
    s.AssertNoError(err)
    s.IsType(bool(true), confirmed)
}

func TestSelectorTestSuite(t *testing.T) {
    suite.Run(t, new(SelectorTestSuite))
}
```

### 5.3 Banner测试用例

#### 5.3.1 Banner核心功能测试
```go
// internal/cli/ui/banner_test.go
package ui_test

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/cli/ui"
    "specify-cli-go/internal/testutil"
)

type BannerTestSuite struct {
    testutil.BaseTestSuite
    banner ui.Banner
}

func (s *BannerTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    s.banner = ui.NewBanner()
}

// TestBanner_Show 测试显示横幅
func (s *BannerTestSuite) TestBanner_Show() {
    // When
    err := s.banner.Show()
    
    // Then
    s.AssertNoError(err)
}

// TestBanner_ShowWithMessage 测试带消息的横幅
func (s *BannerTestSuite) TestBanner_ShowWithMessage() {
    // Given
    message := "欢迎使用Specify CLI"
    
    // When
    err := s.banner.ShowWithMessage(message)
    
    // Then
    s.AssertNoError(err)
}

func TestBannerTestSuite(t *testing.T) {
    suite.Run(t, new(BannerTestSuite))
}
```

---

## 6. GitHub集成测试用例

### 6.1 GitHubDownloader测试用例

#### 6.1.1 GitHubDownloader核心功能测试
```go
// internal/core/downloaders/github_downloader_test.go
package downloaders_test

import (
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/core/downloaders"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type GitHubDownloaderTestSuite struct {
    testutil.BaseTestSuite
    mockHTTPClient *mocks.MockHTTPClient
    mockFileManager *mocks.MockFileManager
    downloader     downloaders.GitHubDownloader
}

func (s *GitHubDownloaderTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockHTTPClient = mocks.NewMockHTTPClient(s.ctrl)
    s.mockFileManager = mocks.NewMockFileManager(s.ctrl)
    
    s.downloader = downloaders.NewGitHubDownloader(
        s.mockHTTPClient,
        s.mockFileManager,
    )
}

// TestGitHubDownloader_Download_Success 测试下载成功
func (s *GitHubDownloaderTestSuite) TestGitHubDownloader_Download_Success() {
    // Given
    url := "https://github.com/user/repo/archive/main.zip"
    expectedFilePath := "/tmp/download.zip"
    
    // 模拟HTTP响应
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/zip")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("fake zip content"))
    }))
    defer server.Close()
    
    s.mockFileManager.EXPECT().CreateTempFile("*.zip").Return(expectedFilePath, nil).Times(1)
    s.mockFileManager.EXPECT().WriteFile(expectedFilePath, gomock.Any()).Return(nil).Times(1)
    
    // When
    filePath, err := s.downloader.Download(s.ctx, server.URL)
    
    // Then
    s.AssertNoError(err)
    s.Equal(expectedFilePath, filePath)
}

// TestGitHubDownloader_Download_HTTPError 测试HTTP错误
func (s *GitHubDownloaderTestSuite) TestGitHubDownloader_Download_HTTPError() {
    // Given
    url := "https://github.com/user/repo/archive/main.zip"
    
    // 模拟HTTP错误响应
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    }))
    defer server.Close()
    
    // When
    _, err := s.downloader.Download(s.ctx, server.URL)
    
    // Then
    s.AssertError(err)
    s.Contains(err.Error(), "下载失败")
}

// TestGitHubDownloader_Download_FileWriteError 测试文件写入错误
func (s *GitHubDownloaderTestSuite) TestGitHubDownloader_Download_FileWriteError() {
    // Given
    url := "https://github.com/user/repo/archive/main.zip"
    expectedFilePath := "/tmp/download.zip"
    expectedErr := errors.New("磁盘空间不足")
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/zip")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("fake zip content"))
    }))
    defer server.Close()
    
    s.mockFileManager.EXPECT().CreateTempFile("*.zip").Return(expectedFilePath, nil).Times(1)
    s.mockFileManager.EXPECT().WriteFile(expectedFilePath, gomock.Any()).Return(expectedErr).Times(1)
    
    // When
    _, err := s.downloader.Download(s.ctx, server.URL)
    
    // Then
    s.AssertError(err)
    s.Equal(expectedErr, err)
}

func TestGitHubDownloaderTestSuite(t *testing.T) {
    suite.Run(t, new(GitHubDownloaderTestSuite))
}
```

---

## 7. 系统集成测试用例

### 7.1 权限管理测试用例

#### 7.1.1 PermissionManager测试用例
```go
// internal/infrastructure/permission/permission_manager_test.go
package permission_test

import (
    "errors"
    "runtime"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/infrastructure/permission"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type PermissionManagerTestSuite struct {
    testutil.BaseTestSuite
    mockFileManager *mocks.MockFileManager
    mockUI          *mocks.MockUI
    manager         permission.PermissionManager
}

func (s *PermissionManagerTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockFileManager = mocks.NewMockFileManager(s.ctrl)
    s.mockUI = mocks.NewMockUI(s.ctrl)
    
    s.manager = permission.NewPermissionManager(
        s.mockFileManager,
        s.mockUI,
    )
}

// TestPermissionManager_SetExecutablePermissions_Unix 测试Unix系统设置可执行权限
func (s *PermissionManagerTestSuite) TestPermissionManager_SetExecutablePermissions_Unix() {
    if runtime.GOOS == "windows" {
        s.T().Skip("跳过Unix权限测试")
    }
    
    // Given
    scriptPath := "/path/to/scripts"
    
    s.mockFileManager.EXPECT().SetFileMode(scriptPath, os.FileMode(0755)).Return(nil).Times(1)
    
    // When
    err := s.manager.SetExecutablePermissions(scriptPath)
    
    // Then
    s.AssertNoError(err)
}

// TestPermissionManager_SetExecutablePermissions_Windows 测试Windows系统权限处理
func (s *PermissionManagerTestSuite) TestPermissionManager_SetExecutablePermissions_Windows() {
    if runtime.GOOS != "windows" {
        s.T().Skip("跳过Windows权限测试")
    }
    
    // Given
    scriptPath := "C:\\path\\to\\scripts"
    
    // Windows系统不需要设置可执行权限
    // When
    err := s.manager.SetExecutablePermissions(scriptPath)
    
    // Then
    s.AssertNoError(err)
}

// TestPermissionManager_CheckAdminPrivileges 测试检查管理员权限
func (s *PermissionManagerTestSuite) TestPermissionManager_CheckAdminPrivileges() {
    // When
    hasAdmin := s.manager.CheckAdminPrivileges()
    
    // Then
    s.IsType(bool(true), hasAdmin)
}

// TestPermissionManager_RequestElevation 测试请求权限提升
func (s *PermissionManagerTestSuite) TestPermissionManager_RequestElevation() {
    // Given
    s.mockUI.EXPECT().PromptConfirm("需要管理员权限才能继续，是否提升权限?").Return(true, nil).Times(1)
    
    // When
    granted, err := s.manager.RequestElevation()
    
    // Then
    s.AssertNoError(err)
    s.True(granted)
}

func TestPermissionManagerTestSuite(t *testing.T) {
    suite.Run(t, new(PermissionManagerTestSuite))
}
```

---

## 8. 基础设施层测试用例

### 8.1 ConfigManager测试用例

#### 8.1.1 ConfigManager核心功能测试
```go
// internal/infrastructure/config/config_manager_test.go
package config_test

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/infrastructure/config"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type ConfigManagerTestSuite struct {
    testutil.BaseTestSuite
    mockFileManager *mocks.MockFileManager
    manager         config.ConfigManager
}

func (s *ConfigManagerTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockFileManager = mocks.NewMockFileManager(s.ctrl)
    s.manager = config.NewConfigManager(s.mockFileManager)
}

// TestConfigManager_LoadConfig 测试加载配置
func (s *ConfigManagerTestSuite) TestConfigManager_LoadConfig() {
    // Given
    configData := `{
        "log_level": "info",
        "color_output": true,
        "available_ais": ["claude", "copilot"],
        "available_script_types": ["bash", "powershell"]
    }`
    
    s.mockFileManager.EXPECT().ReadFile(gomock.Any()).Return([]byte(configData), nil).Times(1)
    
    // When
    err := s.manager.LoadConfig()
    
    // Then
    s.AssertNoError(err)
    s.Equal("info", s.manager.GetLogLevel())
    s.True(s.manager.GetColorOutput())
}

// TestConfigManager_SaveConfig 测试保存配置
func (s *ConfigManagerTestSuite) TestConfigManager_SaveConfig() {
    // Given
    s.manager.SetLogLevel("debug")
    s.manager.SetColorOutput(false)
    
    s.mockFileManager.EXPECT().WriteFile(gomock.Any(), gomock.Any()).Return(nil).Times(1)
    
    // When
    err := s.manager.SaveConfig()
    
    // Then
    s.AssertNoError(err)
}

func TestConfigManagerTestSuite(t *testing.T) {
    suite.Run(t, new(ConfigManagerTestSuite))
}
```

---

## 9. 数据模型测试用例

### 9.1 Agent模型测试用例

#### 9.1.1 Agent模型测试
```go
// internal/models/agent_test.go
package models_test

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/models"
    "specify-cli-go/internal/testutil"
)

type AgentTestSuite struct {
    testutil.BaseTestSuite
}

// TestAgent_NewAgent 测试创建Agent
func (s *AgentTestSuite) TestAgent_NewAgent() {
    // Given
    name := "claude"
    description := "Claude AI助手"
    
    // When
    agent := models.NewAgent(name, description)
    
    // Then
    s.Equal(name, agent.Name)
    s.Equal(description, agent.Description)
    s.NotEmpty(agent.ID)
}

// TestAgent_Validate 测试Agent验证
func (s *AgentTestSuite) TestAgent_Validate() {
    testCases := []struct {
        name        string
        agent       *models.Agent
        expectError bool
    }{
        {
            name: "valid_agent",
            agent: &models.Agent{
                ID:          "agent-1",
                Name:        "claude",
                Description: "Claude AI助手",
            },
            expectError: false,
        },
        {
            name: "empty_name",
            agent: &models.Agent{
                ID:          "agent-1",
                Name:        "",
                Description: "描述",
            },
            expectError: true,
        },
    }
    
    for _, tc := range testCases {
        s.Run(tc.name, func() {
            // When
            err := tc.agent.Validate()
            
            // Then
            if tc.expectError {
                s.AssertError(err)
            } else {
                s.AssertNoError(err)
            }
        })
    }
}

func TestAgentTestSuite(t *testing.T) {
    suite.Run(t, new(AgentTestSuite))
}
```

---

## 10. 错误处理测试用例

### 10.1 ServiceError测试用例

#### 10.1.1 ServiceError核心功能测试
```go
// internal/errors/service_error_test.go
package errors_test

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/errors"
    "specify-cli-go/internal/testutil"
)

type ServiceErrorTestSuite struct {
    testutil.BaseTestSuite
}

// TestServiceError_NewServiceError 测试创建服务错误
func (s *ServiceErrorTestSuite) TestServiceError_NewServiceError() {
    // Given
    errorType := errors.ErrorTypeValidation
    message := "验证失败"
    
    // When
    err := errors.NewServiceError(errorType, message)
    
    // Then
    s.Equal(errorType, err.Type)
    s.Equal(message, err.Message)
    s.NotEmpty(err.Timestamp)
}

// TestServiceError_Error 测试错误字符串
func (s *ServiceErrorTestSuite) TestServiceError_Error() {
    // Given
    err := errors.NewServiceError(errors.ErrorTypeNetwork, "网络连接失败")
    
    // When
    errorStr := err.Error()
    
    // Then
    s.Contains(errorStr, "网络连接失败")
    s.Contains(errorStr, "ErrorTypeNetwork")
}

func TestServiceErrorTestSuite(t *testing.T) {
    suite.Run(t, new(ServiceErrorTestSuite))
}
```

---

## 11. 并发安全测试用例

### 11.1 并发下载测试用例

#### 11.1.1 并发下载安全测试
```go
// internal/core/downloaders/concurrent_download_test.go
package downloaders_test

import (
    "sync"
    "testing"
    "time"
    
    "github.com/stretchr/testify/suite"
    "specify-cli-go/internal/core/downloaders"
    "specify-cli-go/internal/testutil"
    "specify-cli-go/mocks"
)

type ConcurrentDownloadTestSuite struct {
    testutil.BaseTestSuite
    mockDownloader *mocks.MockGitHubDownloader
    manager        *downloaders.ConcurrentDownloadManager
}

func (s *ConcurrentDownloadTestSuite) SetupTest() {
    s.BaseTestSuite.SetupTest()
    
    s.mockDownloader = mocks.NewMockGitHubDownloader(s.ctrl)
    s.manager = downloaders.NewConcurrentDownloadManager(s.mockDownloader, 3)
}

// TestConcurrentDownload_MaxConcurrency 测试最大并发限制
func (s *ConcurrentDownloadTestSuite) TestConcurrentDownload_MaxConcurrency() {
    // Given
    urls := []string{
        "https://example.com/file1.zip",
        "https://example.com/file2.zip",
        "https://example.com/file3.zip",
        "https://example.com/file4.zip",
        "https://example.com/file5.zip",
    }
    
    var activeDownloads int32
    var maxConcurrent int32
    var mu sync.Mutex
    
    s.mockDownloader.EXPECT().Download(gomock.Any(), gomock.Any()).DoAndReturn(
        func(ctx context.Context, url string) (string, error) {
            mu.Lock()
            activeDownloads++
            if activeDownloads > maxConcurrent {
                maxConcurrent = activeDownloads
            }
            mu.Unlock()
            
            time.Sleep(100 * time.Millisecond) // 模拟下载时间
            
            mu.Lock()
            activeDownloads--
            mu.Unlock()
            
            return "/tmp/downloaded.zip", nil
        },
    ).Times(len(urls))
    
    // When
    results, err := s.manager.DownloadAll(s.ctx, urls)
    
    // Then
    s.AssertNoError(err)
    s.Len(results, len(urls))
    s.LessOrEqual(maxConcurrent, int32(3)) // 不应超过最大并发数
}

func TestConcurrentDownloadTestSuite(t *testing.T) {
    suite.Run(t, new(ConcurrentDownloadTestSuite))
}
```

---

## 12. Mock对象设计

### 12.1 Mock接口定义

```go
// mocks/mock_services.go
//go:generate mockgen -source=../internal/core/services/interfaces.go -destination=mock_services.go

// mocks/mock_ui.go  
//go:generate mockgen -source=../internal/cli/ui/interfaces.go -destination=mock_ui.go

// mocks/mock_infrastructure.go
//go:generate mockgen -source=../internal/infrastructure/interfaces.go -destination=mock_infrastructure.go
```

### 12.2 Mock生成命令

```bash
# 生成所有Mock文件
go generate ./...

# 生成特定Mock文件
mockgen -source=internal/core/services/init_service.go -destination=mocks/mock_init_service.go
mockgen -source=internal/cli/ui/step_tracker.go -destination=mocks/mock_step_tracker.go
```

---

## 13. 测试工具和框架

### 13.1 测试依赖管理

```go
// go.mod 测试依赖
require (
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    github.com/onsi/ginkgo/v2 v2.13.0
    github.com/onsi/gomega v1.29.0
    github.com/DATA-DOG/go-sqlmock v1.5.0
    github.com/jarcoal/httpmock v1.3.1
)
```

### 13.2 测试辅助工具

#### 13.2.1 测试数据生成器
```go
// internal/testutil/data_generator.go
package testutil

import (
    "specify-cli-go/internal/models"
)

// GenerateTestAgent 生成测试Agent
func GenerateTestAgent(name string) *models.Agent {
    return &models.Agent{
        ID:          "test-" + name,
        Name:        name,
        Description: "Test " + name + " agent",
    }
}

// GenerateTestSteps 生成测试步骤
func GenerateTestSteps(count int) []*models.Step {
    steps := make([]*models.Step, count)
    for i := 0; i < count; i++ {
        steps[i] = &models.Step{
            ID:     fmt.Sprintf("step-%d", i+1),
            Name:   fmt.Sprintf("步骤%d", i+1),
            Status: models.StepStatusPending,
        }
    }
    return steps
}
```

---

## 14. 测试覆盖率要求

### 14.1 覆盖率目标

| 层级 | 最低覆盖率 | 目标覆盖率 |
|------|------------|------------|
| CLI层 | 80% | 90% |
| 服务层 | 85% | 95% |
| UI组件 | 70% | 85% |
| 基础设施层 | 80% | 90% |
| 数据模型 | 90% | 95% |
| 整体项目 | 80% | 90% |

### 14.2 覆盖率检查命令

```bash
# 运行测试并生成覆盖率报告
go test -v -race -coverprofile=coverage.out ./...

# 查看覆盖率统计
go tool cover -func=coverage.out

# 生成HTML覆盖率报告
go tool cover -html=coverage.out -o coverage.html

# 检查覆盖率是否达标
go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//' | awk '{if($1<80) exit 1}'
```

---

## 15. 测试执行策略

### 15.1 测试分类

#### 15.1.1 单元测试
- **执行频率**: 每次代码提交
- **执行时间**: < 30秒
- **并行执行**: 是
- **Mock依赖**: 是

#### 15.1.2 集成测试
- **执行频率**: 每次Pull Request
- **执行时间**: < 5分钟
- **并行执行**: 部分
- **Mock依赖**: 部分

#### 15.1.3 端到端测试
- **执行频率**: 每日构建
- **执行时间**: < 15分钟
- **并行执行**: 否
- **Mock依赖**: 否

### 15.2 CI/CD集成

#### 15.2.1 GitHub Actions配置
```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run Unit Tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
    
    - name: Check Coverage
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$COVERAGE < 80" | bc -l) )); then
          echo "Coverage $COVERAGE% is below 80%"
          exit 1
        fi
    
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### 15.3 测试执行命令

#### 15.3.1 基本测试命令
```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./internal/cli/commands/...

# 运行特定测试用例
go test -v -run TestInitCommand_Run_Success ./internal/cli/commands/

# 运行基准测试
go test -bench=. ./...

# 运行竞态检测
go test -race ./...
```

#### 15.3.2 Makefile配置
```makefile
# Makefile
.PHONY: test test-unit test-integration test-coverage test-race

test: test-unit test-integration

test-unit:
	go test -v -short ./...

test-integration:
	go test -v -tags=integration ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-race:
	go test -v -race ./...

test-clean:
	go clean -testcache

mock-generate:
	go generate ./...
```

---

## 总结

本单元测试用例文档为Specify CLI Go版本提供了全面的测试规范，涵盖了：

1. **完整的测试架构**: 从CLI层到基础设施层的全栈测试覆盖
2. **详细的测试用例**: 包含正常流程、异常处理、边界条件等场景
3. **并发安全测试**: 确保多线程环境下的数据一致性
4. **Mock对象设计**: 提供完整的依赖隔离方案
5. **测试工具集成**: 包含测试框架、覆盖率工具、CI/CD集成
6. **质量保证机制**: 明确的覆盖率要求和执行策略

通过执行这些测试用例，可以确保Go版本的实现质量，验证所有功能的正确性，并为后续的维护和扩展提供可靠的质量保障。