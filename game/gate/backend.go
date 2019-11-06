package main

import (
	"encoding/json"
	"game/codec"
	"game/util"
	"time"

	"go.uber.org/zap"
)

// var msgSync sync.WaitGroup //保证按顺序发送订阅的消息

func XreadBackend(addr string, db int) {
	go func() {
		for {
			err := util.RedisXread(addr, db, "backend_to_gate", onGameMsg)
			if err != nil {
				log.Errorf("RedisXread err:%s", err.Error())
			}
			time.Sleep(time.Second * 10)
		}
	}()
}

func onGameMsg(channel string, msgData []byte) error {
	go func() {
		var xx struct {
			Msg  *codec.Message
			Uids []uint64
		}
		err := json.Unmarshal(msgData, &xx)
		if err != nil {
			log.Errorf("%s channel recv msg err:%s", channel, err.Error())
			return
		}

		if xx.Msg == nil {
			return
		}

		tlog.Warn("recv backend", zap.String("channel", channel), zap.String("name", xx.Msg.Name), zap.Any("to", xx.Uids))
		for _, uid := range xx.Uids {
			if sess, ok := mgr.GetSession(uid); ok {
				sess.sendMsg(xx.Msg)
			}
		}
	}()
	return nil
}

// func subscribeBackend(addr string, db int) {
// 	go func() {
// 		for {
// 			err := util.Subscribe(addr, db, "backend_to_gate", onGameMsg)
// 			if err != nil {
// 				log.Errorf("subscribe %s", err.Error())
// 			}
// 			time.Sleep(time.Second * 10)
// 		}
// 	}()
// }

// func onGameMsg(channel string, data []byte) error {
// 	msgSync.Add(1)
// 	go func() { //每次订阅的通道有消息时都会启动新协程处理
// 		var xx struct {
// 			Msg  *codec.Message
// 			Uids []uint64
// 		}

// 		err := json.Unmarshal(data, &xx)
// 		if err != nil {
// 			log.Errorf("%s channel recv msg err:%s", channel, err.Error())
// 			return
// 		}

// 		if xx.Msg == nil {
// 			return
// 		}

// 		tlog.Warn("recv backend", zap.String("channel", channel), zap.String("name", xx.Msg.Name), zap.Any("to", xx.Uids))
// 		for _, uid := range xx.Uids {
// 			if sess, ok := mgr.GetSession(uid); ok {
// 				sess.sendMsg(xx.Msg)
// 			}
// 		}
// 		msgSync.Done()
// 	}()
// 	msgSync.Wait()
// 	return nil
// }
