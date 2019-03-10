package game

import (
	"net"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"

	fbsCommon "mahjong.go/fbs/Common"
)

// ClubPool 俱乐部连接池
type ClubPool struct {
	remote        string
	mq            chan *protocal.ImPacket // 消息队列
	conn          *net.TCPConn            // 活动连接
	isConnected   bool
	heartbeatTime int64
}

// NewClubPool 新建一个俱乐部连接池
func NewClubPool() *ClubPool {
	c := &ClubPool{
		mq:            make(chan *protocal.ImPacket, 1024),
		isConnected:   false,
		heartbeatTime: util.GetTime(),
	}
	return c
}

// 添加消息到消息队列
// 如果连接未建立或者连接失败，则不写消息，防止俱乐部服务异常时，导致游戏逻辑出错
func (c *ClubPool) appendMessage(impacket *protocal.ImPacket) {
	if c.isConnected {
		c.mq <- impacket
	}
}

// 从连接池中取出一个活跃的连接
func (c *ClubPool) getConn() *net.TCPConn {
	return c.conn
}

// 关闭连接
func (c *ClubPool) closeConn() {
	c.isConnected = false
	c.conn.Close()
}

// 推送消息
func (c *ClubPool) sendMessage() {
	core.Logger.Debug("开启游戏服向俱乐部的消息推送routine,server:%v, club:%v", GetRemoteAddr(), c.remote)
	for impacket := range c.mq {
		if c.isConnected {
			impacket.Send(c.getConn())
		}
	}
}

// 启动俱乐部连接池
func (c *ClubPool) Run() {
	// 连接俱乐部
	defer util.RecoverPanic()

	c.remote = core.AppConfig.ClubRemote

	if c.remote == "" {
		core.Logger.Warn("俱乐部未配置，不推送俱乐部消息")
		return
	}

	// 启动消息推送服务
	go c.sendMessage()

	// 启动连接检测
	go c.listenConnection()

	// 启动心跳监测
	// go c.listenHeartbeat()

	// 开启心跳发送
	// go c.loopSendHeartbeat()

	core.Logger.Debug("俱乐部功能已开启,server:%v, club:%v", GetRemoteAddr(), c.remote)
}

// 监听俱乐部连接
func (c *ClubPool) listenConnection() {
	defer util.RecoverPanic()
	core.Logger.Debug("开启游戏服与俱乐部连接的监听,server:%v, club:%v", GetRemoteAddr(), c.remote)

	for {
		// 每5秒检测一下连接状态，如果连接断开，则重新建立连接
		time.Sleep(time.Second * 5)
		if !c.isConnected {
			core.Logger.Debug("游戏服与俱乐部未建立连接或者连接已断开，开始连接,server:%v, club:%v", GetRemoteAddr(), c.remote)
			// 连接
			c.connect()
		}
	}
}

func (c *ClubPool) listenHeartbeat() {
	defer util.RecoverPanic()

	for {
		time.Sleep(time.Minute)
		if !c.isConnected {
			continue
		}
		// 心跳超时，关闭连接
		if util.GetTime()-c.heartbeatTime > int64(120) {
			core.Logger.Warn("与俱乐部的连接心跳超时，断开连接,server:%v,club:%v", GetRemoteAddr(), c.remote)
			c.closeConn()
		}
	}
}

func (c *ClubPool) loopSendHeartbeat() {
	defer util.RecoverPanic()

	for {
		time.Sleep(time.Minute)
		if !c.isConnected {
			continue
		}

		// 发送心跳
		message, _ := simplejson.New().Encode()
		imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HEARTBEAT, message)
		c.appendMessage(imPacket)
	}
}

// 连接俱乐部服务
func (c *ClubPool) connect() {
	core.Logger.Info("开始连接俱乐部,server:%v,club:%v", GetRemoteAddr(), c.remote)
	if c.remote == "" {
		core.Logger.Error("连接俱乐部服务失败,remote未设置,server:%v", GetRemoteAddr())
		return
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.remote)
	if err != nil {
		core.Logger.Error("连接俱乐部服务失败(ResolveTCPAddr),server:%v,club:%v,err:%v", GetRemoteAddr(), c.remote, err.Error())
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		core.Logger.Error("连接俱乐部服务失败(DialTCP),server:%v,club:%v,err:%v", GetRemoteAddr(), c.remote, err.Error())
		return
	}
	c.conn = conn
	c.isConnected = true
	go c.listenMessageReceive()

	// 连接成功后，发送一条系统消息，让俱乐部服务知道这条连接是来自游戏服务器的，不要因为没有handshake就踢
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_SYSTEM, nil)
	c.appendMessage(imPacket)

	core.Logger.Info("连接俱乐部成功,server:%v,club:%v", GetRemoteAddr(), c.remote)
}

// 开启俱乐部消息接收
func (c *ClubPool) listenMessageReceive() {
	defer util.RecoverPanic()
	defer func() {
		c.isConnected = false
	}()

	for {
		impacket, err := protocal.ReadPacket(c.conn)
		// 这里只要是接收消息出错，基本就可以认为是连接中断了
		if err != nil {
			core.Logger.Error("[ClubPool.listenMessageReceive]接收消息出错,server:%v,club:%v,error:%v", GetRemoteAddr(), c.remote, err.Error())
			break
		}

		// 解析消息
		switch impacket.GetPackage() {
		case protocal.PACKAGE_TYPE_HEARTBEAT: // 心跳
			c.handleHeartbeat()
		case protocal.PACKAGE_TYPE_DATA:
			c.handleData(int(impacket.GetMessageId()), impacket.GetBody())
		}
	}
}

// 处理服务器返回来的心跳
func (c *ClubPool) handleHeartbeat() {
	c.heartbeatTime = util.GetTime()
	core.Logger.Debug("收到俱乐部的回应心跳,server:%v,club:%v", GetRemoteAddr(), c.remote)
}

// 处理服务器返回来的数据请求
func (c *ClubPool) handleData(messageId int, body []byte) {
	switch messageId {
	case fbsCommon.CommandClubC2GRoomActiveResponse: // 客户端询问房间是否活跃的返回
		c2gRoomActive(body)
	default:
		core.Logger.Error("收到来自俱乐部服务的未知消息,server:%v,club:%v,messageId:%v", GetRemoteAddr(), c.remote, messageId)
	}
}

// 处理俱乐部服务返回来的房间活跃会员
func c2gRoomActive(body []byte) {
	response := fbsCommon.GetRootAsClubC2GRoomActivePesponse(body, 0)
	clubId := int(response.ClubId())
	roomId := int64(response.RoomId())
	isActive := int(response.Active())

	core.Logger.Debug("收到房间活跃检测的回应,clubId:%v,roomId:%v,isActive:%v", clubId, roomId, isActive)
	if isActive == 0 {
		// 如果房间已经不活跃了，需要重新推送房间消息到俱乐部服务
		room, err := RoomMap.GetRoom(roomId)
		if err != nil {
			core.Logger.Warn("收到房间活跃检测的回应, 但是本地房间不在了,clubId:%v,roomId:%v", clubId, roomId)
		} else {
			CPool.appendMessage(ClubG2CReloadRoomPush(clubId, room))
			core.Logger.Info("房间活跃检测失败，重新推送房间信息到俱乐部服务成功,clubId:%v,roomId:%v", clubId, roomId)
		}
	}
}
