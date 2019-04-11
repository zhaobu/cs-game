package main

import (
	"cy/game/cache"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	mj "cy/game/logic/changshu/majiang"
	"cy/game/logic/tpl"
	"cy/game/util"
	"fmt"
	"time"

	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	pbhall "cy/game/pb/hall"
	"sync"

	"github.com/RussellLuo/timingwheel"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

const dissInterval float64 = 2.00000 //解散间隔

type deskUserInfo struct {
	chairId    int32                 //座位号
	info       *pbcommon.UserInfo    //个人信息
	userStatus pbgame.UserDeskStatus //桌子中状态
	lastDiss   time.Time             //上次申请解散时间
	online     bool                  //是否在线
}

type Desk struct {
	mu          sync.Mutex
	gameNode    *tpl.RoomServie
	clubId      int64                    //俱乐部id
	masterUid   uint64                   //房主uid
	deskId      uint64                   //桌子id
	curInning   uint32                   //第几局
	gameStatus  pbgame_logic.GameStatus  //游戏状态
	gameSink    *GameSink                //游戏逻辑
	deskPlayers map[uint64]*deskUserInfo //本桌玩家信息,玩家uid到deskPlayers
	// lookonPlayers map[uint64]*deskUserInfo //观察玩家信息
	playChair   map[int32]*deskUserInfo //玩家chairid到deskPlayers,座位号从0开始
	deskConfig  *pbgame_logic.DeskArg   //桌子参数
	timerManger map[mj.EmtimerID]*timingwheel.Timer
}

func makeDesk(deskArg *pbgame_logic.DeskArg, masterUid, deskID uint64, clubID int64) *Desk {
	d := &Desk{deskId: deskID, clubId: clubID, masterUid: masterUid, deskConfig: deskArg}
	d.gameSink = &GameSink{}
	d.gameSink.Ctor(deskArg.Args)
	d.gameSink.desk = d
	d.playChair = make(map[int32]*deskUserInfo)
	d.deskPlayers = make(map[uint64]*deskUserInfo)
	d.timerManger = make(map[mj.EmtimerID]*timingwheel.Timer)
	return d
}

func (d *Desk) doEnter(uid uint64) pbgame.JoinDeskRspCode {
	// 判断条件 是否能加入

	// 重复加入处理
	return pbgame.JoinDeskRspCode_JoinDeskSucc
}

//找到空闲座位号
// func (d *Desk) getFreeChair() (int32, bool) {
// 	// for i := int32(0); i < d.deskConfig.Args.PlayerCount; i++ {
// 	// 	if _, ok := d.playChair[i]; ok {
// 	// 		continue
// 	// 	}
// 	// 	return i, true
// 	// }
// 	// return -1, false
// }

func (d *Desk) checkPos(pos int32) bool {
	if pos < 0 || pos >= d.deskConfig.Args.PlayerCount {
		return false
	}
	return d.playChair[pos] == nil
}

func (d *Desk) checkStart() bool {
	if len(d.playChair) < int(d.deskConfig.Args.PlayerCount) {
		return false
	}
	//检查是否所有玩家都准备好
	for _, user_info := range d.playChair {
		if user_info.userStatus != pbgame.UserDeskStatus_UDSSitDown {
			return false
		}
	}
	return true
}

//玩家加入桌子后变成观察者
func (d *Desk) doJoin(uid uint64, rsp *pbgame.JoinDeskRsp) {
	tlog.Info("玩家doJoin加入房间", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))

	var err error
	dUserInfo := &deskUserInfo{userStatus: pbgame.UserDeskStatus_UDSLook, chairId: -1}
	userInfo, err := mgo.QueryUserInfo(uid)
	if err != nil {
		tlog.Info("玩家doJoin加入房间时,mgo.QueryUserInfo err", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskUserStatusErr
		return
	}
	dUserInfo.info = userInfo
	d.deskPlayers[uid] = dUserInfo
	rsp.Code = pbgame.JoinDeskRspCode_JoinDeskSucc
	d.SendData(uid, rsp) //先发送加入成功
	d.sendDeskInfo(uid)
}

//坐下后由观察者变为游戏玩家
func (d *Desk) doSitDown(uid uint64, chair int32, rsp *pbgame.SitDownRsp) {
	tlog.Info("玩家doSitDown坐下准备", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
	//检查是否已经准备好
	dUserInfo := d.deskPlayers[uid]
	//检查是否重复坐下
	if dUserInfo.userStatus != pbgame.UserDeskStatus_UDSLook || dUserInfo.chairId != -1 {
		tlog.Info("玩家doSitDown坐下准备时chairId != -1", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
		rsp.Code = pbgame.SitDownRspCode_SitDownGameStatusErr
		return
	}

	//判断座位是否为空
	if !d.checkPos(chair) {
		tlog.Info("玩家doSitDown坐下准备时座位已经有人", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
		rsp.Code = pbgame.SitDownRspCode_SitDownNotEmpty
		return
	}
	//检查钱是否够
	fee := calcFee(d.deskConfig.Args)
	if fee != 0 {
		_, err := mgo.UpdateWealthPre(uid, pbgame.FeeType_FTMasonry, fee)
		if err != nil {
			tlog.Error("err mgo.UpdateWealthPre()", zap.Error(err))
			rsp.Code = pbgame.SitDownRspCode_SitDownNotEnoughMoney
			return
		}
	}
	dUserInfo.chairId = chair
	dUserInfo.userStatus = pbgame.UserDeskStatus_UDSSitDown
	d.playChair[chair] = dUserInfo
	d.gameSink.AddPlayer(chair, uid, dUserInfo.info.GetName())
	//先发送加入成功消息
	rsp.Code = pbgame.SitDownRspCode_SitDownSucc
	d.SendData(uid, rsp)
	//发送玩家信息给所有人
	d.sendDeskInfo(0)
	//再判断游戏开始
	if d.checkStart() {
		d.gameSink.StartGame()
	}
	d.updateDeskInfo(2) //通知俱乐部更新桌子信息
}

//起立后由玩家变为观察者
func (d *Desk) doStandUp(chairId int32) pbgame.ExitDeskRspCode {
	tlog.Info("玩家doStandUp起立", zap.Int32("chairId", chairId), zap.Uint64("uid", d.GetUidByChairid(chairId)), zap.Uint64("deksId", d.deskId))
	// chairId := d.GetChairidByUid(uid)
	// if chairId == -1 {
	// 	tlog.Info("doStandUp时 非游戏玩家")
	// 	return
	// }
	if d.curInning > 0 {
		tlog.Info("doStandUp时 游戏已经开始")
		return pbgame.ExitDeskRspCode_ExitDeskPlaying
	}
	dUserInfo := d.playChair[chairId]
	if dUserInfo.userStatus == pbgame.UserDeskStatus_UDSPlaying {
		tlog.Info("doStandUp时 userStatus=pbgame.UserDeskStatus_UDSPlaying")
		return pbgame.ExitDeskRspCode_ExitDeskPlaying
	}
	if !d.gameSink.Exitlayer(chairId) {
		tlog.Info("doStandUp时 d.gameSink.Exitlayer 出错")
		return pbgame.ExitDeskRspCode_ExitDeskPlaying
	}

	dUserInfo.chairId = -1
	dUserInfo.userStatus = pbgame.UserDeskStatus_UDSLook
	delete(d.playChair, chairId)
	//发送玩家信息给所有人
	d.sendDeskInfo(0)
	d.updateDeskInfo(2) //通知俱乐部更新桌子信息
	return pbgame.ExitDeskRspCode_ExitDeskSucc
}

//玩家退出桌子
func (d *Desk) doExit(uid uint64, rsp *pbgame.ExitDeskRsp) {
	dUserInfo := d.deskPlayers[uid]
	if dUserInfo == nil {
		tlog.Info("doExit时 d.deskPlayers[uid]=nil")
		rsp.Code = pbgame.ExitDeskRspCode_ExitDeskNotInDesk
		return
	}
	//如果游戏玩家先起立
	chairId := d.GetChairidByUid(uid)
	if chairId != -1 {
		rsp.Code = d.doStandUp(chairId)
		if rsp.Code != pbgame.ExitDeskRspCode_ExitDeskSucc {
			tlog.Info("doExit时 d.doStandUp 失败", zap.Any("rsp.Code", rsp.Code))
			return
		}
		d.rollbackMoney(uid)
	}
	d.deleteDeskPlayer(uid)
}

//删除桌子玩家
func (d *Desk) deleteDeskPlayer(uid uint64) {
	delete(d.deskPlayers, uid)
	deleteUser2desk(uid)
}

//更新桌子信息到redis,并通知俱乐部,changeTyp:1create 2update 3delete
func (d *Desk) updateDeskInfo(changeTyp int32) {
	if d.clubId == 0 {
		return
	}
	deskInfo, err := cache.QueryDeskInfo(d.deskId)
	if err != nil {
		tlog.Error("err cache.QueryDeskInfo", zap.Error(err))
		return
	}
	if d.gameStatus <= pbgame_logic.GameStatus_GSWait {
		deskInfo.Status = "1"
	} else {
		deskInfo.Status = "2"
	}
	deskInfo.CurrLoop = int64(d.curInning)
	deskInfo.SdInfos = []*pbcommon.SiteDownPlayerInfo{}
	for k, v := range d.playChair {
		uinfo := &pbcommon.SiteDownPlayerInfo{UserID: v.info.UserID, Dir: k, Name: v.info.Name, Profile: v.info.Profile}
		deskInfo.SdInfos = append(deskInfo.SdInfos, uinfo)
	}
	cache.UpdateDeskInfo(deskInfo)

	//通知俱乐部查询桌子信息
	d.gameNode.SendDeskChangeNotif(d.clubId, d.deskId, changeTyp)
}

//sendDeskInfo 有新玩家坐下,起立,重连,发送玩家信息
func (d *Desk) sendDeskInfo(uid uint64) {
	if len(d.playChair) <= 0 {
		return
	}
	msg := &pbgame_logic.GameDeskInfo{GameName: gameName, Arg: d.deskConfig, GameStatus: d.gameStatus, CurInning: d.curInning}
	msg.BankerId = d.gameSink.bankerId
	msg.MasterUid = d.masterUid
	msg.GameUser = []*pbgame_logic.DeskUserInfo{}
	for chair, user := range d.playChair {
		userInfo := &pbgame_logic.DeskUserInfo{}
		userInfo.Info = user.info
		userInfo.ChairId = chair
		userInfo.UserStatus = user.userStatus
		msg.GameUser = append(msg.GameUser, userInfo)
	}
	d.gameSink.gameReconnect(msg, uid)
	d.SendData(uid, msg)
}

//解散桌子
func (d *Desk) doDestroyDesk(uid uint64, rsp *pbgame.DestroyDeskRsp) {
	//检查是否过于频繁
	if dUserInfo, ok := d.deskPlayers[uid]; ok {
		if dUserInfo.lastDiss.Unix() != 0 && util.Float64Equal(time.Now().Sub(dUserInfo.lastDiss).Seconds(), dissInterval) {
			rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskFrequent
			rsp.ErrMsg = fmt.Sprintf("解散请求过于频繁")
			return
		}
		dUserInfo.lastDiss = time.Now()
	}
	if d.gameStatus == pbgame_logic.GameStatus_GSWait {
		if uid != d.masterUid {
			rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskNotMaster
			rsp.ErrMsg = fmt.Sprintf("游戏开始前只有房主才能解散桌子")
			return
		}
		d.dealDestroyDesk()
	} else if d.gameStatus >= pbgame_logic.GameStatus_GSPlay {
		d.set_timer(mj.TID_Destory, 15*time.Second, func() {

		})
	}

	rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskSucc
}

//
// func (d *Desk) timeOutDestroyDesk(uid uint64, req *pbgame_ddz.UserBreakGameVote) {
// 	if d.voteStartUserID == 0 {
// 		return
// 	}

// 	for _, v := range d.sdPlayers {
// 		if v.uid == uid && v.agreeBreakGame == 0 {
// 			if req.Agree {
// 				v.agreeBreakGame = 1
// 			} else {
// 				v.agreeBreakGame = 2
// 			}
// 			d.toSiteDown(&pbgame_ddz.BreakGameVoteBroadcast{UserID: uid, Agree: req.Agree})

// 			votedCnt := 0
// 			for _, v := range d.sdPlayers {
// 				if v.agreeBreakGame != 0 {
// 					votedCnt++
// 				}
// 			}
// 			if votedCnt == seatNumber || !req.Agree {
// 				d.breakGameTimer.Stop()
// 				d.breakGameVoteEnd()
// 			}
// 			break
// 		}
// 	}
// }

//处理桌子销毁
func (d *Desk) dealDestroyDesk() {
	//处理房间所有人的状态
	for uid, _ := range d.deskPlayers {
		delete(d.deskPlayers, uid)
		deleteUser2desk(uid)
		//通知所有人房间已销毁
		d.SendData(uid, &pbgame.DestroyDeskNotif{DeskID: d.deskId})
		cache.ExitGame(uid, d.gameNode.GameName, d.gameNode.GameID, d.deskId)
	}
	deleteID2desk(d.deskId)
}

// 预扣还回去
func (d *Desk) rollbackMoney(uid uint64) {
	change := int64(0)
	goldChange := int64(0)
	masonryChange := int64(0)
	createArg := d.deskConfig.Args
	if createArg.PaymentType == 1 { //房主支付
		change = int64(createArg.RInfo.Fee)
	} else if createArg.PaymentType == 2 { //个人支付
		change = int64(createArg.RInfo.Fee / uint32(createArg.PlayerCount))
	}

	if d.deskConfig.FeeType == pbgame.FeeType_FTGold {
		goldChange = change
	} else if d.deskConfig.FeeType == pbgame.FeeType_FTMasonry {
		masonryChange = change
	}

	info, err := mgo.UpdateWealthPreSure(uid, d.deskConfig.FeeType, change)
	if err != nil {
		return
	}

	// 通知变化
	d.SendData(uid, &pbhall.UserWealthChange{
		UserID:        uid,
		Gold:          info.Gold,
		GoldChange:    goldChange,
		Masonry:       info.Masonry,
		MasonryChange: masonryChange,
	})
}

// d.breakGameTimer = time.AfterFunc(time.Second*timeOutBreakGameVote, func() {
// 	d.mu.Lock()
// 	defer d.mu.Unlock()

// 	for _, v := range d.sdPlayers {
// 		if v.agreeBreakGame == 0 { // 默认为同意
// 			v.agreeBreakGame = 1
// 			d.toSiteDown(&pbgame_ddz.BreakGameVoteBroadcast{UserID: v.uid, Agree: true})
// 		}
// 	}

// 	d.breakGameVoteEnd()
// })

// func (d *desk) breakGameVoteEnd() {
// 	canBreakGame := true
// 	for _, v := range d.sdPlayers {
// 		if v.agreeBreakGame == 2 { // 有任意一人反对 就不能解散
// 			canBreakGame = false
// 			break
// 		}
// 	}

// 	voteEnd := &pbgame_ddz.BreakGameVoteEnd{}
// 	if canBreakGame {
// 		voteEnd.Code = 1
// 	} else {
// 		voteEnd.Code = 2
// 	}
// 	d.toSiteDown(voteEnd)

// 	if voteEnd.Code == 1 {
// 		if d.arg.Type == gameTypFriend && d.currLoopCnt == 1 { // 第1局还没打完，就解散了，预扣的要还回去
// 			d.preSureCreater()
// 		}

// 		d.gameOverInfo(2)
// 		d.flashWarRecord()
// 		d.toSiteDown(&d.warRecord)
// 		d.enterEnd()
// 		return
// 	}

// 	// clear
// 	d.voteStartUserID = 0
// 	for _, v := range d.sdPlayers {
// 		v.agreeBreakGame = 0
// 	}
// }

//游戏逻辑分发
func (d *Desk) doAction(uid uint64, actionName string, actionValue []byte) {
	pb, err := protobuf.Unmarshal(actionName, actionValue)
	if err != nil {
		log.Warnf("deskid %d invalid action %s", d.deskId, actionName)
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	chairId := d.GetChairidByUid(uid)
	if -1 == chairId {
		log.Infof("can find chairId by uid%d", uid)
		return
	}
	switch v := pb.(type) {
	case *pbgame_logic.C2SThrowDice:
		d.gameSink.ThrowDice(chairId, v)
	default:
		log.Warnf("invalid type %s", actionName)
	}
}

func (d *Desk) SendData(uid uint64, pb proto.Message) {
	//发给所有人,包括观察者
	if uid == 0 {
		uids := make([]uint64, len(d.deskPlayers))
		var i int = 0
		for uid, _ := range d.deskPlayers {
			uids[i] = uid
			i++
		}
		d.gameNode.ToGateNormal(pb, uids...)
	} else {
		if _, ok := d.deskPlayers[uid]; ok {
			d.gameNode.ToGateNormal(pb, uid)
		}
	}
}

func (d *Desk) SendGameMessage(uid uint64, pb proto.Message) {
	//发给所有人
	if uid == 0 {
		uids := make([]uint64, len(d.deskPlayers))
		var i int = 0
		for uid, _ := range d.deskPlayers {
			uids[i] = uid
			i++
		}
		d.gameNode.ToGate(pb, uids...)
	} else {
		if _, ok := d.deskPlayers[uid]; ok {
			d.gameNode.ToGate(pb, uid)
		}
	}
}

//根据uid查找chair_id
func (d *Desk) GetChairidByUid(uid uint64) int32 {
	if user_info, ok := d.deskPlayers[uid]; ok {
		return user_info.chairId
	}
	return -1
}

//根据chairid查找uid
func (d *Desk) GetUidByChairid(chairId int32) uint64 {
	if user_info, ok := d.playChair[chairId]; ok {
		return user_info.info.UserID
	}
	return 0
}

func (d *Desk) set_timer(tID mj.EmtimerID, dura time.Duration, f func()) {
	exefun := func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		f()
		delete(d.timerManger, tID) //闭包,删除已经执行过的定时器
	}
	d.timerManger[tID] = d.gameNode.Timer.AfterFunc(dura, exefun)
}

func (d *Desk) cancel_timer(tID mj.EmtimerID) {
	if t, ok := d.timerManger[tID]; ok == false {
		log.Infof("取消定时器时定时器不存在")
		return
	} else {
		t.Stop()
		//取消后删除该定时器
		delete(d.timerManger, tID)
	}
}

//玩家上下线
func (d *Desk) OnOffLine(uid uint64, online bool) {
	if online {

	}
}
