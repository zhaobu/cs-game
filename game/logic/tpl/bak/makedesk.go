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
)

func (t *RoundTpl) MakeDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			t.Log.Warnf("recover:uid=%d,stack=%s", args.UserID, string(debug.Stack()))
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

	var newDeskID uint64

	defer func() {
		t.ToGateNormal(rsp, args.UserID)

		if rsp.Code == pbgame.MakeDeskRspCode_MakeDeskSucc && req.ClubID != 0 {
			t.SendDeskChangeNotif(req.ClubID, newDeskID, 1)
		}
	}()

	t.Log.Infof("tpl recv:uid=%d,args.Name=%s,reg=%+v", args.UserID, args.Name, *req)
	// 1> 分配桌子ID
	newDeskID, err = cache.AllocDeskID()
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

	// 2> 俱乐部和桌子的关系
	var clubInfo *mgo.Club
	if req.ClubID != 0 {
		clubInfo, err = mgo.QueryClubByID(req.ClubID)
		if err != nil {
			rsp.Code = pbgame.MakeDeskRspCode_MakeDeskCanNotFindClubID
			return nil
		}

		_ = clubInfo

		cache.AddClubDeskRelation(req.ClubID, newDeskID)

		// // 用默认建房参数
		// if !clubInfo.IsAutoCreate && !clubInfo.IsCustomGameArg {
		// 	for idx, a := range clubInfo.GameArgs {
		// 		if a.Enable && a.GameName == req.GameName {
		// 			req.GameArgMsgName = a.GameArgMsgName
		// 			req.GameArgMsgValue = a.GameArgMsgValue
		// 			t.Log.Infof("club:%d use default arg %s %d",
		// 				req.ClubID, req.GameArgMsgName, idx)
		// 			break
		// 		}
		// 	}
		// }

		defer func() {
			if rsp.Code != pbgame.MakeDeskRspCode_MakeDeskSucc {
				cache.DeleteClubDeskRelation(newDeskID)
			}
		}()
	}

	deskInfo := &pbcommon.DeskInfo{}
	deskInfo.ID = newDeskID
	deskInfo.CreateUserID = args.UserID
	if ui, err := mgo.QueryUserInfo(args.UserID); err == nil {
		deskInfo.CreateUserName = ui.Name
		deskInfo.CreateUserProfile = ui.Profile
	}
	deskInfo.CreateTime = time.Now().UTC().Unix()
	// deskInfo.CreateFee =
	deskInfo.ArgName = req.GameArgMsgName
	deskInfo.ArgValue = req.GameArgMsgValue
	deskInfo.Status = "1"
	deskInfo.GameName = t.gameName
	deskInfo.GameID = t.gameID
	deskInfo.ClubID = req.ClubID
	deskInfo.Kind = DeskTypeFriend
	// deskInfo.SdInfos
	deskInfo.TotalLoop = 0
	deskInfo.CurrLoop = 0

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

	// 4> 进入游戏
	// succ, err := cache.EnterGame(args.UserID, t.gameName, t.gameID, newDeskID, false)
	// if err != nil {
	// 	t.Log.Error(err.Error())
	// 	rsp.Code = pbgame.MakeDeskRspCode_MakeDeskUserStatusErr
	// 	return nil
	// }

	// if !succ {
	// 	rsp.Code = pbgame.MakeDeskRspCode_MakeDeskUserStatusErr
	// 	return nil
	// }

	// defer func() {
	// 	if rsp.Code != pbgame.MakeDeskRspCode_MakeDeskSucc {
	// 		cache.ExitGame(args.UserID, t.gameName, t.gameID, newDeskID)
	// 	}
	// }()

	rsp.Info = deskInfo

	t.plugin.HandleMakeDeskReq(args.UserID, newDeskID, req, rsp)

	return nil
}
