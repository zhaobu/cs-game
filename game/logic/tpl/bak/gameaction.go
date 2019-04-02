package tpl

import (
	"context"
	"cy/game/codec"
	pbgame "cy/game/pb/game"
	"fmt"
	"runtime/debug"
)

func (t *RoundTpl) GameAction(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			t.Log.Warnf("recover:uid=%d,stack=%s", args.UserID, string(debug.Stack()))
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

	t.plugin.HandleGameAction(args.UserID, req)

	return
}
