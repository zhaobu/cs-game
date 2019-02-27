package desk

import (
	"cy/game/cache"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	"cy/game/pb/common"
	"cy/game/pb/game"
	"cy/game/pb/game/ddz"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func MatchStart(arg *pbgame_ddz.RoomArg, uids []uint64) bool {
	d := newDesk(arg, 0)

	d.matchSitDwon(uids...)
	return true
}

func MakeDesk(uid uint64, makeDeskReq *pbgame.MakeDeskReq) {
	makeDeskRsp := &pbgame.MakeDeskRsp{}
	if makeDeskReq.Head != nil {
		makeDeskRsp.Head = &pbcommon.RspHead{Seq: makeDeskReq.Head.Seq}
	}

	var newDeskID uint64

	defer func() {
		toGateNormal(makeDeskRsp, uid)

		if makeDeskRsp.Code == pbgame.MakeDeskRspCode_MakeDeskSucc {
			// 发送桌子信息
			d := getDeskByID(makeDeskRsp.Info.ID)
			if d != nil {
				d.toSiteDown(d.deskInfo(0))
			}
		} else {
			// 预扣费用

			// 回滚
			if newDeskID != 0 {
				cache.DeleteClubDeskRelation(newDeskID)
				cache.DelDeskInfo(newDeskID)
				cache.FreeDeskID(newDeskID)

				cache.ExitGame(uid, gameName, gameID, newDeskID)
			}
		}
	}()

	// 1 检查参数
	arg, err := checkMakeDeskArg(makeDeskReq)
	if err != nil {
		makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskArgsErr
		makeDeskRsp.StrCode = err.Error()
		return
	}

	// 2 扣除费用 TODO
	change := int64(0)
	if arg.PaymentType == 1 {
		change = int64(arg.Fee)
	} else if arg.PaymentType == 2 {
		change = int64(arg.Fee / seatNumber)
	}
	_, err = mgo.UpdateWealthPre(uid, arg.FeeType, change)
	if err != nil {
		makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughMoney
		return
	}

	// 3 分配deskid
	newDeskID, err = cache.AllocDeskID()
	if err != nil {
		makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughDesk
		makeDeskRsp.StrCode = err.Error()
		return
	}

	// 4 cache中新加桌子
	makeDeskRsp.Info, err = cacheAddDeskInfo(newDeskID, uid, makeDeskReq.ClubID)
	if err != nil {
		makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskInternalServerError
		makeDeskRsp.StrCode = err.Error()
		return
	}

	// 建桌子的人 默认要进入游戏
	succ, err := cache.EnterGame(uid, gameName, gameID, newDeskID, false)
	if err != nil {
		makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskInternalServerError
		makeDeskRsp.StrCode = err.Error()
		return
	}

	if !succ {
		makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskUserStatusErr
		makeDeskRsp.StrCode = fmt.Sprintf("failed enter game")
		return
	}

	arg.Type = 2
	arg.DeskID = newDeskID
	arg.RoomId = 0

	d := newDesk(arg, uid)

	if d.actionJoinDesk(uid) != pbgame.JoinDeskRspCode_JoinDeskSucc {
		makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskInternalServerError
		return
	}

	makeDeskRsp.Code = pbgame.MakeDeskRspCode_MakeDeskSucc
	makeDeskRsp.GameArgMsgName, makeDeskRsp.GameArgMsgValue, _ = protobuf.Marshal(d.actionQueryDeskArg())

	//
	if makeDeskReq.ClubID != 0 {
		club, err := mgo.QueryClubByID(makeDeskReq.ClubID)
		if err == nil {
			_ = club
			cache.AddClubDeskRelation(makeDeskReq.ClubID, newDeskID)
		}
	}

	return
}

func cacheAddDeskInfo(id, uid uint64, clubID int64) (*pbcommon.DeskInfo, error) {
	deskInfo := &pbcommon.DeskInfo{}
	deskInfo.ID = id
	deskInfo.CreateUserID = uid
	deskInfo.CreateTime = time.Now().UTC().Unix()
	deskInfo.GameName = gameName
	deskInfo.GameID = gameID
	deskInfo.ClubID = clubID
	return deskInfo, cache.AddDeskInfo(deskInfo)
}

func JoinDesk(uid uint64, joinDeskReq *pbgame.JoinDeskReq) {
	joinDeskRsp := &pbgame.JoinDeskRsp{}
	if joinDeskReq.Head != nil {
		joinDeskRsp.Head = &pbcommon.RspHead{Seq: joinDeskReq.Head.Seq}
	}

	deskID := joinDeskReq.DeskID

	d := getDeskByID(deskID)

	defer func() {
		toGateNormal(joinDeskRsp, uid)

		if joinDeskRsp.Code == pbgame.JoinDeskRspCode_JoinDeskSucc {
			// 发送桌子信息
			if d != nil {
				d.toSiteDown(d.deskInfo(0))
			}
		}
	}()

	// 桌子映射存在
	if d == nil {
		joinDeskRsp.Code = pbgame.JoinDeskRspCode_JoinDeskInternalServerError
		return
	}

	deskInfo, err := cache.QueryDeskInfo(deskID)
	if err != nil {
		joinDeskRsp.Code = pbgame.JoinDeskRspCode_JoinDeskInternalServerError
		return
	}

	// 加入游戏
	succ, err := cache.EnterGame(uid, gameName, gameID, deskID, false)
	if err != nil {
		joinDeskRsp.Code = pbgame.JoinDeskRspCode_JoinDeskInternalServerError
		joinDeskRsp.ErrMsg = err.Error()
		return
	}

	if !succ {
		joinDeskRsp.Code = pbgame.JoinDeskRspCode_JoinDeskUserStatusErr
		return
	}

	joinDeskRsp.Code = d.actionJoinDesk(uid)
	if joinDeskRsp.Code != pbgame.JoinDeskRspCode_JoinDeskSucc {
		cache.ExitGame(uid, gameName, gameID, deskID)
		return
	}

	joinDeskRsp.Info = deskInfo
	joinDeskRsp.GameArgMsgName, joinDeskRsp.GameArgMsgValue, _ = protobuf.Marshal(d.actionQueryDeskArg())
}

func DestroyDesk(uid uint64, req *pbgame.DestroyDeskReq) {
	d := getDeskByID(req.DeskID)
	if d == nil {
		return
	}

	d.actionDestroy(uid, req)
	return
}

func (d *desk) actionDestroy(uid uint64, req *pbgame.DestroyDeskReq) {
	rsp := &pbgame.DestroyDeskRsp{Code: 1}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	var uids []uint64
	defer func() {
		toGateNormal(rsp, uid)
		if rsp.Code == 1 {
			toGateNormal(&pbgame.DestroyDeskNotif{DeskID: d.id}, uids...)
			d.mu.Lock()
			d.preSureCreater()
			d.enterEnd()
			d.mu.Unlock()
		}
	}()

	d.mu.Lock()
	if d.createUserID != uid {
		rsp.Code = 3
	} else if d.isStarted {
		rsp.Code = 2
	}
	uids = d.getSdUids()
	d.mu.Unlock()
	return
}

// needReady 匹配场的需要坐下时就准备好
func (d *desk) markUserSitDown(uid uint64, needReady bool) pbgame.JoinDeskRspCode {
	if !d.f.Is("SWait") {
		return pbgame.JoinDeskRspCode_JoinDeskGameStatusErr
	}

	if d.isSitDownUser(uid) {
		return pbgame.JoinDeskRspCode_JoinDeskSucc
	}

	dir := len(d.sdPlayers)
	if dir == seatNumber {
		return pbgame.JoinDeskRspCode_JoinDeskDeskFull
	}

	info, err := mgo.QueryUserInfo(uid)
	if err != nil {
		logrus.Errorf("queryUserInfo[%d] %s", uid, err.Error())
		return pbgame.JoinDeskRspCode_JoinDeskInternalServerError
	}

	pi := &playerInfo{
		uid:    uid,
		status: pbgame_ddz.UserGameStatus_UGSSitDown,
		info:   info,
		dir:    dir,
	}

	if needReady {
		pi.status = pbgame_ddz.UserGameStatus_UGSReady
	}

	d.sdPlayers[dir] = pi
	updateUser2desk(d, uid)

	return pbgame.JoinDeskRspCode_JoinDeskSucc
}

func QueryDeskInfo(uid uint64, req *pbgame.QueryDeskInfoReq) {
	rsp := &pbgame.QueryDeskInfoRsp{Code: 2}

	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	d := getDeskByID(req.DeskID)

	defer func() {
		toGateNormal(rsp, uid)
	}()

	if d == nil {
		return
	}

	var err error
	rsp.Info, err = cache.QueryDeskInfo(req.DeskID)
	if err != nil {
		return
	}

	rsp.Code = 1
	rsp.GameArgMsgName, rsp.GameArgMsgValue, _ = protobuf.Marshal(d.actionQueryDeskArg())

	return
}

func ExitDesk(uid uint64, req *pbgame.ExitDeskReq) {
	d := getDeskByUID(uid)
	if d == nil {
		return
	}
	d.actionExitDesk(uid, req)
}

func UserLogin(uid uint64) {
	d := getDeskByUID(uid)
	if d != nil {
		d.actionLogin(uid)
	}
}

func (d *desk) actionQueryDeskArg() *pbgame_ddz.RoomArg {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.arg
}

func (d *desk) actionJoinDesk(uid uint64) pbgame.JoinDeskRspCode {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.markUserSitDown(uid, false)
}

func (d *desk) actionExitDesk(uid uint64, req *pbgame.ExitDeskReq) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isSitDownUser(uid) {
		return
	}

	rsp := &pbgame.ExitDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		toGateNormal(rsp, uid)

		if rsp.Code == 1 {
			// 通知其他玩家 有人离开
			d.toSiteDown(&pbgame_ddz.UserGameStatusBroadcast{
				UserID: uid,
				Status: pbgame_ddz.UserGameStatus_UGSFree,
			})

			d.userExit(uid)
		}
	}()

	if d.isStarted {
		rsp.Code = 2
		return
	}
	rsp.Code = 1
	return
}

func (d *desk) actionLogin(uid uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 登陆后 要取消托管
	d.changeTrustee(uid, false)

	// 登陆后 要发desk当前信息
	info := d.deskInfo(uid)
	d.toOne(info, uid)
}

// 玩家动作请求
func Action(uid uint64, actionName string, actionValue []byte) {
	d := getDeskByUID(uid)
	if d != nil {
		d.action(uid, actionName, actionValue)
	} else {
		logrus.Warnf("can not find desk by uid %d", uid)
	}
}

func (d *desk) action(uid uint64, actionName string, actionValue []byte) {
	pb, err := protobuf.Unmarshal(actionName, actionValue)
	if err != nil {
		logrus.Warnf("deskid %d invalid action %s", d.id, actionName)
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	switch v := pb.(type) {
	case *pbgame_ddz.UserCall:
		d.handleUserCall(uid, v)
	case *pbgame_ddz.UserRob:
		d.handleUserRob(uid, v)
	case *pbgame_ddz.UserDouble:
		d.handleUserDouble(uid, v)
	case *pbgame_ddz.UserOper:
		d.handleUserOper(uid, v)
	case *pbgame_ddz.MultipleReq:
		d.handleMultipleReq(uid, v)
	case *pbgame_ddz.UserTrustee:
		d.handleUserTrustee(uid, v)
	case *pbgame_ddz.UserReadyReq:
		d.handleUserReadyReq(uid, v)
	case *pbgame_ddz.QueryWarRecord:
		d.handleQueryWarRecord(uid, v)
	case *pbgame_ddz.UserProposeBreakGame:
		d.handleUserProposeBreakGame(uid, v)
	case *pbgame_ddz.UserBreakGameVote:
		d.handleUserBreakGameVote(uid, v)
	case *pbgame_ddz.ChatReq:
		d.handleChatReq(uid, v)
	default:
		logrus.Warnf("invalid type %s", actionName)
	}
}
