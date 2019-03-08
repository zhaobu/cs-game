package util

import (
	"runtime"

	"fmt"
)

// RecoverPanic 捕获panic
func RecoverPanic() {
	// 捕获异常
	if err := recover(); err != nil {
		stack := make([]byte, 1024)
		stack = stack[:runtime.Stack(stack, true)]

		timestamp := GetTimestamp()
		fmt.Println("[", timestamp, "]", "RecoverPanic:", err)
		fmt.Println("[", timestamp, "]", "stack:", string(stack))
	}
}
