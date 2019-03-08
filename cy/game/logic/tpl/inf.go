package tpl

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/pb/game"
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type (
	BeforeMakeDeskReqPlugin interface {
		BeforeMakeDeskReq(*pbgame.MakeDeskReq) error
	}

	AfterMakeDeskReqPlugin interface {
		AfterMakeDeskReq(*pbgame.MakeDeskReq) error
	}

	BeforeGameActionPlugin interface {
		BeforeGameAction(*pbgame.GameAction) error
	}

	AfterGameActionPlugin interface {
		AfterGameAction(*pbgame.GameAction) error
	}
)

type RoundTpl struct {
	Log       *logrus.Entry
	plugins   []interface{}
	redisPool *redis.Pool
}

func (t *RoundTpl) Add(p interface{}) {
	t.plugins = append(t.plugins, p)
}

func (d *RoundTpl) toGate(pb proto.Message, uids ...uint64) error {
	d.Log.Infof("send %v %s %+v", uids, proto.MessageName(pb), pb)

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

	rc := d.redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	return err
}
