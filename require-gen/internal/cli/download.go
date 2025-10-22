package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"specify-cli/internal/business"
	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

var (
	// download命令的标志
	downloadDir  string
	showProgress bool
)

// downloadCmd download子命令
var downloadCmd = &cobra.Command{
	Use:   "download [ai-assistant]",
	Short: "Download AI assistant templates",
	Long: `Download templates for specific AI assistants.

This command allows you to download templates without initializing a project.
Useful for offline development or custom project setups.

Examples:
  specify download claude-code              # Download Claude templates
  specify download github-copilot --dir ./templates  # Download to specific directory
  specify download --progress               # Show download progress`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDownload,
}

// downloadHelpFunc 自定义download命令的help函数，在显示help前先显示banner
func downloadHelpFunc(cmd *cobra.Command, args []string) {
	// 显示banner
	ui.ShowBanner()
	// 调用默认的help函数
	cmd.Parent().HelpFunc()(cmd, args)
}

func init() {
	// 设置自定义help函数
	downloadCmd.SetHelpFunc(downloadHelpFunc)
	
	// 添加download命令的标志
	downloadCmd.Flags().StringVar(&downloadDir, "dir", "", "Directory to download templates to")
	downloadCmd.Flags().BoolVar(&showProgress, "progress", false, "Show download progress")
}

// runDownload 执行download命令
func runDownload(cmd *cobra.Command, args []string) error {
	// 解析AI助手参数
	var assistant string
	if len(args) > 0 {
		assistant = args[0]
	}

	// 如果没有指定AI助手，提示用户选择
	if assistant == "" {
		return fmt.Errorf("AI assistant is required. Use 'specify download <ai-assistant>'")
	}

	// 构建下载选项
	opts := types.DownloadOptions{
		AIAssistant:  assistant,
		DownloadDir:  downloadDir,
		ScriptType:   scriptType,
		Verbose:      GetVerbose(),
		ShowProgress: showProgress,
		GitHubToken:  githubToken,
	}

	// 创建业务逻辑处理器
	downloadHandler := business.NewDownloadHandler()

	// 执行下载流程
	return downloadHandler.Execute(opts)
}