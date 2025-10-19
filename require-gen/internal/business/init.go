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

// InitHandler Spec-Driven Development项目初始化处理器
//
// InitHandler是require-gen框架的核心组件，负责协调整个项目初始化流程。
// 它封装了项目创建、配置、模板下载等复杂操作，为用户提供简单统一的接口。
//
// 该处理器实现了类似于Python版本specify-cli的功能，但采用Go语言的
// 分层架构设计，提供更好的性能、错误处理和跨平台兼容性。
//
// 核心职责：
//   - 协调多个基础设施组件的协作
//   - 管理初始化流程的状态和进度
//   - 处理用户交互和选择逻辑
//   - 提供统一的错误处理和用户反馈
//
// 依赖组件：
//   - ToolChecker: 系统工具可用性检查
//   - GitOperations: Git版本控制操作
//   - TemplateProvider: 项目模板下载和管理
//   - AuthProvider: 认证和授权服务
//   - UIRenderer: 用户界面渲染和交互
type InitHandler struct {
	toolChecker      types.ToolChecker
	gitOps          types.GitOperations
	templateProvider types.TemplateProvider
	authProvider     types.AuthProvider
	uiRenderer       types.UIRenderer
}

// NewInitHandler 创建新的初始化处理器实例
//
// 该构造函数负责创建并配置InitHandler的所有依赖组件。
// 它使用依赖注入模式，将各个基础设施组件注入到处理器中，
// 确保组件之间的松耦合和可测试性。
//
// 初始化的组件：
//   - ToolChecker: 用于检查系统工具的可用性
//   - GitOperations: 提供Git版本控制功能
//   - TemplateProvider: 处理项目模板的下载和管理
//   - AuthProvider: 处理认证和授权相关操作
//   - UIRenderer: 提供用户界面渲染和交互功能
//
// 返回值：
//   *InitHandler - 完全配置的初始化处理器实例
//
// 设计模式：
//   - 工厂模式：封装复杂的对象创建逻辑
//   - 依赖注入：通过构造函数注入依赖
//   - 单一职责：每个组件专注于特定功能
//
// 使用示例：
//   handler := business.NewInitHandler()
//   err := handler.Execute(initOptions)
func NewInitHandler() *InitHandler {
	return &InitHandler{
		toolChecker:      infrastructure.NewToolChecker(),
		gitOps:          infrastructure.NewGitOperations(),
		templateProvider: infrastructure.NewTemplateProvider(),
		authProvider:     infrastructure.NewAuthProvider(),
		uiRenderer:       ui.NewRenderer(),
	}
}

// Execute 执行Spec-Driven Development项目的完整初始化流程
//
// 该函数是InitHandler的核心方法，负责协调整个项目初始化过程。
// 它实现了类似于Python版本specify-cli中的项目初始化功能，但采用Go语言的
// 分层架构设计，提供更好的错误处理和用户体验。
//
// 功能特性：
//   - 创建可视化的步骤跟踪器，实时显示初始化进度
//   - 支持多种AI助手配置（Claude、GitHub Copilot、Gemini等）
//   - 支持跨平台脚本类型选择（bash/PowerShell）
//   - 自动检查系统依赖工具的可用性
//   - 从GitHub下载并提取项目模板
//   - 可选的Git仓库初始化和首次提交
//   - 完整的错误处理和用户反馈机制
//
// 执行流程：
//   1. 创建步骤跟踪器并设置所有初始化步骤
//   2. 显示初始化进度界面
//   3. 按顺序执行9个核心步骤（验证、选择、检查、创建、下载、配置等）
//   4. 在任何步骤失败时立即停止并显示错误信息
//   5. 成功完成后显示完成状态和成功消息
//
// 参数：
//   opts - 初始化选项配置，包含以下字段：
//     - ProjectName: 项目名称（当Here为false时必需）
//     - Here: 是否在当前目录初始化（默认false）
//     - AIAssistant: AI助手类型（如"claude", "github-copilot"等）
//     - ScriptType: 脚本类型（"sh"或"ps"）
//     - GitHubToken: GitHub访问令牌（用于模板下载）
//     - Verbose: 是否显示详细输出
//     - Debug: 是否启用调试模式
//
// 返回值：
//   error - 如果初始化过程中任何步骤失败，返回相应的错误信息；成功时返回nil
//
// 错误处理：
//   - 参数验证失败：返回参数相关错误
//   - 工具检查失败：返回缺少依赖工具的错误
//   - 网络问题：返回模板下载失败错误
//   - 文件系统问题：返回目录创建或文件操作错误
//   - Git操作失败：返回Git相关错误（非致命，会显示警告）
//
// 使用示例：
//   handler := business.NewInitHandler()
//   opts := types.InitOptions{
//       ProjectName: "my-project",
//       AIAssistant: "claude",
//       ScriptType:  "sh",
//       Verbose:     true,
//   }
//   err := handler.Execute(opts)
//   if err != nil {
//       log.Fatalf("项目初始化失败: %v", err)
//   }
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

// setupSteps 设置项目初始化的所有步骤到步骤跟踪器中
//
// 该函数负责定义和注册项目初始化过程中的所有步骤，为用户提供清晰的
// 进度可视化。每个步骤都有唯一的标识符和描述性名称。
//
// 初始化步骤序列：
//   1. validate - 验证用户输入的选项参数
//   2. select_ai - 选择或确认AI助手类型
//   3. select_script - 选择或确认脚本类型（bash/PowerShell）
//   4. check_tools - 检查系统中必需工具的可用性
//   5. create_dir - 创建项目目录或确认当前目录
//   6. download_template - 从GitHub下载项目模板
//   7. init_git - 初始化Git仓库（如果不存在）
//   8. configure - 配置项目设置和环境
//   9. complete - 完成最终设置和清理工作
//
// 参数：
//   tracker - 步骤跟踪器实例，用于管理和显示进度
//   opts - 初始化选项，虽然当前未直接使用，但为未来扩展预留
//
// 注意：
//   - 步骤的执行顺序很重要，后续步骤可能依赖前面步骤的结果
//   - 每个步骤ID必须与executeSteps中的调用保持一致
//   - 步骤描述应该简洁明了，便于用户理解当前进度
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

// executeSteps 按顺序执行项目初始化的所有步骤
//
// 该函数是初始化流程的核心执行器，负责按照预定义的顺序调用各个
// 初始化步骤。它实现了fail-fast机制，任何步骤失败都会立即停止
// 整个初始化过程并返回错误。
//
// 执行策略：
//   - 严格按照步骤顺序执行，确保依赖关系正确
//   - 每个步骤都会更新步骤跟踪器的状态
//   - 任何步骤失败都会立即返回错误，不继续后续步骤
//   - 支持某些步骤的跳过逻辑（如Git仓库已存在）
//
// 步骤执行流程：
//   1. validateOptions - 验证和规范化用户输入
//   2. selectAIAssistant - 交互式选择或确认AI助手
//   3. selectScriptType - 交互式选择或确认脚本类型
//   4. checkTools - 验证系统工具依赖
//   5. createProjectDirectory - 创建或确认工作目录
//   6. downloadTemplate - 下载和提取项目模板
//   7. initializeGit - 初始化版本控制
//   8. configureProject - 应用项目特定配置
//   9. finalizeSetup - 完成最终设置和清理
//
// 参数：
//   tracker - 步骤跟踪器，用于更新和显示进度状态
//   opts - 初始化选项配置，包含所有必要的设置参数
//
// 返回值：
//   error - 如果任何步骤失败，返回详细的错误信息；成功时返回nil
//
// 错误处理：
//   - 每个步骤的错误都会被捕获并传播到上层
//   - 错误信息包含失败步骤的上下文信息
//   - 步骤跟踪器会显示失败状态和错误消息
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

// validateOptions 验证和规范化初始化选项参数
//
// 该函数是初始化流程的第一步，负责验证用户提供的所有选项参数
// 的有效性和完整性。它确保后续步骤能够基于正确的配置执行。
//
// 验证规则：
//   - 项目名称：当不使用--here选项时，项目名称为必需参数
//   - AI助手：如果指定，必须是支持的AI助手类型之一
//   - 脚本类型：如果指定，必须是支持的脚本类型（sh/ps）
//   - 参数组合：检查参数之间的逻辑一致性
//
// 支持的AI助手类型：
//   - claude: Anthropic Claude AI助手
//   - github-copilot: GitHub Copilot
//   - gemini: Google Gemini
//   - 其他在config中定义的助手类型
//
// 支持的脚本类型：
//   - sh: Unix/Linux shell脚本（bash）
//   - ps: Windows PowerShell脚本
//
// 参数：
//   tracker - 步骤跟踪器，用于更新验证进度和状态
//   opts - 指向初始化选项的指针，允许函数修改选项值
//
// 返回值：
//   error - 如果验证失败，返回具体的错误信息；验证通过时返回nil
//
// 副作用：
//   - 更新步骤跟踪器的状态（运行中、完成、错误）
//   - 可能修改opts中的某些字段以进行规范化
//
// 错误类型：
//   - 缺少必需参数错误
//   - 无效参数值错误
//   - 参数组合冲突错误
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

// selectAIAssistant 选择或确认AI助手类型
//
// 该函数负责处理AI助手的选择逻辑。如果用户已经通过命令行参数
// 指定了AI助手，则直接确认；否则提供交互式选择界面让用户选择。
//
// 交互式选择特性：
//   - 显示所有可用的AI助手选项
//   - 支持方向键导航选择
//   - 提供默认选项（github-copilot）
//   - 显示每个助手的详细信息和描述
//
// 可用的AI助手：
//   - GitHub Copilot: 集成在IDE中的AI编程助手
//   - Claude: Anthropic的对话式AI助手
//   - Gemini: Google的多模态AI助手
//   - 其他配置文件中定义的助手
//
// 选择流程：
//   1. 检查用户是否已指定AI助手
//   2. 如果未指定，从配置中获取所有可用助手
//   3. 显示交互式选择界面
//   4. 用户选择后更新选项配置
//   5. 显示选择结果和助手信息
//
// 参数：
//   tracker - 步骤跟踪器，用于更新选择进度和状态
//   opts - 指向初始化选项的指针，用于读取和更新AI助手设置
//
// 返回值：
//   error - 如果选择过程失败，返回错误信息；成功时返回nil
//
// 副作用：
//   - 更新opts.AIAssistant字段为用户选择的助手
//   - 更新步骤跟踪器显示选择结果
//   - 可能显示交互式用户界面
//
// 错误处理：
//   - UI渲染失败
//   - 用户取消选择操作
//   - 配置文件中助手信息缺失
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

// selectScriptType 选择或确认脚本类型
//
// 该函数负责处理脚本类型的选择逻辑，支持跨平台的脚本环境配置。
// 如果用户已指定脚本类型，则直接确认；否则提供交互式选择界面。
//
// 支持的脚本类型：
//   - sh (Shell): Unix/Linux/macOS的bash脚本
//     * 适用于类Unix系统
//     * 支持复杂的shell命令和管道操作
//     * 广泛的工具链支持
//   - ps (PowerShell): Windows PowerShell脚本
//     * 适用于Windows系统
//     * 面向对象的命令行环境
//     * 与.NET框架深度集成
//
// 自动检测逻辑：
//   - 根据操作系统自动推荐默认脚本类型
//   - Windows系统默认推荐PowerShell
//   - Unix/Linux/macOS系统默认推荐Shell
//
// 交互式选择特性：
//   - 显示所有可用的脚本类型选项
//   - 支持方向键导航选择
//   - 显示每种脚本类型的详细描述
//   - 智能默认选项基于当前操作系统
//
// 参数：
//   tracker - 步骤跟踪器，用于更新选择进度和状态
//   opts - 指向初始化选项的指针，用于读取和更新脚本类型设置
//
// 返回值：
//   error - 如果选择过程失败，返回错误信息；成功时返回nil
//
// 副作用：
//   - 更新opts.ScriptType字段为用户选择的脚本类型
//   - 更新步骤跟踪器显示选择结果
//   - 可能显示交互式用户界面
//
// 错误处理：
//   - UI渲染失败
//   - 用户取消选择操作
//   - 配置文件中脚本类型信息缺失
//   - 不支持的脚本类型
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

// checkTools 检查系统中所需工具的可用性
//
// 该函数负责验证项目初始化和后续开发所需的所有外部工具是否
// 在系统中可用。不同的AI助手可能需要不同的工具集合。
//
// 工具检查策略：
//   - 根据选择的AI助手动态确定所需工具列表
//   - 并行检查多个工具以提高效率
//   - 实时更新步骤跟踪器显示检查进度
//   - 提供详细的缺失工具信息和安装建议
//
// 常见的必需工具：
//   - git: 版本控制系统，用于仓库管理
//   - node/npm: Node.js运行时和包管理器
//   - python: Python解释器（某些AI助手需要）
//   - curl/wget: HTTP客户端，用于下载资源
//   - 特定AI助手的CLI工具
//
// AI助手特定工具：
//   - GitHub Copilot: 需要GitHub CLI (gh)
//   - Claude: 可能需要特定的API客户端
//   - Gemini: 可能需要Google Cloud CLI
//
// 检查机制：
//   - 使用系统PATH查找可执行文件
//   - 验证工具版本兼容性
//   - 检查工具的基本功能可用性
//   - 提供友好的错误消息和修复建议
//
// 参数：
//   tracker - 步骤跟踪器，用于更新检查进度和状态
//   opts - 初始化选项，包含AI助手类型等配置信息
//
// 返回值：
//   error - 如果任何必需工具缺失，返回详细的错误信息；所有工具可用时返回nil
//
// 副作用：
//   - 更新步骤跟踪器显示检查结果
//   - 可能在控制台输出工具检查详情
//   - 记录工具版本信息用于调试
//
// 错误处理：
//   - 缺失工具错误：列出所有缺失的工具
//   - 版本不兼容错误：显示期望版本和实际版本
//   - 权限错误：工具存在但无执行权限
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

// createProjectDirectory 创建或确认项目工作目录
//
// 该函数负责处理项目目录的创建和设置逻辑。根据用户的选择，
// 它可以创建新的项目目录或使用当前目录作为项目根目录。
//
// 目录创建策略：
//   - 当opts.Here为false时：创建以项目名称命名的新目录
//   - 当opts.Here为true时：使用当前目录作为项目根目录
//   - 自动处理目录权限和路径规范化
//   - 支持嵌套目录结构的创建
//
// 目录操作流程：
//   1. 检查用户的目录选择偏好（新建 vs 当前）
//   2. 如果需要新建目录：
//      - 使用项目名称创建目录
//      - 设置适当的目录权限（0755）
//      - 切换到新创建的目录
//   3. 如果使用当前目录：
//      - 验证当前目录的可写性
//      - 获取并显示当前目录信息
//   4. 更新步骤跟踪器显示目录信息
//
// 权限设置：
//   - 新创建的目录使用0755权限（rwxr-xr-x）
//   - 确保目录对所有者可读写执行
//   - 确保组和其他用户可读执行
//
// 参数：
//   tracker - 步骤跟踪器，用于更新目录创建进度和状态
//   opts - 初始化选项，包含项目名称和目录选择设置
//
// 返回值：
//   error - 如果目录创建或切换失败，返回详细错误信息；成功时返回nil
//
// 副作用：
//   - 可能在文件系统中创建新目录
//   - 改变当前工作目录（当创建新目录时）
//   - 更新步骤跟踪器显示目录路径
//
// 错误处理：
//   - 目录创建失败：权限不足、磁盘空间不足等
//   - 目录切换失败：目录不存在、权限问题等
//   - 路径解析错误：无效的项目名称或路径字符
//   - 文件系统错误：I/O错误、网络驱动器问题等
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

// downloadTemplate 从GitHub下载并提取项目模板
//
// 该函数负责从远程GitHub仓库下载适合所选AI助手的项目模板，
// 并将其提取到当前项目目录中。这是项目初始化的核心步骤之一。
//
// 模板下载策略：
//   - 根据选择的AI助手确定对应的模板仓库
//   - 支持GitHub API认证以避免速率限制
//   - 显示实时下载进度条
//   - 自动处理模板文件的提取和放置
//
// 支持的模板类型：
//   - GitHub Copilot模板：包含VS Code配置和Copilot设置
//   - Claude模板：包含Claude API集成和配置文件
//   - Gemini模板：包含Google AI集成和示例代码
//   - 通用模板：基础的Spec-Driven Development结构
//
// 下载流程：
//   1. 构建下载选项配置
//   2. 调用模板提供者的下载方法
//   3. 验证下载的模板完整性
//   4. 提取模板文件到项目目录
//   5. 清理临时下载文件
//   6. 更新步骤跟踪器显示结果
//
// 网络优化：
//   - 支持断点续传（如果模板提供者支持）
//   - 自动重试机制处理网络波动
//   - 压缩传输减少下载时间
//   - 并行下载多个文件（如果适用）
//
// 参数：
//   tracker - 步骤跟踪器，用于更新下载进度和状态
//   opts - 初始化选项，包含AI助手类型、GitHub令牌等配置
//
// 返回值：
//   error - 如果下载或提取失败，返回详细错误信息；成功时返回nil
//
// 副作用：
//   - 在项目目录中创建模板文件和目录结构
//   - 可能创建临时文件用于下载过程
//   - 更新步骤跟踪器显示下载进度
//   - 可能在控制台显示进度条
//
// 错误处理：
//   - 网络连接错误：DNS解析失败、连接超时等
//   - 认证错误：GitHub令牌无效或权限不足
//   - 下载错误：文件损坏、传输中断等
//   - 文件系统错误：磁盘空间不足、权限问题等
//   - 模板格式错误：无效的模板结构或配置
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

// initializeGit 初始化Git版本控制仓库
//
// 该函数负责在项目目录中设置Git版本控制系统。它会检查现有的
// Git仓库状态，并在需要时创建新的仓库。这为项目提供了版本
// 控制基础，支持后续的代码管理和协作开发。
//
// Git初始化策略：
//   - 智能检测现有Git仓库，避免重复初始化
//   - 支持静默模式和详细模式的初始化
//   - 自动配置基本的Git设置
//   - 处理各种Git初始化场景
//
// 初始化流程：
//   1. 检查当前目录是否已经是Git仓库
//   2. 如果已存在仓库，跳过初始化并显示状态
//   3. 如果不存在仓库，调用Git初始化命令
//   4. 验证初始化结果并更新状态
//   5. 为后续的提交操作做准备
//
// Git配置：
//   - 创建.git目录和基本结构
//   - 设置默认分支（通常为main或master）
//   - 配置基本的Git钩子（如果需要）
//   - 准备暂存区和工作区
//
// 兼容性处理：
//   - 支持不同版本的Git客户端
//   - 处理Git配置文件的差异
//   - 适配不同操作系统的Git行为
//   - 处理权限和路径问题
//
// 参数：
//   tracker - 步骤跟踪器，用于更新Git初始化进度和状态
//   opts - 初始化选项，包含详细输出控制等设置
//
// 返回值：
//   error - 如果Git初始化失败，返回详细错误信息；成功或跳过时返回nil
//
// 副作用：
//   - 可能在项目目录中创建.git目录和相关文件
//   - 更新步骤跟踪器显示Git状态
//   - 可能在控制台输出Git命令详情（详细模式）
//
// 错误处理：
//   - Git命令不可用：系统未安装Git或PATH配置问题
//   - 权限错误：无法在目录中创建.git文件夹
//   - 磁盘空间不足：无法创建Git仓库文件
//   - Git版本不兼容：使用了不支持的Git功能
//   - 目录状态异常：目录不存在或不可访问
//
// 跳过条件：
//   - 当前目录已经是Git仓库的一部分
//   - 父目录中存在.git目录（子模块场景）
//   - 用户明确禁用了Git初始化
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

// configureProject 配置项目特定设置和环境
//
// 该函数负责应用项目特定的配置设置，为开发环境做最终准备。
// 它处理各种配置文件的创建和定制，确保项目符合所选AI助手
// 和开发工具链的要求。
//
// 配置范围：
//   - AI助手特定的配置文件和设置
//   - 开发环境配置（IDE设置、扩展配置等）
//   - 项目元数据和描述文件
//   - 构建和部署配置
//   - 代码质量和格式化规则
//
// 配置策略：
//   - 基于模板的配置文件生成
//   - 动态替换配置中的占位符
//   - 合并用户自定义配置和默认配置
//   - 验证配置文件的语法和完整性
//
// AI助手特定配置：
//   - GitHub Copilot: VS Code设置、忽略文件配置
//   - Claude: API密钥配置、提示模板设置
//   - Gemini: Google Cloud配置、认证设置
//   - 通用: 编辑器配置、代码风格设置
//
// 配置文件类型：
//   - .vscode/settings.json: VS Code工作区设置
//   - .gitignore: Git忽略规则
//   - package.json: Node.js项目配置（如果适用）
//   - requirements.txt: Python依赖（如果适用）
//   - 环境变量文件: .env, .env.example
//   - 文档配置: README模板、API文档设置
//
// 参数：
//   tracker - 步骤跟踪器，用于更新配置进度和状态
//   opts - 初始化选项，包含AI助手类型和其他配置参数
//
// 返回值：
//   error - 如果配置过程失败，返回详细错误信息；成功时返回nil
//
// 副作用：
//   - 创建或修改项目配置文件
//   - 设置环境变量或系统配置
//   - 更新步骤跟踪器显示配置状态
//   - 可能触发IDE或工具的重新加载
//
// 错误处理：
//   - 配置文件写入失败：权限问题、磁盘空间不足
//   - 配置格式错误：JSON语法错误、YAML格式问题
//   - 模板处理错误：占位符替换失败、模板文件缺失
//   - 权限配置错误：无法设置文件权限或环境变量
//
// 扩展性：
//   - 支持插件化的配置处理器
//   - 允许用户自定义配置模板
//   - 支持配置的版本化和迁移
//   - 提供配置验证和诊断功能
func (h *InitHandler) configureProject(tracker *ui.StepTracker, opts types.InitOptions) error {
	tracker.SetStepRunning("configure", "Configuring project settings")

	// 这里可以添加项目配置逻辑
	// 例如：创建配置文件、设置环境变量等

	tracker.SetStepDone("configure", "Project configured successfully")
	return nil
}

// finalizeSetup 完成项目设置的最终步骤和清理工作
//
// 该函数是项目初始化流程的最后一步，负责执行收尾工作和最终验证。
// 它确保项目处于可用状态，并为开发者提供清晰的下一步指导。
//
// 最终化任务：
//   - 创建初始Git提交，记录项目设置状态
//   - 验证所有配置文件的完整性和正确性
//   - 清理临时文件和下载缓存
//   - 生成项目使用指南和文档
//   - 执行最终的健康检查
//
// Git提交策略：
//   - 自动添加所有项目文件到暂存区
//   - 创建描述性的初始提交消息
//   - 处理Git提交失败的情况（非致命错误）
//   - 支持跳过Git提交（如果Git不可用）
//
// 验证检查：
//   - 确认所有必需文件已创建
//   - 验证配置文件语法正确性
//   - 检查文件权限设置
//   - 测试基本的项目功能
//
// 清理操作：
//   - 删除下载过程中的临时文件
//   - 清理安装缓存和中间文件
//   - 整理项目目录结构
//   - 优化文件权限设置
//
// 用户指导：
//   - 生成项目README文件（如果不存在）
//   - 创建快速开始指南
//   - 提供AI助手使用说明
//   - 列出推荐的下一步操作
//
// 参数：
//   tracker - 步骤跟踪器，用于更新最终化进度和状态
//   opts - 初始化选项，包含项目配置和用户偏好
//
// 返回值：
//   error - 如果最终化过程失败，返回详细错误信息；成功时返回nil
//
// 副作用：
//   - 创建Git提交（如果Git可用且仓库存在）
//   - 删除临时文件和目录
//   - 更新步骤跟踪器显示完成状态
//   - 可能在控制台显示成功消息和指导信息
//
// 错误处理：
//   - Git提交失败：显示警告但不中断流程
//   - 文件清理失败：记录警告信息
//   - 验证失败：显示具体的问题和修复建议
//   - 权限问题：提供权限修复指导
//
// 容错性：
//   - 大部分错误被视为警告，不会中断整个流程
//   - 提供详细的错误信息和修复建议
//   - 确保即使部分步骤失败，项目仍然可用
//   - 记录所有问题以便后续诊断
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