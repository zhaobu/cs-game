package clientgame

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbgame_csmj "cy/game/pb/game/mj/changshu"
	"fmt"

	"go.uber.org/zap"
)

type Changshu struct {
	waitchan chan int
}

var (
	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

type deskInfo struct {
	DeskId uint64 `json:"DeskId"`
}

func (self *Changshu) DetailMsg(msg *codec.Message) {
	// zap.Info("recv", zap.String("msgName", msg.Name))
	pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch v := pb.(type) {
	case *pbgame_csmj.S2CThrowDice:
		tlog.Info("recv", zap.String("msgName", msg.Name), zap.Any("msgValue", v))

		// case *pbgame.MakeDeskRsp:
		// 	tlog.Info("recv", zap.String("msgName", msg.Name), zap.Any("msgValue", v))
		// 	desk := &deskInfo{DeskId: v.Info.ID}
		// 	//写入到文件中
		// 	buf, err := json.MarshalIndent(desk, "", "	") //格式化编码
		// 	if err != nil {
		// 		fmt.Println("err = ", err)
		// 		return
		// 	}
		// 	writebuf(*fileName, string(buf))
		// 	self.waitchan <- 1
	}
}

func (self *Changshu) MakeDeskReq() (string, []byte) {
	name, value, _ := protobuf.Marshal(&pbgame_csmj.CreateArg{
		Rule:        []*pbgame_csmj.CyU32String{},
		Barhead:     5,
		PlayerCount: 4,
		Dipiao:      1,
		RInfo:       &pbgame_csmj.RoundInfo{},
		PaymentType: 3,
		LimitIP:     1,
		Voice:       0,
	})
	return name, value
}
