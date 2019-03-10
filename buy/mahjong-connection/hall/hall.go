package hall

import (
	"mahjong-connection/config"
	"mahjong-connection/core"
	"mahjong-connection/model"
	"mahjong-connection/protocal"
	"net"
	"sync"
	"time"

	"github.com/fwhappy/util"
)

// UserRemote 用户remote与userId对照表
type UserRemote struct {
	ID         int
	RemoteAddr string
}

// 全局变量
var (
	RemoteAddr        string // 服务器的地址:port
	Version           string // 服务器版本
	UserMap           *model.UserSet
	UserRemoteMap     *sync.Map
	ConnectionSet     *sync.Map // 连接的用户集合
	WaitConnectionSet *sync.Map // 等候连接的用户集合
)

func init() {
	UserMap = model.NewUserSet()
	UserRemoteMap = &sync.Map{}
	ConnectionSet = &sync.Map{}
	WaitConnectionSet = &sync.Map{}
}

// BroadcastMessage 推送全局消息
func BroadcastMessage(p *protocal.ImPacket) {
	UserMap.LoadUsers().Range(func(k, v interface{}) bool {
		u := v.(*model.User)
		u.SendMessageToClient(p)
		return true
	})
}

// PrivateMessage 推送私人消息消息
func PrivateMessage(userId int, p *protocal.ImPacket) {
	if u := UserMap.Load(userId); u != nil {
		u.SendMessageToClient(p)
	}
}

// ListenHandShakeTimeout 监听用户连接-握手是否超时
func ListenHandShakeTimeout(conn *net.TCPConn) {
	core.Logger.Trace("[ListenHandShakeTimeout]监听连接handshake, remote:%v", conn.RemoteAddr().String())
	c := make(chan int, 1)
	WaitConnectionSet.Store(conn.RemoteAddr().String(), c)

	go func() {
		// 捕获异常
		defer util.RecoverPanic()
		select {
		case code := <-c:
			core.Logger.Debug("收到退出监听连接handshake的指令，退出监听, code:%v, remote:%v", code, conn.RemoteAddr().String())
			WaitConnectionSet.Delete(conn.RemoteAddr().String())
			break
		case <-time.After(time.Duration(config.HANDSHAKE_TIMEOUT) * time.Second):
			core.Logger.Debug("用户长时间未handshark成功，被踢下线:%s", conn.RemoteAddr().String())
			conn.Close()
			break
		}
	}()
}

// FinishListenHandShakeTimeout 结束监听用户连接-握手超时
func FinishListenHandShakeTimeout(conn *net.TCPConn, code int) {
	// 删除handshake超时监听
	c, ok := WaitConnectionSet.Load(conn.RemoteAddr().String())
	if ok {
		select {
		case c.(chan int) <- code:
		default:
			core.Logger.Warn("[ClientHandShakeAck]wait connection channel 删除失败, code:%v, remote:%v", code, conn.RemoteAddr().String())
		}
	}
}

// StatLastActionTime 记录服务器在线人数、活跃数据
func StatLastActionTime() {
	// 捕获异常
	defer util.RecoverPanic()
	for {
		// 设置服务器在线用户数
		// model.SetRemoteUserCnt(ServerRemote, UserSet.Len())
		// 设置服务器最后活动时间
		// model.SetRemoteActionTime(ServerRemote, util.GetTime())
		time.Sleep(time.Second)
	}
}

// KickConn 踢出连接
func KickConn(conn *net.TCPConn) {
	// 检查是否处理等待handshake状态
	remoteAddr := conn.RemoteAddr().String()
	core.Logger.Debug("[KickConn]conn:%v", remoteAddr)
	// 清除等待连接的监听
	FinishListenHandShakeTimeout(conn, 1)
	id, ok := ConnectionSet.Load(remoteAddr)
	if !ok {
		// 未找到连接与userId的对应关系
		// 直接关闭连接
		conn.Close()
		return
	}
	KickUser(id.(int), conn)
}

// KickUser 踢出用户
func KickUser(id int, conn *net.TCPConn) {
	core.Logger.Debug("[KickUser]userId:%v", id)
	if u := UserMap.Load(id); u != nil {
		if u.ClientConn != conn {
			return
		}
		u.QuitOnce.Do(func() {
			// 关闭用户消息队列
			u.ClientMQ.Close()
			// 删除用户
			ConnectionSet.Delete(id)
			UserMap.Delete(id)
			// 关闭用户消息队列
			u.ClientConn.Close()
			if u.League != nil {
				u.League.Close()
			}
			if u.Club != nil {
				u.Club.Close()
			}
			if u.Game != nil {
				u.Game.Close()
			}
		})
	}
}
