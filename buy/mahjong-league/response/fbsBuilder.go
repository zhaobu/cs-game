package response

import (
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/ierror"

	flatbuffers "github.com/google/flatbuffers/go"
)

// BuidGameResult 构建一个fbs coommonResult
func BuidGameResult(builder *flatbuffers.Builder, err *ierror.Error) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	code := 0
	msg := ""
	if err != nil {
		code = err.GetCode()
		msg = err.Error()
	}
	errmsg := builder.CreateString(msg)
	fbsCommon.GameResultStart(builder)
	fbsCommon.GameResultAddCode(builder, int32(code))
	fbsCommon.GameResultAddMsg(builder, errmsg)
	commonResult := fbsCommon.GameResultEnd(builder)
	return builder, commonResult
}
