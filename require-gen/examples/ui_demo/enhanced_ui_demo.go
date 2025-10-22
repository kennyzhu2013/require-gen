package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"specify-cli/internal/ui"
)

// main 演示增强UI功能的主函数
//
// 该程序展示了以下增强功能：
// 1. 增强的Banner显示（渐变色彩和居中对齐）
// 2. Panel组件的各种用法
// 3. 增强的Tree组件功能
// 4. 组合使用效果
func main() {
	fmt.Println("=== Require-Gen 增强UI功能演示 ===")
	fmt.Println()
	
	// 演示1: 增强的Banner显示
	demonstrateBanner()
	
	// 演示2: Panel组件功能
	demonstratePanel()
	
	// 演示3: 增强的Tree组件
	demonstrateTree()
	
	// 演示4: 组合使用效果
	demonstrateCombined()
	
	fmt.Println("演示完成！")
}

// demonstrateBanner 演示增强的Banner功能
func demonstrateBanner() {
	fmt.Println("🎨 演示1: 增强的Banner显示")
	fmt.Println(strings.Repeat("-", 50))
	
	// 显示增强版Banner
	ui.ShowBannerEnhanced()
	
	// 对比原版Banner
	fmt.Println("对比 - 原版Banner:")
	ui.ShowBanner()
	
	waitForUser()
}

// demonstratePanel 演示Panel组件功能
func demonstratePanel() {
	fmt.Println("📋 演示2: Panel组件功能")
	fmt.Println(strings.Repeat("-", 50))
	
	// 基础Panel
	basicPanel := ui.NewPanel(
		"这是一个基础的Panel组件\n支持多行文本显示\n可以包含各种信息",
		"基础Panel",
		ui.WithBorderStyle(color.FgCyan),
		ui.WithPadding(1, 2))
	basicPanel.Render()
	fmt.Println()
	
	// 信息Panel
	infoItems := map[string]string{
		"项目名称":   "require-gen",
		"版本":     "v1.0.0",
		"作者":     "Spec-Kit Team",
		"语言":     "Go",
		"状态":     "开发中",
	}
	infoPanel := ui.CreateInfoPanel("项目信息", infoItems)
	infoPanel.Render()
	fmt.Println()
	
	// 不同类型的消息Panel
	successPanel := ui.CreateMessagePanel("项目初始化成功完成！", "success")
	successPanel.Render()
	fmt.Println()
	
	warningPanel := ui.CreateMessagePanel("检测到配置文件已存在，将进行覆盖", "warning")
	warningPanel.Render()
	fmt.Println()
	
	errorPanel := ui.CreateMessagePanel("网络连接失败，请检查网络设置", "error")
	errorPanel.Render()
	fmt.Println()
	
	waitForUser()
}

// demonstrateTree 演示增强的Tree组件
func demonstrateTree() {
	fmt.Println("🌳 演示3: 增强的Tree组件")
	fmt.Println(strings.Repeat("-", 50))
	
	// 创建步骤跟踪树
	tree := ui.CreateProgressTree("项目初始化进度")
	
	// 添加步骤
	tree.Add("验证项目选项", "completed", "检查项目名称和路径")
	tree.Add("选择AI助手", "completed", "已选择: GitHub Copilot")
	step3 := tree.Add("检查工具依赖", "running", "正在检查Git和Node.js")
	tree.Add("创建项目目录", "pending", "")
	tree.Add("下载模板文件", "pending", "")
	
	// 为某些步骤添加子步骤
	tree.AddChild(step3, "检查Git", "completed", "版本 2.40.0")
	tree.AddChild(step3, "检查Node.js", "running", "正在验证...")
	tree.AddChild(step3, "检查npm", "pending", "")
	
	// 渲染树
	tree.Render()
	fmt.Println()
	
	// 模拟进度更新
	fmt.Println("模拟进度更新...")
	time.Sleep(1 * time.Second)
	
	// 更新状态
	if step3Node := tree.FindNode("检查工具依赖"); step3Node != nil {
		tree.UpdateNodeStatus(step3Node, "completed", "所有工具检查完成")
	}
	if step4Node := tree.FindNode("创建项目目录"); step4Node != nil {
		tree.UpdateNodeStatus(step4Node, "running", "正在创建目录结构")
	}
	
	// 重新渲染
	fmt.Print("\033[2J\033[H") // 清屏
	fmt.Println("🌳 演示3: 增强的Tree组件 (更新后)")
	fmt.Println(strings.Repeat("-", 50))
	tree.Render()
	fmt.Println()
	
	waitForUser()
}

// demonstrateCombined 演示组合使用效果
func demonstrateCombined() {
	fmt.Println("🎭 演示4: 组合使用效果")
	fmt.Println(strings.Repeat("-", 50))
	
	// 显示增强Banner
	ui.ShowBannerEnhanced()
	
	// 创建项目设置Panel
	setupLines := []string{
		"项目初始化设置",
		"",
		fmt.Sprintf("%-15s %s", "项目名称:", color.GreenString("my-awesome-project")),
		fmt.Sprintf("%-15s %s", "工作路径:", color.New(color.Faint).Sprint("/path/to/project")),
		fmt.Sprintf("%-15s %s", "AI助手:", color.CyanString("GitHub Copilot")),
		fmt.Sprintf("%-15s %s", "脚本类型:", color.YellowString("PowerShell")),
	}
	
	setupPanel := ui.NewPanel(
		strings.Join(setupLines, "\n"),
		"",
		ui.WithBorderStyle(color.FgCyan),
		ui.WithPadding(1, 2))
	setupPanel.Render()
	fmt.Println()
	
	// 创建进度树
	progressTree := ui.NewTree("初始化进度",
		ui.WithTitleColor(color.FgHiCyan),
		ui.WithCompactMode(false))
	
	progressTree.Add("✓ 验证配置", "completed", "配置验证通过")
	progressTree.Add("✓ 创建目录", "completed", "项目目录已创建")
	progressTree.Add("◐ 下载模板", "running", "正在下载...")
	progressTree.Add("○ 初始化Git", "pending", "")
	progressTree.Add("○ 安装依赖", "pending", "")
	
	progressTree.Render()
	fmt.Println()
	
	// 显示完成消息
	completionPanel := ui.CreateMessagePanel(
		"项目初始化即将完成！\n请稍等片刻...",
		"info")
	completionPanel.Render()
	
	waitForUser()
}

// waitForUser 等待用户按键继续
func waitForUser() {
	fmt.Print("按Enter键继续...")
	fmt.Scanln()
	fmt.Println()
}