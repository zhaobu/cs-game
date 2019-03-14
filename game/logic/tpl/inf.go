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

const (
	// 结算类型
	FeeTypeGold    = 1 // 金币
	FeeTypeMasonry = 2 // 砖石
	// 桌子类型
	DeskTypeMatch  = 1 // 匹配
	DeskTypeFriend = 2 // 好友、俱乐部
	DeskTypeLadder = 3 // 比赛
)

type (
	GameLogicPlugin interface {
		HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp)
		HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp)
		HandleGameAction(uid uint64, req *pbgame.GameAction)
		HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq, rsp *pbgame.JoinDeskRsp)
		HandleMakeDeskReq(uid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp)
		HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp)
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

func (t *RoundTpl) InitRedis(redisAddr string, redisDb int) {
	t.redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisAddr)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", redisDb); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}

func (t *RoundTpl) SetPlugin(p GameLogicPlugin) {
	t.plugin = p
}

func (t *RoundTpl) toGateNormal(pb proto.Message, uids ...uint64) error {
	t.Log.Infof("tpl send %v %s %+v", uids, proto.MessageName(pb), pb)

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
		t.Log.Error(err.Error())
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
