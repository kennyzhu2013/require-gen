package infrastructure

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"specify-cli/internal/types"
)

// MockSystemOperations 模拟系统操作，用于测试
type MockSystemOperations struct {
	files       map[string]bool
	directories map[string]bool
	tempDirs    []string
}

func NewMockSystemOperations() *MockSystemOperations {
	return &MockSystemOperations{
		files:       make(map[string]bool),
		directories: make(map[string]bool),
		tempDirs:    make([]string, 0),
	}
}

func (m *MockSystemOperations) GetOS() string                   { return "linux" }
func (m *MockSystemOperations) GetArch() string                 { return "amd64" }
func (m *MockSystemOperations) GetSystemInfo() types.SystemInfo { return types.SystemInfo{} }
func (m *MockSystemOperations) ExecuteCommand(name string, args ...string) (*types.CommandResult, error) {
	return nil, nil
}
func (m *MockSystemOperations) ExecuteCommandInDir(dir, name string, args ...string) (*types.CommandResult, error) {
	return nil, nil
}
func (m *MockSystemOperations) ExecuteCommandWithOptions(name string, args []string, options *types.CommandOptions) (*types.CommandResult, error) {
	return nil, nil
}
func (m *MockSystemOperations) ExecuteCommandAsync(name string, args []string, options *types.CommandOptions) (<-chan *types.CommandResult, error) {
	return nil, nil
}
func (m *MockSystemOperations) CreateDirectory(path string) error {
	// 在真实文件系统中创建目录
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	
	// 确保父目录也被标记为存在
	dir := filepath.Clean(path)
	for dir != "." && dir != "/" && dir != "\\" {
		m.directories[dir] = true
		dir = filepath.Dir(dir)
		if dir == filepath.Dir(dir) {
			break // 避免无限循环
		}
	}
	return nil
}
func (m *MockSystemOperations) RemoveDirectory(path string) error {
	delete(m.directories, path)
	return nil
}
func (m *MockSystemOperations) DirectoryExists(path string) bool {
	// 检查路径是否存在，同时处理根目录情况
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == "/" || cleanPath == "\\" {
		return true // 根目录总是存在
	}
	return m.directories[cleanPath]
}
func (m *MockSystemOperations) ListDirectory(path string) ([]string, error) { return nil, nil }
func (m *MockSystemOperations) CopyFile(src, dst string) error              { return nil }
func (m *MockSystemOperations) MoveFile(src, dst string) error              { return nil }
func (m *MockSystemOperations) FileExists(path string) bool {
	return m.files[path]
}
func (m *MockSystemOperations) GetFileSize(path string) (int64, error)         { return 0, nil }
func (m *MockSystemOperations) GetFileModTime(path string) (int64, error)      { return 0, nil }
func (m *MockSystemOperations) GetCurrentDirectory() (string, error)           { return "/tmp", nil }
func (m *MockSystemOperations) ChangeDirectory(path string) error              { return nil }
func (m *MockSystemOperations) GetHomeDirectory() (string, error)              { return "/home/user", nil }
func (m *MockSystemOperations) GetTempDirectory() string                       { return "/tmp" }
func (m *MockSystemOperations) GetPathSeparator() string                       { return "/" }
func (m *MockSystemOperations) NormalizePath(path string) string               { return path }
func (m *MockSystemOperations) JoinPath(paths ...string) string                { return filepath.Join(paths...) }
func (m *MockSystemOperations) SetEnvironmentVariable(key, value string) error { return nil }
func (m *MockSystemOperations) GetEnvironmentVariable(key string) string       { return "" }
func (m *MockSystemOperations) GetAllEnvironmentVariables() []string           { return nil }
func (m *MockSystemOperations) IsExecutable(path string) bool                  { return false }
func (m *MockSystemOperations) FindExecutable(name string) (string, error)     { return "", nil }
func (m *MockSystemOperations) CheckPermissions(path, permission string) (bool, error) {
	return true, nil
}
func (m *MockSystemOperations) CreateTempFile(pattern string) (string, error) {
	return "/tmp/test", nil
}
func (m *MockSystemOperations) CreateTempDirectory(pattern string) (string, error) {
	dir := fmt.Sprintf("/tmp/%s", pattern)
	m.tempDirs = append(m.tempDirs, dir)
	return dir, nil
}
func (m *MockSystemOperations) GetExecutableExtension() string      { return "" }
func (m *MockSystemOperations) GetShellCommand() (string, []string) { return "sh", []string{"-c"} }
func (m *MockSystemOperations) DetectShellType() string             { return "bash" }
func (m *MockSystemOperations) GetPlatformSpecificPath(pathType string) (string, error) {
	return "/tmp", nil
}
func (m *MockSystemOperations) IsWindowsSystem() bool { return false }
func (m *MockSystemOperations) IsUnixSystem() bool    { return true }
func (m *MockSystemOperations) IsMacOSSystem() bool   { return false }
func (m *MockSystemOperations) IsLinuxSystem() bool   { return true }
func (m *MockSystemOperations) IsPathSafe(path string) (bool, error) {
	return !strings.Contains(path, ".."), nil
}
func (m *MockSystemOperations) IsCommandSafe(command string, args []string) (bool, error) {
	return true, nil
}
func (m *MockSystemOperations) ExtractZipArchive(zipPath, targetDir string, overwrite bool) error {
	return nil
}
func (m *MockSystemOperations) ValidateZipArchive(zipPath string) error { return nil }
func (m *MockSystemOperations) ListZipArchiveContents(zipPath string) ([]string, error) {
	return []string{"file1.txt", "dir1/file2.txt"}, nil
}

// createTestZip 创建测试用的ZIP文件
func createTestZip(t *testing.T, filename string, files map[string]string) {
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create test zip file: %v", err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("Failed to create zip entry %s: %v", name, err)
		}
		_, err = io.WriteString(writer, content)
		if err != nil {
			t.Fatalf("Failed to write zip entry %s: %v", name, err)
		}
	}
}

func TestNewZipProcessor(t *testing.T) {
	mockSysOps := NewMockSystemOperations()
	processor := NewZipProcessor(mockSysOps)

	if processor == nil {
		t.Fatal("NewZipProcessor returned nil")
	}

	impl, ok := processor.(*ZipProcessorImpl)
	if !ok {
		t.Fatal("NewZipProcessor did not return ZipProcessorImpl")
	}

	if impl.sysOps != mockSysOps {
		t.Fatal("ZipProcessorImpl does not have correct SystemOperations")
	}
}

func TestZipProcessor_ValidateZip(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ziptest_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"file1.txt":      "Hello World",
		"dir1/file2.txt": "Test Content",
		"dir1/file3.txt": "More Content",
	}
	createTestZip(t, zipPath, testFiles)

	mockSysOps := NewMockSystemOperations()
	mockSysOps.files[zipPath] = true
	processor := NewZipProcessor(mockSysOps)

	// 测试有效的ZIP文件
	err = processor.ValidateZip(zipPath)
	if err != nil {
		t.Errorf("ValidateZip failed for valid zip: %v", err)
	}

	// 测试空路径
	err = processor.ValidateZip("")
	if err == nil {
		t.Error("ValidateZip should fail for empty path")
	}

	// 测试不存在的文件
	err = processor.ValidateZip("/nonexistent/file.zip")
	if err == nil {
		t.Error("ValidateZip should fail for nonexistent file")
	}
}

func TestZipProcessor_ListZipContents(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ziptest_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"file1.txt":      "Hello World",
		"dir1/file2.txt": "Test Content",
		"dir1/file3.txt": "More Content",
	}
	createTestZip(t, zipPath, testFiles)

	mockSysOps := NewMockSystemOperations()
	mockSysOps.files[zipPath] = true
	processor := NewZipProcessor(mockSysOps)

	// 测试列出ZIP内容
	contents, err := processor.ListZipContents(zipPath)
	if err != nil {
		t.Errorf("ListZipContents failed: %v", err)
	}

	expectedFiles := []string{"file1.txt", "dir1/file2.txt", "dir1/file3.txt"}
	if len(contents) != len(expectedFiles) {
		t.Errorf("Expected %d files, got %d", len(expectedFiles), len(contents))
	}

	for _, expected := range expectedFiles {
		found := false
		for _, actual := range contents {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %s not found in contents", expected)
		}
	}
}

func TestZipProcessor_ExtractZip(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ziptest_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"file1.txt":      "Hello World",
		"dir1/file2.txt": "Test Content",
	}
	createTestZip(t, zipPath, testFiles)

	// 创建提取目录
	extractDir := filepath.Join(tempDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create extract dir: %v", err)
	}

	mockSysOps := NewMockSystemOperations()
	mockSysOps.files[zipPath] = true
	processor := NewZipProcessor(mockSysOps)

	// 配置提取选项
	opts := &ExtractOptions{
		OverwriteExisting: true,
		Verbose:           false,
	}

	// 测试ZIP提取
	err = processor.ExtractZip(zipPath, extractDir, opts)
	if err != nil {
		t.Errorf("ExtractZip failed: %v", err)
	}

	// 验证提取的文件
	extractedFile1 := filepath.Join(extractDir, "file1.txt")
	if _, err := os.Stat(extractedFile1); os.IsNotExist(err) {
		t.Errorf("Extracted file %s does not exist", extractedFile1)
	}

	extractedFile2 := filepath.Join(extractDir, "dir1", "file2.txt")
	if _, err := os.Stat(extractedFile2); os.IsNotExist(err) {
		t.Errorf("Extracted file %s does not exist", extractedFile2)
	}
}

func TestZipProcessor_ExtractWithProgress(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ziptest_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"file1.txt": "Hello World",
		"file2.txt": "Test Content",
		"file3.txt": "More Content",
	}
	createTestZip(t, zipPath, testFiles)

	// 创建提取目录
	extractDir := filepath.Join(tempDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create extract dir: %v", err)
	}

	mockSysOps := NewMockSystemOperations()
	mockSysOps.files[zipPath] = true
	processor := NewZipProcessor(mockSysOps)

	// 配置提取选项
	opts := &ExtractOptions{
		OverwriteExisting: true,
		Verbose:           true,
	}

	// 进度回调计数器
	progressCallCount := 0
	progressCallback := func(current, total int64) {
		progressCallCount++
		if total > 0 {
			// 简单验证进度回调被调用
			t.Logf("Progress: %d/%d bytes", current, total)
		}
	}

	// 测试带进度的ZIP提取
	err = processor.ExtractWithProgress(zipPath, extractDir, opts, progressCallback)
	if err != nil {
		t.Errorf("ExtractWithProgress failed: %v", err)
	}

	// 验证进度回调被调用
	if progressCallCount == 0 {
		t.Error("Progress callback was not called")
	}
}

func TestZipError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	zipErr := &ZipError{
		Operation: "extract",
		Path:      "/test/path.zip",
		Cause:     originalErr,
	}

	// 测试Error方法
	expectedMsg := "zip extract failed for /test/path.zip: original error"
	if zipErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, zipErr.Error())
	}

	// 测试Unwrap方法
	if zipErr.Unwrap() != originalErr {
		t.Error("Unwrap did not return original error")
	}
}

func TestExtractOptions_Defaults(t *testing.T) {
	opts := &ExtractOptions{}

	// 测试默认值
	if opts.OverwriteExisting {
		t.Error("Default Overwrite should be false")
	}
	if opts.FlattenStructure {
		t.Error("Default FlattenDirs should be false")
	}
	if opts.MaxFileSize != 0 {
		t.Error("Default MaxFileSize should be 0")
	}
}

func TestSystemOperations_ZipIntegration(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ziptest_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试ZIP文件
	zipPath := filepath.Join(tempDir, "test.zip")
	testFiles := map[string]string{
		"file1.txt": "Hello World",
		"file2.txt": "Test Content",
	}
	createTestZip(t, zipPath, testFiles)

	// 创建提取目录
	extractDir := filepath.Join(tempDir, "extract")

	// 使用真实的SystemOperations
	sysOps := NewSystemOperations()

	// 测试ValidateZipArchive
	err = sysOps.ValidateZipArchive(zipPath)
	if err != nil {
		t.Errorf("ValidateZipArchive failed: %v", err)
	}

	// 测试ListZipArchiveContents
	contents, err := sysOps.ListZipArchiveContents(zipPath)
	if err != nil {
		t.Errorf("ListZipArchiveContents failed: %v", err)
	}
	if len(contents) != 2 {
		t.Errorf("Expected 2 files, got %d", len(contents))
	}

	// 测试ExtractZipArchive
	err = sysOps.ExtractZipArchive(zipPath, extractDir, true)
	if err != nil {
		t.Errorf("ExtractZipArchive failed: %v", err)
	}

	// 验证提取的文件
	extractedFile1 := filepath.Join(extractDir, "file1.txt")
	if !sysOps.FileExists(extractedFile1) {
		t.Errorf("Extracted file %s does not exist", extractedFile1)
	}
}
