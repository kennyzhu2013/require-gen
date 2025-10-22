package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"specify-cli/internal/ui"

	"github.com/fatih/color"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "ui-demo" {
		runUIDemo()
		return
	}

	// 原有的main逻辑保持不变
	fmt.Println("Use 'specify ui-demo' to run UI components demonstration")
}

func runUIDemo() {
	fmt.Println("🎨 Go UI Components Demo - Rich Library Experience")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println()

	// 创建UI管理器
	uiManager := ui.NewUIManager()

	// 演示1: Panel组件
	demoPanel(uiManager)

	// 演示2: Table组件
	demoTable(uiManager)

	// 演示3: Tree组件
	demoTree(uiManager)

	// 演示4: Progress组件（多列和嵌套）
	demoProgress(uiManager)

	// 演示5: Spinner组件
	demoSpinner(uiManager)

	// 演示6: Live组件
	demoLive(uiManager)

	// 演示7: Syntax高亮
	demoSyntax()

	// 演示8: Align对齐
	demoAlign()

	fmt.Println("\n✨ Demo completed! All UI components are working perfectly.")
}

func demoPanel(uiManager *ui.UIManager) {
	fmt.Println("📦 Panel Component Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 基础面板
	panel1 := uiManager.CreatePanel(
		"This is a basic panel with border and title.\nIt supports multi-line content and automatic formatting.",
		"Basic Panel",
		ui.WithPanelPadding(2, 2),
		ui.WithMinWidth(50),
	)
	fmt.Println(panel1.Render())
	fmt.Println()

	// 嵌套面板
	innerContent := "Nested panel content\nwith multiple lines"
	innerPanel := uiManager.CreatePanel(innerContent, "Inner Panel", ui.WithPanelPadding(1, 1))

	outerContent := fmt.Sprintf("This panel contains another panel:\n\n%s\n\nPretty cool, right?", innerPanel.Render())
	outerPanel := uiManager.CreatePanel(outerContent, "Outer Panel", ui.WithPanelPadding(2, 2))
	fmt.Println(outerPanel.Render())
	fmt.Println()
}

func demoTable(uiManager *ui.UIManager) {
	fmt.Println("📊 Table Component Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 创建表格
	table := uiManager.CreateTable(
		ui.WithHeaderColor(color.New(color.FgCyan, color.Bold)),
		ui.WithBorderColor(color.New(color.FgHiBlack)),
		ui.WithRowColors(color.New(color.FgWhite), color.New(color.FgHiBlack)),
	)

	// 设置表头
	table.SetHeaders("Name", "Age", "City", "Occupation")

	// 添加数据行
	table.AddRow("Alice Johnson", "28", "New York", "Software Engineer")
	table.AddRow("Bob Smith", "34", "San Francisco", "Product Manager")
	table.AddRow("Carol Davis", "29", "Seattle", "UX Designer")
	table.AddRow("David Wilson", "31", "Austin", "DevOps Engineer")
	table.AddRow("Eva Brown", "26", "Boston", "Data Scientist")

	fmt.Println(table.Render())
	fmt.Println()
}

func demoTree(uiManager *ui.UIManager) {
	fmt.Println("🌳 Tree Component Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 创建树形结构
	tree := ui.NewTree("Project Structure")

	// 添加根节点
	root := tree.Add("src/", "directory", "Source code directory")

	// 添加子节点
	main := tree.AddChild(root, "main/", "directory", "Main package directory")
	tree.AddChild(main, "main.go", "file", "Main entry point")
	tree.AddChild(main, "config.go", "file", "Configuration file")

	internal := tree.AddChild(root, "internal/", "directory", "Internal packages")
	ui_pkg := tree.AddChild(internal, "ui/", "directory", "UI components")
	tree.AddChild(ui_pkg, "panel.go", "file", "Panel component")
	tree.AddChild(ui_pkg, "table.go", "file", "Table component")
	tree.AddChild(ui_pkg, "progress.go", "file", "Progress component")

	tests := tree.AddChild(root, "tests/", "directory", "Test files")
	tree.AddChild(tests, "ui_test.go", "file", "UI tests")
	tree.AddChild(tests, "integration_test.go", "file", "Integration tests")

	// 添加一些状态指示
	tree.AddChild(root, "README.md", "modified", "Documentation file")
	tree.AddChild(root, "go.mod", "new", "Go module file")
	tree.AddChild(root, "go.sum", "new", "Go dependencies")

	tree.Render()
	fmt.Println()
}

func demoProgress(uiManager *ui.UIManager) {
	fmt.Println("📈 Progress Component Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 演示多列进度条
	fmt.Println("Multi-column Progress Bar:")

	// 定义列格式化器
	nameFormatter := func(pb *ui.ProgressBar) string {
		return pb.GetDescription()
	}

	barFormatter := func(pb *ui.ProgressBar) string {
		percentage := float64(pb.GetCurrent()) / float64(pb.GetTotal()) * 100
		filled := int(percentage / 5) // 20个字符的进度条
		bar := strings.Repeat("█", filled) + strings.Repeat("░", 20-filled)
		return fmt.Sprintf("[%s]", bar)
	}

	percentFormatter := func(pb *ui.ProgressBar) string {
		percentage := float64(pb.GetCurrent()) / float64(pb.GetTotal()) * 100
		return fmt.Sprintf("%.1f%%", percentage)
	}

	speedFormatter := func(pb *ui.ProgressBar) string {
		return fmt.Sprintf("%.1f/s", pb.GetSpeed())
	}

	// 创建多列进度条
	columns := []ui.ProgressColumn{
		{Name: "Task", Width: 20, Alignment: ui.AlignLeft, Formatter: nameFormatter},
		{Name: "Progress", Width: 22, Alignment: ui.AlignCenter, Formatter: barFormatter},
		{Name: "Percent", Width: 8, Alignment: ui.AlignRight, Formatter: percentFormatter},
		{Name: "Speed", Width: 10, Alignment: ui.AlignRight, Formatter: speedFormatter},
	}

	// 创建多进度条管理器
	multiProgress := ui.NewMultiProgressBar()

	// 创建主任务进度条
	mainTask := uiManager.CreateProgressBar(100, "Main Task",
		ui.WithColumns(columns...),
		ui.WithDisplayOptions(false, false, false, false),
	)
	multiProgress.AddBar(mainTask)

	// 创建子任务进度条
	subTask1 := uiManager.CreateProgressBar(50, "Subtask 1",
		ui.WithColumns(columns...),
		ui.WithParent(mainTask),
		ui.WithDisplayOptions(false, false, false, false),
	)
	multiProgress.AddBar(subTask1)

	subTask2 := uiManager.CreateProgressBar(30, "Subtask 2",
		ui.WithColumns(columns...),
		ui.WithParent(mainTask),
		ui.WithDisplayOptions(false, false, false, false),
	)
	multiProgress.AddBar(subTask2)

	// 模拟进度更新
	multiProgress.Start()

	for i := 0; i <= 100; i += 5 {
		mainTask.SetCurrent(int64(i))
		if i <= 50 {
			subTask1.SetCurrent(int64(i))
		}
		if i >= 20 && i <= 50 {
			subTask2.SetCurrent(int64(i - 20))
		}
		time.Sleep(100 * time.Millisecond)
	}

	multiProgress.Stop()
	fmt.Println()
}

func demoSpinner(uiManager *ui.UIManager) {
	fmt.Println("⏳ Spinner Component Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 演示不同样式的Spinner
	styles := []ui.SpinnerStyle{
		ui.SpinnerDots,
		ui.SpinnerLine,
		ui.SpinnerCircle,
		ui.SpinnerArrow,
		ui.SpinnerBounce,
	}

	styleNames := []string{"Dots", "Line", "Circle", "Arrow", "Bounce"}

	for i, style := range styles {
		fmt.Printf("Testing %s spinner: ", styleNames[i])

		spinner := uiManager.CreateSpinner(style,
			ui.WithSpinnerText(fmt.Sprintf("Loading %s...", styleNames[i])),
			ui.WithSpinnerColor(color.New(color.FgCyan)),
		)

		spinner.Start()
		time.Sleep(2 * time.Second)
		spinner.Stop()

		fmt.Println(" ✓ Done")
	}
	fmt.Println()
}

func demoLive(uiManager *ui.UIManager) {
	fmt.Println("🔄 Live Component Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 创建Live组件
	live := uiManager.CreateLive(
		ui.WithRefreshRate(200*time.Millisecond),
		ui.WithAutoRefresh(true),
	)

	live.Start()

	// 模拟实时数据更新
	for i := 0; i < 20; i++ {
		content := fmt.Sprintf("Live Update #%d\n", i+1)
		content += fmt.Sprintf("Timestamp: %s\n", time.Now().Format("15:04:05"))
		content += fmt.Sprintf("Random Value: %d\n", rand.Intn(1000))
		content += fmt.Sprintf("Progress: %s", strings.Repeat("█", i%10+1))

		live.Update(content)
		time.Sleep(300 * time.Millisecond)
	}

	live.Stop()
	fmt.Println("Live demo completed!")
	fmt.Println()
}

func demoSyntax() {
	fmt.Println("🎨 Syntax Highlighting Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	// Go代码示例
	goCode := `package main

import (
	"fmt"
	"time"
)

func main() {
	// This is a comment
	message := "Hello, World!"
	count := 42
	
	for i := 0; i < count; i++ {
		fmt.Println(message)
		time.Sleep(100 * time.Millisecond)
	}
}`

	fmt.Println("Go Code with Syntax Highlighting:")
	syntax := ui.NewSyntax(goCode, "go",
		ui.WithSyntaxTheme(ui.DefaultSyntaxTheme),
		ui.WithLineNumbers(true),
	)
	fmt.Println(syntax.Render())
	fmt.Println()

	// Python代码示例
	pythonCode := `def fibonacci(n):
    """Generate Fibonacci sequence up to n terms"""
    if n <= 0:
        return []
    elif n == 1:
        return [0]
    elif n == 2:
        return [0, 1]
    
    sequence = [0, 1]
    for i in range(2, n):
        sequence.append(sequence[i-1] + sequence[i-2])
    
    return sequence

# Main execution
if __name__ == "__main__":
    result = fibonacci(10)
    print(f"Fibonacci sequence: {result}")`

	fmt.Println("Python Code with Syntax Highlighting:")
	pythonSyntax := ui.NewSyntax(pythonCode, "python",
		ui.WithSyntaxTheme(ui.DarkSyntaxTheme),
		ui.WithLineNumbers(true),
	)
	fmt.Println(pythonSyntax.Render())
	fmt.Println()
}

func demoAlign() {
	fmt.Println("📐 Text Alignment Demo")
	fmt.Println("-" + strings.Repeat("-", 30))

	text := "This is a sample text for alignment demonstration"
	width := 60

	fmt.Println("Original text:")
	fmt.Printf("'%s'\n\n", text)

	fmt.Printf("Left aligned (width %d):\n", width)
	fmt.Printf("'%s'\n\n", ui.Left(text, width))

	fmt.Printf("Right aligned (width %d):\n", width)
	fmt.Printf("'%s'\n\n", ui.Right(text, width))

	fmt.Printf("Center aligned (width %d):\n", width)
	fmt.Printf("'%s'\n\n", ui.Center(text, width))

	fmt.Printf("Justified (width %d):\n", width)
	fmt.Printf("'%s'\n\n", ui.Justify(text, width))

	// 多行文本对齐
	multilineText := "Line 1: Short\nLine 2: This is a longer line\nLine 3: Medium length"
	fmt.Println("Multi-line text alignment:")
	fmt.Printf("Center aligned:\n%s\n\n", ui.Center(multilineText, 40))
}
