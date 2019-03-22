package main

import (
	mj "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"math/rand"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
)

var DebugCard []int32

//游戏公共信息
type gameAllInfo struct {
	game_player_operation PriorityOper    //最高优先级操作
	CanOperInfo                           //玩家能做的操作
	game_balance_info     gameBalanceInfo //游戏结束信息
	diceResult            [4][2]int32     //投色子结果
	banker_id             int32           //庄家id
	left_rand_card        []int32         //发完牌后剩余的牌
	curThrowDice          int32           //当前投色子的玩家
	isdeuce               bool            //是否流局
}

//游戏结束信息
type gameBalanceInfo struct {
}

type mjLib struct {
	cardDef mj.CardDef      //牌定义
	record  mj.GameRecord   //游戏回放
	huLib   mj.HuLib        //胡牌算法
	players []mj.PlayerInfo //玩家游戏信息
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
	makeCards    bool                    //是否做牌
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
	self.cardDef.Init(log)
	self.baseCard = self.cardDef.GetBaseCard(config.PlayerCount)
	self.reset()
	return nil
}

//重置游戏
func (self *GameSink) reset() {
	//all_info
	self.gameAllInfo = gameAllInfo{}
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
		log.Errorf("人数已满,游戏开始人数为%d", self.game_config.PlayerCount)
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
		log.Warnf("当前应投色子玩家为%d,操作玩家为%d", self.curThrowDice, chairId)
		return
	}
	//检查玩家是否已经投过
	if self.diceResult[chairId][0] != 0 {
		log.Warnf("玩家%d已经投过色子", chairId)
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
	player_cards, self.left_rand_card = self.cardDef.DealCard(self.left_rand_card, self.game_config.PlayerCount, self.banker_id)
	msg.LeftCardNum = int32(len(self.left_rand_card))
	for k, v := range self.players {
		v.CardInfo.HandCards = player_cards[k]
		msg.UserInfo = &pbgame_logic.StartGameInfo{HandCard: player_cards[k]}
		log.Warnf("房间[%d] 玩家[%s,%d]手牌为:%v", self.desk.id, v.BaseInfo.Nickname, k, player_cards[k])
		//统计每个玩家手牌数量
		v.CardInfo.StackCards = self.cardDef.StackCards(player_cards[k])
		//给每个玩家发送游戏开始消息
		self.sendData(int32(k), msg)
	}
	//庄家开始第一次补花
	self.firstBuHua(self.banker_id)
	//检查庄家能否胡
	if ok, huTypeList := self.huLib.CheckHuType(&(self.players[self.banker_id].CardInfo)); ok {
		self.CanOperInfo.CanHu[self.banker_id] = &CanHuOper{HuList: huTypeList}
	}
	//检查能否杠
}

//玩家第一次补花
func (self *GameSink) firstBuHua(chairId int32) {
	cardInfo := self.players[chairId].CardInfo
	leftCard := self.left_rand_card
	huaIndex := make(map[int32]int32) //下次要补的花牌
	leftCount := len(leftCard)

	msg := pbgame_logic.S2CBuHua{ChairId: chairId}      //发送给补花玩家
	msgOther := pbgame_logic.S2CBuHua{ChairId: chairId} //发送给其他玩家

	//补一张花牌
	operOnce := func(card int32) int32 {
		//减一张花牌
		self.cardDef.Sub_stack(cardInfo.StackCards, card)
		cardInfo.StackCards[card]--
		//从牌库摸一张牌
		moCard := leftCard[leftCount-1]
		leftCount--
		//摸的牌加到手牌
		self.cardDef.Add_stack(cardInfo.StackCards, moCard)
		//记录到消息
		msg.BuHuaResult = append(msg.BuHuaResult, &pbgame_logic.BuHuaOnce{HuaCard: card, BuCard: moCard})
		msgOther.BuHuaResult = append(msgOther.BuHuaResult, &pbgame_logic.BuHuaOnce{HuaCard: card, BuCard: 0})
		if self.cardDef.IsHuaCard(moCard) {
			self.cardDef.Add_stack(huaIndex, moCard)
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
					if self.cardDef.IsHuaCard(operOnce(huaCard)) {
						bFin = false
					}
					//补一张减一张
					self.cardDef.Sub_stack(huaIndex, huaCard)
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

//摸牌 last(-1摸最后一张 1第一次摸牌 0正常摸牌),lose_chair在明杠时为放杠玩家id,包赔
func (self *GameSink) draw_card(chairId, last, lose_chair int32) error {
	//检查游戏是否结束
	if len(self.left_rand_card) <= 0 {
		self.isdeuce = true
		self.gameEnd()
		return nil
	}
	self.CanOperInfo.ResetCanOper()
	self.game_player_operation.ResetPriorityOper()

	var card, index int32 = 0, 0
	if last == -1 {
		index = int32(len(self.left_rand_card)) - 1
	}
	card = self.left_rand_card[index]
	self.left_rand_card = self.left_rand_card[:index]
	log.Debugf("房间[%d] 玩家[%s,%d]摸牌[%d]剩余[%d]张", self.desk.id, self.players[chairId].BaseInfo.Nickname, chairId, card, index)

	return nil
}

//游戏结束
func (self *GameSink) gameEnd() {

}

//洗牌
func (self *GameSink) shuffle_cards() {
	if self.makeCards {
		self.left_rand_card = DebugCard
		return
	}
	self.left_rand_card = self.cardDef.RandCards(self.baseCard)
}

//吃
func (self *GameSink) chiCard(chairId int32) error {
	//TODO 1:一些检查
	//TODO 2:构建玩家信息,并保存到self.players
	//TODO 3:执行游戏记录的初始化
	//TODO 4:设置玩家在线
	return nil

}

//出牌
func (self *GameSink) outCard(chairId int32) error {

	return nil
}
