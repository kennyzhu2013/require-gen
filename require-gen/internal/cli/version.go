package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"specify-cli/internal/config"
)

// versionCmd version子命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display version, build information, and runtime details.",
	Run:   runVersion,
}

// configCmd config子命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration information",
	Long:  "Display current configuration including available AI assistants and script types.",
	Run:   runConfig,
}

// runVersion 执行version命令
func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Specify CLI v%s\n", rootCmd.Version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Built with: %s\n", runtime.Compiler)
}

// runConfig 执行config命令
func runConfig(cmd *cobra.Command, args []string) {
	fmt.Println("=== Specify CLI Configuration ===")
	fmt.Printf("Version: %s\n", rootCmd.Version)
	fmt.Printf("Default Script Type: %s\n", config.GetDefaultScriptType())
	
	fmt.Println("\n=== Available AI Assistants ===")
	agents := config.GetAllAgents()
	for key, name := range agents {
		info, _ := config.GetAgentInfo(key)
		fmt.Printf("  %-15s %s", key, name)
		if info.RequiresCLI {
			fmt.Printf(" (requires CLI)")
		}
		if info.InstallURL != "" {
			fmt.Printf("\n                  Install: %s", info.InstallURL)
		}
		fmt.Println()
	}

	fmt.Println("\n=== Available Script Types ===")
	scripts := config.GetAllScriptTypes()
	for key, desc := range scripts {
		scriptInfo, _ := config.GetScriptType(key)
		fmt.Printf("  %-5s %s (%s)\n", key, desc, scriptInfo.Extension)
	}

	fmt.Println("\n=== Runtime Information ===")
	fmt.Printf("  OS: %s\n", runtime.GOOS)
	fmt.Printf("  Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("  Go Version: %s\n", runtime.Version())
}