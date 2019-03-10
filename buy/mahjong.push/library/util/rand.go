package util

import (
	"math/rand"
	"time"
)

// RandIntn 获取一个 0 ~ n 之间的随机值
func RandIntn(n int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return r.Intn(n)
}
