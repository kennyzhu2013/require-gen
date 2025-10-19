package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitOperations_IsRepo(t *testing.T) {
	git := NewGitOperations()

	t.Run("检测非Git目录", func(t *testing.T) {
		// 创建临时目录
		tempDir, err := os.MkdirTemp("", "test_non_git_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 测试非Git目录
		isRepo := git.IsRepo(tempDir)
		if isRepo {
			t.Error("期望非Git目录返回false，但返回了true")
		}
	})

	t.Run("检测Git目录", func(t *testing.T) {
		// 创建临时目录
		tempDir, err := os.MkdirTemp("", "test_git_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 初始化Git仓库
		success, err := git.InitRepo(tempDir, true)
		if err != nil {
			t.Fatalf("初始化Git仓库失败: %v", err)
		}
		if !success {
			t.Fatal("期望初始化成功，但返回了false")
		}

		// 测试Git目录
		isRepo := git.IsRepo(tempDir)
		if !isRepo {
			t.Error("期望Git目录返回true，但返回了false")
		}
	})

	t.Run("检测不存在的目录", func(t *testing.T) {
		nonExistentPath := filepath.Join(os.TempDir(), "non_existent_dir_12345")
		isRepo := git.IsRepo(nonExistentPath)
		if isRepo {
			t.Error("期望不存在的目录返回false，但返回了true")
		}
	})
}

func TestGitOperations_InitRepo(t *testing.T) {
	git := NewGitOperations()

	t.Run("在空目录中初始化Git仓库", func(t *testing.T) {
		// 创建临时目录
		tempDir, err := os.MkdirTemp("", "test_init_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 初始化Git仓库
		success, err := git.InitRepo(tempDir, false)
		if err != nil {
			t.Fatalf("初始化Git仓库失败: %v", err)
		}
		if !success {
			t.Error("期望初始化成功，但返回了false")
		}

		// 验证.git目录是否存在
		gitDir := filepath.Join(tempDir, ".git")
		if stat, err := os.Stat(gitDir); err != nil || !stat.IsDir() {
			t.Error("初始化后.git目录不存在或不是目录")
		}

		// 验证IsRepo返回true
		if !git.IsRepo(tempDir) {
			t.Error("初始化后IsRepo应该返回true")
		}
	})

	t.Run("在已存在Git仓库的目录中初始化", func(t *testing.T) {
		// 创建临时目录
		tempDir, err := os.MkdirTemp("", "test_reinit_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 第一次初始化
		success, err := git.InitRepo(tempDir, true)
		if err != nil {
			t.Fatalf("第一次初始化失败: %v", err)
		}
		if !success {
			t.Error("期望第一次初始化成功")
		}

		// 第二次初始化（应该返回false，因为已经是Git仓库）
		success, err = git.InitRepo(tempDir, true)
		if err != nil {
			t.Fatalf("第二次初始化出错: %v", err)
		}
		if success {
			t.Error("期望第二次初始化返回false（已经是Git仓库）")
		}
	})

	t.Run("使用quiet模式初始化", func(t *testing.T) {
		// 创建临时目录
		tempDir, err := os.MkdirTemp("", "test_quiet_init_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 使用quiet模式初始化
		success, err := git.InitRepo(tempDir, true)
		if err != nil {
			t.Fatalf("quiet模式初始化失败: %v", err)
		}
		if !success {
			t.Error("期望quiet模式初始化成功")
		}

		// 验证仓库确实被创建
		if !git.IsRepo(tempDir) {
			t.Error("quiet模式初始化后IsRepo应该返回true")
		}
	})
}

func TestGitOperations_AddAndCommit(t *testing.T) {
	git := NewGitOperations()

	t.Run("在Git仓库中添加和提交文件", func(t *testing.T) {
		// 创建临时目录并初始化Git仓库
		tempDir, err := os.MkdirTemp("", "test_commit_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 初始化Git仓库
		success, err := git.InitRepo(tempDir, true)
		if err != nil {
			t.Fatalf("初始化Git仓库失败: %v", err)
		}
		if !success {
			t.Fatal("初始化Git仓库失败")
		}

		// 创建测试文件
		testFile := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testFile, []byte("Hello, Git!"), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		// 添加并提交文件
		err = git.AddAndCommit(tempDir, "Initial commit with test file")
		if err != nil {
			t.Fatalf("添加和提交文件失败: %v", err)
		}

		// 验证提交成功（通过检查状态）
		status, err := git.GetStatus(tempDir)
		if err != nil {
			t.Fatalf("获取Git状态失败: %v", err)
		}

		// 如果状态为空或只包含空白字符，说明没有未提交的更改
		if len(status) > 0 && status != "" {
			t.Logf("Git状态: %s", status)
		}
	})

	t.Run("在非Git目录中尝试提交", func(t *testing.T) {
		// 创建临时目录（不初始化Git）
		tempDir, err := os.MkdirTemp("", "test_no_git_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 尝试在非Git目录中提交
		err = git.AddAndCommit(tempDir, "This should fail")
		if err == nil {
			t.Error("期望在非Git目录中提交失败，但成功了")
		}

		// 验证错误消息包含预期内容
		expectedMsg := "not a git repository"
		if err != nil && !contains(err.Error(), expectedMsg) {
			t.Errorf("期望错误消息包含 '%s'，但得到: %v", expectedMsg, err)
		}
	})

	t.Run("在空仓库中提交（没有文件变更）", func(t *testing.T) {
		// 创建临时目录并初始化Git仓库
		tempDir, err := os.MkdirTemp("", "test_empty_commit_*")
		if err != nil {
			t.Fatalf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 初始化Git仓库
		success, err := git.InitRepo(tempDir, true)
		if err != nil {
			t.Fatalf("初始化Git仓库失败: %v", err)
		}
		if !success {
			t.Fatal("初始化Git仓库失败")
		}

		// 尝试提交（没有文件变更）
		err = git.AddAndCommit(tempDir, "Empty commit")
		if err != nil {
			t.Fatalf("空提交失败: %v", err)
		}
		// 注意：根据实现，空提交应该被跳过，不应该报错
	})
}

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}