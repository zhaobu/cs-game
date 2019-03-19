package main

import (
	majiang "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"math/rand"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
)

var DebugCard []uint8

//游戏公共信息
type gameAllInfo struct {
	diceResult     [4][2]int32 //投色子结果
	banker_id      int32       //庄家id
	left_rand_card []uint8     //发完牌后剩余的牌
}

//游戏私有信息
type gamePrivateInfo struct {
}

//游戏结束信息
type gameBalanceInfo struct {
}

type majiangLib struct {
	cardDef majiang.CardDef    //牌定义
	record  majiang.GameRecord //游戏回放
}

//游戏主逻辑
type GameSink struct {
	majiangLib
	desk              *Desk
	game_config       *pbgame_logic.CreateArg //游戏参数
	players           []playerInfo            //玩家游戏信息
	game_all_info     gameAllInfo             //游戏公共信息
	game_privite_info gamePrivateInfo         //游戏私有信息
	game_balance_info gameBalanceInfo         //游戏结束信息
	onlinePlayer      []bool                  //在线玩家
	baseCard          []uint8                 //基础牌库
	isPlaying         bool                    //是否在游戏中
	makeCards         bool                    //是否做牌
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

//游戏定时器
func (self *GameSink) timeOut() {

}

////////////////////////调用desk接口函数END/////////////////////////////

//构建游戏
func (self *GameSink) Ctor(config *pbgame_logic.CreateArg) error {
	self.game_config = config
	self.isPlaying = false
	self.onlinePlayer = make([]bool, config.PlayerCount)
	self.players = make([]playerInfo, config.PlayerCount)
	self.cardDef.Init(log)
	self.baseCard = self.cardDef.GetBaseCard(config.PlayerCount)
	self.reset()
	return nil
}

//重置游戏
func (self *GameSink) reset() {
	//game_all_info
	self.game_all_info = gameAllInfo{}
}

//玩家投色子
func (self *GameSink) ThrowDice(chairId int32, req *pbgame_logic.C2SThrowDice) {
	//检查玩家是否已经投过
	if self.game_all_info.diceResult[chairId][0] != 0 {
		log.Warn("玩家%d已经投过色子", chairId)
		return
	}

	//给玩家随机2个色子
	msg := &pbgame_logic.S2CThrowDiceResult{ChairId: chairId}
	rand.Seed(time.Now().Unix())
	for i := 0; i < 2; i++ {
		rnd := int32(rand.Intn(5) + 1)
		msg.DiceValue[i] = &pbgame_logic.Cyint32{T: rnd}
		self.game_all_info.diceResult[chairId][i] = rnd
	}
	//广播色子结果
	self.sendData(-1, msg)
	//判断是否所有人都投色子
	for i := int32(0); i < self.game_config.PlayerCount; i++ {
		if self.game_all_info.diceResult[i][0] == 0 {
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
	for i := 0; i < len(self.game_all_info.diceResult); i++ {
		diceRes[i].dice = self.game_all_info.diceResult[i][0] + self.game_all_info.diceResult[i][1]
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
	self.game_all_info.banker_id = diceRes[0].chairId
	msg := &pbgame_logic.S2CChangePos{PosInfo: posInfo}
	self.sendData(-1, msg)
	//1s后发送游戏开始消息
	self.deal_card()
}

//开始发牌
func (self *GameSink) deal_card() {
	msg := &pbgame_logic.S2CStartGame{BankerId: self.game_all_info.banker_id, CurInning: int32(self.desk.curInning)}
	msg.TotalCardNum = int32(len(self.baseCard))
	//洗牌
	self.shuffle_cards()
	var player_cards [][]uint8
	player_cards, self.game_all_info.left_rand_card = self.cardDef.DealCard(self.game_all_info.left_rand_card, self.game_config.PlayerCount, self.game_all_info.banker_id)
	msg.LeftCardNum = int32(len(self.game_all_info.left_rand_card))
	// var userInfo []*pbgame_logic.StartGameInfo
	for k, v := range self.players {
		v.cardInfo.handCards = player_cards[k]
		log.Warnf("房间[%d] 玩家[%s,%d]手牌为:%v", self.desk.id, v.baseInfo.nickname, k, player_cards[k])
		//统计每个玩家手牌数量
		v.cardInfo.stackCards = self.cardDef.StackCards(player_cards[k])
		//给每个玩家发送游戏开始消息
		self.sendData(int32(k), msg)
	}
	/*
			BankerId             int32
		TotalCardNum         int32
		LeftCardNum          int32
		CurInning            int32
		UserInfo             []*StartGameInfo
	*/
}

//洗牌
func (self *GameSink) shuffle_cards() {
	if self.makeCards {
		self.game_all_info.left_rand_card = DebugCard
		return
	}
	self.game_all_info.left_rand_card = self.cardDef.RandCards(self.baseCard)
}

//开始游戏
func (self *GameSink) StartGame() {
	self.isPlaying = true
	//通知第一个玩家投色子
	self.sendData(-1, &pbgame_logic.S2CThrowDice{ChairId: 0})
}

//玩家加入游戏
func (self *GameSink) AddPlayer(chairId int32, uid uint64, nickName string) bool {
	if self.game_config.PlayerCount < chairId {
		return false
	}

	info := playerInfo{baseInfo: playerBaseInfo{chairId: chairId, uid: uid, nickname: nickName, point: 0}}
	self.players[chairId] = info
	self.onlinePlayer[chairId] = true
	return true
}

//吃
func (self *GameSink) chiCard(chairId int32, uid uint64, nickName string) error {
	//TODO 1:一些检查
	//TODO 2:构建玩家信息,并保存到self.players
	//TODO 3:执行游戏记录的初始化
	//TODO 4:设置玩家在线
	return nil

}
