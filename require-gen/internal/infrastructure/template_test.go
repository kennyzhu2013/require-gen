package infrastructure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"specify-cli/internal/types"
)

// mockAuthProvider 模拟认证提供者
type mockAuthProvider struct{}

func (m *mockAuthProvider) GetHeaders() map[string]string {
	return map[string]string{
		"User-Agent": "Test-Agent/1.0.0",
	}
}

func (m *mockAuthProvider) GetToken() string {
	return "mock-token"
}

func (m *mockAuthProvider) GetAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer mock-token",
		"User-Agent":    "Test-Agent/1.0.0",
	}
}

func (m *mockAuthProvider) SetToken(token string) {}

func (m *mockAuthProvider) SetCLIToken(token string) {}

func (m *mockAuthProvider) IsAuthenticated() bool {
	return true
}

func (m *mockAuthProvider) ValidateToken() error {
	return nil
}

func (m *mockAuthProvider) GetTokenScopes() ([]string, error) {
	return []string{"repo"}, nil
}

func (m *mockAuthProvider) GetAuthType() string {
	return "bearer"
}

func (m *mockAuthProvider) ClearToken() {}

func (m *mockAuthProvider) ClearCLIToken() {}

func (m *mockAuthProvider) GetUserInfo() (map[string]interface{}, error) {
	return map[string]interface{}{"login": "testuser"}, nil
}

func (m *mockAuthProvider) GetTokenSource() string {
	return "mock"
}

func (m *mockAuthProvider) LoadFromFile(filePath string) error {
	return nil
}

func (m *mockAuthProvider) SaveToFile(filePath string) error {
	return nil
}

func (m *mockAuthProvider) IsTokenExpired() (bool, error) {
	return false, nil
}

// MockTemplateProvider 模拟模板提供者，用于测试
type MockTemplateProvider struct {
	downloadResult string
	downloadError  error
	validateError  error
	templateInfo   map[string]interface{}
	templates      []string
}

func (m *MockTemplateProvider) Download(opts types.DownloadOptions) (string, error) {
	if m.downloadError != nil {
		return "", m.downloadError
	}
	return m.downloadResult, nil
}

func (m *MockTemplateProvider) Validate(path string) error {
	return m.validateError
}

func (m *MockTemplateProvider) GetTemplateInfo(path string) (map[string]interface{}, error) {
	if m.templateInfo == nil {
		return nil, fmt.Errorf("template info not found")
	}
	return m.templateInfo, nil
}

func (m *MockTemplateProvider) ListTemplates(token string) ([]string, error) {
	return m.templates, nil
}

// 创建测试用的GitHub Release响应
func createMockGitHubRelease() *types.GitHubRelease {
	return &types.GitHubRelease{
		TagName: "v1.0.0",
		Assets: []types.Asset{
			{
				Name:               "claude-powershell.zip",
				Size:               1024,
				BrowserDownloadURL: "https://github.com/test/repo/releases/download/v1.0.0/claude-powershell.zip",
			},
			{
				Name:               "copilot-bash.zip",
				Size:               2048,
				BrowserDownloadURL: "https://github.com/test/repo/releases/download/v1.0.0/copilot-bash.zip",
			},
			{
				Name:               "templates.zip",
				Size:               4096,
				BrowserDownloadURL: "https://github.com/test/repo/releases/download/v1.0.0/templates.zip",
			},
		},
	}
}

// 创建模拟的HTTP服务器
func createMockGitHubServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/releases/latest"):
			release := createMockGitHubRelease()
			data, err := json.Marshal(release)
			require.NoError(t, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case strings.Contains(r.URL.Path, "/releases/download/"):
			// 模拟文件下载
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("mock zip content"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// 创建测试目录结构
func createTestTemplateStructure(t *testing.T, baseDir string) {
	// 创建必需的文件
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "scripts"), 0755))
	
	// 创建README.md
	readmePath := filepath.Join(baseDir, "README.md")
	require.NoError(t, ioutil.WriteFile(readmePath, []byte("# Test Template"), 0644))
	
	// 创建spec-template.md
	specPath := filepath.Join(baseDir, "spec-template.md")
	require.NoError(t, ioutil.WriteFile(specPath, []byte("# Spec Template"), 0644))
	
	// 创建template-info.json
	infoPath := filepath.Join(baseDir, "template-info.json")
	info := map[string]interface{}{
		"name":        "test-template",
		"version":     "1.0.0",
		"description": "Test template for unit testing",
		"author":      "Test Author",
	}
	data, err := json.Marshal(info)
	require.NoError(t, err)
	require.NoError(t, ioutil.WriteFile(infoPath, data, 0644))
}

func TestNewTemplateProvider(t *testing.T) {
	provider := NewTemplateProvider()
	assert.NotNil(t, provider)
	
	// 验证类型
	_, ok := provider.(*TemplateProvider)
	assert.True(t, ok, "Should return *TemplateProvider")
}

func TestTemplateProvider_findAsset(t *testing.T) {
	provider := &TemplateProvider{}
	release := createMockGitHubRelease()
	
	tests := []struct {
		name      string
		assistant string
		scriptType string
		expected  string
		shouldErr bool
	}{
		{
			name:      "exact match claude-powershell",
			assistant: "claude",
			scriptType: "powershell",
			expected:  "claude-powershell.zip",
			shouldErr: false,
		},
		{
			name:      "exact match copilot-bash",
			assistant: "copilot",
			scriptType: "bash",
			expected:  "copilot-bash.zip",
			shouldErr: false,
		},
		{
			name:      "fallback to templates.zip",
			assistant: "unknown",
			scriptType: "unknown",
			expected:  "templates.zip",
			shouldErr: false,
		},
		{
			name:      "no match found",
			assistant: "nonexistent",
			scriptType: "nonexistent",
			expected:  "",
			shouldErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为不匹配的测试修改release
			testRelease := release
			if tt.shouldErr && tt.name == "no match found" {
				testRelease = &types.GitHubRelease{
					TagName: "v1.0.0",
					Assets: []types.Asset{
						{
							Name: "other-file.zip",
							Size: 1024,
						},
					},
				}
			}
			
			asset, err := provider.findAsset(testRelease, tt.assistant, tt.scriptType)
			
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, asset)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, asset)
				assert.Equal(t, tt.expected, asset.Name)
			}
		})
	}
}

func TestTemplateProvider_Validate(t *testing.T) {
	provider := &TemplateProvider{}
	
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "template_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	t.Run("valid template structure", func(t *testing.T) {
		validDir := filepath.Join(tempDir, "valid")
		createTestTemplateStructure(t, validDir)
		
		err := provider.Validate(validDir)
		assert.NoError(t, err)
	})
	
	t.Run("missing directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		
		err := provider.Validate(nonExistentDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
	
	t.Run("missing required files", func(t *testing.T) {
		incompleteDir := filepath.Join(tempDir, "incomplete")
		require.NoError(t, os.MkdirAll(incompleteDir, 0755))
		
		err := provider.Validate(incompleteDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required file missing")
	})
	
	t.Run("missing required directories", func(t *testing.T) {
		noDirsDir := filepath.Join(tempDir, "nodirs")
		require.NoError(t, os.MkdirAll(noDirsDir, 0755))
		
		// 只创建文件，不创建目录
		readmePath := filepath.Join(noDirsDir, "README.md")
		require.NoError(t, ioutil.WriteFile(readmePath, []byte("# Test"), 0644))
		specPath := filepath.Join(noDirsDir, "spec-template.md")
		require.NoError(t, ioutil.WriteFile(specPath, []byte("# Spec"), 0644))
		
		err := provider.Validate(noDirsDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid template structure")
	})
}

func TestTemplateProvider_GetTemplateInfo(t *testing.T) {
	provider := &TemplateProvider{}
	
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "template_info_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	t.Run("valid template info", func(t *testing.T) {
		createTestTemplateStructure(t, tempDir)
		
		info, err := provider.GetTemplateInfo(tempDir)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test-template", info["name"])
		assert.Equal(t, "1.0.0", info["version"])
	})
	
	t.Run("missing template info file", func(t *testing.T) {
		emptyDir := filepath.Join(tempDir, "empty")
		require.NoError(t, os.MkdirAll(emptyDir, 0755))
		
		info, err := provider.GetTemplateInfo(emptyDir)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "template info file not found")
	})
	
	t.Run("invalid json in template info", func(t *testing.T) {
		invalidDir := filepath.Join(tempDir, "invalid")
		require.NoError(t, os.MkdirAll(invalidDir, 0755))
		
		infoPath := filepath.Join(invalidDir, "template-info.json")
		require.NoError(t, ioutil.WriteFile(infoPath, []byte("invalid json"), 0644))
		
		info, err := provider.GetTemplateInfo(invalidDir)
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "failed to parse template info")
	})
}

func TestTemplateProvider_validateTemplateStructure(t *testing.T) {
	provider := &TemplateProvider{}
	
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "structure_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	t.Run("valid structure", func(t *testing.T) {
		validDir := filepath.Join(tempDir, "valid")
		createTestTemplateStructure(t, validDir)
		
		err := provider.validateTemplateStructure(validDir)
		assert.NoError(t, err)
	})
	
	t.Run("missing templates directory", func(t *testing.T) {
		invalidDir := filepath.Join(tempDir, "invalid")
		require.NoError(t, os.MkdirAll(filepath.Join(invalidDir, "scripts"), 0755))
		
		err := provider.validateTemplateStructure(invalidDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected directory not found: templates")
	})
	
	t.Run("missing scripts directory", func(t *testing.T) {
		invalidDir := filepath.Join(tempDir, "invalid2")
		require.NoError(t, os.MkdirAll(filepath.Join(invalidDir, "templates"), 0755))
		
		err := provider.validateTemplateStructure(invalidDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected directory not found: scripts")
	})
}

// 测试进度读取器
func TestProgressReader(t *testing.T) {
	data := []byte("test data for progress reader")
	reader := &progressReader{
		Reader: strings.NewReader(string(data)),
		total:  int64(len(data)),
	}
	
	buffer := make([]byte, 10)
	n, err := reader.Read(buffer)
	
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, "test data ", string(buffer))
}

// TestTemplateProvider_Download_Integration 集成测试
func TestTemplateProvider_Download_Integration(t *testing.T) {
	tests := []struct {
		name        string
		opts        types.DownloadOptions
		expectError bool
	}{
		{
			name: "invalid AI assistant",
			opts: types.DownloadOptions{
				AIAssistant: "invalid-assistant",
				ScriptType:  "powershell",
				DownloadDir: t.TempDir(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewTemplateProvider()
			_, err := provider.Download(tt.opts)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTemplateProvider_getLatestRelease 测试获取最新发布信息
func TestTemplateProvider_getLatestRelease(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "without token",
			token:       "",
			expectError: true, // 预期会失败，因为没有真实的GitHub仓库
		},
		{
			name:        "with token",
			token:       "fake-token",
			expectError: true, // 预期会失败，因为没有真实的GitHub仓库
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &TemplateProvider{
				client: resty.New(),
			}
			
			_, err := provider.getLatestRelease(tt.token)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTemplateProvider_downloadAsset 测试资源下载
func TestTemplateProvider_downloadAsset(t *testing.T) {
	// 创建模拟HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake zip content"))
	}))
	defer server.Close()

	asset := &types.Asset{
		Name:               "test.zip",
		BrowserDownloadURL: server.URL,
	}

	tempDir := t.TempDir()
	downloadPath := filepath.Join(tempDir, "test.zip")

	opts := types.DownloadOptions{
		ShowProgress: false, // 关闭进度显示以避免复杂的测试设置
	}

	provider := &TemplateProvider{
		client:       resty.New(),
		authProvider: &mockAuthProvider{}, // 使用模拟的认证提供者
	}

	err := provider.downloadAsset(asset, downloadPath, opts)
	assert.NoError(t, err)

	// 验证文件是否存在
	_, err = os.Stat(downloadPath)
	assert.NoError(t, err)
}

// TestTemplateProvider_extractAsset 测试资源提取
func TestTemplateProvider_extractAsset(t *testing.T) {
	tempDir := t.TempDir()
	
	// 创建一个假的zip文件
	zipPath := filepath.Join(tempDir, "test.zip")
	file, err := os.Create(zipPath)
	assert.NoError(t, err)
	file.WriteString("fake zip content")
	file.Close()

	opts := types.DownloadOptions{
		ShowProgress: true,
	}

	provider := &TemplateProvider{}
	
	// 这个测试预期会失败，因为文件不是真正的zip格式
	err = provider.extractAsset(zipPath, tempDir, opts)
	assert.Error(t, err) // 预期错误，因为不是真正的zip文件
}

// TestTemplateProvider_ListTemplates 测试列出模板
func TestTemplateProvider_ListTemplates(t *testing.T) {
	provider := NewTemplateProvider()
	
	// 这个测试预期会失败，因为没有真实的GitHub仓库
	_, err := provider.ListTemplates("")
	assert.Error(t, err)
}

// 基准测试
func BenchmarkTemplateProvider_Validate(b *testing.B) {
	provider := &TemplateProvider{}
	
	// 创建测试目录
	tempDir, err := ioutil.TempDir("", "benchmark_test")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)
	
	// 创建测试结构 - 使用testing.TB接口
	createBenchmarkTemplateStructure(b, tempDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.Validate(tempDir)
	}
}

func BenchmarkTemplateProvider_GetTemplateInfo(b *testing.B) {
	provider := &TemplateProvider{}
	
	// 创建测试目录
	tempDir, err := ioutil.TempDir("", "benchmark_info_test")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)
	
	// 创建测试结构 - 使用testing.TB接口
	createBenchmarkTemplateStructure(b, tempDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.GetTemplateInfo(tempDir)
	}
}

// 为基准测试创建模板结构的辅助函数
func createBenchmarkTemplateStructure(b *testing.B, baseDir string) {
	// 创建必需的文件
	require.NoError(b, os.MkdirAll(filepath.Join(baseDir, "templates"), 0755))
	require.NoError(b, os.MkdirAll(filepath.Join(baseDir, "scripts"), 0755))
	
	// 创建README.md
	readmePath := filepath.Join(baseDir, "README.md")
	require.NoError(b, ioutil.WriteFile(readmePath, []byte("# Test Template"), 0644))
	
	// 创建spec-template.md
	specPath := filepath.Join(baseDir, "spec-template.md")
	require.NoError(b, ioutil.WriteFile(specPath, []byte("# Spec Template"), 0644))
	
	// 创建template-info.json
	infoPath := filepath.Join(baseDir, "template-info.json")
	info := map[string]interface{}{
		"name":        "test-template",
		"version":     "1.0.0",
		"description": "Test template for unit testing",
		"author":      "Test Author",
	}
	data, err := json.Marshal(info)
	require.NoError(b, err)
	require.NoError(b, ioutil.WriteFile(infoPath, data, 0644))
}