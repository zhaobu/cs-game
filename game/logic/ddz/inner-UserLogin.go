package main

import (
	"context"
	"cy/game/codec"
	"cy/game/logic/ddz/desk"
	"cy/game/pb/inner"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

// gate通知玩家登陆
func (p *ddz) UserLogin(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	userLogin, ok := pb.(*pbinner.UserLogin)
	if !ok {
		err = fmt.Errorf("not *pbinner.UserLogin")
		logrus.Error(err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{"req": *userLogin, "name": args.Name}).Info("recv")

	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"req":   *userLogin,
				"r":     r,
				"stack": string(debug.Stack()),
				"name":  args.Name,
			}).Info("recv")
		}
	}()

	desk.UserLogin(userLogin.UserID)

	return nil
}
