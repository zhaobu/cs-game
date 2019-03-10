package response

import (
	"mahjong.club/ierror"
	"github.com/fwhappy/mahjong/protocal"

	simplejson "github.com/bitly/go-simplejson"
)

// JSON 构建一个json类的返回消息
func JSON(packageID uint8, js *simplejson.Json) *protocal.ImPacket {
	message, _ := js.Encode()
	return protocal.NewImPacket(packageID, message)
}

// GenFbs 构建一个flatbuffers格式的返回消息
func GenFbs(packageID uint8, mID, mType uint16, mNumber uint32, body []byte) *protocal.ImPacket {
	message := protocal.NewImMessage(mID, mType, mNumber, body)
	return protocal.NewImPacket(packageID, message)
}

// GenEmpty 构建一个空的返回消息
func GenEmpty(packageID uint8) *protocal.ImPacket {
	return protocal.NewImPacket(packageID, []byte{})
}

// JSONError 返回一个json的Error
func JSONError(packageID uint8, err *ierror.Error) *protocal.ImPacket {
	js := simplejson.New()
	js.Set("code", err.GetCode())
	js.Set("message", err.Error())
	return JSON(packageID, js)
}

// JSONSuccess 返回一个json的回应
func JSONSuccess(packageID uint8, js *simplejson.Json) *protocal.ImPacket {
	if js == nil {
		js = simplejson.New()
	}
	js.Set("code", 0)
	js.Set("message", "")
	return JSON(packageID, js)
}
