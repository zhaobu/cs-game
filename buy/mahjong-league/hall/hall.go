package hall

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/model"
	"mahjong-league/user"
	"net"
	"sync"
	"time"

	"github.com/fwhappy/util"
)

// 定义全局变量
var (
	UserSet           *user.Set // handshake的用户集合
	ConnectionSet     *sync.Map // 连接的用户集合
	WaitConnectionSet *sync.Map // 连接的用户集合
	ServerRemote      string
	ServerVersion     string
	GameServers       *sync.Map // map[string]*net.Conn // 已连接的游戏服列表
	LastRaceResultSet *sync.Map // 用户最后比赛结果
)

func init() {
	UserSet = user.NewSet()
	ConnectionSet = &sync.Map{}
	WaitConnectionSet = &sync.Map{}
	GameServers = &sync.Map{}
	LastRaceResultSet = &sync.Map{}
}

// ListenHandShakeTimeout 监听用户连接-握手是否超时
func ListenHandShakeTimeout(conn *net.TCPConn) {
	c := make(chan int, 1)
	WaitConnectionSet.Store(conn.RemoteAddr().String(), c)

	go func() {
		// 捕获异常
		defer util.RecoverPanic()
		select {
		case <-c:
			core.Logger.Debug("收到退出监听连接handshake的指令，退出监听:%v", conn.RemoteAddr().String())
			WaitConnectionSet.Delete(conn.RemoteAddr().String())
			break
		case <-time.After(time.Duration(config.HANDSHAKE_TIMEOUT) * time.Second):
			core.Logger.Debug("用户长时间未handshark成功，被踢下线:%s", conn.RemoteAddr().String())
			conn.Close()
			break
		}
	}()
}

// KickConn 踢出连接
func KickConn(conn *net.TCPConn) {
	// 检查是否处理等待handshake状态
	remoteAddr := conn.RemoteAddr().String()
	core.Logger.Debug("[KickConn]conn:%v", remoteAddr)
	c, ok := WaitConnectionSet.Load(remoteAddr)
	if ok {
		select {
		case c.(chan int) <- 1:
		default:
			core.Logger.Warn("[KickConn]wait connection channel is full:%v", remoteAddr)
		}
	}
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
func KickUser(userID int, conn *net.TCPConn) {
	core.Logger.Debug("[KickUser]userID:%v", userID)
	u, exists := UserSet.Get(userID)
	if exists {
		if u.Conn != conn {
			return
		}
		u.QuitOnce.Do(func() {
			UserSet.Del(userID)
			u.Conn.Close()
			u.Mq.Close()
		})
	}
}

// StatLastActionTime 记录服务器在线人数、活跃数据
func StatLastActionTime() {
	// 捕获异常
	defer util.RecoverPanic()
	for {
		// 设置服务器在线用户数
		model.SetRemoteUserCnt(ServerRemote, UserSet.Len())
		// 设置服务器最后活动时间
		model.SetRemoteActionTime(ServerRemote, util.GetTime())
		time.Sleep(time.Second)
	}
}
