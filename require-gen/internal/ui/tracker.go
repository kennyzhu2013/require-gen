package ui

import (
	"fmt"
	"sort"
	"sync"

	"github.com/fatih/color"
	"specify-cli/internal/types"
)

// StepTracker 步骤跟踪器实现
type StepTracker struct {
	title       string
	steps       map[string]*types.Step
	statusOrder map[string]int
	mutex       sync.RWMutex
	observers   []types.StepObserver
}

// NewStepTracker 创建新的步骤跟踪器
func NewStepTracker(title string) *StepTracker {
	return &StepTracker{
		title: title,
		steps: make(map[string]*types.Step),
		statusOrder: map[string]int{
			types.StatusPending: 0,
			types.StatusRunning: 1,
			types.StatusDone:    2,
			types.StatusError:   3,
			types.StatusSkipped: 4,
		},
		observers: make([]types.StepObserver, 0),
	}
}

// AddStep 添加步骤
func (st *StepTracker) AddStep(key, label string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	step := &types.Step{
		Key:    key,
		Label:  label,
		Status: types.StatusPending,
		Detail: "",
	}
	st.steps[key] = step
}

// UpdateStep 更新步骤状态
func (st *StepTracker) UpdateStep(key, status, detail string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if step, exists := st.steps[key]; exists {
		step.Status = status
		step.Detail = detail

		// 通知观察者
		for _, observer := range st.observers {
			observer.OnStepChanged(step)
		}
	}
}

// SetStepRunning 设置步骤为运行状态
func (st *StepTracker) SetStepRunning(key, detail string) {
	st.UpdateStep(key, types.StatusRunning, detail)
}

// SetStepDone 设置步骤为完成状态
func (st *StepTracker) SetStepDone(key, detail string) {
	st.UpdateStep(key, types.StatusDone, detail)
}

// SetStepError 设置步骤为错误状态
func (st *StepTracker) SetStepError(key, detail string) {
	st.UpdateStep(key, types.StatusError, detail)
}

// SetStepSkipped 设置步骤为跳过状态
func (st *StepTracker) SetStepSkipped(key, detail string) {
	st.UpdateStep(key, types.StatusSkipped, detail)
}

// GetStep 获取步骤信息
func (st *StepTracker) GetStep(key string) (*types.Step, bool) {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	step, exists := st.steps[key]
	return step, exists
}

// GetAllSteps 获取所有步骤
func (st *StepTracker) GetAllSteps() []*types.Step {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	steps := make([]*types.Step, 0, len(st.steps))
	for _, step := range st.steps {
		steps = append(steps, step)
	}

	// 按状态排序
	sort.Slice(steps, func(i, j int) bool {
		return st.statusOrder[steps[i].Status] < st.statusOrder[steps[j].Status]
	})

	return steps
}

// Display 显示步骤跟踪器
func (st *StepTracker) Display() {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// 显示标题
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Printf("\n=== %s ===\n", st.title)

	// 显示步骤
	for _, step := range st.GetAllSteps() {
		st.displayStep(step)
	}
	fmt.Println()
}

// displayStep 显示单个步骤
func (st *StepTracker) displayStep(step *types.Step) {
	var statusIcon string
	var statusColor *color.Color

	switch step.Status {
	case types.StatusPending:
		statusIcon = "⏳"
		statusColor = color.New(color.FgYellow)
	case types.StatusRunning:
		statusIcon = "🔄"
		statusColor = color.New(color.FgBlue, color.Bold)
	case types.StatusDone:
		statusIcon = "✅"
		statusColor = color.New(color.FgGreen, color.Bold)
	case types.StatusError:
		statusIcon = "❌"
		statusColor = color.New(color.FgRed, color.Bold)
	case types.StatusSkipped:
		statusIcon = "⏭️"
		statusColor = color.New(color.FgMagenta)
	default:
		statusIcon = "❓"
		statusColor = color.New(color.FgWhite)
	}

	// 显示步骤信息
	statusColor.Printf("%s %s", statusIcon, step.Label)
	if step.Detail != "" {
		fmt.Printf(" - %s", step.Detail)
	}
	fmt.Println()
}

// AddObserver 添加观察者
func (st *StepTracker) AddObserver(observer types.StepObserver) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.observers = append(st.observers, observer)
}

// RemoveObserver 移除观察者
func (st *StepTracker) RemoveObserver(observer types.StepObserver) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	for i, obs := range st.observers {
		if obs == observer {
			st.observers = append(st.observers[:i], st.observers[i+1:]...)
			break
		}
	}
}

// IsCompleted 检查是否所有步骤都已完成
func (st *StepTracker) IsCompleted() bool {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	for _, step := range st.steps {
		if step.Status == types.StatusPending || step.Status == types.StatusRunning {
			return false
		}
	}
	return true
}

// HasErrors 检查是否有错误步骤
func (st *StepTracker) HasErrors() bool {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	for _, step := range st.steps {
		if step.Status == types.StatusError {
			return true
		}
	}
	return false
}

// GetProgress 获取进度百分比
func (st *StepTracker) GetProgress() float64 {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	if len(st.steps) == 0 {
		return 0
	}

	completed := 0
	for _, step := range st.steps {
		if step.Status == types.StatusDone || step.Status == types.StatusError || step.Status == types.StatusSkipped {
			completed++
		}
	}

	return float64(completed) / float64(len(st.steps)) * 100
}

// Reset 重置所有步骤
func (st *StepTracker) Reset() {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	for _, step := range st.steps {
		step.Status = types.StatusPending
		step.Detail = ""
	}
}