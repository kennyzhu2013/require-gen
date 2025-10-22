package ui

import (
	"strings"
	"unicode/utf8"
)

// Alignment 对齐方式
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
	AlignJustify
)

// Align 文本对齐组件
type Align struct {
	content   string
	alignment Alignment
	width     int
	padding   string
}

// AlignOption 对齐选项函数类型
type AlignOption func(*Align)

// WithAlignment 设置对齐方式
func WithAlignment(alignment Alignment) AlignOption {
	return func(a *Align) {
		a.alignment = alignment
	}
}

// WithWidth 设置宽度
func WithWidth(width int) AlignOption {
	return func(a *Align) {
		a.width = width
	}
}

// WithPadding 设置填充字符
func WithPadding(padding string) AlignOption {
	return func(a *Align) {
		a.padding = padding
	}
}

// NewAlign 创建新的对齐组件
func NewAlign(content string, options ...AlignOption) *Align {
	align := &Align{
		content:   content,
		alignment: AlignLeft,
		width:     0,
		padding:   " ",
	}
	
	for _, option := range options {
		option(align)
	}
	
	return align
}

// Render 渲染对齐的文本
func (a *Align) Render() string {
	if a.width <= 0 {
		return a.content
	}
	
	lines := strings.Split(a.content, "\n")
	var result []string
	
	for _, line := range lines {
		result = append(result, a.alignLine(line))
	}
	
	return strings.Join(result, "\n")
}

// alignLine 对齐单行文本
func (a *Align) alignLine(line string) string {
	lineWidth := utf8.RuneCountInString(line)
	
	if lineWidth >= a.width {
		return line
	}
	
	switch a.alignment {
	case AlignLeft:
		return line + strings.Repeat(a.padding, a.width-lineWidth)
	case AlignRight:
		return strings.Repeat(a.padding, a.width-lineWidth) + line
	case AlignCenter:
		leftPad := (a.width - lineWidth) / 2
		rightPad := a.width - lineWidth - leftPad
		return strings.Repeat(a.padding, leftPad) + line + strings.Repeat(a.padding, rightPad)
	case AlignJustify:
		return a.justifyLine(line)
	default:
		return line
	}
}

// justifyLine 两端对齐单行文本
func (a *Align) justifyLine(line string) string {
	words := strings.Fields(line)
	if len(words) <= 1 {
		return line
	}
	
	totalChars := 0
	for _, word := range words {
		totalChars += utf8.RuneCountInString(word)
	}
	
	totalSpaces := a.width - totalChars
	gaps := len(words) - 1
	
	if gaps <= 0 || totalSpaces <= 0 {
		return line
	}
	
	spacePerGap := totalSpaces / gaps
	extraSpaces := totalSpaces % gaps
	
	var result strings.Builder
	for i, word := range words {
		result.WriteString(word)
		if i < len(words)-1 {
			spaces := spacePerGap
			if i < extraSpaces {
				spaces++
			}
			result.WriteString(strings.Repeat(" ", spaces))
		}
	}
	
	return result.String()
}

// Left 左对齐文本
func Left(content string, width int) string {
	return NewAlign(content, WithAlignment(AlignLeft), WithWidth(width)).Render()
}

// Right 右对齐文本
func Right(content string, width int) string {
	return NewAlign(content, WithAlignment(AlignRight), WithWidth(width)).Render()
}

// Center 居中对齐文本
func Center(content string, width int) string {
	return NewAlign(content, WithAlignment(AlignCenter), WithWidth(width)).Render()
}

// Justify 两端对齐文本
func Justify(content string, width int) string {
	return NewAlign(content, WithAlignment(AlignJustify), WithWidth(width)).Render()
}