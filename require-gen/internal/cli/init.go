package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"specify-cli/internal/business"
	"specify-cli/internal/config"
	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

var (
	// init命令的标志
	projectName string
	here        bool
	aiAssistant string
	scriptType  string
	githubToken string
	// 新增的CLI标志
	force       bool
	noGit       bool
	ignoreTools bool
	skipTLS     bool
)

// initCmd init子命令
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new spec-driven project",
	Long: `Initialize a new spec-driven development project with AI assistant integration.

This command will:
1. Create a new project directory (unless --here is used)
2. Download the appropriate template for your chosen AI assistant
3. Set up the project structure and configuration
4. Initialize Git repository (if not already present)
5. Create initial commit with project setup

Examples:
  specify init my-project                    # Create new project in ./my-project
  specify init --here                       # Initialize in current directory
  specify init my-project --ai claude-code  # Use specific AI assistant
  specify init my-project --script ps       # Use PowerShell scripts
  specify init my-project --force           # Force overwrite existing directory
  specify init my-project --no-git          # Skip Git repository initialization
  specify init my-project --ignore-agent-tools  # Ignore tool availability checks
  specify init my-project --skip-tls        # Skip TLS certificate verification`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

// initHelpFunc 自定义init命令的help函数，在显示help前先显示banner
func initHelpFunc(cmd *cobra.Command, args []string) {
	// 显示banner
	ui.ShowBanner()
	// 调用默认的help函数
	cmd.Parent().HelpFunc()(cmd, args)
}

func init() {
	// 设置自定义help函数
	initCmd.SetHelpFunc(initHelpFunc)
	
	// 添加init命令的标志
	initCmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name (overrides positional argument)")
	initCmd.Flags().BoolVar(&here, "here", false, "Initialize in current directory")
	initCmd.Flags().StringVarP(&aiAssistant, "ai", "a", "", "AI assistant to use")
	initCmd.Flags().StringVarP(&scriptType, "script", "s", "", "Script type (sh/ps)")
	initCmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub token for private repositories")
	
	// 新增的CLI标志
	initCmd.Flags().BoolVar(&force, "force", false, "Force overwrite existing project directory")
	initCmd.Flags().BoolVar(&noGit, "no-git", false, "Skip Git repository initialization")
	initCmd.Flags().BoolVar(&ignoreTools, "ignore-agent-tools", false, "Ignore AI assistant tool availability checks")
	initCmd.Flags().BoolVar(&skipTLS, "skip-tls", false, "Skip TLS certificate verification")
}

// runInit 执行init命令
func runInit(cmd *cobra.Command, args []string) error {
	// 解析项目名称
	if len(args) > 0 && projectName == "" {
		projectName = args[0]
	}

	// 如果没有指定项目名称且没有使用--here，则默认在当前目录初始化
	if !here && projectName == "" {
		here = true
	}

	// 构建初始化选项
	opts := types.InitOptions{
		ProjectName:  projectName,
		Here:         here,
		AIAssistant:  aiAssistant,
		ScriptType:   scriptType,
		GitHubToken:  githubToken,
		Verbose:      GetVerbose(),
		Debug:        GetDebug(),
		// 新增的CLI标志
		Force:        force,
		NoGit:        noGit,
		IgnoreTools:  ignoreTools,
		SkipTLS:      skipTLS,
	}

	// 显示横幅
	ui.ShowBanner()

	// 创建业务逻辑处理器
	initHandler := business.NewInitHandler()

	// 执行初始化流程
	return initHandler.Execute(opts)
}

// validateInitOptions 验证初始化选项
func validateInitOptions(opts *types.InitOptions) error {
	// 验证AI助手
	if opts.AIAssistant != "" {
		if _, exists := config.GetAgentInfo(opts.AIAssistant); !exists {
			return fmt.Errorf("unknown AI assistant: %s", opts.AIAssistant)
		}
	}

	// 验证脚本类型
	if opts.ScriptType != "" {
		if _, exists := config.GetScriptType(opts.ScriptType); !exists {
			return fmt.Errorf("unknown script type: %s", opts.ScriptType)
		}
	}

	// 验证项目目录
	if !opts.Here && opts.ProjectName != "" {
		// 检查目录是否已存在
		if _, err := os.Stat(opts.ProjectName); err == nil {
			if !opts.Force {
				return fmt.Errorf("directory '%s' already exists. Use --force to overwrite", opts.ProjectName)
			}
		}
	} else if opts.Here {
		// 检查当前目录是否为空
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		entries, err := os.ReadDir(cwd)
		if err != nil {
			return fmt.Errorf("failed to read current directory: %w", err)
		}

		if len(entries) > 0 && !opts.Force {
			return fmt.Errorf("current directory '%s' is not empty. Use --force to overwrite", filepath.Base(cwd))
		}
	}

	return nil
}