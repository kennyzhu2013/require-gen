package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ErrorLevel 错误级别
type ErrorLevel int

const (
	ErrorLevelInfo ErrorLevel = iota
	ErrorLevelWarning
	ErrorLevelError
	ErrorLevelCritical
)

// ErrorContext 错误上下文信息
type ErrorContext struct {
	Operation   string            // 操作名称
	Component   string            // 组件名称
	Details     map[string]string // 详细信息
	Timestamp   time.Time         // 时间戳
	Level       ErrorLevel        // 错误级别
	Suggestions []string          // 建议解决方案
}

// EnhancedErrorHandler 增强的错误处理器
type EnhancedErrorHandler struct {
	theme Theme  // 改为接口类型而不是指针
}

// NewEnhancedErrorHandler 创建增强错误处理器
func NewEnhancedErrorHandler() *EnhancedErrorHandler {
	return &EnhancedErrorHandler{
		theme: GetGlobalTheme(),  // 直接使用接口
	}
}

// HandleError 处理错误并显示友好的错误信息
func (eh *EnhancedErrorHandler) HandleError(err error, context *ErrorContext) {
	if err == nil {
		return
	}

	// 创建错误面板
	panel := eh.createErrorPanel(err, context)
	panel.Render()
}

// createErrorPanel 创建错误显示面板
func (eh *EnhancedErrorHandler) createErrorPanel(err error, context *ErrorContext) *Panel {
	var content strings.Builder
	
	// 错误标题
	levelIcon := eh.getLevelIcon(context.Level)
	levelColor := eh.getLevelColor(context.Level)
	
	content.WriteString(fmt.Sprintf("%s %s\n\n", 
		levelColor.Sprint(levelIcon), 
		levelColor.Sprint(eh.getLevelText(context.Level))))
	
	// 操作信息
	if context.Operation != "" {
		content.WriteString(fmt.Sprintf("Operation: %s\n", 
			eh.theme.Info().Sprint(context.Operation)))
	}
	
	if context.Component != "" {
		content.WriteString(fmt.Sprintf("Component: %s\n", 
			eh.theme.Info().Sprint(context.Component)))
	}
	
	// 错误消息
	content.WriteString(fmt.Sprintf("Error: %s\n", 
		eh.theme.Error().Sprint(err.Error())))
	
	// 详细信息
	if len(context.Details) > 0 {
		content.WriteString("\nDetails:\n")
		for key, value := range context.Details {
			content.WriteString(fmt.Sprintf("  %s: %s\n", 
				eh.theme.TextMuted().Sprint(key), 
				eh.theme.Info().Sprint(value)))
		}
	}
	
	// 建议解决方案
	if len(context.Suggestions) > 0 {
		content.WriteString("\nSuggested Solutions:\n")
		for i, suggestion := range context.Suggestions {
			content.WriteString(fmt.Sprintf("  %d. %s\n", 
				i+1, eh.theme.Success().Sprint(suggestion)))
		}
	}
	
	// 时间戳
	content.WriteString(fmt.Sprintf("\nTime: %s", 
		eh.theme.TextMuted().Sprint(context.Timestamp.Format("2006-01-02 15:04:05"))))
	
	// 创建面板
	borderColor := getColorAttribute(context.Level)
	return NewPanel(content.String(), "Error Details", 
		WithBorderStyle(borderColor),
		WithPadding(1, 2))
}

// getLevelIcon 获取错误级别图标
func (eh *EnhancedErrorHandler) getLevelIcon(level ErrorLevel) string {
	switch level {
	case ErrorLevelInfo:
		return "ℹ"
	case ErrorLevelWarning:
		return "⚠"
	case ErrorLevelError:
		return "✗"
	case ErrorLevelCritical:
		return "💥"
	default:
		return "?"
	}
}

// getLevelText 获取错误级别文本
func (eh *EnhancedErrorHandler) getLevelText(level ErrorLevel) string {
	switch level {
	case ErrorLevelInfo:
		return "Information"
	case ErrorLevelWarning:
		return "Warning"
	case ErrorLevelError:
		return "Error"
	case ErrorLevelCritical:
		return "Critical Error"
	default:
		return "Unknown"
	}
}

// getLevelColor 获取错误级别颜色
func (eh *EnhancedErrorHandler) getLevelColor(level ErrorLevel) *color.Color {
	switch level {
	case ErrorLevelInfo:
		return eh.theme.Info()
	case ErrorLevelWarning:
		return eh.theme.Warning()
	case ErrorLevelError:
		return eh.theme.Error()
	case ErrorLevelCritical:
		return color.New(color.FgRed, color.Bold, color.BlinkSlow)
	default:
		return eh.theme.TextMuted()
	}
}

// ShowProgressError 显示进度相关错误
func (eh *EnhancedErrorHandler) ShowProgressError(operation string, err error, progress int) {
	context := &ErrorContext{
		Operation: operation,
		Component: "Progress Tracker",
		Details: map[string]string{
			"Progress": fmt.Sprintf("%d%%", progress),
		},
		Timestamp: time.Now(),
		Level:     ErrorLevelError,
		Suggestions: []string{
			"Check network connection",
			"Verify file permissions",
			"Retry the operation",
		},
	}
	
	eh.HandleError(err, context)
}

// ShowValidationError 显示验证错误
func (eh *EnhancedErrorHandler) ShowValidationError(field string, value string, err error) {
	context := &ErrorContext{
		Operation: "Input Validation",
		Component: "Form Validator",
		Details: map[string]string{
			"Field": field,
			"Value": value,
		},
		Timestamp: time.Now(),
		Level:     ErrorLevelWarning,
		Suggestions: []string{
			"Check the input format",
			"Refer to the documentation",
			"Use the default value",
		},
	}
	
	eh.HandleError(err, context)
}

// ShowNetworkError 显示网络错误
func (eh *EnhancedErrorHandler) ShowNetworkError(url string, err error) {
	context := &ErrorContext{
		Operation: "Network Request",
		Component: "HTTP Client",
		Details: map[string]string{
			"URL": url,
		},
		Timestamp: time.Now(),
		Level:     ErrorLevelError,
		Suggestions: []string{
			"Check internet connection",
			"Verify the URL is correct",
			"Check firewall settings",
			"Try again later",
		},
	}
	
	eh.HandleError(err, context)
}

// ShowFileError 显示文件操作错误
func (eh *EnhancedErrorHandler) ShowFileError(filepath string, operation string, err error) {
	context := &ErrorContext{
		Operation: fmt.Sprintf("File %s", operation),
		Component: "File System",
		Details: map[string]string{
			"File Path": filepath,
			"Operation": operation,
		},
		Timestamp: time.Now(),
		Level:     ErrorLevelError,
		Suggestions: []string{
			"Check file permissions",
			"Verify the file path exists",
			"Ensure sufficient disk space",
			"Check if file is in use",
		},
	}
	
	eh.HandleError(err, context)
}

// ShowConfigError 显示配置错误
func (eh *EnhancedErrorHandler) ShowConfigError(configKey string, err error) {
	context := &ErrorContext{
		Operation: "Configuration Loading",
		Component: "Config Manager",
		Details: map[string]string{
			"Config Key": configKey,
		},
		Timestamp: time.Now(),
		Level:     ErrorLevelCritical,
		Suggestions: []string{
			"Check configuration file syntax",
			"Verify all required fields are present",
			"Reset to default configuration",
			"Check file permissions",
		},
	}
	
	eh.HandleError(err, context)
}

// ShowRecoveryMessage 显示恢复消息
func (eh *EnhancedErrorHandler) ShowRecoveryMessage(message string) {
	panel := NewPanel(
		fmt.Sprintf("%s %s", 
			eh.theme.Success().Sprint("✓"), 
			eh.theme.Info().Sprint(message)),
		"Recovery",
		WithBorderStyle(color.FgGreen),
		WithPadding(1, 2))
	panel.Render()
}

// 全局错误处理器实例
var globalErrorHandler *EnhancedErrorHandler

// GetGlobalErrorHandler 获取全局错误处理器
func GetGlobalErrorHandler() *EnhancedErrorHandler {
	if globalErrorHandler == nil {
		globalErrorHandler = NewEnhancedErrorHandler()
	}
	return globalErrorHandler
}

// 便捷函数

// HandleError 处理错误（全局函数）
func HandleError(err error, context *ErrorContext) {
	GetGlobalErrorHandler().HandleError(err, context)
}

// ShowEnhancedError 显示增强错误
func ShowEnhancedError(message string) {
	context := &ErrorContext{
		Operation: "General Operation",
		Timestamp: time.Now(),
		Level:     ErrorLevelError,
	}
	GetGlobalErrorHandler().HandleError(fmt.Errorf(message), context)
}

// ShowEnhancedWarning 显示增强警告
func ShowEnhancedWarning(message string) {
	context := &ErrorContext{
		Operation: "General Operation",
		Timestamp: time.Now(),
		Level:     ErrorLevelWarning,
	}
	GetGlobalErrorHandler().HandleError(fmt.Errorf(message), context)
}

// ShowEnhancedInfo 显示增强信息
func ShowEnhancedInfo(message string) {
	context := &ErrorContext{
		Operation: "General Operation",
		Timestamp: time.Now(),
		Level:     ErrorLevelInfo,
	}
	GetGlobalErrorHandler().HandleError(fmt.Errorf(message), context)
}

// GetAttribute 获取颜色属性的辅助函数
func getColorAttribute(level ErrorLevel) color.Attribute {
	switch level {
	case ErrorLevelInfo:
		return color.FgCyan
	case ErrorLevelWarning:
		return color.FgYellow
	case ErrorLevelError:
		return color.FgRed
	case ErrorLevelCritical:
		return color.FgRed
	default:
		return color.FgWhite
	}
}