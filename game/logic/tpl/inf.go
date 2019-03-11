package tpl

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/pb/game"
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type (
	DestroyDeskReqPlugin interface {
		HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq) error
	}

	ExitDeskReqPlugin interface {
		HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq) error
	}

	GameActionPlugin interface {
		HandleGameAction(uid uint64, req *pbgame.GameAction) error
	}

	JoinDeskReqPlugin interface {
		HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq) error
	}

	MakeDeskReqPlugin interface {
		HandleMakeDeskReq(uid uint64, req *pbgame.MakeDeskReq, deskID uint64) error
	}

	QueryGameConfigReqPlugin interface {
		HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq) error
	}
)

type RoundTpl struct {
	Log      *logrus.Entry
	plugins  []interface{}
	gameName string
	gameID   string

	redisPool *redis.Pool
}

func (t *RoundTpl) SetName(gameName, gameID string) {
	t.gameName = gameName
	t.gameID = gameID
}

func (t *RoundTpl) Add(p interface{}) {
	t.plugins = append(t.plugins, p)
}

func (t *RoundTpl) toGateNormal(loge *logrus.Entry, pb proto.Message, uids ...uint64) error {
	t.Log.Infof("send %v %s %+v", uids, proto.MessageName(pb), pb)

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
	t.Log.Infof("send %v %s %+v", uids, proto.MessageName(pb), pb)

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
