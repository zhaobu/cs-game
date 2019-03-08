package game

import (
	"net"
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
)

// RClub 房间所属俱乐部
type RClub struct {
	ClubId        int   // 俱乐部id
	RoomId        int64 // 房间id
	Remote        string
	Conn          *net.TCPConn            // 活动连接
	SendMq        chan *protocal.ImPacket // 消息发送队列
	ConnectStatus int                     // 连接状态,0:未连接;1:连接中;2:已连接
	Mux           *sync.RWMutex
}

func (rc *RClub) isNotConnect() bool {
	return rc.ConnectStatus == 0
}

func (rc *RClub) isConnecting() bool {
	return rc.ConnectStatus == 1
}

func (rc *RClub) isConnected() bool {
	return rc.ConnectStatus == 2
}

// 获取房间对应的俱乐部的id
func getRClubRemote(hashValue int64) string {
	return core.AppConfig.ClubRemote
}

// NewRClub 创建一个俱乐部
func NewRClub(roomId int64, clubId int) *RClub {
	rClub := &RClub{
		ClubId:        clubId,
		RoomId:        roomId,
		Remote:        getRClubRemote(roomId),
		Mux:           &sync.RWMutex{},
		SendMq:        make(chan *protocal.ImPacket, 1024),
		ConnectStatus: 0,
	}
	return rClub
}

// RClubRun 开始运行俱乐部方式
func (rc *RClub) RClubRun() {
	defer util.RecoverPanic()

	// 开始连接
	// 一直重试，直至连上
	// for {
	rc.RClubConnect()
	if !rc.isConnected() {
		return
	}
	// 	if rc.isConnected() {
	// 		break
	// 	}
	// 	time.Sleep(10 * time.Second)
	// }

	// 开启发送队列
	go rc.sendMsgToClub()

	// 开启房间心跳
	go rc.heartBeat()

	// 开启消息接受
	go rc.receiveMsgFromClub()

}

// RClubConnect 连接俱乐部
func (rc *RClub) RClubConnect() {
	rc.Mux.Lock()
	defer rc.Mux.Unlock()

	// 防止重复连接
	if !rc.isNotConnect() {
		return
	}
	// 设置为连接中
	rc.ConnectStatus = 1

	// 开始连接
	tcpAddr, err := net.ResolveTCPAddr("tcp", rc.Remote)
	if err != nil {
		core.Logger.Error("[RClubConnect]Error:ResolveTCPAddr:%v", err.Error())
		rc.ConnectStatus = 0
		return
	}
	rc.Conn, err = net.DialTCP("[RClubConnect] tcp", nil, tcpAddr)
	if err != nil {
		core.Logger.Error("Error:DialTCP:%v", err.Error())
		rc.ConnectStatus = 0
		return
	}
	rc.ConnectStatus = 2 // 设置为已连接
	core.Logger.Info("[RClubConnect]roomId:%v, clubId:%v, remote:%v", rc.RoomId, rc.ClubId, rc.Remote)
}

func (rc *RClub) appendMsg(impacket *protocal.ImPacket) {
	rc.SendMq <- impacket
}

// 发送消息到club服务器
func (rc *RClub) sendMsgToClub() {
	defer util.RecoverPanic()
	for impacket := range rc.SendMq {
		if rc.isConnected() {
			impacket.Send(rc.Conn)
		}
	}
}

func (rc *RClub) receiveMsgFromClub() {
	defer util.RecoverPanic()
	defer func() {
		if rc.isConnected() {
			rc.ConnectStatus = 0
			rc.Conn.Close()
		}
	}()

	for {
		// 读取包内容
		impacket, err := protocal.ReadPacket(rc.Conn)
		if err != nil {
			rc.ConnectStatus = 0
			break
		}
		switch impacket.GetPackage() {
		}
	}
}

// 心跳
func (rc *RClub) heartBeat() {

}

// 发送重载消息
func (rc *RClub) reload() {
	core.Logger.Info("[rc.reload]roomId:%v,clubId:%v", rc.RoomId, rc.ClubId)
}

// 用户加入房间
func (rc *RClub) join(userId int) {
	core.Logger.Info("[rc.join]roomId:%v,clubId:%v, userId:%v", rc.RoomId, rc.ClubId, userId)
}

// 用户退出
func (rc *RClub) quit(userId int) {
	core.Logger.Info("[rc.quit]roomId:%v,clubId:%v, userId:%v", rc.RoomId, rc.ClubId, userId)
}

// 房间解散
func (rc *RClub) dismiss() {
	core.Logger.Info("[rc.dismiss]roomId:%v,clubId:%v", rc.RoomId, rc.ClubId)
}

// 房间开始
func (rc *RClub) start() {
	core.Logger.Info("[rc.start]roomId:%v,clubId:%v", rc.RoomId, rc.ClubId)
}
