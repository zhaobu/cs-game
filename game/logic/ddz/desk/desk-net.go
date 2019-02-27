package desk

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	"cy/game/pb/common"
	"cy/game/pb/game"
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

var (
	gameName  string
	gameID    string
	redisPool *redis.Pool
)

func Init(redisAddr string, redisDb int, mgoURI, name, id string) error {
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

func toGateNormal(pb proto.Message, uids ...uint64) error {
	logrus.WithFields(logrus.Fields{"pb": pb, "to": uids, "name": proto.MessageName(pb)}).Info("send")

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
		logrus.Error(err.Error())
	}
	return err
}

func toGate(pb proto.Message, uids ...uint64) error {
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
	if err != nil {
		logrus.Error(err.Error())
	}
	return err
}

func (d *desk) toOne(pb proto.Message, uid uint64) {
	logrus.WithFields(logrus.Fields{"deskid": d.id, "pb": pb, "to": uid, "name": proto.MessageName(pb)}).Info("send")
	toGate(pb, uid)
}

func (d *desk) toOther(pb proto.Message, notUID uint64) {
	uids := make([]uint64, 0)
	for i := 0; i < seatNumber; i++ {
		if d.sdPlayers[i] != nil && d.sdPlayers[i].uid != notUID {
			uids = append(uids, d.sdPlayers[i].uid)
		}
	}
	logrus.WithFields(logrus.Fields{"deskid": d.id, "pb": pb, "to": uids, "name": proto.MessageName(pb)}).Info("send")
	toGate(pb, uids...)
}

func (d *desk) toSiteDown(pb proto.Message) {
	uids := d.getSdUids()
	logrus.WithFields(logrus.Fields{"deskid": d.id, "pb": pb, "to": uids, "name": proto.MessageName(pb)}).Info("send")
	toGate(pb, uids...)
}
