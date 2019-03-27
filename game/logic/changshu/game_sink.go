package main

import (
	mj "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
)

var (
	cardDef mj.CardDef //牌定义
	huLib   mj.HuLib   //胡牌算法
)

//游戏公共信息
type gameAllInfo struct {
	waitHigestOper    *OperPriority                     //当前等待中的最高优先级的操作
	operOrder         map[PriorityOrder][]*OperPriority //操作优先级
	canOperInfo       map[int32]*CanOperInfo            //玩家能做的操作
	game_balance_info gameBalanceInfo                   //游戏结束信息
	diceResult        [4][2]int32                       //投色子结果
	banker_id         int32                             //庄家id
	leftCard          []int32                           //发完牌后剩余的牌
	curThrowDice      int32                             //当前投色子的玩家
	isdeuce           bool                              //是否流局
	curOutChair       int32                             //当前出牌玩家
	lastOutChair      int32                             //上次出牌玩家
	lastOutCard       int32                             //上次出的牌
	makeCards         bool                              //是否做牌
	debugCard         []int32                           //配牌
	laiziCard         map[int32]int32                   //癞子牌
	hasHu             bool                              //是否胡牌
}

//游戏结束信息
type gameBalanceInfo struct {
}

type mjLib struct {
	operAction OperAtion       //操作
	record     mj.GameRecord   //游戏回放
	players    []mj.PlayerInfo //玩家游戏信息
}

//游戏主逻辑
type GameSink struct {
	mjLib
	gameAllInfo  //游戏公共信息
	desk         *Desk
	game_config  *pbgame_logic.CreateArg //游戏参数
	onlinePlayer []bool                  //在线玩家
	baseCard     []int32                 //基础牌库
	isPlaying    bool                    //是否在游戏中
}

func init() {
	cardDef.Init(log)
}

////////////////////////调用desk接口函数START/////////////////////////////
//发送消息给玩家(chairId为-1时发送给所有玩家)
func (self *GameSink) sendData(chairId int32, msg proto.Message) {
	if -1 == chairId {
		self.desk.SendGameMessage(0, msg)
	} else {
		self.desk.SendGameMessage(self.desk.GetUidByChairid(chairId), msg)
	}
}

////////////////////////调用desk接口函数END/////////////////////////////

//构建游戏
func (self *GameSink) Ctor(config *pbgame_logic.CreateArg) error {
	self.game_config = config
	self.isPlaying = false
	self.onlinePlayer = make([]bool, config.PlayerCount)
	self.players = make([]mj.PlayerInfo, config.PlayerCount)
	self.baseCard = cardDef.GetBaseCard(config.PlayerCount)
	self.reset()
	self.operAction.Init(config, self.laiziCard)
	return nil
}

//重置游戏
func (self *GameSink) reset() {
	//all_info
	self.gameAllInfo = gameAllInfo{}
	self.laiziCard = map[int32]int32{}
}

//开始游戏
func (self *GameSink) StartGame() {
	self.isPlaying = true
	//通知第一个玩家投色子
	self.sendData(-1, &pbgame_logic.S2CThrowDice{ChairId: 0})
	self.curThrowDice = 0
}

//玩家加入游戏
func (self *GameSink) AddPlayer(chairId int32, uid uint64, nickName string) bool {
	if self.game_config.PlayerCount <= chairId {
		log.Errorf("%s 加入房间失败,人数已满,游戏开始人数为%d", self.logHeadUser(chairId), self.game_config.PlayerCount)
		return false
	}

	info := mj.PlayerInfo{BaseInfo: mj.PlayerBaseInfo{ChairId: chairId, Uid: uid, Nickname: nickName, Point: 0}}
	self.players[chairId] = info
	self.onlinePlayer[chairId] = true
	return true
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
	msg := &pbgame_logic.S2CThrowDiceResult{ChairId: chairId}
	rand.Seed(time.Now().Unix())
	msg.DiceValue = make([]*pbgame_logic.Cyint32, 2)
	for i := 0; i < 2; i++ {
		rnd := int32(rand.Intn(5) + 1)
		msg.DiceValue[i] = &pbgame_logic.Cyint32{T: rnd}
		self.diceResult[chairId][i] = rnd
	}
	//广播色子结果
	self.sendData(-1, msg)
	//判断是否所有人都投色子
	for i := int32(0); i < self.game_config.PlayerCount; i++ {
		if self.diceResult[i][0] == 0 {
			//通知下一个玩家投色子
			self.sendData(-1, &pbgame_logic.S2CThrowDice{ChairId: i})
			return
		}
	}
	self.dealDiceResult()
}

//处理投色子结果
func (self *GameSink) dealDiceResult() {
	diceRes := make([]struct {
		dice    int32
		chairId int32
	}, self.game_config.PlayerCount)
	for i := 0; i < len(self.diceResult); i++ {
		diceRes[i].dice = self.diceResult[i][0] + self.diceResult[i][1]
		diceRes[i].chairId = int32(i)
	}
	//排序，实现比较方法即可
	sort.Slice(diceRes, func(i, j int) bool {
		if diceRes[i].dice == diceRes[j].dice {
			return diceRes[i].chairId < diceRes[j].chairId
		}
		return diceRes[i].dice > diceRes[j].dice
	})
	//发送换座位结果
	posInfo := make([]*pbgame_logic.ChangePosInfo, len(diceRes))
	for i, res := range diceRes {
		posInfo[i] = &pbgame_logic.ChangePosInfo{UserPos: res.chairId, DiceValue: res.dice}
	}
	//记录庄家
	self.banker_id = diceRes[0].chairId
	msg := &pbgame_logic.S2CChangePos{PosInfo: posInfo}
	self.sendData(-1, msg)
	//1s后发送游戏开始消息
	self.desk.set_timer(mj.TID_DealCard, 4*time.Second, func() {
		self.deal_card()
	})
}

//开始发牌
func (self *GameSink) deal_card() {
	msg := &pbgame_logic.S2CStartGame{BankerId: self.banker_id, CurInning: int32(self.desk.curInning)}
	msg.TotalCardNum = int32(len(self.baseCard))
	//洗牌
	self.shuffle_cards()
	var player_cards [][]int32
	player_cards, self.leftCard = cardDef.DealCard(self.leftCard, self.game_config.PlayerCount, self.banker_id)
	msg.LeftCardNum = int32(len(self.leftCard))
	for k, v := range self.players {
		v.CardInfo.HandCards = player_cards[k]
		msg.UserInfo = &pbgame_logic.StartGameInfo{HandCard: player_cards[k]}
		log.Warnf("%s手牌为:%v", self.logHeadUser(int32(k)), player_cards[k])
		//统计每个玩家手牌数量
		v.CardInfo.StackCards = cardDef.StackCards(player_cards[k])
		//给每个玩家发送游戏开始消息
		self.sendData(int32(k), msg)
	}
	//庄家开始第一次补花
	self.firstBuHua(self.banker_id)
	//检查庄家能否胡
	if ok, huTypeList := huLib.CheckHuType(&(self.players[self.banker_id].CardInfo)); ok {
		self.canOperInfo[self.banker_id] = &CanOperInfo{CanHu: CanHuOper{HuList: huTypeList}}
	}
	//检查能否杠
}

//玩家第一次补花
func (self *GameSink) firstBuHua(chairId int32) {
	cardInfo := self.players[chairId].CardInfo
	leftCard := self.leftCard
	huaIndex := make(map[int32]int32) //下次要补的花牌
	leftCount := len(leftCard)

	msg := pbgame_logic.S2CBuHua{ChairId: chairId}      //发送给补花玩家
	msgOther := pbgame_logic.S2CBuHua{ChairId: chairId} //发送给其他玩家

	//补一张花牌
	operOnce := func(card int32) int32 {
		//减一张花牌
		mj.Sub_stack(cardInfo.StackCards, card)
		cardInfo.StackCards[card]--
		//从牌库摸一张牌
		moCard := leftCard[leftCount-1]
		leftCount--
		//摸的牌加到手牌
		mj.Add_stack(cardInfo.StackCards, moCard)
		//记录到消息
		msg.BuHuaResult = append(msg.BuHuaResult, &pbgame_logic.BuHuaOnce{HuaCard: card, BuCard: moCard})
		msgOther.BuHuaResult = append(msgOther.BuHuaResult, &pbgame_logic.BuHuaOnce{HuaCard: card, BuCard: 0})
		if cardDef.IsHuaCard(moCard) {
			mj.Add_stack(huaIndex, moCard)
		}
		return moCard
	}

	//先遍历一次所有花牌
	for huaCard := int32(51); huaCard <= 59; huaCard++ {
		//遇到一张花牌,补一张
		if huaCount, ok := cardInfo.StackCards[huaCard]; ok {
			for j := int32(0); j < huaCount; j++ {
				operOnce(huaCard)
			}
		}
	}

	//再从第一次结果补花
	if len(huaIndex) > 0 {
		bFin := true //补花结束
		num := 0
		for {
			//遍历所有
			for huaCard, huaCount := range huaIndex {
				for j := int32(0); j < huaCount; j++ {
					if cardDef.IsHuaCard(operOnce(huaCard)) {
						bFin = false
					}
					//补一张减一张
					mj.Sub_stack(huaIndex, huaCard)
				}
			}
			if bFin {
				break
			}
			num++
			if num >= 12 {
				log.Errorf("补花死循环")
				return
			}
		}
	}
	for i := int32(0); i < self.game_config.PlayerCount; i++ {
		if i == chairId {
			self.sendData(i, &msg)
		} else {
			self.sendData(i, &msgOther)
		}
	}
}

func (self *GameSink) resetOper() {
	self.canOperInfo = map[int32]*CanOperInfo{}
	self.waitHigestOper = nil
	self.operOrder = map[PriorityOrder][]OperPriority{}
}

//摸牌 last(-1摸最后一张 1第一次摸牌 0正常摸牌),lose_chair在明杠时为放杠玩家id,包赔
func (self *GameSink) drawCard(chairId, last, lose_chair int32) error {
	//检查游戏是否结束
	if len(self.leftCard) <= 0 {
		self.isdeuce = true
		self.gameEnd()
		return nil
	}

	self.resetOper()

	var card, index int32 = 0, 0
	if last == -1 { //杠后摸最后一张牌
		index = int32(len(self.leftCard)) - 1
	}
	card = self.leftCard[index]
	self.leftCard = self.leftCard[:index]
	log.Debugf("%s 摸牌[%d]剩余[%d]张", self.logHeadUser(chairId), card, index)

	//发送摸牌
	for i := int32(0); i < self.game_config.PlayerCount; i++ {
		if i == chairId {
			self.sendData(i, &pbgame_logic.S2CDrawCard{ChairId: chairId, Card: card, LeftNum: int32(len(self.leftCard))})
		} else {
			self.sendData(i, &pbgame_logic.S2CDrawCard{ChairId: chairId, Card: 0, LeftNum: int32(len(self.leftCard))})
		}
	}
	self.curOutChair = chairId

	cardInfo := &self.players[chairId].CardInfo
	//分析能否暗杠,补杠,自摸胡
	ret := self.operAction.DrawcardAnalysis(cardInfo, card, int32(len(self.leftCard)))
	fmt.Printf("%s 摸牌后操作分析ret=%+v", self.logHeadUser(chairId), ret)
	//发送倒计时玩家
	self.sendData(-1, &pbgame_logic.S2CTimeoutChair{ChairId: chairId, Time: 15})
	//更新玩家card_info表
	self.operAction.HandleDrawCard(cardInfo, card)
	//游戏回放记录
	//统计能做的操作
	if !ret.Empty() {
		if last == -1 && !ret.CanHu.Empty() {
			//杠上开花
		}
		msg := &pbgame_logic.S2CHaveOperation{Card: card}
		self.countCanOper(ret, chairId, mj.HuMode_ZIMO, -1, card, -1, msg)
		//发送玩家可进行的操作
		self.sendData(chairId, msg)
	}
	return nil
}

//出牌
func (self *GameSink) outCard(chairId, card int32) error {
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
	if _, ok := self.canOperInfo[chairId]; ok {
		log.Errorf("%s 出牌失败,还有其他操作，先取消", self.logHeadUser(chairId))
		return nil
	}
	cardInfo := &self.players[chairId].CardInfo
	//判断是否有这张牌
	if _, ok := cardInfo.StackCards[card]; !ok {
		log.Errorf("%s 出牌失败,手上没有这张牌", self.logHeadUser(chairId))
		return nil
	}

	//更新玩家card_info表
	self.operAction.HandleOutCard(cardInfo, card)

	self.sendData(-1, &pbgame_logic.S2COutCard{ChairId: chairId, Card: card})

	//游戏回放记录

	self.lastOutChair, self.lastOutCard, self.curOutChair = chairId, card, 0

	//检查出牌后能做的操作
	willWait := false
	for k, v := range self.players {
		if int32(k) != chairId {
			ret := self.operAction.OutCardAnalysis(&v.CardInfo, card, int32(k), chairId, int32(len(self.leftCard)))
			if !ret.Empty() {
				//统计并记录玩家可以进行的操作
				msg := &pbgame_logic.S2CHaveOperation{Card: card}
				self.countCanOper(ret, int32(k), mj.HuMode_PAOHU, chairId, card, chairId, msg)
				willWait = true
				//发送玩家可进行的操作
				log.Tracef("%s 可进行的操作%+v", self.logHeadUser(int32(k)), ret)
				self.sendData(chairId, msg)
			}
		}
	}
	if !willWait { //出牌后无其他人能操作
		self.drawCard(GetNextChair(chairId, self.game_config.PlayerCount), 0, -1)
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

//统计并记录玩家可以进行的操作
func (self *GameSink) countCanOper(ret *CanOperInfo, chairId int32, huMode mj.EmHuMode, loseChair, card, opChair int32, msg *pbgame_logic.S2CHaveOperation) {
	//记录能吃
	if !ret.CanChi.Empty() {
		self.canOperInfo[chairId].CanChi = CanChiOper{Card: card, ChairId: chairId, ChiType: ret.CanChi.ChiType}
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskChi) | msg.OperMask
		msg.CanChi = &pbgame_logic.CanChiMsg{ChiType: ret.CanChi.ChiType}
		self.addOperOrder(ChiOrder, &OperPriority{ChairId: chairId, Op: ChiOrder, Info: &self.canOperInfo[chairId].CanChi})
	}
	//记录能碰
	if !ret.CanPeng.Empty() {
		self.canOperInfo[chairId].CanPeng = CanPengOper{ChairId: chairId, LoseChair: loseChair, Card: card}
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskPeng) | msg.OperMask
		self.addOperOrder(PengOrder, &OperPriority{ChairId: chairId, Op: PengOrder, Info: &self.canOperInfo[chairId].CanPeng})
	}
	//记录能杠
	if !ret.CanGang.Empty() {
		self.canOperInfo[chairId].CanGang = CanGangOper{ChairId: chairId, GangList: ret.CanGang.GangList}
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskGang) | msg.OperMask
		msg.CanGang = &pbgame_logic.CanGangMsg{}
		for _, gangCard := range ret.CanGang.GangList {
			msg.CanGang.Cards = append(msg.CanGang.Cards, gangCard)
		}
		self.addOperOrder(GangOrder, &OperPriority{ChairId: chairId, Op: GangOrder, Info: &self.canOperInfo[chairId].CanGang})
	}
	//记录能胡
	if !ret.CanHu.Empty() {
		self.canOperInfo[chairId].CanHu = CanHuOper{HuMode: huMode, LoseChair: loseChair, Card: card, OpChair: opChair, HuList: ret.CanHu.HuList}
		msg.OperMask = uint32(pbgame_logic.CanOperMask_OperMaskHu) | msg.OperMask
		self.addOperOrder(HuOrder, &OperPriority{ChairId: chairId, Op: HuOrder, Info: &self.canOperInfo[chairId].CanHu})
	}
}

//洗牌
func (self *GameSink) shuffle_cards() {
	if self.makeCards {
		self.leftCard = self.debugCard
		return
	}
	self.leftCard = cardDef.RandCards(self.baseCard)
}

//返回1表示能直接进行该操作,返回2表示还需要等待,返回3表示需要唤醒等待中的操作
func (self *GameSink) checkPlayerOperationNeedWait(chairId int32, curOrder PriorityOrder) int {
	var otherOrder, waitOrder PriorityOrder = NoneOrder, NoneOrder

	//检查其他人能做的最高优先级操作
	for i := HuOrder; i >= ChiOrder; i-- {
		if curOper, ok := self.operOrder[i]; ok {
			for _, v := range curOper {
				if v.ChairId != chairId {
					otherOrder = i
					break
				}
			}
		}
	}
	if self.waitHigestOper != nil {
		waitOrder = self.waitHigestOper.Op
	}
	//比较当前操作,当前等待中的最高优先级的操作,和其他人能做的最高优先级操作
	return 0
}

//删除玩家所有能做的操作
func (self *GameSink) deletePlayerCanOper(chairId int32) {
	delete(self.canOperInfo, chairId)
	for i := HuOrder; i >= ChiOrder; i-- {
		if curOper, ok := self.operOrder[i]; ok {
			for index, v := range curOper {
				if v.ChairId == chairId { //每种操作最多只可能有一个
					curOper = append(curOper[:index], curOper[index+1:]...)
					break
				}
			}
		}
	}
}

//判断是否存在需要等待的操作,存在则返回false,不存在则返回true
func (self *GameSink) checkCanOperEmpty() bool {
	if len(self.operOrder[HuOrder]) > 0 { //还有人能胡
		return false
	}
	// if  {

	// }
	return true
}

//吃
func (self *GameSink) chiCard(chairId, card int32, chiType uint32) error {
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

	//校验吃牌
	if chiType == 0 || card != self.canOperInfo[chairId].CanChi.Card || chiType != (self.canOperInfo[chairId].CanChi.ChiType&chiType) {
		log.Errorf("%s 吃牌失败,没有该吃类型,或者吃的牌不对,CanChi=%+v", self.logHeadUser(chairId), self.canOperInfo[chairId].CanChi)
		return nil
	}

	if self.canOperInfo[chairId] != nil && !self.canOperInfo[chairId].CanHu.Empty() { //玩家能胡
		if self.hasHu { //已经有人胡牌
			log.Debugf("%s 吃牌时已经有人选择胡牌", self.logHeadUser(chairId))
			self.deletePlayerCanOper(chairId)
			if self.checkCanOperEmpty() {
				log.Debugf("%s 吃牌时已经有人选择胡牌,并且不需要再等待其他人操作,游戏结束", self.logHeadUser(chairId))
				self.gameEnd()
			}
			return nil
		} else { //没有其他人操作胡

		}
	}

	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, ChiOrder)
	if res == 1 {

	}
	return nil
}

//碰
func (self *GameSink) pengCard(chairId, card int32) error {

	return nil
}

//杠
func (self *GameSink) gangCard(chairId, card int32) error {
	return nil
}

//胡
func (self *GameSink) huCard(chairId int32) error {
	return nil
}

//游戏结束
func (self *GameSink) gameEnd() {

}

func (self *GameSink) logHeadUser(chairId int32) string {
	return fmt.Sprintf("房间[%d] 玩家[%s,%d]:", self.desk.id, self.players[chairId].BaseInfo.Nickname, chairId)
}
