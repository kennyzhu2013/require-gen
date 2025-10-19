package ui

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
)

// Theme 主题接口
//
// Theme 定义了UI主题的基本接口，提供统一的颜色管理功能。
// 类似于Python Rich库的主题系统，支持多种预定义主题和自定义主题。
//
// 主要功能：
// - 统一的颜色定义和管理
// - 多种预定义主题（暗色、亮色、高对比度等）
// - 支持自定义主题创建
// - 动态主题切换
// - 颜色一致性保证
//
// 设计原则：
// - 单一职责：专注于颜色和样式管理
// - 开放封闭：易于扩展新主题，不修改现有代码
// - 依赖倒置：通过接口而非具体实现依赖
type Theme interface {
	// 基础颜色
	Primary() *color.Color     // 主色调
	Secondary() *color.Color   // 次要色调
	Success() *color.Color     // 成功状态色
	Warning() *color.Color     // 警告状态色
	Error() *color.Color       // 错误状态色
	Info() *color.Color        // 信息状态色
	
	// 文本颜色
	Text() *color.Color        // 主要文本色
	TextSecondary() *color.Color // 次要文本色
	TextMuted() *color.Color   // 弱化文本色
	
	// 背景颜色
	Background() *color.Color  // 主背景色
	BackgroundAlt() *color.Color // 替代背景色
	
	// 边框和分隔符
	Border() *color.Color      // 边框色
	Separator() *color.Color   // 分隔符色
	
	// 进度条颜色
	ProgressFill() *color.Color   // 进度条填充色
	ProgressEmpty() *color.Color  // 进度条空白色
	
	// 主题信息
	Name() string              // 主题名称
	Description() string       // 主题描述
	IsDark() bool             // 是否为暗色主题
}

// BaseTheme 基础主题结构
//
// BaseTheme 提供了Theme接口的基础实现，包含所有必要的颜色定义。
// 其他主题可以通过组合BaseTheme来快速创建。
type BaseTheme struct {
	name        string
	description string
	isDark      bool
	
	// 颜色定义
	primary       *color.Color
	secondary     *color.Color
	success       *color.Color
	warning       *color.Color
	error         *color.Color
	info          *color.Color
	text          *color.Color
	textSecondary *color.Color
	textMuted     *color.Color
	background    *color.Color
	backgroundAlt *color.Color
	border        *color.Color
	separator     *color.Color
	progressFill  *color.Color
	progressEmpty *color.Color
}

// 实现Theme接口
func (t *BaseTheme) Primary() *color.Color     { return t.primary }
func (t *BaseTheme) Secondary() *color.Color   { return t.secondary }
func (t *BaseTheme) Success() *color.Color     { return t.success }
func (t *BaseTheme) Warning() *color.Color     { return t.warning }
func (t *BaseTheme) Error() *color.Color       { return t.error }
func (t *BaseTheme) Info() *color.Color        { return t.info }
func (t *BaseTheme) Text() *color.Color        { return t.text }
func (t *BaseTheme) TextSecondary() *color.Color { return t.textSecondary }
func (t *BaseTheme) TextMuted() *color.Color   { return t.textMuted }
func (t *BaseTheme) Background() *color.Color  { return t.background }
func (t *BaseTheme) BackgroundAlt() *color.Color { return t.backgroundAlt }
func (t *BaseTheme) Border() *color.Color      { return t.border }
func (t *BaseTheme) Separator() *color.Color   { return t.separator }
func (t *BaseTheme) ProgressFill() *color.Color   { return t.progressFill }
func (t *BaseTheme) ProgressEmpty() *color.Color  { return t.progressEmpty }
func (t *BaseTheme) Name() string              { return t.name }
func (t *BaseTheme) Description() string       { return t.description }
func (t *BaseTheme) IsDark() bool             { return t.isDark }

// ThemeManager 主题管理器
//
// ThemeManager 负责管理所有可用的主题，提供主题注册、切换和获取功能。
// 支持线程安全的主题操作和全局主题状态管理。
//
// 功能特性：
// - 线程安全的主题管理
// - 动态主题注册和注销
// - 全局主题状态管理
// - 主题切换事件通知
// - 主题验证和错误处理
type ThemeManager struct {
	themes      map[string]Theme
	currentTheme Theme
	mutex       sync.RWMutex
	observers   []ThemeObserver
}

// ThemeObserver 主题观察者接口
//
// ThemeObserver 定义了主题变化时的回调接口，
// 允许UI组件在主题切换时自动更新样式。
type ThemeObserver interface {
	OnThemeChanged(oldTheme, newTheme Theme)
}

// 全局主题管理器实例
var (
	globalThemeManager *ThemeManager
	once              sync.Once
)

// GetThemeManager 获取全局主题管理器
//
// 使用单例模式确保全局只有一个主题管理器实例。
// 首次调用时会初始化管理器并注册所有预定义主题。
func GetThemeManager() *ThemeManager {
	once.Do(func() {
		globalThemeManager = &ThemeManager{
			themes:    make(map[string]Theme),
			observers: make([]ThemeObserver, 0),
		}
		
		// 注册预定义主题
		globalThemeManager.registerBuiltinThemes()
		
		// 设置默认主题
		globalThemeManager.SetTheme("default")
	})
	return globalThemeManager
}

// RegisterTheme 注册主题
//
// 将新主题注册到主题管理器中，使其可以被使用。
// 如果主题名称已存在，将覆盖原有主题。
//
// 参数：
//   theme - 要注册的主题实例
//
// 返回值：
//   error - 注册过程中的错误，如果成功则为nil
func (tm *ThemeManager) RegisterTheme(theme Theme) error {
	if theme == nil {
		return fmt.Errorf("theme cannot be nil")
	}
	
	if theme.Name() == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.themes[theme.Name()] = theme
	return nil
}

// SetTheme 设置当前主题
//
// 根据主题名称切换到指定主题。
// 如果主题不存在，返回错误。
// 成功切换后会通知所有观察者。
//
// 参数：
//   name - 主题名称
//
// 返回值：
//   error - 切换过程中的错误，如果成功则为nil
func (tm *ThemeManager) SetTheme(name string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	theme, exists := tm.themes[name]
	if !exists {
		return fmt.Errorf("theme '%s' not found", name)
	}
	
	oldTheme := tm.currentTheme
	tm.currentTheme = theme
	
	// 通知观察者
	for _, observer := range tm.observers {
		observer.OnThemeChanged(oldTheme, theme)
	}
	
	return nil
}

// GetCurrentTheme 获取当前主题
func (tm *ThemeManager) GetCurrentTheme() Theme {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.currentTheme
}

// GetTheme 根据名称获取主题
func (tm *ThemeManager) GetTheme(name string) (Theme, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	theme, exists := tm.themes[name]
	if !exists {
		return nil, fmt.Errorf("theme '%s' not found", name)
	}
	
	return theme, nil
}

// ListThemes 列出所有可用主题
func (tm *ThemeManager) ListThemes() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	names := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		names = append(names, name)
	}
	
	return names
}

// AddObserver 添加主题观察者
func (tm *ThemeManager) AddObserver(observer ThemeObserver) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.observers = append(tm.observers, observer)
}

// RemoveObserver 移除主题观察者
func (tm *ThemeManager) RemoveObserver(observer ThemeObserver) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	for i, obs := range tm.observers {
		if obs == observer {
			tm.observers = append(tm.observers[:i], tm.observers[i+1:]...)
			break
		}
	}
}

// registerBuiltinThemes 注册内置主题
func (tm *ThemeManager) registerBuiltinThemes() {
	// 注册默认主题
	tm.RegisterTheme(NewDefaultTheme())
	
	// 注册暗色主题
	tm.RegisterTheme(NewDarkTheme())
	
	// 注册亮色主题
	tm.RegisterTheme(NewLightTheme())
	
	// 注册高对比度主题
	tm.RegisterTheme(NewHighContrastTheme())
	
	// 注册彩色主题
	tm.RegisterTheme(NewColorfulTheme())
	
	// 注册简约主题
	tm.RegisterTheme(NewMinimalTheme())
}

// 预定义主题构造函数

// NewDefaultTheme 创建默认主题
//
// 默认主题提供平衡的颜色搭配，适合大多数使用场景。
// 使用中等对比度和舒适的颜色组合。
func NewDefaultTheme() Theme {
	return &BaseTheme{
		name:        "default",
		description: "默认主题，平衡的颜色搭配",
		isDark:      false,
		
		primary:       color.New(color.FgBlue, color.Bold),
		secondary:     color.New(color.FgCyan),
		success:       color.New(color.FgGreen, color.Bold),
		warning:       color.New(color.FgYellow, color.Bold),
		error:         color.New(color.FgRed, color.Bold),
		info:          color.New(color.FgCyan),
		text:          color.New(color.FgWhite),
		textSecondary: color.New(color.FgWhite, color.Faint),
		textMuted:     color.New(color.FgBlack, color.Faint),
		background:    color.New(color.BgBlack),
		backgroundAlt: color.New(color.BgBlack, color.Faint),
		border:        color.New(color.FgWhite, color.Faint),
		separator:     color.New(color.FgBlack, color.Bold),
		progressFill:  color.New(color.FgGreen, color.Bold),
		progressEmpty: color.New(color.FgWhite, color.Faint),
	}
}

// NewDarkTheme 创建暗色主题
//
// 暗色主题适合在低光环境下使用，减少眼部疲劳。
// 使用深色背景和亮色文本。
func NewDarkTheme() Theme {
	return &BaseTheme{
		name:        "dark",
		description: "暗色主题，适合低光环境",
		isDark:      true,
		
		primary:       color.New(color.FgBlue, color.Bold),
		secondary:     color.New(color.FgMagenta),
		success:       color.New(color.FgGreen),
		warning:       color.New(color.FgYellow),
		error:         color.New(color.FgRed),
		info:          color.New(color.FgCyan),
		text:          color.New(color.FgWhite),
		textSecondary: color.New(color.FgWhite, color.Faint),
		textMuted:     color.New(color.FgBlack, color.Bold),
		background:    color.New(color.BgBlack),
		backgroundAlt: color.New(color.BgBlack, color.Faint),
		border:        color.New(color.FgBlack, color.Bold),
		separator:     color.New(color.FgBlack, color.Bold),
		progressFill:  color.New(color.FgBlue, color.Bold),
		progressEmpty: color.New(color.FgBlack, color.Bold),
	}
}

// NewLightTheme 创建亮色主题
//
// 亮色主题适合在明亮环境下使用，提供清晰的视觉效果。
// 使用浅色背景和深色文本。
func NewLightTheme() Theme {
	return &BaseTheme{
		name:        "light",
		description: "亮色主题，适合明亮环境",
		isDark:      false,
		
		primary:       color.New(color.FgBlue, color.Bold),
		secondary:     color.New(color.FgMagenta),
		success:       color.New(color.FgGreen, color.Bold),
		warning:       color.New(color.FgYellow, color.Bold),
		error:         color.New(color.FgRed, color.Bold),
		info:          color.New(color.FgBlue),
		text:          color.New(color.FgBlack, color.Bold),
		textSecondary: color.New(color.FgBlack),
		textMuted:     color.New(color.FgBlack, color.Faint),
		background:    color.New(color.BgWhite),
		backgroundAlt: color.New(color.BgWhite, color.Faint),
		border:        color.New(color.FgBlack, color.Faint),
		separator:     color.New(color.FgBlack, color.Faint),
		progressFill:  color.New(color.FgGreen, color.Bold),
		progressEmpty: color.New(color.FgBlack, color.Faint),
	}
}

// NewHighContrastTheme 创建高对比度主题
//
// 高对比度主题提供最大的视觉对比度，适合视觉障碍用户。
// 使用强烈的颜色对比来确保可读性。
func NewHighContrastTheme() Theme {
	return &BaseTheme{
		name:        "high-contrast",
		description: "高对比度主题，适合视觉障碍用户",
		isDark:      true,
		
		primary:       color.New(color.FgWhite, color.Bold, color.BgBlue),
		secondary:     color.New(color.FgBlack, color.Bold, color.BgWhite),
		success:       color.New(color.FgWhite, color.Bold, color.BgGreen),
		warning:       color.New(color.FgBlack, color.Bold, color.BgYellow),
		error:         color.New(color.FgWhite, color.Bold, color.BgRed),
		info:          color.New(color.FgWhite, color.Bold, color.BgCyan),
		text:          color.New(color.FgWhite, color.Bold),
		textSecondary: color.New(color.FgWhite),
		textMuted:     color.New(color.FgWhite, color.Faint),
		background:    color.New(color.BgBlack),
		backgroundAlt: color.New(color.BgBlack),
		border:        color.New(color.FgWhite, color.Bold),
		separator:     color.New(color.FgWhite, color.Bold),
		progressFill:  color.New(color.FgBlack, color.Bold, color.BgWhite),
		progressEmpty: color.New(color.FgWhite, color.Bold),
	}
}

// NewColorfulTheme 创建彩色主题
//
// 彩色主题使用丰富的颜色搭配，提供生动的视觉体验。
// 适合喜欢鲜艳颜色的用户。
func NewColorfulTheme() Theme {
	return &BaseTheme{
		name:        "colorful",
		description: "彩色主题，丰富的颜色搭配",
		isDark:      false,
		
		primary:       color.New(color.FgMagenta, color.Bold),
		secondary:     color.New(color.FgCyan, color.Bold),
		success:       color.New(color.FgGreen, color.Bold),
		warning:       color.New(color.FgYellow, color.Bold),
		error:         color.New(color.FgRed, color.Bold),
		info:          color.New(color.FgBlue, color.Bold),
		text:          color.New(color.FgWhite, color.Bold),
		textSecondary: color.New(color.FgCyan),
		textMuted:     color.New(color.FgMagenta, color.Faint),
		background:    color.New(color.BgBlack),
		backgroundAlt: color.New(color.BgBlack, color.Faint),
		border:        color.New(color.FgMagenta),
		separator:     color.New(color.FgCyan),
		progressFill:  color.New(color.FgMagenta, color.Bold),
		progressEmpty: color.New(color.FgCyan, color.Faint),
	}
}

// NewMinimalTheme 创建简约主题
//
// 简约主题使用最少的颜色，提供简洁的视觉效果。
// 适合喜欢简洁界面的用户。
func NewMinimalTheme() Theme {
	return &BaseTheme{
		name:        "minimal",
		description: "简约主题，简洁的视觉效果",
		isDark:      false,
		
		primary:       color.New(color.FgWhite, color.Bold),
		secondary:     color.New(color.FgWhite),
		success:       color.New(color.FgWhite, color.Bold),
		warning:       color.New(color.FgWhite, color.Bold),
		error:         color.New(color.FgWhite, color.Bold),
		info:          color.New(color.FgWhite),
		text:          color.New(color.FgWhite),
		textSecondary: color.New(color.FgWhite, color.Faint),
		textMuted:     color.New(color.FgBlack, color.Bold),
		background:    color.New(color.BgBlack),
		backgroundAlt: color.New(color.BgBlack),
		border:        color.New(color.FgWhite, color.Faint),
		separator:     color.New(color.FgBlack, color.Bold),
		progressFill:  color.New(color.FgWhite, color.Bold),
		progressEmpty: color.New(color.FgBlack, color.Bold),
	}
}

// 便捷函数

// SetGlobalTheme 设置全局主题
//
// 这是一个便捷函数，用于快速设置全局主题。
// 等同于 GetThemeManager().SetTheme(name)
func SetGlobalTheme(name string) error {
	return GetThemeManager().SetTheme(name)
}

// GetGlobalTheme 获取当前全局主题
//
// 这是一个便捷函数，用于快速获取当前全局主题。
// 等同于 GetThemeManager().GetCurrentTheme()
func GetGlobalTheme() Theme {
	return GetThemeManager().GetCurrentTheme()
}

// RegisterGlobalTheme 注册全局主题
//
// 这是一个便捷函数，用于快速注册主题到全局管理器。
// 等同于 GetThemeManager().RegisterTheme(theme)
func RegisterGlobalTheme(theme Theme) error {
	return GetThemeManager().RegisterTheme(theme)
}

// ListGlobalThemes 列出所有全局主题
//
// 这是一个便捷函数，用于快速列出所有可用主题。
// 等同于 GetThemeManager().ListThemes()
func ListGlobalThemes() []string {
	return GetThemeManager().ListThemes()
}