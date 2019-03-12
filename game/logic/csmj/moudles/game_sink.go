package moudles

import (
	pbgamemsg "cy/game/pb/game/mj"
)

//游戏主逻辑
type GameSink struct {
	GameConfig   pbgamemsg.RoomArg //游戏参数
	players      []PlayerInfo      //玩家游戏信息
	onlinePlayer []bool            //在线玩家

}

//构建游戏
func (self *GameSink) Ctor(config *pbgamemsg.RoomArg) error {
	self.GameConfig = config
	self.isPlaying = false
	self.onlinePlayer = make([]bool, 5)
	self.players = make([]PlayerInfo, 5)
	self.gameRecord = nil
	self.reset()
	return
}

//重置游戏
func (self *GameSink) reset() {

}

func (self *GameSink) StartGame() {
	//TODO 1:洗牌
	//TODO 2:定庄

}

//玩家加入游戏
func (self *GameSink) AddPlayer(chairId int16, uid int32, nickName string) error {
	//TODO 1:一些检查
	//TODO 2:构建玩家信息,并保存到self.players
	//TODO 3:执行游戏记录的初始化
	//TODO 4:设置玩家在线
}

//玩家加入游戏
func (self *GameSink) (chairId int16, uid int32, nickName string) error {
	//TODO 1:一些检查
	//TODO 2:构建玩家信息,并保存到self.players
	//TODO 3:执行游戏记录的初始化
	//TODO 4:设置玩家在线
}
