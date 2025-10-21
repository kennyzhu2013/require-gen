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
//
// Execute 是CLI应用程序的主入口点，负责启动Cobra命令行框架并处理
// 用户输入的命令和参数。该函数初始化根命令并开始命令解析和执行流程。
//
// 执行流程：
// 1. 解析命令行参数和标志
// 2. 验证命令语法和参数有效性
// 3. 路由到相应的子命令处理器
// 4. 执行命令逻辑并处理结果
// 5. 返回执行状态和错误信息
//
// 支持的全局标志：
// - --verbose, -v: 启用详细输出模式
// - --debug: 启用调试模式，显示详细的诊断信息
// - --help, -h: 显示帮助信息
// - --version: 显示版本信息
//
// 支持的子命令：
// - init: 初始化新的spec-driven项目
// - download: 下载项目模板和资源
// - version: 显示版本和系统信息
// - config: 管理配置设置
//
// 返回值：
// - error: 命令执行过程中的错误，nil表示成功
//
// 错误处理：
// - 命令语法错误：显示使用帮助
// - 参数验证失败：显示具体错误信息
// - 执行失败：返回详细的错误上下文
//
// 使用示例：
//   if err := Execute(); err != nil {
//       fmt.Fprintf(os.Stderr, "Error: %v\n", err)
//       os.Exit(1)
//   }
//
// 注意事项：
// - 该函数会阻塞直到命令执行完成
// - 错误信息会自动格式化并输出到stderr
// - 支持信号处理和优雅退出
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
	rootCmd.AddCommand(checkCmd)
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