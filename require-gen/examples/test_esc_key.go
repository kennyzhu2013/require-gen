package main

import (
	"fmt"
	"specify-cli/internal/ui"
)

func main() {
	fmt.Println("=== ESC键检测测试 ===")
	fmt.Println("请按ESC键测试检测功能")
	fmt.Println("按其他键查看检测结果")
	fmt.Println()

	// 初始化键盘
	err := ui.InitKeyboard()
	if err != nil {
		fmt.Printf("键盘初始化失败: %v\n", err)
		return
	}
	defer ui.CloseKeyboard()

	for i := 0; i < 10; i++ { // 最多测试10次
		fmt.Printf("测试 %d/10 - 请按键: ", i+1)
		
		key, err := ui.GetKey()
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}

		fmt.Printf("检测到: '%s'", key)
		
		if key == "Escape" {
			fmt.Println(" ← ESC键检测成功！")
			fmt.Println("✅ ESC键检测正常工作")
			break
		} else {
			fmt.Printf(" (不是ESC键)\n")
		}
	}
	
	fmt.Println("\n测试完成")
}