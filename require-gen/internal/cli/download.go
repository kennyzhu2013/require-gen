package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"specify-cli/internal/business"
	"specify-cli/internal/types"
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

func init() {
	// 添加download命令的标志
	downloadCmd.Flags().StringVarP(&downloadDir, "dir", "d", ".", "Download directory")
	downloadCmd.Flags().StringVarP(&scriptType, "script", "s", "", "Script type (sh/ps)")
	downloadCmd.Flags().BoolVar(&showProgress, "progress", true, "Show download progress")
	downloadCmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub token for private repositories")
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