package hall

import (
	"mahjong.club/club"
	"mahjong.club/core"
	"mahjong.club/room"
	"mahjong.club/user"

	"github.com/fwhappy/mahjong/protocal"
)

// 定义全局变量
var (
	UserSet *user.Set
	RoomSet *room.Set
	ClubSet *club.Set
)

func init() {
	UserSet = user.NewSet()
	RoomSet = room.NewSet()
	ClubSet = club.NewSet()
}

// SendUserMessage 通过Id给用户发送一条消息
func SendUserMessage(id int, imPacket *protocal.ImPacket) bool {
	if u, online := UserSet.Get(id); online {
		u.AppendMessage(imPacket)
		return true
	}
	return false
}

// SendClubMessage 推送消息给房间用户
func SendClubMessage(c *club.Club, impacket *protocal.ImPacket) {
	c.Users.Range(func(k, v interface{}) bool {
		SendUserMessage(k.(int), impacket)
		return true
	})
}

// RemoveClubUser 移除房间成员
func RemoveClubUser(clubID int, u *user.User) {
	// 读取房间信息
	c, exists := ClubSet.Get(clubID)
	if !exists {
		core.Logger.Warn("[RemoveClubUser]club not exists, clubID:%v, userID:%v", clubID, u.ID)
		return
	}

	// 从成员列表删除
	c.DelUser(u.ID)
	core.Logger.Debug("[RemoveClubUser]clubID:%v, userID:%v", c.ID, u.ID)
}
