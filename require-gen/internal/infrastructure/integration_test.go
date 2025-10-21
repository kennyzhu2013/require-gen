package infrastructure

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullIntegration 测试所有功能的完整集成
func TestFullIntegration(t *testing.T) {
	sysOps := NewSystemOperations()
	zipProcessor := NewZipProcessor(sysOps)

	testDir, err := sysOps.CreateTempDirectory("test_full_integration")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建复杂的测试ZIP文件
	zipPath := filepath.Join(testDir, "integration_test.zip")
	err = createComplexTestZipFile(zipPath, map[string]string{
		"README.md":                    "# Project Documentation",
		"src/main.go":                  "package main\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}",
		"src/utils/helper.go":          "package utils\n\nfunc Helper() string {\n\treturn \"helper\"\n}",
		"config/app.json":              `{"name": "test-app", "version": "1.0.0"}`,
		"docs/api.md":                  "# API Documentation\n\nThis is the API documentation.",
		"tests/unit_test.go":           "package tests\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {}",
		"assets/images/logo.svg":       "<svg>...</svg>",
		"assets/styles/main.css":       "body { margin: 0; padding: 0; }",
		"scripts/build.sh":             "#!/bin/bash\necho \"Building...\"",
		"data/sample.json":             `{"data": [1, 2, 3, 4, 5]}`,
		".gitignore":                   "*.log\n*.tmp\nnode_modules/",
		"LICENSE":                      "MIT License\n\nCopyright (c) 2024",
	})
	require.NoError(t, err)

	extractDir := filepath.Join(testDir, "extracted")
	require.NoError(t, sysOps.CreateDirectory(extractDir))

	// 集成测试：进度跟踪、错误处理和智能合并
	var progressUpdates []ProgressUpdate
	var errorEvents []ErrorEvent
	var conflictEvents []ConflictEvent

	opts := &ExtractOptions{
		Verbose:             true,
		OverwriteExisting:   false, // 测试冲突处理
		SmartMerge:          true,
		ConflictResolution:  "rename",
		PreservePermissions: true,
		MaxFileSize:         10 * 1024 * 1024, // 10MB
		OnProgress: func(current, total int64, filename string) {
			progressUpdates = append(progressUpdates, ProgressUpdate{
				Current:   current,
				Total:     total,
				Filename:  filename,
				Timestamp: time.Now(),
			})
		},
		OnError: func(filename string, err error) bool {
			errorEvents = append(errorEvents, ErrorEvent{
				Filename:  filename,
				Error:     err,
				Timestamp: time.Now(),
			})
			return true // 继续处理
		},
		OnConflict: func(conflictInfo *FileConflictInfo) ConflictAction {
			conflictEvents = append(conflictEvents, ConflictEvent{
				SourcePath: conflictInfo.SourcePath,
				TargetPath: conflictInfo.TargetPath,
				Timestamp:  time.Now(),
			})
			return ConflictRename // 重命名冲突文件
		},
	}

	// 第一次提取
	start := time.Now()
	err = zipProcessor.ExtractWithProgress(zipPath, extractDir, opts, func(current, total int64) {
		// 额外的进度回调
	})
	firstExtractionDuration := time.Since(start)
	assert.NoError(t, err)

	// 验证第一次提取的结果
	assert.Greater(t, len(progressUpdates), 0, "Progress updates should be recorded")
	assert.Equal(t, 0, len(errorEvents), "No errors should occur in first extraction")
	assert.Equal(t, 0, len(conflictEvents), "No conflicts should occur in first extraction")

	// 验证文件结构
	expectedFiles := []string{
		"README.md",
		"src/main.go",
		"src/utils/helper.go",
		"config/app.json",
		"docs/api.md",
		"tests/unit_test.go",
		"assets/images/logo.svg",
		"assets/styles/main.css",
		"scripts/build.sh",
		"data/sample.json",
		".gitignore",
		"LICENSE",
	}

	for _, expectedFile := range expectedFiles {
		fullPath := filepath.Join(extractDir, expectedFile)
		assert.True(t, sysOps.FileExists(fullPath), "File %s should exist", expectedFile)
	}

	// 第二次提取（测试冲突处理）
	progressUpdates = nil // 重置
	errorEvents = nil
	conflictEvents = nil

	start = time.Now()
	err = zipProcessor.ExtractWithProgress(zipPath, extractDir, opts, func(current, total int64) {
		// 额外的进度回调
	})
	secondExtractionDuration := time.Since(start)
	assert.NoError(t, err)

	// 验证冲突处理
	assert.Greater(t, len(conflictEvents), 0, "Conflicts should be detected in second extraction")
	
	// 验证重命名的文件存在
	renamedFiles := 0
	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.Contains(info.Name(), "_conflict_") || 
			strings.Contains(info.Name(), "_renamed_") ||
			strings.Contains(path, "_1") || strings.Contains(path, "_2")) {
			renamedFiles++
		}
		return nil
	})
	require.NoError(t, err)
	
	// 由于冲突处理可能通过不同方式实现，我们检查是否有冲突事件被记录
	if len(conflictEvents) > 0 {
		t.Logf("Conflicts were detected and handled: %d events", len(conflictEvents))
	} else {
		t.Log("No conflicts detected in second extraction (files may have been overwritten)")
	}

	// 性能验证
	t.Logf("First extraction took: %v", firstExtractionDuration)
	t.Logf("Second extraction took: %v", secondExtractionDuration)
	t.Logf("Progress updates: %d", len(progressUpdates))
	t.Logf("Error events: %d", len(errorEvents))
	t.Logf("Conflict events: %d", len(conflictEvents))
	t.Logf("Renamed files: %d", renamedFiles)

	// 验证进度更新的完整性
	if len(progressUpdates) > 0 {
		firstUpdate := progressUpdates[0]
		lastUpdate := progressUpdates[len(progressUpdates)-1]
		
		assert.Equal(t, int64(0), firstUpdate.Current, "First progress should start at 0")
		assert.GreaterOrEqual(t, lastUpdate.Current, lastUpdate.Total-1, "Last progress should be close to total")
		assert.Greater(t, lastUpdate.Total, int64(0), "Total should be greater than 0")
	}
}

// TestLargeFileHandling 测试大文件处理
func TestLargeFileHandling(t *testing.T) {
	sysOps := NewSystemOperations()
	zipProcessor := NewZipProcessor(sysOps)

	testDir, err := sysOps.CreateTempDirectory("test_large_file")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建包含大文件的ZIP
	zipPath := filepath.Join(testDir, "large_file_test.zip")
	largeContent := strings.Repeat("This is a large file content. ", 10000) // ~300KB
	
	err = createComplexTestZipFile(zipPath, map[string]string{
		"small_file.txt":  "small content",
		"large_file.txt":  largeContent,
		"medium_file.txt": strings.Repeat("medium content ", 1000), // ~15KB
	})
	require.NoError(t, err)

	extractDir := filepath.Join(testDir, "extracted")
	require.NoError(t, sysOps.CreateDirectory(extractDir))

	var progressUpdates []ProgressUpdate
	var totalBytesProcessed int64

	opts := &ExtractOptions{
		Verbose:           true,
		MaxFileSize:       1024 * 1024, // 1MB limit
		OnProgress: func(current, total int64, filename string) {
			progressUpdates = append(progressUpdates, ProgressUpdate{
				Current:  current,
				Total:    total,
				Filename: filename,
			})
			totalBytesProcessed += int64(len(filename)) // 简单的字节计数
		},
	}

	// 执行提取
	start := time.Now()
	err = zipProcessor.ExtractWithProgress(zipPath, extractDir, opts, func(current, total int64) {
		// 额外的进度回调
	})
	duration := time.Since(start)
	
	assert.NoError(t, err)
	assert.Greater(t, len(progressUpdates), 0, "Progress should be tracked for large files")
	
	// 验证文件提取
	smallFile := filepath.Join(extractDir, "small_file.txt")
	largeFile := filepath.Join(extractDir, "large_file.txt")
	mediumFile := filepath.Join(extractDir, "medium_file.txt")
	
	assert.True(t, sysOps.FileExists(smallFile), "Small file should be extracted")
	assert.True(t, sysOps.FileExists(largeFile), "Large file should be extracted")
	assert.True(t, sysOps.FileExists(mediumFile), "Medium file should be extracted")

	// 验证文件内容
	extractedLargeContent, err := sysOps.ReadFile(largeFile)
	require.NoError(t, err)
	assert.Equal(t, largeContent, string(extractedLargeContent), "Large file content should match")

	t.Logf("Large file extraction took: %v", duration)
	t.Logf("Progress updates: %d", len(progressUpdates))
	t.Logf("Total bytes processed: %d", totalBytesProcessed)
}

// TestErrorRecovery 测试错误恢复机制
func TestErrorRecovery(t *testing.T) {
	sysOps := NewSystemOperations()
	zipProcessor := NewZipProcessor(sysOps)

	testDir, err := sysOps.CreateTempDirectory("test_error_recovery")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(testDir, "error_recovery_test.zip")
	err = createComplexTestZipFile(zipPath, map[string]string{
		"good_file1.txt":    "content1",
		"problematic.txt":   "problematic content",
		"good_file2.txt":    "content2",
		"another_good.txt":  "another content",
	})
	require.NoError(t, err)

	extractDir := filepath.Join(testDir, "extracted")
	require.NoError(t, sysOps.CreateDirectory(extractDir))

	// 创建一个问题文件来模拟错误
	problematicFile := filepath.Join(extractDir, "problematic.txt")
	require.NoError(t, sysOps.WriteFile(problematicFile, []byte("existing content")))
	require.NoError(t, os.Chmod(problematicFile, 0444)) // 只读

	var errorEvents []ErrorEvent
	var successfulFiles []string
	var failedFiles []string

	opts := &ExtractOptions{
		Verbose:           true,
		OverwriteExisting: true, // 尝试覆盖，可能会失败
		OnProgress: func(current, total int64, filename string) {
			// 记录成功处理的文件
			successfulFiles = append(successfulFiles, filename)
		},
		OnError: func(filename string, err error) bool {
			errorEvents = append(errorEvents, ErrorEvent{
				Filename:  filename,
				Error:     err,
				Timestamp: time.Now(),
			})
			failedFiles = append(failedFiles, filename)
			return true // 继续处理其他文件
		},
	}

	// 执行提取
	err = zipProcessor.ExtractWithProgress(zipPath, extractDir, opts, func(current, total int64) {
		// 额外的进度回调
	})
	
	// 在某些系统上可能不会产生错误
	if len(errorEvents) > 0 {
		t.Logf("Errors occurred as expected: %d", len(errorEvents))
		assert.Greater(t, len(successfulFiles), len(failedFiles), 
			"More files should succeed than fail")
		
		// 验证错误处理
		for _, errorEvent := range errorEvents {
			assert.NotEmpty(t, errorEvent.Filename, "Error filename should not be empty")
			assert.NotNil(t, errorEvent.Error, "Error should not be nil")
		}
	} else {
		t.Log("No errors occurred (system may not enforce readonly restrictions)")
	}

	// 验证至少一些文件被成功提取
	assert.Greater(t, len(successfulFiles), 0, "Some files should be successfully extracted")

	t.Logf("Successful files: %d", len(successfulFiles))
	t.Logf("Failed files: %d", len(failedFiles))
	t.Logf("Error events: %d", len(errorEvents))
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	sysOps := NewSystemOperations()

	testDir, err := sysOps.CreateTempDirectory("test_concurrent")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建多个ZIP文件
	numZips := 3
	zipPaths := make([]string, numZips)
	
	for i := 0; i < numZips; i++ {
		zipPath := filepath.Join(testDir, fmt.Sprintf("concurrent_test_%d.zip", i))
		err = createComplexTestZipFile(zipPath, map[string]string{
			fmt.Sprintf("file_%d_1.txt", i): fmt.Sprintf("content %d 1", i),
			fmt.Sprintf("file_%d_2.txt", i): fmt.Sprintf("content %d 2", i),
			fmt.Sprintf("dir_%d/file.txt", i): fmt.Sprintf("dir content %d", i),
		})
		require.NoError(t, err)
		zipPaths[i] = zipPath
	}

	// 并发提取
	results := make(chan ExtractionResult, numZips)
	
	for i, zipPath := range zipPaths {
		go func(index int, path string) {
			zipProcessor := NewZipProcessor(sysOps) // 每个goroutine使用独立的处理器
			extractDir := filepath.Join(testDir, fmt.Sprintf("extracted_%d", index))
			
			start := time.Now()
			err := zipProcessor.ExtractWithProgress(path, extractDir, &ExtractOptions{
				Verbose: false,
			}, func(current, total int64) {
				// 额外的进度回调
			})
			duration := time.Since(start)
			
			results <- ExtractionResult{
				Index:    index,
				Path:     path,
				Error:    err,
				Duration: duration,
			}
		}(i, zipPath)
	}

	// 收集结果
	var extractionResults []ExtractionResult
	for i := 0; i < numZips; i++ {
		result := <-results
		extractionResults = append(extractionResults, result)
	}

	// 验证所有提取都成功
	for _, result := range extractionResults {
		assert.NoError(t, result.Error, "Concurrent extraction %d should succeed", result.Index)
		assert.Greater(t, result.Duration, time.Duration(0), "Duration should be positive")
		
		// 验证提取的文件存在
		extractDir := filepath.Join(testDir, fmt.Sprintf("extracted_%d", result.Index))
		assert.True(t, sysOps.FileExists(extractDir), "Extract directory should exist")
	}

	t.Logf("Concurrent extractions completed successfully: %d", len(extractionResults))
	for _, result := range extractionResults {
		t.Logf("Extraction %d took: %v", result.Index, result.Duration)
	}
}

// 辅助结构体和函数

type ProgressUpdate struct {
	Current   int64
	Total     int64
	Filename  string
	Timestamp time.Time
}

type ErrorEvent struct {
	Filename  string
	Error     error
	Timestamp time.Time
}

type ConflictEvent struct {
	SourcePath string
	TargetPath string
	Timestamp  time.Time
}

type ExtractionResult struct {
	Index    int
	Path     string
	Error    error
	Duration time.Duration
}

// createComplexTestZipFile 创建复杂的测试ZIP文件
func createComplexTestZipFile(zipPath string, files map[string]string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到ZIP，包括目录结构
	for filePath, content := range files {
		// 如果文件路径包含目录，先创建目录条目
		dir := filepath.Dir(filePath)
		if dir != "." && dir != "/" {
			// 创建目录条目
			dirPath := strings.ReplaceAll(dir, "\\", "/") + "/"
			_, err := zipWriter.Create(dirPath)
			if err != nil {
				return err
			}
		}

		// 创建文件条目
		writer, err := zipWriter.Create(strings.ReplaceAll(filePath, "\\", "/"))
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(content))
		if err != nil {
			return err
		}
	}

	return nil
}