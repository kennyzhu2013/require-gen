package infrastructure

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"specify-cli/internal/types"
)

// GitOperations Git操作实现
type GitOperations struct{}

// NewGitOperations 创建新的Git操作实例
func NewGitOperations() types.GitOperations {
	return &GitOperations{}
}

// IsRepo 检查指定路径是否为Git仓库
func (g *GitOperations) IsRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}
	
	// 检查是否在Git工作树中
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

// InitRepo 初始化Git仓库
func (g *GitOperations) InitRepo(path string, quiet bool) (bool, error) {
	// 检查是否已经是Git仓库
	if g.IsRepo(path) {
		return false, nil
	}

	// 构建git init命令
	args := []string{"init"}
	if quiet {
		args = append(args, "--quiet")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git init failed: %w, output: %s", err, string(output))
	}

	return true, nil
}

// AddAndCommit 添加文件并提交
func (g *GitOperations) AddAndCommit(path string, message string) error {
	// 检查是否为Git仓库
	if !g.IsRepo(path) {
		return fmt.Errorf("not a git repository: %s", path)
	}

	// 添加所有文件
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = path
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w, output: %s", err, string(output))
	}

	// 检查是否有文件需要提交
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = path
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	// 如果没有变更，跳过提交
	if len(strings.TrimSpace(string(statusOutput))) == 0 {
		return nil
	}

	// 提交变更
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = path
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GetStatus 获取Git状态
func (g *GitOperations) GetStatus(path string) (string, error) {
	if !g.IsRepo(path) {
		return "", fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	return string(output), nil
}

// GetBranch 获取当前分支
func (g *GitOperations) GetBranch(path string) (string, error) {
	if !g.IsRepo(path) {
		return "", fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CreateBranch 创建新分支
func (g *GitOperations) CreateBranch(path, branchName string) error {
	if !g.IsRepo(path) {
		return fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b failed: %w, output: %s", err, string(output))
	}

	return nil
}

// SwitchBranch 切换分支
func (g *GitOperations) SwitchBranch(path, branchName string) error {
	if !g.IsRepo(path) {
		return fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w, output: %s", err, string(output))
	}

	return nil
}

// AddRemote 添加远程仓库
func (g *GitOperations) AddRemote(path, name, url string) error {
	if !g.IsRepo(path) {
		return fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "remote", "add", name, url)
	cmd.Dir = path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git remote add failed: %w, output: %s", err, string(output))
	}

	return nil
}

// Push 推送到远程仓库
func (g *GitOperations) Push(path, remote, branch string) error {
	if !g.IsRepo(path) {
		return fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "push", remote, branch)
	cmd.Dir = path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w, output: %s", err, string(output))
	}

	return nil
}

// Pull 从远程仓库拉取
func (g *GitOperations) Pull(path, remote, branch string) error {
	if !g.IsRepo(path) {
		return fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "pull", remote, branch)
	cmd.Dir = path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w, output: %s", err, string(output))
	}

	return nil
}

// Clone 克隆仓库
func (g *GitOperations) Clone(url, targetPath string) error {
	cmd := exec.Command("git", "clone", url, targetPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GetCommitHash 获取当前提交哈希
func (g *GitOperations) GetCommitHash(path string) (string, error) {
	if !g.IsRepo(path) {
		return "", fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL 获取远程仓库URL
func (g *GitOperations) GetRemoteURL(path, remote string) (string, error) {
	if !g.IsRepo(path) {
		return "", fmt.Errorf("not a git repository: %s", path)
	}

	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsClean 检查工作目录是否干净
func (g *GitOperations) IsClean(path string) (bool, error) {
	status, err := g.GetStatus(path)
	if err != nil {
		return false, err
	}

	return len(strings.TrimSpace(status)) == 0, nil
}

// HasUncommittedChanges 检查是否有未提交的变更
func (g *GitOperations) HasUncommittedChanges(path string) (bool, error) {
	clean, err := g.IsClean(path)
	return !clean, err
}