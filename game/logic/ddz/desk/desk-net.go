package desk

import (
	"game/codec"
	"game/codec/protobuf"
	"game/db/mgo"
	pbcommon "game/pb/common"
	pbgame "game/pb/game"
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

var (
	gameName  string
	gameID    string
	redisPool *redis.Pool
	log       *logrus.Entry
)

func Init(redisAddr string, redisDb int, mgoURI, name, id string, log_ *logrus.Entry) error {
	log = log_

	if err := mgo.Init(mgoURI); err != nil {
		return err
	}

	gameName = name
	gameID = id

	redisPool = &redis.Pool{
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

	return nil
}

func QueryUserInfo(uid uint64) (info *pbcommon.UserInfo, err error) {
	return mgo.QueryUserInfo(uid)
}

func toGateNormal(loge *logrus.Entry, pb proto.Message, uids ...uint64) error {
	loge.Infof("send %v %s %+v", uids, proto.MessageName(pb), pb)

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

	rc := redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	if err != nil {
		loge.Error(err.Error())
	}
	return err
}

func (d *desk) toGate(pb proto.Message, uids ...uint64) error {
	d.loge.Infof("send %v %s %+v", uids, proto.MessageName(pb), pb)

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

	rc := redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	return err
}

func (d *desk) toOne(pb proto.Message, uid uint64) {
	d.toGate(pb, uid)
}

func (d *desk) toOther(pb proto.Message, notUID uint64) {
	uids := make([]uint64, 0)
	for i := 0; i < seatNumber; i++ {
		if d.sdPlayers[i] != nil && d.sdPlayers[i].uid != notUID {
			uids = append(uids, d.sdPlayers[i].uid)
		}
	}
	d.toGate(pb, uids...)
}

func (d *desk) toSiteDown(pb proto.Message) {
	uids := d.getSdUids()
	d.toGate(pb, uids...)
}
