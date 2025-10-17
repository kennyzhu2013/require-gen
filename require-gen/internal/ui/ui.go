package ui

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"specify-cli/internal/config"
	"specify-cli/internal/types"
)

// UIManager UI管理器
type UIManager struct {
	colorEnabled bool
}

// NewUIManager 创建新的UI管理器
func NewUIManager() *UIManager {
	return &UIManager{
		colorEnabled: true,
	}
}

// ShowBanner 显示应用横幅
func ShowBanner() {
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow)
	
	cyan.Println(config.Banner)
	yellow.Println(config.Tagline)
	fmt.Println()
}

// SelectWithArrows 使用箭头键选择选项
func SelectWithArrows(options map[string]string, promptText, defaultKey string) (string, error) {
	// 构建选项列表
	var items []string
	var keys []string
	
	for key, desc := range options {
		items = append(items, fmt.Sprintf("%s - %s", key, desc))
		keys = append(keys, key)
	}

	// 创建选择提示
	prompt := promptui.Select{
		Label:     promptText,
		Items:     items,
		Size:      10,
		Templates: getSelectTemplates(),
	}

	// 执行选择
	index, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("selection cancelled: %w", err)
	}

	return keys[index], nil
}

// GetKey 获取单个按键输入
func GetKey() (string, error) {
	// 这里应该实现获取单个按键的逻辑
	// 由于Go标准库限制，这里提供一个简化版本
	var input string
	fmt.Print("Press any key and Enter: ")
	_, err := fmt.Scanln(&input)
	return input, err
}

// ConfirmAction 确认操作
func ConfirmAction(message string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return false, err
	}

	return result == "y" || result == "Y", nil
}

// InputText 文本输入
func InputText(label, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}

	return prompt.Run()
}

// ShowProgress 显示进度信息
func ShowProgress(message string) {
	blue := color.New(color.FgBlue)
	blue.Printf("⏳ %s\n", message)
}

// ShowSuccess 显示成功信息
func ShowSuccess(message string) {
	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✅ %s\n", message)
}

// ShowError 显示错误信息
func ShowError(message string) {
	red := color.New(color.FgRed, color.Bold)
	red.Printf("❌ %s\n", message)
}

// ShowWarning 显示警告信息
func ShowWarning(message string) {
	yellow := color.New(color.FgYellow, color.Bold)
	yellow.Printf("⚠️  %s\n", message)
}

// ShowInfo 显示信息
func ShowInfo(message string) {
	cyan := color.New(color.FgCyan)
	cyan.Printf("ℹ️  %s\n", message)
}

// getSelectTemplates 获取选择模板
func getSelectTemplates() *promptui.SelectTemplates {
	return &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "▶ {{ .| cyan | bold }}",
		Inactive: "  {{ . | faint }}",
		Selected: "{{ \"✓\" | green | bold }} {{ . | bold }}",
	}
}

// Renderer UI渲染器实现
type Renderer struct {
	manager *UIManager
}

// NewRenderer 创建新的渲染器
func NewRenderer() types.UIRenderer {
	return &Renderer{
		manager: NewUIManager(),
	}
}

// ShowBanner 实现UIRenderer接口
func (r *Renderer) ShowBanner() {
	ShowBanner()
}

// SelectWithArrows 实现UIRenderer接口
func (r *Renderer) SelectWithArrows(options map[string]string, prompt, defaultKey string) (string, error) {
	return SelectWithArrows(options, prompt, defaultKey)
}

// GetKey 实现UIRenderer接口
func (r *Renderer) GetKey() (string, error) {
	return GetKey()
}

// PrintTable 打印表格
func PrintTable(headers []string, rows [][]string) {
	// 计算列宽
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// 打印表头
	cyan := color.New(color.FgCyan, color.Bold)
	for i, header := range headers {
		cyan.Printf("%-*s", colWidths[i]+2, header)
	}
	fmt.Println()

	// 打印分隔线
	for i := range headers {
		fmt.Print(fmt.Sprintf("%-*s", colWidths[i]+2, ""))
	}
	fmt.Println()

	// 打印数据行
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				fmt.Printf("%-*s", colWidths[i]+2, cell)
			}
		}
		fmt.Println()
	}
}

// ClearScreen 清屏
func ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

// MoveCursor 移动光标
func MoveCursor(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

// HideCursor 隐藏光标
func HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor 显示光标
func ShowCursor() {
	fmt.Print("\033[?25h")
}

// SetTitle 设置终端标题
func SetTitle(title string) {
	fmt.Printf("\033]0;%s\007", title)
}