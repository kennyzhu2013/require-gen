# Specify CLI 详细设计文档

## 1. 系统概述

Specify CLI是一个基于Python的命令行工具，用于Spec-Driven Development项目的初始化和管理。系统采用分层架构设计，支持多种AI助手集成和跨平台操作。

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Interface Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Typer App     │  │   Commands      │  │   Callbacks  │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    UI Components Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │  StepTracker    │  │  Selector       │  │   Banner     │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Business Logic Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │  Template Mgmt  │  │  Git Operations │  │  Tool Check  │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │  HTTP Client    │  │  File System    │  │  Process     │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 3. 核心模块设计

### 3.1 配置管理模块

```pseudocode
MODULE ConfigurationManager:
    CONSTANTS:
        AGENT_CONFIG = {
            "copilot": {name, folder, install_url, requires_cli},
            "claude": {name, folder, install_url, requires_cli},
            // ... 其他AI助手配置
        }
        SCRIPT_TYPE_CHOICES = {"sh": "POSIX Shell", "ps": "PowerShell"}
        BANNER = ASCII_ART_STRING
        CLAUDE_LOCAL_PATH = Path.home() / ".claude" / "local" / "claude"

    FUNCTION get_agent_config(agent_name):
        RETURN AGENT_CONFIG[agent_name]
    
    FUNCTION get_supported_agents():
        RETURN list(AGENT_CONFIG.keys())
```

### 3.2 认证管理模块

```pseudocode
MODULE AuthenticationManager:
    FUNCTION get_github_token(cli_token):
        INPUT: cli_token (optional string)
        OUTPUT: sanitized_token or None
        
        ALGORITHM:
            token = cli_token OR env("GH_TOKEN") OR env("GITHUB_TOKEN") OR ""
            cleaned_token = token.strip()
            IF cleaned_token is empty:
                RETURN None
            ELSE:
                RETURN cleaned_token
    
    FUNCTION get_auth_headers(cli_token):
        INPUT: cli_token (optional string)
        OUTPUT: headers dictionary
        
        ALGORITHM:
            token = get_github_token(cli_token)
            IF token exists:
                RETURN {"Authorization": "Bearer " + token}
            ELSE:
                RETURN empty_dict
```

### 3.3 系统工具模块

```pseudocode
MODULE SystemToolsManager:
    FUNCTION run_command(cmd, check_return, capture, shell):
        INPUT: 
            cmd (list of strings)
            check_return (boolean)
            capture (boolean) 
            shell (boolean)
        OUTPUT: command_output or None
        
        ALGORITHM:
            TRY:
                IF capture:
                    result = subprocess.run(cmd, check=check_return, capture_output=True, text=True, shell=shell)
                    RETURN result.stdout.strip()
                ELSE:
                    subprocess.run(cmd, check=check_return, shell=shell)
                    RETURN None
            CATCH CalledProcessError as e:
                IF check_return:
                    display_error(e)
                    RAISE exception
                RETURN None
    
    FUNCTION check_tool(tool_name, tracker):
        INPUT: 
            tool_name (string)
            tracker (StepTracker, optional)
        OUTPUT: boolean
        
        ALGORITHM:
            // 特殊处理Claude CLI
            IF tool_name == "claude":
                IF CLAUDE_LOCAL_PATH exists and is_file:
                    IF tracker:
                        tracker.complete(tool_name, "available")
                    RETURN True
            
            found = shutil.which(tool_name) is not None
            
            IF tracker:
                IF found:
                    tracker.complete(tool_name, "available")
                ELSE:
                    tracker.error(tool_name, "not found")
            
            RETURN found
```

### 3.4 Git操作模块

```pseudocode
MODULE GitOperationsManager:
    FUNCTION is_git_repo(path):
        INPUT: path (Path, optional)
        OUTPUT: boolean
        
        ALGORITHM:
            IF path is None:
                path = current_working_directory
            
            IF path is not directory:
                RETURN False
            
            TRY:
                subprocess.run(["git", "rev-parse", "--is-inside-work-tree"], 
                              check=True, capture_output=True, cwd=path)
                RETURN True
            CATCH (CalledProcessError, FileNotFoundError):
                RETURN False
    
    FUNCTION init_git_repo(project_path, quiet):
        INPUT:
            project_path (Path)
            quiet (boolean)
        OUTPUT: (success_boolean, error_message)
        
        ALGORITHM:
            TRY:
                original_cwd = current_working_directory
                change_directory(project_path)
                
                IF not quiet:
                    display("Initializing git repository...")
                
                run_command(["git", "init"])
                run_command(["git", "add", "."])
                run_command(["git", "commit", "-m", "Initial commit from Specify template"])
                
                IF not quiet:
                    display("✓ Git repository initialized")
                
                RETURN (True, None)
            
            CATCH CalledProcessError as e:
                error_msg = format_error_message(e)
                IF not quiet:
                    display_error(error_msg)
                RETURN (False, error_msg)
            
            FINALLY:
                change_directory(original_cwd)
```

### 3.5 UI组件模块

```pseudocode
MODULE UIComponentsManager:
    CLASS StepTracker:
        ATTRIBUTES:
            title (string)
            steps (list of step_objects)
            status_order (dictionary)
            refresh_callback (function, optional)
        
        METHOD __init__(title):
            self.title = title
            self.steps = empty_list
            self.status_order = {"pending": 0, "running": 1, "done": 2, "error": 3, "skipped": 4}
            self.refresh_callback = None
        
        METHOD attach_refresh(callback):
            self.refresh_callback = callback
        
        METHOD add(key, label):
            IF key not in existing_step_keys:
                step = {key: key, label: label, status: "pending", detail: ""}
                self.steps.append(step)
                self.maybe_refresh()
        
        METHOD start(key, detail):
            self.update_step(key, status="running", detail=detail)
        
        METHOD complete(key, detail):
            self.update_step(key, status="done", detail=detail)
        
        METHOD error(key, detail):
            self.update_step(key, status="error", detail=detail)
        
        METHOD skip(key, detail):
            self.update_step(key, status="skipped", detail=detail)
        
        METHOD update_step(key, status, detail):
            FOR each step in self.steps:
                IF step.key == key:
                    step.status = status
                    IF detail:
                        step.detail = detail
                    self.maybe_refresh()
                    RETURN
            
            // 如果步骤不存在，创建新步骤
            new_step = {key: key, label: key, status: status, detail: detail}
            self.steps.append(new_step)
            self.maybe_refresh()
        
        METHOD maybe_refresh():
            IF self.refresh_callback exists:
                TRY:
                    self.refresh_callback()
                CATCH any_exception:
                    // 忽略刷新错误
                    pass
        
        METHOD render():
            tree = create_rich_tree(self.title)
            FOR each step in self.steps:
                symbol = get_status_symbol(step.status)
                line = format_step_line(symbol, step.label, step.detail, step.status)
                tree.add(line)
            RETURN tree
    
    FUNCTION get_key():
        OUTPUT: standardized_key_name
        
        ALGORITHM:
            key = readchar.readkey()
            
            IF key == UP_ARROW or key == CTRL_P:
                RETURN 'up'
            IF key == DOWN_ARROW or key == CTRL_N:
                RETURN 'down'
            IF key == ENTER:
                RETURN 'enter'
            IF key == ESCAPE:
                RETURN 'escape'
            IF key == CTRL_C:
                RAISE KeyboardInterrupt
            
            RETURN key
    
    FUNCTION select_with_arrows(options, prompt_text, default_key):
        INPUT:
            options (dictionary)
            prompt_text (string)
            default_key (string, optional)
        OUTPUT: selected_option_key
        
        ALGORITHM:
            option_keys = list(options.keys())
            
            IF default_key exists and default_key in option_keys:
                selected_index = option_keys.index(default_key)
            ELSE:
                selected_index = 0
            
            selected_key = None
            
            FUNCTION create_selection_panel():
                table = create_rich_table()
                FOR i, key in enumerate(option_keys):
                    IF i == selected_index:
                        table.add_row("▶", highlight(key, options[key]))
                    ELSE:
                        table.add_row(" ", normal(key, options[key]))
                
                table.add_row("", "")
                table.add_row("", "Use ↑/↓ to navigate, Enter to select, Esc to cancel")
                
                RETURN create_panel(table, title=prompt_text)
            
            WITH live_display(create_selection_panel()) as live:
                WHILE True:
                    TRY:
                        key = get_key()
                        IF key == 'up':
                            selected_index = (selected_index - 1) % len(option_keys)
                        ELIF key == 'down':
                            selected_index = (selected_index + 1) % len(option_keys)
                        ELIF key == 'enter':
                            selected_key = option_keys[selected_index]
                            BREAK
                        ELIF key == 'escape':
                            display("Selection cancelled")
                            EXIT_PROGRAM(1)
                        
                        live.update(create_selection_panel())
                    
                    CATCH KeyboardInterrupt:
                        display("Selection cancelled")
                        EXIT_PROGRAM(1)
            
            IF selected_key is None:
                display_error("Selection failed")
                EXIT_PROGRAM(1)
            
            RETURN selected_key
    
    FUNCTION show_banner():
        ALGORITHM:
            banner_lines = BANNER.split('\n')
            colors = ["bright_blue", "blue", "cyan", "bright_cyan", "white", "bright_white"]
            
            styled_banner = create_text_object()
            FOR i, line in enumerate(banner_lines):
                color = colors[i % len(colors)]
                styled_banner.append(line + "\n", style=color)
            
            display_centered(styled_banner)
            display_centered(TAGLINE, style="italic bright_yellow")
            display_newline()
```

### 3.6 模板管理模块

```pseudocode
MODULE TemplateManager:
    FUNCTION download_template_from_github(ai_assistant, download_dir, script_type, verbose, show_progress, client, debug, github_token):
        INPUT:
            ai_assistant (string)
            download_dir (Path)
            script_type (string, default "sh")
            verbose (boolean, default True)
            show_progress (boolean, default True)
            client (httpx.Client, optional)
            debug (boolean, default False)
            github_token (string, optional)
        OUTPUT: (zip_path, release_data)
        
        ALGORITHM:
            repo_owner = "github"
            repo_name = "spec-kit"
            
            IF client is None:
                client = create_httpx_client()
            
            IF verbose:
                display("Fetching latest release information...")
            
            api_url = f"https://api.github.com/repos/{repo_owner}/{repo_name}/releases/latest"
            
            TRY:
                response = client.get(api_url, timeout=30, follow_redirects=True, 
                                    headers=get_auth_headers(github_token))
                
                IF response.status_code != 200:
                    error_msg = f"GitHub API returned {response.status_code} for {api_url}"
                    IF debug:
                        error_msg += format_debug_info(response)
                    RAISE RuntimeError(error_msg)
                
                TRY:
                    release_data = response.json()
                CATCH ValueError as json_error:
                    RAISE RuntimeError(f"Failed to parse release JSON: {json_error}")
            
            CATCH Exception as e:
                display_error("Error fetching release information")
                display_error_panel(str(e))
                EXIT_PROGRAM(1)
            
            assets = release_data.get("assets", [])
            pattern = f"spec-kit-template-{ai_assistant}-{script_type}"
            
            matching_assets = filter_assets_by_pattern(assets, pattern)
            
            IF no matching_assets:
                display_error(f"No matching release asset found for {ai_assistant}")
                display_available_assets(assets)
                EXIT_PROGRAM(1)
            
            asset = matching_assets[0]
            download_url = asset["browser_download_url"]
            filename = asset["name"]
            file_size = asset["size"]
            
            IF verbose:
                display_asset_info(filename, file_size, release_data["tag_name"])
            
            zip_path = download_dir / filename
            
            IF verbose:
                display("Downloading template...")
            
            TRY:
                WITH client.stream("GET", download_url, timeout=60, follow_redirects=True,
                                  headers=get_auth_headers(github_token)) as response:
                    
                    IF response.status_code != 200:
                        RAISE RuntimeError(f"Download failed with {response.status_code}")
                    
                    total_size = int(response.headers.get('content-length', 0))
                    
                    WITH open(zip_path, 'wb') as file:
                        IF total_size == 0:
                            FOR chunk in response.iter_bytes(chunk_size=8192):
                                file.write(chunk)
                        ELSE:
                            IF show_progress:
                                WITH progress_bar(total=total_size) as progress:
                                    task = progress.add_task("Downloading...", total=total_size)
                                    downloaded = 0
                                    FOR chunk in response.iter_bytes(chunk_size=8192):
                                        file.write(chunk)
                                        downloaded += len(chunk)
                                        progress.update(task, completed=downloaded)
                            ELSE:
                                FOR chunk in response.iter_bytes(chunk_size=8192):
                                    file.write(chunk)
            
            CATCH Exception as e:
                display_error(f"Download failed: {e}")
                EXIT_PROGRAM(1)
            
            RETURN (zip_path, release_data)
```

### 3.7 主命令模块

```pseudocode
MODULE MainCommandsManager:
    CLASS BannerGroup(TyperGroup):
        METHOD format_help(ctx, formatter):
            show_banner()
            CALL parent.format_help(ctx, formatter)
    
    FUNCTION create_app():
        app = create_typer_app(
            name="specify",
            help="Setup tool for Specify spec-driven development projects",
            add_completion=False,
            invoke_without_command=True,
            cls=BannerGroup
        )
        RETURN app
    
    FUNCTION callback(ctx):
        INPUT: ctx (typer.Context)
        
        ALGORITHM:
            IF ctx.invoked_subcommand is None AND "--help" not in sys.argv AND "-h" not in sys.argv:
                show_banner()
                display_centered("Run 'specify --help' for usage information")
                display_newline()
    
    FUNCTION init_command(project_name, here, ai_assistant, script_type, github_token, verbose, debug):
        INPUT:
            project_name (string, optional)
            here (boolean)
            ai_assistant (string, optional)
            script_type (string, optional)
            github_token (string, optional)
            verbose (boolean)
            debug (boolean)
        
        ALGORITHM:
            // 步骤1: 确定项目路径和名称
            IF here:
                project_path = current_working_directory
                project_name = project_path.name
            ELIF project_name == ".":
                project_path = current_working_directory
                project_name = project_path.name
            ELIF project_name:
                project_path = current_working_directory / project_name
            ELSE:
                project_name = prompt_for_project_name()
                project_path = current_working_directory / project_name
            
            // 步骤2: 创建步骤跟踪器
            tracker = create_step_tracker(f"Initializing {project_name}")
            
            WITH live_display(tracker.render()) as live:
                tracker.attach_refresh(live.refresh)
                
                // 步骤3: 选择AI助手
                IF ai_assistant is None:
                    tracker.add("select_ai", "Select AI assistant")
                    tracker.start("select_ai")
                    
                    ai_options = get_ai_assistant_options()
                    ai_assistant = select_with_arrows(ai_options, "Choose your AI assistant")
                    
                    tracker.complete("select_ai", ai_assistant)
                
                // 步骤4: 选择脚本类型
                IF script_type is None:
                    tracker.add("select_script", "Select script type")
                    tracker.start("select_script")
                    
                    script_type = select_with_arrows(SCRIPT_TYPE_CHOICES, "Choose script type")
                    
                    tracker.complete("select_script", script_type)
                
                // 步骤5: 检查必需工具
                tracker.add("check_tools", "Check required tools")
                tracker.start("check_tools")
                
                required_tools = get_required_tools(ai_assistant)
                all_tools_available = True
                
                FOR tool in required_tools:
                    IF not check_tool(tool, tracker):
                        all_tools_available = False
                
                IF all_tools_available:
                    tracker.complete("check_tools", "All tools available")
                ELSE:
                    tracker.error("check_tools", "Some tools missing")
                    display_installation_instructions(ai_assistant)
                    EXIT_PROGRAM(1)
                
                // 步骤6: 创建项目目录
                tracker.add("create_dir", "Create project directory")
                tracker.start("create_dir")
                
                TRY:
                    create_directory(project_path, exist_ok=True)
                    tracker.complete("create_dir", str(project_path))
                CATCH Exception as e:
                    tracker.error("create_dir", str(e))
                    EXIT_PROGRAM(1)
                
                // 步骤7: 下载模板
                tracker.add("download", "Download template")
                tracker.start("download")
                
                TRY:
                    WITH temporary_directory() as temp_dir:
                        zip_path, release_data = download_template_from_github(
                            ai_assistant, temp_dir, script_type=script_type,
                            verbose=False, show_progress=False,
                            github_token=github_token, debug=debug
                        )
                        
                        tracker.complete("download", release_data["tag_name"])
                        
                        // 步骤8: 解压模板
                        tracker.add("extract", "Extract template")
                        tracker.start("extract")
                        
                        extract_template(zip_path, project_path)
                        tracker.complete("extract", "Template files extracted")
                
                CATCH Exception as e:
                    tracker.error("download", str(e))
                    EXIT_PROGRAM(1)
                
                // 步骤9: 初始化Git仓库
                tracker.add("git_init", "Initialize Git repository")
                tracker.start("git_init")
                
                IF not is_git_repo(project_path):
                    success, error = init_git_repo(project_path, quiet=True)
                    IF success:
                        tracker.complete("git_init", "Repository initialized")
                    ELSE:
                        tracker.error("git_init", error)
                ELSE:
                    tracker.skip("git_init", "Already a Git repository")
                
                // 步骤10: 完成设置
                tracker.add("finalize", "Finalize setup")
                tracker.start("finalize")
                
                setup_project_configuration(project_path, ai_assistant, script_type)
                tracker.complete("finalize", "Project ready")
            
            // 显示成功信息
            display_success_message(project_name, project_path, ai_assistant)
```

## 4. 数据流设计

```
用户输入 → 参数解析 → 命令路由 → 业务逻辑 → 系统调用 → 结果反馈

具体流程:
1. CLI参数解析 (Typer)
2. 配置验证和默认值设置
3. AI助手选择 (交互式或参数指定)
4. 脚本类型选择 (交互式或参数指定)
5. 工具依赖检查 (并行检查多个工具)
6. 项目目录创建
7. GitHub模板下载 (带进度显示)
8. 模板解压和文件复制
9. Git仓库初始化 (可选)
10. 项目配置文件生成
11. 成功反馈和后续步骤提示
```

## 5. 错误处理策略

```pseudocode
ERROR_HANDLING_STRATEGY:
    1. 网络错误:
        - HTTP超时: 重试机制 + 用户友好提示
        - API限制: 显示认证建议
        - 下载失败: 清理临时文件 + 重试选项
    
    2. 文件系统错误:
        - 权限不足: 提示管理员权限需求
        - 磁盘空间不足: 显示空间需求
        - 路径不存在: 自动创建父目录
    
    3. 工具依赖错误:
        - 工具未安装: 显示安装链接和说明
        - 版本不兼容: 显示版本要求
        - 配置错误: 提供配置修复建议
    
    4. 用户输入错误:
        - 无效选择: 重新提示选择
        - 中断操作: 优雅退出 + 清理
        - 参数冲突: 显示正确用法
```

## 6. 性能优化设计

```pseudocode
PERFORMANCE_OPTIMIZATIONS:
    1. 并发处理:
        - 工具检查: 并行执行多个工具检查
        - 文件操作: 异步文件I/O
        - 网络请求: 连接池复用
    
    2. 缓存策略:
        - GitHub API响应缓存
        - 工具检查结果缓存
        - 模板文件本地缓存
    
    3. 资源管理:
        - 临时文件自动清理
        - 内存使用监控
        - 网络连接超时控制
    
    4. 用户体验:
        - 实时进度反馈
        - 增量状态更新
        - 响应式UI刷新
```

## 7. 扩展性设计

```pseudocode
EXTENSIBILITY_DESIGN:
    1. 插件架构:
        - AI助手插件接口
        - 自定义命令扩展
        - 模板提供者扩展
    
    2. 配置驱动:
        - 外部配置文件支持
        - 环境变量覆盖
        - 运行时配置更新
    
    3. 国际化支持:
        - 多语言消息系统
        - 本地化UI组件
        - 区域特定配置
    
    4. 平台适配:
        - 操作系统特定逻辑
        - 包管理器集成
        - 系统服务集成
```

这个详细设计文档提供了完整的伪代码实现，涵盖了系统的各个层面和模块，可以作为实际开发的参考指南。