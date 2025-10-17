# Specify CLI - Go Framework

基于 `golanganalyse.md` 文档分析生成的 Go 语言 CLI 框架代码。

## 项目概述

Specify CLI 是一个用于初始化和管理 AI 助手项目模板的命令行工具。支持多种 AI 助手（如 GitHub Copilot、Claude、Gemini 等）和脚本类型（Shell、PowerShell）。

## 架构设计

项目采用四层架构设计：

```
├── cmd/specify/           # 应用入口
│   └── main.go           # 主程序入口
├── internal/             # 内部包
│   ├── cli/              # CLI 接口层
│   │   ├── cli.go        # 根命令和全局配置
│   │   ├── init.go       # init 命令实现
│   │   ├── download.go   # download 命令实现
│   │   └── version.go    # version 和 config 命令
│   ├── ui/               # UI 组件层
│   │   ├── ui.go         # UI 管理器和渲染组件
│   │   └── tracker.go    # 步骤跟踪器
│   ├── business/         # 业务逻辑层
│   │   ├── init.go       # 初始化业务逻辑
│   │   └── download.go   # 下载业务逻辑
│   ├── infrastructure/   # 基础设施层
│   │   ├── git.go        # Git 操作
│   │   ├── tools.go      # 工具检查
│   │   ├── template.go   # 模板提供者
│   │   ├── auth.go       # 认证提供者
│   │   └── system.go     # 系统操作
│   ├── types/            # 类型定义
│   │   └── types.go      # 核心数据结构和接口
│   └── config/           # 配置管理
│       └── config.go     # 配置信息和默认值
├── go.mod                # Go 模块定义
└── README.md             # 项目说明文档
```

## 核心特性

### 1. CLI 接口层
- **根命令**: `specify` - 主命令入口
- **子命令**:
  - `init` - 初始化项目
  - `download` - 下载模板
  - `version` - 显示版本信息
  - `config` - 显示配置信息
- **全局标志**: `--verbose`, `--debug`

### 2. UI 组件层
- **交互式选择**: 支持箭头键选择
- **进度跟踪**: 实时显示执行步骤
- **彩色输出**: 成功、错误、警告等状态显示
- **表格显示**: 格式化信息展示

### 3. 业务逻辑层
- **初始化处理器**: 完整的项目初始化流程
- **下载处理器**: 模板下载和提取流程
- **步骤编排**: 有序执行各个操作步骤

### 4. 基础设施层
- **Git 操作**: 仓库初始化、提交、分支管理
- **工具检查**: 验证系统依赖工具
- **模板提供者**: GitHub 模板下载和管理
- **认证提供者**: GitHub Token 认证
- **系统操作**: 文件系统和命令执行

## 支持的 AI 助手

| 助手名称 | 文件夹 | CLI 要求 |
|---------|--------|----------|
| GitHub Copilot | copilot | gh |
| Claude Code | claude | claude |
| Gemini CLI | gemini | gemini |
| OpenAI CLI | openai | openai |
| Hugging Face | huggingface | huggingface-hub |
| Ollama | ollama | ollama |
| Mistral AI | mistral | mistral |
| Cohere | cohere | cohere |
| Perplexity AI | perplexity | perplexity |
| Anthropic | anthropic | anthropic |
| DeepSeek | deepseek | deepseek |
| Qwen | qwen | qwen |
| ChatGLM | chatglm | chatglm |

## 使用方法

### 安装依赖

```bash
go mod tidy
```

### 构建项目

```bash
go build -o specify ./cmd/specify
```

### 基本命令

```bash
# 初始化项目
./specify init --name my-project --ai copilot --script sh

# 在当前目录初始化
./specify init --here --ai claude --script ps

# 下载模板
./specify download --ai gemini --dir ./templates

# 查看版本信息
./specify version

# 查看配置信息
./specify config
```

### 环境变量

- `GITHUB_TOKEN`: GitHub 访问令牌
- `GH_TOKEN`: GitHub CLI 令牌（备用）

## 设计模式

### 1. 命令模式 (Command Pattern)
- CLI 命令结构化设计
- 每个命令独立封装

### 2. 策略模式 (Strategy Pattern)
- 不同 AI 助手的处理策略
- 多种脚本类型支持

### 3. 观察者模式 (Observer Pattern)
- 步骤执行进度通知
- UI 状态更新

### 4. 工厂模式 (Factory Pattern)
- 组件实例创建
- 依赖注入管理

## 扩展性设计

### 1. 配置驱动
- 通过配置文件添加新的 AI 助手
- 灵活的脚本类型配置

### 2. 插件架构
- 接口化设计支持插件扩展
- 模块化组件便于替换

### 3. 多平台支持
- 跨平台文件操作
- 系统特定的命令执行

## 开发指南

### 添加新的 AI 助手

1. 在 `config/config.go` 中添加助手信息
2. 更新 `AgentInfoMap` 配置
3. 添加相应的模板文件

### 添加新的命令

1. 在 `cli/` 目录创建命令文件
2. 实现命令逻辑和参数解析
3. 在 `cli.go` 中注册命令

### 扩展业务逻辑

1. 在 `business/` 目录添加处理器
2. 实现相应的业务流程
3. 集成到 CLI 命令中

## 代码质量特性

- **类型安全**: 强类型定义和接口约束
- **错误处理**: 显式错误处理和传播
- **并发安全**: 使用 `sync.RWMutex` 保护共享资源
- **资源管理**: 使用 `defer` 确保资源释放
- **用户体验**: 丰富的交互和反馈机制

## 许可证

MIT License

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

---

**注意**: 这是基于 `golanganalyse.md` 文档分析生成的框架代码，专注于架构设计和代码结构，具体的实现细节需要根据实际需求进一步完善。