package main

import (
	pbgamemsg "cy/game/pb/game/mj/changshu"

	"github.com/gogo/protobuf/proto"
)

//游戏主逻辑
type GameSink struct {
	*Desk
	GameConfig   *pbgamemsg.DeskArg //游戏参数
	players      []playerInfo       //玩家游戏信息
	onlinePlayer []bool             //在线玩家
	isPlaying    bool
	record       gameRecord
}

func (self *GameSink) sendData(chairId uint16, msg proto.Message) {

}

//构建游戏
func (self *GameSink) Ctor(config *pbgamemsg.DeskArg) error {
	self.GameConfig = config
	self.isPlaying = false
	self.onlinePlayer = make([]bool, config.PlayerCount)
	self.players = make([]playerInfo, config.PlayerCount)
	self.reset()
	return nil
}

//重置游戏
func (self *GameSink) reset() {

}

func (self *GameSink) StartGame() {
	self.isPlaying = true
	//通知玩家投色子

}

//玩家加入游戏
func (self *GameSink) AddPlayer(chairId uint16, uid uint64, nickName string) bool {
	if self.GameConfig.PlayerCount < uint32(chairId) {
		return false
	}

	info := playerInfo{baseInfo: playerBaseInfo{chairId: chairId, uid: uid, nickname: nickName, point: 0}}
	self.players[chairId] = info
	self.onlinePlayer[chairId] = true
	return true
}

//吃
func (self *GameSink) chiCard(chairId int16, uid int32, nickName string) error {
	//TODO 1:一些检查
	//TODO 2:构建玩家信息,并保存到self.players
	//TODO 3:执行游戏记录的初始化
	//TODO 4:设置玩家在线
	return nil

}
