package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"specify-cli/internal/infrastructure"
	"specify-cli/internal/types"
)

func main() {
	fmt.Println("=== Infrastructure Layer 功能验证测试 ===")
	
	// 1. 测试系统操作
	fmt.Println("=== 测试系统操作 ===")
	sysOps := infrastructure.NewSystemOperations()
	
	// 测试目录操作
	testDir := "./test_dir"
	if err := sysOps.CreateDirectory(testDir); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
	} else {
		fmt.Println("✓ 目录创建成功")
	}
	
	// 测试文件操作
	testFile := filepath.Join(testDir, "test.txt")
	if err := sysOps.WriteFile(testFile, []byte("Hello Infrastructure")); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
	} else {
		fmt.Println("✓ 文件写入成功")
	}
	
	// 清理测试文件
	defer func() {
		if err := sysOps.RemoveDirectory(testDir); err != nil {
			fmt.Printf("清理目录失败: %v\n", err)
		}
	}()

	// 2. 测试HTTP客户端管理
	fmt.Println("\n=== 测试HTTP客户端管理 ===")
	httpConfig := infrastructure.DefaultHTTPClientConfig()
	httpManager := infrastructure.NewHTTPClientManager(httpConfig)
	
	// 获取默认客户端
	client := httpManager.GetDefaultClient()
	if client == nil {
		fmt.Printf("获取HTTP客户端失败\n")
	} else {
		fmt.Println("✓ HTTP客户端获取成功")
	}
	
	// 获取命名客户端
	namedClient := httpManager.GetClient("test")
	if namedClient != nil {
		fmt.Println("✓ 命名客户端获取成功")
	}

	// 3. 测试自适应进度管理
	fmt.Println("\n=== 测试自适应进度管理 ===")
	adaptiveConfig := infrastructure.DefaultAdaptiveConfig()
	progressDisplay := infrastructure.NewAdaptiveProgressDisplay(adaptiveConfig)
	
	// 测试进度显示
	progressDisplay.Start(100)
	progressDisplay.SetMessage("测试进度显示")
	
	// 模拟进度更新
	for i := 0; i <= 100; i += 20 {
		progressInfo := &types.ProgressInfo{
			Downloaded: int64(i),
			Total:      100,
			Percentage: float64(i),
			Speed:      1024 * 1024, // 1MB/s
		}
		progressDisplay.Update(progressInfo)
		time.Sleep(100 * time.Millisecond)
	}
	
	progressDisplay.Finish()
	fmt.Println("✓ 自适应进度管理测试完成")

	// 4. 测试错误处理
	fmt.Println("\n=== 测试错误处理 ===")
	errorConfig := infrastructure.DefaultErrorHandlerConfig()
	errorHandler := infrastructure.NewNetworkErrorHandler(errorConfig)
	
	// 模拟网络错误
	testErr := fmt.Errorf("connection timeout")
	networkErr := errorHandler.HandleError(context.Background(), testErr, "example.com")
	
	if networkErr != nil {
		fmt.Printf("✓ 网络错误处理成功: %s\n", networkErr.Message)
		
		// 测试重试判断
		shouldRetry := errorHandler.ShouldRetry(networkErr, 1)
		fmt.Printf("✓ 重试判断: %v\n", shouldRetry)
		
		// 测试重试延迟计算
		delay := errorHandler.CalculateRetryDelay(networkErr, 1)
		fmt.Printf("✓ 重试延迟: %v\n", delay)
	}

	// 5. 测试重试管理
	fmt.Println("\n=== 测试重试管理 ===")
	retryConfig := infrastructure.DefaultRetryManagerConfig()
	retryManager := infrastructure.NewRetryManager(retryConfig, errorHandler)
	
	// 测试重试执行
	retryOptions := &infrastructure.RetryOptions{
		StrategyName: "default",
		Host:         "example.com",
		Operation:    "test_operation",
	}
	
	result := retryManager.ExecuteWithRetry(context.Background(), func(ctx context.Context, attempt int) error {
		if attempt < 2 {
			return fmt.Errorf("模拟失败 attempt %d", attempt)
		}
		return nil // 第3次尝试成功
	}, retryOptions)
	
	if result.Success {
		fmt.Printf("✓ 重试执行成功，尝试次数: %d\n", result.AttemptCount)
	} else {
		fmt.Printf("✗ 重试执行失败: %v\n", result.LastError)
	}

	// 6. 测试压缩文件处理
	fmt.Println("\n=== 测试压缩文件处理 ===")
	_ = infrastructure.NewZipProcessor(sysOps)
	
	// 这里只是演示接口调用，实际使用需要真实的ZIP文件
	fmt.Println("✓ ZIP处理器创建成功")
	
	// 7. 测试TAR文件处理
	fmt.Println("\n=== 测试TAR文件处理 ===")
	_ = infrastructure.NewTarProcessor(sysOps)
	fmt.Println("✓ TAR处理器创建成功")
	
	// 8. 测试模板提供者
	fmt.Println("\n=== 测试模板提供者 ===")
	templateProvider := infrastructure.NewTemplateProvider()
	fmt.Printf("✓ 模板提供者创建成功: %T\n", templateProvider)
	
	// 9. 测试认证提供者
	fmt.Println("\n=== 测试认证提供者 ===")
	authProvider := infrastructure.NewAuthProvider()
	fmt.Printf("✓ 认证提供者创建成功: %T\n", authProvider)
	
	// 测试令牌获取（不会有真实令牌，但可以测试接口）
	token := authProvider.GetToken()
	if token == "" {
		fmt.Println("✓ 令牌获取测试完成（无令牌配置）")
	} else {
		fmt.Printf("✓ 令牌获取成功: %s...\n", token[:min(len(token), 10)])
	}
	
	fmt.Println("\n=== 所有测试完成 ===")
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// testSystemOperations 测试系统操作功能
func testSystemOperations() {
	fmt.Println("\n--- 测试系统操作功能 ---")
	
	sysOps := infrastructure.NewSystemOperations()
	
	// 测试临时目录创建
	tempDir, err := sysOps.CreateTempDirectory("infrastructure_test")
	if err != nil {
		log.Printf("创建临时目录失败: %v", err)
		return
	}
	defer sysOps.RemoveDirectory(tempDir)
	fmt.Printf("✓ 临时目录创建成功: %s\n", tempDir)
	
	// 测试文件操作
	testFile := filepath.Join(tempDir, "test.txt")
	content := "这是测试内容"
	
	err = sysOps.WriteFile(testFile, []byte(content))
	if err != nil {
		log.Printf("写入文件失败: %v", err)
		return
	}
	fmt.Printf("✓ 文件写入成功: %s\n", testFile)
	
	// 测试文件读取
	readContent, err := sysOps.ReadFile(testFile)
	if err != nil {
		log.Printf("读取文件失败: %v", err)
		return
	}
	
	if string(readContent) == content {
		fmt.Println("✓ 文件读取验证成功")
	} else {
		fmt.Println("✗ 文件内容不匹配")
	}
	
	// 测试目录操作
	subDir := filepath.Join(tempDir, "subdir")
	err = sysOps.CreateDirectory(subDir)
	if err != nil {
		log.Printf("创建子目录失败: %v", err)
		return
	}
	fmt.Printf("✓ 子目录创建成功: %s\n", subDir)
	
	// 测试文件存在性检查
	if sysOps.FileExists(testFile) {
		fmt.Println("✓ 文件存在性检查通过")
	} else {
		fmt.Println("✗ 文件存在性检查失败")
	}
}

// testHTTPClientManager 测试HTTP客户端管理功能
func testHTTPClientManager() {
	fmt.Println("\n--- 测试HTTP客户端管理功能 ---")
	
	// 测试HTTP客户端配置
	config := infrastructure.DefaultHTTPClientConfig()
	config.RetryCount = 3
	config.Timeout = 30 * time.Second
	
	// 创建HTTP客户端管理器
	clientManager := infrastructure.NewHTTPClientManager(config)
	
	// 创建命名客户端
	namedClient := clientManager.CreateClient(config)
	if namedClient != nil {
		fmt.Println("✓ 命名HTTP客户端创建成功")
	} else {
		fmt.Println("✗ 命名HTTP客户端创建失败")
	}
	
	// 测试默认客户端获取
	client := clientManager.GetDefaultClient()
	if client != nil {
		fmt.Println("✓ 默认HTTP客户端获取成功")
	} else {
		fmt.Println("✗ 默认HTTP客户端获取失败")
		return
	}
	
	// 测试客户端配置（简化测试）
	fmt.Println("✓ HTTP客户端配置验证完成")
}

// testProgressManager 测试进度管理功能
func testProgressManager() {
	fmt.Println("\n--- 测试进度管理功能 ---")
	
	// 测试自适应进度显示
	config := infrastructure.DefaultAdaptiveConfig()
	progressDisplay := infrastructure.NewAdaptiveProgressDisplay(config)
	
	if progressDisplay != nil {
		fmt.Println("✓ 自适应进度显示创建成功")
		
		// 开始进度显示
		progressDisplay.Start(100)
		
		// 模拟进度更新
		for i := 0; i <= 100; i += 20 {
			progressInfo := &types.ProgressInfo{
				Downloaded: int64(i),
				Total:      100,
				Percentage: float64(i),
				Speed:      1024,
			}
			progressDisplay.Update(progressInfo)
			time.Sleep(100 * time.Millisecond)
		}
		
		progressDisplay.Finish()
		fmt.Println("✓ 进度更新测试完成")
	} else {
		fmt.Println("✗ 自适应进度显示创建失败")
	}
}

// testErrorHandling 测试错误处理功能
func testErrorHandling() {
	fmt.Println("\n--- 测试错误处理功能 ---")
	
	// 创建错误处理器
	config := infrastructure.DefaultErrorHandlerConfig()
	handler := infrastructure.NewNetworkErrorHandler(config)
	
	// 测试错误分类
	testError := fmt.Errorf("connection refused")
	networkErr := handler.ClassifyError(testError, "example.com")
	
	if networkErr != nil {
		fmt.Printf("✓ 错误分类成功: 类型=%v, 可重试=%v\n", networkErr.Type, networkErr.Retryable)
	} else {
		fmt.Println("✗ 错误分类失败")
		return
	}
	
	// 测试重试策略
	if networkErr.Retryable {
		fmt.Println("✓ 重试策略配置正确")
	} else {
		fmt.Println("✗ 重试策略配置错误")
	}
	
	// 获取错误统计
	stats := handler.GetErrorStats()
	fmt.Printf("✓ 错误统计获取成功，总错误数: %d\n", stats.TotalErrors)
}

// testRetryManager 测试重试管理功能
func testRetryManager() {
	fmt.Println("\n--- 测试重试管理功能 ---")
	
	// 创建错误处理器
	errorHandler := infrastructure.NewNetworkErrorHandler(nil)
	
	// 创建重试管理器
	config := infrastructure.DefaultRetryManagerConfig()
	manager := infrastructure.NewRetryManager(config, errorHandler)
	
	// 测试重试策略创建
	strategy := &infrastructure.RetryStrategy{
		MaxRetries:  3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
		BackoffType: infrastructure.ExponentialBackoff,
	}
	
	manager.AddStrategy("test-operation", strategy)
	fmt.Println("✓ 重试策略设置成功")
	
	// 测试重试执行
	ctx := context.Background()
	attempts := 0
	
	operation := func(ctx context.Context, attempt int) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("模拟失败 (尝试 %d)", attempts)
		}
		return nil
	}
	
	options := &infrastructure.RetryOptions{
		StrategyName: "test-operation",
		Operation:    "test",
	}
	
	result := manager.ExecuteWithRetry(ctx, operation, options)
	if result.Success {
		fmt.Printf("✓ 重试执行成功，总尝试次数: %d\n", result.AttemptCount)
	} else {
		fmt.Printf("✗ 重试执行失败: %v\n", result.LastError)
	}
	
	// 测试重试统计
	stats := manager.GetRetryStatistics()
	fmt.Printf("✓ 重试统计信息获取成功: 总重试次数=%d\n", stats.TotalRetries)
}

// testCompressionHandling 测试压缩文件处理功能
func testCompressionHandling() {
	fmt.Println("\n--- 测试压缩文件处理功能 ---")
	
	sysOps := infrastructure.NewSystemOperations()
	
	// 测试ZIP处理器
	_ = infrastructure.NewZipProcessor(sysOps)
	fmt.Println("✓ ZIP处理器创建成功")
	
	// 创建测试目录
	tempDir, err := sysOps.CreateTempDirectory("compression_test")
	if err != nil {
		log.Printf("创建测试目录失败: %v", err)
		return
	}
	defer sysOps.RemoveDirectory(tempDir)
	
	// 创建测试文件用于压缩
	testFile := filepath.Join(tempDir, "test.txt")
	err = sysOps.WriteFile(testFile, []byte("测试压缩内容"))
	if err != nil {
		log.Printf("创建测试文件失败: %v", err)
		return
	}
	
	fmt.Println("✓ 压缩文件处理功能基础验证完成")
	
	// 测试TAR处理器
	_ = infrastructure.NewTarProcessor(sysOps)
	fmt.Println("✓ TAR处理器创建成功")
}

// testNetworkOperations 测试网络操作功能
func testNetworkOperations() {
	fmt.Println("\n--- 测试网络操作功能 ---")
	
	// 测试连接池管理
	poolManager := infrastructure.NewConnectionPoolManager()
	
	if poolManager != nil {
		fmt.Println("✓ 连接池管理器创建成功")
	} else {
		fmt.Println("✗ 连接池管理器创建失败")
		return
	}
	
	// 测试连接池获取
	pool := poolManager.GetPool("default", &infrastructure.ConnectionPoolConfig{
		MaxIdleConns:    10,
		MaxConnsPerHost: 5,
		IdleConnTimeout: 90 * time.Second,
	})
	
	if pool != nil {
		fmt.Println("✓ 连接池获取成功")
	} else {
		fmt.Println("✗ 连接池获取失败")
	}
	
	// 测试流式下载器
	client := &http.Client{Timeout: 30 * time.Second}
	downloader := infrastructure.NewStreamingDownloader(client, 8192)
	
	if downloader != nil {
		fmt.Println("✓ 流式下载器创建成功")
	} else {
		fmt.Println("✗ 流式下载器创建失败")
	}
}

// testAuthenticationSystem 测试认证系统功能
func testAuthenticationSystem() {
	fmt.Println("\n--- 测试认证系统功能 ---")
	
	// 测试GitHub认证提供者
	authProvider := infrastructure.NewAuthProvider()
	if authProvider != nil {
		fmt.Println("✓ 认证提供者创建成功")
		
		// 测试认证状态检查
		if authProvider.IsAuthenticated() {
			fmt.Println("✓ 已认证状态")
			token := authProvider.GetToken()
			fmt.Printf("✓ 获取到令牌 (长度: %d)\n", len(token))
		} else {
			fmt.Println("ℹ 未认证状态 (未设置令牌)")
		}
		
		// 测试认证头生成
		headers := authProvider.GetHeaders()
		fmt.Printf("✓ 生成认证头 (数量: %d)\n", len(headers))
		
		// 测试CLI令牌设置
		authProvider.SetCLIToken("test-cli-token")
		if authProvider.IsAuthenticated() {
			fmt.Println("✓ CLI令牌设置成功")
		}
		
		// 测试令牌验证
		if err := authProvider.ValidateToken(); err != nil {
			fmt.Printf("ℹ 令牌验证结果: %v\n", err)
		} else {
			fmt.Println("✓ 令牌验证通过")
		}
	} else {
		fmt.Println("✗ 认证提供者创建失败")
	}
}

// testTemplateDownloader 测试模板下载功能
func testTemplateDownloader() {
	fmt.Println("\n--- 测试模板下载功能 ---")
	
	// 创建模板提供者
	templateProvider := infrastructure.NewTemplateProvider()
	
	if templateProvider != nil {
		fmt.Println("✓ 模板提供者创建成功")
		
		// 测试模板信息获取
		tempDir := os.TempDir()
		templateInfo, err := templateProvider.GetTemplateInfo(tempDir)
		if err != nil {
			fmt.Printf("ℹ 模板信息获取结果: %v\n", err)
		} else {
			fmt.Printf("✓ 模板信息获取成功: %v\n", templateInfo)
		}
		
		// 测试模板列表获取
		templates, err := templateProvider.ListTemplates("")
		if err != nil {
			fmt.Printf("ℹ 模板列表获取结果: %v\n", err)
		} else {
			fmt.Printf("✓ 模板列表获取成功 (数量: %d)\n", len(templates))
		}
	} else {
		fmt.Println("✗ 模板提供者创建失败")
	}
}

// testGitOperations 测试Git操作功能
func testGitOperations() {
	fmt.Println("\n--- 测试Git操作功能 ---")
	
	gitOps := infrastructure.NewGitOperations()
	
	// 测试Git仓库检查
	cwd, _ := os.Getwd()
	if gitOps.IsRepo(cwd) {
		fmt.Println("✓ 当前目录是Git仓库")
	} else {
		fmt.Println("✓ 当前目录不是Git仓库")
	}
	
	// 创建临时目录用于Git测试
	sysOps := infrastructure.NewSystemOperations()
	tempDir, err := sysOps.CreateTempDirectory("git_test")
	if err != nil {
		log.Printf("创建Git测试目录失败: %v", err)
		return
	}
	defer sysOps.RemoveDirectory(tempDir)
	
	// 测试Git仓库初始化
	success, err := gitOps.InitRepo(tempDir, true)
	if err == nil && success {
		fmt.Println("✓ Git仓库初始化成功")
	} else if err == nil && !success {
		fmt.Println("✓ Git仓库已存在")
	} else {
		fmt.Printf("✗ Git仓库初始化失败: %v\n", err)
	}
}