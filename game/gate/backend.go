package main

import (
	"cy/game/codec"
	"cy/game/util"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

func subscribeBackend(addr string, db int) {
	go func() {
		for {
			err := util.Subscribe(addr, db, "backend_to_gate", onGameMsg)
			if err != nil {
				log.Errorf("subscribe %s", err.Error())
			}
			time.Sleep(time.Second * 10)
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
			log.Errorf("recv game msg %s", err.Error())
			return
		}

		if xx.Msg == nil {
			return
		}

		tlog.Warn("recv backend", zap.String("name", xx.Msg.Name), zap.Any("to", xx.Uids))
		for _, uid := range xx.Uids {
			if sess, ok := mgr.GetSession(uid); ok {
				sess.sendMsg(xx.Msg)
			}
		}
	}()
	return nil
}
