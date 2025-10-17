package infrastructure

import (
	"fmt"
	"os"
	"strings"

	"specify-cli/internal/types"
)

// AuthProvider 认证提供者实现
type AuthProvider struct {
	token string
}

// NewAuthProvider 创建新的认证提供者实例
func NewAuthProvider() types.AuthProvider {
	return &AuthProvider{}
}

// SetToken 设置认证令牌
func (ap *AuthProvider) SetToken(token string) {
	ap.token = strings.TrimSpace(token)
}

// GetToken 获取认证令牌
func (ap *AuthProvider) GetToken() string {
	if ap.token != "" {
		return ap.token
	}

	// 尝试从环境变量获取
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}

	return ""
}

// GetHeaders 获取认证头
func (ap *AuthProvider) GetHeaders() map[string]string {
	headers := make(map[string]string)
	
	// 设置User-Agent
	headers["User-Agent"] = "Specify-CLI/1.0.0"
	
	// 添加认证头
	if token := ap.GetToken(); token != "" {
		headers["Authorization"] = fmt.Sprintf("token %s", token)
	}
	
	// 设置Accept头
	headers["Accept"] = "application/vnd.github.v3+json"
	
	return headers
}

// IsAuthenticated 检查是否已认证
func (ap *AuthProvider) IsAuthenticated() bool {
	return ap.GetToken() != ""
}

// ValidateToken 验证令牌有效性
func (ap *AuthProvider) ValidateToken() error {
	token := ap.GetToken()
	if token == "" {
		return fmt.Errorf("no authentication token provided")
	}

	// 这里可以添加实际的令牌验证逻辑
	// 例如调用GitHub API验证令牌
	
	return nil
}

// GetAuthType 获取认证类型
func (ap *AuthProvider) GetAuthType() string {
	if ap.IsAuthenticated() {
		return "token"
	}
	return "none"
}

// ClearToken 清除令牌
func (ap *AuthProvider) ClearToken() {
	ap.token = ""
}

// LoadFromFile 从文件加载令牌
func (ap *AuthProvider) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return fmt.Errorf("token file is empty")
	}

	ap.SetToken(token)
	return nil
}

// SaveToFile 保存令牌到文件
func (ap *AuthProvider) SaveToFile(filePath string) error {
	token := ap.GetToken()
	if token == "" {
		return fmt.Errorf("no token to save")
	}

	err := os.WriteFile(filePath, []byte(token), 0600)
	if err != nil {
		return fmt.Errorf("failed to save token file: %w", err)
	}

	return nil
}