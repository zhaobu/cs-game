package robot

import (
	"fmt"
	"io"
	"net"
	"sync"

	"mahjong.go/fbs/Common"

	"github.com/fwhappy/util"
	"mahjong.go/mi/protocal"
	"mahjong.go/mi/suggest"
)

var (
	huTimes    int // 胡牌次数
	roundTimes int // 总局数
)

// Robot 机器人配置
type Robot struct {
	UserId            int          // 用户id
	RoomId            int64        // 房间id
	MType             int          // 麻将类型
	CType             int          // 创建类型
	TileCnt           int          // 麻将张数
	RoomNumber        string       // 房间号
	Index             int          // 房间内索引
	HandTileList      []int        // 手牌列表(仅自己)
	DiscardTileList   []int        // 弃牌列表(所有人)
	ShowTileList      [][]int      // 明牌列表(所有人)
	Lack              int          // 定的缺
	Conn              *net.TCPConn // 用户socket连接
	Mux               *sync.Mutex  // 用户锁
	QuitChan          chan int     // 退出信号
	IsQuit            bool         // 是否已退出
	QuitOnce          sync.Once    // 只执行一次退出
	Round             int          // 当前局
	TRound            int          // 总局数
	CreateTime        int64        // 机器人创建时间
	LastHeartBeatTime int64        // 上次心跳时间

	NextNetworkTime        int64 // 下次网络广播时间
	LastCheckSayTime       int64 // 上次检查说话时间
	LastSayTime            int64 // 上次说话时间
	LastCheckUrgeTime      int64 // 上次检查“催人”时间
	LastOtherOperationTime int64 // 最后收到“其他人”操作时间

	DismissInterval int // 消息回复间隔
	DismissRandom   int // 消息回复间隔，随机值
	AILevel         int // AI级别；0：普通；1：中等；2：高级

	// 选牌器
	ms *suggest.MSelector

	// 机器人启动参数
	GameInfo *GameInfo
}

func NewRobot(userId int) *Robot {
	robot := &Robot{}
	robot.UserId = userId
	robot.QuitChan = make(chan int)
	robot.Mux = &sync.Mutex{}
	robot.AILevel = 2 // 默认高级AI
	robot.IsQuit = false
	robot.ms = suggest.NewMSelector()

	// 铂金级别的ai
	robot.ms.SetAILevel(suggest.AI_PLATINUM)
	// 创建时间
	robot.CreateTime = util.GetTime()

	return robot
}

// 启动机器人
func (this *Robot) Run() {
	// 连接服务器
	var err error
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", this.GameInfo.Remote)
	if err != nil {
		this.show("Error:ResolveTCPAddr:%s", err.Error())
		return
	}
	this.Conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		this.show("Error:DialTCP:%s", err.Error())
		return
	}
	defer func() {
		this.Conn.Close()
	}()

	this.trace("socket connected, userId:%d", this.UserId)

	// 开启消息监听
	go this.OnMessageReceive()

	// 握手
	this.HandShake()

	this.show("启动机器人,userId:%v, GradeId:%v, AILevel:%v", this.UserId, this.GameInfo.GradeId, this.AILevel)

	// 监听机器人退出
	code := <-this.QuitChan
	switch code {
	case 1:
		this.show("执行完成或者房间解散，退出:%d", this.UserId)
	case 2:
		this.show("消息异常，退出:%d", this.UserId)
	case 3:
		this.show("进入了错误的房间，退出:%d", this.UserId)
	case 4:
		this.show("加入房间错误，退出:%d", this.UserId)
	case 6:
		this.show("用户牌错误，退出:%d", this.UserId)
	case 7:
		this.show("联赛报名结束，退出:%d", this.UserId)
	case 8:
		this.show("接受操作超时，退出:%d", this.UserId)
	default:
		this.show("unkown code,userId:%d, code:%d", this.UserId, code)
	}
}

// 模拟机器人的网络变化
func (this *Robot) network(chatId int16) {
	this.chat(chatId, "")
	this.trace("发送网络信号, userId:%v, chatId:%v", this.UserId, chatId)
}

// 机器人接收消息
func (this *Robot) OnMessageReceive() {
	// 捕获异常
	defer util.RecoverPanic()

	for {
		// 读取包内容
		impacket, err := protocal.ReadPacket(this.Conn)
		if this.IsQuit {
			// 机器人已退出
			break
		}
		// 检查解析错误
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				fmt.Println("disconnected, userId:", this.UserId)
				// os.Exit(0)
			} else {
				// 协议解析错误
				fmt.Println(err.Error())
			}

			// 解析协议错误
			this.quit(2)
			break
		}

		packageId := impacket.GetPackage()
		switch packageId {
		case protocal.PACKAGE_TYPE_HANDSHAKE: // 握手成功
			this.handleHandShake(impacket)
		case protocal.PACKAGE_TYPE_HEARTBEAT: // 心跳回复
			this.handleHeartBeat(impacket)
		case protocal.PACKAGE_TYPE_DATA: // 数据包
			this.handleData(impacket)
		case Common.CommandGameRestorePush: // 重连
		default:
			break
		}
	}
}

// 处理数据包
func (this *Robot) handleData(impacket *protocal.ImPacket) {
	switch impacket.GetMessageId() {
	case Common.CommandJoinRoomResponse: // 加入房间结果
		this.handleJoinRoom(impacket)
	case Common.CommandDismissRoomPush: // 同意解散房间
		this.handleDismissRoom(impacket)
	case Common.CommandGameEnterPush: // 游戏初始化
		this.Prepare(true)
	case Common.CommandGameInitPush: // 游戏初始化
		this.handleGameInit(impacket)
	case Common.CommandOperationPush: // 提示用户操作
		this.handleOperation(impacket)
	case Common.CommandUserOperationPush: // 用户操作
		this.handleUserOperation(impacket)
	case Common.CommandClientOperationPush: // 服务器操作
		this.handleClientOperation(impacket)
	case Common.CommandGameSettlementPush: // 单局结算
		this.handleGameSettlement(impacket)
	case Common.CommandCloseRoomPush: // 房间解散
		this.handleCloseRoom(impacket)
	case Common.CommandGameRestorePush: // 重连
		this.handleGameRestore(impacket)
	case Common.CommandLeagueApplyResponse: // 重连
		this.handleLeagueApplyResponse(impacket)
	}
}

// 退出
func (robot *Robot) quit(code int) {
	robot.QuitOnce.Do(func() {
		robot.QuitChan <- code
		robot.IsQuit = true
	})
}

// 检查概率是否命中
func (robot *Robot) checkRate(rate int) bool {
	rand := util.RandIntn(100)
	return rand < rate
}
