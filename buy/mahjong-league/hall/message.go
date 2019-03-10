package hall

import (
	"mahjong-league/protocal"
	"mahjong-league/user"
	"net"
	"strings"

	"github.com/fwhappy/util"
)

// BrodcastMessage 广播消息给所有的用户
func BrodcastMessage(impacket *protocal.ImPacket) {
	// 防止用户量比较大，这里开启新routine去推送
	go func() {
		defer util.RecoverPanic()
		UserSet.Users.Range(func(k, v interface{}) bool {
			v.(*user.User).SendMessage(impacket)
			return true
		})
	}()
}

// BrodcastMessageWithVersion 广播消息给所有符合版本的用户
func BrodcastMessageWithVersion(impacket *protocal.ImPacket, requireVersion string) {
	// 防止用户量比较大，这里开启新routine去推送
	go func() {
		defer util.RecoverPanic()
		UserSet.Users.Range(func(k, v interface{}) bool {
			u := v.(*user.User)
			if strings.Compare(u.Version, requireVersion) > -1 {
				u.SendMessage(impacket)
			}
			return true
		})
	}()
}

// RaceMessage 推送消息给所有参加比赛的用户
func RaceMessage(raceID int, impacket *protocal.ImPacket) {
	go func() {
		defer util.RecoverPanic()
	}()
}

// PrivateMessage 私人消息 推送私人消息
func PrivateMessage(userID int, impacket *protocal.ImPacket) bool {
	v, ok := UserSet.Users.Load(userID)
	if ok {
		v.(*user.User).SendMessage(impacket)
	}
	return ok
}

// BrodcastMessageToGameServers 广播消息给所有的游戏服
func BrodcastMessageToGameServers(impacket *protocal.ImPacket) {
	GameServers.Range(func(k, v interface{}) bool {
		v.(*net.TCPConn).Write(impacket.Serialize())
		return true
	})
}
