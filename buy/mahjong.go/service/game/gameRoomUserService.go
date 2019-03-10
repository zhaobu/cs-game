package game

import (
	"github.com/fwhappy/util"
	"mahjong.go/library/core"
	configService "mahjong.go/service/config"
)

// 房间用户信息
type RoomUser struct {
	UserId    int       // 用户id
	Index     int       // 房间位置
	Info      *UserInfo // 用户扩展信息
	CType     int       // 房间类型，为了能准确拿到info中的积分，故添加此冗余字段
	Longitude float64   // 经度
	Latitude  float64   // 纬度
}

// 创建一个房间内的用户
func NewRoomUser(userId int) *RoomUser {
	roomUser := &RoomUser{}
	roomUser.UserId = userId
	roomUser.Longitude = float64(-1)
	roomUser.Latitude = float64(-1)

	return roomUser
}

// 判断用户是否可以发送push
func (u *RoomUser) CanSendPush() bool {
	// return u.Info.Device == "ios" && len(u.Info.DeviceToken) > 0
	return len(u.Info.DeviceToken) > 0
}

// 给用户发送push
func (u *RoomUser) SendPush(langId int, senderUserId int, senderNickname, content string) {
	if !u.CanSendPush() {
		return
	}

	data := make(map[string]interface{})
	// 发送者信息
	data["senderId"] = senderUserId
	data["senderNickname"] = senderNickname
	// 接受者信息
	data["device"] = u.Info.Device
	data["deviceToken"] = u.Info.DeviceToken
	data["receiverId"] = u.UserId
	// 语言包id
	data["langId"] = langId
	if len(content) > 0 {
		// fixme 这里直接拼接了字符串，不够规范
		data["content"] = "[" + senderNickname + "]" + content
	}
	// 发送时间
	data["time"] = util.GetTime()

	// 写入redis队列
	redisListPush(data)

	core.Logger.Debug("推送消息:%#v", data)
}

// GetAccumulativeScore 读取房间用户的累积积分
// 如果是随机房间, 返回累计随机积分
// 如果是比赛房间, 返回比赛积分
// 如果是自主创建的房间, 返回总积分
func (u *RoomUser) GetAccumulativeScore() int {
	var score int
	if configService.IsCreateRoom(u.CType) || configService.IsTVRoom(u.CType) {
		score = u.Info.Score
	} else if configService.IsRandomRoom(u.CType) {
		score = u.Info.ScoreRandom
	} else if configService.IsMatchRoom(u.CType) {
		score = u.Info.ScoreMatch
	} else if configService.IsClubMatchRoom(u.CType) {
		// 俱乐部淘汰赛
		score = u.Info.ScoreClub
	} else if configService.IsCoinRoom(u.CType) {
		score = u.Info.ScoreCoin
	} else if configService.IsLeagueRoom(u.CType) {
		score = u.Info.ScoreLeague
	} else if configService.IsRankRoom(u.CType) {
		// score = u.Info.ScoreRank
		score = u.Info.RankExp
	}
	return score
}
