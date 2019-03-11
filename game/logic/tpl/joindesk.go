package tpl

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/pb/common"
	"cy/game/pb/game"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (t *RoundTpl) JoinDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		t.Log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.JoinDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.JoinDeskReq")
		t.Log.Error(err.Error())
		return
	}

	rsp := &pbgame.JoinDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("recv %s %+v ", args.Name, *req)

	rsp.Info, err = cache.QueryDeskInfo(req.DeskID)
	if err != nil {
		// TODO send
		return
	}

	succ, err := cache.EnterGame(args.UserID, t.gameName, t.gameID, req.DeskID, false)
	if err != nil {
		t.Log.Error(err.Error())
		// TODO send
		return
	}

	if !succ {
		// TODO send
		return
	}

	defer func() {
		if err != nil {
			cache.ExitGame(args.UserID, t.gameName, t.gameID, req.DeskID)
		}
	}()

	for _, v := range t.plugins {
		if plugin, ok := v.(JoinDeskReqPlugin); ok {
			err = plugin.HandleJoinDeskReq(args.UserID, req)
			if err != nil {
				break
			}
		}
	}

	return err
}
