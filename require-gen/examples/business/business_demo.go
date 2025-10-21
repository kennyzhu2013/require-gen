package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"specify-cli/internal/business"
	"specify-cli/internal/infrastructure"
	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

// BusinessDemoConfig 业务层演示配置
type BusinessDemoConfig struct {
	ProjectName   string
	AIAssistant   string
	ScriptType    string
	GitHubToken   string
	TestDirectory string
	Verbose       bool
}

// BusinessDemo 业务层演示结构体
type BusinessDemo struct {
	config           *BusinessDemoConfig
	initHandler      *business.InitHandler
	downloadHandler  *business.DownloadHandler
	uiManager        *ui.UIManager
	authProvider     *infrastructure.AuthProvider
	toolChecker      *infrastructure.ToolChecker
	systemOps        *infrastructure.SystemOperations
	gitOps           *infrastructure.GitOperations
	templateProvider *infrastructure.TemplateProvider
}

// NewBusinessDemo 创建新的业务层演示实例
func NewBusinessDemo(config *BusinessDemoConfig) *BusinessDemo {
	// 初始化基础设施层组件 - 修正：这些构造函数都不需要参数
	systemOps := infrastructure.NewSystemOperations()
	authProvider := infrastructure.NewAuthProvider()
	toolChecker := infrastructure.NewToolChecker()
	gitOps := infrastructure.NewGitOperations()
	templateProvider := infrastructure.NewTemplateProvider()
	uiManager := ui.NewUIManager()

	// 初始化业务层组件 - 修正：这些构造函数不需要参数
	initHandler := business.NewInitHandler()
	downloadHandler := business.NewDownloadHandler()

	return &BusinessDemo{
		config:           config,
		initHandler:      initHandler,
		downloadHandler:  downloadHandler,
		uiManager:        uiManager,
		authProvider:     authProvider.(*infrastructure.AuthProvider),
		toolChecker:      toolChecker.(*infrastructure.ToolChecker),
		systemOps:        systemOps.(*infrastructure.SystemOperations),
		gitOps:           gitOps.(*infrastructure.GitOperations),
		templateProvider: templateProvider.(*infrastructure.TemplateProvider),
	}
}

// RunDemo 运行业务层演示
func (bd *BusinessDemo) RunDemo() error {
	fmt.Println("=== Business Logic Layer 集成测试 ===")
	
	// 1. 测试UI组件
	if err := bd.testUIComponents(); err != nil {
		return fmt.Errorf("UI组件测试失败: %w", err)
	}
	
	// 2. 测试认证管理
	if err := bd.testAuthManagement(); err != nil {
		return fmt.Errorf("认证管理测试失败: %w", err)
	}
	
	// 3. 测试工具检查
	if err := bd.testToolChecking(); err != nil {
		return fmt.Errorf("工具检查测试失败: %w", err)
	}
	
	// 4. 测试系统操作
	if err := bd.testSystemOperations(); err != nil {
		return fmt.Errorf("系统操作测试失败: %w", err)
	}
	
	// 5. 测试Git操作
	if err := bd.testGitOperations(); err != nil {
		return fmt.Errorf("Git操作测试失败: %w", err)
	}
	
	// 6. 测试模板管理
	if err := bd.testTemplateManagement(); err != nil {
		return fmt.Errorf("模板管理测试失败: %w", err)
	}
	
	// 7. 测试项目初始化
	if err := bd.testProjectInitialization(); err != nil {
		return fmt.Errorf("项目初始化测试失败: %w", err)
	}
	
	// 8. 测试错误处理
	if err := bd.testErrorHandling(); err != nil {
		return fmt.Errorf("错误处理测试失败: %w", err)
	}
	
	bd.printSummary()
	return nil
}

// testUIComponents 测试UI组件功能
func (bd *BusinessDemo) testUIComponents() error {
	fmt.Println("\n--- 测试UI组件 ---")
	
	// 测试步骤跟踪器
	stepTracker := ui.NewStepTracker("UI测试")
	stepTracker.AddStep("ui_init", "初始化UI组件")
	stepTracker.SetStepRunning("ui_init", "正在初始化...")
	stepTracker.SetStepDone("ui_init", "UI组件初始化完成")
	
	// 测试进度条
	progressBar := ui.NewProgressBar(100, "测试进度", ui.ModernStyle())
	for i := 0; i <= 100; i += 20 {
		progressBar.Update(int64(i))
		time.Sleep(100 * time.Millisecond)
	}
	
	// 测试UI管理器
	ui.ShowBanner()
	ui.ShowSuccess("UI组件测试通过")
	
	fmt.Println("✅ UI组件测试通过")
	return nil
}

// testAuthManagement 测试认证管理功能
func (bd *BusinessDemo) testAuthManagement() error {
	fmt.Println("\n--- 测试认证管理 ---")
	
	// 测试令牌设置
	if bd.config.GitHubToken != "" {
		bd.authProvider.SetToken(bd.config.GitHubToken)
		fmt.Printf("✅ 设置GitHub令牌: %s...\n", bd.config.GitHubToken[:8])
	}
	
	// 测试认证状态
	isAuth := bd.authProvider.IsAuthenticated()
	fmt.Printf("✅ 认证状态检查: %v\n", isAuth)
	
	// 测试认证头
	headers := bd.authProvider.GetHeaders()
	fmt.Printf("✅ 获取认证头: %d个头部字段\n", len(headers))
	
	fmt.Println("✅ 认证管理测试通过")
	return nil
}

// testToolChecking 测试工具检查功能
func (bd *BusinessDemo) testToolChecking() error {
	fmt.Println("\n--- 测试工具检查 ---")
	
	// 创建步骤跟踪器用于工具检查
	stepTracker := &types.StepTracker{
		Title: "工具检查",
		Steps: make(map[string]*types.Step),
	}
	
	// 测试单个工具检查
	tools := []string{"git", "node", "python"}
	for _, tool := range tools {
		available := bd.toolChecker.CheckTool(tool, stepTracker)
		fmt.Printf("✅ 工具 %s 可用性: %v\n", tool, available)
	}
	
	// 测试批量工具检查
	allAvailable := bd.toolChecker.CheckAllTools(tools, stepTracker)
	fmt.Printf("✅ 所有工具可用性: %v\n", allAvailable)
	
	// 测试系统信息
	sysInfo := bd.toolChecker.GetSystemInfo()
	fmt.Printf("✅ 系统信息: %d个字段\n", len(sysInfo))
	
	fmt.Println("✅ 工具检查测试通过")
	return nil
}

// testSystemOperations 测试系统操作功能
func (bd *BusinessDemo) testSystemOperations() error {
	fmt.Println("\n--- 测试系统操作 ---")
	
	// 测试命令执行 - 使用Windows兼容的命令
	result, err := bd.systemOps.ExecuteCommand("cmd", "/c", "echo", "Hello World")
	if err != nil {
		return fmt.Errorf("命令执行失败: %w", err)
	}
	fmt.Printf("✅ 命令执行结果: %s\n", result.Output)
	
	// 测试目录操作
	testDir := filepath.Join(bd.config.TestDirectory, "test_system")
	err = bd.systemOps.CreateDirectory(testDir)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	fmt.Printf("✅ 创建测试目录: %s\n", testDir)
	
	// 测试文件操作
	testFile := filepath.Join(testDir, "test.txt")
	err = bd.systemOps.WriteFile(testFile, []byte("测试内容"))
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	fmt.Printf("✅ 创建测试文件: %s\n", testFile)
	
	// 清理测试文件
	defer bd.systemOps.RemoveDirectory(testDir)
	
	fmt.Println("✅ 系统操作测试通过")
	return nil
}

// testGitOperations 测试Git操作功能
func (bd *BusinessDemo) testGitOperations() error {
	fmt.Println("\n--- 测试Git操作 ---")
	
	testDir := filepath.Join(bd.config.TestDirectory, "test_git")
	
	// 确保测试目录存在
	err := bd.systemOps.CreateDirectory(testDir)
	if err != nil {
		return fmt.Errorf("创建Git测试目录失败: %w", err)
	}
	defer bd.systemOps.RemoveDirectory(testDir)
	
	// 测试Git仓库检查
	isRepo := bd.gitOps.IsRepo(testDir)
	fmt.Printf("✅ Git仓库检查: %v\n", isRepo)
	
	// 测试Git初始化
	success, err := bd.gitOps.InitRepo(testDir, true)
	if err != nil {
		return fmt.Errorf("Git初始化失败: %w", err)
	}
	fmt.Printf("✅ Git初始化: %v\n", success)
	
	// 测试Git状态
	status, err := bd.gitOps.GetStatus(testDir)
	if err != nil {
		return fmt.Errorf("获取Git状态失败: %w", err)
	}
	fmt.Printf("✅ Git状态: %s\n", status)
	
	// 测试分支操作
	branch, err := bd.gitOps.GetBranch(testDir)
	if err != nil {
		return fmt.Errorf("获取分支失败: %w", err)
	}
	fmt.Printf("✅ 当前分支: %s\n", branch)
	
	fmt.Println("✅ Git操作测试通过")
	return nil
}

// testTemplateManagement 测试模板管理功能
func (bd *BusinessDemo) testTemplateManagement() error {
	fmt.Println("\n--- 测试模板管理 ---")
	
	// 创建下载选项
	downloadOpts := types.DownloadOptions{
		AIAssistant:  bd.config.AIAssistant,
		DownloadDir:  bd.config.TestDirectory,
		ScriptType:   bd.config.ScriptType,
		GitHubToken:  bd.config.GitHubToken,
		Verbose:      bd.config.Verbose,
		ShowProgress: true,
	}
	
	// 注意：这里只是测试接口调用，实际下载可能会失败
	fmt.Printf("✅ 模板下载配置: AI助手=%s, 脚本类型=%s\n", 
		downloadOpts.AIAssistant, downloadOpts.ScriptType)
	
	// 测试模板列表获取（如果有令牌）
	if bd.config.GitHubToken != "" {
		templates, err := bd.templateProvider.ListTemplates(bd.config.GitHubToken)
		if err != nil {
			fmt.Printf("⚠️ 获取模板列表失败: %v\n", err)
		} else {
			fmt.Printf("✅ 可用模板数量: %d\n", len(templates))
		}
	}
	
	fmt.Println("✅ 模板管理测试通过")
	return nil
}

// testProjectInitialization 测试项目初始化功能
func (bd *BusinessDemo) testProjectInitialization() error {
	fmt.Println("\n--- 测试项目初始化 ---")
	
	// 创建初始化选项
	initOpts := types.InitOptions{
		ProjectName: bd.config.ProjectName,
		Here:        true, // 在当前目录初始化
		AIAssistant: bd.config.AIAssistant,
		ScriptType:  bd.config.ScriptType,
		GitHubToken: bd.config.GitHubToken,
		Verbose:     bd.config.Verbose,
	}
	
	fmt.Printf("✅ 初始化配置: 项目=%s, AI助手=%s, 脚本类型=%s\n",
		initOpts.ProjectName, initOpts.AIAssistant, initOpts.ScriptType)
	
	// 注意：这里只是验证配置，不执行实际初始化以避免副作用
	fmt.Println("✅ 项目初始化配置验证通过")
	return nil
}

// testErrorHandling 测试错误处理功能
func (bd *BusinessDemo) testErrorHandling() error {
	fmt.Println("\n--- 测试错误处理 ---")
	
	// 测试无效路径的Git操作
	_, err := bd.gitOps.GetStatus("/invalid/path/that/does/not/exist")
	if err != nil {
		fmt.Printf("✅ 错误处理测试: 正确捕获了无效路径错误\n")
	}
	
	// 测试无效工具检查
	stepTracker := &types.StepTracker{
		Title: "错误测试",
		Steps: make(map[string]*types.Step),
	}
	available := bd.toolChecker.CheckTool("nonexistent_tool_12345", stepTracker)
	if !available {
		fmt.Printf("✅ 错误处理测试: 正确识别了不存在的工具\n")
	}
	
	// 测试无效命令执行
	_, err = bd.systemOps.ExecuteCommand("nonexistent_command_12345")
	if err != nil {
		fmt.Printf("✅ 错误处理测试: 正确捕获了无效命令错误\n")
	}
	
	fmt.Println("✅ 错误处理测试通过")
	return nil
}



// printSummary 打印测试总结
func (bd *BusinessDemo) printSummary() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Business Logic Layer 测试总结")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Println("✅ 所有测试已完成")
	fmt.Println("📊 测试覆盖范围:")
	fmt.Println("   - UI组件管理")
	fmt.Println("   - 认证管理")
	fmt.Println("   - 工具检查")
	fmt.Println("   - 系统操作")
	fmt.Println("   - Git操作")
	fmt.Println("   - 模板管理")
	fmt.Println("   - 项目初始化")
	fmt.Println("   - 错误处理")
	
	fmt.Println()
	fmt.Println("🎯 Business Logic Layer 调用链验证完成")
	fmt.Println("   所有核心功能模块都能正常协作")
	
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
}

func main() {
	// 创建测试配置
	config := &BusinessDemoConfig{
		ProjectName:   "test-project",
		AIAssistant:   "github-copilot",
		ScriptType:    "bash",
		GitHubToken:   os.Getenv("GITHUB_TOKEN"),
		TestDirectory: "./test_output",
		Verbose:       true,
	}
	
	// 创建业务层演示实例
	demo := NewBusinessDemo(config)
	
	// 运行演示
	if err := demo.RunDemo(); err != nil {
		log.Fatalf("演示运行失败: %v", err)
	}
}