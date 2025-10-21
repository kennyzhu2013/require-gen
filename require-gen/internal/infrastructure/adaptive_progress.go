package infrastructure

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"specify-cli/internal/types"
)

// AdaptiveProgressDisplay 自适应进度显示器
type AdaptiveProgressDisplay struct {
	mu             sync.Mutex
	context        context.Context
	cancel         context.CancelFunc
	display        types.ProgressDisplay
	displayType    ProgressDisplayType
	isInteractive  bool
	terminalWidth  int
	updateInterval time.Duration
	lastUpdate     time.Time
	config         *AdaptiveConfig
}

// ProgressDisplayType 进度显示类型
type ProgressDisplayType int

const (
	DisplayTypeAuto ProgressDisplayType = iota
	DisplayTypeConsole
	DisplayTypeEnhanced
	DisplayTypeMinimal
	DisplayTypeSilent
)

// AdaptiveConfig 自适应配置
type AdaptiveConfig struct {
	// 显示配置
	PreferredType ProgressDisplayType
	MinUpdateRate time.Duration
	MaxUpdateRate time.Duration
	ShowSpeed     bool
	ShowETA       bool
	ShowDetails   bool

	// 自适应配置
	AutoDetectTerminal bool
	FallbackToMinimal  bool
	SilentMode         bool

	// 性能配置
	BufferUpdates bool
	MaxBufferSize int
	FlushInterval time.Duration
}

// DefaultAdaptiveConfig 默认自适应配置
func DefaultAdaptiveConfig() *AdaptiveConfig {
	return &AdaptiveConfig{
		PreferredType:      DisplayTypeAuto,
		MinUpdateRate:      50 * time.Millisecond,
		MaxUpdateRate:      500 * time.Millisecond,
		ShowSpeed:          true,
		ShowETA:            true,
		ShowDetails:        true,
		AutoDetectTerminal: true,
		FallbackToMinimal:  true,
		SilentMode:         false,
		BufferUpdates:      true,
		MaxBufferSize:      100,
		FlushInterval:      200 * time.Millisecond,
	}
}

// NewAdaptiveProgressDisplay 创建自适应进度显示器
func NewAdaptiveProgressDisplay(config *AdaptiveConfig) types.ProgressDisplay {
	if config == nil {
		config = DefaultAdaptiveConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	display := &AdaptiveProgressDisplay{
		context:        ctx,
		cancel:         cancel,
		config:         config,
		updateInterval: config.MinUpdateRate,
	}

	// 检测终端环境
	display.detectEnvironment()

	// 选择合适的显示器
	display.selectDisplay()

	return display
}

// detectEnvironment 检测环境
func (apd *AdaptiveProgressDisplay) detectEnvironment() {
	// 检测是否为交互式终端
	apd.isInteractive = isInteractiveTerminal()

	// 获取终端宽度
	apd.terminalWidth = getTerminalWidth()

	// 检测是否在CI环境中
	if isCI() {
		apd.config.SilentMode = true
		apd.config.PreferredType = DisplayTypeSilent
	}

	// 检测是否支持颜色
	if !supportsColor() {
		apd.config.PreferredType = DisplayTypeMinimal
	}
}

// selectDisplay 选择显示器
func (apd *AdaptiveProgressDisplay) selectDisplay() {
	displayType := apd.config.PreferredType

	// 自动检测最佳显示类型
	if displayType == DisplayTypeAuto {
		displayType = apd.autoSelectDisplayType()
	}

	apd.displayType = displayType

	switch displayType {
	case DisplayTypeEnhanced:
		apd.display = NewEnhancedProgressDisplay(
			WithDisplayOptions(apd.config.ShowDetails, apd.config.ShowSpeed, apd.config.ShowETA),
			WithUpdateRate(apd.config.MinUpdateRate),
		)
	case DisplayTypeConsole:
		apd.display = NewMinimalProgressDisplay()
	case DisplayTypeMinimal:
		apd.display = NewMinimalProgressDisplay()
	case DisplayTypeSilent:
		apd.display = NewSilentProgressDisplay()
	default:
		// 回退到最小显示
		apd.display = NewMinimalProgressDisplay()
	}
}

// autoSelectDisplayType 自动选择显示类型
func (apd *AdaptiveProgressDisplay) autoSelectDisplayType() ProgressDisplayType {
	// 静默模式
	if apd.config.SilentMode {
		return DisplayTypeSilent
	}

	// 非交互式终端
	if !apd.isInteractive {
		return DisplayTypeMinimal
	}

	// 终端宽度太小
	if apd.terminalWidth < 60 {
		return DisplayTypeMinimal
	}

	// 支持增强显示
	if apd.terminalWidth >= 80 && supportsColor() {
		return DisplayTypeEnhanced
	}

	// 默认控制台显示
	return DisplayTypeConsole
}

// Start 开始进度显示
func (apd *AdaptiveProgressDisplay) Start(total int64) {
	apd.mu.Lock()
	defer apd.mu.Unlock()

	if apd.display != nil {
		apd.display.Start(total)
	}

	// 启动自适应更新
	if apd.config.BufferUpdates {
		go apd.startBufferedUpdates()
	}
}

// Update 更新进度
func (apd *AdaptiveProgressDisplay) Update(info *types.ProgressInfo) {
	apd.mu.Lock()
	defer apd.mu.Unlock()

	// 检查更新频率
	now := time.Now()
	if now.Sub(apd.lastUpdate) < apd.updateInterval {
		return
	}
	apd.lastUpdate = now

	// 自适应调整更新频率
	apd.adaptUpdateRate(info)

	if apd.display != nil {
		apd.display.Update(info)
	}
}

// Finish 完成进度显示
func (apd *AdaptiveProgressDisplay) Finish() {
	apd.mu.Lock()
	defer apd.mu.Unlock()

	if apd.display != nil {
		apd.display.Finish()
	}

	// 取消上下文
	if apd.cancel != nil {
		apd.cancel()
	}
}

// SetMessage 设置显示消息
func (apd *AdaptiveProgressDisplay) SetMessage(message string) {
	apd.mu.Lock()
	defer apd.mu.Unlock()

	if apd.display != nil {
		apd.display.SetMessage(message)
	}
}

// adaptUpdateRate 自适应调整更新频率
func (apd *AdaptiveProgressDisplay) adaptUpdateRate(info *types.ProgressInfo) {
	// 根据下载速度调整更新频率
	if info.Speed > 10*1024*1024 { // 10MB/s以上，降低更新频率
		apd.updateInterval = apd.config.MaxUpdateRate
	} else if info.Speed > 1024*1024 { // 1MB/s以上，中等更新频率
		apd.updateInterval = (apd.config.MinUpdateRate + apd.config.MaxUpdateRate) / 2
	} else { // 低速下载，高更新频率
		apd.updateInterval = apd.config.MinUpdateRate
	}
}

// startBufferedUpdates 启动缓冲更新
func (apd *AdaptiveProgressDisplay) startBufferedUpdates() {
	ticker := time.NewTicker(apd.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-apd.context.Done():
			return
		case <-ticker.C:
			// 执行缓冲刷新逻辑
			apd.flushBufferedUpdates()
		}
	}
}

// flushBufferedUpdates 刷新缓冲更新
func (apd *AdaptiveProgressDisplay) flushBufferedUpdates() {
	// 这里可以实现缓冲更新的刷新逻辑
	// 例如批量更新UI，减少频繁的屏幕刷新
}

// MinimalProgressDisplay 最小化进度显示器
type MinimalProgressDisplay struct {
	mu       sync.Mutex
	message  string
	started  bool
	finished bool
	lastPct  int
}

// NewMinimalProgressDisplay 创建最小化进度显示器
func NewMinimalProgressDisplay() types.ProgressDisplay {
	return &MinimalProgressDisplay{}
}

// Start 开始进度显示
func (mpd *MinimalProgressDisplay) Start(total int64) {
	mpd.mu.Lock()
	defer mpd.mu.Unlock()

	mpd.started = true
	mpd.finished = false

	if mpd.message != "" {
		fmt.Printf("%s...\n", mpd.message)
	}
}

// Update 更新进度
func (mpd *MinimalProgressDisplay) Update(info *types.ProgressInfo) {
	mpd.mu.Lock()
	defer mpd.mu.Unlock()

	if !mpd.started || mpd.finished {
		return
	}

	// 只在百分比有显著变化时更新
	currentPct := int(info.Percentage)
	if currentPct-mpd.lastPct >= 10 {
		fmt.Printf("Progress: %d%%\n", currentPct)
		mpd.lastPct = currentPct
	}
}

// Finish 完成进度显示
func (mpd *MinimalProgressDisplay) Finish() {
	mpd.mu.Lock()
	defer mpd.mu.Unlock()

	if !mpd.started || mpd.finished {
		return
	}

	mpd.finished = true
	fmt.Println("Completed.")
}

// SetMessage 设置显示消息
func (mpd *MinimalProgressDisplay) SetMessage(message string) {
	mpd.mu.Lock()
	defer mpd.mu.Unlock()

	mpd.message = message
}

// SilentProgressDisplay 静默进度显示器
type SilentProgressDisplay struct{}

// NewSilentProgressDisplay 创建静默进度显示器
func NewSilentProgressDisplay() types.ProgressDisplay {
	return &SilentProgressDisplay{}
}

// Start 开始进度显示（静默）
func (spd *SilentProgressDisplay) Start(total int64) {}

// Update 更新进度（静默）
func (spd *SilentProgressDisplay) Update(info *types.ProgressInfo) {}

// Finish 完成进度显示（静默）
func (spd *SilentProgressDisplay) Finish() {}

// SetMessage 设置显示消息（静默）
func (spd *SilentProgressDisplay) SetMessage(message string) {}

// 环境检测辅助函数

// isInteractiveTerminal 检测是否为交互式终端
func isInteractiveTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// getTerminalWidth 获取终端宽度
func getTerminalWidth() int {
	// 这里可以使用系统调用获取终端宽度
	// 简化实现，返回默认值
	return 80
}

// isCI 检测是否在CI环境中
func isCI() bool {
	ciEnvs := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL"}
	for _, env := range ciEnvs {
		if os.Getenv(env) != "" {
			return true
		}
	}
	return false
}

// supportsColor 检测是否支持颜色
func supportsColor() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	// 检测常见的支持颜色的终端
	colorTerms := []string{"xterm", "xterm-256color", "screen", "tmux"}
	for _, colorTerm := range colorTerms {
		if term == colorTerm {
			return true
		}
	}

	return false
}
