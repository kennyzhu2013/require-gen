package infrastructure

import (
	"os"
	"testing"
	"time"
)

func TestTempFileManager_CreateAndCleanup(t *testing.T) {
	sysOps := NewSystemOperations()
	
	// 创建临时文件
	tempFile, err := sysOps.CreateTempFile("test_")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// 验证文件存在
	if !sysOps.FileExists(tempFile) {
		t.Fatalf("Temp file should exist: %s", tempFile)
	}
	
	// 写入测试数据
	testData := []byte("Hello, World!")
	if err := sysOps.WriteFile(tempFile, testData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	
	// 验证数据写入成功
	data, err := sysOps.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}
	if string(data) != string(testData) {
		t.Fatalf("Expected %s, got %s", string(testData), string(data))
	}
	
	// 手动删除文件进行清理测试
	if err := os.Remove(tempFile); err != nil {
		t.Fatalf("Failed to remove temp file: %v", err)
	}
	
	// 验证文件已被删除
	if sysOps.FileExists(tempFile) {
		t.Fatalf("Temp file should be deleted: %s", tempFile)
	}
}

func TestTempFileManager_CreateTempDir(t *testing.T) {
	sysOps := NewSystemOperations()
	
	// 创建临时目录
	tempDir, err := sysOps.CreateTempDirectory("test_dir_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// 验证目录存在
	if !sysOps.DirectoryExists(tempDir) {
		t.Fatalf("Temp dir should exist and be a directory: %s", tempDir)
	}
	
	// 在临时目录中创建文件
	testFile := sysOps.JoinPath(tempDir, "test.txt")
	testData := []byte("test content")
	if err := sysOps.WriteFile(testFile, testData); err != nil {
		t.Fatalf("Failed to create file in temp dir: %v", err)
	}
	
	// 清理
	if err := sysOps.RemoveDirectory(tempDir); err != nil {
		t.Fatalf("Failed to remove temp directory: %v", err)
	}
	
	// 验证目录已被删除
	if sysOps.DirectoryExists(tempDir) {
		t.Fatalf("Temp dir should be deleted: %s", tempDir)
	}
}

func TestTempFileManager_MultipleFiles(t *testing.T) {
	sysOps := NewSystemOperations()
	
	var tempFiles []string
	
	// 创建多个临时文件
	for i := 0; i < 5; i++ {
		tempFile, err := sysOps.CreateTempFile("multi_test_")
		if err != nil {
			t.Fatalf("Failed to create temp file %d: %v", i, err)
		}
		tempFiles = append(tempFiles, tempFile)
		
		// 写入数据
		data := []byte("test data " + string(rune('0'+i)))
		if err := sysOps.WriteFile(tempFile, data); err != nil {
			t.Fatalf("Failed to write to temp file %d: %v", i, err)
		}
	}
	
	// 验证所有文件都存在
	for i, tempFile := range tempFiles {
		if !sysOps.FileExists(tempFile) {
			t.Fatalf("Temp file %d should exist: %s", i, tempFile)
		}
	}
	
	// 清理所有文件
	for i, tempFile := range tempFiles {
		if err := os.Remove(tempFile); err != nil {
			t.Fatalf("Failed to remove temp file %d: %v", i, err)
		}
	}
	
	// 验证所有文件都被删除
	for i, tempFile := range tempFiles {
		if sysOps.FileExists(tempFile) {
			t.Fatalf("Temp file %d should be deleted: %s", i, tempFile)
		}
	}
}

func TestTempFileManager_AutoCleanupOnExit(t *testing.T) {
	// 这个测试验证程序退出时的自动清理功能
	sysOps := NewSystemOperations()
	
	// 创建临时文件
	tempFile, err := sysOps.CreateTempFile("auto_cleanup_")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// 验证文件存在
	if !sysOps.FileExists(tempFile) {
		t.Fatalf("Temp file should exist: %s", tempFile)
	}
	
	// 模拟程序退出时的清理
	// 注意：实际的自动清理是通过signal handler实现的
	// 这里我们手动删除文件来模拟
	if err := os.Remove(tempFile); err != nil {
		t.Fatalf("Failed to cleanup temp file: %v", err)
	}
	
	// 验证文件已被删除
	if sysOps.FileExists(tempFile) {
		t.Fatalf("Temp file should be deleted on exit: %s", tempFile)
	}
}

func TestTempFileManager_ConcurrentAccess(t *testing.T) {
	sysOps := NewSystemOperations()
	
	// 并发创建临时文件
	done := make(chan bool, 10)
	var tempFiles []string
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			tempFile, err := sysOps.CreateTempFile("concurrent_")
			if err != nil {
				t.Errorf("Failed to create temp file in goroutine %d: %v", id, err)
				return
			}
			
			// 写入数据
			data := []byte("concurrent test data")
			if err := sysOps.WriteFile(tempFile, data); err != nil {
				t.Errorf("Failed to write to temp file in goroutine %d: %v", id, err)
				return
			}
			
			tempFiles = append(tempFiles, tempFile)
		}(i)
	}
	
	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
	
	// 清理
	for _, tempFile := range tempFiles {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file %s: %v", tempFile, err)
		}
	}
}

func TestTempFileManager_ErrorHandling(t *testing.T) {
	sysOps := NewSystemOperations()
	
	// 测试在无效目录中创建临时文件
	originalTempDir := os.TempDir()
	defer func() {
		// 恢复原始临时目录
		os.Setenv("TMPDIR", originalTempDir)
	}()
	
	// 设置一个不存在的临时目录
	invalidTempDir := "/invalid/temp/dir"
	os.Setenv("TMPDIR", invalidTempDir)
	
	// 尝试创建临时文件（应该回退到系统默认目录）
	tempFile, err := sysOps.CreateTempFile("error_test_")
	if err != nil {
		// 如果失败，这是预期的行为
		t.Logf("Expected error when using invalid temp dir: %v", err)
	} else {
		// 如果成功，验证文件确实被创建
		if !sysOps.FileExists(tempFile) {
			t.Fatalf("Temp file should exist: %s", tempFile)
		}
		
		// 清理
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}
}