package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Live 实时更新显示组件
//
// Live提供了实时刷新显示功能，类似于Python Rich库的Live组件。
// 支持动态内容更新、减少屏幕闪烁、并发安全等特性。
//
// 特性：
// - 实时内容更新
// - 减少屏幕闪烁
// - 并发安全操作
// - 自动清理和恢复
// - 可配置刷新频率
// - 支持多行内容
type Live struct {
	content     string        // 当前内容
	lastContent string        // 上次内容
	mutex       sync.RWMutex  // 读写锁
	isActive    bool          // 是否活跃
	refreshRate time.Duration // 刷新频率
	stopChan    chan bool     // 停止信号
	lastLines   int           // 上次显示的行数
	autoRefresh bool          // 自动刷新
}

// LiveOption Live配置选项函数类型
type LiveOption func(*Live)

// WithRefreshRate 设置刷新频率
func WithRefreshRate(rate time.Duration) LiveOption {
	return func(l *Live) {
		l.refreshRate = rate
	}
}

// WithAutoRefresh 设置自动刷新
func WithAutoRefresh(auto bool) LiveOption {
	return func(l *Live) {
		l.autoRefresh = auto
	}
}

// NewLive 创建新的Live实例
//
// 参数：
// - options: 配置选项
//
// 返回值：
// - *Live: Live实例
func NewLive(options ...LiveOption) *Live {
	live := &Live{
		refreshRate: 100 * time.Millisecond,
		stopChan:    make(chan bool, 1),
		autoRefresh: true,
	}
	
	// 应用配置选项
	for _, opt := range options {
		opt(live)
	}
	
	return live
}

// Start 启动Live显示
//
// 启动后Live会在独立的goroutine中运行刷新循环
func (l *Live) Start() {
	l.mutex.Lock()
	if l.isActive {
		l.mutex.Unlock()
		return
	}
	l.isActive = true
	l.mutex.Unlock()
	
	if l.autoRefresh {
		go l.refreshLoop()
	}
}

// Stop 停止Live显示
//
// 停止刷新循环并清理显示内容
func (l *Live) Stop() {
	l.mutex.Lock()
	if !l.isActive {
		l.mutex.Unlock()
		return
	}
	l.isActive = false
	l.mutex.Unlock()
	
	// 发送停止信号
	select {
	case l.stopChan <- true:
	default:
	}
	
	// 清理显示内容
	l.clearLastContent()
}

// Update 更新显示内容
//
// 参数：
// - content: 新的显示内容
func (l *Live) Update(content string) {
	l.mutex.Lock()
	l.content = content
	l.mutex.Unlock()
	
	// 如果没有自动刷新，立即刷新
	if !l.autoRefresh {
		l.refresh()
	}
}

// UpdateAndRefresh 更新内容并立即刷新
//
// 参数：
// - content: 新的显示内容
func (l *Live) UpdateAndRefresh(content string) {
	l.mutex.Lock()
	l.content = content
	l.mutex.Unlock()
	
	l.refresh()
}

// GetContent 获取当前内容
//
// 返回值：
// - string: 当前内容
func (l *Live) GetContent() string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.content
}

// IsActive 检查Live是否活跃
//
// 返回值：
// - bool: 是否活跃
func (l *Live) IsActive() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.isActive
}

// refreshLoop 刷新循环
func (l *Live) refreshLoop() {
	ticker := time.NewTicker(l.refreshRate)
	defer ticker.Stop()
	
	for {
		select {
		case <-l.stopChan:
			return
		case <-ticker.C:
			l.mutex.RLock()
			if !l.isActive {
				l.mutex.RUnlock()
				return
			}
			l.mutex.RUnlock()
			
			l.refresh()
		}
	}
}

// refresh 刷新显示内容
func (l *Live) refresh() {
	l.mutex.RLock()
	currentContent := l.content
	lastContent := l.lastContent
	isActive := l.isActive
	l.mutex.RUnlock()
	
	if !isActive || currentContent == lastContent {
		return
	}
	
	// 清除之前的内容
	l.clearLastContent()
	
	// 显示新内容
	if currentContent != "" {
		fmt.Print(currentContent)
		
		// 更新行数记录
		l.mutex.Lock()
		l.lastContent = currentContent
		l.lastLines = strings.Count(currentContent, "\n")
		if !strings.HasSuffix(currentContent, "\n") {
			l.lastLines++
		}
		l.mutex.Unlock()
	}
}

// clearLastContent 清除上次显示的内容
func (l *Live) clearLastContent() {
	l.mutex.RLock()
	lastLines := l.lastLines
	l.mutex.RUnlock()
	
	if lastLines > 0 {
		// 移动光标到开始位置并清除内容
		for i := 0; i < lastLines; i++ {
			fmt.Print("\033[1A\033[2K") // 上移一行并清除该行
		}
	}
}

// Clear 清除所有内容
func (l *Live) Clear() {
	l.mutex.Lock()
	l.content = ""
	l.lastContent = ""
	l.mutex.Unlock()
	
	l.clearLastContent()
}

// Append 追加内容
//
// 参数：
// - content: 要追加的内容
func (l *Live) Append(content string) {
	l.mutex.Lock()
	l.content += content
	l.mutex.Unlock()
	
	if !l.autoRefresh {
		l.refresh()
	}
}

// AppendLine 追加一行内容
//
// 参数：
// - line: 要追加的行内容
func (l *Live) AppendLine(line string) {
	l.Append(line + "\n")
}

// SetRefreshRate 设置刷新频率
//
// 参数：
// - rate: 新的刷新频率
func (l *Live) SetRefreshRate(rate time.Duration) {
	l.mutex.Lock()
	l.refreshRate = rate
	l.mutex.Unlock()
}

// LiveRenderer 实时渲染器
//
// 用于渲染复杂的实时内容
type LiveRenderer struct {
	live     *Live
	builders []ContentBuilder
	mutex    sync.RWMutex
}

// ContentBuilder 内容构建器接口
type ContentBuilder interface {
	Build() string
}

// NewLiveRenderer 创建实时渲染器
//
// 返回值：
// - *LiveRenderer: LiveRenderer实例
func NewLiveRenderer() *LiveRenderer {
	return &LiveRenderer{
		live:     NewLive(),
		builders: make([]ContentBuilder, 0),
	}
}

// AddBuilder 添加内容构建器
//
// 参数：
// - builder: 内容构建器
func (lr *LiveRenderer) AddBuilder(builder ContentBuilder) {
	lr.mutex.Lock()
	lr.builders = append(lr.builders, builder)
	lr.mutex.Unlock()
}

// Start 启动渲染器
func (lr *LiveRenderer) Start() {
	lr.live.Start()
	go lr.renderLoop()
}

// Stop 停止渲染器
func (lr *LiveRenderer) Stop() {
	lr.live.Stop()
}

// renderLoop 渲染循环
func (lr *LiveRenderer) renderLoop() {
	ticker := time.NewTicker(lr.live.refreshRate)
	defer ticker.Stop()
	
	for lr.live.IsActive() {
		select {
		case <-ticker.C:
			lr.render()
		}
	}
}

// render 渲染内容
func (lr *LiveRenderer) render() {
	lr.mutex.RLock()
	builders := make([]ContentBuilder, len(lr.builders))
	copy(builders, lr.builders)
	lr.mutex.RUnlock()
	
	var content strings.Builder
	for _, builder := range builders {
		content.WriteString(builder.Build())
		content.WriteString("\n")
	}
	
	lr.live.Update(content.String())
}

// ProgressBuilder 进度构建器
//
// 实现ContentBuilder接口，用于构建进度显示内容
type ProgressBuilder struct {
	label    string
	current  int64
	total    int64
	width    int
	showPct  bool
}

// NewProgressBuilder 创建进度构建器
//
// 参数：
// - label: 标签
// - total: 总量
//
// 返回值：
// - *ProgressBuilder: ProgressBuilder实例
func NewProgressBuilder(label string, total int64) *ProgressBuilder {
	return &ProgressBuilder{
		label:   label,
		total:   total,
		width:   40,
		showPct: true,
	}
}

// SetProgress 设置进度
//
// 参数：
// - current: 当前进度
func (pb *ProgressBuilder) SetProgress(current int64) {
	pb.current = current
}

// Build 构建内容
//
// 返回值：
// - string: 构建的内容
func (pb *ProgressBuilder) Build() string {
	if pb.total == 0 {
		return fmt.Sprintf("%s: 0%%", pb.label)
	}
	
	percentage := float64(pb.current) / float64(pb.total)
	filled := int(percentage * float64(pb.width))
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)
	
	if pb.showPct {
		return fmt.Sprintf("%s: [%s] %.1f%%", pb.label, bar, percentage*100)
	}
	
	return fmt.Sprintf("%s: [%s] %d/%d", pb.label, bar, pb.current, pb.total)
}

// StatusBuilder 状态构建器
//
// 实现ContentBuilder接口，用于构建状态显示内容
type StatusBuilder struct {
	items map[string]string
	mutex sync.RWMutex
}

// NewStatusBuilder 创建状态构建器
//
// 返回值：
// - *StatusBuilder: StatusBuilder实例
func NewStatusBuilder() *StatusBuilder {
	return &StatusBuilder{
		items: make(map[string]string),
	}
}

// SetStatus 设置状态
//
// 参数：
// - key: 状态键
// - value: 状态值
func (sb *StatusBuilder) SetStatus(key, value string) {
	sb.mutex.Lock()
	sb.items[key] = value
	sb.mutex.Unlock()
}

// Build 构建内容
//
// 返回值：
// - string: 构建的内容
func (sb *StatusBuilder) Build() string {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	
	var content strings.Builder
	for key, value := range sb.items {
		content.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	
	return content.String()
}

// CreateLiveProgress 创建实时进度显示的便捷函数
//
// 参数：
// - label: 进度标签
// - total: 总量
//
// 返回值：
// - *Live: Live实例
// - *ProgressBuilder: 进度构建器
func CreateLiveProgress(label string, total int64) (*Live, *ProgressBuilder) {
	live := NewLive()
	builder := NewProgressBuilder(label, total)
	
	// 设置更新函数
	updateFunc := func() {
		live.Update(builder.Build())
	}
	
	// 启动定时更新
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		for live.IsActive() {
			select {
			case <-ticker.C:
				updateFunc()
			}
		}
	}()
	
	return live, builder
}