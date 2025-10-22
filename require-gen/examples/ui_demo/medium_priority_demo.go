package main

import (
	"fmt"
	"os"
	"time"

	"specify-cli/internal/ui"
)

func main() {
	fmt.Println("=== 中优先级UI改进功能演示 ===\n")

	// 1. 演示交互选择器增强
	demonstrateEnhancedSelector()

	// 2. 演示颜色主题统一
	demonstrateThemeManager()

	// 3. 演示错误处理优化
	demonstrateErrorHandler()

	fmt.Println("\n=== 演示完成 ===")
}

// demonstrateEnhancedSelector 演示增强的交互选择器
func demonstrateEnhancedSelector() {
	fmt.Println("1. 增强交互选择器演示")
	fmt.Println("   - 实时反馈和更好的视觉效果")
	fmt.Println("   - 支持方向键导航和ESC取消")
	fmt.Println()

	// 创建选项
	options := map[string]string{
		"copilot": "GitHub Copilot AI助手 - 代码补全和建议",
		"claude":  "Claude AI助手 - 智能对话和分析",
		"gemini":  "Google Gemini - 多模态AI助手",
		"gpt":     "OpenAI GPT - 强大的语言模型",
	}

	fmt.Println("请选择一个AI助手（使用方向键导航，Enter确认，Esc取消）:")

	// 使用增强选择器
	selected, err := ui.SelectWithEnhancedUI(options, "选择AI助手", "claude")
	if err != nil {
		fmt.Printf("选择被取消: %v\n", err)
	} else {
		fmt.Printf("您选择了: %s - %s\n", selected, options[selected])
	}

	fmt.Println()
	time.Sleep(2 * time.Second)
}

// demonstrateThemeManager 演示颜色主题统一
func demonstrateThemeManager() {
	fmt.Println("2. 颜色主题统一演示")
	fmt.Println("   - 统一所有UI组件的颜色方案")
	fmt.Println("   - 支持多种预定义主题")
	fmt.Println()

	tm := ui.GetThemeManager()

	// 展示所有可用主题
	themes := tm.ListThemes()
	fmt.Println("可用主题:")
	for _, themeName := range themes {
		fmt.Printf("- %s\n", themeName)
	}
	fmt.Println()

	// 演示不同主题的效果
	for _, themeName := range themes {
		fmt.Printf("=== %s 主题效果 ===\n", themeName)

		// 切换主题
		if err := tm.SetTheme(themeName); err != nil {
			fmt.Printf("切换主题失败: %v\n", err)
			continue
		}

		// 显示主题预览
		if theme, err := tm.GetTheme(themeName); err == nil {
			fmt.Printf("Primary: %s, Success: %s, Warning: %s, Error: %s\n",
				theme.Primary().Sprint("■"),
				theme.Success().Sprint("■"),
				theme.Warning().Sprint("■"),
				theme.Error().Sprint("■"))
		}

		// 创建示例面板展示主题效果
		panel := ui.NewPanel(
			fmt.Sprintf("这是使用 %s 主题的示例面板\n包含不同颜色的文本展示", themeName),
			fmt.Sprintf("%s 主题示例", themeName),
			ui.WithPadding(1, 2))
		panel.Render()

		fmt.Println()
		time.Sleep(1 * time.Second)
	}

	// 恢复默认主题
	tm.SetTheme("default")
}

// demonstrateErrorHandler 演示错误处理优化
func demonstrateErrorHandler() {
	fmt.Println("3. 错误处理优化演示")
	fmt.Println("   - 提供更清晰的错误信息和处理建议")
	fmt.Println("   - 支持不同错误级别和上下文信息")
	fmt.Println()

	eh := ui.GetGlobalErrorHandler()

	// 演示不同类型的错误处理

	// 1. 信息级别
	fmt.Println("信息级别示例:")
	ui.ShowInfo("这是一条信息消息")
	time.Sleep(2 * time.Second)

	// 2. 警告级别
	fmt.Println("警告级别示例:")
	ui.ShowWarning("这是一条警告消息")
	time.Sleep(2 * time.Second)

	// 3. 错误级别
	fmt.Println("错误级别示例:")
	ui.ShowError("这是一条错误消息")
	time.Sleep(2 * time.Second)

	// 4. 增强错误处理示例
	fmt.Println("增强错误处理示例:")
	ui.ShowEnhancedError("这是一条增强错误消息")
	time.Sleep(2 * time.Second)

	// 5. 网络错误示例
	fmt.Println("网络错误示例:")
	eh.ShowNetworkError("https://api.example.com/data",
		fmt.Errorf("connection timeout"))
	time.Sleep(2 * time.Second)

	// 6. 文件错误示例
	fmt.Println("文件错误示例:")
	eh.ShowFileError("/path/to/config.json", "read",
		fmt.Errorf("permission denied"))
	time.Sleep(2 * time.Second)

	// 7. 验证错误示例
	fmt.Println("验证错误示例:")
	eh.ShowValidationError("email", "invalid-email",
		fmt.Errorf("invalid email format"))
	time.Sleep(2 * time.Second)

	// 8. 配置错误示例
	fmt.Println("配置错误示例:")
	eh.ShowConfigError("database.host",
		fmt.Errorf("missing required configuration"))
	time.Sleep(2 * time.Second)

	// 9. 恢复消息示例
	fmt.Println("恢复消息示例:")
	eh.ShowRecoveryMessage("系统已成功恢复正常运行")
	time.Sleep(2 * time.Second)

	// 10. 自定义错误上下文示例
	fmt.Println("自定义错误上下文示例:")
	context := &ui.ErrorContext{
		Operation: "数据处理",
		Component: "数据分析器",
		Details: map[string]string{
			"数据源":  "user_data.csv",
			"处理行数": "1000",
			"失败行数": "5",
		},
		Timestamp: time.Now(),
		Level:     ui.ErrorLevelWarning,
		Suggestions: []string{
			"检查数据格式是否正确",
			"验证必填字段是否完整",
			"查看详细日志获取更多信息",
			"联系技术支持获取帮助",
		},
	}

	eh.HandleError(fmt.Errorf("数据处理过程中发现格式错误"), context)
	time.Sleep(3 * time.Second)
}

// 辅助函数：模拟用户输入
func simulateUserInput() {
	fmt.Println("（模拟用户交互，实际使用时会等待用户输入）")
}

// 辅助函数：等待用户按键
func waitForKey() {
	fmt.Print("按任意键继续...")
	var input string
	fmt.Scanln(&input)
}

// 程序入口点检查
func init() {
	// 检查是否在正确的目录中运行
	if _, err := os.Stat("../../internal/ui"); err != nil {
		fmt.Println("警告: 请在正确的项目目录中运行此演示程序")
		fmt.Println("当前工作目录应该包含 internal/ui 目录")
	}
}
