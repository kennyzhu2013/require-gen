package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"specify-cli/internal/ui"
)

// main 简化的UI组件测试程序
func main() {
	fmt.Println("=== UI组件功能测试 ===")
	fmt.Println()
	
	// 测试1: 增强Banner
	testEnhancedBanner()
	
	// 测试2: Panel组件
	testPanelComponent()
	
	// 测试3: Tree组件
	testTreeComponent()
	
	fmt.Println("所有测试完成！")
}

// testEnhancedBanner 测试增强的Banner功能
func testEnhancedBanner() {
	fmt.Println("🎨 测试增强Banner功能")
	fmt.Println(strings.Repeat("-", 40))
	
	// 显示增强版Banner
	ui.ShowBannerEnhanced()
	
	fmt.Println("✓ 增强Banner测试完成")
	fmt.Println()
}

// testPanelComponent 测试Panel组件
func testPanelComponent() {
	fmt.Println("📋 测试Panel组件")
	fmt.Println(strings.Repeat("-", 40))
	
	// 基础Panel测试
	panel := ui.NewPanel(
		"这是一个测试Panel\n包含多行内容\n用于验证功能",
		"测试Panel",
		ui.WithBorderStyle(color.FgGreen),
		ui.WithPadding(1, 2))
	panel.Render()
	fmt.Println()
	
	// 信息Panel测试
	infoItems := map[string]string{
		"组件":   "Panel",
		"状态":   "正常",
		"功能":   "信息展示",
	}
	infoPanel := ui.CreateInfoPanel("组件信息", infoItems)
	infoPanel.Render()
	fmt.Println()
	
	// 消息Panel测试
	successPanel := ui.CreateMessagePanel("Panel组件测试成功！", "success")
	successPanel.Render()
	fmt.Println()
	
	fmt.Println("✓ Panel组件测试完成")
	fmt.Println()
}

// testTreeComponent 测试Tree组件
func testTreeComponent() {
	fmt.Println("🌳 测试Tree组件")
	fmt.Println(strings.Repeat("-", 40))
	
	// 创建测试树
	tree := ui.NewTree("测试进度树",
		ui.WithTitleColor(color.FgCyan),
		ui.WithCompactMode(false))
	
	// 添加测试节点
	tree.Add("初始化", "completed", "已完成")
	node2 := tree.Add("处理中", "running", "正在执行")
	tree.Add("等待中", "pending", "")
	tree.Add("已跳过", "skipped", "跳过执行")
	
	// 添加子节点
	tree.AddChild(node2, "子任务1", "completed", "完成")
	tree.AddChild(node2, "子任务2", "running", "进行中")
	tree.AddChild(node2, "子任务3", "pending", "")
	
	// 渲染树
	tree.Render()
	fmt.Println()
	
	fmt.Println("✓ Tree组件测试完成")
	fmt.Println()
}