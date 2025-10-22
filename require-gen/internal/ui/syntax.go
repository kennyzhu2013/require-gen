package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// SyntaxHighlighter 语法高亮器
type SyntaxHighlighter struct {
	language string
	theme    *SyntaxTheme
}

// SyntaxTheme 语法高亮主题
type SyntaxTheme struct {
	Keyword    *color.Color
	String     *color.Color
	Comment    *color.Color
	Number     *color.Color
	Function   *color.Color
	Type       *color.Color
	Variable   *color.Color
	Operator   *color.Color
	Bracket    *color.Color
	Background *color.Color
}

// DefaultSyntaxTheme 默认语法高亮主题
var DefaultSyntaxTheme = &SyntaxTheme{
	Keyword:    color.New(color.FgMagenta, color.Bold),
	String:     color.New(color.FgGreen),
	Comment:    color.New(color.FgHiBlack),
	Number:     color.New(color.FgYellow),
	Function:   color.New(color.FgBlue, color.Bold),
	Type:       color.New(color.FgCyan, color.Bold),
	Variable:   color.New(color.FgWhite),
	Operator:   color.New(color.FgRed),
	Bracket:    color.New(color.FgHiWhite),
	Background: color.New(color.BgBlack),
}

// DarkSyntaxTheme 暗色语法高亮主题
var DarkSyntaxTheme = &SyntaxTheme{
	Keyword:    color.New(color.FgHiMagenta, color.Bold),
	String:     color.New(color.FgHiGreen),
	Comment:    color.New(color.FgHiBlack),
	Number:     color.New(color.FgHiYellow),
	Function:   color.New(color.FgHiBlue, color.Bold),
	Type:       color.New(color.FgHiCyan, color.Bold),
	Variable:   color.New(color.FgHiWhite),
	Operator:   color.New(color.FgHiRed),
	Bracket:    color.New(color.FgWhite),
	Background: color.New(color.BgHiBlack),
}

// LanguageRules 语言规则
type LanguageRules struct {
	Keywords  []string
	Operators []string
	Brackets  []string
	Comments  []string
}

// 预定义语言规则
var languageRules = map[string]*LanguageRules{
	"go": {
		Keywords: []string{
			"break", "case", "chan", "const", "continue", "default", "defer",
			"else", "fallthrough", "for", "func", "go", "goto", "if",
			"import", "interface", "map", "package", "range", "return",
			"select", "struct", "switch", "type", "var",
		},
		Operators: []string{"++", "--", "==", "!=", "<=", ">=", "&&", "||", ":=", "...", "->"},
		Brackets:  []string{"(", ")", "[", "]", "{", "}"},
		Comments:  []string{"//", "/*", "*/"},
	},
	"python": {
		Keywords: []string{
			"and", "as", "assert", "break", "class", "continue", "def",
			"del", "elif", "else", "except", "exec", "finally", "for",
			"from", "global", "if", "import", "in", "is", "lambda",
			"not", "or", "pass", "print", "raise", "return", "try",
			"while", "with", "yield",
		},
		Operators: []string{"==", "!=", "<=", ">=", "and", "or", "not", "in", "is"},
		Brackets:  []string{"(", ")", "[", "]", "{", "}"},
		Comments:  []string{"#", "'''", `"""`},
	},
	"javascript": {
		Keywords: []string{
			"break", "case", "catch", "class", "const", "continue", "debugger",
			"default", "delete", "do", "else", "export", "extends", "finally",
			"for", "function", "if", "import", "in", "instanceof", "let",
			"new", "return", "super", "switch", "this", "throw", "try",
			"typeof", "var", "void", "while", "with", "yield",
		},
		Operators: []string{"===", "!==", "==", "!=", "<=", ">=", "&&", "||", "++", "--", "=>"},
		Brackets:  []string{"(", ")", "[", "]", "{", "}"},
		Comments:  []string{"//", "/*", "*/"},
	},
}

// NewSyntaxHighlighter 创建新的语法高亮器
func NewSyntaxHighlighter(language string, theme *SyntaxTheme) *SyntaxHighlighter {
	if theme == nil {
		theme = DefaultSyntaxTheme
	}
	
	return &SyntaxHighlighter{
		language: language,
		theme:    theme,
	}
}

// Highlight 高亮代码
func (sh *SyntaxHighlighter) Highlight(code string) string {
	rules, exists := languageRules[sh.language]
	if !exists {
		return code // 不支持的语言，返回原始代码
	}
	
	result := code
	
	// 高亮字符串
	result = sh.highlightStrings(result)
	
	// 高亮注释
	result = sh.highlightComments(result, rules.Comments)
	
	// 高亮关键字
	result = sh.highlightKeywords(result, rules.Keywords)
	
	// 高亮数字
	result = sh.highlightNumbers(result)
	
	// 高亮操作符
	result = sh.highlightOperators(result, rules.Operators)
	
	// 高亮括号
	result = sh.highlightBrackets(result, rules.Brackets)
	
	return result
}

// highlightStrings 高亮字符串
func (sh *SyntaxHighlighter) highlightStrings(code string) string {
	// 双引号字符串
	doubleQuoteRegex := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	code = doubleQuoteRegex.ReplaceAllStringFunc(code, func(match string) string {
		return sh.theme.String.Sprint(match)
	})
	
	// 单引号字符串
	singleQuoteRegex := regexp.MustCompile(`'([^'\\]|\\.)*'`)
	code = singleQuoteRegex.ReplaceAllStringFunc(code, func(match string) string {
		return sh.theme.String.Sprint(match)
	})
	
	// 反引号字符串（Go语言）
	if sh.language == "go" {
		backQuoteRegex := regexp.MustCompile("`[^`]*`")
		code = backQuoteRegex.ReplaceAllStringFunc(code, func(match string) string {
			return sh.theme.String.Sprint(match)
		})
	}
	
	return code
}

// highlightComments 高亮注释
func (sh *SyntaxHighlighter) highlightComments(code string, comments []string) string {
	for _, comment := range comments {
		switch comment {
		case "//":
			// 单行注释
			regex := regexp.MustCompile(`//.*$`)
			code = regex.ReplaceAllStringFunc(code, func(match string) string {
				return sh.theme.Comment.Sprint(match)
			})
		case "/*":
			// 多行注释开始
			regex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
			code = regex.ReplaceAllStringFunc(code, func(match string) string {
				return sh.theme.Comment.Sprint(match)
			})
		case "#":
			// Python风格注释
			regex := regexp.MustCompile(`#.*$`)
			code = regex.ReplaceAllStringFunc(code, func(match string) string {
				return sh.theme.Comment.Sprint(match)
			})
		}
	}
	return code
}

// highlightKeywords 高亮关键字
func (sh *SyntaxHighlighter) highlightKeywords(code string, keywords []string) string {
	for _, keyword := range keywords {
		// 使用单词边界确保完整匹配
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(keyword))
		regex := regexp.MustCompile(pattern)
		code = regex.ReplaceAllStringFunc(code, func(match string) string {
			return sh.theme.Keyword.Sprint(match)
		})
	}
	return code
}

// highlightNumbers 高亮数字
func (sh *SyntaxHighlighter) highlightNumbers(code string) string {
	// 匹配整数和浮点数
	numberRegex := regexp.MustCompile(`\b\d+\.?\d*\b`)
	code = numberRegex.ReplaceAllStringFunc(code, func(match string) string {
		return sh.theme.Number.Sprint(match)
	})
	return code
}

// highlightOperators 高亮操作符
func (sh *SyntaxHighlighter) highlightOperators(code string, operators []string) string {
	for _, op := range operators {
		// 转义特殊字符
		escaped := regexp.QuoteMeta(op)
		regex := regexp.MustCompile(escaped)
		code = regex.ReplaceAllStringFunc(code, func(match string) string {
			return sh.theme.Operator.Sprint(match)
		})
	}
	return code
}

// highlightBrackets 高亮括号
func (sh *SyntaxHighlighter) highlightBrackets(code string, brackets []string) string {
	for _, bracket := range brackets {
		escaped := regexp.QuoteMeta(bracket)
		regex := regexp.MustCompile(escaped)
		code = regex.ReplaceAllStringFunc(code, func(match string) string {
			return sh.theme.Bracket.Sprint(match)
		})
	}
	return code
}

// Syntax 语法高亮组件
type Syntax struct {
	code        string
	language    string
	theme       *SyntaxTheme
	lineNumbers bool
	background  bool
}

// SyntaxOption 语法高亮选项函数类型
type SyntaxOption func(*Syntax)

// WithSyntaxTheme 设置语法高亮主题
func WithSyntaxTheme(theme *SyntaxTheme) SyntaxOption {
	return func(s *Syntax) {
		s.theme = theme
	}
}

// WithLineNumbers 设置是否显示行号
func WithLineNumbers(show bool) SyntaxOption {
	return func(s *Syntax) {
		s.lineNumbers = show
	}
}

// WithBackground 设置是否显示背景
func WithBackground(show bool) SyntaxOption {
	return func(s *Syntax) {
		s.background = show
	}
}

// NewSyntax 创建新的语法高亮组件
func NewSyntax(code, language string, options ...SyntaxOption) *Syntax {
	syntax := &Syntax{
		code:        code,
		language:    language,
		theme:       DefaultSyntaxTheme,
		lineNumbers: false,
		background:  false,
	}
	
	for _, option := range options {
		option(syntax)
	}
	
	return syntax
}

// Render 渲染语法高亮的代码
func (s *Syntax) Render() string {
	highlighter := NewSyntaxHighlighter(s.language, s.theme)
	highlighted := highlighter.Highlight(s.code)
	
	if s.lineNumbers {
		highlighted = s.addLineNumbers(highlighted)
	}
	
	if s.background && s.theme.Background != nil {
		lines := strings.Split(highlighted, "\n")
		for i, line := range lines {
			lines[i] = s.theme.Background.Sprint(line)
		}
		highlighted = strings.Join(lines, "\n")
	}
	
	return highlighted
}

// addLineNumbers 添加行号
func (s *Syntax) addLineNumbers(code string) string {
	lines := strings.Split(code, "\n")
	maxWidth := len(fmt.Sprintf("%d", len(lines)))
	
	for i, line := range lines {
		lineNum := fmt.Sprintf("%*d", maxWidth, i+1)
		lineNumColored := color.New(color.FgHiBlack).Sprint(lineNum)
		lines[i] = fmt.Sprintf("%s │ %s", lineNumColored, line)
	}
	
	return strings.Join(lines, "\n")
}

// HighlightCode 快速语法高亮函数
func HighlightCode(code, language string) string {
	return NewSyntax(code, language).Render()
}

// HighlightCodeWithLineNumbers 带行号的快速语法高亮函数
func HighlightCodeWithLineNumbers(code, language string) string {
	return NewSyntax(code, language, WithLineNumbers(true)).Render()
}