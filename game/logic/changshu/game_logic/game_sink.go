package game_logic

import (
	pbgamemsg "cy/game/pb/game/mj/changshu"
)

//游戏主逻辑
type GameSink struct {
	GameConfig   *pbgamemsg.DeskArg //游戏参数
	players      []playerInfo       //玩家游戏信息
	onlinePlayer []bool             //在线玩家
	isPlaying    bool
	record       gameRecord
}

//构建游戏
func (self *GameSink) Ctor(config *pbgamemsg.DeskArg) error {
	self.GameConfig = config
	self.isPlaying = false
	self.onlinePlayer = make([]bool, 5)
	self.players = make([]playerInfo, 5)
	self.reset()
	return nil
}

//重置游戏
func (self *GameSink) reset() {

}

func (self *GameSink) StartGame() {

}

//玩家加入游戏
func (self *GameSink) AddPlayer(chairId int16, uid int32, nickName string) error {
	//TODO 1:一些检查
	//TODO 2:构建玩家信息,并保存到self.players
	//TODO 3:执行游戏记录的初始化
	//TODO 4:设置玩家在线
	return nil
}

//吃
func (self *GameSink) chiCard(chairId int16, uid int32, nickName string) error {
	//TODO 1:一些检查
	//TODO 2:构建玩家信息,并保存到self.players
	//TODO 3:执行游戏记录的初始化
	//TODO 4:设置玩家在线
	return nil

}
