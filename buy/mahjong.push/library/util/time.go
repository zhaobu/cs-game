package util

import (
	"fmt"
	"time"
)

// 获取当前时间戳
func GetTime() int64 {
	return time.Now().Unix()
}

// 获取当前格式化时间
func GetTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 将时间戳格式化
func FormatUnixTime(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
}

// 获取微秒时间
func GetMicrotime() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
