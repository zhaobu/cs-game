package main

import (
	. "cy/chat/def"
	"reflect"
)

type typeAndHandle struct {
	T reflect.Type
	F Handle
}

var (
	jsonType = make(map[OpKind]*typeAndHandle)
)

func init() {
	jsonType[OpQueryGroupReq] = &typeAndHandle{T: reflect.TypeOf(QueryGroupReq{}), F: handleQueryGroupReq}
	jsonType[OpJoinGroupReq] = &typeAndHandle{T: reflect.TypeOf(JoinGroupReq{}), F: handleJoinGroupReq}
	jsonType[OpExitGroupReq] = &typeAndHandle{T: reflect.TypeOf(ExitGroupReq{}), F: handleExitGroupReq}
	jsonType[OpSendGroupMsgReq] = &typeAndHandle{T: reflect.TypeOf(SendGroupMsgReq{}), F: handleSendGroupMsgReq}
}
