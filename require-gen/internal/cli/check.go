package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"specify-cli/internal/config"
	"specify-cli/internal/infrastructure"
	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

var (
	// check命令的标志
	showVersions bool
	showDetails  bool
)

// checkCmd check子命令
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that all required tools are installed",
	Long: `Check the availability of required development tools and AI assistants.

This command will:
1. Check for Git version control system
2. Check for available AI assistant tools (Claude, Gemini, etc.)
3. Check for Visual Studio Code variants
4. Display system information and tool versions
5. Provide installation suggestions for missing tools

Examples:
  specify check                    # Basic tool availability check
  specify check --versions         # Show tool versions
  specify check --details          # Show detailed system information`,
	Run: runCheck,
}

func init() {
	// 添加check命令的标志
	checkCmd.Flags().BoolVar(&showVersions, "versions", false, "Show versions of available tools")
	checkCmd.Flags().BoolVar(&showDetails, "details", false, "Show detailed system information")
}

// runCheck 执行check命令
func runCheck(cmd *cobra.Command, args []string) {
	// 显示横幅
	ui.ShowBanner()
	
	fmt.Println("🔍 Checking for installed tools...")
	fmt.Println()

	// 创建步骤跟踪器
	tracker := ui.NewStepTracker("Check Available Tools")
	
	// 创建工具检查器
	toolChecker := infrastructure.NewToolChecker()

	// 检查Git
	tracker.AddStep("git", "Git version control")
	gitAvailable := toolChecker.CheckTool("git", &types.StepTracker{
		Title: "Git Check",
		Steps: make(map[string]*types.Step),
	})
	
	if gitAvailable {
		tracker.SetStepDone("git", "available")
		if showVersions {
			if version, err := toolChecker.GetToolVersion("git"); err == nil {
				tracker.SetStepDone("git", fmt.Sprintf("available (v%s)", version))
			}
		}
	} else {
		tracker.SetStepError("git", "not found")
	}

	// 检查AI助手工具
	agents := config.GetAllAgents()
	agentResults := make(map[string]bool)
	
	for agentKey, agentName := range agents {
		tracker.AddStep(agentKey, agentName)
		
		// 获取AI助手信息
		agentInfo, exists := config.GetAgentInfo(agentKey)
		if !exists || !agentInfo.RequiresCLI {
			tracker.SetStepSkipped(agentKey, "no CLI required")
			agentResults[agentKey] = true // 不需要CLI的视为可用
			continue
		}
		
		// 检查需要CLI的AI助手
		available := toolChecker.CheckTool(agentKey, &types.StepTracker{
			Title: fmt.Sprintf("%s Check", agentName),
			Steps: make(map[string]*types.Step),
		})
		
		agentResults[agentKey] = available
		
		if available {
			tracker.SetStepDone(agentKey, "available")
			if showVersions {
				if version, err := toolChecker.GetToolVersion(agentKey); err == nil {
					tracker.SetStepDone(agentKey, fmt.Sprintf("available (v%s)", version))
				}
			}
		} else {
			tracker.SetStepError(agentKey, "not found")
		}
	}

	// 检查VS Code变体
	tracker.AddStep("code", "Visual Studio Code")
	codeAvailable := toolChecker.CheckTool("code", &types.StepTracker{
		Title: "VS Code Check",
		Steps: make(map[string]*types.Step),
	})
	
	if codeAvailable {
		tracker.SetStepDone("code", "available")
		if showVersions {
			if version, err := toolChecker.GetToolVersion("code"); err == nil {
				tracker.SetStepDone("code", fmt.Sprintf("available (v%s)", version))
			}
		}
	} else {
		tracker.SetStepError("code", "not found")
	}

	tracker.AddStep("code-insiders", "Visual Studio Code Insiders")
	codeInsidersAvailable := toolChecker.CheckTool("code-insiders", &types.StepTracker{
		Title: "VS Code Insiders Check",
		Steps: make(map[string]*types.Step),
	})
	
	if codeInsidersAvailable {
		tracker.SetStepDone("code-insiders", "available")
		if showVersions {
			if version, err := toolChecker.GetToolVersion("code-insiders"); err == nil {
				tracker.SetStepDone("code-insiders", fmt.Sprintf("available (v%s)", version))
			}
		}
	} else {
		tracker.SetStepError("code-insiders", "not found")
	}

	// 显示检查结果
	tracker.Display()
	fmt.Println()

	// 显示总结信息
	fmt.Println("✅ Specify CLI is ready to use!")
	fmt.Println()

	// 提供安装建议
	if !gitAvailable {
		fmt.Println("💡 Tip: Install git for repository management")
	}

	availableAgents := 0
	for _, available := range agentResults {
		if available {
			availableAgents++
		}
	}

	if availableAgents == 0 {
		fmt.Println("💡 Tip: Install an AI assistant for the best experience")
		fmt.Println()
		fmt.Println("Available AI assistants:")
		for agentKey, agentName := range agents {
			if agentInfo, exists := config.GetAgentInfo(agentKey); exists && agentInfo.RequiresCLI && agentInfo.InstallURL != "" {
				fmt.Printf("  • %s: %s\n", agentName, agentInfo.InstallURL)
			}
		}
	}

	// 显示详细系统信息
	if showDetails {
		fmt.Println()
		fmt.Println("=== System Information ===")
		systemInfo := toolChecker.GetSystemInfo()
		for key, value := range systemInfo {
			fmt.Printf("  %-15s: %s\n", key, value)
		}
		
		fmt.Printf("  %-15s: %s\n", "Go Version", runtime.Version())
		fmt.Printf("  %-15s: %s/%s\n", "OS/Arch", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("  %-15s: %s\n", "Compiler", runtime.Compiler)
	}
}