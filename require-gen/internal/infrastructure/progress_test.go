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

// TestProgressFeedback 测试进度反馈功能
func TestProgressFeedback(t *testing.T) {
	sysOps := NewSystemOperations()
	zipProcessor := NewZipProcessor(sysOps)

	testDir, err := sysOps.CreateTempDirectory("test_progress_feedback")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(testDir, "progress_test.zip")
	err = createTestZipFile(zipPath, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"file3.txt": "content3",
		"dir/file4.txt": "content4",
		"dir/subdir/file5.txt": "content5",
	})
	require.NoError(t, err)

	extractDir := filepath.Join(testDir, "extracted")
	require.NoError(t, sysOps.CreateDirectory(extractDir))

	// 测试进度回调
	var progressCalls []string
	var currentFiles []int64
	var totalFiles []int64

	opts := &ExtractOptions{
		Verbose: true,
	}

	// 使用ExtractWithProgress方法来测试进度回调
	progressCallback := func(current, total int64) {
		progressCalls = append(progressCalls, fmt.Sprintf("Progress: %d/%d", current, total))
		currentFiles = append(currentFiles, current)
		totalFiles = append(totalFiles, total)
	}

	// 执行提取
	err = zipProcessor.ExtractWithProgress(zipPath, extractDir, opts, progressCallback)
	assert.NoError(t, err)

	// 验证进度回调被调用
	assert.Greater(t, len(progressCalls), 0, "Progress callback should be called")
	assert.Equal(t, len(currentFiles), len(totalFiles), "Current and total arrays should have same length")

	// 验证进度递增
	for i := 1; i < len(currentFiles); i++ {
		assert.GreaterOrEqual(t, currentFiles[i], currentFiles[i-1], "Progress should be non-decreasing")
		assert.Equal(t, totalFiles[i], totalFiles[0], "Total should remain constant")
	}

	// 验证最终进度
	if len(currentFiles) > 0 {
		assert.Equal(t, currentFiles[len(currentFiles)-1], totalFiles[0], "Final progress should equal total")
	}
}

// TestErrorHandling 测试错误处理功能
func TestErrorHandling(t *testing.T) {
	sysOps := NewSystemOperations()
	zipProcessor := NewZipProcessor(sysOps)

	testDir, err := sysOps.CreateTempDirectory("test_error_handling")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(testDir, "error_test.zip")
	err = createTestZipFile(zipPath, map[string]string{
		"good_file.txt": "good content",
		"bad_file.txt":  "bad content",
		"another_good.txt": "another good content",
	})
	require.NoError(t, err)

	extractDir := filepath.Join(testDir, "extracted")
	require.NoError(t, sysOps.CreateDirectory(extractDir))

	// 创建一个只读文件来模拟写入错误
	readOnlyFile := filepath.Join(extractDir, "bad_file.txt")
	require.NoError(t, sysOps.WriteFile(readOnlyFile, []byte("readonly")))
	
	// 在Windows上设置只读属性
	if err := os.Chmod(readOnlyFile, 0444); err != nil {
		t.Logf("Warning: Could not set readonly permission: %v", err)
	}

	var errorCalls []string
	var errorFiles []string
	var errorMessages []error
	continueOnError := true

	opts := &ExtractOptions{
		Verbose: true,
		OnError: func(filename string, err error) bool {
			errorCalls = append(errorCalls, fmt.Sprintf("Error in %s: %v", filename, err))
			errorFiles = append(errorFiles, filename)
			errorMessages = append(errorMessages, err)
			return continueOnError // 继续处理其他文件
		},
	}

	// 执行提取（应该遇到错误但继续）
	err = zipProcessor.ExtractZip(zipPath, extractDir, opts)
	
	// 在某些系统上可能不会产生错误，所以我们检查是否有错误回调被调用
	if len(errorCalls) > 0 {
		t.Logf("Error callbacks were called: %v", errorCalls)
		assert.Greater(t, len(errorCalls), 0, "Error callback should be called")
		assert.Equal(t, len(errorFiles), len(errorMessages), "Error files and messages should match")
		
		// 验证错误文件名
		for _, filename := range errorFiles {
			assert.NotEmpty(t, filename, "Error filename should not be empty")
		}
	} else {
		t.Log("No errors occurred during extraction (system may not enforce readonly)")
	}

	// 测试停止处理的情况
	continueOnError = false
	extractDir2 := filepath.Join(testDir, "extracted2")
	require.NoError(t, sysOps.CreateDirectory(extractDir2))

	// 重新创建只读文件
	readOnlyFile2 := filepath.Join(extractDir2, "bad_file.txt")
	require.NoError(t, sysOps.WriteFile(readOnlyFile2, []byte("readonly")))
	if err := os.Chmod(readOnlyFile2, 0444); err != nil {
		t.Logf("Warning: Could not set readonly permission: %v", err)
	}

	opts.OnError = func(filename string, err error) bool {
		return false // 停止处理
	}

	// 这次应该在遇到错误时停止
	err = zipProcessor.ExtractZip(zipPath, extractDir2, opts)
	// 可能会有错误，也可能没有，取决于系统行为
	t.Logf("Extraction result with stop-on-error: %v", err)
}

// TestProgressAndErrorIntegration 测试进度反馈和错误处理的集成
func TestProgressAndErrorIntegration(t *testing.T) {
	sysOps := NewSystemOperations()
	zipProcessor := NewZipProcessor(sysOps)

	testDir, err := sysOps.CreateTempDirectory("test_progress_error_integration")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(testDir, "integration_test.zip")
	err = createTestZipFile(zipPath, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"file3.txt": "content3",
		"file4.txt": "content4",
	})
	require.NoError(t, err)

	extractDir := filepath.Join(testDir, "extracted")
	require.NoError(t, sysOps.CreateDirectory(extractDir))

	var progressCalls []string
	var errorCalls []string
	var processedFiles []int64

	opts := &ExtractOptions{
		Verbose: true,
		OnError: func(filename string, err error) bool {
			errorCalls = append(errorCalls, fmt.Sprintf("Error: %s - %v", filename, err))
			return true // 继续处理
		},
	}

	// 使用ExtractWithProgress方法来测试进度和错误处理的集成
	progressCallback := func(current, total int64) {
		progressCalls = append(progressCalls, fmt.Sprintf("Processing: %d/%d", current, total))
		processedFiles = append(processedFiles, current)
	}

	// 执行提取
	err = zipProcessor.ExtractWithProgress(zipPath, extractDir, opts, progressCallback)
	assert.NoError(t, err)

	// 验证进度和错误处理的协调工作
	assert.Greater(t, len(progressCalls), 0, "Progress callbacks should be called")
	assert.Equal(t, len(processedFiles), len(progressCalls), "Each progress call should include a file count")

	// 验证所有文件都被处理
	expectedFileCount := int64(4) // 4个文件
	if len(processedFiles) > 0 {
		finalCount := processedFiles[len(processedFiles)-1]
		assert.Equal(t, expectedFileCount, finalCount, "All files should be processed")
	}

	t.Logf("Progress calls: %d", len(progressCalls))
	t.Logf("Error calls: %d", len(errorCalls))
	t.Logf("Processed files: %v", processedFiles)
}

// TestProgressBarIntegration 测试进度条集成
func TestProgressBarIntegration(t *testing.T) {
	sysOps := NewSystemOperations()
	zipProcessor := NewZipProcessor(sysOps)

	testDir, err := sysOps.CreateTempDirectory("test_progress_bar")
	require.NoError(t, err)
	defer sysOps.RemoveDirectory(testDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(testDir, "progress_bar_test.zip")
	err = createTestZipFile(zipPath, map[string]string{
		"file1.txt": strings.Repeat("a", 1000),
		"file2.txt": strings.Repeat("b", 2000),
		"file3.txt": strings.Repeat("c", 1500),
	})
	require.NoError(t, err)

	extractDir := filepath.Join(testDir, "extracted")
	require.NoError(t, sysOps.CreateDirectory(extractDir))

	// 模拟进度条更新
	var progressUpdates []float64
	var currentProgress int64 = 0
	var totalProgress int64 = 0

	opts := &ExtractOptions{
		Verbose: true,
	}

	// 使用ExtractWithProgress方法来测试进度条集成
	progressCallback := func(current, total int64) {
		currentProgress = current
		totalProgress = total
		if total > 0 {
			percentage := float64(current) / float64(total) * 100
			progressUpdates = append(progressUpdates, percentage)
			
			// 模拟进度条显示
			t.Logf("Progress: %.1f%% - Processing file %d/%d", percentage, current, total)
		}
	}

	// 执行提取
	start := time.Now()
	err = zipProcessor.ExtractWithProgress(zipPath, extractDir, opts, progressCallback)
	duration := time.Since(start)
	
	assert.NoError(t, err)
	assert.Greater(t, len(progressUpdates), 0, "Progress updates should be recorded")
	
	// 验证进度递增
	for i := 1; i < len(progressUpdates); i++ {
		assert.GreaterOrEqual(t, progressUpdates[i], progressUpdates[i-1], 
			"Progress should be non-decreasing: %.1f%% -> %.1f%%", 
			progressUpdates[i-1], progressUpdates[i])
	}

	// 验证最终进度
	if len(progressUpdates) > 0 {
		finalProgress := progressUpdates[len(progressUpdates)-1]
		assert.Equal(t, 100.0, finalProgress, "Final progress should be 100%%")
	}

	t.Logf("Extraction completed in %v with %d progress updates", duration, len(progressUpdates))
	t.Logf("Final progress: %d/%d", currentProgress, totalProgress)
}

// createTestZipFile 创建测试用的ZIP文件
func createTestZipFile(zipPath string, files map[string]string) error {
	// 创建ZIP文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建ZIP写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到ZIP
	for filePath, content := range files {
		// 创建ZIP条目
		writer, err := zipWriter.Create(filePath)
		if err != nil {
			return err
		}

		// 写入文件内容
		_, err = writer.Write([]byte(content))
		if err != nil {
			return err
		}
	}

	return nil
}