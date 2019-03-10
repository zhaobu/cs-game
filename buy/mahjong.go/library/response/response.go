package response

import (
	"github.com/bitly/go-simplejson"
	"mahjong.go/mi/protocal"

	fbsCommon "mahjong.go/fbs/Common"
)

// GenJson 构建一个json类的返回消息
func GenJson(packageId uint8, js *simplejson.Json) *protocal.ImPacket {
	message, _ := js.Encode()

	return protocal.NewImPacket(packageId, message)
}

// GenFbs 构建一个flatbuffers格式的返回消息
func GenFbs(packageId uint8, mId, mType, mIndex, mNumber uint16, body []byte) *protocal.ImPacket {
	message := protocal.NewImMessage(mId, mType, mIndex, mNumber, body)

	return protocal.NewImPacket(packageId, message)
}

// GenEmpty 构建一个空的返回消息
func GenEmpty(packageId uint8) *protocal.ImPacket {
	return protocal.NewImPacket(packageId, []byte{})
}

// GetMessageName 根据messageId获取到messageId的map名字，好辨识
func GetMessageName(messageId uint16) string {
	mId := int(messageId)
	mIdName, exists := fbsCommon.EnumNamesCommand[mId]

	if exists {
		return mIdName
	}
	return "UndefinedCommandName"
}
