package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
	"specify-cli/internal/config"
)

// ShowBannerEnhanced 显示增强版横幅，支持渐变色彩和居中对齐
//
// 该函数实现了与Python版本相同的视觉效果：
// - 6色渐变循环显示Banner
// - 自动居中对齐
// - 斜体黄色标语
// - 终端宽度自适应
//
// 颜色序列：
// - 亮蓝色 -> 蓝色 -> 青色 -> 亮青色 -> 白色 -> 亮白色
//
// 示例输出：
//   [渐变色彩的ASCII艺术横幅，居中显示]
//   [居中的斜体黄色标语]
func ShowBannerEnhanced() {
	banner := config.Banner
	tagline := config.Tagline
	
	// 定义颜色序列（与Python版本保持一致）
	colors := []color.Attribute{
		color.FgHiBlue,   // bright_blue
		color.FgBlue,     // blue
		color.FgCyan,     // cyan
		color.FgHiCyan,   // bright_cyan
		color.FgWhite,    // white
		color.FgHiWhite,  // bright_white
	}
	
	// 分行处理Banner
	lines := strings.Split(strings.TrimSpace(banner), "\n")
	
	for i, line := range lines {
		// 循环使用颜色
		colorAttr := colors[i%len(colors)]
		colorFunc := color.New(colorAttr, color.Bold)
		
		// 左对齐显示，不添加padding
		colorFunc.Println(line)
	}
	
	// 标语左对齐显示（斜体黄色）
	if tagline != "" {
		// 注意：终端中的斜体支持有限，使用亮黄色代替
		taglineColor := color.New(color.FgHiYellow, color.Bold)
		taglineColor.Println(tagline)
	}
	
	fmt.Println()
}

// getTerminalWidth 获取终端宽度
//
// 使用golang.org/x/term包获取终端尺寸，如果获取失败则返回默认宽度80
//
// 返回值：
// - int: 终端宽度（字符数）
func getTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	return 80 // 默认宽度
}

// getDisplayWidth 获取字符串的显示宽度
//
// 考虑到ASCII艺术可能包含特殊字符，这里简化处理，
// 直接返回字符串长度。在实际应用中可能需要更复杂的
// Unicode宽度计算。
//
// 参数：
// - s: 要计算宽度的字符串
//
// 返回值：
// - int: 字符串的显示宽度
func getDisplayWidth(s string) int {
	// 简化实现：直接返回字符串长度
	// 在更复杂的场景中，可能需要使用github.com/mattn/go-runewidth
	// 来正确处理Unicode字符的显示宽度
	return len(s)
}

// ShowBannerWithStyle 显示带样式的横幅（备用实现）
//
// 提供更多自定义选项的横幅显示函数
//
// 参数：
// - banner: 横幅内容
// - tagline: 标语内容
// - colors: 自定义颜色序列
// - centerAlign: 是否居中对齐
func ShowBannerWithStyle(banner, tagline string, colors []color.Attribute, centerAlign bool) {
	if len(colors) == 0 {
		// 使用默认颜色序列
		colors = []color.Attribute{
			color.FgHiBlue, color.FgBlue, color.FgCyan,
			color.FgHiCyan, color.FgWhite, color.FgHiWhite,
		}
	}
	
	width := getTerminalWidth()
	lines := strings.Split(strings.TrimSpace(banner), "\n")
	
	for i, line := range lines {
		colorAttr := colors[i%len(colors)]
		colorFunc := color.New(colorAttr, color.Bold)
		
		if centerAlign {
			padding := (width - getDisplayWidth(line)) / 2
			if padding > 0 {
				fmt.Print(strings.Repeat(" ", padding))
			}
		}
		
		colorFunc.Println(line)
	}
	
	if tagline != "" {
		taglineColor := color.New(color.FgHiYellow, color.Bold)
		if centerAlign {
			taglinePadding := (width - getDisplayWidth(tagline)) / 2
			if taglinePadding > 0 {
				fmt.Print(strings.Repeat(" ", taglinePadding))
			}
		}
		taglineColor.Println(tagline)
	}
	
	fmt.Println()
}