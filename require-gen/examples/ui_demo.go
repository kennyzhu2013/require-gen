package main

import (
	"fmt"
	"time"

	"specify-cli/internal/ui"
)

// UI组件演示程序
//
// 这个演示程序展示了改进后的UI组件功能，包括：
// 1. 改进的GetKey函数 - 真正的单键捕获
// 2. 进度条组件 - 类似Python Rich的进度条
// 3. 主题系统 - 统一的颜色主题管理
//
// 运行方式：
//   go run examples/ui_demo.go

func main() {
	fmt.Println("=== require-gen UI组件演示 ===\n")

	// 演示1: 主题系统
	demonstrateThemes()

	// 演示2: 改进的GetKey函数
	demonstrateGetKey()

	// 演示3: 进度条组件
	demonstrateProgressBar()

	// 演示4: 多进度条
	demonstrateMultiProgressBar()

	fmt.Println("\n=== 演示完成 ===")
}

// demonstrateThemes 演示主题系统
func demonstrateThemes() {
	fmt.Println("📋 主题系统演示")
	fmt.Println("可用主题:", ui.ListGlobalThemes())
	fmt.Println()

	// 演示不同主题
	themes := []string{"default", "dark", "light", "colorful", "minimal"}

	for _, themeName := range themes {
		fmt.Printf("切换到主题: %s\n", themeName)
		err := ui.SetGlobalTheme(themeName)
		if err != nil {
			fmt.Printf("切换主题失败: %v\n", err)
			continue
		}

		// 显示横幅
		ui.ShowBanner()

		// 显示各种消息类型
		ui.ShowSuccess("这是成功消息")
		ui.ShowError("这是错误消息")
		ui.ShowWarning("这是警告消息")
		ui.ShowInfo("这是信息消息")

		fmt.Println("按任意键继续到下一个主题...")
		ui.GetKey()
		fmt.Println()
	}

	// 恢复默认主题
	ui.SetGlobalTheme("default")
}

// demonstrateGetKey 演示改进的GetKey函数
func demonstrateGetKey() {
	fmt.Println("⌨️  改进的GetKey函数演示")
	fmt.Println("现在支持真正的单键捕获，无需按Enter！")
	fmt.Println()

	fmt.Println("请按以下键进行测试：")
	fmt.Println("- 字母键 (a-z)")
	fmt.Println("- 数字键 (0-9)")
	fmt.Println("- 方向键 (↑↓←→)")
	fmt.Println("- 功能键 (F1-F12)")
	fmt.Println("- ESC键退出测试")
	fmt.Println()

	// 初始化键盘监听
	err := ui.InitKeyboard()
	if err != nil {
		ui.ShowError(fmt.Sprintf("键盘初始化失败: %v", err))
		return
	}
	defer ui.CloseKeyboard() // 确保在函数结束时清理资源

	for {
		fmt.Print("按键测试 (ESC退出): ")
		key, err := ui.GetKey()
		if err != nil {
			ui.ShowError(fmt.Sprintf("获取按键失败: %v", err))
			continue
		}

		// 立即显示检测到的按键
		ui.ShowSuccess(fmt.Sprintf("检测到按键: %s", key))

		if key == "Escape" {
			ui.ShowInfo("退出按键测试")
			break
		}
	}
	fmt.Println()
}

// demonstrateProgressBar 演示进度条组件
func demonstrateProgressBar() {
	fmt.Println("📊 进度条组件演示")
	fmt.Println()

	// 基础进度条
	fmt.Println("1. 基础进度条:")
	bar1 := ui.NewProgressBar(100, "下载文件")
	for i := int64(0); i <= 100; i += 5 {
		bar1.Update(i)
		time.Sleep(100 * time.Millisecond)
	}
	bar1.Finish()
	fmt.Println()

	// 自定义样式进度条
	fmt.Println("2. 自定义样式进度条:")
	bar2 := ui.NewProgressBar(50, "处理数据",
		ui.WithWidth(60),
		ui.ModernStyle(),
		ui.WithDisplayOptions(true, true, true, true),
	)
	for i := int64(0); i <= 50; i += 2 {
		bar2.Update(i)
		time.Sleep(80 * time.Millisecond)
	}
	bar2.Finish()
	fmt.Println()

	// 不同样式展示
	fmt.Println("3. 不同样式展示:")
	styles := []struct {
		name   string
		option ui.ProgressBarOption
	}{
		{"经典样式", ui.ClassicStyle()},
		{"现代样式", ui.ModernStyle()},
		{"简约样式", ui.MinimalStyle()},
		{"箭头样式", ui.ArrowStyle()},
		{"点状样式", ui.DotStyle()},
	}

	for _, style := range styles {
		fmt.Printf("%s: ", style.name)
		bar := ui.NewProgressBar(20, "",
			ui.WithWidth(30),
			style.option,
		)
		for i := int64(0); i <= 20; i++ {
			bar.Update(i)
			time.Sleep(50 * time.Millisecond)
		}
		bar.Finish()
	}
	fmt.Println()
}

// demonstrateMultiProgressBar 演示多进度条
func demonstrateMultiProgressBar() {
	fmt.Println("📊 多进度条演示")
	fmt.Println()

	// 创建多进度条管理器
	multiBar := ui.NewMultiProgressBar()

	// 创建多个进度条
	bar1 := ui.NewProgressBar(100, "任务1: 下载依赖", ui.WithWidth(40))
	bar2 := ui.NewProgressBar(80, "任务2: 编译代码", ui.WithWidth(40))
	bar3 := ui.NewProgressBar(60, "任务3: 运行测试", ui.WithWidth(40))

	// 添加到管理器
	multiBar.AddBar(bar1)
	multiBar.AddBar(bar2)
	multiBar.AddBar(bar3)

	// 开始显示
	multiBar.Start()

	// 模拟并发任务
	go func() {
		for i := int64(0); i < 100; i += 2 {
			bar1.Update(i)
			time.Sleep(50 * time.Millisecond)
		}
		bar1.Update(100) // 确保完成
	}()

	go func() {
		time.Sleep(1 * time.Second) // 延迟启动
		for i := int64(0); i < 80; i += 3 {
			bar2.Update(i)
			time.Sleep(60 * time.Millisecond)
		}
		bar2.Update(80) // 确保完成
	}()

	go func() {
		time.Sleep(2 * time.Second) // 延迟启动
		for i := int64(0); i < 60; i += 4 {
			bar3.Update(i)
			time.Sleep(70 * time.Millisecond)
		}
		bar3.Update(60) // 确保完成
	}()

	// 定期渲染
	for {
		multiBar.Render()
		time.Sleep(100 * time.Millisecond)

		// 检查是否全部完成
		if bar1.IsCompleted() && bar2.IsCompleted() && bar3.IsCompleted() {
			break
		}
	}

	// 停止显示
	multiBar.Stop()
	ui.ShowSuccess("所有任务完成！")
	fmt.Println()
}
