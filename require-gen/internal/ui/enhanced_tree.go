package ui

import (
	"fmt"

	"github.com/fatih/color"
)

// TreeNode 树节点结构
//
// TreeNode表示树形结构中的一个节点，包含标签、状态、详细信息等
type TreeNode struct {
	label    string      // 节点标签
	status   string      // 节点状态
	detail   string      // 详细信息
	children []*TreeNode // 子节点
}

// Tree 树形显示组件
//
// Tree提供了层次化的信息展示，类似于Rich库的Tree组件。
// 支持多种状态指示、颜色编码、层次结构等功能。
//
// 特性：
// - 多种状态指示符
// - 颜色编码状态
// - 层次化结构
// - 详细信息显示
// - 自定义样式
type Tree struct {
	title     string      // 树标题
	nodes     []*TreeNode // 根节点列表
	style     TreeStyle   // 树样式配置
}

// TreeStyle 树样式配置
type TreeStyle struct {
	TitleColor    color.Attribute // 标题颜色
	GuideStyle    color.Attribute // 连接线颜色
	ShowDetails   bool           // 是否显示详细信息
	CompactMode   bool           // 紧凑模式
}

// TreeOption Tree配置选项函数类型
type TreeOption func(*Tree)

// WithTreeStyle 设置树样式
func WithTreeStyle(style TreeStyle) TreeOption {
	return func(t *Tree) {
		t.style = style
	}
}

// WithTitleColor 设置标题颜色
func WithTitleColor(color color.Attribute) TreeOption {
	return func(t *Tree) {
		t.style.TitleColor = color
	}
}

// WithCompactMode 设置紧凑模式
func WithCompactMode(compact bool) TreeOption {
	return func(t *Tree) {
		t.style.CompactMode = compact
	}
}

// NewTree 创建新的Tree实例
//
// 参数：
// - title: 树标题
// - options: 配置选项
//
// 返回值：
// - *Tree: Tree实例
func NewTree(title string, options ...TreeOption) *Tree {
	tree := &Tree{
		title: title,
		nodes: make([]*TreeNode, 0),
		style: TreeStyle{
			TitleColor:  color.FgCyan,
			GuideStyle:  color.FgWhite,
			ShowDetails: true,
			CompactMode: false,
		},
	}
	
	// 应用配置选项
	for _, opt := range options {
		opt(tree)
	}
	
	return tree
}

// Add 添加根节点
//
// 参数：
// - label: 节点标签
// - status: 节点状态
// - detail: 详细信息
//
// 返回值：
// - *TreeNode: 添加的节点
func (t *Tree) Add(label, status, detail string) *TreeNode {
	node := &TreeNode{
		label:    label,
		status:   status,
		detail:   detail,
		children: make([]*TreeNode, 0),
	}
	t.nodes = append(t.nodes, node)
	return node
}

// AddChild 为节点添加子节点
//
// 参数：
// - parent: 父节点
// - label: 子节点标签
// - status: 子节点状态
// - detail: 详细信息
//
// 返回值：
// - *TreeNode: 添加的子节点
func (t *Tree) AddChild(parent *TreeNode, label, status, detail string) *TreeNode {
	child := &TreeNode{
		label:    label,
		status:   status,
		detail:   detail,
		children: make([]*TreeNode, 0),
	}
	parent.children = append(parent.children, child)
	return child
}

// Render 渲染Tree到终端
//
// 该方法将Tree渲染为层次化的文本结构，包括：
// - 标题行
// - 根节点和子节点
// - 状态指示符
// - 连接线
// - 详细信息
func (t *Tree) Render() {
	// 渲染标题
	if t.title != "" {
		titleColor := color.New(t.style.TitleColor, color.Bold)
		titleColor.Println(t.title)
		if !t.style.CompactMode {
			fmt.Println()
		}
	}
	
	// 渲染根节点
	for i, node := range t.nodes {
		isLast := i == len(t.nodes)-1
		t.renderNode(node, "", isLast, 0)
	}
}

// renderNode 渲染单个节点
//
// 参数：
// - node: 要渲染的节点
// - prefix: 前缀字符串（用于层次缩进）
// - isLast: 是否为最后一个节点
// - depth: 节点深度
func (t *Tree) renderNode(node *TreeNode, prefix string, isLast bool, depth int) {
	// 选择连接符
	var connector string
	var childPrefix string
	
	if depth == 0 {
		// 根节点
		if isLast {
			connector = "└── "
			childPrefix = "    "
		} else {
			connector = "├── "
			childPrefix = "│   "
		}
	} else {
		// 子节点
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		} else {
			connector = "├── "
			childPrefix = prefix + "│   "
		}
	}
	
	// 渲染连接线
	guideColor := color.New(t.style.GuideStyle)
	fmt.Print(prefix)
	guideColor.Print(connector)
	
	// 渲染状态符号
	symbol, symbolColor := t.getStatusSymbol(node.status)
	symbolColor.Print(symbol + " ")
	
	// 渲染标签
	t.renderNodeLabel(node)
	
	// 渲染详细信息
	if t.style.ShowDetails && node.detail != "" {
		faintColor := color.New(color.Faint)
		faintColor.Printf(" (%s)", node.detail)
	}
	
	fmt.Println()
	
	// 渲染子节点
	for i, child := range node.children {
		isChildLast := i == len(node.children)-1
		t.renderNode(child, childPrefix, isChildLast, depth+1)
	}
}

// getStatusSymbol 获取状态符号和颜色
//
// 参数：
// - status: 状态字符串
//
// 返回值：
// - string: 状态符号
// - *color.Color: 符号颜色
func (t *Tree) getStatusSymbol(status string) (string, *color.Color) {
	switch status {
	case "done", "completed", "success":
		return "●", color.New(color.FgGreen, color.Bold)
	case "running", "in_progress", "active":
		return "◐", color.New(color.FgCyan, color.Bold)
	case "pending", "waiting":
		return "○", color.New(color.FgWhite, color.Faint)
	case "error", "failed":
		return "●", color.New(color.FgRed, color.Bold)
	case "skipped", "cancelled":
		return "○", color.New(color.FgYellow, color.Bold)
	case "warning":
		return "◐", color.New(color.FgYellow, color.Bold)
	default:
		return "○", color.New(color.FgWhite)
	}
}

// renderNodeLabel 渲染节点标签
//
// 根据节点状态应用不同的文本样式
//
// 参数：
// - node: 要渲染的节点
func (t *Tree) renderNodeLabel(node *TreeNode) {
	switch node.status {
	case "done", "completed", "success":
		color.New(color.FgWhite, color.Bold).Print(node.label)
	case "running", "in_progress", "active":
		color.New(color.FgCyan, color.Bold).Print(node.label)
	case "pending", "waiting":
		color.New(color.Faint).Print(node.label)
	case "error", "failed":
		color.New(color.FgRed, color.Bold).Print(node.label)
	case "skipped", "cancelled":
		color.New(color.FgYellow).Print(node.label)
	case "warning":
		color.New(color.FgYellow, color.Bold).Print(node.label)
	default:
		fmt.Print(node.label)
	}
}

// UpdateNodeStatus 更新节点状态
//
// 参数：
// - node: 要更新的节点
// - status: 新状态
// - detail: 新的详细信息（可选）
func (t *Tree) UpdateNodeStatus(node *TreeNode, status string, detail ...string) {
	node.status = status
	if len(detail) > 0 {
		node.detail = detail[0]
	}
}

// FindNode 查找节点
//
// 参数：
// - label: 节点标签
//
// 返回值：
// - *TreeNode: 找到的节点，未找到返回nil
func (t *Tree) FindNode(label string) *TreeNode {
	for _, node := range t.nodes {
		if found := t.findNodeRecursive(node, label); found != nil {
			return found
		}
	}
	return nil
}

// findNodeRecursive 递归查找节点
func (t *Tree) findNodeRecursive(node *TreeNode, label string) *TreeNode {
	if node.label == label {
		return node
	}
	
	for _, child := range node.children {
		if found := t.findNodeRecursive(child, label); found != nil {
			return found
		}
	}
	
	return nil
}

// Clear 清空所有节点
func (t *Tree) Clear() {
	t.nodes = make([]*TreeNode, 0)
}

// GetNodeCount 获取节点总数
//
// 返回值：
// - int: 节点总数（包括子节点）
func (t *Tree) GetNodeCount() int {
	count := 0
	for _, node := range t.nodes {
		count += t.countNodeRecursive(node)
	}
	return count
}

// countNodeRecursive 递归计算节点数量
func (t *Tree) countNodeRecursive(node *TreeNode) int {
	count := 1 // 当前节点
	for _, child := range node.children {
		count += t.countNodeRecursive(child)
	}
	return count
}

// CreateStepTree 创建步骤跟踪树的便捷函数
//
// 参数：
// - title: 树标题
// - steps: 步骤列表
//
// 返回值：
// - *Tree: 配置好的步骤树
func CreateStepTree(title string, steps []string) *Tree {
	tree := NewTree(title,
		WithTitleColor(color.FgCyan),
		WithCompactMode(false))
	
	for _, step := range steps {
		tree.Add(step, "pending", "")
	}
	
	return tree
}

// CreateProgressTree 创建进度跟踪树的便捷函数
//
// 参数：
// - title: 树标题
//
// 返回值：
// - *Tree: 配置好的进度树
func CreateProgressTree(title string) *Tree {
	return NewTree(title,
		WithTitleColor(color.FgHiCyan),
		WithCompactMode(false))
}