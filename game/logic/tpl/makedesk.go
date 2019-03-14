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
		return
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

	defer func() {
		t.toGateNormal(rsp, args.UserID)
	}()

	t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("tpl recv %s %+v ", args.Name, *req)

	var newDeskID uint64
	newDeskID, err = cache.AllocDeskID() // 1> 分配桌子ID
	if err != nil {
		t.Log.Error(err.Error())
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughDesk
		return nil
	}

	defer func() {
		if rsp.Code != pbgame.MakeDeskRspCode_MakeDeskSucc {
			cache.FreeDeskID(newDeskID)
		}
	}()

	deskInfo := &pbcommon.DeskInfo{}
	deskInfo.ID = newDeskID
	deskInfo.CreateUserID = args.UserID
	//deskInfo.CreateUserName = ""
	deskInfo.CreateTime = time.Now().UTC().Unix()
	deskInfo.ArgName = req.GameArgMsgName
	deskInfo.ArgValue = req.GameArgMsgValue
	deskInfo.GameName = t.gameName
	deskInfo.GameID = t.gameID
	deskInfo.ClubID = req.ClubID
	deskInfo.Kind = DeskTypeFriend

	err = cache.AddDeskInfo(deskInfo) // 2> 保存桌子信息
	if err != nil {
		t.Log.Error(err.Error())
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskInternalServerError
		return nil
	}

	defer func() {
		if rsp.Code != pbgame.MakeDeskRspCode_MakeDeskSucc {
			cache.DelDeskInfo(newDeskID)
		}
	}()

	// 3> 俱乐部和桌子的关系
	if req.ClubID != 0 {
		_, err = mgo.QueryClubByID(req.ClubID)
		if err != nil {
			rsp.Code = pbgame.MakeDeskRspCode_MakeDeskCanNotFindClubID
			return
		}
		// TODO 数量限制
		cache.AddClubDeskRelation(req.ClubID, newDeskID)
		defer func() {
			if rsp.Code != pbgame.MakeDeskRspCode_MakeDeskSucc {
				cache.DeleteClubDeskRelation(newDeskID)
			}
		}()
	}

	// 4> 进入游戏
	succ, err := cache.EnterGame(args.UserID, t.gameName, t.gameID, newDeskID, false)
	if err != nil {
		t.Log.Error(err.Error())
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskUserStatusErr
		return nil
	}

	if !succ {
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskUserStatusErr
		return nil
	}

	defer func() {
		if rsp.Code != pbgame.MakeDeskRspCode_MakeDeskSucc {
			cache.ExitGame(args.UserID, t.gameName, t.gameID, newDeskID)
		}
	}()

	rsp.Info = deskInfo

	t.plugin.HandleMakeDeskReq(args.UserID, newDeskID, req, rsp)

	return nil
}
