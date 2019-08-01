package main

import (
	"cy/game/cache"
	"cy/game/codec/protobuf"
	zaplog "cy/game/common/logger"
	"cy/game/db/mgo"
	mj "cy/game/logic/changshu/majiang"
	"cy/game/logic/tpl"
	"cy/game/util"
	"fmt"
	"runtime/debug"
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

var (
	roomlog  *zap.SugaredLogger //printf风格
	troomlog *zap.Logger
)

const dissInterval time.Duration = time.Second * 2  //解散间隔2s
const dissTimeOut time.Duration = time.Second * 120 //投票解散时间120s

type deskUserInfo struct {
	chairId    int32                 //座位号
	info       *pbcommon.UserInfo    //个人信息
	userStatus pbgame.UserDeskStatus //桌子中状态
	lastDiss   time.Time             //上次申请解散时间
	online     bool                  //是否在线
}

//投票解散信息
type voteInfo struct {
	voteUser   uint64                       //发起投票解散的玩家
	voteOption map[uint64]pbgame.VoteOption //玩家投票选择
	voteTime   time.Time                    //玩家发起投票时间点
}
type Desk struct {
	mu            sync.Mutex
	rmu           sync.RWMutex
	gameNode      *tpl.RoomServie
	clubId        int64                    //俱乐部id
	masterUid     uint64                   //房主uid
	clubMasterUid uint64                   //俱乐部群主uid
	deskId        uint64                   //桌子id
	curInning     uint32                   //第几局
	gameStatus    pbgame_logic.GameStatus  //游戏状态
	gameSink      *GameSink                //游戏逻辑
	deskPlayers   map[uint64]*deskUserInfo //本桌玩家信息,玩家uid到deskPlayers
	playChair     map[int32]*deskUserInfo  //玩家chairid到deskPlayers,座位号从0开始
	deskConfig    *pbgame_logic.DeskArg    //桌子参数
	timerManger   map[mj.EmtimerID]*timingwheel.Timer
	*voteInfo
}

func initRoomLog(deskId uint64) {
	var logName, logLevel string
	logName = fmt.Sprintf("./log/roomlog/%s_%d_%s.log", gameName, deskId, time.Now().Format("01021504"))
	if *release {
		logLevel = "info"
	} else {
		logLevel = "debug"
	}
	troomlog = zaplog.InitLogger(logName, logLevel, !*release)
	roomlog = troomlog.Sugar()
}

func makeDesk(deskArg *pbgame_logic.DeskArg, masterUid, clubMasterUid, deskID uint64, clubID int64, gameNode *tpl.RoomServie) *Desk {
	initRoomLog(deskID)
	d := &Desk{deskId: deskID, clubId: clubID, masterUid: masterUid, clubMasterUid: clubMasterUid, deskConfig: deskArg, gameNode: gameNode}
	d.gameSink = &GameSink{desk: d}
	d.gameSink.Ctor(deskArg.Args)
	d.playChair = make(map[int32]*deskUserInfo)
	d.deskPlayers = make(map[uint64]*deskUserInfo)
	d.timerManger = make(map[mj.EmtimerID]*timingwheel.Timer)
	d.set_timer(mj.TID_LongTime, 30*time.Minute, func() {
		if d.curInning == 0 {
			d.dealDestroyDesk(pbgame.DestroyDeskType_DestroyTypeTimeOut)
		}
	})
	return d
}

func (d *Desk) getVoteResult() (voteResult []*pbgame.VoteDestroyDeskInfo) {
	for _, userInfo := range d.playChair {
		if option, ok := d.voteOption[userInfo.info.UserID]; ok {
			voteResult = append(voteResult, &pbgame.VoteDestroyDeskInfo{UserID: userInfo.info.UserID, Option: option})
		} else {
			voteResult = append(voteResult, &pbgame.VoteDestroyDeskInfo{UserID: userInfo.info.UserID, Option: pbgame.VoteOption_VoteOptionNone})
		}
	}
	return
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
	troomlog.Info("玩家doJoin加入房间", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))

	var err error
	dUserInfo := &deskUserInfo{userStatus: pbgame.UserDeskStatus_UDSLook, chairId: -1}
	userInfo, err := mgo.QueryUserInfo(uid)
	if err != nil {
		troomlog.Info("玩家doJoin加入房间时,mgo.QueryUserInfo err", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
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
	troomlog.Info("玩家doSitDown坐下准备", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
	userInfo, err := mgo.QueryUserInfo(uid)
	if err != nil {
		troomlog.Info("玩家doSitDown坐下准备时,mgo.QueryUserInfo err", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
		rsp.Code = pbgame.SitDownRspCode_SitDownNotInDesk
		return
	}
	// ip限制,同一个局域网,不允许加入房间
	if d.deskConfig.Args.LimitIP == 3 || d.deskConfig.Args.LimitIP == 2 {
		for _, v := range d.playChair {
			if userInfo.IP == v.info.IP {
				rsp.Code = pbgame.SitDownRspCode_SitDownIPLimit
				troomlog.Info("玩家doSitDown坐下准备时IP限制,不允许坐下", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
				return
			}
		}
	}
	// ip+距离判断 新加入的和每个已加入的玩家比较 <?的才能加入
	if d.deskConfig.Args.LimitIP == 3 {
		for _, v := range d.playChair {
			if util.DistanceGeo(userInfo.Latitude, userInfo.Longitude, v.info.Latitude, v.info.Longitude) < 500.00 {
				rsp.Code = pbgame.SitDownRspCode_SitDownDistanceSoClose
				troomlog.Info("玩家doSitDown坐下准备时距离限制,不允许坐下", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
				return
			}
		}
	}
	//检查是否已经准备好
	dUserInfo := d.deskPlayers[uid]
	//检查是否重复坐下
	if dUserInfo.userStatus != pbgame.UserDeskStatus_UDSLook || dUserInfo.chairId != -1 {
		troomlog.Info("玩家doSitDown坐下准备时chairId != -1", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
		rsp.Code = pbgame.SitDownRspCode_SitDownGameStatusErr
		return
	}

	//判断座位是否为空
	if !d.checkPos(chair) {
		troomlog.Info("玩家doSitDown坐下准备时座位已经有人", zap.Uint64("uid", uid), zap.Uint64("deksId", d.deskId))
		rsp.Code = pbgame.SitDownRspCode_SitDownNotEmpty
		return
	}
	//检查钱是否够(房主建房时已经预扣)
	if uid != d.masterUid && d.deskConfig.Args.PaymentType == 2 {
		if err := d.updateWealth(uid, 0); err != nil {
			troomlog.Error("err mgo.UpdateWealthPre()", zap.Error(err))
			rsp.Code = pbgame.SitDownRspCode_SitDownNotEnoughMoney
			return
		}
	}
	dUserInfo.chairId = chair
	d.changUserState(uid, pbgame.UserDeskStatus_UDSSitDown)
	if len(d.playChair) == 0 {
		d.gameSink.changGameState(pbgame_logic.GameStatus_GSWait)
	}
	d.playChair[chair] = dUserInfo
	d.gameSink.AddPlayer(chair, uid, dUserInfo.info)
	//先发送加入成功消息
	rsp.Code = pbgame.SitDownRspCode_SitDownSucc
	d.SendData(uid, rsp)
	//发送玩家信息给所有人
	d.sendDeskInfo(0)
	//再判断游戏开始
	if d.checkStart() {
		//游戏开始时,如果是aa支付,房主没有在房间玩游戏,就退还房主预扣的钱
		if d.deskConfig.Args.PaymentType == 2 && d.GetChairidByUid(d.masterUid) == -1 {
			d.updateWealth(d.masterUid, 1)
		}
		d.curInning = 1
		d.changUserState(0, pbgame.UserDeskStatus_UDSPlaying)
		d.gameSink.changGameState(pbgame_logic.GameStatus_GSDice)
		d.gameSink.StartGame()
		d.set_timer(mj.TID_LongTime, 2*time.Hour, func() {
			d.dealDestroyDesk(pbgame.DestroyDeskType_DestroyTypeTimeOut)
		})
	}
	d.updateDeskInfo(2) //通知俱乐部更新桌子信息
}

//改变玩家状态
func (d *Desk) changUserState(uid uint64, uState pbgame.UserDeskStatus) {
	if uid == 0 {
		for _, userInfo := range d.playChair {
			userInfo.userStatus = uState
		}
	} else {
		d.deskPlayers[uid].userStatus = uState
	}
}

//起立后由玩家变为观察者
func (d *Desk) doStandUp(chairId int32) pbgame.ExitDeskRspCode {
	troomlog.Info("玩家doStandUp起立", zap.Int32("chairId", chairId), zap.Uint64("uid", d.GetUidByChairid(chairId)), zap.Uint64("deksId", d.deskId))
	if d.curInning > 0 {
		troomlog.Info("doStandUp时 游戏已经开始")
		return pbgame.ExitDeskRspCode_ExitDeskPlaying
	}
	dUserInfo := d.playChair[chairId]
	if dUserInfo.userStatus == pbgame.UserDeskStatus_UDSPlaying {
		troomlog.Info("doStandUp时 userStatus=pbgame.UserDeskStatus_UDSPlaying")
		return pbgame.ExitDeskRspCode_ExitDeskPlaying
	}
	if !d.gameSink.Exitlayer(chairId) {
		troomlog.Info("doStandUp时 d.gameSink.Exitlayer 出错")
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
		troomlog.Info("doExit时 d.deskPlayers[uid]=nil")
		rsp.Code = pbgame.ExitDeskRspCode_ExitDeskNotInDesk
		return
	}
	//如果游戏玩家先起立
	chairId := d.GetChairidByUid(uid)
	if chairId != -1 {
		rsp.Code = d.doStandUp(chairId)
		if rsp.Code != pbgame.ExitDeskRspCode_ExitDeskSucc {
			troomlog.Info("doExit时 d.doStandUp 失败", zap.Any("rsp.Code", rsp.Code))
			return
		}
		//退还扣的钱(房主预扣的钱要等游戏开始再判断是否需要退还)
		if uid != d.masterUid && d.deskConfig.Args.PaymentType == 2 {
			d.updateWealth(uid, 1)
		}
	}
	delete(d.deskPlayers, uid)
	deleteUser2desk(uid)
	cache.ExitGame(uid, d.gameNode.GameName, d.gameNode.GameID, d.deskId)
	rsp.Code = pbgame.ExitDeskRspCode_ExitDeskSucc
}

//删除桌子玩家
// func (d *Desk) deleteDeskPlayer(uid uint64) {
// 	delete(d.deskPlayers, uid)
// 	deleteUser2desk(uid)
// 	cache.ExitGame(uid, d.gameNode.GameName, d.gameNode.GameID, d.deskId)
// }

//更新桌子信息到redis,并通知俱乐部,changeTyp:1create 2update 3delete
func (d *Desk) updateDeskInfo(changeTyp int32) {
	if d.clubId == 0 {
		return
	}
	deskInfo, err := cache.QueryDeskInfo(d.deskId)
	if err != nil {
		roomlog.Errorf("debug stack info=%s", string(debug.Stack()))
		troomlog.Error("err cache.QueryDeskInfo", zap.Error(err))
		return
	}
	if d.gameStatus <= pbgame_logic.GameStatus_GSWait {
		deskInfo.Status = "1" //等待中
	} else {
		deskInfo.Status = "2" //开始了
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
	if len(d.deskPlayers) <= 0 {
		return
	}
	msg := d.getBaseDeskInfo()
	msg.CurInning = d.curInning
	d.gameSink.gameReconnect(msg, uid)
	d.SendData(uid, msg)
}

func (d *Desk) getBaseDeskInfo() *pbgame_logic.GameDeskInfo {
	msg := &pbgame_logic.GameDeskInfo{GameName: gameName, Arg: d.deskConfig, GameStatus: d.gameStatus, MasterUid: d.masterUid}
	msg.GameUser = make([]*pbgame_logic.DeskUserInfo, 0, d.deskConfig.Args.PlayerCount)
	// 按照座位号从0开始遍历d.playChair
	for chair, user := range d.playChair {
		userInfo := &pbgame_logic.DeskUserInfo{}
		userInfo.Info = user.info
		userInfo.ChairId = chair
		userInfo.UserStatus = user.userStatus
		msg.GameUser = append(msg.GameUser, userInfo)
	}
	return msg
}

//解散桌子
func (d *Desk) doDestroyDesk(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp) {
	if req.Type == pbgame.DestroyDeskType_DestroyTypeGame { //游戏开始房主申请解散
		if uid != d.masterUid {
			rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskNotMaster
			rsp.ErrMsg = fmt.Sprintf("游戏开始前只有房主才能解散桌子")
			return
		}
		d.dealDestroyDesk(req.Type)
	} else if req.Type == pbgame.DestroyDeskType_DestroyTypeVote { //游戏中投票解散
		//检查是否重复申请
		if d.voteInfo != nil {
			rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskRepeated
			rsp.ErrMsg = fmt.Sprintf("已经有玩家申请解散:%d", req.DeskID)
			return
		}
		if d.curInning <= 0 { //游戏中申请解散
			rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskVoteWithOutPlaying
			rsp.ErrMsg = fmt.Sprintf("游戏不在进行中不能进行投票解散:%d", req.DeskID)
			return
		}
		dUserInfo, _ := d.deskPlayers[uid]
		//检查是否过于频繁
		if dUserInfo.lastDiss.Unix() != 0 && time.Since(dUserInfo.lastDiss) < dissInterval {
			rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskFrequent
			rsp.ErrMsg = fmt.Sprintf("解散请求过于频繁")
			return
		}
		dUserInfo.lastDiss = time.Now()
		d.voteInfo = &voteInfo{voteUser: uid, voteTime: time.Now(), voteOption: map[uint64]pbgame.VoteOption{uid: pbgame.VoteOption_VoteOptionAgree}}
		msg := &pbgame.VoteDestroyDeskNotif{DeskID: d.deskId, VoteUser: uid, LeftTime: int32(dissTimeOut.Seconds()), VoteResult: d.getVoteResult()}
		d.SendData(0, msg)
		d.set_timer(mj.TID_Destory, dissTimeOut, func() {
			if d.voteInfo == nil {
				return
			}
			tmpReq := &pbgame.VoteDestroyDeskReq{Option: pbgame.VoteOption_VoteOptionAgree}
			for _, v := range d.playChair {
				//超时默认为同意解散
				if _, ok := d.voteOption[v.info.UserID]; !ok {
					d.doVoteDestroyDesk(v.info.UserID, tmpReq)
				}
			}
		})
	} else if req.Type == pbgame.DestroyDeskType_DestroyTypeClub { //俱乐部群主申请解散房间
		//TODO查询俱乐部接口
		d.dealDestroyDesk(req.Type)
	} else if req.Type == pbgame.DestroyDeskType_DestroyTypeDebug { //强制解散
		d.dealDestroyDesk(req.Type)
	}
	rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskSucc
}

//玩家选择解散请求
func (d *Desk) doVoteDestroyDesk(uid uint64, req *pbgame.VoteDestroyDeskReq) {
	if d.voteOption[uid] != pbgame.VoteOption_VoteOptionNone {
		troomlog.Info("doVoteDestroyDesk already do option", zap.Uint64("uid", uid))
		return
	}
	d.voteOption[uid] = req.Option
	leftTime := dissTimeOut - time.Now().Sub(d.voteInfo.voteTime)
	//广播选择
	d.SendData(0, &pbgame.VoteDestroyDeskNotif{DeskID: d.deskId, VoteUser: d.voteInfo.voteUser, LeftTime: int32(leftTime.Seconds()), VoteResult: d.getVoteResult()})
	if req.Option == pbgame.VoteOption_VoteOptionReject { //拒绝解散
		d.SendData(0, &pbgame.DestroyDeskResultNotif{DeskID: d.deskId, Result: 2, Type: pbgame.DestroyDeskType_DestroyTypeVote})
		d.voteInfo = nil
		d.cancel_timer(mj.TID_Destory)
	} else if len(d.voteOption) == int(d.deskConfig.Args.PlayerCount) { //所有人都同意解散
		d.dealDestroyDesk(pbgame.DestroyDeskType_DestroyTypeVote)
		d.cancel_timer(mj.TID_Destory)
	}
}

//处理桌子销毁
func (d *Desk) dealDestroyDesk(reqType pbgame.DestroyDeskType) {
	switch reqType {
	case pbgame.DestroyDeskType_DestroyTypeClub, pbgame.DestroyDeskType_DestroyTypeVote, pbgame.DestroyDeskType_DestroyTypeTimeOut:
		//在游戏中解散
		if d.curInning > 0 {
			d.gameSink.gameEnd(pbgame_logic.GameEndType_EndDissmiss)
		}
	}
	d.realDestroyDesk(reqType)
}

//游戏结束
func (d *Desk) gameEnd(endType pbgame_logic.GameEndType) {
	if endType != pbgame_logic.GameEndType_EndDissmiss { //胡牌或者流局
		d.changUserState(0, pbgame.UserDeskStatus_UDSGameEnd)
		if d.curInning == d.deskConfig.Args.RInfo.LoopCnt {
			d.realDestroyDesk(pbgame.DestroyDeskType_DestroyTypeGameEnd)
		} else {
			d.curInning++
			d.updateDeskInfo(2) //通知俱乐部更新桌子信息
		}
	}
}

//真正销毁桌子
func (d *Desk) realDestroyDesk(reqType pbgame.DestroyDeskType) {
	//如果房主没有在房间玩游戏,就退还房主预扣的钱
	if d.curInning == 0 {
		if d.deskConfig.Args.PaymentType == 3 { //俱乐部群主支付
			d.updateWealth(d.clubMasterUid, 1)
		} else {
			d.updateWealth(d.masterUid, 1)
		}
	}
	//处理房间所有人的状态
	msg := &pbgame.DestroyDeskResultNotif{DeskID: d.deskId, Result: 1, Type: reqType}

	for uid, _ := range d.deskPlayers {
		d.SendData(uid, msg)
		delete(d.deskPlayers, uid)
		deleteUser2desk(uid)
		//通知所有人房间已销毁
		cache.ExitGame(uid, d.gameNode.GameName, d.gameNode.GameID, d.deskId)
	}
	d.updateDeskInfo(3) //通知俱乐部更新桌子信息
	deleteID2desk(d.deskId)
	cache.DeleteClubDeskRelation(d.deskId)
	cache.DelDeskInfo(d.deskId)
	cache.FreeDeskID(d.deskId)
}

// 更新财富(tag:0:表示预扣,1:返还)
func (d *Desk) updateWealth(uid uint64, tag int) (err error) {
	goldChange := int64(0)
	masonryChange := int64(0)
	change := calcFee(d.deskConfig.Args)

	if d.deskConfig.FeeType == pbgame.FeeType_FTGold {
		goldChange = change
	} else if d.deskConfig.FeeType == pbgame.FeeType_FTMasonry {
		masonryChange = change
	}
	var info *pbcommon.UserInfo
	if tag == 1 {
		info, err = mgo.UpdateWealthPreSure(uid, d.deskConfig.FeeType, change)
	} else {
		info, err = mgo.UpdateWealthPre(uid, d.deskConfig.FeeType, change)
	}
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
	return
}

//游戏逻辑分发
func (d *Desk) doAction(uid uint64, actionName string, actionValue []byte) {
	pb, err := protobuf.Unmarshal(actionName, actionValue)
	if err != nil {
		roomlog.Warnf("deskid %d invalid action %s", d.deskId, actionName)
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	roomlog.Infof("doAction uid: %v,actionName: %s,actionValue: %s", uid, actionName, util.PB2JSON(pb, true))

	chairId := d.GetChairidByUid(uid)
	if -1 == chairId {
		roomlog.Infof("can find chairId by uid%d", uid)
		return
	}
	switch v := pb.(type) {
	case *pbgame_logic.C2SThrowDice:
		d.gameSink.ThrowDice(chairId, v)
	case *pbgame_logic.C2SChiCard:
		d.gameSink.chiCard(chairId, v.Card, v.ChiType)
	case *pbgame_logic.C2SPengCard:
		d.gameSink.pengCard(chairId, v.Card)
	case *pbgame_logic.C2SGangCard:
		d.gameSink.gangCard(chairId, v.Card)
	case *pbgame_logic.C2SHuCard:
		d.gameSink.huCard(chairId)
	case *pbgame_logic.C2SCancelAction:
		d.gameSink.cancelOper(chairId)
	case *pbgame_logic.C2SOutCard:
		d.gameSink.outCard(chairId, v.Card)
	case *pbgame_logic.C2SGetReady:
		d.gameSink.getReady(uid)
	case *pbgame_logic.C2SGetGameRecord:
		d.gameSink.getGameRecord(chairId)
	default:
		roomlog.Warnf("invalid type %s", actionName)
	}
}

//方便打印日志到单个房间
func (d *Desk) ToGate(pb proto.Message, uids ...uint64) {
	roomlog.Infof("ToGate uid: %v,msgName: %s,msgValue: %s", uids, proto.MessageName(pb), util.PB2JSON(pb, true))
	d.gameNode.ToGate(pb, uids...)
}

func (d *Desk) ToGateNormal(pb proto.Message, uids ...uint64) {
	roomlog.Infof("ToGateNormal uid: %v,msgName: %s,msgValue: %s", uids, proto.MessageName(pb), util.PB2JSON(pb, true))
	d.gameNode.ToGateNormal(pb, uids...)
}

//_uid为0时发送给所有人,包括观察者
func (d *Desk) SendData(_uid uint64, pb proto.Message) {
	if _uid == 0 {
		uids := make([]uint64, len(d.deskPlayers))
		var i int = 0
		for uid, _ := range d.deskPlayers {
			uids[i] = uid
			i++
		}
		d.ToGateNormal(pb, uids...)
		return
	}
	d.ToGateNormal(pb, _uid)
}

//_uid为0时发送给所有人,包括观察者
func (d *Desk) SendGameMessage(_uid uint64, pb proto.Message) {
	if _uid == 0 {
		uids := make([]uint64, len(d.deskPlayers))
		var i int = 0
		for uid, _ := range d.deskPlayers {
			uids[i] = uid
			i++
		}
		d.ToGate(pb, uids...)
		return
	}
	d.ToGate(pb, _uid)
}

//发送消息给除所有观察者
func (d *Desk) SendGameMessageAllLook(pb proto.Message) {
	uids := []uint64{}
	for uid, v := range d.deskPlayers {
		if v.userStatus == pbgame.UserDeskStatus_UDSLook {
			uids = append(uids, uid)
		}
	}
	d.ToGate(pb, uids...)
}

//发送消息给除_uid其他所有人
func (d *Desk) SendGameMessageOther(_uid uint64, pb proto.Message) {
	uids := []uint64{}
	for uid, _ := range d.deskPlayers {
		if uid != _uid {
			uids = append(uids, uid)
		}
	}
	d.ToGate(pb, uids...)
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
		if d != nil {
			f()
			d.timerMangerDelete(tID) //闭包,删除已经执行过的定时器
		}
	}
	d.rmu.Lock()
	d.timerManger[tID] = d.gameNode.Timer.AfterFunc(dura, exefun)
	d.rmu.Unlock()
}

func (d *Desk) cancel_timer(tID mj.EmtimerID) {
	d.rmu.Lock()
	t, ok := d.timerManger[tID]
	d.rmu.Unlock()
	if !ok {
		roomlog.Infof("取消定时器时定时器不存在")
		return
	}
	t.Stop()
	//取消后删除该定时器
	d.timerMangerDelete(tID)
}

func (d *Desk) timerMangerDelete(tID mj.EmtimerID) {
	d.rmu.Lock()
	delete(d.timerManger, tID)
	d.rmu.Unlock()
}

//玩家上下线
func (d *Desk) OnOffLine(uid uint64, online bool) {
	dUserInfo := d.deskPlayers[uid]
	dUserInfo.online = online
	if d.curInning == 0 { //游戏开始前
		if !online { //下线
			d.doExit(uid, &pbgame.ExitDeskRsp{})
		}
	} else { //游戏中
		if dUserInfo.userStatus >= pbgame.UserDeskStatus_UDSSitDown {
			if online { //游戏玩家上线
				//重新从mgo获取session信息
				dUserInfo.info, _ = mgo.QueryUserInfo(uid)
				d.sendDeskInfo(uid)
				//如果有解散信息,就广播当前解散信息
				if d.voteInfo != nil {
					leftTime := dissTimeOut - time.Now().Sub(d.voteInfo.voteTime)
					//广播选择
					d.SendData(uid, &pbgame.VoteDestroyDeskNotif{DeskID: d.deskId, VoteUser: d.voteInfo.voteUser, LeftTime: int32(leftTime.Seconds()), VoteResult: d.getVoteResult()})
				}
			}
		} else {
			if !online { //观察者下线
				d.doExit(uid, &pbgame.ExitDeskRsp{})
			}
		}
	}
}
