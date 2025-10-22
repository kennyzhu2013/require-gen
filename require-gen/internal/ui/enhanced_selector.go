package ui

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// EnhancedSelector 增强的交互选择器
type EnhancedSelector struct {
	options     map[string]string
	title       string
	defaultKey  string
	selected    int
	keys        []string
	theme       Theme  // 改为接口类型而不是指针
}

// NewEnhancedSelector 创建新的增强选择器
func NewEnhancedSelector(options map[string]string, prompt, defaultKey string) *EnhancedSelector {
	keys := make([]string, 0, len(options))
	for k := range options {
		keys = append(keys, k)
	}
	
	selectedIdx := 0
	if defaultKey != "" {
		for i, k := range keys {
			if k == defaultKey {
				selectedIdx = i
				break
			}
		}
	}
	
	return &EnhancedSelector{
		options:    options,
		title:      prompt,
		defaultKey: defaultKey,
		selected:   selectedIdx,
		keys:       keys,
		theme:      GetGlobalTheme(),  // 直接使用接口
	}
}

// getTerminalSize 获取终端尺寸
func getTerminalSize() (int, int) {
	if width, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return width, height
	}
	return 80, 24 // 默认尺寸
}

// clearScreen 清屏并移动到顶部
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// Render 渲染选择界面
// Render 渲染选择器界面
func (s *EnhancedSelector) Render() {
	clearScreen()
	
	// 创建选择面板
	panel := NewPanel(s.createContent(), s.title, 
		WithBorderStyle(color.FgCyan),
		WithPadding(1, 2))
	panel.Render()
}

// createContent 创建选择内容
func (s *EnhancedSelector) createContent() string {
	var lines []string
	
	for i, key := range s.keys {
		desc := s.options[key]
		if i == s.selected {
			// 高亮选中项
			line := fmt.Sprintf("▶ %s (%s)", 
				s.theme.Primary().Sprint(key), 
				s.theme.TextMuted().Sprint(desc))
			lines = append(lines, line)
		} else {
			line := fmt.Sprintf("  %s (%s)", 
				s.theme.Primary().Sprint(key), 
				s.theme.TextMuted().Sprint(desc))
			lines = append(lines, line)
		}
	}
	
	lines = append(lines, "")
	lines = append(lines, s.theme.TextMuted().Sprint("Use ↑/↓ to navigate, Enter to select, Esc to cancel"))
	
	return strings.Join(lines, "\n")
}

// getKey 获取按键输入
func (s *EnhancedSelector) getKey() (string, error) {
	// 设置终端为原始模式
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	
	// 读取按键
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return "", err
	}
	
	// 解析按键
	switch {
	case n == 1:
		switch buf[0] {
		case 13: // Enter
			return "enter", nil
		case 27: // Escape
			return "escape", nil
		case 3: // Ctrl+C
			return "ctrl_c", nil
		}
	case n == 3 && buf[0] == 27 && buf[1] == 91:
		switch buf[2] {
		case 65: // Up arrow
			return "up", nil
		case 66: // Down arrow
			return "down", nil
		}
	}
	
	return "", nil
}

// Run 运行选择器
func (s *EnhancedSelector) Run() (string, error) {
	// 初始渲染
	s.Render()
	
	// 键盘事件循环
	for {
		key, err := s.getKey()
		if err != nil {
			return "", err
		}
		
		switch key {
		case "up":
			s.selected = (s.selected - 1 + len(s.keys)) % len(s.keys)
			s.Render()
		case "down":
			s.selected = (s.selected + 1) % len(s.keys)
			s.Render()
		case "enter":
			// 清屏并显示选择结果
			clearScreen()
			selectedKey := s.keys[s.selected]
			s.theme.Success().Printf("✓ Selected: %s\n", selectedKey)
			return selectedKey, nil
		case "escape", "ctrl_c":
			clearScreen()
			s.theme.Warning().Println("Selection cancelled")
			return "", fmt.Errorf("selection cancelled")
		}
	}
}

// SelectWithEnhancedUI 使用增强UI进行选择
func SelectWithEnhancedUI(options map[string]string, prompt, defaultKey string) (string, error) {
	selector := NewEnhancedSelector(options, prompt, defaultKey)
	return selector.Run()
}

// Windows特定的键盘输入处理
var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// enableVirtualTerminalProcessing 启用Windows虚拟终端处理
func enableVirtualTerminalProcessing() error {
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	
	ret, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return err
	}
	
	mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	ret, _, err = procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if ret == 0 {
		return err
	}
	
	return nil
}

// init 初始化函数，启用Windows虚拟终端处理
func init() {
	// 在Windows上启用虚拟终端处理
	if os.Getenv("OS") == "Windows_NT" {
		enableVirtualTerminalProcessing()
	}
}