package infrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"specify-cli/internal/types"
)

// AuthError 认证错误类型
//
// AuthError 提供了详细的认证错误信息，包括错误类型、原因和建议的解决方案。
// 这有助于用户理解认证失败的具体原因并采取相应的修复措施。
type AuthError struct {
	Type        string `json:"type"`        // 错误类型
	Message     string `json:"message"`     // 错误消息
	Cause       string `json:"cause"`       // 错误原因
	Suggestion  string `json:"suggestion"`  // 解决建议
	StatusCode  int    `json:"status_code"` // HTTP状态码（如果适用）
	TokenSource string `json:"token_source"` // 令牌来源
}

// Error 实现error接口
func (e *AuthError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// DetailedError 返回详细的错误信息
func (e *AuthError) DetailedError() string {
	var details []string
	details = append(details, fmt.Sprintf("Type: %s", e.Type))
	details = append(details, fmt.Sprintf("Message: %s", e.Message))
	
	if e.Cause != "" {
		details = append(details, fmt.Sprintf("Cause: %s", e.Cause))
	}
	
	if e.Suggestion != "" {
		details = append(details, fmt.Sprintf("Suggestion: %s", e.Suggestion))
	}
	
	if e.StatusCode > 0 {
		details = append(details, fmt.Sprintf("Status Code: %d", e.StatusCode))
	}
	
	if e.TokenSource != "" {
		details = append(details, fmt.Sprintf("Token Source: %s", e.TokenSource))
	}
	
	return strings.Join(details, "\n")
}

// NewAuthError 创建新的认证错误
func NewAuthError(errorType, message string) *AuthError {
	return &AuthError{
		Type:    errorType,
		Message: message,
	}
}

// NewAuthErrorWithDetails 创建带详细信息的认证错误
func NewAuthErrorWithDetails(errorType, message, cause, suggestion string) *AuthError {
	return &AuthError{
		Type:       errorType,
		Message:    message,
		Cause:      cause,
		Suggestion: suggestion,
	}
}

// NewAuthErrorWithStatus 创建带HTTP状态码的认证错误
func NewAuthErrorWithStatus(errorType, message string, statusCode int) *AuthError {
	return &AuthError{
		Type:       errorType,
		Message:    message,
		StatusCode: statusCode,
	}
}

// 预定义的错误类型常量
const (
	ErrorTypeTokenNotFound    = "TOKEN_NOT_FOUND"
	ErrorTypeTokenInvalid     = "TOKEN_INVALID"
	ErrorTypeTokenExpired     = "TOKEN_EXPIRED"
	ErrorTypeTokenFormat      = "TOKEN_FORMAT_ERROR"
	ErrorTypeAPIError         = "API_ERROR"
	ErrorTypeNetworkError     = "NETWORK_ERROR"
	ErrorTypePermissionDenied = "PERMISSION_DENIED"
	ErrorTypeFileError        = "FILE_ERROR"
	ErrorTypeValidationError  = "VALIDATION_ERROR"
)

// 预定义的错误创建函数
func NewTokenNotFoundError(source string) *AuthError {
	return &AuthError{
		Type:        ErrorTypeTokenNotFound,
		Message:     "Authentication token not found",
		Cause:       fmt.Sprintf("No valid token found from %s", source),
		Suggestion:  "Please set GITHUB_TOKEN or GH_TOKEN environment variable, or provide token via CLI",
		TokenSource: source,
	}
}

func NewTokenInvalidError(reason string) *AuthError {
	return &AuthError{
		Type:       ErrorTypeTokenInvalid,
		Message:    "Authentication token is invalid",
		Cause:      reason,
		Suggestion: "Please check your token and ensure it has the required permissions",
	}
}

func NewTokenExpiredError() *AuthError {
	return &AuthError{
		Type:       ErrorTypeTokenExpired,
		Message:    "Authentication token has expired",
		Cause:      "The provided token is no longer valid",
		Suggestion: "Please generate a new token and update your configuration",
	}
}

func NewTokenFormatError(details string) *AuthError {
	return &AuthError{
		Type:       ErrorTypeTokenFormat,
		Message:    "Token format is invalid",
		Cause:      details,
		Suggestion: "GitHub tokens should be 40 characters long and contain only alphanumeric characters and underscores",
	}
}

func NewAPIError(statusCode int, message string) *AuthError {
	return &AuthError{
		Type:       ErrorTypeAPIError,
		Message:    fmt.Sprintf("GitHub API error: %s", message),
		StatusCode: statusCode,
		Suggestion: "Please check GitHub API status and your network connection",
	}
}

func NewNetworkError(err error) *AuthError {
	return &AuthError{
		Type:       ErrorTypeNetworkError,
		Message:    "Network connection failed",
		Cause:      err.Error(),
		Suggestion: "Please check your internet connection and try again",
	}
}

func NewPermissionDeniedError(resource string) *AuthError {
	return &AuthError{
		Type:       ErrorTypePermissionDenied,
		Message:    fmt.Sprintf("Permission denied for %s", resource),
		Cause:      "Token does not have sufficient permissions",
		Suggestion: "Please ensure your token has the required scopes for this operation",
	}
}

func NewFileError(operation, path string, err error) *AuthError {
	return &AuthError{
		Type:       ErrorTypeFileError,
		Message:    fmt.Sprintf("File %s failed for %s", operation, path),
		Cause:      err.Error(),
		Suggestion: "Please check file permissions and path accessibility",
	}
}

// AuthProvider 认证提供者实现
type AuthProvider struct {
	token    string
	cliToken string // CLI参数传入的令牌
}

// TokenSource 令牌来源类型
type TokenSource int

const (
	TokenSourceCLI TokenSource = iota
	TokenSourceGHToken
	TokenSourceGitHubToken
	TokenSourceFile
	TokenSourceNone
)

// TokenInfo 令牌信息
type TokenInfo struct {
	Token  string
	Source TokenSource
	Valid  bool
}

// NewAuthProvider 创建新的认证提供者实例
//
// 该函数创建一个新的AuthProvider实例，用于管理GitHub API的认证。
// 创建的实例支持多种令牌来源和完整的认证功能。
//
// 返回值：
//   - types.AuthProvider: 新创建的认证提供者实例
//
// 使用示例：
//   auth := NewAuthProvider()
//   auth.SetCLIToken("your-cli-token")  // 设置CLI令牌（最高优先级）
//   if auth.IsAuthenticated() {
//       headers := auth.GetHeaders()
//       // 使用headers进行API调用
//   }
func NewAuthProvider() types.AuthProvider {
	return &AuthProvider{}
}

// ValidateAndSetToken 验证并设置令牌
//
// 该方法结合了令牌验证和设置功能，确保只有有效的令牌被设置。
// 这提供了一个安全的方式来设置令牌，避免设置无效的令牌。
//
// 参数：
//   - token: 要验证和设置的令牌
//
// 返回值：
//   - error: 如果令牌无效或验证失败，返回相应的错误
//
// 使用示例：
//   auth := NewAuthProvider()
//   if err := auth.ValidateAndSetToken("your-token"); err != nil {
//       log.Printf("Token validation failed: %v", err)
//   }
func (ap *AuthProvider) ValidateAndSetToken(token string) error {
	if token == "" {
		return NewTokenFormatError("token cannot be empty")
	}
	
	// 格式验证
	if err := ap.validateTokenFormat(token); err != nil {
		return err
	}
	
	// 临时设置令牌进行API验证
	originalToken := ap.token
	ap.token = token
	
	// API验证
	if err := ap.validateTokenWithAPI(token); err != nil {
		// 验证失败，恢复原始令牌
		ap.token = originalToken
		return err
	}
	
	// 验证成功，令牌已设置
	return nil
}

// HasValidToken 检查是否有有效的令牌
//
// 该方法检查当前是否有可用的有效令牌，包括格式验证和API验证。
// 这是一个便捷方法，用于快速检查认证状态。
//
// 返回值：
//   - bool: 如果有有效令牌返回true，否则返回false
//
// 使用示例：
//   auth := NewAuthProvider()
//   if auth.HasValidToken() {
//       // 可以安全地进行API调用
//       userInfo, _ := auth.GetUserInfo()
//   }
func (ap *AuthProvider) HasValidToken() bool {
	return ap.ValidateToken() == nil
}

// RefreshTokenValidation 刷新令牌验证状态
//
// 该方法重新验证当前令牌的有效性，用于定期检查令牌状态。
// 这对于长时间运行的应用程序特别有用。
//
// 返回值：
//   - error: 如果令牌无效或验证失败，返回相应的错误
//
// 使用示例：
//   auth := NewAuthProvider()
//   if err := auth.RefreshTokenValidation(); err != nil {
//       log.Printf("Token validation refresh failed: %v", err)
//       // 可能需要重新获取令牌
//   }
func (ap *AuthProvider) RefreshTokenValidation() error {
	return ap.ValidateToken()
}

// NewAuthProviderWithToken 使用CLI令牌创建认证提供者
func NewAuthProviderWithToken(cliToken string) types.AuthProvider {
	return &AuthProvider{
		cliToken: strings.TrimSpace(cliToken),
	}
}

// SetToken 设置认证令牌
func (ap *AuthProvider) SetToken(token string) {
	ap.token = strings.TrimSpace(token)
}

// SetCLIToken 设置CLI参数令牌（最高优先级）
func (ap *AuthProvider) SetCLIToken(token string) {
	ap.cliToken = strings.TrimSpace(token)
}

// GetToken 获取认证令牌（按优先级顺序）
func (ap *AuthProvider) GetToken() string {
	// 1. CLI参数令牌（最高优先级）
	if ap.cliToken != "" {
		return ap.cliToken
	}

	// 2. 手动设置的令牌
	if ap.token != "" {
		return ap.token
	}

	// 3. GH_TOKEN环境变量（GitHub CLI标准）
	if token := strings.TrimSpace(os.Getenv("GH_TOKEN")); token != "" {
		return token
	}

	// 4. GITHUB_TOKEN环境变量（GitHub Actions标准）
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		return token
	}

	return ""
}

// GetTokenInfo 获取详细的令牌信息
//
// 该方法返回当前令牌的详细信息，包括来源、有效性状态等。
// 这对于调试和监控认证状态非常有用。
//
// 返回值：
//   - *TokenInfo: 令牌详细信息，包含来源、有效性等信息
//
// 使用示例：
//   auth := NewAuthProvider()
//   info := auth.GetTokenInfo()
//   if info.Source != TokenSourceNone {
//       log.Printf("Token source: %d, Valid: %t", info.Source, info.Valid)
//   }
func (ap *AuthProvider) GetTokenInfo() *TokenInfo {
	info := &TokenInfo{}

	// 按优先级检查令牌来源
	if ap.cliToken != "" {
		info.Token = ap.cliToken
		info.Source = TokenSourceCLI
	} else if ap.token != "" {
		info.Token = ap.token
		info.Source = TokenSourceFile
	} else if token := strings.TrimSpace(os.Getenv("GH_TOKEN")); token != "" {
		info.Token = token
		info.Source = TokenSourceGHToken
	} else if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		info.Token = token
		info.Source = TokenSourceGitHubToken
	} else {
		info.Source = TokenSourceNone
	}

	// 设置令牌有效性
	if info.Token != "" {
		info.Valid = ap.ValidateToken() == nil
	}

	return info
}

// GetHeaders 获取认证头（使用Bearer格式）
func (ap *AuthProvider) GetHeaders() map[string]string {
	headers := make(map[string]string)
	
	// 设置User-Agent
	headers["User-Agent"] = "Specify-CLI/1.0.0"
	
	// 设置Accept头
	headers["Accept"] = "application/vnd.github.v3+json"
	
	// 添加认证头（仅在有效令牌存在时）
	if token := ap.GetToken(); token != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}
	
	return headers
}

// GetAuthHeaders 获取GitHub API认证头（兼容方法）
func (ap *AuthProvider) GetAuthHeaders() map[string]string {
	token := ap.GetToken()
	if token == "" {
		return map[string]string{
			"User-Agent": "Specify-CLI/1.0.0",
			"Accept":     "application/vnd.github.v3+json",
		}
	}
	
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"User-Agent":    "Specify-CLI/1.0.0",
		"Accept":        "application/vnd.github.v3+json",
	}
}

// IsAuthenticated 检查是否已认证
func (ap *AuthProvider) IsAuthenticated() bool {
	return ap.GetToken() != ""
}

// ValidateToken 验证令牌有效性
func (ap *AuthProvider) ValidateToken() error {
	token := ap.GetToken()
	if token == "" {
		return NewTokenNotFoundError(ap.GetTokenSource())
	}

	// 验证令牌格式
	if err := ap.validateTokenFormat(token); err != nil {
		return err
	}

	// 通过GitHub API验证令牌
	return ap.validateTokenWithAPI(token)
}

// validateTokenFormat 验证令牌格式
func (ap *AuthProvider) validateTokenFormat(token string) error {
	// GitHub Personal Access Token 格式验证
	if len(token) < 20 {
		return NewTokenFormatError("token too short (minimum 20 characters)")
	}

	// 检查是否包含非法字符
	for _, char := range token {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_') {
			return NewTokenFormatError("token contains invalid characters")
		}
	}

	return nil
}

// validateTokenWithAPI 通过GitHub API验证令牌
func (ap *AuthProvider) validateTokenWithAPI(token string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return NewNetworkError(fmt.Errorf("failed to create validation request: %w", err))
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "Specify-CLI/1.0.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return NewNetworkError(fmt.Errorf("failed to validate token: %w", err))
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 401:
		return NewTokenInvalidError("invalid or expired token")
	case 403:
		return NewPermissionDeniedError("user information access")
	default:
		return NewAPIError(resp.StatusCode, fmt.Sprintf("token validation failed with status %d", resp.StatusCode))
	}
}

// GetTokenScopes 获取令牌权限范围
func (ap *AuthProvider) GetTokenScopes() ([]string, error) {
	token := ap.GetToken()
	if token == "" {
		return nil, NewTokenNotFoundError(ap.GetTokenSource())
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, NewNetworkError(fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "Specify-CLI/1.0.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, NewNetworkError(fmt.Errorf("failed to get token scopes: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 401:
			return nil, NewTokenInvalidError("token is invalid or expired")
		case 403:
			return nil, NewPermissionDeniedError("token scope information")
		default:
			return nil, NewAPIError(resp.StatusCode, fmt.Sprintf("failed to get scopes with status %d", resp.StatusCode))
		}
	}

	// 从响应头获取权限范围
	scopes := resp.Header.Get("X-OAuth-Scopes")
	if scopes == "" {
		return []string{}, nil
	}

	return strings.Split(scopes, ", "), nil
}

// GetAuthType 获取认证类型
func (ap *AuthProvider) GetAuthType() string {
	if ap.IsAuthenticated() {
		return "bearer"
	}
	return "none"
}

// ClearToken 清除令牌
func (ap *AuthProvider) ClearToken() {
	ap.token = ""
	ap.cliToken = ""
}

// ClearCLIToken 仅清除CLI令牌
func (ap *AuthProvider) ClearCLIToken() {
	ap.cliToken = ""
}

// LoadFromFile 从文件加载令牌
func (ap *AuthProvider) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return NewFileError("read", filePath, err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return NewTokenFormatError("file contains empty token")
	}

	// 验证令牌格式
	if err := ap.validateTokenFormat(token); err != nil {
		return err
	}

	ap.SetToken(token)
	return nil
}

// SaveToFile 保存令牌到文件
func (ap *AuthProvider) SaveToFile(filePath string) error {
	token := ap.GetToken()
	if token == "" {
		return NewTokenNotFoundError(ap.GetTokenSource())
	}

	// 使用安全的文件权限（仅所有者可读写）
	err := os.WriteFile(filePath, []byte(token), 0600)
	if err != nil {
		return NewFileError("write", filePath, err)
	}

	return nil
}

// GetUserInfo 获取认证用户信息
func (ap *AuthProvider) GetUserInfo() (map[string]interface{}, error) {
	token := ap.GetToken()
	if token == "" {
		return nil, NewTokenNotFoundError(ap.GetTokenSource())
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, NewNetworkError(fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "Specify-CLI/1.0.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, NewNetworkError(fmt.Errorf("failed to get user info: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 401:
			return nil, NewTokenInvalidError("token is invalid or expired")
		case 403:
			return nil, NewPermissionDeniedError("user information")
		default:
			return nil, NewAPIError(resp.StatusCode, fmt.Sprintf("failed to get user info with status %d", resp.StatusCode))
		}
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, NewAPIError(0, fmt.Sprintf("failed to parse user info: %v", err))
	}

	return userInfo, nil
}

// IsTokenExpired 检查令牌是否过期
func (ap *AuthProvider) IsTokenExpired() (bool, error) {
	token := ap.GetToken()
	if token == "" {
		return true, NewTokenNotFoundError(ap.GetTokenSource())
	}

	// 通过调用API来检查令牌是否有效
	err := ap.validateTokenWithAPI(token)
	if err != nil {
		// 检查是否是令牌过期错误
		if authErr, ok := err.(*AuthError); ok {
			switch authErr.Type {
			case ErrorTypeTokenInvalid:
				return true, nil // 令牌无效，可能已过期
			case ErrorTypeTokenExpired:
				return true, nil // 明确的过期错误
			default:
				return false, err // 其他错误，不是过期问题
			}
		}
		return false, err
	}

	return false, nil // 令牌有效，未过期
}

// GetTokenSource 获取当前令牌来源的字符串描述
func (ap *AuthProvider) GetTokenSource() string {
	info := ap.GetTokenInfo()
	switch info.Source {
	case TokenSourceCLI:
		return "CLI argument"
	case TokenSourceGHToken:
		return "GH_TOKEN environment variable"
	case TokenSourceGitHubToken:
		return "GITHUB_TOKEN environment variable"
	case TokenSourceFile:
		return "manually set token"
	default:
		return "none"
	}
}