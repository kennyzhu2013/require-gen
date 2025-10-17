package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"specify-cli/internal/config"
)

var (
	// 全局标志
	verbose bool
	debug   bool
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "specify",
	Short: "GitHub Spec Kit - Spec-Driven Development Toolkit",
	Long: fmt.Sprintf(`%s

%s

A powerful toolkit for spec-driven development with AI assistants.
Supports multiple AI platforms and script types for cross-platform development.`, 
		config.Banner, config.Tagline),
	Version: "1.0.0",
}

// Execute 执行CLI命令
func Execute() error {
	return rootCmd.Execute()
}

// init 初始化CLI命令
func init() {
	// 添加全局标志
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")

	// 添加子命令
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

// GetVerbose 获取verbose标志状态
func GetVerbose() bool {
	return verbose
}

// GetDebug 获取debug标志状态
func GetDebug() bool {
	return debug
}

// printError 打印错误信息
func printError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// printVerbose 打印详细信息
func printVerbose(format string, args ...interface{}) {
	if verbose {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

// printDebug 打印调试信息
func printDebug(format string, args ...interface{}) {
	if debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}