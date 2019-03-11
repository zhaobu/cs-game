package tpl

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/db/mgo"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

func (t *RoundTpl) MakeDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
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

	req, ok := pb.(*pbgame.MakeDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.MakeDeskReq")
		t.Log.Error(err.Error())
		return
	}

	rsp := &pbgame.MakeDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("recv %s %+v ", args.Name, *req)

	var newDeskID uint64
	newDeskID, err = cache.AllocDeskID()
	if err != nil {
		t.Log.Error(err.Error())
		// TODO send
		return
	}

	defer func() {
		if err != nil {
			cache.FreeDeskID(newDeskID)
		}
	}()

	deskInfo := &pbcommon.DeskInfo{}
	deskInfo.ID = newDeskID
	deskInfo.CreateUserID = args.UserID
	deskInfo.CreateTime = time.Now().UTC().Unix()
	deskInfo.GameName = t.gameName
	deskInfo.GameID = t.gameID
	deskInfo.ClubID = req.ClubID

	err = cache.AddDeskInfo(deskInfo)
	if err != nil {
		t.Log.Error(err.Error())
		// TODO send
		return
	}

	defer func() {
		if err != nil {
			cache.DelDeskInfo(newDeskID)
		}
	}()

	if req.ClubID != 0 {
		_, err = mgo.QueryClubByID(req.ClubID)
		if err == nil {
			cache.AddClubDeskRelation(req.ClubID, newDeskID)
			defer func() {
				if err != nil {
					cache.DeleteClubDeskRelation(newDeskID)
				}
			}()
		}
	}

	succ, err := cache.EnterGame(args.UserID, t.gameName, t.gameID, newDeskID, false)
	if err != nil {
		t.Log.Error(err.Error())
		// TODO send
		return
	}

	if !succ {
		// TODO send
		return fmt.Errorf("entergame failed %d", args.UserID)
	}

	defer func() {
		if err != nil {
			cache.ExitGame(args.UserID, t.gameName, t.gameID, newDeskID)
		}
	}()

	rsp.Info = deskInfo

	err = t.plugin.HandleMakeDeskReq(args.UserID, req, newDeskID)

	return err
}
