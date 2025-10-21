package infrastructure

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"specify-cli/internal/types"
	"specify-cli/internal/ui"
)

// EnhancedProgressDisplay 增强的进度显示器
type EnhancedProgressDisplay struct {
	mu           sync.Mutex
	progressBar  *ui.ProgressBar
	message      string
	started      bool
	finished     bool
	startTime    time.Time
	lastUpdate   time.Time
	updateRate   time.Duration
	showDetails  bool
	showSpeed    bool
	showETA      bool
	customFormat string
}

// NewEnhancedProgressDisplay 创建增强的进度显示器
func NewEnhancedProgressDisplay(options ...ProgressDisplayOption) types.ProgressDisplay {
	display := &EnhancedProgressDisplay{
		updateRate:  100 * time.Millisecond,
		showDetails: true,
		showSpeed:   true,
		showETA:     true,
	}
	
	// 应用配置选项
	for _, option := range options {
		option(display)
	}
	
	return display
}

// ProgressDisplayOption 进度显示配置选项
type ProgressDisplayOption func(*EnhancedProgressDisplay)

// WithUpdateRate 设置更新频率
func WithUpdateRate(rate time.Duration) ProgressDisplayOption {
	return func(display *EnhancedProgressDisplay) {
		display.updateRate = rate
	}
}

// WithDisplayOptions 设置显示选项
func WithDisplayOptions(showDetails, showSpeed, showETA bool) ProgressDisplayOption {
	return func(display *EnhancedProgressDisplay) {
		display.showDetails = showDetails
		display.showSpeed = showSpeed
		display.showETA = showETA
	}
}

// WithCustomFormat 设置自定义格式
func WithCustomFormat(format string) ProgressDisplayOption {
	return func(display *EnhancedProgressDisplay) {
		display.customFormat = format
	}
}

// Start 开始进度显示
func (epd *EnhancedProgressDisplay) Start(total int64) {
	epd.mu.Lock()
	defer epd.mu.Unlock()
	
	epd.started = true
	epd.finished = false
	epd.startTime = time.Now()
	epd.lastUpdate = time.Now()
	
	// 创建进度条
	options := []ui.ProgressBarOption{
		ui.WithWidth(50),
		ui.ClassicStyle(),
		ui.WithColors(color.FgGreen, color.FgWhite),
		ui.WithDisplayOptions(true, epd.showDetails, epd.showSpeed, epd.showETA),
	}
	
	epd.progressBar = ui.NewProgressBar(total, epd.message, options...)
	
	if epd.message != "" {
		ui.ShowInfo(epd.message)
	}
}

// Update 更新进度
func (epd *EnhancedProgressDisplay) Update(info *types.ProgressInfo) {
	epd.mu.Lock()
	defer epd.mu.Unlock()
	
	if !epd.started || epd.finished {
		return
	}
	
	// 限制更新频率
	now := time.Now()
	if now.Sub(epd.lastUpdate) < epd.updateRate {
		return
	}
	epd.lastUpdate = now
	
	// 更新进度条
	if epd.progressBar != nil {
		epd.progressBar.Update(info.Downloaded)
	}
	
	// 显示详细信息
	if epd.showDetails && epd.customFormat != "" {
		epd.displayCustomInfo(info)
	}
}

// Finish 完成进度显示
func (epd *EnhancedProgressDisplay) Finish() {
	epd.mu.Lock()
	defer epd.mu.Unlock()
	
	if !epd.started || epd.finished {
		return
	}
	
	epd.finished = true
	
	if epd.progressBar != nil {
		epd.progressBar.Finish()
	}
	
	elapsed := time.Since(epd.startTime)
	ui.ShowSuccess(fmt.Sprintf("✓ 下载完成 (耗时: %s)", elapsed.Round(time.Second)))
}

// SetMessage 设置显示消息
func (epd *EnhancedProgressDisplay) SetMessage(message string) {
	epd.mu.Lock()
	defer epd.mu.Unlock()
	
	epd.message = message
	if epd.progressBar != nil {
		epd.progressBar.SetDescription(message)
	}
}

// displayCustomInfo 显示自定义信息
func (epd *EnhancedProgressDisplay) displayCustomInfo(info *types.ProgressInfo) {
	if epd.customFormat == "" {
		return
	}
	
	// 替换格式化字符串中的占位符
	format := epd.customFormat
	format = strings.ReplaceAll(format, "{downloaded}", formatBytes(info.Downloaded))
	format = strings.ReplaceAll(format, "{total}", formatBytes(info.Total))
	format = strings.ReplaceAll(format, "{percentage}", fmt.Sprintf("%.1f%%", info.Percentage))
	format = strings.ReplaceAll(format, "{speed}", formatBytes(int64(info.Speed))+"/s")
	format = strings.ReplaceAll(format, "{eta}", info.ETA.Round(time.Second).String())
	
	fmt.Printf("\r%s", format)
}

// MultiTaskProgressDisplay 多任务进度显示器
type MultiTaskProgressDisplay struct {
	mu           sync.Mutex
	tasks        map[string]*TaskProgress
	multiBar     *ui.MultiProgressBar
	started      bool
	finished     bool
	updateRate   time.Duration
	lastUpdate   time.Time
}

// TaskProgress 任务进度信息
type TaskProgress struct {
	ID          string
	Name        string
	Total       int64
	Current     int64
	ProgressBar *ui.ProgressBar
	StartTime   time.Time
	Status      string
}

// NewMultiTaskProgressDisplay 创建多任务进度显示器
func NewMultiTaskProgressDisplay() *MultiTaskProgressDisplay {
	return &MultiTaskProgressDisplay{
		tasks:      make(map[string]*TaskProgress),
		multiBar:   ui.NewMultiProgressBar(),
		updateRate: 100 * time.Millisecond,
	}
}

// AddTask 添加任务
func (mtpd *MultiTaskProgressDisplay) AddTask(id, name string, total int64) {
	mtpd.mu.Lock()
	defer mtpd.mu.Unlock()
	
	progressBar := ui.NewProgressBar(total, name,
		ui.WithWidth(40),
		ui.ModernStyle(),
		ui.WithDisplayOptions(true, true, false, false),
	)
	
	task := &TaskProgress{
		ID:          id,
		Name:        name,
		Total:       total,
		Current:     0,
		ProgressBar: progressBar,
		StartTime:   time.Now(),
		Status:      "pending",
	}
	
	mtpd.tasks[id] = task
	mtpd.multiBar.AddBar(progressBar)
}

// UpdateTask 更新任务进度
func (mtpd *MultiTaskProgressDisplay) UpdateTask(id string, current int64) {
	mtpd.mu.Lock()
	defer mtpd.mu.Unlock()
	
	task, exists := mtpd.tasks[id]
	if !exists {
		return
	}
	
	task.Current = current
	task.Status = "running"
	
	if task.ProgressBar != nil {
		task.ProgressBar.Update(current)
	}
	
	// 限制更新频率
	now := time.Now()
	if now.Sub(mtpd.lastUpdate) >= mtpd.updateRate {
		mtpd.multiBar.Render()
		mtpd.lastUpdate = now
	}
}

// CompleteTask 完成任务
func (mtpd *MultiTaskProgressDisplay) CompleteTask(id string) {
	mtpd.mu.Lock()
	defer mtpd.mu.Unlock()
	
	task, exists := mtpd.tasks[id]
	if !exists {
		return
	}
	
	task.Current = task.Total
	task.Status = "completed"
	
	if task.ProgressBar != nil {
		task.ProgressBar.Finish()
	}
	
	elapsed := time.Since(task.StartTime)
	ui.ShowSuccess(fmt.Sprintf("✓ %s 完成 (耗时: %s)", task.Name, elapsed.Round(time.Second)))
}

// Start 开始多任务显示
func (mtpd *MultiTaskProgressDisplay) Start() {
	mtpd.mu.Lock()
	defer mtpd.mu.Unlock()
	
	mtpd.started = true
	mtpd.multiBar.Start()
}

// Stop 停止多任务显示
func (mtpd *MultiTaskProgressDisplay) Stop() {
	mtpd.mu.Lock()
	defer mtpd.mu.Unlock()
	
	mtpd.finished = true
	mtpd.multiBar.Stop()
}

// GetTaskStatus 获取任务状态
func (mtpd *MultiTaskProgressDisplay) GetTaskStatus(id string) string {
	mtpd.mu.Lock()
	defer mtpd.mu.Unlock()
	
	task, exists := mtpd.tasks[id]
	if !exists {
		return "not_found"
	}
	
	return task.Status
}

// GetAllTasksStatus 获取所有任务状态
func (mtpd *MultiTaskProgressDisplay) GetAllTasksStatus() map[string]string {
	mtpd.mu.Lock()
	defer mtpd.mu.Unlock()
	
	status := make(map[string]string)
	for id, task := range mtpd.tasks {
		status[id] = task.Status
	}
	
	return status
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}