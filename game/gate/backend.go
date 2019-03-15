package main

import (
	"cy/game/codec"
	"cy/game/util"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func subscribeBackend(addr string, db int) {
	go func() {
		err := util.Subscribe(addr, db, "backend_to_gate", onGameMsg)
		if err != nil {
			logrus.Errorf("subscribe %s", err.Error())
		}
	}()
}

func onGameMsg(channel string, data []byte) error {
	go func() {
		var xx struct {
			Msg  *codec.Message
			Uids []uint64
		}

		err := json.Unmarshal(data, &xx)
		if err != nil {
			logrus.Errorf("recv game msg %s", err.Error())
			return
		}

		if xx.Msg == nil {
			return
		}

		logrus.WithFields(logrus.Fields{"name": xx.Msg.Name, "to": xx.Uids}).Info("recv backend")

		for _, uid := range xx.Uids {
			if sess, ok := mgr.GetSession(uid); ok {
				sess.sendMsg(xx.Msg)
			}
		}
	}()
	return nil
}
