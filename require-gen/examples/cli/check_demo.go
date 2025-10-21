package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"specify-cli/internal/cli"
)

func checkMain() {
	fmt.Println("=== Go版本 Check 命令测试演示 ===")
	fmt.Println()

	// 设置测试环境
	setupTestEnvironment()

	// 测试基本check命令
	fmt.Println("1. 测试基本check命令:")
	testBasicCheck()

	// 测试带版本信息的check命令
	fmt.Println("\n2. 测试带版本信息的check命令:")
	testCheckWithVersions()

	// 测试带详细信息的check命令
	fmt.Println("\n3. 测试带详细信息的check命令:")
	testCheckWithDetails()

	// 测试帮助信息
	fmt.Println("\n4. 测试check命令帮助信息:")
	testCheckHelp()

	fmt.Println("\n=== Check 命令测试完成 ===")
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() {
	// 设置工作目录到项目根目录
	projectRoot := filepath.Join("..", "..")
	if err := os.Chdir(projectRoot); err != nil {
		log.Printf("Warning: Could not change to project root: %v", err)
	}

	// 设置环境变量
	os.Setenv("SPECIFY_DEBUG", "false")
	os.Setenv("SPECIFY_VERBOSE", "false")
}

// testBasicCheck 测试基本check命令
func testBasicCheck() {
	fmt.Println("执行: specify check")

	// 模拟命令行参数
	oldArgs := os.Args
	os.Args = []string{"specify", "check"}

	defer func() {
		os.Args = oldArgs
		if r := recover(); r != nil {
			fmt.Printf("命令执行完成 (recovered: %v)\n", r)
		}
	}()

	// 执行check命令
	if err := cli.Execute(); err != nil {
		fmt.Printf("命令执行错误: %v\n", err)
	} else {
		fmt.Println("命令执行成功")
	}
}

// testCheckWithVersions 测试带版本信息的check命令
func testCheckWithVersions() {
	fmt.Println("执行: specify check --versions")

	// 模拟命令行参数
	oldArgs := os.Args
	os.Args = []string{"specify", "check", "--versions"}

	defer func() {
		os.Args = oldArgs
		if r := recover(); r != nil {
			fmt.Printf("命令执行完成 (recovered: %v)\n", r)
		}
	}()

	// 执行check命令
	if err := cli.Execute(); err != nil {
		fmt.Printf("命令执行错误: %v\n", err)
	} else {
		fmt.Println("命令执行成功")
	}
}

// testCheckWithDetails 测试带详细信息的check命令
func testCheckWithDetails() {
	fmt.Println("执行: specify check --details")

	// 模拟命令行参数
	oldArgs := os.Args
	os.Args = []string{"specify", "check", "--details"}

	defer func() {
		os.Args = oldArgs
		if r := recover(); r != nil {
			fmt.Printf("命令执行完成 (recovered: %v)\n", r)
		}
	}()

	// 执行check命令
	if err := cli.Execute(); err != nil {
		fmt.Printf("命令执行错误: %v\n", err)
	} else {
		fmt.Println("命令执行成功")
	}
}

// testCheckHelp 测试check命令帮助信息
func testCheckHelp() {
	fmt.Println("执行: specify check --help")

	// 模拟命令行参数
	oldArgs := os.Args
	os.Args = []string{"specify", "check", "--help"}

	defer func() {
		os.Args = oldArgs
		if r := recover(); r != nil {
			fmt.Printf("命令执行完成 (recovered: %v)\n", r)
		}
	}()

	// 执行check命令
	if err := cli.Execute(); err != nil {
		fmt.Printf("命令执行错误: %v\n", err)
	} else {
		fmt.Println("命令执行成功")
	}
}
