package tpl

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbgame "cy/game/pb/game"
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type (
	GameLogicPlugin interface {
		HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq) error
		HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq) error
		HandleGameAction(uid uint64, req *pbgame.GameAction) error
		HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq) error
		HandleMakeDeskReq(uid uint64, req *pbgame.MakeDeskReq, deskID uint64) error
		HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq) error
	}
)

type RoundTpl struct {
	Log      *logrus.Entry
	plugin   GameLogicPlugin
	gameName string
	gameID   string

	redisPool *redis.Pool
}

func (t *RoundTpl) SetName(gameName, gameID string) {
	t.gameName = gameName
	t.gameID = gameID
}

func (t *RoundTpl) SetPlugin(p GameLogicPlugin) {
	t.plugin = p
}

func (t *RoundTpl) toGateNormal(loge *logrus.Entry, pb proto.Message, uids ...uint64) error {
	t.Log.Infof("toGateNormal send %v %s %+v", uids, proto.MessageName(pb), pb)

	msg := &codec.Message{}
	err := codec.Pb2Msg(pb, msg)
	if err != nil {
		return err
	}

	var xx struct {
		Msg  *codec.Message
		Uids []uint64
	}
	xx.Msg = msg
	xx.Uids = append(xx.Uids, uids...)

	data, err := json.Marshal(xx)
	if err != nil {
		return err
	}

	rc := t.redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	if err != nil {
		loge.Error(err.Error())
	}
	return err
}

func (t *RoundTpl) toGate(pb proto.Message, uids ...uint64) error {
	t.Log.Infof("toGate send %v %s %+v", uids, proto.MessageName(pb), pb)

	notif := &pbgame.GameNotif{}
	var err error
	notif.NotifName, notif.NotifValue, err = protobuf.Marshal(pb)
	if err != nil {
		return err
	}

	msg := &codec.Message{}
	err = codec.Pb2Msg(notif, msg)
	if err != nil {
		return err
	}

	var xx struct {
		Msg  *codec.Message
		Uids []uint64
	}
	xx.Msg = msg
	xx.Uids = append(xx.Uids, uids...)

	data, err := json.Marshal(xx)
	if err != nil {
		return err
	}

	rc := t.redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	return err
}
