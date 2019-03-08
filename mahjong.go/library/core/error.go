package core

import (
	"fmt"

	"mahjong.go/config"
)

type Error struct {
	code    int
	message string
}

// 新建一个错误对象
func NewError(code int, argv ...interface{}) *Error {
	var error *Error

	msg, exists := config.ErrorConfig[code]
	if !exists {
		error = &Error{-1, fmt.Sprintf("错误号[%d]未定义。", code)}
	} else {
		error = &Error{code, fmt.Sprintf(msg, argv...)}
	}

	return error
}

// 获取错误号
func (this *Error) GetCode() int {
	return this.code
}

// 输出错误内容
func (this *Error) Error() string {
	return this.message
}
