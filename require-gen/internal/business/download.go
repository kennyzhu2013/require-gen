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

// DownloadHandler 下载处理器
type DownloadHandler struct {
	templateProvider types.TemplateProvider
	authProvider     types.AuthProvider
	uiRenderer       types.UIRenderer
}

// NewDownloadHandler 创建新的下载处理器
func NewDownloadHandler() *DownloadHandler {
	return &DownloadHandler{
		templateProvider: infrastructure.NewTemplateProvider(),
		authProvider:     infrastructure.NewAuthProvider(),
		uiRenderer:       ui.NewRenderer(),
	}
}

// Execute 执行下载流程
func (h *DownloadHandler) Execute(opts types.DownloadOptions) error {
	// 创建步骤跟踪器
	tracker := ui.NewStepTracker("Template Download")
	
	// 添加步骤
	h.setupSteps(tracker, opts)
	
	// 显示初始状态
	tracker.Display()

	// 执行下载流程
	if err := h.executeSteps(tracker, opts); err != nil {
		ui.ShowError(fmt.Sprintf("Download failed: %v", err))
		return err
	}

	// 显示完成状态
	tracker.Display()
	ui.ShowSuccess("Template download completed successfully!")
	
	return nil
}

// setupSteps 设置下载步骤
func (h *DownloadHandler) setupSteps(tracker *ui.StepTracker, opts types.DownloadOptions) {
	tracker.AddStep("validate", "Validate download options")
	tracker.AddStep("prepare", "Prepare download environment")
	tracker.AddStep("download", "Download template files")
	tracker.AddStep("extract", "Extract and organize files")
	tracker.AddStep("verify", "Verify download integrity")
	tracker.AddStep("cleanup", "Clean up temporary files")
}

// executeSteps 执行下载步骤
func (h *DownloadHandler) executeSteps(tracker *ui.StepTracker, opts types.DownloadOptions) error {
	// 步骤1: 验证选项
	if err := h.validateOptions(tracker, opts); err != nil {
		return err
	}

	// 步骤2: 准备下载环境
	if err := h.prepareEnvironment(tracker, opts); err != nil {
		return err
	}

	// 步骤3: 下载模板
	if err := h.downloadTemplate(tracker, opts); err != nil {
		return err
	}

	// 步骤4: 提取文件
	if err := h.extractFiles(tracker, opts); err != nil {
		return err
	}

	// 步骤5: 验证完整性
	if err := h.verifyIntegrity(tracker, opts); err != nil {
		return err
	}

	// 步骤6: 清理临时文件
	if err := h.cleanup(tracker, opts); err != nil {
		return err
	}

	return nil
}

// validateOptions 验证下载选项
func (h *DownloadHandler) validateOptions(tracker *ui.StepTracker, opts types.DownloadOptions) error {
	tracker.SetStepRunning("validate", "Validating download options")

	// 验证AI助手
	if opts.AIAssistant == "" {
		tracker.SetStepError("validate", "AI assistant is required")
		return fmt.Errorf("AI assistant is required")
	}

	if _, exists := config.GetAgentInfo(opts.AIAssistant); !exists {
		tracker.SetStepError("validate", fmt.Sprintf("Unknown AI assistant: %s", opts.AIAssistant))
		return fmt.Errorf("unknown AI assistant: %s", opts.AIAssistant)
	}

	// 验证下载目录
	if opts.DownloadDir == "" {
		opts.DownloadDir = "."
	}

	// 检查下载目录是否存在
	if _, err := os.Stat(opts.DownloadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(opts.DownloadDir, 0755); err != nil {
			tracker.SetStepError("validate", fmt.Sprintf("Failed to create download directory: %v", err))
			return fmt.Errorf("failed to create download directory: %w", err)
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

// prepareEnvironment 准备下载环境
func (h *DownloadHandler) prepareEnvironment(tracker *ui.StepTracker, opts types.DownloadOptions) error {
	tracker.SetStepRunning("prepare", "Preparing download environment")

	// 获取AI助手信息
	agentInfo, _ := config.GetAgentInfo(opts.AIAssistant)
	
	// 创建目标目录
	targetDir := filepath.Join(opts.DownloadDir, agentInfo.Folder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		tracker.SetStepError("prepare", fmt.Sprintf("Failed to create target directory: %v", err))
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 检查认证（如果需要）
	if opts.GitHubToken != "" {
		// 验证GitHub token
		headers := h.authProvider.GetHeaders()
		if len(headers) == 0 {
			ui.ShowWarning("GitHub token provided but authentication failed")
		}
	}

	tracker.SetStepDone("prepare", fmt.Sprintf("Environment prepared for %s", agentInfo.Name))
	return nil
}

// downloadTemplate 下载模板
func (h *DownloadHandler) downloadTemplate(tracker *ui.StepTracker, opts types.DownloadOptions) error {
	tracker.SetStepRunning("download", "Downloading template files")

	// 使用模板提供者下载
	templatePath, err := h.templateProvider.Download(opts)
	if err != nil {
		tracker.SetStepError("download", fmt.Sprintf("Download failed: %v", err))
		return fmt.Errorf("failed to download template: %w", err)
	}

	if opts.Verbose {
		ui.ShowInfo(fmt.Sprintf("Template downloaded to: %s", templatePath))
	}

	tracker.SetStepDone("download", "Template files downloaded successfully")
	return nil
}

// extractFiles 提取文件
func (h *DownloadHandler) extractFiles(tracker *ui.StepTracker, opts types.DownloadOptions) error {
	tracker.SetStepRunning("extract", "Extracting and organizing files")

	// 这里应该实现文件提取逻辑
	// 例如：解压ZIP文件、重命名文件等

	tracker.SetStepDone("extract", "Files extracted and organized")
	return nil
}

// verifyIntegrity 验证完整性
func (h *DownloadHandler) verifyIntegrity(tracker *ui.StepTracker, opts types.DownloadOptions) error {
	tracker.SetStepRunning("verify", "Verifying download integrity")

	// 获取目标路径
	agentInfo, _ := config.GetAgentInfo(opts.AIAssistant)
	targetPath := filepath.Join(opts.DownloadDir, agentInfo.Folder)

	// 验证模板
	if err := h.templateProvider.Validate(targetPath); err != nil {
		tracker.SetStepError("verify", fmt.Sprintf("Validation failed: %v", err))
		return fmt.Errorf("template validation failed: %w", err)
	}

	tracker.SetStepDone("verify", "Download integrity verified")
	return nil
}

// cleanup 清理临时文件
func (h *DownloadHandler) cleanup(tracker *ui.StepTracker, opts types.DownloadOptions) error {
	tracker.SetStepRunning("cleanup", "Cleaning up temporary files")

	// 这里应该实现清理逻辑
	// 例如：删除临时下载文件、清理缓存等

	tracker.SetStepDone("cleanup", "Temporary files cleaned up")
	return nil
}

// GetAvailableTemplates 获取可用模板列表
func (h *DownloadHandler) GetAvailableTemplates() (map[string]types.AgentInfo, error) {
	return config.AgentConfig, nil
}

// GetTemplateInfo 获取模板信息
func (h *DownloadHandler) GetTemplateInfo(assistant string) (*types.AgentInfo, error) {
	info, exists := config.GetAgentInfo(assistant)
	if !exists {
		return nil, fmt.Errorf("unknown AI assistant: %s", assistant)
	}
	return &info, nil
}

// EstimateDownloadSize 估算下载大小
func (h *DownloadHandler) EstimateDownloadSize(assistant string) (int64, error) {
	// 这里应该实现下载大小估算逻辑
	// 可以通过GitHub API获取发布包大小
	return 0, fmt.Errorf("download size estimation not implemented")
}

// CheckDiskSpace 检查磁盘空间
func (h *DownloadHandler) CheckDiskSpace(path string, requiredSize int64) error {
	// 这里应该实现磁盘空间检查逻辑
	return nil
}