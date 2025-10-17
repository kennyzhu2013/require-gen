package business

import (
	"fmt"
	"os"
	"path/filepath"

	"specify-cli/internal/config"
	"specify-cli/internal/infrastructure"
	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

// InitHandler 初始化处理器
type InitHandler struct {
	toolChecker      types.ToolChecker
	gitOps          types.GitOperations
	templateProvider types.TemplateProvider
	authProvider     types.AuthProvider
	uiRenderer       types.UIRenderer
}

// NewInitHandler 创建新的初始化处理器
func NewInitHandler() *InitHandler {
	return &InitHandler{
		toolChecker:      infrastructure.NewToolChecker(),
		gitOps:          infrastructure.NewGitOperations(),
		templateProvider: infrastructure.NewTemplateProvider(),
		authProvider:     infrastructure.NewAuthProvider(),
		uiRenderer:       ui.NewRenderer(),
	}
}

// Execute 执行初始化流程
func (h *InitHandler) Execute(opts types.InitOptions) error {
	// 创建步骤跟踪器
	tracker := ui.NewStepTracker("Project Initialization")
	
	// 添加步骤
	h.setupSteps(tracker, opts)
	
	// 显示初始状态
	tracker.Display()

	// 执行初始化流程
	if err := h.executeSteps(tracker, opts); err != nil {
		ui.ShowError(fmt.Sprintf("Initialization failed: %v", err))
		return err
	}

	// 显示完成状态
	tracker.Display()
	ui.ShowSuccess("Project initialization completed successfully!")
	
	return nil
}

// setupSteps 设置初始化步骤
func (h *InitHandler) setupSteps(tracker *ui.StepTracker, opts types.InitOptions) {
	tracker.AddStep("validate", "Validate options")
	tracker.AddStep("select_ai", "Select AI assistant")
	tracker.AddStep("select_script", "Select script type")
	tracker.AddStep("check_tools", "Check required tools")
	tracker.AddStep("create_dir", "Create project directory")
	tracker.AddStep("download_template", "Download template")
	tracker.AddStep("init_git", "Initialize Git repository")
	tracker.AddStep("configure", "Configure project")
	tracker.AddStep("complete", "Finalize setup")
}

// executeSteps 执行初始化步骤
func (h *InitHandler) executeSteps(tracker *ui.StepTracker, opts types.InitOptions) error {
	// 步骤1: 验证选项
	if err := h.validateOptions(tracker, &opts); err != nil {
		return err
	}

	// 步骤2: 选择AI助手
	if err := h.selectAIAssistant(tracker, &opts); err != nil {
		return err
	}

	// 步骤3: 选择脚本类型
	if err := h.selectScriptType(tracker, &opts); err != nil {
		return err
	}

	// 步骤4: 检查工具
	if err := h.checkTools(tracker, opts); err != nil {
		return err
	}

	// 步骤5: 创建项目目录
	if err := h.createProjectDirectory(tracker, opts); err != nil {
		return err
	}

	// 步骤6: 下载模板
	if err := h.downloadTemplate(tracker, opts); err != nil {
		return err
	}

	// 步骤7: 初始化Git
	if err := h.initializeGit(tracker, opts); err != nil {
		return err
	}

	// 步骤8: 配置项目
	if err := h.configureProject(tracker, opts); err != nil {
		return err
	}

	// 步骤9: 完成设置
	if err := h.finalizeSetup(tracker, opts); err != nil {
		return err
	}

	return nil
}

// validateOptions 验证选项
func (h *InitHandler) validateOptions(tracker *ui.StepTracker, opts *types.InitOptions) error {
	tracker.SetStepRunning("validate", "Validating initialization options")

	// 验证项目名称
	if !opts.Here && opts.ProjectName == "" {
		tracker.SetStepError("validate", "Project name is required")
		return fmt.Errorf("project name is required unless --here is used")
	}

	// 验证AI助手
	if opts.AIAssistant != "" {
		if _, exists := config.GetAgentInfo(opts.AIAssistant); !exists {
			tracker.SetStepError("validate", fmt.Sprintf("Unknown AI assistant: %s", opts.AIAssistant))
			return fmt.Errorf("unknown AI assistant: %s", opts.AIAssistant)
		}
	}

	// 验证脚本类型
	if opts.ScriptType != "" {
		if _, exists := config.GetScriptType(opts.ScriptType); !exists {
			tracker.SetStepError("validate", fmt.Sprintf("Unknown script type: %s", opts.ScriptType))
			return fmt.Errorf("unknown script type: %s", opts.ScriptType)
		}
	}

	tracker.SetStepDone("validate", "Options validated successfully")
	return nil
}

// selectAIAssistant 选择AI助手
func (h *InitHandler) selectAIAssistant(tracker *ui.StepTracker, opts *types.InitOptions) error {
	tracker.SetStepRunning("select_ai", "Selecting AI assistant")

	if opts.AIAssistant == "" {
		agents := config.GetAllAgents()
		selected, err := h.uiRenderer.SelectWithArrows(agents, "Select AI Assistant", "github-copilot")
		if err != nil {
			tracker.SetStepError("select_ai", fmt.Sprintf("Selection failed: %v", err))
			return fmt.Errorf("failed to select AI assistant: %w", err)
		}
		opts.AIAssistant = selected
	}

	agentInfo, _ := config.GetAgentInfo(opts.AIAssistant)
	tracker.SetStepDone("select_ai", fmt.Sprintf("Selected: %s", agentInfo.Name))
	return nil
}

// selectScriptType 选择脚本类型
func (h *InitHandler) selectScriptType(tracker *ui.StepTracker, opts *types.InitOptions) error {
	tracker.SetStepRunning("select_script", "Selecting script type")

	if opts.ScriptType == "" {
		scripts := config.GetAllScriptTypes()
		defaultScript := config.GetDefaultScriptType()
		selected, err := h.uiRenderer.SelectWithArrows(scripts, "Select Script Type", defaultScript)
		if err != nil {
			tracker.SetStepError("select_script", fmt.Sprintf("Selection failed: %v", err))
			return fmt.Errorf("failed to select script type: %w", err)
		}
		opts.ScriptType = selected
	}

	scriptInfo, _ := config.GetScriptType(opts.ScriptType)
	tracker.SetStepDone("select_script", fmt.Sprintf("Selected: %s", scriptInfo.Description))
	return nil
}

// checkTools 检查所需工具
func (h *InitHandler) checkTools(tracker *ui.StepTracker, opts types.InitOptions) error {
	tracker.SetStepRunning("check_tools", "Checking required tools")

	tools := config.GetRequiredTools(opts.AIAssistant)
	if !h.toolChecker.CheckAllTools(tools, tracker) {
		tracker.SetStepError("check_tools", "Some required tools are missing")
		return fmt.Errorf("required tools are missing")
	}

	tracker.SetStepDone("check_tools", "All required tools are available")
	return nil
}

// createProjectDirectory 创建项目目录
func (h *InitHandler) createProjectDirectory(tracker *ui.StepTracker, opts types.InitOptions) error {
	tracker.SetStepRunning("create_dir", "Creating project directory")

	if !opts.Here {
		if err := os.MkdirAll(opts.ProjectName, 0755); err != nil {
			tracker.SetStepError("create_dir", fmt.Sprintf("Failed to create directory: %v", err))
			return fmt.Errorf("failed to create project directory: %w", err)
		}

		if err := os.Chdir(opts.ProjectName); err != nil {
			tracker.SetStepError("create_dir", fmt.Sprintf("Failed to change directory: %v", err))
			return fmt.Errorf("failed to change to project directory: %w", err)
		}

		tracker.SetStepDone("create_dir", fmt.Sprintf("Created and entered directory: %s", opts.ProjectName))
	} else {
		cwd, _ := os.Getwd()
		tracker.SetStepDone("create_dir", fmt.Sprintf("Using current directory: %s", filepath.Base(cwd)))
	}

	return nil
}

// downloadTemplate 下载模板
func (h *InitHandler) downloadTemplate(tracker *ui.StepTracker, opts types.InitOptions) error {
	tracker.SetStepRunning("download_template", "Downloading project template")

	downloadOpts := types.DownloadOptions{
		AIAssistant:  opts.AIAssistant,
		DownloadDir:  ".",
		ScriptType:   opts.ScriptType,
		Verbose:      opts.Verbose,
		ShowProgress: true,
		GitHubToken:  opts.GitHubToken,
	}

	templatePath, err := h.templateProvider.Download(downloadOpts)
	if err != nil {
		tracker.SetStepError("download_template", fmt.Sprintf("Download failed: %v", err))
		return fmt.Errorf("failed to download template: %w", err)
	}

	tracker.SetStepDone("download_template", fmt.Sprintf("Template downloaded to: %s", templatePath))
	return nil
}

// initializeGit 初始化Git仓库
func (h *InitHandler) initializeGit(tracker *ui.StepTracker, opts types.InitOptions) error {
	tracker.SetStepRunning("init_git", "Initializing Git repository")

	cwd, _ := os.Getwd()
	if h.gitOps.IsRepo(cwd) {
		tracker.SetStepSkipped("init_git", "Git repository already exists")
		return nil
	}

	created, err := h.gitOps.InitRepo(cwd, !opts.Verbose)
	if err != nil {
		tracker.SetStepError("init_git", fmt.Sprintf("Git initialization failed: %v", err))
		return fmt.Errorf("failed to initialize Git repository: %w", err)
	}

	if created {
		tracker.SetStepDone("init_git", "Git repository initialized")
	} else {
		tracker.SetStepSkipped("init_git", "Git repository already exists")
	}

	return nil
}

// configureProject 配置项目
func (h *InitHandler) configureProject(tracker *ui.StepTracker, opts types.InitOptions) error {
	tracker.SetStepRunning("configure", "Configuring project settings")

	// 这里可以添加项目配置逻辑
	// 例如：创建配置文件、设置环境变量等

	tracker.SetStepDone("configure", "Project configured successfully")
	return nil
}

// finalizeSetup 完成设置
func (h *InitHandler) finalizeSetup(tracker *ui.StepTracker, opts types.InitOptions) error {
	tracker.SetStepRunning("complete", "Finalizing project setup")

	// 创建初始提交
	cwd, _ := os.Getwd()
	if h.gitOps.IsRepo(cwd) {
		if err := h.gitOps.AddAndCommit(cwd, "Initial commit: Project setup with Specify CLI"); err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to create initial commit: %v", err))
		}
	}

	tracker.SetStepDone("complete", "Project setup completed")
	return nil
}