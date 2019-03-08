package ierror

import (
	"fmt"
	"mahjong-league/core"

	"github.com/juju/errors"
)

// Error 错误
type Error struct {
	code int
	error
}

// NewError 新建一个错误对象
// 若错误号未配置，返回一个通用描述
func NewError(code int, argv ...interface{}) *Error {
	err := &Error{code: code}
	if msg, exists := errorConfig[code]; exists {
		err.error = fmt.Errorf(msg, argv...)
	} else {
		err.error = fmt.Errorf("错误号[%d]未定义", code)
	}
	return err
}

// GetCode 获取错误号
func (err *Error) GetCode() int {
	return err.code
}

// MustNil 判断err是否是nil
func MustNil(err error) bool {
	if err != nil {
		core.Logger.Error(errors.ErrorStack(err))
		return false
	}
	return true
}
