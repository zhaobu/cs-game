package desk

import (
	"cy/game/logic/ddz/card"
	"cy/game/pb/common"
	"cy/game/pb/game/ddz"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/RussellLuo/timingwheel"
	"github.com/looplab/fsm"
)

const (
	seatNumber = 3

	// 动作超时时间--秒
	timeOutCallrob       = 6   // 叫抢地主
	timeOutDouble        = 5   // 加倍
	timeOutCanOut        = 20  // 要得起
	timeOutCanNotOut     = 5   // 要不起
	timeOutTrustee       = 1   // 托管出牌时间
	timeOutBreakGameVote = 120 // 解散游戏投票时间

	gameTypMatch  = 1 // 匹配
	gameTypFriend = 2 // 好友
	gameTypLadder = 3 // 比赛
)

var (
	tw = timingwheel.NewTimingWheel(time.Second, 60) // 所有桌子公用
)

func init() {
	tw.Start()
}

// 公共倍数信息
type pubMultiple struct {
	callRob uint32
	back    uint32 // 直接倍数表示
	bomb    uint32
	spring  uint32
}

func (m *pubMultiple) clear() {
	m.callRob = 0
	m.back = 1
	m.bomb = 0
	m.spring = 0
}

func (m *pubMultiple) get() uint32 {
	return (1 << (m.callRob + m.bomb + m.spring)) * m.back
}

type playerInfo struct {
	uid            uint64
	status         pbgame_ddz.UserGameStatus
	info           *pbcommon.UserInfo
	dir            int                   // 方位
	currCard       *card.SetCard         // 当前手牌
	historyCard    []*card.SetCard       // 出牌记录
	call           pbgame_ddz.CallCode   // 叫地主标记
	rob            pbgame_ddz.RobCode    // 抢地主标记
	double         pbgame_ddz.DoubleCode // 加倍标记
	doubleMul      uint32                // 加倍倍数 相互影响
	lastOper       pbgame_ddz.OperMask   // 上次出牌操作
	lastCard       *card.SetCard         // 上次出牌
	lastCardType   pbgame_ddz.CardType   // 上次出牌型
	isTrustee      bool                  // 是否托管
	isWin          bool                  // 是否赢了
	change         int64                 // 输赢变化
	agreeBreakGame int                   // 0 init 1 agree 2 against
}

type desk struct {
	mu              *sync.Mutex
	id              uint64 // desk id
	createUserID    uint64
	createTime      time.Time
	arg             *pbgame_ddz.RoomArg
	isStarted       bool   // 是否开局了
	currLoopCnt     uint32 // 当前局数 好友场中用
	warRecord       pbgame_ddz.WarRecord
	voteStartUserID uint64

	sdPlayers            map[int]*playerInfo     // 坐下的玩家 key: 方位 0,1,2
	backCard             *card.SetCard           // 底牌
	backCardCt           pbgame_ddz.BackCardType // 底牌类型
	currPlayer           int                     // 当前操作玩家方位
	currPlayerID         uint64                  // 当前操作玩家userid
	callUID              uint64                  // 叫地主的userid
	robUID               uint64                  // 最后一次抢地主的userid
	landlord             uint64                  // 地主的userid
	lastGiveCardPlayerID uint64                  // 最近出牌的userid
	lastGiveCard         *card.SetCard           // 最近出的牌
	mul                  pubMultiple             // 倍数信息

	reqTime                time.Time // 玩家操作开始时间
	breakGameVoteStartTime time.Time // 解散游戏投票开始时间
	breakGameTimer         *time.Timer
	seq                    uint64             // 操作序号
	timer                  *timingwheel.Timer // 定时器
	f                      *fsm.FSM
	loge                   *logrus.Entry
}

func newDesk(arg *pbgame_ddz.RoomArg, createUserID uint64) *desk {
	d := &desk{}
	d.mu = &sync.Mutex{}
	d.id = arg.DeskID
	d.createUserID = createUserID
	d.createTime = time.Now().UTC()
	d.arg = arg
	d.currLoopCnt = 1
	d.sdPlayers = make(map[int]*playerInfo)
	d.mul.clear()
	d.loge = log.WithFields(logrus.Fields{"deskid": arg.DeskID})

	d.f = fsm.NewFSM(
		"SWait",
		[]fsm.EventDesc{
			{Name: "wait_end", Src: []string{"SWait"}, Dst: "SCall"},
			{Name: "call_end", Src: []string{"SCall"}, Dst: "SRob"},
			{Name: "rob_end", Src: []string{"SRob"}, Dst: "SDouble"},
			{Name: "rouble_end", Src: []string{"SDouble"}, Dst: "SPlay"},
			{Name: "play_end", Src: []string{"SPlay"}, Dst: "SCalc"},
			{Name: "calc_end", Src: []string{"SCalc"}, Dst: "SWait"},
			{Name: "game_end", Src: []string{"SCalc"}, Dst: "SEnd"},
		},
		fsm.Callbacks{},
	)

	updateID2desk(d)

	return d
}

func (d *desk) getSdUids() []uint64 {
	uids := make([]uint64, 0)
	for _, v := range d.sdPlayers {
		uids = append(uids, v.uid)
	}
	return uids
}

func (d *desk) mulUser(uid uint64) uint32 {
	for _, v := range d.sdPlayers {
		if v.uid == uid {
			return d.mul.get() * (1 << v.doubleMul)
		}
	}
	return 1
}

func (d *desk) getDeskInfo() *pbgame_ddz.DeskInfo {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.deskInfo(0)
}

func (d *desk) deskInfo(needuid uint64) *pbgame_ddz.DeskInfo {
	di := &pbgame_ddz.DeskInfo{}
	di.GameName = gameName
	di.Arg = d.arg
	di.Status = pbgame_ddz.GameStatus(pbgame_ddz.GameStatus_value[d.f.Current()])
	di.Landlord = d.landlord
	di.Current = d.currPlayerID
	if d.backCard != nil {
		di.BackCards = d.backCard.Dump()
		di.Bct = d.backCardCt
		di.BackMul = d.mul.back
	}
	di.CurrLoopCnt = d.currLoopCnt
	di.BreakGameStartUserID = d.voteStartUserID
	di.CreateUserID = d.createUserID
	di.CurrSeq = d.seq
	di.BreakGameLeftTime = timeOutBreakGameVote - uint32(time.Now().UTC().Sub(d.breakGameVoteStartTime).Seconds())

	for _, v := range d.sdPlayers {
		dui := &pbgame_ddz.DeskUserInfo{}
		dui.Info = v.info
		dui.Dir = uint32(v.dir)
		dui.Status = v.status

		if d.currPlayerID == v.uid {
			switch d.f.Current() {
			case "SCall", "SRob":
				dui.Time = (timeOutCallrob - uint32(time.Now().UTC().Sub(d.reqTime).Seconds()))
				if dui.Time > timeOutCallrob {
					dui.Time = 2
				}
			case "SPlay":
				freeOut := d.isFreeOut(v.uid)
				operTime := uint32(timeOutCanOut)

				var bigerLast bool
				if !freeOut {
					bigerLast = v.currCard.HaveBiger(d.lastGiveCard)
					if !bigerLast {
						operTime = timeOutCanNotOut
					}
				}
				operMask := mask(freeOut, bigerLast)

				dui.Time = (operTime - uint32(time.Now().UTC().Sub(d.reqTime).Seconds()))
				if dui.Time > operTime {
					dui.Time = 2
				}
				dui.Mask = operMask
			}
		}

		dui.Call = v.call
		dui.Rob = v.rob
		dui.Double = v.double
		dui.Oper = v.lastOper

		if v.lastCard != nil {
			dui.LastCards = v.lastCard.Dump()
			dui.Lct = v.lastCard.Type()
		}

		if needuid != 0 && v.uid == needuid && v.currCard != nil {
			dui.HaveCards = v.currCard.Dump()
			dui.HaveCardCount = uint32(len(dui.HaveCards))
		}

		dui.IsTrustee = v.isTrustee
		dui.Mul = d.mulUser(v.uid)
		dui.DoubleEnable = (d.arg.Type == gameTypMatch)
		dui.BreakGameAgree = uint32(v.agreeBreakGame)

		di.GameUser = append(di.GameUser, dui)
	}
	return di
}
