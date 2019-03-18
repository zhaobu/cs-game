package main

import (
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"math/rand"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
)

//游戏公共信息
type gameAllInfo struct {
	throwDice [4][2]int32 //投色子结果
}

//游戏私有信息
type gamePrivateInfo struct {
}

//游戏结束信息
type gameBalanceInfo struct {
}

//游戏主逻辑
type GameSink struct {
	desk              *Desk
	game_config       *pbgame_logic.CreateArg //游戏参数
	players           []playerInfo            //玩家游戏信息
	record            *gameRecord             //游戏回放
	game_all_info     gameAllInfo             //游戏公共信息
	game_privite_info gamePrivateInfo         //游戏私有信息
	game_balance_info gameBalanceInfo         //游戏结束信息
	onlinePlayer      []bool                  //在线玩家
	isPlaying         bool
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
	self.players = make([]playerInfo, config.PlayerCount)
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
	if self.game_all_info.throwDice[chairId][0] != 0 {
		log.Warn("玩家%d已经投过色子", chairId)
		return
	}

	//给玩家随机2个色子
	msg := &pbgame_logic.S2CThrowDiceResult{ChairId: chairId}
	rand.Seed(time.Now().Unix())
	for i := 0; i < 2; i++ {
		rnd := int32(rand.Intn(6-1) + 1)
		msg.DiceValue[i] = &pbgame_logic.Cyint32{T: rnd}
		self.game_all_info.throwDice[chairId][i] = rnd
	}
	//广播色子结果
	self.sendData(-1, msg)
	//判断是否所有人都投色子
	for i := int32(0); i < self.game_config.PlayerCount; i++ {
		if self.game_all_info.throwDice[i][0] == 0 {
			//通知下一个玩家投色子
			self.sendData(-1, &pbgame_logic.S2CThrowDice{ChairId: i})
			return
		}
	}
	self.dealDiceResult()
}

//处理投色子结果
func (self *GameSink) dealDiceResult() {
	diceRes := make([][]int32, self.game_config.PlayerCount)
	for i := 0; i < len(self.game_all_info.throwDice); i++ {
		diceRes[i][0] = self.game_all_info.throwDice[i][0] + self.game_all_info.throwDice[i][1]
	}
	//排序，实现比较方法即可
	sort.Slice(diceRes, func(i, j int) bool {
		return diceRes[i][0] > diceRes[j][0]
	})
}

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
