package main

import (
	"encoding/json"

	"mahjong-process-task.go/ierror"
)

func buildError(err *ierror.Error) []byte {
	var data = map[string]interface{}{
		"code":    err.GetCode(),
		"message": err.Error(),
	}
	byts, _ := json.Marshal(data)
	return byts
}

func buildSuccess(msg string) []byte {
	if msg == "" {
		msg = "操作成功"
	}
	var data = map[string]interface{}{
		"code":    0,
		"message": msg,
	}
	byts, _ := json.Marshal(data)
	return byts
}
