package main

import (
	"cy/game/configs"
	mj "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"cy/game/util"
	"fmt"
	"math/rand"
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
	guoChiCards    map[int32][]int32                 //过吃的牌
	guoPengCards   map[int32]int32                   //过碰的牌
	louPeng        map[int32]bool                    //是否过碰
	wantCards      map[int32][]int32                 //玩家指定要的牌
}

type mjLib struct {
	gameBalance GameBalance      //游戏结束信息
	operAction  OperAtion        //操作
	record      mj.GameRecord    //游戏回放
	players     []*mj.PlayerInfo //玩家游戏信息
}

//游戏主逻辑
type GameSink struct {
	mjLib
	desk        *Desk                   //桌子
	gameAllInfo                         //游戏公共信息
	game_config *pbgame_logic.CreateArg //游戏参数
	baseCard    []int32                 //基础牌库
	isPlaying   bool                    //是否在游戏中
	// onlinePlayer []bool                  //在线玩家
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

//发送消息给其他人
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
	// self.onlinePlayer = make([]bool, config.PlayerCount)
	self.players = make([]*mj.PlayerInfo, config.PlayerCount)
	self.baseCard = cardDef.GetBaseCard(config.PlayerCount)
	self.operAction.Init(config, self.laiziCard)
	self.gameBalance.Init(config)
	return nil
}

//重置游戏
func (self *GameSink) reset() {
	//all_info
	self.gameAllInfo = gameAllInfo{}
	self.operOrder = map[PriorityOrder][]*OperPriority{}
	self.canOperInfo = map[int32]*CanOperInfo{}
	self.haswaitOper = make([]bool, self.game_config.PlayerCount)
	self.diceResult = make([][2]int32, self.game_config.PlayerCount)
	self.hasFirstBuHua = make([]bool, self.game_config.PlayerCount)
	self.bankerId = -1
	self.curThrowDice = -1
	self.curOutChair = -1
	self.lastOutChair = -1
	self.gameBalance.Reset()
	self.laiziCard = map[int32]int32{}
}

//开始游戏
func (self *GameSink) StartGame() {
	self.isPlaying = true
	self.reset()
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
	self.players[chairId] = mj.MakePlayers()
	self.players[chairId].BaseInfo = mj.PlayerBaseInfo{ChairId: chairId, Uid: uid, Nickname: nickName, Point: 0}
	// self.onlinePlayer[chairId] = true
	return true
}

//玩家退出游戏
func (self *GameSink) Exitlayer(chairId int32) bool {
	if int(chairId) >= len(self.players) {
		log.Error("Exitlayer 时int(chairId) >= len(self.players)")
		return false
	}
	self.players[chairId] = nil
	//self.players = append(self.players[:chairId], self.players[chairId+1:]...)
	// self.onlinePlayer[chairId] = false
	return true
}

//改变游戏状态
func (self *GameSink) changGameState(gState pbgame_logic.GameStatus) {
	self.desk.gameStatus = gState
	if gState > pbgame_logic.GameStatus_GSWait {
		self.sendData(-1, &pbgame_logic.BS2CUpdateGameStatus{GameStatus: gState})
	}
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
	// sort.Slice(diceRes, func(i, j int) bool {
	// 	if diceRes[i].dice == diceRes[j].dice {
	// 		return diceRes[i].oldChairId < diceRes[j].oldChairId
	// 	}
	// 	return diceRes[i].dice > diceRes[j].dice
	// })
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
		self.changGameState(pbgame_logic.GameStatus_GSPlaying)
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
	//随机2个色子,用于客户端选择从牌堆摸牌的方向
	msg := &pbgame_logic.S2CStartGame{BankerId: self.bankerId, CurInning: self.desk.curInning, LeftTime: 15}
	msg.DiceValue = make([]*pbgame_logic.Cyint32, 2)
	randRes := self.randDice()
	self.gameBalance.DealStartDice(randRes)
	for i, rnd := range randRes {
		msg.DiceValue[i] = &pbgame_logic.Cyint32{T: rnd}
	}

	//洗牌
	self.shuffle_cards()
	var player_cards [][]int32
	player_cards, self.leftCard = cardDef.DealCard(self.leftCard, self.game_config.PlayerCount, self.bankerId)

	//庄家手牌
	bankerCardInfo := &self.players[self.bankerId].CardInfo
	bankerCardInfo.HandCards = player_cards[self.bankerId]
	bankerCardInfo.StackCards = mj.CalStackCards(player_cards[self.bankerId])
	//庄家开始第一次补花
	huaCards, _ := self.firstBuHuaCards(self.bankerId)
	bankerCardInfo.HuaCards = append(bankerCardInfo.HuaCards, huaCards...)
	msg.HuaCards = switchToCyint32(huaCards)
	msg.LeftNum = int32(len(self.leftCard))
	msg.TotalNum = int32(len(self.baseCard))
	self.curOutChair = self.bankerId
	self.resetOper()
	//分析庄家能做的操作
	ret := self.operAction.BankerAnalysis(bankerCardInfo)
	//统计能做的操作
	if !ret.Empty() {
		msg.BankerOper = &pbgame_logic.S2CHaveOperation{}
		self.countCanOper(ret, self.bankerId, mj.HuMode_ZIMO, -1, bankerCardInfo.HandCards[13], -1, msg.BankerOper)
	}

	for k, v := range self.players {
		if int32(k) != self.bankerId {
			v.CardInfo.HandCards = player_cards[k]
			v.CardInfo.StackCards = mj.CalStackCards(player_cards[k])
		}

		tmp := &pbgame_logic.Json_UserCardInfo{HandCards: map[int32]*pbgame_logic.Json_UserCardInfoCards{}}
		tmp.HandCards[int32(k)] = &pbgame_logic.Json_UserCardInfoCards{Cards: v.CardInfo.HandCards}
		msg.JsonAllCards = util.PB2JSON(tmp, false)
		log.Warnf("%s手牌为:%v", self.logHeadUser(int32(k)), player_cards[k])
		//给每个玩家发送游戏开始消息
		self.sendData(int32(k), msg)
	}
	//游戏回放记录
}

func switchToCyint32(cards []int32) []*pbgame_logic.Cyint32 {
	res := []*pbgame_logic.Cyint32{}
	for _, card := range cards {
		res = append(res, &pbgame_logic.Cyint32{T: card})
	}
	return res
}

//玩家第一次补花,返回所有的花牌
// func (self *GameSink) firstBuHuaCards(chairId int32) []int32 {
// 	cardInfo := &self.players[chairId].CardInfo
// 	tlog.Debug("庄家补花前手牌数据为", zap.Any("cardInfo", cardInfo))
// 	leftCard := self.leftCard
// 	huaIndex := make(map[int32]int32) //下次要补的花牌

// 	huaCards := []int32{} //所有的花牌
// 	//补掉一张花牌
// 	operOnce := func(card int32) int32 {
// 		//减一张花牌
// 		self.operAction.updateCardInfo(cardInfo, nil, []int32{card})
// 		//从牌库摸一张牌
// 		moCard := leftCard[len(leftCard)-1]
// 		self.leftCard = self.leftCard[:len(leftCard)-1]
// 		tlog.Debug("补花", zap.Int32("huaCard", card), zap.Int32("moCard", moCard))
// 		//摸的牌加到手牌
// 		self.operAction.updateCardInfo(cardInfo, []int32{moCard}, nil)
// 		//记录到消息
// 		huaCards = append(huaCards, card)
// 		if mj.IsHuaCard(moCard) {
// 			mj.Add_stack(huaIndex, moCard)
// 		}
// 		return moCard
// 	}

// 	//先遍历一次所有花牌
// 	for huaCard := int32(51); huaCard <= 59; huaCard++ {
// 		//遇到一张花牌,补一张
// 		if huaCount, ok := cardInfo.StackCards[huaCard]; ok {
// 			for j := int32(0); j < huaCount; j++ {
// 				operOnce(huaCard)
// 			}
// 		}
// 	}

// 	//再从第一次结果补花
// 	if len(huaIndex) > 0 {
// 		bFin := true //补花结束
// 		num := 0
// 		for {
// 			//遍历所有
// 			for huaCard, huaCount := range huaIndex {
// 				for j := int32(0); j < huaCount; j++ {
// 					if mj.IsHuaCard(operOnce(huaCard)) {
// 						bFin = false
// 					}
// 					//补一张减一张
// 					mj.Sub_stack(huaIndex, huaCard)
// 				}
// 			}
// 			if bFin {
// 				break
// 			}
// 			num++
// 			if num >= 12 {
// 				log.Errorf("补花死循环")
// 				return nil
// 			}
// 		}
// 	}
// 	tlog.Debug("庄家补花后手牌数据为", zap.Any("cardInfo", cardInfo))
// 	return huaCards
// }

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
			tmpHuaCards, moCard := self.drawOneCard(chairId)
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
func (self *GameSink) drawOneCard(chairId int32) (huaCards, moCard []int32) {
	var num int32
	for {
		card := self.popLeftCard(chairId)
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

//摸牌 last(-1摸最后一张 1第一次摸牌 0正常摸牌),lose_chair在明杠时为放杠玩家id,包赔
func (self *GameSink) drawCard(chairId, last, lose_chair int32) error {
	log.Debugf("%s,摸牌操作,last=%d, lose_chair=%d", self.logHeadUser(chairId), last, lose_chair)
	//检查游戏是否结束
	if len(self.leftCard) <= 0 {
		self.gameEnd()
		return nil
	}

	self.resetOper()

	msg := &pbgame_logic.BS2CDrawCard{ChairId: chairId}
	var huaCards, moCards []int32
	//游戏中摸牌
	if self.hasFirstBuHua[chairId] {
		huaCards, moCards = self.drawOneCard(chairId)
		log.Debugf("%s 游戏中摸牌moCards=%v,huaCards=%v,剩余[%d]张", self.logHeadUser(chairId), moCards, huaCards, len(self.leftCard))
	} else { //第一次摸牌
		if mj.GetHuaCount(self.players[chairId].CardInfo.StackCards) > 0 {
			huaCards, moCards = self.firstBuHuaCards(chairId)
			//删掉一张牌
			self.operAction.updateCardInfo(&self.players[chairId].CardInfo, nil, moCards[len(moCards)-1:])
		} else {
			huaCards, moCards = self.drawOneCard(chairId)
			self.hasFirstBuHua[chairId] = true
		}
		log.Debugf("%s 第一次摸牌moCards=%v,huaCards=%v,剩余[%d]张", self.logHeadUser(chairId), moCards, huaCards, len(self.leftCard))
	}
	msg.LeftNum = int32(len(self.leftCard))
	card := moCards[len(moCards)-1] //最后摸到的牌
	//发送自己
	msg.JsonDrawInfo = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: moCards}, false)
	self.sendData(chairId, msg)
	//发送别人
	msg.JsonDrawInfo = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: make([]int32, len(moCards))}, false)
	self.sendDataOther(chairId, msg)

	cardInfo := &self.players[chairId].CardInfo
	cardInfo.HuaCards = append(cardInfo.HuaCards, huaCards...)

	self.curOutChair = chairId

	//分析能否暗杠,补杠,自摸胡
	ret := self.operAction.DrawcardAnalysis(cardInfo, card, int32(len(self.leftCard)))
	log.Infof("%s 摸牌后操作分析ret=%+v", self.logHeadUser(chairId), ret)
	//发送倒计时玩家
	// self.sendData(-1, &pbgame_logic.BS2CCurOutChair{ChairId: chairId})
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

//检查吃碰后能做的操作
func (self *GameSink) checkAfterChiPeng(chairId int32) {
	if !self.hasFirstBuHua[chairId] { //第一次补花
		huaCards, moCards := self.firstBuHuaCards(chairId)
		if len(huaCards) > 0 {
			self.players[chairId].CardInfo.HuaCards = append(self.players[chairId].CardInfo.HuaCards, huaCards...)
			msg := &pbgame_logic.BS2CFirstBuHua{ChairId: chairId, LeftNum: int32(len(self.leftCard))}
			//发给自己
			msg.JsonFirstBuhua = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: moCards}, false)
			self.sendData(chairId, msg)
			//发给别人
			msg.JsonFirstBuhua = util.PB2JSON(&pbgame_logic.Json_FirstBuHua{HuaCards: huaCards, MoCards: make([]int32, len(moCards))}, false)
			self.sendDataOther(chairId, msg)
		}
	}

	cardInfo := &self.players[chairId].CardInfo
	//分析能否暗杠,补杠
	ret := self.operAction.AfterChiPengAnalysis(cardInfo)
	log.Infof("%s 吃碰后操作分析ret=%+v", self.logHeadUser(chairId), ret)
	//发送倒计时玩家
	// self.sendData(-1, &pbgame_logic.BS2CCurOutChair{ChairId: chairId})
	//游戏回放记录
	//统计能做的操作
	if !ret.Empty() {
		msg := &pbgame_logic.S2CHaveOperation{}
		self.countCanOper(ret, chairId, mj.HuMode_ZIMO, -1, 0, -1, msg)
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
	self.sendData(-1, &pbgame_logic.BS2COutCard{ChairId: chairId, Card: card})
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
				log.Infof("%s 可进行的操作%+v", self.logHeadUser(int32(k)), ret)
				self.sendData(int32(k), msg)
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
	if *release && configs.Conf.GameNode[gameName].GameTest != "" {
		self.leftCard = cardDef.DebugCards(gameName, self.baseCard, self.game_config.PlayerCount)
		return
	}
	self.leftCard = mj.RandCards(self.baseCard)
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
	switch info := self.waitHigestOper.Info.(type) {
	case *CanChiOper:
		log.Debugf("%s 唤醒操作吃", self.logHeadUser(chairId))
		self.chiCard(info.ChairId, info.Card, info.ChiType)
	case *CanPengOper:
		log.Debugf("%s 唤醒操作碰", self.logHeadUser(chairId))
	case *CanGangOper:
		log.Debugf("%s 唤醒操作杠", self.logHeadUser(chairId))

	case *CanHuOper:
		log.Debugf("%s 唤醒操作胡", self.logHeadUser(chairId))

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
		self.insertWaitOper(chairId, ChiOrder, &self.canOperInfo[chairId].CanChi)
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
		self.gameEnd()
		return nil
	}
	//更新玩家card_info表
	self.operAction.HandleChiCard(&self.players[chairId].CardInfo, &self.players[self.lastOutChair].CardInfo, card, chiType)
	//回放记录

	// self.sendData(-1, &pbgame_logic.BS2CCurOutChair{ChairId: chairId})
	//变量维护
	self.AddGuoChiCards(chairId, card, chiType)
	self.curOutChair = chairId
	self.haswaitOper[chairId] = false
	self.resetOper()

	self.sendData(-1, &pbgame_logic.BS2CChiCard{ChairId: chairId, Card: card, ChiType: chiType})

	self.checkAfterChiPeng(chairId)
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
		self.insertWaitOper(chairId, PengOrder, &self.canOperInfo[chairId].CanPeng)
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
		self.gameEnd()
		return nil
	}
	//更新玩家card_info表
	self.operAction.HandlePengCard(&self.players[chairId].CardInfo, &self.players[self.lastOutChair].CardInfo, card, self.canOperInfo[chairId].CanPeng.LoseChair)
	//回放记录

	// self.sendData(-1, &pbgame_logic.BS2CCurOutChair{ChairId: chairId})
	self.sendData(-1, &pbgame_logic.BS2CPengCard{ChairId: chairId, Card: card})
	//变量维护
	self.curOutChair = chairId
	self.haswaitOper[chairId] = false
	self.resetOper()

	self.checkAfterChiPeng(chairId)
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
	if card != self.canOperInfo[chairId].CanGang.GangList[card] {
		log.Errorf("%s 杠牌失败,牌不对,card=%d,CanGang=%+v", self.logHeadUser(chairId), card, self.canOperInfo[chairId].CanGang)
		return nil
	}
	self.deletePlayerOperOrder(chairId)
	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, GangOrder)
	if res == 2 { //需要等待其他人操作
		log.Debugf("%s 操作杠需要等待其他人", self.logHeadUser(chairId))
		self.insertWaitOper(chairId, GangOrder, &self.canOperInfo[chairId].CanGang)
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
		self.gameEnd()
		return nil
	}

	//变量维护
	self.resetOper() //先清除操作,以免影响抢杠胡判断
	self.haswaitOper[chairId] = false

	self.curOutChair = -1 //玩家杠牌后 当前出牌玩家还不是自己 要摸牌后才能出牌
	gangType := self.operAction.GetGangType(&self.players[chairId].CardInfo, card)

	var loseChair int32 = -1             //如果是补杠,为碰牌时被碰玩家
	if gangType == mj.OperType_BU_GANG { //补杠
		log.Debugf("%s 补杠,杠牌为%d", self.logHeadUser(chairId), card)
		loseChair = self.players[chairId].CardInfo.PengCards[card]
	} else if gangType == mj.OperType_MING_GANG { //明杠
		log.Debugf("%s 明杠,杠牌为%d", self.logHeadUser(chairId), card)
		loseChair = self.lastOutChair
	} else if gangType == mj.OperType_AN_GANG { //暗杠
		log.Debugf("%s 暗杠,杠牌为%d", self.logHeadUser(chairId), card)
	}

	willWait := false
	//判断抢杠胡
	if gangType == mj.OperType_BU_GANG {
		for k, v := range self.players {
			if int32(k) != chairId {
				ret := self.operAction.QiangGangAnalysis(&v.CardInfo, card, int32(k), int32(chairId))
				if !ret.Empty() {
					//统计并记录玩家可以进行的操作
					msg := &pbgame_logic.S2CHaveOperation{Card: card}
					self.countCanOper(ret, int32(k), mj.HuMode_QIANGHU, chairId, card, chairId, msg)
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
		return nil
	}

	self.sendData(-1, &pbgame_logic.BS2CGangCard{ChairId: chairId, Card: card, Type: pbgame_logic.GangType(gangType), LoseChair: loseChair})

	//更新玩家card_info表
	if gangType == mj.OperType_MING_GANG {
		self.operAction.HandleGangCard(&self.players[chairId].CardInfo, &self.players[self.lastOutChair].CardInfo, card, gangType)
	} else {
		self.operAction.HandleGangCard(&self.players[chairId].CardInfo, nil, card, gangType)
	}
	//回放记录

	//变量维护
	self.haswaitOper[chairId] = false
	self.resetOper()

	self.afterGangCard(chairId, card, gangType)
	return nil
}

//杠之后,计算杠分,统计杠的次数,摸牌
func (self *GameSink) afterGangCard(chairId, card int32, gangType mj.EmOperType) {
	if gangType == mj.OperType_BU_GANG { //补杠
		self.gameBalance.AddScoreTimes(&self.players[chairId].BalanceResult, mj.ScoreTimes_BuGang)
		self.gameBalance.CalGangScore(chairId, -1, gangType)
		self.drawCard(chairId, -1, -1)
	} else if gangType == mj.OperType_MING_GANG { //明杠
		self.gameBalance.AddScoreTimes(&self.players[chairId].BalanceResult, mj.ScoreTimes_MingGang)
		self.gameBalance.CalGangScore(chairId, -1, gangType)
		self.drawCard(chairId, -1, -1)
	} else if gangType == mj.OperType_AN_GANG { //暗杠
		self.gameBalance.AddScoreTimes(&self.players[chairId].BalanceResult, mj.ScoreTimes_AnGang)
		self.gameBalance.CalGangScore(chairId, -1, gangType)
		self.drawCard(chairId, -1, self.lastOutChair)
	}
}

//胡
func (self *GameSink) huCard(chairId int32) error {
	log.Debugf("%s,胡牌操作,card=%d", self.logHeadUser(chairId))
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
		self.insertWaitOper(chairId, HuOrder, &self.canOperInfo[chairId].CanGang)
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
	if huInfo.LoseChair != -1 && huInfo.HuMode != mj.HuMode_QIANGHU {
		self.operAction.updateCardInfo(cardInfo, []int32{huInfo.Card}, nil)
	}

	//记录胡牌牌型
	self.gameBalance.loseChair = huInfo.LoseChair
	self.gameBalance.huCard = huInfo.Card
	self.gameBalance.huMode = huInfo.HuMode
	//判断附属胡牌类型
	huTypeExtra := []mj.EmExtraHuType{}
	if huInfo.HuMode == mj.HuMode_QIANGHU { //抢杠胡
		huTypeExtra = append(huTypeExtra, mj.ExtraHuType_QiangGang)
		if huInfo.LoseChair != -1 {
			self.operAction.updateCardInfo(&self.players[huInfo.LoseChair].CardInfo, nil, []int32{huInfo.Card})
		}
	} else if self.gameBalance.gangHuaChair == chairId { //杠上花
		huTypeExtra = append(huTypeExtra, mj.ExtraHuType_GangShangHua)
	}

	if huInfo.HuMode == mj.HuMode_PAOHU && self.gameBalance.gangPaoHu { //杠上炮
		huTypeExtra = append(huTypeExtra, mj.ExtraHuType_GangShangPao)
	}
	self.gameBalance.huChairs[chairId] = &HuScoreInfo{HuTypeList: huInfo.HuList, HuTypeExtra: huTypeExtra}

	//统计总结算次数
	if huInfo.LoseChair == -1 {
		self.gameBalance.AddScoreTimes(&self.players[chairId].BalanceResult, mj.ScoreTimes_ZiMo)
	} else {
		self.gameBalance.AddScoreTimes(&self.players[chairId].BalanceResult, mj.ScoreTimes_JiePao)
		self.gameBalance.AddScoreTimes(&self.players[huInfo.LoseChair].BalanceResult, mj.ScoreTimes_JiePao)
	}

	self.sendData(-1, &pbgame_logic.BS2CHuCard{ChairId: chairId, HandCards: cardInfo.HandCards})
	//变量维护
	self.haswaitOper[chairId] = false
	self.deletePlayerOperOrder(chairId)

	if len(self.operOrder[HuOrder]) == 0 {
		self.gameBalance.lastHuChair = chairId
		self.gameEnd()
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

	//检查玩家当前操作是否需要等待
	res := self.checkPlayerOperationNeedWait(chairId, NoneOrder)
	if res == 2 { //需要等待其他人操作
		log.Debugf("%s 取消操作需要等待其他人", self.logHeadUser(chairId))
		// self.insertWaitOper(chairId, GangOrder, &self.canOperInfo[chairId].CanGang)
		// self.haswaitOper[chairId] = true
		return nil
	} else if res == 3 { //唤醒等待中的操作
		log.Debugf("%s 取消操作,唤醒等待中的操作", self.logHeadUser(chairId))
		self.dealWaitOper(chairId)
		return nil
	}
	//判断是否已经胡
	if self.hasHu {
		log.Debugf("%s 取消操作,因为已经有人胡牌,游戏结束", self.logHeadUser(chairId))
		self.gameEnd()
		return nil
	}

	//回放记录

	//变量维护
	self.haswaitOper[chairId] = false
	self.resetOper()

	//取消操作后,由上次出牌玩家下家抓牌
	self.drawCard(GetNextChair(self.lastOutChair, self.game_config.PlayerCount), 0, -1)
	return nil
}

//游戏结束
func (self *GameSink) gameEnd() {
	log.Debugf("%s 第%d局游戏结束", self.logHeadUser(self.gameBalance.lastHuChair), self.desk.curInning)
	self.isPlaying = false
	self.dealGameBalance()

	msg := &pbgame_logic.BS2CGameEnd{CurInning: self.desk.curInning, Banker: self.bankerId, Isdeuce: self.gameBalance.lastHuChair == -1}
	msg.PlayerBalance = []*pbgame_logic.PlayerBalanceInfo{}
	for _, v := range self.players {
		Info := &pbgame_logic.PlayerBalanceInfo{}
		Info.HandCards = v.CardInfo.HandCards
		Info.HuCard = v.BalanceInfo.HuCard
		Info.Point = v.BalanceInfo.HuPoint
	}

	//游戏记录
	self.sendData(-1, msg)
	self.afterGameEnd()
}

//处理算分
func (self *GameSink) dealGameBalance() {

}

//小局结束后数据清理
func (self *GameSink) afterGameEnd() {

}

//断线重连
func (self *GameSink) gameReconnect(recInfo *pbgame_logic.GameDeskInfo, uid uint64) {
	switch recInfo.GameStatus {
	case pbgame_logic.GameStatus_GSWait:
	default:
	}
}

func (self *GameSink) logHeadUser(chairId int32) string {
	if chairId == -1 {
		return fmt.Sprintf("房间[%d] :", self.desk.deskId)
	} else {
		return fmt.Sprintf("房间[%d] 玩家[%s,%d]:", self.desk.deskId, self.players[chairId].BaseInfo.Nickname, chairId)
	}
}

//过吃
func (self *GameSink) AddGuoChiCards(chairId, card int32, chiType uint32) {

}

//过碰
func (self *GameSink) AddGuoPengCards(chairId, card int32) {

}

func (self *GameSink) popLeftCard(chairId int32) (card int32) {
	if _, ok := self.wantCards[chairId]; ok { //如果玩家要牌,优先发要的牌
		var index int //要牌索引
		for _, wcard := range self.wantCards[chairId] {
			for i, lcard := range self.leftCard {
				if wcard == lcard { //找到
					card = wcard
					self.leftCard = append(self.leftCard[:i], self.leftCard[i+1:]...)
					break
				}
			}
			index++
			if card != 0 {
				break
			}
		}
		//删除这张要过的牌
		if index >= len(self.wantCards[chairId]) {
			self.wantCards[chairId] = nil
		} else {
			self.wantCards[chairId] = self.wantCards[chairId][index:]
		}
		return
	}

	card = self.leftCard[len(self.leftCard)-1]
	self.leftCard = self.leftCard[:len(self.leftCard)-1]
	return
}

func (self *GameSink) doWantCards(chairId int32, cards []int32) {
	for _, v := range cards {
		if mj.IsVaildCard(v) {
			self.wantCards[chairId] = append(self.wantCards[chairId], v)
		}
	}
}
