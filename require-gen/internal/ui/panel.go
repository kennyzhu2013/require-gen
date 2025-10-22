package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Panel 信息展示面板组件
//
// Panel提供了一个带边框的信息展示区域，类似于Rich库的Panel组件。
// 支持标题、自定义边框样式、内边距等功能。
//
// 特性：
// - 自定义标题和内容
// - 可配置的边框颜色和样式
// - 内边距控制
// - 自动宽度计算
// - Unicode边框字符
type Panel struct {
	content     string           // 面板内容
	title       string           // 面板标题
	borderStyle color.Attribute  // 边框颜色
	padding     [2]int          // 内边距 [垂直, 水平]
	minWidth    int             // 最小宽度
}

// PanelOption Panel配置选项函数类型
type PanelOption func(*Panel)

// WithBorderStyle 设置边框样式
func WithBorderStyle(style color.Attribute) PanelOption {
	return func(p *Panel) {
		p.borderStyle = style
	}
}

// WithPanelPadding 设置面板内边距
func WithPanelPadding(vertical, horizontal int) PanelOption {
	return func(p *Panel) {
		p.padding = [2]int{vertical, horizontal}
	}
}

// WithMinWidth 设置最小宽度
func WithMinWidth(width int) PanelOption {
	return func(p *Panel) {
		p.minWidth = width
	}
}

// NewPanel 创建新的Panel实例
//
// 参数：
// - content: 面板内容
// - title: 面板标题（可为空）
// - options: 配置选项
//
// 返回值：
// - *Panel: Panel实例
func NewPanel(content, title string, options ...PanelOption) *Panel {
	panel := &Panel{
		content:     content,
		title:       title,
		borderStyle: color.FgCyan,
		padding:     [2]int{1, 2}, // 默认垂直1，水平2
		minWidth:    0,
	}
	
	// 应用配置选项
	for _, opt := range options {
		opt(panel)
	}
	
	return panel
}

// Render 渲染Panel到终端
//
// 该方法将Panel渲染为带边框的文本框，包括：
// - 顶部边框（可选标题）
// - 内容区域（带内边距）
// - 底部边框
//
// 边框字符使用Unicode字符：
// - ╔ ╗ ╚ ╝ (角落)
// - ═ (水平线)
// - ║ (垂直线)
func (p *Panel) Render() string {
	lines := strings.Split(p.content, "\n")
	
	// 计算内容的最大宽度
	maxContentWidth := 0
	for _, line := range lines {
		if width := getDisplayWidth(line); width > maxContentWidth {
			maxContentWidth = width
		}
	}
	
	// 计算总宽度（内容 + 水平内边距）
	totalWidth := maxContentWidth + p.padding[1]*2
	if totalWidth < p.minWidth {
		totalWidth = p.minWidth
	}
	
	var result strings.Builder
	borderColor := color.New(p.borderStyle)
	
	// 绘制顶部边框（圆角）
	if p.title != "" {
		titleLen := getDisplayWidth(p.title)
		if titleLen+6 > totalWidth {
			totalWidth = titleLen + 6 // 确保标题能完整显示
		}
		
		result.WriteString(borderColor.Sprint("╭─── "))
		result.WriteString(borderColor.Sprint(p.title))
		result.WriteString(borderColor.Sprint(" "))
		result.WriteString(borderColor.Sprint(strings.Repeat("─", totalWidth-titleLen-5)))
		result.WriteString(borderColor.Sprintln("╮"))
	} else {
		result.WriteString(borderColor.Sprint("╭"))
		result.WriteString(borderColor.Sprint(strings.Repeat("─", totalWidth)))
		result.WriteString(borderColor.Sprintln("╮"))
	}
	
	// 绘制顶部内边距
	for i := 0; i < p.padding[0]; i++ {
		result.WriteString(borderColor.Sprint("│"))
		result.WriteString(strings.Repeat(" ", totalWidth))
		result.WriteString(borderColor.Sprintln("│"))
	}
	
	// 绘制内容行
	for _, line := range lines {
		result.WriteString(borderColor.Sprint("│"))
		result.WriteString(strings.Repeat(" ", p.padding[1])) // 左内边距
		result.WriteString(line)
		// 右内边距（填充到总宽度）
		rightPadding := totalWidth - getDisplayWidth(line) - p.padding[1]
		if rightPadding > 0 {
			result.WriteString(strings.Repeat(" ", rightPadding))
		}
		result.WriteString(borderColor.Sprintln("│"))
	}
	
	// 绘制底部内边距
	for i := 0; i < p.padding[0]; i++ {
		result.WriteString(borderColor.Sprint("│"))
		result.WriteString(strings.Repeat(" ", totalWidth))
		result.WriteString(borderColor.Sprintln("│"))
	}
	
	// 绘制底部边框（圆角）
	result.WriteString(borderColor.Sprint("╰"))
	result.WriteString(borderColor.Sprint(strings.Repeat("─", totalWidth)))
	result.WriteString(borderColor.Sprintln("╯"))
	
	return result.String()
}

// RenderCompact 紧凑模式渲染
//
// 不添加垂直内边距的紧凑渲染模式
func (p *Panel) RenderCompact() {
	originalPadding := p.padding[0]
	p.padding[0] = 0
	p.Render()
	p.padding[0] = originalPadding
}

// SetContent 设置面板内容
func (p *Panel) SetContent(content string) {
	p.content = content
}

// SetTitle 设置面板标题
func (p *Panel) SetTitle(title string) {
	p.title = title
}

// GetContent 获取面板内容
func (p *Panel) GetContent() string {
	return p.content
}

// GetTitle 获取面板标题
func (p *Panel) GetTitle() string {
	return p.title
}

// CreateInfoPanel 创建信息面板的便捷函数
//
// 参数：
// - title: 面板标题
// - items: 信息项目（key-value对）
//
// 返回值：
// - *Panel: 配置好的信息面板
func CreateInfoPanel(title string, items map[string]string) *Panel {
	var lines []string
	
	for key, value := range items {
		// 格式化为 "Key: Value" 的形式
		line := fmt.Sprintf("%-15s %s", key+":", value)
		lines = append(lines, line)
	}
	
	content := strings.Join(lines, "\n")
	
	return NewPanel(content, title,
		WithBorderStyle(color.FgCyan),
		WithPanelPadding(1, 2))
}

// CreateMessagePanel 创建消息面板的便捷函数
//
// 参数：
// - message: 消息内容
// - messageType: 消息类型 ("info", "success", "warning", "error")
//
// 返回值：
// - *Panel: 配置好的消息面板
func CreateMessagePanel(message, messageType string) *Panel {
	var borderColor color.Attribute
	var title string
	
	switch messageType {
	case "success":
		borderColor = color.FgGreen
		title = "✓ Success"
	case "warning":
		borderColor = color.FgYellow
		title = "⚠ Warning"
	case "error":
		borderColor = color.FgRed
		title = "✗ Error"
	default:
		borderColor = color.FgBlue
		title = "ℹ Information"
	}
	
	return NewPanel(message, title,
		WithBorderStyle(borderColor),
		WithPanelPadding(1, 2))
}