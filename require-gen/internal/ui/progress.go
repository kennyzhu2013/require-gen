package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ProgressBar 进度条组件
//
// ProgressBar 提供了一个功能丰富的进度条实现，类似于Python Rich库的进度条。
// 支持多种样式、动画效果和实时更新功能。
//
// 主要特性：
// - 可自定义宽度和样式
// - 支持百分比和数值显示
// - 彩色进度条和状态指示
// - 实时更新和动画效果
// - 多种进度条样式选择
// - 支持任务描述和状态消息
//
// 使用场景：
// - 文件下载进度显示
// - 任务执行进度跟踪
// - 数据处理进度监控
// - 安装和配置过程显示
//
// 设计模式：
// - 建造者模式：通过链式调用配置进度条
// - 观察者模式：支持进度变化事件监听
// - 状态模式：不同状态下的不同显示效果
type ProgressBar struct {
	// 基本属性
	total       int64         // 总量
	current     int64         // 当前进度
	width       int           // 进度条宽度
	description string        // 任务描述
	
	// 样式配置
	fillChar    string        // 填充字符
	emptyChar   string        // 空白字符
	leftBorder  string        // 左边框
	rightBorder string        // 右边框
	
	// 颜色配置
	fillColor   *color.Color  // 填充颜色
	emptyColor  *color.Color  // 空白颜色
	textColor   *color.Color  // 文本颜色
	
	// 状态管理
	isCompleted bool          // 是否完成
	startTime   time.Time     // 开始时间
	lastUpdate  time.Time     // 最后更新时间
	
	// 显示选项
	showPercent bool          // 显示百分比
	showNumbers bool          // 显示数字
	showSpeed   bool          // 显示速度
	showETA     bool          // 显示预计完成时间
}

// ProgressBarOption 进度条配置选项
type ProgressBarOption func(*ProgressBar)

// NewProgressBar 创建新的进度条
//
// 创建一个具有默认配置的进度条实例。默认配置包括：
// - 宽度：50字符
// - 样式：经典进度条样式
// - 颜色：绿色填充，灰色空白
// - 显示：百分比和数字
//
// 参数：
//   total - 进度条的总量（最大值）
//   description - 进度条的描述文本
//   options - 可选的配置选项
//
// 返回值：
//   *ProgressBar - 配置好的进度条实例
//
// 示例用法：
//   bar := NewProgressBar(100, "下载文件",
//       WithWidth(60),
//       WithStyle("█", "░"),
//       WithColors(color.FgGreen, color.FgWhite),
//   )
func NewProgressBar(total int64, description string, options ...ProgressBarOption) *ProgressBar {
	bar := &ProgressBar{
		total:       total,
		current:     0,
		width:       50,
		description: description,
		
		// 默认样式
		fillChar:    "█",
		emptyChar:   "░",
		leftBorder:  "[",
		rightBorder: "]",
		
		// 默认颜色
		fillColor:   color.New(color.FgGreen),
		emptyColor:  color.New(color.FgWhite, color.Faint),
		textColor:   color.New(color.FgCyan),
		
		// 默认显示选项
		showPercent: true,
		showNumbers: true,
		showSpeed:   false,
		showETA:     false,
		
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
	
	// 应用配置选项
	for _, option := range options {
		option(bar)
	}
	
	return bar
}

// WithWidth 设置进度条宽度
func WithWidth(width int) ProgressBarOption {
	return func(bar *ProgressBar) {
		if width > 0 {
			bar.width = width
		}
	}
}

// WithStyle 设置进度条样式
func WithStyle(fillChar, emptyChar string) ProgressBarOption {
	return func(bar *ProgressBar) {
		bar.fillChar = fillChar
		bar.emptyChar = emptyChar
	}
}

// WithBorders 设置进度条边框
func WithBorders(left, right string) ProgressBarOption {
	return func(bar *ProgressBar) {
		bar.leftBorder = left
		bar.rightBorder = right
	}
}

// WithColors 设置进度条颜色
func WithColors(fillColor, emptyColor color.Attribute) ProgressBarOption {
	return func(bar *ProgressBar) {
		bar.fillColor = color.New(fillColor)
		bar.emptyColor = color.New(emptyColor)
	}
}

// WithTextColor 设置文本颜色
func WithTextColor(textColor color.Attribute) ProgressBarOption {
	return func(bar *ProgressBar) {
		bar.textColor = color.New(textColor)
	}
}

// WithDisplayOptions 设置显示选项
func WithDisplayOptions(showPercent, showNumbers, showSpeed, showETA bool) ProgressBarOption {
	return func(bar *ProgressBar) {
		bar.showPercent = showPercent
		bar.showNumbers = showNumbers
		bar.showSpeed = showSpeed
		bar.showETA = showETA
	}
}

// Update 更新进度
//
// 更新进度条的当前值并重新渲染显示。
// 如果当前值达到或超过总量，进度条将标记为完成状态。
//
// 参数：
//   current - 当前进度值
//
// 示例用法：
//   bar.Update(50)  // 更新到50%
func (pb *ProgressBar) Update(current int64) {
	pb.current = current
	pb.lastUpdate = time.Now()
	
	if pb.current >= pb.total {
		pb.isCompleted = true
		pb.current = pb.total
	}
	
	pb.render()
}

// Increment 增加进度
//
// 将当前进度增加指定的数量。
//
// 参数：
//   delta - 要增加的数量
func (pb *ProgressBar) Increment(delta int64) {
	pb.Update(pb.current + delta)
}

// SetDescription 设置描述文本
func (pb *ProgressBar) SetDescription(description string) {
	pb.description = description
}

// GetProgress 获取当前进度百分比
func (pb *ProgressBar) GetProgress() float64 {
	if pb.total == 0 {
		return 0
	}
	return float64(pb.current) / float64(pb.total) * 100
}

// IsCompleted 检查是否完成
func (pb *ProgressBar) IsCompleted() bool {
	return pb.isCompleted
}

// render 渲染进度条
func (pb *ProgressBar) render() {
	// 清除当前行
	fmt.Print("\r")
	
	// 计算进度
	progress := pb.GetProgress()
	filledWidth := int(float64(pb.width) * progress / 100)
	emptyWidth := pb.width - filledWidth
	
	// 构建进度条
	var bar strings.Builder
	
	// 添加描述
	if pb.description != "" {
		pb.textColor.Fprintf(&bar, "%s ", pb.description)
	}
	
	// 添加左边框
	bar.WriteString(pb.leftBorder)
	
	// 添加填充部分
	if filledWidth > 0 {
		pb.fillColor.Fprint(&bar, strings.Repeat(pb.fillChar, filledWidth))
	}
	
	// 添加空白部分
	if emptyWidth > 0 {
		pb.emptyColor.Fprint(&bar, strings.Repeat(pb.emptyChar, emptyWidth))
	}
	
	// 添加右边框
	bar.WriteString(pb.rightBorder)
	
	// 添加百分比
	if pb.showPercent {
		pb.textColor.Fprintf(&bar, " %.1f%%", progress)
	}
	
	// 添加数字显示
	if pb.showNumbers {
		pb.textColor.Fprintf(&bar, " (%d/%d)", pb.current, pb.total)
	}
	
	// 添加速度显示
	if pb.showSpeed && !pb.startTime.IsZero() {
		elapsed := pb.lastUpdate.Sub(pb.startTime).Seconds()
		if elapsed > 0 {
			speed := float64(pb.current) / elapsed
			pb.textColor.Fprintf(&bar, " %.1f/s", speed)
		}
	}
	
	// 添加预计完成时间
	if pb.showETA && !pb.isCompleted && pb.current > 0 {
		elapsed := pb.lastUpdate.Sub(pb.startTime).Seconds()
		if elapsed > 0 {
			speed := float64(pb.current) / elapsed
			remaining := float64(pb.total-pb.current) / speed
			eta := time.Duration(remaining) * time.Second
			pb.textColor.Fprintf(&bar, " ETA: %s", eta.Round(time.Second))
		}
	}
	
	// 输出进度条
	fmt.Print(bar.String())
	
	// 如果完成，换行
	if pb.isCompleted {
		fmt.Println()
	}
}

// Finish 完成进度条
//
// 将进度条设置为完成状态并进行最终渲染。
// 通常在任务完成时调用，确保进度条显示100%。
func (pb *ProgressBar) Finish() {
	// 如果已经完成，不需要重复渲染
	if pb.isCompleted {
		return
	}
	
	pb.current = pb.total
	pb.isCompleted = true
	pb.render()
}

// Reset 重置进度条
//
// 将进度条重置到初始状态，可以重新开始使用。
func (pb *ProgressBar) Reset() {
	pb.current = 0
	pb.isCompleted = false
	pb.startTime = time.Now()
	pb.lastUpdate = time.Now()
}

// MultiProgressBar 多进度条管理器
//
// MultiProgressBar 允许同时管理和显示多个进度条，
// 类似于Python Rich库的多进度条功能。
type MultiProgressBar struct {
	bars    []*ProgressBar
	active  bool
}

// NewMultiProgressBar 创建多进度条管理器
func NewMultiProgressBar() *MultiProgressBar {
	return &MultiProgressBar{
		bars:   make([]*ProgressBar, 0),
		active: false,
	}
}

// AddBar 添加进度条
func (mpb *MultiProgressBar) AddBar(bar *ProgressBar) {
	mpb.bars = append(mpb.bars, bar)
}

// Start 开始显示多进度条
func (mpb *MultiProgressBar) Start() {
	mpb.active = true
	// 隐藏光标
	fmt.Print("\033[?25l")
}

// Stop 停止显示多进度条
func (mpb *MultiProgressBar) Stop() {
	mpb.active = false
	// 显示光标
	fmt.Print("\033[?25h")
	fmt.Println()
}

// Render 渲染所有进度条
func (mpb *MultiProgressBar) Render() {
	if !mpb.active {
		return
	}
	
	// 移动到起始位置
	fmt.Printf("\033[%dA", len(mpb.bars))
	
	// 渲染每个进度条
	for _, bar := range mpb.bars {
		bar.render()
		fmt.Println()
	}
}

// 预定义的进度条样式
var (
	// ClassicStyle 经典样式
	ClassicStyle = func() ProgressBarOption {
		return WithStyle("█", "░")
	}
	
	// ModernStyle 现代样式
	ModernStyle = func() ProgressBarOption {
		return WithStyle("▓", "▒")
	}
	
	// MinimalStyle 简约样式
	MinimalStyle = func() ProgressBarOption {
		return WithStyle("■", "□")
	}
	
	// ArrowStyle 箭头样式
	ArrowStyle = func() ProgressBarOption {
		return WithStyle("►", "─")
	}
	
	// DotStyle 点状样式
	DotStyle = func() ProgressBarOption {
		return WithStyle("●", "○")
	}
)