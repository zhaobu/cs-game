package tpl

import (
	"context"
	"cy/game/codec"
	"cy/game/pb/game"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (t *RoundTpl) GameAction(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Warnf("r:%v stack:%s", r, string(debug.Stack()))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		t.Log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.GameAction)
	if !ok {
		err = fmt.Errorf("not *pbgame.GameAction")
		t.Log.Error(err.Error())
		return
	}

	err = t.plugin.HandleGameAction(args.UserID, req)

	return
}
