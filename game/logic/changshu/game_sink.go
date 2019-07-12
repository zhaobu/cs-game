package main

import (
	"cy/game/configs"
	mj "cy/game/logic/changshu/majiang"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"cy/game/util"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/gogo/protobuf/proto"
)

var (
	cardDef mj.CardDef //牌定义
	huLib   mj.HuLib   //胡牌算法
)

//游戏公共信息
type gameAllInfo struct {
	waitHigestOper *OperPriority                     //当前等待中的最高优先级的操作
	operOrder      map[PriorityOrder][]*OperPriority //操作优先级
	canOperInfo    map[int32]*CanOperInfo            //玩家能做的操作
	haswaitOper    []bool                            //玩家是否有等待中的操作
	diceResult     [][2]int32                        //投色子结果
	bankerId       int32                             //庄家id
	leftCard       []int32                           //发完牌后剩余的牌
	curThrowDice   int32                             //当前投色子的玩家
	curOutChair    int32                             //当前出牌玩家
	lastOutChair   int32                             //上次出牌玩家
	lastOutCard    int32                             //上次出的牌
	laiziCard      map[int32]int32                   //癞子牌
	hasHu          bool                              //是否有人胡牌
	hasFirstBuHua  []bool                            //是否已经进行过第一次补花
	wantCards      [][]int32                         //玩家指定要的牌
	drawDiceValue  [2]int32                          //发牌色子
	// readyInfo      map[int32]bool                    //玩家准备下一局状态
}

type gamePrivateInfo struct {
	gameBalance  GameBalance             //游戏结束信息
	operAction   OperAtion               //操作
	record       mj.GameRecord           //游戏战绩回放
	players      []*mj.PlayerInfo        //玩家游戏信息
	game_config  *pbgame_logic.CreateArg //游戏参数
	baseCard     []int32                 //基础牌库
	nextBankerId int32                   //下局庄家id,每局游戏结束后确定
}

//游戏主逻辑
type GameSink struct {
	desk            *Desk //桌子
	isPlaying       bool  //是否在游戏中
	gameAllInfo           //游戏公共信息(每局所有数据都初始化)
	gamePrivateInfo       //游戏私有信息(每局部分数据初始化)
}

////////////////////////调用desk接口函数START/////////////////////////////
//发送消息给玩家(chairId为-1时发送给所有玩家,包括观察者)
func (self *GameSink) sendData(chairId int32, msg proto.Message) {
	if -1 == chairId {
		self.desk.SendGameMessage(0, msg)
	} else {
		self.desk.SendGameMessage(self.desk.GetUidByChairid(chairId), msg)
	}
}

//发送消息给所有观察者
func (self *GameSink) sendDataAllLook(msg proto.Message) {
	self.desk.SendDataAllLook(msg)
}

//发送消息给其他人,包括观察者
func (self *GameSink) sendDataOther(chairId int32, msg proto.Message) {
	uid := self.desk.GetUidByChairid(chairId)
	if uid == 0 {
		tlog.Error("sendDataOther时uid=0")
	}
	self.desk.SendGameMessageOther(uid, msg)
}

////////////////////////调用desk接口函数END/////////////////////////////

//构建游戏
func (self *GameSink) Ctor(config *pbgame_logic.CreateArg) error {
	self.game_config = config
	cardDef.Init(log)
	self.isPlaying = false
	self.players = make([]*mj.PlayerInfo, config.PlayerCount)
	self.baseCard = cardDef.GetBaseCard(config.PlayerCount)
	self.operAction.Init(config, self.laiziCard)
	self.gameBalance.Init(config)
	return nil
}

//重置游戏
func (self *GameSink) reset() {
	if self.desk.curInning == 1 {
		//所有玩家坐下后才初始化战绩记录
		// log.Debugf("传入前:%v", self.desk.getBaseDeskInfo())
		self.record.Init(self.desk.getBaseDeskInfo(), self.players, self.desk.clubId, self.desk.masterUid)
		// log.Debugf("传入后:%v", self.desk.getBaseDeskInfo())
	}
	self.record.Reset(self.desk.curInning)
	//AllInfo
	self.gameAllInfo = gameAllInfo{}
	self.operOrder = make(map[PriorityOrder][]*OperPriority, self.game_config.PlayerCount)
	self.canOperInfo = make(map[int32]*CanOperInfo, self.game_config.PlayerCount)
	self.haswaitOper = make([]bool, self.game_config.PlayerCount)
	self.diceResult = make([][2]int32, self.game_config.PlayerCount)
	self.drawDiceValue = [2]int32{}
	self.hasFirstBuHua = make([]bool, self.game_config.PlayerCount)
	self.laiziCard = make(map[int32]int32)
	self.wantCards = make([][]int32, self.game_config.PlayerCount)
	self.bankerId = -1
	self.curThrowDice = -1
	self.curOutChair = -1
	self.lastOutChair = -1

	//PrivateInfo
	self.gameBalance.Reset()
	for _, v := range self.players {
		v.Reset()
	}
}

//开始游戏
func (self *GameSink) StartGame() {
	log.Infof("%s 第%d局游戏开始", self.logHeadUser(-1), self.desk.curInning)
	self.isPlaying = true
	self.reset()
	if self.desk.curInning == 1 {
		//通知第一个玩家投色子
		self.sendData(-1, &pbgame_logic.S2CThrowDice{ChairId: 0})
		self.curThrowDice = 0
	} else {
		//由上局确定本局庄家
		self.bankerId = self.nextBankerId
		self.deal_card()
	}
}

//玩家加入游戏
func (self *GameSink) AddPlayer(chairId int32, uid uint64, uinfo *pbcommon.UserInfo) bool {
	if self.game_config.PlayerCount <= chairId {
		log.Errorf("%s 加入房间失败,人数已满,游戏开始人数为%d", self.logHeadUser(chairId), self.game_config.PlayerCount)
		return false
	}
	self.players[chairId] = mj.MakePlayers()
	self.players[chairId].BaseInfo = mj.PlayerBaseInfo{ChairId: chairId, Uid: uid, Nickname: uinfo.Name}
	return true
}

//玩家退出游戏
func (self *GameSink) Exitlayer(chairId int32) bool {
	if int(chairId) >= len(self.players) {
		log.Error("Exitlayer 时int(chairId) >= len(self.players)")
		return false
	}
	self.players[chairId] = nil
	return true
}

//改变游戏状态
func (self *GameSink) changGameState(gState pbgame_logic.GameStatus) {
	self.desk.gameStatus = gState
	self.sendData(-1, &pbgame_logic.BS2CUpdateGameStatus{GameStatus: gState})
}

//玩家投色子
func (self *GameSink) ThrowDice(chairId int32, req *pbgame_logic.C2SThrowDice) {
	//检查是否当前投色子的玩家
	if self.curThrowDice != chairId {
		log.Warnf("%s 投色子失败,当前应投色子玩家为%d", self.logHeadUser(chairId), self.curThrowDice)
		return
	}
	//检查玩家是否已经投过
	if self.diceResult[chairId][0] != 0 {
		log.Warnf("%s 已经投过色子", self.logHeadUser(chairId))
		return
	}

	//给玩家随机2个色子
	msg := &pbgame_logic.BS2CThrowDiceResult{ChairId: chairId}
	msg.DiceValue = make([]*pbgame_logic.Cyint32, 2)
	for i, rnd := range self.randDice() {
		msg.DiceValue[i] = &pbgame_logic.Cyint32{T: rnd}
		self.diceResult[chairId][i] = rnd
	}

	//广播色子结果
	self.sendData(-1, msg)
	//投色子后给客户端动画时间
	self.desk.set_timer(mj.TID_Common, 2*time.Second, func() {
		//判断是否所有人都投色子
		for i := int32(0); i < self.game_config.PlayerCount; i++ {
			if self.diceResult[i][0] == 0 {
				//2s后通知下一个玩家投色子
				self.curThrowDice = i
				self.sendData(-1, &pbgame_logic.S2CThrowDice{ChairId: i})
				return
			}
		}
		self.dealDiceResult()
	})
}

//处理投色子结果
func (self *GameSink) dealDiceResult() {
	diceRes := make([]struct {
		dice       int32
		oldChairId int32 //坐下时的chairid
	}, self.game_config.PlayerCount)
	for i := int32(0); i < self.game_config.PlayerCount; i++ {
		diceRes[i].dice = self.diceResult[i][0] + self.diceResult[i][1]
		diceRes[i].oldChairId = int32(i)
	}
	log.Debugf("排序前,dices=%+v", diceRes)
	//排序，实现比较方法即可
	sort.Slice(diceRes, func(i, j int) bool {
		if diceRes[i].dice == diceRes[j].dice {
			return diceRes[i].oldChairId < diceRes[j].oldChairId
		}
		return diceRes[i].dice > diceRes[j].dice
	})
	log.Debugf("排序后,dices=%+v", diceRes)

	//发送换座位结果
	posInfo := make([]*pbgame_logic.ChangePosInfo, len(diceRes))
	for newChair, res := range diceRes {
		uid := self.desk.GetUidByChairid(res.oldChairId)
		posInfo[newChair] = &pbgame_logic.ChangePosInfo{UserPos: int32(newChair), UserId: uid}
	}
	//切换位子
	newdiceResult := ([4][2]int32{})
	newplayers := make([]*mj.PlayerInfo, len(self.players))
	newplayChair := map[int32]*deskUserInfo{}
	log.Debugf("切换内存前,self.diceResult=%+v,self.players=%+v,self.desk.playChair=%+v", self.diceResult, self.players, self.desk.playChair)

	for newChair, res := range diceRes {
		newdiceResult[newChair] = self.diceResult[res.oldChairId]
		newplayers[newChair] = self.players[res.oldChairId]
		newplayChair[int32(newChair)] = self.desk.playChair[res.oldChairId]
		newplayChair[int32(newChair)].chairId = int32(newChair)
	}
	self.diceResult = newdiceResult[:]
	self.players = newplayers
	self.desk.playChair = newplayChair
	log.Debugf("切换内存后,self.diceResult=%+v,self.players=%+v,self.desk.playChair=%+v", self.diceResult, self.players, self.desk.playChair)

	//记录庄家
	self.bankerId = 0
	msg := &pbgame_logic.S2CChangePos{PosInfo: posInfo}
	self.sendData(-1, msg)
	//1s后发送游戏开始消息
	self.desk.set_timer(mj.TID_Common, 2*time.Second, func() {
		self.deal_card()
	})
}

//随机摇一次色子
func (self *GameSink) randDice() [2]int32 {
	rand.Seed(time.Now().Unix())
	res := [2]int32{}
	for i := 0; i < 2; i++ {
		res[i] = int32(rand.Intn(5) + 1)
	}
	return res
}

//开始发牌
func (self *GameSink) deal_card() {
	self.changGameState(pbgame_logic.GameStatus_GSPlaying)
	//随机2个色子,用于客户端选择从牌堆摸牌的方向
	msg := &pbgame_logic.S2CStartGame{BankerId: self.bankerId, CurInning: self.desk.curInning, LeftTime: 15}
	msg.DiceValue = make([]*pbgame_logic.Cyint32, 2)
	self.drawDiceValue = self.randDice()
	self.gameBalance.DealStartDice(self.drawDiceValue)
	for i, rnd := range self.drawDiceValue {
		msg.DiceValue[i] = &pbgame_logic.Cyint32{T: rnd}
	}

	//洗牌
	self.shuffle_cards()
	tlog.Debug("发牌前的牌库为", zap.Any("self.leftCard", self.leftCard))
	var player_cards [][]int32
	player_cards, self.leftCard = cardDef.DealCard(self.leftCard, self.game_config.PlayerCount, self.bankerId)

	//庄家手牌
	bankerCardInfo := &self.players[self.bankerId].CardInfo
	bankerCardInfo.HandCards = player_cards[self.bankerId]
	bankerCardInfo.StackCards = mj.CalStackCards(player_cards[self.bankerId], false)
	//庄家开始第一次补花
	huaCards, _ := self.firstBuHuaCards(self.bankerId)
	self.operAction.HandleBuHua(self.players[self.bankerId], huaCards)
	msg.HuaCards = switchToCyint32(huaCards)
	msg.LeftNum = int32(len(self.leftCard))
	msg.TotalNum = int32(len(self.baseCard))
	self.curOutChair = self.bankerId
	self.resetOper()
	//分析庄家能做的操作
	ret := self.operAction.BankerAnalysis(self.players[self.bankerId], self.bankerId, mj.HuMode_ZIMO)
	//统计能做的操作
	if !ret.Empty() {
		msg.BankerOper = &pbgame_logic.S2CHaveOperation{ChairId: self.bankerId}
		self.countCanOper(ret, self.bankerId, msg.BankerOper)
	}
	//发送观察者
	self.sendDataAllLook(msg)
	//回放记录所有人手牌
	recordCard := &pbgame_logic.Json_UserCardInfo{HandCards: map[int32]*pbgame_logic.Json_UserCardInfoCards{}}
	for k, v := range self.players {
		if int32(k) != self.bankerId {
			v.CardInfo.HandCards = player_cards[k]
			v.CardInfo.StackCards = mj.CalStackCards(player_cards[k], false)
		}

		tmp := &pbgame_logic.Json_UserCardInfo{HandCards: map[int32]*pbgame_logic.Json_UserCardInfoCards{}}
		tmp.HandCards[int32(k)] = &pbgame_logic.Json_UserCardInfoCards{Cards: v.CardInfo.HandCards}
		recordCard.HandCards[int32(k)] = tmp.HandCards[int32(k)]
		msg.JsonAllCards = util.PB2JSON(tmp, false)
		log.Warnf("%s手牌为:%v", self.logHeadUser(int32(k)), player_cards[k])
		//给每个玩家发送游戏开始消息
		self.sendData(int32(k), msg)
	}
	//发送听牌信息
	if listenMsg, ok := self.operAction.GetListenInfo(self.bankerId, self.players, nil, self.leftCard); ok {
		self.sendData(self.bankerId, listenMsg)
	}
	//游戏回放记录
	msg.JsonAllCards = util.PB2JSON(recordCard, false)
	self.record.RecordGameAction(msg)
}

func switchToCyint32(cards []int32) []*pbgame_logic.Cyint32 {
	res := []*pbgame_logic.Cyint32{}
	for _, card := range cards {
		res = append(res, &pbgame_logic.Cyint32{T: card})
	}
	return res
}

//玩家第一次补花,返回所有的花牌,所有摸到的牌
func (self *GameSink) firstBuHuaCards(chairId int32) (huaCards, moCards []int32) {
	if self.hasFirstBuHua[chairId] {
		tlog.Error("玩家第一次补花执行了多次", zap.Int32("chairId", chairId))
	}
	self.hasFirstBuHua[chairId] = true
	cardInfo := &self.players[chairId].CardInfo
	tlog.Debug("玩家第一次补花前手牌数据为", zap.Int32("chairId", chairId), zap.Any("cardInfo", cardInfo))

	tmpHandCards := make([]int32, len(cardInfo.HandCards))
	copy(tmpHandCards, cardInfo.HandCards)
	for _, card := range tmpHandCards {
		if mj.IsHuaCard(card) {
			tmpHuaCards, moCard := self.drawOneCard()
			tlog.Debug("补花", zap.Int32("huacard", card), zap.Any("moCard", moCard), zap.Any("tmpHuaCards", tmpHuaCards))
			self.operAction.updateCardInfo(cardInfo, nil, []int32{card}) //减掉原有的花
			self.operAction.updateCardInfo(cardInfo, moCard, nil)        //加上摸到的牌
			tmpHuaCards = append(tmpHuaCards, card)                      //加上原有的花
			huaCards = append(huaCards, tmpHuaCards...)                  //记录花牌
			moCards = append(moCards, moCard...)                         //记录摸到的牌
		}
	}
	tlog.Debug("玩家第一次补花后手牌数据为", zap.Int32("chairId", chairId), zap.Any("cardInfo", cardInfo))
	return
}

//从牌堆摸一张牌,摸到不是花牌为止
func (self *GameSink) drawOneCard() (huaCards, moCard []int32) {
	var num int32
	huaCards, moCard = []int32{}, []int32{}
	for {
		if len(self.leftCard) == 0 {
			break
		}
		card := self.leftCard[len(self.leftCard)-1]
		self.leftCard = self.leftCard[:len(self.leftCard)-1]
		if !mj.IsHuaCard(card) {
			moCard = append(moCard, card)
			break
		}
		huaCards = append(huaCards, card)
		num++
		if num > 12 {
			tlog.Error("drawOneCard死循环")
			break
		}
	}
	tlog.Debug("drawOneCard 结果", zap.Any("huaCards", huaCards), zap.Any("moCard", moCard))
	return
}

func (self *GameSink) resetOper() {
	self.canOperInfo = map[int32]*CanOperInfo{}
	for i := int32(0); i < self.game_config.PlayerCount; i++ {
		self.canOperInfo[i] = &CanOperInfo{}
	}
	self.waitHigestOper = nil
	self.operOrder = map[PriorityOrder][]*OperPriority{}
}

//摸牌 last(0从牌前摸,1从摸牌尾)
func (self *GameSink) drawCard(chairId, last int32) error {
	log.Debugf("%s,摸牌操作,last=%d", self.logHeadUser(chairId), last)
	//检查游戏是否结束
	if len(self.leftCard) <= 0 {
		self.gameEnd(pbgame_logic.GameEndType_EndDeuce)
		return nil
	}

	self.resetOper()

	msg := &pbgame_logic.BS2CDrawCard{ChairId: chairId, DrawPos: last}
	var huaCards, moCards []int32
	var moCount int //总共摸牌的次数
	log.Debugf("%s,玩家%d摸牌前手牌数据为%+v,牌库剩余牌:%v", self.logHeadUser(chairId), chairId, self.players[chairId].CardInfo, self.leftCard)
	if !self.hasFirstBuHua[chairId] { //没有进行过第一次补花,先补掉手上的牌
		if mj.GetHuaCount(self.players[chairId].CardInfo.StackCards) > 0 {
			log.Debugf("%s 第一次摸牌,需要补花,补花前剩余[%d]张", self.logHeadUser(chairId), len(self.leftCard))
			huaCards, moCards = self.firstBuHuaCards(chairId)
			moCount += len(huaCards)
		}
		self.hasFirstBuHua[chairId] = true
	}
	huaCards2, moCards2 := self.drawOneCard()
	moCount += len(huaCards2) + len(moCards2)
	huaCards, moCards = append(huaCards, huaCards2...), append(moCards, moCards2...)

	msg.LeftNum = int32(len(self.leftCard))
	var card int32
	if len(moCards) > 0 {
		card = moCards[len(moCards)-1] //最后摸到的牌
	}
	//发送自己
	msg.JsonDrawInfo = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: moCards, MoCount: int32(moCount)}, false)
	self.sendData(chairId, msg)

	//游戏回放记录
	self.record.RecordGameAction(msg)
	//发送别人
	msg.JsonDrawInfo = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: make([]int32, len(moCards)), MoCount: int32(moCount)}, false)
	self.sendDataOther(chairId, msg)

	cardInfo := &self.players[chairId].CardInfo
	if len(huaCards) > 0 {
		self.operAction.HandleBuHua(self.players[chairId], huaCards)
	}

	//如果摸牌时摸到的全部是花牌,游戏结束
	if len(moCards) == 0 {
		self.gameEnd(pbgame_logic.GameEndType_EndDeuce)
		return nil
	}
	self.curOutChair = chairId
	//当抓到花或杠牌后，补上一张牌,能胡,杠上开花
	huModeTags := map[mj.EmHuModeTag]bool{}
	if last == 1 || len(huaCards2) > 0 {
		huModeTags[mj.HuModeTag_GangShangHua] = true
	}
	//分析能否暗杠,补杠,自摸胡
	ret := self.operAction.DrawcardAnalysis(self.players[chairId], chairId, card, int32(len(self.leftCard)), huModeTags)
	log.Infof("%s 摸牌后操作分析ret=%+v", self.logHeadUser(chairId), ret)
	//更新玩家card_info表
	self.operAction.HandleDrawCard(cardInfo, card)
	//发送听牌信息
	if listenMsg, ok := self.operAction.GetListenInfo(chairId, self.players, huModeTags, self.leftCard); ok {
		self.sendData(chairId, listenMsg)
	}
	//统计能做的操作
	if !ret.Empty() {
		msg := &pbgame_logic.S2CHaveOperation{ChairId: chairId}
		self.countCanOper(ret, chairId, msg)
		//发送玩家可进行的操作
		self.sendData(chairId, msg)
	}
	return nil
}

//检查吃碰后能做的操作
func (self *GameSink) checkAfterChiPeng(chairId, pengCard int32) {
	if !self.hasFirstBuHua[chairId] { //第一次补花
		huaCards, moCards := self.firstBuHuaCards(chairId)
		if len(huaCards) > 0 {
			self.operAction.HandleBuHua(self.players[chairId], huaCards)
			msg := &pbgame_logic.BS2CFirstBuHua{ChairId: chairId, LeftNum: int32(len(self.leftCard))}
			//发给自己
			msg.JsonFirstBuhua = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: moCards, MoCount: int32(len(huaCards))}, false)
			self.sendData(chairId, msg)
			//游戏回放记录
			self.record.RecordGameAction(msg)
			//发给别人
			msg.JsonFirstBuhua = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: make([]int32, len(moCards)), MoCount: int32(len(huaCards))}, false)
			self.sendDataOther(chairId, msg)
		}
	}

	cardInfo := &self.players[chairId].CardInfo
	//分析能否暗杠,补杠,不能补杠刚刚碰的那张牌
	ret := self.operAction.AfterChiPengAnalysis(cardInfo, chairId, pengCard)
	log.Infof("%s 吃碰后操作分析ret=%+v", self.logHeadUser(chairId), ret)
	//统计能做的操作
	if !ret.Empty() {
		msg := &pbgame_logic.S2CHaveOperation{ChairId: chairId}
		self.countCanOper(ret, chairId, msg)
		//发送玩家可进行的操作
		self.sendData(chairId, msg)
	}
}

//出牌
func (self *GameSink) outCard(chairId, card int32) error {
	log.Debugf("%s,出牌操作,card=%d", self.logHeadUser(chairId), card)
	//检查是否在游戏中
	if !self.isPlaying {
		log.Errorf("%s 出牌失败,不在游戏中", self.logHeadUser(chairId))
		return nil
	}
	//检查是否轮到自己出牌
	if self.curOutChair != chairId {
		log.Errorf("%s 出牌失败,还没轮到你", self.logHeadUser(chairId))
		return nil
	}
	//出牌前检测是否还有其可执行的操作没有完成
	if !self.canOperInfo[chairId].Empty() {
		log.Infof("%s 出牌时还有其他操作，取消所有能做的操作", self.logHeadUser(chairId))
		self.resetOper()
	}
	cardInfo := &self.players[chairId].CardInfo
	//判断是否有这张牌
	if _, ok := cardInfo.StackCards[card]; !ok {
		log.Errorf("%s 出牌失败,手上没有这张牌", self.logHeadUser(chairId))
		return nil
	}
	//检查是否是吃碰后不能打的牌
	if cardInfo.CanNotOut[card] == card {
		log.Errorf("%s 出牌失败,是吃碰后不能打的牌", self.logHeadUser(chairId))
		return nil
	}
	cardInfo.GuoPeng = false
	cardInfo.CanNotOut = map[int32]int32{}
	//更新玩家card_info表
	self.operAction.HandleOutCard(cardInfo, card)
	msg := &pbgame_logic.BS2COutCard{ChairId: chairId, Card: card}
	self.sendData(-1, msg)
	//游戏回放记录
	self.record.RecordGameAction(msg)
	self.lastOutChair, self.lastOutCard, self.curOutChair = chairId, card, -1
	//检查出牌后能做的操作
	willWait := false
	for k, v := range self.players {
		if int32(k) != chairId {
			ret := self.operAction.OutCardAnalysis(v, card, int32(k), chairId, int32(len(self.leftCard)))
			if !ret.Empty() {
				//统计并记录玩家可以进行的操作
				msg := &pbgame_logic.S2CHaveOperation{ChairId: int32(k)}
				self.countCanOper(ret, int32(k), msg)
				willWait = true
				//发送玩家可进行的操作
				log.Infof("%s 可进行的操作%+v", self.logHeadUser(int32(k)), ret)
				self.sendData(int32(k), msg)
			}
		}
	}
	if !willWait { //出牌后无其他人能操作
		self.drawCard(GetNextChair(chairId, self.game_config.PlayerCount), 0)
	}
	return nil
}

//新增优先级操作
func (self *GameSink) addOperOrder(order PriorityOrder, oper *OperPriority) {
	if _, ok := self.operOrder[order]; ok {
		self.operOrder[order] = append(self.operOrder[order], oper)
	} else {
		self.operOrder[order] = []*OperPriority{oper}
	}
}

//构建玩家可进行操作消息
func (self *GameSink) countCanOperMsg(ret *CanOperInfo, msg *pbgame_logic.S2CHaveOperation) {
	//记录能吃
	if !ret.CanChi.Empty() {
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskChi) | msg.OperMask
		msg.CanChi = ret.CanChi.ChiType
		msg.Card = ret.CanChi.Card
	}
	//记录能碰
	if !ret.CanPeng.Empty() {
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskPeng) | msg.OperMask
		msg.Card = ret.CanPeng.Card
	}
	//记录能杠
	if !ret.CanGang.Empty() {
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskGang) | msg.OperMask
		for _, gangCard := range ret.CanGang.GangList {
			msg.CanGang = append(msg.CanGang, &pbgame_logic.Cyint32{T: gangCard})
		}
	}
	//记录能胡
	if !ret.CanHu.Empty() {
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskHu) | msg.OperMask
		msg.Card = ret.CanHu.Card
	}
}

//统计并记录玩家可以进行的操作
func (self *GameSink) countCanOper(ret *CanOperInfo, chairId int32, msg *pbgame_logic.S2CHaveOperation) {
	//记录能吃
	if !ret.CanChi.Empty() {
		self.canOperInfo[chairId].CanChi = ret.CanChi
		self.addOperOrder(ChiOrder, &OperPriority{ChairId: chairId, Op: ChiOrder, Info: &self.canOperInfo[chairId].CanChi})
	}
	//记录能碰
	if !ret.CanPeng.Empty() {
		self.canOperInfo[chairId].CanPeng = ret.CanPeng
		self.addOperOrder(PengOrder, &OperPriority{ChairId: chairId, Op: PengOrder, Info: &self.canOperInfo[chairId].CanPeng})
	}
	//记录能杠
	if !ret.CanGang.Empty() {
		self.canOperInfo[chairId].CanGang = ret.CanGang
		self.addOperOrder(GangOrder, &OperPriority{ChairId: chairId, Op: GangOrder, Info: &self.canOperInfo[chairId].CanGang})
	}
	//记录能胡
	if !ret.CanHu.Empty() {
		self.canOperInfo[chairId].CanHu = ret.CanHu
		self.addOperOrder(HuOrder, &OperPriority{ChairId: chairId, Op: HuOrder, Info: &self.canOperInfo[chairId].CanHu})
	}
	self.countCanOperMsg(ret, msg)
}

//洗牌
func (self *GameSink) shuffle_cards() {
	if !*release && configs.Conf.GameNode[gameName].GameTest != "" {
		self.leftCard = cardDef.GetDebugCards(gameName, self.baseCard, self.game_config.PlayerCount)
		return
	}
	log.Debugf("*release=%v,configs.Conf.GameNode[gameName].GameTest=%v", *release, configs.Conf.GameNode[gameName].GameTest)
	self.leftCard = mj.RandCards(self.baseCard)
}

//返回1表示能直接进行该操作,返回2表示还需要等待,返回3表示需要唤醒等待中的操作
func (self *GameSink) checkPlayerOperationNeedWait(chairId int32, curOrder PriorityOrder) int {
	var otherOrder, waitOrder PriorityOrder = NoneOrder, NoneOrder
	for k, v := range self.operOrder {
		for k1, v1 := range v {
			log.Debugf("self.operOrder[%d][%d]=%+v", k, k1, v1)
		}
	}
	//检查其他人能做的最高优先级操作
	for i := HuOrder; i >= ChiOrder; i-- {
		if curOper, ok := self.operOrder[i]; ok {
			for _, v := range curOper {
				if v.ChairId != chairId {
					otherOrder = i
					break
				}
			}
			if otherOrder != NoneOrder {
				break
			}
		}
	}
	if self.waitHigestOper != nil {
		waitOrder = self.waitHigestOper.Op
	}
	//比较当前操作,当前等待中的最高优先级的操作,和其他人能做的最高优先级操作
	maxOrder := self.operAction.GetMaxOrder(self.operAction.GetMaxOrder(otherOrder, waitOrder), curOrder)
	if maxOrder == curOrder {
		return 1
	} else if maxOrder == otherOrder {
		return 2
	} else if maxOrder == waitOrder {
		return 3
	}
	log.Errorf("异常操作优先结果")
	return 0
}

//删除玩家操作优先级记录,bdel表示是否删除成功
func (self *GameSink) deletePlayerOperOrder(chairId int32) (bdel bool) {
	for order, allOper := range self.operOrder {
		for i, oneOper := range allOper {
			if oneOper.ChairId == chairId { //每种操作最多只可能有一个
				self.operOrder[order] = append(allOper[:i], allOper[i+1:]...)
				bdel = true
			}
			if len(self.operOrder[order]) == 0 {
				delete(self.operOrder, order)
			}
		}
	}
	return
}

//插入等待中的操作(info为指针)
func (self *GameSink) insertWaitOper(chairId int32, op PriorityOrder, info interface{}) {
	if self.waitHigestOper != nil && op < self.waitHigestOper.Op {
		return
	}
	if self.waitHigestOper != nil {
		log.Debugf("%s 高优先级操作%v替换掉低优先级操作%v", self.logHeadUser(chairId), op, self.waitHigestOper.Op)
	}
	self.waitHigestOper = &OperPriority{ChairId: chairId, Op: op, Info: info}
}

//唤醒等待中的操作
func (self *GameSink) dealWaitOper(chairId int32) {
	if self.waitHigestOper == nil {
		log.Errorf("%s 唤醒操作时self.waitHigestOper == nil", self.logHeadUser(chairId))
	}
	info, _ := self.waitHigestOper.Info.(*WaitOperRecord)
	switch self.waitHigestOper.Op {
	case ChiOrder:
		log.Debugf("%s 唤醒操作吃", self.logHeadUser(chairId))
		self.chiCard(self.waitHigestOper.ChairId, info.Card, info.ChiType)
	case PengOrder:
		log.Debugf("%s 唤醒操作碰", self.logHeadUser(chairId))
		self.pengCard(self.waitHigestOper.ChairId, info.Card)
	case GangOrder:
		log.Debugf("%s 唤醒操作杠", self.logHeadUser(chairId))
		self.gangCard(self.waitHigestOper.ChairId, info.Card)
	case HuOrder:
		log.Debugf("%s 唤醒操作胡", self.logHeadUser(chairId))
		self.huCard(self.waitHigestOper.ChairId)
	default:
		log.Debugf("%s 唤醒操作时,类型转换失败", self.logHeadUser(chairId))
		return
	}
	self.waitHigestOper = nil
}

//吃
func (self *GameSink) chiCard(chairId, card int32, chiType uint32) error {
	log.Debugf("%s,吃牌操作,card=%d,chiType=%d", self.logHeadUser(chairId), card, chiType)
	//检查是否在游戏中
	if !self.isPlaying {
		log.Errorf("%s 吃牌失败,不在游戏中", self.logHeadUser(chairId))
		return nil
	}
	//检测是否能吃
	if self.canOperInfo[chairId] == nil || self.canOperInfo[chairId].CanChi.Empty() {
		log.Errorf("%s 吃牌失败,没有该操作", self.logHeadUser(chairId))
		return nil
	}
	//校验操作参数合法性
	if chiType == 0 || card != self.canOperInfo[chairId].CanChi.Card || chiType != (self.canOperInfo[chairId].CanChi.ChiType&chiType) {
		log.Errorf("%s 吃牌失败,没有该吃类型,或者牌不对,card=%d,chiType=%d,CanChi=%+v", self.logHeadUser(chairId), card, chiType, self.canOperInfo[chairId].CanChi)
		return nil
	}
	self.deletePlayerOperOrder(chairId)
	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, ChiOrder)
	if res == 2 { //需要等待其他人操作
		log.Debugf("%s 操作吃需要等待其他人", self.logHeadUser(chairId))
		self.insertWaitOper(chairId, ChiOrder, &WaitOperRecord{Card: card, ChiType: chiType})
		self.haswaitOper[chairId] = true
		return nil
	} else if res == 3 { //唤醒等待中的操作
		log.Debugf("%s 操作吃,唤醒等待中的操作", self.logHeadUser(chairId))
		self.dealWaitOper(chairId)
		return nil
	}
	//判断是否已经胡
	if self.hasHu {
		log.Debugf("%s 操作吃,因为已经有人胡牌,游戏结束", self.logHeadUser(chairId))
		self.gameEnd(pbgame_logic.GameEndType_EndHu)
		return nil
	}
	//更新玩家card_info表
	self.operAction.HandleChiCard(&self.players[chairId].CardInfo, &self.players[self.lastOutChair].CardInfo, card, self.lastOutChair, chiType)
	//回放记录

	//变量维护
	self.addCanNotOut(chairId, card, chiType)
	self.curOutChair = chairId
	self.haswaitOper[chairId] = false
	self.resetOper()

	msg := &pbgame_logic.BS2CChiCard{ChairId: chairId, Card: card, ChiType: chiType}
	self.sendData(-1, msg)

	//游戏回放记录
	self.record.RecordGameAction(msg)
	self.checkAfterChiPeng(chairId, 0)
	//发送听牌信息
	if listenMsg, ok := self.operAction.GetListenInfo(chairId, self.players, nil, self.leftCard); ok {
		self.sendData(chairId, listenMsg)
	}
	return nil
}

//碰
func (self *GameSink) pengCard(chairId, card int32) error {
	log.Debugf("%s,碰牌操作,card=%d", self.logHeadUser(chairId), card)
	//检查是否在游戏中
	if !self.isPlaying {
		log.Errorf("%s 碰牌失败,不在游戏中", self.logHeadUser(chairId))
		return nil
	}
	//检测是否能碰
	if self.canOperInfo[chairId] == nil || self.canOperInfo[chairId].CanPeng.Empty() {
		log.Errorf("%s 碰牌失败,没有该操作", self.logHeadUser(chairId))
		return nil
	}
	//校验操作参数合法性
	if card != self.canOperInfo[chairId].CanPeng.Card {
		log.Errorf("%s 碰牌失败,碰的牌不对,card=%d,CanPeng=%+v", self.logHeadUser(chairId), card, self.canOperInfo[chairId].CanPeng)
		return nil
	}
	self.deletePlayerOperOrder(chairId)
	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, PengOrder)
	if res == 2 { //需要等待其他人操作
		log.Debugf("%s 操作碰需要等待其他人", self.logHeadUser(chairId))
		self.insertWaitOper(chairId, PengOrder, &WaitOperRecord{Card: card})
		self.haswaitOper[chairId] = true
		return nil
	} else if res == 3 { //唤醒等待中的操作
		log.Debugf("%s 操作碰,唤醒等待中的操作", self.logHeadUser(chairId))
		self.dealWaitOper(chairId)
		return nil
	}
	//判断是否已经胡
	if self.hasHu {
		log.Debugf("%s 操作碰,因为已经有人胡牌,游戏结束", self.logHeadUser(chairId))
		self.gameEnd(pbgame_logic.GameEndType_EndHu)
		return nil
	}
	//更新玩家card_info表
	self.operAction.HandlePengCard(self.players[chairId], &self.players[self.lastOutChair].CardInfo, card, self.canOperInfo[chairId].CanPeng.LoseChair)
	//回放记录

	msg := &pbgame_logic.BS2CPengCard{ChairId: chairId, LoseChair: self.lastOutChair, Card: card}
	self.sendData(-1, msg)
	//游戏回放记录
	self.record.RecordGameAction(msg)

	//变量维护
	self.addCanNotOut(chairId, card, 0)
	self.curOutChair = chairId
	self.haswaitOper[chairId] = false
	self.resetOper()

	self.checkAfterChiPeng(chairId, card)
	//发送听牌信息
	if listenMsg, ok := self.operAction.GetListenInfo(chairId, self.players, nil, self.leftCard); ok {
		self.sendData(chairId, listenMsg)
	}
	return nil
}

//杠
func (self *GameSink) gangCard(chairId, card int32) error {
	log.Debugf("%s,杠牌操作,card=%d", self.logHeadUser(chairId), card)
	//检查是否在游戏中
	if !self.isPlaying {
		log.Errorf("%s 杠牌失败,不在游戏中", self.logHeadUser(chairId))
		return nil
	}
	//检测是否能杠
	if self.canOperInfo[chairId] == nil || self.canOperInfo[chairId].CanGang.Empty() {
		log.Errorf("%s 杠牌失败,没有该操作", self.logHeadUser(chairId))
		return nil
	}
	//校验操作参数合法性
	if _, ok := self.canOperInfo[chairId].CanGang.GangList[card]; !ok {
		log.Errorf("%s 杠牌失败,牌不对,card=%d,CanGang=%+v", self.logHeadUser(chairId), card, self.canOperInfo[chairId].CanGang)
		return nil
	}
	self.deletePlayerOperOrder(chairId)
	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, GangOrder)
	if res == 2 { //需要等待其他人操作
		log.Debugf("%s 操作杠需要等待其他人", self.logHeadUser(chairId))
		self.insertWaitOper(chairId, GangOrder, &WaitOperRecord{Card: card})
		self.haswaitOper[chairId] = true
		return nil
	} else if res == 3 { //唤醒等待中的操作
		log.Debugf("%s 操作杠,唤醒等待中的操作", self.logHeadUser(chairId))
		self.dealWaitOper(chairId)
		return nil
	}
	//判断是否已经胡
	if self.hasHu {
		log.Debugf("%s 操作杠,因为已经有人胡牌,游戏结束", self.logHeadUser(chairId))
		self.gameEnd(pbgame_logic.GameEndType_EndHu)
		return nil
	}

	//变量维护

	self.haswaitOper[chairId] = false

	self.curOutChair = -1 //玩家杠牌后 当前出牌玩家还不是自己 要摸牌后才能出牌
	gangType := self.operAction.GetGangType(&self.players[chairId].CardInfo, card)

	var loseChair int32 = -1                            //如果是补杠,为碰牌时被碰玩家
	if gangType == pbgame_logic.OperType_Oper_BU_GANG { //补杠
		log.Debugf("%s 补杠,杠牌为%d", self.logHeadUser(chairId), card)
		loseChair = self.players[chairId].CardInfo.PengCards[card]
	} else if gangType == pbgame_logic.OperType_Oper_MING_GANG { //明杠
		log.Debugf("%s 明杠,杠牌为%d", self.logHeadUser(chairId), card)
		loseChair = self.lastOutChair
	} else if gangType == pbgame_logic.OperType_Oper_AN_GANG { //暗杠
		log.Debugf("%s 暗杠,杠牌为%d", self.logHeadUser(chairId), card)
	}

	willWait := false
	//判断抢杠胡
	if gangType == pbgame_logic.OperType_Oper_BU_GANG {
		for k, v := range self.players {
			if int32(k) != chairId && self.canOperInfo[int32(k)].CanHu.Empty() { //避免抢杠胡玩家取消后重复判断抢杠胡
				ret := self.operAction.QiangGangAnalysis(v, card, int32(k), int32(chairId))
				if !ret.Empty() {
					//统计并记录玩家可以进行的操作
					msg := &pbgame_logic.S2CHaveOperation{ChairId: int32(k)}
					self.countCanOper(ret, int32(k), msg)
					//标记有一次机会可以抢杠胡
					willWait = true

					//发送玩家可进行的操作
					log.Infof("%s 可进行的操作%+v", self.logHeadUser(int32(k)), ret)
					self.sendData(int32(k), msg)
				}
			}
		}
	}

	//如果能抢杠胡,需要等待玩家操作
	if willWait {
		log.Debugf("%s 操作杠时其他玩家可以抢杠胡,需要等待", self.logHeadUser(chairId))
		self.insertWaitOper(chairId, GangOrder, &WaitOperRecord{Card: card})
		self.haswaitOper[chairId] = true
		return nil
	}
	self.resetOper()

	msg := &pbgame_logic.BS2CGangCard{ChairId: chairId, Card: card, Type: pbgame_logic.GangType(gangType), LoseChair: loseChair}
	self.sendData(-1, msg)
	//游戏回放记录
	self.record.RecordGameAction(msg)

	//更新玩家card_info表
	if gangType == pbgame_logic.OperType_Oper_MING_GANG {
		self.operAction.HandleGangCard(self.players[chairId], &self.players[self.lastOutChair].CardInfo, card, gangType, loseChair)
	} else {
		self.operAction.HandleGangCard(self.players[chairId], nil, card, gangType, loseChair)
	}

	//变量维护
	self.haswaitOper[chairId] = false
	self.resetOper()

	self.drawCard(chairId, 1)
	return nil
}

//胡
func (self *GameSink) huCard(chairId int32) error {
	log.Debugf("%s,胡牌操作", self.logHeadUser(chairId))
	//检查是否在游戏中
	if !self.isPlaying {
		log.Errorf("%s 胡牌失败,不在游戏中", self.logHeadUser(chairId))
		return nil
	}
	//检测是否能胡
	if self.canOperInfo[chairId] == nil || self.canOperInfo[chairId].CanHu.Empty() {
		log.Errorf("%s 胡牌失败,没有该操作", self.logHeadUser(chairId))
		return nil
	}
	self.deletePlayerOperOrder(chairId)
	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, HuOrder)
	if res == 2 { //需要等待其他人操作
		log.Debugf("%s 操作胡需要等待其他人", self.logHeadUser(chairId))
		self.insertWaitOper(chairId, HuOrder, nil)
		self.haswaitOper[chairId] = true
		return nil
	} else if res == 3 { //唤醒等待中的操作
		log.Errorf("%s 操作胡执行了唤醒操作,checking!!!!", self.logHeadUser(chairId))
		return nil
	}

	huInfo := &self.canOperInfo[chairId].CanHu
	cardInfo := &self.players[chairId].CardInfo
	//回放记录

	self.hasHu = true
	self.players[chairId].BalanceInfo.HuCard = huInfo.Card

	//接炮 or 抢杠胡 把胡的牌加到手牌里
	if huInfo.HuMode != mj.HuMode_ZIMO {
		self.operAction.updateCardInfo(cardInfo, []int32{huInfo.Card}, nil)
	}

	//记录胡牌牌型
	self.gameBalance.loseChair = huInfo.LoseChair
	self.gameBalance.huCard = huInfo.Card
	self.gameBalance.huMode = huInfo.HuMode

	self.gameBalance.huChairs[chairId] = &HuScoreInfo{HuTypeList: huInfo.HuList}

	//统计总结算次数
	self.gameBalance.AddScoreTimes(&self.players[chairId].BalanceResult, mj.ScoreTimes_Win)

	msg := &pbgame_logic.BS2CHuCard{ChairId: chairId, HuCard: huInfo.Card}
	self.sendData(-1, msg)
	//游戏回放记录
	self.record.RecordGameAction(msg)

	//变量维护
	self.haswaitOper[chairId] = false

	if len(self.operOrder[HuOrder]) == 0 {
		self.gameEnd(pbgame_logic.GameEndType_EndHu)
	}
	return nil
}

//取消操作
func (self *GameSink) cancelOper(chairId int32) error {
	log.Debugf("%s,取消操作", self.logHeadUser(chairId))
	//检查是否在游戏中
	if !self.isPlaying {
		log.Errorf("%s 取消失败,不在游戏中", self.logHeadUser(chairId))
		return nil
	}
	//检测是否能取消
	if self.canOperInfo[chairId] == nil || self.canOperInfo[chairId].Empty() || !self.deletePlayerOperOrder(chairId) {
		log.Errorf("%s 取消失败,没有可取消的操作", self.logHeadUser(chairId))
		return nil
	}

	if !self.canOperInfo[chairId].CanPeng.Empty() {
		self.players[chairId].CardInfo.GuoPeng = true
	}
	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, NoneOrder)
	if res == 2 { //需要等待其他人操作
		log.Debugf("%s 取消操作需要等待其他人", self.logHeadUser(chairId))
		return nil
	} else if res == 3 { //唤醒等待中的操作
		log.Debugf("%s 取消操作,唤醒等待中的操作", self.logHeadUser(chairId))
		self.dealWaitOper(chairId)
		return nil
	}
	//判断是否已经胡
	if self.hasHu {
		log.Debugf("%s 取消操作,因为已经有人胡牌,游戏结束", self.logHeadUser(chairId))
		self.gameEnd(pbgame_logic.GameEndType_EndHu)
		return nil
	}

	//回放记录

	//变量维护
	self.haswaitOper[chairId] = false
	self.resetOper()

	//取消操作后,由上次出牌玩家下家抓牌
	if self.curOutChair != chairId {
		self.drawCard(GetNextChair(self.lastOutChair, self.game_config.PlayerCount), 0)
	}
	return nil
}

//游戏结束
func (self *GameSink) gameEnd(endType pbgame_logic.GameEndType) {
	self.changGameState(pbgame_logic.GameStatus_GSGameEnd)
	log.Infof("%s 第%d局游戏结束,结束原因%d,结束时剩余牌为%v", self.logHeadUser(-1), self.desk.curInning, endType, self.leftCard)
	if !self.isPlaying { //可能是解散导致游戏结束
		self.desk.gameEnd(endType)
		return
	}
	self.isPlaying = false
	//发送小结算信息
	msg := &pbgame_logic.BS2CGameEnd{CurInning: self.desk.curInning, Banker: self.bankerId, EndType: endType}
	if self.game_config.Barhead == 3 && self.hasHu {
		msg.DulongCard = self.leftCard[len(self.leftCard)-1]
		//如果是独龙杠,把独龙杠的牌从牌堆去掉
		self.leftCard = self.leftCard[:len(self.leftCard)-1]
	}
	self.gameBalance.CalGangTou(self.leftCard, self.bankerId)
	self.gameBalance.CalGameBalance(self.players, self.bankerId)
	strPlayerBalance := &pbgame_logic.Json_PlayerBalance{PlayerBalanceInfo: self.gameBalance.GetPlayerBalanceInfo(self.players)}
	msg.JsonPlayerBalance = util.PB2JSON(strPlayerBalance, false)

	self.sendData(-1, msg)
	//游戏回放记录
	self.record.RecordGameAction(msg)
	scoreInfo := map[int32]int32{}
	for k, v := range self.players {
		scoreInfo[int32(k)] = v.BalanceInfo.Point
	}
	self.record.AddGameRecord(scoreInfo)
	self.record.RecordGameEnd(self.players)
	//主动发送一下战绩
	self.getGameRecord(-1)
	//游戏记录
	self.afterGameEnd(endType)
}

//小局结束后数据清理
func (self *GameSink) afterGameEnd(endType pbgame_logic.GameEndType) {
	//判断下一局庄家
	self.nextBankerId = self.gameBalance.CalNextBankerId(self.bankerId)
	// self.readyInfo = make(map[int32]bool, self.game_config.PlayerCount)
	self.desk.gameEnd(endType)
}

//断线重连
func (self *GameSink) gameReconnect(recInfo *pbgame_logic.GameDeskInfo, uid uint64) {
	chairId := self.desk.GetChairidByUid(uid)
	log.Infof("%s 第%d局玩家%d断线重连,chairId=%d", self.logHeadUser(-1), self.desk.curInning, uid, chairId)
	//chairId为-1时为观察者游戏中途进入房间
	switch recInfo.GameStatus {
	case pbgame_logic.GameStatus_GSDice: //投色子
		recInfo.CurDiceChair = self.curThrowDice
		for _, v := range recInfo.GameUser {
			k := v.ChairId
			if mj.GetNextChair(k, self.game_config.PlayerCount) == self.curThrowDice {
				recInfo.LastDiceValue = switchToCyint32(self.diceResult[k][:])
			}
			v.DiceValue = self.diceResult[k][0] + self.diceResult[k][1]
		}
	case pbgame_logic.GameStatus_GSPlaying: //行牌中
		recInfo.BankerId = self.bankerId
		canOper := &pbgame_logic.S2CHaveOperation{ChairId: chairId}
		if info, ok := self.canOperInfo[chairId]; ok {
			self.countCanOperMsg(info, canOper)
			recInfo.CanOper = canOper
		}
		recInfo.TotalNum = int32(len(self.baseCard))
		recInfo.LeftOperTime = 15
		recInfo.DrawDiceValue = switchToCyint32(self.drawDiceValue[:])
		recInfo.LastOutChair = self.lastOutChair
		recInfo.CurOutChair = self.curOutChair

		recInfo.LastOutCard = self.lastOutCard
		recInfo.LeftNum = int32(len(self.leftCard))
		if chairId != -1 {
			//发送听牌信息
			if self.curOutChair != -1 {
				if listenMsg, ok := self.operAction.GetListenInfo(chairId, self.players, nil, self.leftCard); ok {
					recInfo.ListenResult = listenMsg.ListenResult
				}
			} else {
				if listenMsg, ok := self.operAction.GetListenInfo2(chairId, self.players, nil, self.leftCard); ok {
					recInfo.ListenResult = listenMsg.ListenResult
				}
			}
			canNotOut := make([]int32, 0, len(self.players[chairId].CardInfo.CanNotOut))
			for _, v := range self.players[chairId].CardInfo.CanNotOut {
				canNotOut = append(canNotOut, v)
			}
			recInfo.UnableOutCards = switchToCyint32(canNotOut)
		}
		for _, v := range recInfo.GameUser {
			k := v.ChairId
			userInfo := self.players[k]
			v.CardNum = int32(len(userInfo.CardInfo.HandCards))
			v.Point = userInfo.BalanceResult.Point
			v.RecordCards = userInfo.CardInfo.RiverCards
			tmp := &pbgame_logic.Json_PlayerCard{}
			tmp.HuaCards = userInfo.CardInfo.HuaCards
			tmp.OutCards = userInfo.CardInfo.OutCards
			if k == chairId {
				tmp.HandCards = userInfo.CardInfo.HandCards
			}
			v.JsonCardInfo = util.PB2JSON(tmp, false)
		}
	case pbgame_logic.GameStatus_GSGameEnd: //游戏结束等待下局中
		for _, v := range recInfo.GameUser {
			k := v.ChairId
			userInfo := self.players[k]
			v.Point = userInfo.BalanceResult.Point
		}
	}
}

func (self *GameSink) logHeadUser(chairId int32) string {
	if chairId == -1 {
		return fmt.Sprintf("房间[%d] :", self.desk.deskId)
	} else {
		return fmt.Sprintf("房间[%d] 玩家[%s,%d]:", self.desk.deskId, self.players[chairId].BaseInfo.Nickname, chairId)
	}
}

//吃碰后不能出的牌
func (self *GameSink) addCanNotOut(chairId, card int32, chiType uint32) {
	if isFlag(chiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)) { //左吃
		if mj.IsVaildCard(card + 3) {
			self.players[chairId].CardInfo.CanNotOut[card+3] = card + 3
		}
	} else if isFlag(chiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)) { //右吃
		if mj.IsVaildCard(card - 3) {
			self.players[chairId].CardInfo.CanNotOut[card-3] = card - 3
		}
	}
	self.players[chairId].CardInfo.CanNotOut[card] = card
}

// func (self *GameSink) popLeftCard(chairId int32) (card int32) {
// 	if self.wantCards[chairId] != nil { //如果玩家要牌,优先发要的牌
// 		var index int //要牌索引
// 		for _, wcard := range self.wantCards[chairId] {
// 			index++
// 			if leftCards, ok := mj.RemoveCard(self.leftCard, wcard, false); ok {
// 				card = wcard
// 				self.leftCard = leftCards
// 				break
// 			}
// 		}
// 		//删除这张要过的牌
// 		if index >= len(self.wantCards[chairId]) {
// 			self.wantCards[chairId] = nil
// 		} else {
// 			self.wantCards[chairId] = self.wantCards[chairId][index:]
// 		}
// 	}
// 	if card == 0 {
// 		card = self.leftCard[len(self.leftCard)-1]
// 		self.leftCard = self.leftCard[:len(self.leftCard)-1]
// 	}
// 	return
// }

func (self *GameSink) doWantCards(chairId int32, cards []int32) (errMsg string) {
	var leftCardsStack map[int32]int32
	if self.isPlaying { //游戏中要牌
		leftCardsStack = mj.CalStackCards(self.leftCard, false)
	} else { //游戏开始前配牌
		leftCardsStack = mj.CalStackCards(self.baseCard, false)
	}
	cardsStack := mj.CalStackCards(cards, false)
	log.Debugf("%s 玩家要牌,cards=%v", self.logHeadUser(chairId), cards)
	for _, v := range cards {
		if !mj.IsVaildCard(v) {
			errMsg = fmt.Sprintf("指定的牌%v不合法,要牌失败", v)
			return
		}
		//检验每张牌,牌库是否有剩余的
		if leftCardsStack[v] < cardsStack[v] {
			errMsg = fmt.Sprintf("指定的牌%v剩余数量为%d,指定数量为%d,要牌失败", v, leftCardsStack[v], cardsStack[v])
			return
		}
	}
	if self.isPlaying { //已经发过牌
		//调整牌库的顺序
		tmpLeftCards := mj.DelCards(cardsStack, cards, self.leftCard)
		self.leftCard = append(tmpLeftCards, mj.ReversaCards(cards)...)
		tlog.Debug("已经发过牌要牌")
	} else { //没发过牌
		tlog.Debug("没发过牌要牌")
		cardDef.DebugCardsFromClient(gameName, cards)
	}

	return ""
}

//准备下一局
func (self *GameSink) getReady(uid uint64) {
	if self.desk.deskPlayers[uid].userStatus == pbgame.UserDeskStatus_UDSSitDown {
		log.Debugf("%s 准备下一局时重复准备", self.logHeadUser(self.desk.GetChairidByUid(uid)))
		return
	}
	// if self.readyInfo[chairId] {
	// }
	self.desk.changUserState(uid, pbgame.UserDeskStatus_UDSSitDown)

	// self.readyInfo[chairId] = true
	self.sendData(-1, &pbgame_logic.BS2CGetReady{UserId: uid})
	var readyNum int32 = 0
	for _, v := range self.desk.playChair {
		if v.userStatus == pbgame.UserDeskStatus_UDSSitDown {
			readyNum++
		}
	}
	if readyNum == 1 {
		self.changGameState(pbgame_logic.GameStatus_GSWait)
	} else if readyNum == self.game_config.PlayerCount {
		self.desk.changUserState(0, pbgame.UserDeskStatus_UDSPlaying)
		self.StartGame()
	}
}

// 查询战绩
func (self *GameSink) getGameRecord(chairId int32) {
	self.sendData(chairId, self.record.GetGameRecord())
}
