package ierror

import (
	"fmt"
	"mahjong-connection/config"
	"mahjong-connection/core"

	"github.com/juju/errors"
)

// Error 错误
type Error struct {
	code  int
	msg   string // 带参数的具体错误描述
	error        // 返还给客户端的错误
}

// NewError 新建一个错误对象
// 若错误号未配置，返回一个通用描述
func NewError(code int, argv ...interface{}) *Error {
	err := &Error{code: code}
	if msg, exists := config.Errors[code]; exists {
		err.msg = fmt.Sprintf(msg[1], argv...)
		err.error = fmt.Errorf(msg[0])
	} else {
		err.msg = fmt.Sprintf("错误号[%v]未定义", code)
		err.error = fmt.Errorf("请求失败")
	}
	return err
}

// GetCode 获取错误号
func (err *Error) GetCode() int {
	return err.code
}

// GetMsg 获取详细的错误描述
func (err *Error) GetMsg() string {
	return err.msg
}

// MustNil 判断error是否是nil
func MustNil(err error) bool {
	if err != nil {
		core.Logger.Error(errors.ErrorStack(err))
		return false
	}
	return true
}

// IMustNil 判断ierror是否为nil
func IMustNil(err *Error) bool {
	if err != nil {
		core.Logger.Error(errors.ErrorStack(err.error))
		return false
	}
	return true
}
