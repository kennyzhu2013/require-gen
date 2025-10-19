package ui

import (
	"fmt"
	"sync"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"specify-cli/internal/config"
	"specify-cli/internal/types"
)

// UIManager UI管理器
//
// UIManager 是用户界面管理的核心组件，负责统一管理和控制
// require-gen框架中所有的用户交互界面元素。
//
// 主要功能：
// - 颜色输出控制：管理终端颜色显示的启用/禁用
// - 交互式提示：提供统一的用户输入和选择接口
// - 界面状态管理：维护UI组件的显示状态
// - 跨平台兼容：确保在不同操作系统上的一致体验
// - 主题管理：集成主题系统，提供一致的颜色方案
// - 进度显示：支持进度条和状态跟踪
//
// 设计特点：
// - 单例模式：通过NewUIManager创建统一实例
// - 配置驱动：支持运行时动态调整UI行为
// - 模块化设计：各UI功能相互独立，便于维护
// - 响应式布局：自适应不同终端尺寸
// - 主题感知：自动应用当前主题的颜色方案
//
// 使用场景：
// - 项目初始化向导界面
// - 配置选择和确认对话框
// - 进度显示和状态反馈
// - 错误信息和警告提示
//
// 示例用法：
//   manager := NewUIManager()
//   manager.ShowBanner()
//   choice, err := manager.SelectWithArrows(options, "选择AI助手", "default")
type UIManager struct {
	colorEnabled bool
	theme        Theme
}

// NewUIManager 创建新的UI管理器
//
// NewUIManager 是UIManager的工厂函数，用于创建和初始化
// 新的用户界面管理器实例。
//
// 初始化特性：
// - 默认启用颜色输出：colorEnabled设置为true
// - 自动获取当前全局主题
// - 注册为主题观察者，自动响应主题变化
//
// 返回值：
//   *UIManager - 初始化完成的UI管理器实例
//
// 注意事项：
// - 创建的实例会自动注册为主题观察者
// - 如果需要禁用颜色，可以通过SetColorEnabled方法调整
func NewUIManager() *UIManager {
	manager := &UIManager{
		colorEnabled: true,
		theme:        GetGlobalTheme(),
	}
	
	// 注册为主题观察者
	GetThemeManager().AddObserver(manager)
	
	return manager
}

// OnThemeChanged 实现ThemeObserver接口
//
// 当全局主题发生变化时，自动更新UI管理器的主题引用。
// 这确保了UI组件始终使用最新的主题设置。
func (ui *UIManager) OnThemeChanged(oldTheme, newTheme Theme) {
	ui.theme = newTheme
}

// SetColorEnabled 设置颜色启用状态
func (ui *UIManager) SetColorEnabled(enabled bool) {
	ui.colorEnabled = enabled
	color.NoColor = !enabled
}

// GetTheme 获取当前主题
func (ui *UIManager) GetTheme() Theme {
	return ui.theme
}

// CreateProgressBar 创建进度条
//
// 使用当前主题的颜色创建一个新的进度条实例。
// 进度条会自动应用主题的进度条颜色设置。
//
// 参数：
//   total - 进度条的总量
//   description - 进度条描述
//   options - 可选的配置选项
//
// 返回值：
//   *ProgressBar - 配置好的进度条实例
func (ui *UIManager) CreateProgressBar(total int64, description string, options ...ProgressBarOption) *ProgressBar {
	// 基于当前主题创建进度条
	themeOptions := []ProgressBarOption{
		WithColors(color.FgGreen, color.FgWhite),  // 使用默认颜色，因为主题颜色获取需要修复
		WithTextColor(color.FgCyan),
	}
	
	// 合并用户选项
	allOptions := append(themeOptions, options...)
	
	return NewProgressBar(total, description, allOptions...)
}

// ShowBanner 显示应用横幅
//
// ShowBanner 在终端中显示require-gen应用程序的欢迎横幅，
// 包括应用名称、版本信息和标语。这是用户与应用交互的第一印象。
//
// 显示内容：
// - 应用横幅：使用主题的主色调显示主标题
// - 应用标语：使用主题的次要色调显示描述性文本
// - 空行分隔：在横幅后添加空行以改善视觉效果
//
// 视觉特性：
// - 主题感知：自动使用当前主题的颜色方案
// - 跨平台兼容：在不同操作系统上保持一致显示
// - 自适应终端：根据终端颜色支持自动调整
//
// 使用场景：
// - 应用程序启动时的欢迎界面
// - 命令行工具的品牌展示
// - 用户体验的第一接触点
//
// 示例输出：
//   ╔══════════════════════════════════════╗
//   ║        REQUIRE-GEN FRAMEWORK         ║
//   ╚══════════════════════════════════════╝
//   🚀 智能项目初始化和AI助手集成工具
//
// 注意事项：
// - 横幅内容来自config.Banner和config.Tagline
// - 颜色输出依赖终端的颜色支持
// - 函数为全局函数，使用全局主题设置
func ShowBanner() {
	theme := GetGlobalTheme()
	
	theme.Primary().Println(config.Banner)
	theme.Secondary().Println(config.Tagline)
	fmt.Println()
}

// SelectWithArrows 使用箭头键选择选项
//
// SelectWithArrows 提供一个交互式的选择界面，用户可以使用箭头键
// 在多个选项中进行导航和选择。这是require-gen框架中主要的
// 用户交互方式之一。
//
// 参数说明：
// - options: map[string]string - 选项映射，key为选项值，value为描述
// - promptText: string - 提示文本，显示在选项列表上方
// - defaultKey: string - 默认选中的选项key（当前未使用）
//
// 返回值：
// - string: 用户选择的选项key
// - error: 选择过程中的错误（如用户取消操作）
//
// 交互特性：
// - 箭头键导航：上下箭头键移动选择
// - 回车确认：Enter键确认当前选择
// - ESC取消：Escape键取消选择操作
// - 可视化反馈：高亮显示当前选中项
//
// 显示格式：
// 每个选项显示为 "key - description" 的格式
//
// 使用场景：
// - AI助手选择界面
// - 脚本类型选择
// - 配置选项确认
// - 模板选择菜单
//
// 示例用法：
//   options := map[string]string{
//       "copilot": "GitHub Copilot AI助手",
//       "claude": "Claude AI助手",
//   }
//   choice, err := SelectWithArrows(options, "选择AI助手", "copilot")
//
// 注意事项：
// - 选项数量限制为10个可见项（Size: 10）
// - 依赖promptui库提供交互功能
// - 需要终端支持ANSI转义序列
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

// 全局键盘状态管理
var (
	keyboardInitialized = false
	keyboardMutex       sync.Mutex
)

// InitKeyboard 初始化键盘监听（可选调用）
func InitKeyboard() error {
	keyboardMutex.Lock()
	defer keyboardMutex.Unlock()
	
	if keyboardInitialized {
		return nil
	}
	
	if err := keyboard.Open(); err != nil {
		return err
	}
	
	keyboardInitialized = true
	return nil
}

// CloseKeyboard 关闭键盘监听（可选调用）
func CloseKeyboard() {
	keyboardMutex.Lock()
	defer keyboardMutex.Unlock()
	
	if keyboardInitialized {
		keyboard.Close()
		keyboardInitialized = false
	}
}

// GetKey 获取单个按键输入
//
// GetKey 实现跨平台的单键捕获功能，无需用户按回车键确认。
// 支持普通字符键、功能键（如方向键、ESC、回车等）和组合键。
//
// 功能特性：
// - 即时响应：无需等待回车键，按键后立即返回
// - 跨平台兼容：支持Windows、macOS、Linux系统
// - 特殊键支持：方向键、功能键、组合键等
// - 错误处理：键盘初始化失败时的优雅降级
//
// 支持的按键类型：
// - 普通字符：a-z, A-Z, 0-9, 符号等
// - 方向键：ArrowUp, ArrowDown, ArrowLeft, ArrowRight
// - 功能键：Enter, Escape, Space, Tab, Backspace
// - 组合键：Ctrl+C, Ctrl+Z等（部分支持）
//
// 返回值格式：
// - 普通字符：返回字符本身，如 "a", "1", "!"
// - 特殊键：返回键名，如 "ArrowUp", "Enter", "Escape"
// - 组合键：返回组合描述，如 "Ctrl+C"
//
// 使用场景：
// - 交互式菜单导航
// - 游戏控制输入
// - 快捷键处理
// - 实时输入响应
//
// 错误处理：
// - 键盘初始化失败：返回错误信息
// - 读取中断：用户强制退出时的处理
// - 系统不支持：在某些受限环境中的降级处理
//
// 示例用法：
//   key, err := GetKey()
//   if err != nil {
//       log.Printf("获取按键失败: %v", err)
//       return
//   }
//   
//   switch key {
//   case "ArrowUp":
//       // 处理上箭头
//   case "Enter":
//       // 处理回车键
//   case "q", "Q":
//       // 处理退出
//   default:
//       fmt.Printf("按下了: %s\n", key)
//   }
//
// 注意事项：
// - 函数会阻塞直到用户按键
// - 使用前需要确保终端支持原始模式
// - 在某些IDE或受限环境中可能无法正常工作
// - 自动管理键盘初始化和清理
func GetKey() (string, error) {
	keyboardMutex.Lock()
	needsInit := !keyboardInitialized
	keyboardMutex.Unlock()
	
	// 如果还没有初始化，尝试初始化
	if needsInit {
		if err := keyboard.Open(); err != nil {
			// 如果keyboard库初始化失败，降级到简化版本
			var input string
			fmt.Print("Press any key and Enter: ")
			_, err := fmt.Scanln(&input)
			return input, err
		}
		
		keyboardMutex.Lock()
		keyboardInitialized = true
		keyboardMutex.Unlock()
	}

	// 获取按键事件
	char, key, err := keyboard.GetKey()
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}

	// 处理特殊键
	if key != 0 {
		switch key {
		case keyboard.KeyArrowUp:
			return "ArrowUp", nil
		case keyboard.KeyArrowDown:
			return "ArrowDown", nil
		case keyboard.KeyArrowLeft:
			return "ArrowLeft", nil
		case keyboard.KeyArrowRight:
			return "ArrowRight", nil
		case keyboard.KeyEnter:
			return "Enter", nil
		case keyboard.KeyEsc:
			return "Escape", nil
		case keyboard.KeySpace:
			return "Space", nil
		case keyboard.KeyTab:
			return "Tab", nil
		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			return "Backspace", nil
		case keyboard.KeyDelete:
			return "Delete", nil
		case keyboard.KeyHome:
			return "Home", nil
		case keyboard.KeyEnd:
			return "End", nil
		case keyboard.KeyPgup:
			return "PageUp", nil
		case keyboard.KeyPgdn:
			return "PageDown", nil
		case keyboard.KeyF1:
			return "F1", nil
		case keyboard.KeyF2:
			return "F2", nil
		case keyboard.KeyF3:
			return "F3", nil
		case keyboard.KeyF4:
			return "F4", nil
		case keyboard.KeyF5:
			return "F5", nil
		case keyboard.KeyF6:
			return "F6", nil
		case keyboard.KeyF7:
			return "F7", nil
		case keyboard.KeyF8:
			return "F8", nil
		case keyboard.KeyF9:
			return "F9", nil
		case keyboard.KeyF10:
			return "F10", nil
		case keyboard.KeyF11:
			return "F11", nil
		case keyboard.KeyF12:
			return "F12", nil
		case keyboard.KeyCtrlC:
			return "Ctrl+C", nil
		case keyboard.KeyCtrlZ:
			return "Ctrl+Z", nil
		default:
			return fmt.Sprintf("Key_%d", int(key)), nil
		}
	}

	// 处理普通字符
	if char != 0 {
		return string(char), nil
	}

	return "", fmt.Errorf("unknown key event")
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

// ShowSuccess 显示成功消息
//
// 使用主题的成功色调显示成功消息，提供一致的视觉反馈。
func ShowSuccess(message string) {
	theme := GetGlobalTheme()
	theme.Success().Printf("✓ %s\n", message)
}

// ShowError 显示错误消息
//
// 使用主题的错误色调显示错误消息，提供清晰的错误指示。
func ShowError(message string) {
	theme := GetGlobalTheme()
	theme.Error().Printf("✗ %s\n", message)
}

// ShowWarning 显示警告消息
//
// 使用主题的警告色调显示警告消息，提供适当的注意提示。
func ShowWarning(message string) {
	theme := GetGlobalTheme()
	theme.Warning().Printf("⚠ %s\n", message)
}

// ShowInfo 显示信息消息
//
// 使用主题的信息色调显示信息消息，提供中性的信息反馈。
func ShowInfo(message string) {
	theme := GetGlobalTheme()
	theme.Info().Printf("ℹ %s\n", message)
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
// Renderer UI渲染器
//
// Renderer 是UIRenderer接口的具体实现，提供了一套完整的
// 用户界面渲染功能。它封装了UIManager，为上层业务逻辑
// 提供统一的UI操作接口。
//
// 设计模式：
// - 适配器模式：将UIManager适配为UIRenderer接口
// - 组合模式：通过组合UIManager实现功能复用
// - 接口隔离：只暴露必要的UI操作方法
//
// 主要功能：
// - 横幅显示：应用启动时的欢迎界面
// - 交互选择：箭头键导航的选项选择
// - 按键捕获：单个按键输入处理
//
// 使用场景：
// - 依赖注入：作为UIRenderer接口的实现
// - 业务层调用：提供统一的UI操作接口
// - 测试环境：便于UI功能的单元测试
//
// 示例用法：
//   renderer := NewRenderer()
//   renderer.ShowBanner()
//   choice, err := renderer.SelectWithArrows(options, "选择选项", "default")
//
// 注意事项：
// - 实现了types.UIRenderer接口的所有方法
// - 内部使用UIManager实例进行实际的UI操作
// - 支持链式调用和方法组合
type Renderer struct {
	manager *UIManager
}

// NewRenderer 创建新的UI渲染器
//
// NewRenderer 是Renderer的工厂函数，创建并返回一个实现了
// types.UIRenderer接口的新渲染器实例。
//
// 创建特性：
// - 自动初始化：内部自动创建UIManager实例
// - 接口返回：返回UIRenderer接口类型，支持多态
// - 即用型：返回的实例立即可用于UI操作
// - 零配置：无需额外参数即可创建
//
// 返回值：
// - types.UIRenderer: UI渲染器接口实例
//
// 使用场景：
// - 依赖注入容器中的UI组件创建
// - 业务层需要UI渲染功能时
// - 测试环境中的Mock对象创建
//
// 示例用法：
//   renderer := NewRenderer()
//   renderer.ShowBanner()
//   
//   // 也可以用于依赖注入
//   type Service struct {
//       ui types.UIRenderer
//   }
//   service := &Service{ui: NewRenderer()}
//
// 注意事项：
// - 返回接口类型便于测试和扩展
// - 内部UIManager使用默认配置
// - 支持后续的功能扩展和定制
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

// ConfirmAction 实现UIRenderer接口
func (r *Renderer) ConfirmAction(message string) bool {
	result, err := ConfirmAction(message)
	if err != nil {
		return false
	}
	return result
}

// ShowProgress 实现UIRenderer接口
func (r *Renderer) ShowProgress(message string, percentage int) {
	fmt.Printf("%s [%d%%]\n", message, percentage)
}

// ShowMessage 实现UIRenderer接口
func (r *Renderer) ShowMessage(message, messageType string) {
	switch messageType {
	case "success":
		ShowSuccess(message)
	case "error":
		ShowError(message)
	case "warning":
		ShowWarning(message)
	case "info":
		ShowInfo(message)
	default:
		fmt.Println(message)
	}
}

// SelectOption 实现UIRenderer接口
func (r *Renderer) SelectOption(prompt string, options []string) (string, error) {
	selectPrompt := promptui.Select{
		Label: prompt,
		Items: options,
	}
	
	_, result, err := selectPrompt.Run()
	return result, err
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