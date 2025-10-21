package infrastructure

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"specify-cli/internal/types"
)

// CreateNetworkError 创建网络错误
func CreateNetworkError(err error, url string, status int) *types.NetworkError {
	if err == nil {
		return nil
	}
	
	errorType := classifyError(err, status)
	message := buildErrorMessage(errorType, err, url, status)
	
	return &types.NetworkError{
		Type:    errorType,
		Message: message,
		Cause:   err,
		URL:     url,
		Status:  status,
	}
}

// classifyError 分类错误类型
func classifyError(err error, status int) types.NetworkErrorType {
	// 检查HTTP状态码
	if status > 0 {
		switch {
		case status == 401 || status == 403:
			return types.NetworkErrorTypeAuthentication
		case status == 404:
			return types.NetworkErrorTypeNotFound
		case status >= 500:
			return types.NetworkErrorTypeServerError
		}
	}
	
	// 检查错误类型
	switch {
	case isTimeoutError(err):
		return types.NetworkErrorTypeTimeout
	case isConnectionError(err):
		return types.NetworkErrorTypeConnection
	case isSSLError(err):
		return types.NetworkErrorTypeSSL
	case isProxyError(err):
		return types.NetworkErrorTypeProxy
	default:
		return types.NetworkErrorTypeUnknown
	}
}

// isTimeoutError 检查是否为超时错误
func isTimeoutError(err error) bool {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	
	if err == context.DeadlineExceeded {
		return true
	}
	
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") || 
		   strings.Contains(errStr, "deadline exceeded")
}

// isConnectionError 检查是否为连接错误
func isConnectionError(err error) bool {
	if _, ok := err.(*net.OpError); ok {
		return true
	}
	
	if _, ok := err.(*net.DNSError); ok {
		return true
	}
	
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection refused") ||
		   strings.Contains(errStr, "no such host") ||
		   strings.Contains(errStr, "network unreachable") ||
		   strings.Contains(errStr, "connection reset")
}

// isSSLError 检查是否为SSL/TLS错误
func isSSLError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "tls") ||
		   strings.Contains(errStr, "ssl") ||
		   strings.Contains(errStr, "certificate") ||
		   strings.Contains(errStr, "x509")
}

// isProxyError 检查是否为代理错误
func isProxyError(err error) bool {
	if urlErr, ok := err.(*url.Error); ok {
		return strings.Contains(strings.ToLower(urlErr.Error()), "proxy")
	}
	
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "proxy")
}

// buildErrorMessage 构建错误消息
func buildErrorMessage(errorType types.NetworkErrorType, err error, url string, status int) string {
	switch errorType {
	case types.NetworkErrorTypeTimeout:
		return fmt.Sprintf("Request timeout when accessing %s: %v", url, err)
	case types.NetworkErrorTypeConnection:
		return fmt.Sprintf("Connection failed to %s: %v", url, err)
	case types.NetworkErrorTypeAuthentication:
		return fmt.Sprintf("Authentication failed for %s (HTTP %d): %v", url, status, err)
	case types.NetworkErrorTypeNotFound:
		return fmt.Sprintf("Resource not found at %s (HTTP %d)", url, status)
	case types.NetworkErrorTypeServerError:
		return fmt.Sprintf("Server error at %s (HTTP %d): %v", url, status, err)
	case types.NetworkErrorTypeSSL:
		return fmt.Sprintf("SSL/TLS error when accessing %s: %v", url, err)
	case types.NetworkErrorTypeProxy:
		return fmt.Sprintf("Proxy error when accessing %s: %v", url, err)
	default:
		return fmt.Sprintf("Network error when accessing %s: %v", url, err)
	}
}

// HandleHTTPError 处理HTTP错误
func HandleHTTPError(resp *http.Response, err error, url string) error {
	if err != nil {
		return CreateNetworkError(err, url, 0)
	}
	
	if resp == nil {
		return CreateNetworkError(fmt.Errorf("empty response"), url, 0)
	}
	
	if resp.StatusCode >= 400 {
		return CreateNetworkError(
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status),
			url,
			resp.StatusCode,
		)
	}
	
	return nil
}

// ExecuteWithRetry 带重试的执行函数
func ExecuteWithRetry(operation func() error, maxRetries int, waitTime time.Duration, maxWaitTime time.Duration) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 检查是否为可重试的错误
		if netErr, ok := err.(*types.NetworkError); ok {
			if !netErr.IsRetryable() {
				return err
			}
		}
		
		// 如果是最后一次尝试，直接返回错误
		if attempt == maxRetries {
			break
		}
		
		// 计算等待时间（指数退避）
		currentWait := time.Duration(attempt+1) * waitTime
		if currentWait > maxWaitTime {
			currentWait = maxWaitTime
		}
		
		time.Sleep(currentWait)
	}
	
	return lastErr
}

// RetryableHTTPClient 可重试的HTTP客户端包装器
type RetryableHTTPClient struct {
	client       *http.Client
	maxRetries   int
	waitTime     time.Duration
	maxWaitTime  time.Duration
}

// NewRetryableHTTPClient 创建可重试的HTTP客户端
func NewRetryableHTTPClient(client *http.Client, maxRetries int, waitTime, maxWaitTime time.Duration) *RetryableHTTPClient {
	return &RetryableHTTPClient{
		client:      client,
		maxRetries:  maxRetries,
		waitTime:    waitTime,
		maxWaitTime: maxWaitTime,
	}
}

// Do 执行HTTP请求（带重试）
func (rhc *RetryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error
	
	err := ExecuteWithRetry(func() error {
		resp, err := rhc.client.Do(req)
		lastResp = resp
		lastErr = err
		
		return HandleHTTPError(resp, err, req.URL.String())
	}, rhc.maxRetries, rhc.waitTime, rhc.maxWaitTime)
	
	if err != nil {
		return lastResp, err
	}
	
	return lastResp, lastErr
}

// Get 执行GET请求（带重试）
func (rhc *RetryableHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, CreateNetworkError(err, url, 0)
	}
	
	return rhc.Do(req)
}