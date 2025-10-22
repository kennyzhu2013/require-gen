package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// SpinnerStyle Spinner样式枚举
type SpinnerStyle int

const (
	SpinnerDots SpinnerStyle = iota
	SpinnerLine
	SpinnerCircle
	SpinnerArrow
	SpinnerBounce
	SpinnerClock
	SpinnerMoon
	SpinnerStar
)

// Spinner 加载动画组件
//
// Spinner提供了多种样式的加载动画，类似于Python Rich库的Spinner组件。
// 支持自定义文本、颜色、速度等配置。
//
// 特性：
// - 多种动画样式
// - 可配置的动画速度
// - 自定义文本和颜色
// - 线程安全的启动/停止
// - 自动清理和恢复
type Spinner struct {
	style       SpinnerStyle  // 动画样式
	text        string        // 显示文本
	color       *color.Color  // 颜色
	isActive    bool          // 是否活跃
	mutex       sync.RWMutex  // 读写锁
	stopChan    chan bool     // 停止信号
	frames      []string      // 动画帧
	frameIndex  int           // 当前帧索引
	speed       time.Duration // 动画速度
	prefix      string        // 前缀
	suffix      string        // 后缀
}

// SpinnerOption Spinner配置选项函数类型
type SpinnerOption func(*Spinner)

// WithSpinnerText 设置显示文本
func WithSpinnerText(text string) SpinnerOption {
	return func(s *Spinner) {
		s.text = text
	}
}

// WithSpinnerColor 设置颜色
func WithSpinnerColor(color *color.Color) SpinnerOption {
	return func(s *Spinner) {
		s.color = color
	}
}

// WithSpinnerSpeed 设置动画速度
func WithSpinnerSpeed(speed time.Duration) SpinnerOption {
	return func(s *Spinner) {
		s.speed = speed
	}
}

// WithSpinnerPrefix 设置前缀
func WithSpinnerPrefix(prefix string) SpinnerOption {
	return func(s *Spinner) {
		s.prefix = prefix
	}
}

// WithSpinnerSuffix 设置后缀
func WithSpinnerSuffix(suffix string) SpinnerOption {
	return func(s *Spinner) {
		s.suffix = suffix
	}
}

// NewSpinner 创建新的Spinner实例
//
// 参数：
// - style: 动画样式
// - options: 配置选项
//
// 返回值：
// - *Spinner: Spinner实例
func NewSpinner(style SpinnerStyle, options ...SpinnerOption) *Spinner {
	s := &Spinner{
		style:    style,
		color:    color.New(color.FgCyan),
		speed:    100 * time.Millisecond,
		stopChan: make(chan bool, 1),
		prefix:   "",
		suffix:   "",
	}
	
	// 应用配置选项
	for _, opt := range options {
		opt(s)
	}
	
	s.setFrames()
	return s
}

// setFrames 设置动画帧
func (s *Spinner) setFrames() {
	switch s.style {
	case SpinnerDots:
		s.frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	case SpinnerLine:
		s.frames = []string{"|", "/", "-", "\\"}
	case SpinnerCircle:
		s.frames = []string{"◐", "◓", "◑", "◒"}
	case SpinnerArrow:
		s.frames = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	case SpinnerBounce:
		s.frames = []string{"⠁", "⠂", "⠄", "⠂"}
	case SpinnerClock:
		s.frames = []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}
	case SpinnerMoon:
		s.frames = []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	case SpinnerStar:
		s.frames = []string{"✦", "✧", "✦", "✧"}
	default:
		s.frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}
}

// Start 启动Spinner动画
//
// 启动后Spinner会在独立的goroutine中运行动画循环
func (s *Spinner) Start() {
	s.mutex.Lock()
	if s.isActive {
		s.mutex.Unlock()
		return
	}
	s.isActive = true
	s.mutex.Unlock()
	
	go s.animationLoop()
}

// Stop 停止Spinner动画
//
// 停止动画并清理显示内容
func (s *Spinner) Stop() {
	s.mutex.Lock()
	if !s.isActive {
		s.mutex.Unlock()
		return
	}
	s.isActive = false
	s.mutex.Unlock()
	
	// 发送停止信号
	select {
	case s.stopChan <- true:
	default:
	}
	
	// 清理显示内容
	s.clearLine()
}

// SetText 设置显示文本
//
// 参数：
// - text: 新的显示文本
func (s *Spinner) SetText(text string) {
	s.mutex.Lock()
	s.text = text
	s.mutex.Unlock()
}

// SetColor 设置颜色
//
// 参数：
// - color: 新的颜色
func (s *Spinner) SetColor(color *color.Color) {
	s.mutex.Lock()
	s.color = color
	s.mutex.Unlock()
}

// IsActive 检查Spinner是否活跃
//
// 返回值：
// - bool: 是否活跃
func (s *Spinner) IsActive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isActive
}

// animationLoop 动画循环
func (s *Spinner) animationLoop() {
	ticker := time.NewTicker(s.speed)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.mutex.RLock()
			if !s.isActive {
				s.mutex.RUnlock()
				return
			}
			s.mutex.RUnlock()
			
			s.render()
			s.nextFrame()
		}
	}
}

// render 渲染当前帧
func (s *Spinner) render() {
	s.mutex.RLock()
	frame := s.frames[s.frameIndex]
	text := s.text
	prefix := s.prefix
	suffix := s.suffix
	color := s.color
	s.mutex.RUnlock()
	
	// 构建显示内容
	var content string
	if text != "" {
		content = fmt.Sprintf("%s%s %s%s", prefix, frame, text, suffix)
	} else {
		content = fmt.Sprintf("%s%s%s", prefix, frame, suffix)
	}
	
	// 清除当前行并显示新内容
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
	color.Print(content)
}

// nextFrame 切换到下一帧
func (s *Spinner) nextFrame() {
	s.mutex.Lock()
	s.frameIndex = (s.frameIndex + 1) % len(s.frames)
	s.mutex.Unlock()
}

// clearLine 清除当前行
func (s *Spinner) clearLine() {
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
}

// SpinnerWithTimeout 带超时的Spinner
//
// 在指定时间后自动停止Spinner
//
// 参数：
// - style: 动画样式
// - text: 显示文本
// - timeout: 超时时间
//
// 返回值：
// - *Spinner: Spinner实例
func SpinnerWithTimeout(style SpinnerStyle, text string, timeout time.Duration) *Spinner {
	spinner := NewSpinner(style, WithSpinnerText(text))
	
	go func() {
		time.Sleep(timeout)
		spinner.Stop()
	}()
	
	return spinner
}

// CreateLoadingSpinner 创建加载Spinner的便捷函数
//
// 参数：
// - text: 显示文本
//
// 返回值：
// - *Spinner: 配置好的Spinner实例
func CreateLoadingSpinner(text string) *Spinner {
	return NewSpinner(SpinnerDots,
		WithSpinnerText(text),
		WithSpinnerColor(color.New(color.FgCyan)),
		WithSpinnerSpeed(100*time.Millisecond),
	)
}

// CreateProcessingSpinner 创建处理Spinner的便捷函数
//
// 参数：
// - text: 显示文本
//
// 返回值：
// - *Spinner: 配置好的Spinner实例
func CreateProcessingSpinner(text string) *Spinner {
	return NewSpinner(SpinnerCircle,
		WithSpinnerText(text),
		WithSpinnerColor(color.New(color.FgYellow)),
		WithSpinnerSpeed(150*time.Millisecond),
	)
}

// CreateWaitingSpinner 创建等待Spinner的便捷函数
//
// 参数：
// - text: 显示文本
//
// 返回值：
// - *Spinner: 配置好的Spinner实例
func CreateWaitingSpinner(text string) *Spinner {
	return NewSpinner(SpinnerBounce,
		WithSpinnerText(text),
		WithSpinnerColor(color.New(color.FgMagenta)),
		WithSpinnerSpeed(200*time.Millisecond),
	)
}

// MultiSpinner 多Spinner管理器
//
// 用于管理多个同时运行的Spinner
type MultiSpinner struct {
	spinners map[string]*Spinner
	mutex    sync.RWMutex
}

// NewMultiSpinner 创建多Spinner管理器
//
// 返回值：
// - *MultiSpinner: MultiSpinner实例
func NewMultiSpinner() *MultiSpinner {
	return &MultiSpinner{
		spinners: make(map[string]*Spinner),
	}
}

// Add 添加Spinner
//
// 参数：
// - name: Spinner名称
// - spinner: Spinner实例
func (ms *MultiSpinner) Add(name string, spinner *Spinner) {
	ms.mutex.Lock()
	ms.spinners[name] = spinner
	ms.mutex.Unlock()
}

// Start 启动指定Spinner
//
// 参数：
// - name: Spinner名称
func (ms *MultiSpinner) Start(name string) {
	ms.mutex.RLock()
	spinner, exists := ms.spinners[name]
	ms.mutex.RUnlock()
	
	if exists {
		spinner.Start()
	}
}

// Stop 停止指定Spinner
//
// 参数：
// - name: Spinner名称
func (ms *MultiSpinner) Stop(name string) {
	ms.mutex.RLock()
	spinner, exists := ms.spinners[name]
	ms.mutex.RUnlock()
	
	if exists {
		spinner.Stop()
	}
}

// StopAll 停止所有Spinner
func (ms *MultiSpinner) StopAll() {
	ms.mutex.RLock()
	spinners := make([]*Spinner, 0, len(ms.spinners))
	for _, spinner := range ms.spinners {
		spinners = append(spinners, spinner)
	}
	ms.mutex.RUnlock()
	
	for _, spinner := range spinners {
		spinner.Stop()
	}
}

// Remove 移除Spinner
//
// 参数：
// - name: Spinner名称
func (ms *MultiSpinner) Remove(name string) {
	ms.mutex.Lock()
	if spinner, exists := ms.spinners[name]; exists {
		spinner.Stop()
		delete(ms.spinners, name)
	}
	ms.mutex.Unlock()
}

// GetSpinner 获取Spinner
//
// 参数：
// - name: Spinner名称
//
// 返回值：
// - *Spinner: Spinner实例
// - bool: 是否存在
func (ms *MultiSpinner) GetSpinner(name string) (*Spinner, bool) {
	ms.mutex.RLock()
	spinner, exists := ms.spinners[name]
	ms.mutex.RUnlock()
	return spinner, exists
}