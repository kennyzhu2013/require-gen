package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Table 表格组件
//
// Table提供了结构化数据展示功能，类似于Python Rich库的Table组件。
// 支持表头、数据行、自动列宽调整、颜色配置等功能。
//
// 特性：
// - 自动列宽计算
// - 表头和数据行支持
// - 可配置的边框样式
// - 颜色和样式自定义
// - 对齐方式控制
// - 行交替颜色
type Table struct {
	headers      []string        // 表头
	rows         [][]string      // 数据行
	columnWidths []int          // 列宽度
	headerColor  *color.Color   // 表头颜色
	rowColors    []*color.Color // 行颜色（交替使用）
	borderColor  *color.Color   // 边框颜色
	showBorder   bool           // 是否显示边框
	showHeader   bool           // 是否显示表头
	minWidth     int            // 最小宽度
	maxWidth     int            // 最大宽度
	alignment     []TableAlignment    // 列对齐方式
}

// TableAlignment 表格对齐方式
type TableAlignment int

const (
	TableAlignLeft TableAlignment = iota
	TableAlignCenter
	TableAlignRight
)

// TableOption Table配置选项函数类型
type TableOption func(*Table)

// WithHeaderColor 设置表头颜色
func WithHeaderColor(color *color.Color) TableOption {
	return func(t *Table) {
		t.headerColor = color
	}
}

// WithRowColors 设置行颜色
func WithRowColors(colors ...*color.Color) TableOption {
	return func(t *Table) {
		t.rowColors = colors
	}
}

// WithBorderColor 设置边框颜色
func WithBorderColor(color *color.Color) TableOption {
	return func(t *Table) {
		t.borderColor = color
	}
}

// WithBorder 设置是否显示边框
func WithBorder(show bool) TableOption {
	return func(t *Table) {
		t.showBorder = show
	}
}

// WithHeader 设置是否显示表头
func WithHeader(show bool) TableOption {
	return func(t *Table) {
		t.showHeader = show
	}
}

// WithColumnAlignment 设置列对齐方式
func WithColumnAlignment(alignments ...TableAlignment) TableOption {
	return func(t *Table) {
		t.alignment = alignments
	}
}

// WithWidthLimits 设置宽度限制
func WithWidthLimits(min, max int) TableOption {
	return func(t *Table) {
		t.minWidth = min
		t.maxWidth = max
	}
}

// NewTable 创建新的Table实例
//
// 参数：
// - options: 配置选项
//
// 返回值：
// - *Table: Table实例
func NewTable(options ...TableOption) *Table {
	table := &Table{
		headers:     make([]string, 0),
		rows:        make([][]string, 0),
		columnWidths: make([]int, 0),
		headerColor: color.New(color.FgCyan, color.Bold),
		rowColors:   []*color.Color{color.New(color.FgWhite), color.New(color.FgHiBlack)},
		borderColor: color.New(color.FgHiBlack),
		showBorder:  true,
		showHeader:  true,
		minWidth:    0,
		maxWidth:    120,
		alignment:   make([]TableAlignment, 0),
	}
	
	// 应用配置选项
	for _, opt := range options {
		opt(table)
	}
	
	return table
}

// SetHeaders 设置表头
//
// 参数：
// - headers: 表头列表
//
// 返回值：
// - *Table: 返回自身以支持链式调用
func (t *Table) SetHeaders(headers ...string) *Table {
	t.headers = headers
	t.calculateColumnWidths()
	t.ensureAlignment()
	return t
}

// AddRow 添加数据行
//
// 参数：
// - row: 行数据
//
// 返回值：
// - *Table: 返回自身以支持链式调用
func (t *Table) AddRow(row ...string) *Table {
	t.rows = append(t.rows, row)
	t.calculateColumnWidths()
	t.ensureAlignment()
	return t
}

// AddRows 批量添加数据行
//
// 参数：
// - rows: 行数据列表
//
// 返回值：
// - *Table: 返回自身以支持链式调用
func (t *Table) AddRows(rows [][]string) *Table {
	t.rows = append(t.rows, rows...)
	t.calculateColumnWidths()
	t.ensureAlignment()
	return t
}

// Clear 清空表格数据
func (t *Table) Clear() {
	t.headers = make([]string, 0)
	t.rows = make([][]string, 0)
	t.columnWidths = make([]int, 0)
	t.alignment = make([]TableAlignment, 0)
}

// calculateColumnWidths 计算列宽度
func (t *Table) calculateColumnWidths() {
	if len(t.headers) == 0 && len(t.rows) == 0 {
		return
	}
	
	colCount := len(t.headers)
	if colCount == 0 && len(t.rows) > 0 {
		colCount = len(t.rows[0])
	}
	
	t.columnWidths = make([]int, colCount)
	
	// 计算表头宽度
	for i, header := range t.headers {
		if i < len(t.columnWidths) {
			t.columnWidths[i] = getDisplayWidth(header)
		}
	}
	
	// 计算数据行宽度
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(t.columnWidths) {
				width := getDisplayWidth(cell)
				if width > t.columnWidths[i] {
					t.columnWidths[i] = width
				}
			}
		}
	}
	
	// 应用最小宽度限制
	for i := range t.columnWidths {
		if t.columnWidths[i] < 3 {
			t.columnWidths[i] = 3
		}
	}
}

// ensureAlignment 确保对齐方式数组长度正确
func (t *Table) ensureAlignment() {
	colCount := len(t.columnWidths)
	if len(t.alignment) < colCount {
		// 扩展对齐方式数组，默认左对齐
		for len(t.alignment) < colCount {
			t.alignment = append(t.alignment, TableAlignLeft)
		}
	}
}

// Render 渲染表格到终端
func (t *Table) Render() string {
	if len(t.headers) == 0 && len(t.rows) == 0 {
		return ""
	}
	
	var builder strings.Builder
	
	// 渲染顶部边框
	if t.showBorder {
		builder.WriteString(t.renderTopBorderString())
	}
	
	// 渲染表头
	if t.showHeader && len(t.headers) > 0 {
		builder.WriteString(t.renderHeaderString())
		if t.showBorder {
			builder.WriteString(t.renderSeparatorBorderString())
		}
	}
	
	// 渲染数据行
	for i, row := range t.rows {
		builder.WriteString(t.renderRowString(row, i))
	}
	
	// 渲染底部边框
	if t.showBorder {
		builder.WriteString(t.renderBottomBorderString())
	}
	
	return builder.String()
}

// renderTopBorderString 渲染顶部边框并返回字符串
func (t *Table) renderTopBorderString() string {
	var border strings.Builder
	border.WriteString("┌")
	
	for i, width := range t.columnWidths {
		border.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			border.WriteString("┬")
		}
	}
	
	border.WriteString("┐\n")
	return t.borderColor.Sprint(border.String())
}

// renderTopBorder 渲染顶部边框
func (t *Table) renderTopBorder() {
	var border strings.Builder
	border.WriteString("┌")
	
	for i, width := range t.columnWidths {
		border.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			border.WriteString("┬")
		}
	}
	
	border.WriteString("┐")
	t.borderColor.Println(border.String())
}

// renderBottomBorderString 渲染底部边框并返回字符串
func (t *Table) renderBottomBorderString() string {
	var border strings.Builder
	border.WriteString("└")
	
	for i, width := range t.columnWidths {
		border.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			border.WriteString("┴")
		}
	}
	
	border.WriteString("┘\n")
	return t.borderColor.Sprint(border.String())
}

// renderBottomBorder 渲染底部边框
func (t *Table) renderBottomBorder() {
	var border strings.Builder
	border.WriteString("└")
	
	for i, width := range t.columnWidths {
		border.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			border.WriteString("┴")
		}
	}
	
	border.WriteString("┘")
	t.borderColor.Println(border.String())
}

// renderSeparatorBorderString 渲染分隔边框并返回字符串
func (t *Table) renderSeparatorBorderString() string {
	var border strings.Builder
	border.WriteString("├")
	
	for i, width := range t.columnWidths {
		border.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			border.WriteString("┼")
		}
	}
	
	border.WriteString("┤\n")
	return t.borderColor.Sprint(border.String())
}

// renderSeparatorBorder 渲染分隔边框
func (t *Table) renderSeparatorBorder() {
	var border strings.Builder
	border.WriteString("├")
	
	for i, width := range t.columnWidths {
		border.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			border.WriteString("┼")
		}
	}
	
	border.WriteString("┤")
	t.borderColor.Println(border.String())
}

// renderHeaderString 渲染表头并返回字符串
func (t *Table) renderHeaderString() string {
	var line strings.Builder
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	for i, header := range t.headers {
		if i < len(t.columnWidths) {
			cell := t.formatCell(header, t.columnWidths[i], t.alignment[i])
			line.WriteString(" ")
			line.WriteString(t.headerColor.Sprint(cell))
			line.WriteString(" ")
			
			if t.showBorder && i < len(t.headers)-1 {
				line.WriteString(t.borderColor.Sprint("│"))
			}
		}
	}
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	line.WriteString("\n")
	return line.String()
}

// renderHeader 渲染表头
func (t *Table) renderHeader() {
	var line strings.Builder
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	for i, header := range t.headers {
		if i < len(t.columnWidths) {
			cell := t.formatCell(header, t.columnWidths[i], t.alignment[i])
			line.WriteString(" ")
			line.WriteString(t.headerColor.Sprint(cell))
			line.WriteString(" ")
			
			if t.showBorder && i < len(t.headers)-1 {
				line.WriteString(t.borderColor.Sprint("│"))
			}
		}
	}
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	fmt.Println(line.String())
}

// renderRowString 渲染数据行并返回字符串
func (t *Table) renderRowString(row []string, rowIndex int) string {
	var line strings.Builder
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	// 选择行颜色
	rowColor := t.rowColors[rowIndex%len(t.rowColors)]
	
	for i, cell := range row {
		if i < len(t.columnWidths) {
			formattedCell := t.formatCell(cell, t.columnWidths[i], t.alignment[i])
			line.WriteString(" ")
			line.WriteString(rowColor.Sprint(formattedCell))
			line.WriteString(" ")
			
			if t.showBorder && i < len(row)-1 {
				line.WriteString(t.borderColor.Sprint("│"))
			}
		}
	}
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	line.WriteString("\n")
	return line.String()
}

// renderRow 渲染数据行
func (t *Table) renderRow(row []string, rowIndex int) {
	var line strings.Builder
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	// 选择行颜色
	rowColor := t.rowColors[rowIndex%len(t.rowColors)]
	
	for i, cell := range row {
		if i < len(t.columnWidths) {
			formattedCell := t.formatCell(cell, t.columnWidths[i], t.alignment[i])
			line.WriteString(" ")
			line.WriteString(rowColor.Sprint(formattedCell))
			line.WriteString(" ")
			
			if t.showBorder && i < len(row)-1 {
				line.WriteString(t.borderColor.Sprint("│"))
			}
		}
	}
	
	if t.showBorder {
		line.WriteString(t.borderColor.Sprint("│"))
	}
	
	fmt.Println(line.String())
}

// formatCell 格式化单元格内容
func (t *Table) formatCell(content string, width int, align TableAlignment) string {
	contentWidth := getDisplayWidth(content)
	
	if contentWidth >= width {
		// 内容过长，截断
		if width > 3 {
			return content[:width-3] + "..."
		}
		return content[:width]
	}
	
	padding := width - contentWidth
	
	switch align {
	case TableAlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad)
	case TableAlignRight:
		return strings.Repeat(" ", padding) + content
	default: // TableAlignLeft
		return content + strings.Repeat(" ", padding)
	}
}

// GetRowCount 获取行数
func (t *Table) GetRowCount() int {
	return len(t.rows)
}

// GetColumnCount 获取列数
func (t *Table) GetColumnCount() int {
	return len(t.columnWidths)
}

// CreateDataTable 创建数据表格的便捷函数
//
// 参数：
// - headers: 表头
// - data: 数据行
//
// 返回值：
// - *Table: 配置好的表格实例
func CreateDataTable(headers []string, data [][]string) *Table {
	table := NewTable()
	table.SetHeaders(headers...)
	table.AddRows(data)
	return table
}

// CreateSimpleTable 创建简单表格的便捷函数
//
// 参数：
// - data: 键值对数据
//
// 返回值：
// - *Table: 配置好的表格实例
func CreateSimpleTable(data map[string]string) *Table {
	table := NewTable(
		WithBorder(true),
		WithHeader(false),
	)
	
	for key, value := range data {
		table.AddRow(key, value)
	}
	
	return table
}