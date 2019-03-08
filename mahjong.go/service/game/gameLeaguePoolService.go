package game

import (
	"net"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	"mahjong.go/library/core"
	"mahjong.go/library/response"
	"mahjong.go/mi/protocal"

	fbsCommon "mahjong.go/fbs/Common"
)

// LeaguePool 联赛连接池
type LeaguePool struct {
	remote        string
	mq            chan *protocal.ImPacket // 消息队列
	conn          *net.TCPConn            // 活动连接
	isConnected   bool
	heartbeatTime int64
}

// NewLeaguePool 新建一个联赛连接池
func NewLeaguePool() *LeaguePool {
	c := &LeaguePool{
		mq:            make(chan *protocal.ImPacket, 1024),
		isConnected:   false,
		heartbeatTime: util.GetTime(),
	}
	return c
}

// 添加消息到消息队列
// 如果连接未建立或者连接失败，则不写消息，防止联赛服务异常时，导致游戏逻辑出错
func (l *LeaguePool) appendMessage(impacket *protocal.ImPacket) {
	if l.isConnected {
		l.mq <- impacket
	}
}

// 从连接池中取出一个活跃的连接
func (l *LeaguePool) getConn() *net.TCPConn {
	return l.conn
}

// 关闭连接
func (l *LeaguePool) closeConn() {
	l.isConnected = false
	l.conn.Close()
}

// 推送消息
func (l *LeaguePool) sendMessage() {
	core.Logger.Debug("开启游戏服向联赛的消息推送routine,server:%v, league:%v", GetRemoteAddr(), l.remote)
	for impacket := range l.mq {
		if l.isConnected {
			impacket.Send(l.getConn())
		}
	}
}

// Run 启动联赛连接池
func (l *LeaguePool) Run() {
	// 连接联赛
	defer util.RecoverPanic()

	l.remote = core.AppConfig.LeagueRemote
	if l.remote == "" {
		core.Logger.Warn("联赛未配置，不推送联赛消息")
		return
	}

	// 启动消息推送服务
	go l.sendMessage()

	// 启动连接检测
	go l.listenConnection()

	core.Logger.Debug("联赛功能已开启,server:%v, league:%v", GetRemoteAddr(), l.remote)
}

// 监听联赛连接
func (l *LeaguePool) listenConnection() {
	defer util.RecoverPanic()
	core.Logger.Debug("开启游戏服与联赛连接的监听,server:%v, league:%v", GetRemoteAddr(), l.remote)

	for {
		// 每5秒检测一下连接状态，如果连接断开，则重新建立连接
		time.Sleep(time.Second * 5)
		if !l.isConnected {
			core.Logger.Debug("游戏服与联赛未建立连接或者连接已断开，开始连接,server:%v, league:%v", GetRemoteAddr(), l.remote)
			// 连接
			l.connect()
		}
	}
}

// 连接联赛服务
func (l *LeaguePool) connect() {
	core.Logger.Info("开始连接联赛,server:%v,league:%v", GetRemoteAddr(), l.remote)
	if l.remote == "" {
		core.Logger.Error("连接联赛服务失败,remote未设置,server:%v", GetRemoteAddr())
		return
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", l.remote)
	if err != nil {
		core.Logger.Error("连接联赛服务失败(ResolveTCPAddr),server:%v,league:%v,err:%v", GetRemoteAddr(), l.remote, err.Error())
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		core.Logger.Error("连接联赛服务失败(DialTCP),server:%v,league:%v,err:%v", GetRemoteAddr(), l.remote, err.Error())
		return
	}
	l.conn = conn
	l.isConnected = true
	go l.listenMessageReceive()

	// 连接成功后，发送一条系统消息，让联赛服务知道这条连接是来自游戏服务器的，不要因为没有handshake就踢
	js := simplejson.New()
	js.Set("remote", GetRemoteAddr())
	l.appendMessage(response.GenJson(protocal.PACKAGE_TYPE_SYSTEM, js))

	core.Logger.Info("连接联赛成功,server:%v,league:%v", GetRemoteAddr(), l.remote)
}

// 开启联赛消息接收
func (l *LeaguePool) listenMessageReceive() {
	defer util.RecoverPanic()
	defer func() {
		l.isConnected = false
		core.Logger.Info("[LeaguePool.listenMessageReceive]退出接收消息,server:%v,league:%v", GetRemoteAddr(), l.remote)
	}()

	core.Logger.Info("[LeaguePool.listenMessageReceive]开启消息接收,server:%v,league:%v", GetRemoteAddr(), l.remote)

	for {
		impacket, err := protocal.ReadPacket(l.conn)
		// 这里只要是接收消息出错，基本就可以认为是连接中断了
		if err != nil {
			core.Logger.Error("[LeaguePool.listenMessageReceive]接收消息出错,server:%v,league:%v,error:%v", GetRemoteAddr(), l.remote, err.Error())
			break
		}

		core.Logger.Debug("[listenMessageReceive]收到消息,packageId:%v, messageId:%v", impacket.GetPackage(), impacket.GetMessageId())

		// 解析消息
		switch impacket.GetPackage() {
		case protocal.PACKAGE_TYPE_DATA:
			go l.handleData(impacket)
		}
	}
}

// 处理服务器返回来的数据请求
func (l *LeaguePool) handleData(impacket *protocal.ImPacket) {
	core.Logger.Debug("[handleData]收到消息,packageId:%v, messageId:%v", impacket.GetPackage(), impacket.GetMessageId())
	defer util.RecoverPanic()
	switch impacket.GetMessageId() {
	case fbsCommon.CommandLeagueL2SPlanPush: // 联赛通知游戏服排赛
		l2sPlanPush(impacket)
	case fbsCommon.CommandLeagueL2SRankRefreshPush:
		l2sRankRefresh(impacket)
	default:
		core.Logger.Error("收到来自联赛服务的未知消息,server:%v,league:%v,messageId:%v", GetRemoteAddr(), l.remote, impacket.GetMessageId())
	}
}
