package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	flatbuffers "github.com/google/flatbuffers/go"
	"mahjong.go/config"
	"mahjong.go/fbs/Common"
	"mahjong.go/mi/protocal"
)

// 初始化参数
var (
	// s = flag.String("s", "g2-8980", "-s=local")
	// s = flag.String("s", "qa-8990", "-s=local")
	// s = flag.String("s", "g5-8950", "-s=local")
	//s = flag.String("s", "g6-8940", "-s=local")
	s = flag.String("s", "local", "-s=local")
	// s = flag.String("s", "connection", "-s=local")
	h string
	p string
)

// 命令翻译
var codeText = map[int]string{
	1:  "换牌",
	2:  "换牌结果",
	3:  "定缺",
	4:  "定缺结果",
	5:  "抓牌",
	6:  "杠牌",
	7:  "出牌",
	8:  "吃牌",
	9:  "碰",
	10: "明杠",
	11: "暗杠",
	12: "转弯杠",
	13: "听",
	14: "报听",
	15: "胡",
	16: "抢杠胡",
	17: "自摸",
	18: "过",
	19: "取消",
	20: "显示手牌",
	21: "翻牌鸡",
	22: "前后鸡",
	23: "责任鸡",
	24: "冲锋鸡",
	25: "冲锋乌骨",
	26: "杠上开花",
	27: "热炮",
	28: "憨包杠",
	29: "已定缺",
	30: "滚筒鸡",
	31: "推荐",
	32: "补花",
	33: "分数倍数",
}

// 命令行提示
var operateHelpText = map[string]string{
	"握手":      "1.userId",
	"心跳":      "3",
	"下线":      "5 or cancel",
	"随机组局":    "4.18.gType",
	"加入比赛":    "4.36.100(100~103)",
	"创建房间":    "4.2.游戏类型.局数.俱乐部id",
	"加入房间":    "4.12.number",
	"退出房间":    "4.22",
	"解散房间":    "4.3.操作类型,-1:申请;0:同意;1:拒绝",
	"游戏准备":    "4.7",
	"麻将操作":    "4.19.operationCode.tile, 可输入help op 查看所有的操作列表",
	"聊天":      "4.24.chatId.memberId.content",
	"托管":      "4.35.0|1",
	"观察房间":    "4.37.number",
	"结束房间":    "4.38",
	"加入金币场房间": "4.40.coinType.gType.lastRoomId",
	"查看位置信息":  "4.41",
	"手动重连":    "4.30.seq",
	"重连完成":    "4.33",
	"H5创建房间":  "100.create.userId.clubId.gType.round",
	"H5解散房间":  "100.dismiss.userId.clubId.roomId",
	"H5加入房间":  "100.jr.token.number, 不需要验证token的环境,token直接传userId",
	"查看统计消息":  "100.stat",
	"查看房间信息":  "100.roomInfo",
	"查看游戏信息":  "100.gameDetail",
	"设置心跳标志":  "100.hf.1",
	"排位赛":     "4.44",
}

var setting = []int{
	1,   // 0满堂鸡
	0,   // 1连庄
	0,   // 2上下鸡
	1,   // 3乌骨鸡
	0,   // 4前后鸡
	0,   // 5星期鸡
	1,   // 6意外鸡
	0,   // 7吹分鸡
	0,   // 8滚筒鸡
	4,   // 9麻将人数
	108, // 10麻将张数
	0,   // 11本鸡
	0,   // 12站鸡
	0,   // 13翻倍鸡
	0,   // 14首圈冲锋鸡
	0,   // 15清一色奖励三分
	0,   // 16自摸翻倍
	0,   // 17自摸加1分
	0,   // 18通三
	0,   // 19大牌翻倍
	0,   // 20：包杠
	0,   // 21：爬坡鸡
	0,   // 22：查缺不查叫
	0,   // 23: 见7挖
	0,   // 24: 高挖弹
	0,   // 25: 龙七对奖3分
	0,   // 26: 最后一局翻倍
	3,   // 27:换3张
	0,   // 28:换4张
}

var (
	globalUserId int    // 当前连接用户id
	globalRoomId uint64 // 已加入房间id
	globalRound  uint8  // 当前游戏round
)

func init() {
	// 解析url参数
	flag.Parse()

	switch *s {
	case "connection":
		h = "127.0.0.1"
		p = "9000"
	case "local":
		h = "0.0.0.0"
		p = "9090"
	case "qa-9000":
		h = "114.55.227.47"
		p = "8999"
	case "qa-8999":
		h = "114.55.227.47"
		p = "8999"
	case "qa-8998":
		h = "114.55.227.47"
		p = "8998"
	case "qa-8990":
		h = "114.55.227.47"
		p = "8990"
	case "g1-9000":
		h = "118.178.190.132"
		p = "9000"
	case "g1-8999":
		h = "118.178.190.132"
		p = "8999"
	case "g2-8980":
		h = "101.37.226.22"
		p = "8980"
	case "g3-9000":
		h = "118.178.127.24"
		p = "9000"
	case "g3-8999":
		h = "118.178.127.24"
		p = "8999"
	case "g5-8950":
		h = "118.31.183.211"
		p = "8950"
	case "g6-8940":
		h = "101.37.224.155"
		p = "8940"
	default:
		fmt.Println("server not supported!")
		os.Exit(-1)
	}
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", h+":"+p)
	if err != nil {
		fmt.Println("Error:ResolveTCPAddr:", err.Error())
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Error:DialTCP:", err.Error())
		return
	}
	defer conn.Close()

	fmt.Println(fmt.Sprintf("connect at: %s:%s", h, p))

	// 定义接收消息的协程
	go onMessageRecived(conn)

	// 控制台接收输入
	onMessageInput(conn)
}

// 接受控制台输入，并发送消息给服务器
func onMessageInput(conn *net.TCPConn) {
	for {
		var msg string

		msgReader := bufio.NewReader(os.Stdin)
		msg, _ = msgReader.ReadString('\n')
		msg = strings.TrimSuffix(msg, "\n")

		// 跳过空数据包
		if len(msg) == 0 {
			continue
		} else if msg == "quit" || msg == "exit" {
			break
		} else if msg == "cancel" {
			conn.Close()
		} else if msg == "help" {
			for k, v := range operateHelpText {
				fmt.Println(k, ":", v)
			}
			continue
		} else if msg == "help op" {
			for i := 1; i <= len(codeText); i++ {
				fmt.Println(i, ":", codeText[i])
			}
			continue
		}

		// 分隔消息
		receiveMessageSplit := strings.SplitN(string(msg), ".", -1)

		// 包id
		pId, _ := strconv.Atoi(receiveMessageSplit[0])
		packageId := uint8(pId)
		// 包体
		// 生成包体
		// body := make(map[string]interface{})
		js := simplejson.New()
		// 消息内容
		var message []byte

		switch packageId {
		case protocal.PACKAGE_TYPE_HANDSHAKE: // 握手，eg：1|1
			globalUserId, _ = strconv.Atoi(receiveMessageSplit[1])
			user := make(map[string]interface{})
			user["token"] = util.GenToken(globalUserId, "latest", config.TOKEN_SECRET_KEY)
			if len(receiveMessageSplit) > 2 {
				// 纬度
				user["lat"], _ = strconv.ParseFloat(receiveMessageSplit[3], 64)
				// 经度
				user["lng"], _ = strconv.ParseFloat(receiveMessageSplit[2], 64)
			}
			user["no_heartbeat"] = 1
			js.Set("user", user)
			message, _ = js.Encode()
		case protocal.PACKAGE_TYPE_HANDSHAKE_ACK: // 握手回复，eg：
			message, _ = js.Encode()
		case protocal.PACKAGE_TYPE_HEARTBEAT: // 心跳，eg：3
			message, _ = js.Encode()
		case protocal.PACKAGE_TYPE_DATA: // 数据包
			mId, _ := strconv.Atoi(receiveMessageSplit[1])
			var mType int
			var buf []byte

			switch mId {
			case Common.CommandRandomRoomRequest: // eg: 4|18|0
				// 随机加入房间
				mType = protocal.MSG_TYPE_REQUEST
				gType, _ := strconv.Atoi(receiveMessageSplit[2]) // 游戏类型
				builder := flatbuffers.NewBuilder(0)
				Common.RandomRoomRequestStart(builder)
				Common.RandomRoomRequestAddGameType(builder, uint16(gType))
				orc := Common.RandomRoomRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandMatchRoomRequest: // eg: 4|36|1001
				// 随机加入房间
				mType = protocal.MSG_TYPE_REQUEST
				gType, _ := strconv.Atoi(receiveMessageSplit[2]) // 游戏类型
				builder := flatbuffers.NewBuilder(0)
				Common.MatchRoomRequestStart(builder)
				Common.MatchRoomRequestAddGameType(builder, uint16(gType))
				orc := Common.MatchRoomRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandCreateRoomRequest:
				// 创建房间
				mType = protocal.MSG_TYPE_REQUEST
				gType, _ := strconv.Atoi(receiveMessageSplit[2]) // 游戏类型
				round, _ := strconv.Atoi(receiveMessageSplit[3]) // 游戏局数
				// 俱乐部id
				var clubId int
				if len(receiveMessageSplit) >= 5 {
					clubId, _ = strconv.Atoi(receiveMessageSplit[4])
				}
				switch gType {
				case Common.GameTypeMAHJONG_LD:
					setting[9] = 2
				case Common.GameTypeMAHJONG_SD:
					setting[9] = 3
				case Common.GameTypeMAHJONG_STT:
					setting[9] = 3
					setting[10] = 72
				default:
					break
				}

				builder := flatbuffers.NewBuilder(0)
				var settingBinary flatbuffers.UOffsetT
				Common.CreateRoomRequestStartSettingVector(builder, len(setting))
				for i := len(setting) - 1; i >= 0; i-- {
					builder.PrependByte(byte(setting[i]))
				}
				settingBinary = builder.EndVector(len(setting))

				Common.CreateRoomRequestStart(builder)
				Common.CreateRoomRequestAddGameType(builder, uint16(gType))
				Common.CreateRoomRequestAddRound(builder, uint8(round))
				Common.CreateRoomRequestAddSetting(builder, settingBinary)
				Common.CreateRoomRequestAddClubId(builder, int32(clubId))
				orc := Common.CreateRoomRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandJoinRoomRequest:
				// 加入房间
				mType = protocal.MSG_TYPE_REQUEST

				builder := flatbuffers.NewBuilder(0)
				number := builder.CreateString(receiveMessageSplit[2]) // 房间id

				Common.JoinRoomRequestStart(builder)
				Common.JoinRoomRequestAddNumber(builder, number)
				orc := Common.JoinRoomRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandQuitRoomNotify:
				// 退出房间
				mType = protocal.MSG_TYPE_NOTIFY

				builder := flatbuffers.NewBuilder(0)
				Common.QuitRoomNotifyStart(builder)
				orc := Common.QuitRoomNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandDismissRoomNotify:
				// 解散房间
				mType = protocal.MSG_TYPE_NOTIFY

				op, _ := strconv.Atoi(receiveMessageSplit[2]) // 操作类型：-1 申请 0 同意 1拒绝
				builder := flatbuffers.NewBuilder(0)
				Common.DismissRoomNotifyStart(builder)
				Common.DismissRoomNotifyAddOp(builder, int8(op))
				orc := Common.DismissRoomNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandGameReadyNotify:
				// 游戏准备
				mType = protocal.MSG_TYPE_NOTIFY

				builder := flatbuffers.NewBuilder(0)
				Common.GameReadyNotifyStart(builder)
				if len(receiveMessageSplit) == 2 {
					// 只有2个参数的时候，表示准备
					Common.GameReadyNotifyAddReadying(builder, uint8(1))
				} else {
					// 3个参数，表示有经纬度
					lng, _ := strconv.ParseFloat(receiveMessageSplit[2], 64)
					lat, _ := strconv.ParseFloat(receiveMessageSplit[3], 64)
					Common.GameReadyNotifyAddLng(builder, lng)
					Common.GameReadyNotifyAddLat(builder, lat)
				}
				orc := Common.GameReadyNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandUserOperationNotify:
				// 用户操作
				mType = protocal.MSG_TYPE_NOTIFY

				operation, _ := strconv.Atoi(receiveMessageSplit[2])
				tiles := receiveMessageSplit[3:]
				tileCnt := len(tiles)

				builder := flatbuffers.NewBuilder(0)
				var tilesBinary flatbuffers.UOffsetT
				Common.OperationStartTilesVector(builder, tileCnt)
				for i := tileCnt - 1; i >= 0; i-- {
					tile, _ := strconv.Atoi(tiles[i])
					builder.PrependByte(byte(byte(tile)))
				}
				tilesBinary = builder.EndVector(tileCnt)

				// 构建一个operation对象
				Common.OperationStart(builder)
				Common.OperationAddOp(builder, byte(operation))
				Common.OperationAddTiles(builder, tilesBinary)
				op := Common.OperationEnd(builder)

				Common.UserOperationNotifyStart(builder)
				Common.UserOperationNotifyAddOp(builder, op)
				orc := Common.UserOperationNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandRoomChatNotify: // 游戏聊天
				mType = protocal.MSG_TYPE_NOTIFY
				chatId, _ := strconv.Atoi(receiveMessageSplit[2])
				memberId, _ := strconv.Atoi(receiveMessageSplit[3])
				var content string
				if len(receiveMessageSplit) > 4 {
					content = receiveMessageSplit[4]
				}

				builder := flatbuffers.NewBuilder(0)
				str := builder.CreateString(content)
				Common.RoomChatNotifyStart(builder)
				Common.RoomChatNotifyAddChatId(builder, int16(chatId))
				Common.RoomChatNotifyAddMemberId(builder, uint8(memberId))
				Common.RoomChatNotifyAddContent(builder, str)
				orc := Common.RoomChatNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandGameRestoreNotify: // 请求重连
				mType = protocal.MSG_TYPE_NOTIFY

				seq, _ := strconv.Atoi(receiveMessageSplit[2])
				builder := flatbuffers.NewBuilder(0)
				Common.GameRestoreNotifyStart(builder)
				Common.GameRestoreNotifyAddRoomId(builder, globalRoomId)
				Common.GameRestoreNotifyAddRound(builder, uint16(globalRound))
				Common.GameRestoreNotifyAddStep(builder, uint16(seq))
				orc := Common.GameRestoreNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandGameRestoreDoneNotify: // 重连完成
				mType = protocal.MSG_TYPE_NOTIFY
				builder := flatbuffers.NewBuilder(0)
				Common.GameRestoreDoneNotifyStart(builder)
				orc := Common.GameRestoreDoneNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandGameHostingNotify: // 托管, 4.35.0|1
				mType = protocal.MSG_TYPE_NOTIFY
				hostingStatus, _ := strconv.Atoi(receiveMessageSplit[2])
				builder := flatbuffers.NewBuilder(0)
				Common.GameHostingNotifyStart(builder)
				Common.GameHostingNotifyAddHostingStatus(builder, uint8(hostingStatus))
				orc := Common.GameHostingNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandObRoomRequest: // 观察房间, 4.37.number
				mType = protocal.MSG_TYPE_REQUEST
				builder := flatbuffers.NewBuilder(0)
				number := builder.CreateString(receiveMessageSplit[2]) // 房间号
				Common.ObRoomRequestStart(builder)
				Common.ObRoomRequestAddNumber(builder, number)
				orc := Common.ObRoomRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandEndRoomNotify: // 结束房间 4.38
				// 退出房间
				mType = protocal.MSG_TYPE_NOTIFY
				builder := flatbuffers.NewBuilder(0)
				Common.GameEndNotifyStart(builder)
				orc := Common.GameEndNotifyEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandGeneralRequest:
				var d flatbuffers.UOffsetT
				mType = protocal.MSG_TYPE_REQUEST
				builder := flatbuffers.NewBuilder(0)
				act := builder.CreateString(receiveMessageSplit[2])
				tmp := make(map[string]interface{})
				tmp["nickname"] = "haha"
				tmp["user"] = map[string]interface{}{"avatar": "www.baidu.com"}
				b, _ := json.Marshal(tmp)
				Common.GeneralRequestStartDataVector(builder, len(b))
				for i := len(b) - 1; i >= 0; i-- {
					builder.PrependByte(byte(b[i]))
				}
				d = builder.EndVector(len(b))

				Common.GeneralRequestStart(builder)
				Common.GeneralRequestAddAct(builder, act)
				Common.GeneralRequestAddData(builder, d)
				orc := Common.GeneralRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandCoinRoomRequest:
				mType = protocal.MSG_TYPE_REQUEST
				builder := flatbuffers.NewBuilder(0)
				coinType, _ := strconv.Atoi(receiveMessageSplit[2])
				gType, _ := strconv.Atoi(receiveMessageSplit[3])
				lastRoomId, _ := strconv.Atoi(receiveMessageSplit[4])
				Common.CoinRoomRequestStart(builder)
				Common.CoinRoomRequestAddCoinType(builder, byte(coinType))
				Common.CoinRoomRequestAddGameType(builder, uint16(gType))
				Common.CoinRoomRequestAddLastRoomId(builder, uint64(lastRoomId))
				orc := Common.CoinRoomRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandRankRoomRequest:
				mType = protocal.MSG_TYPE_REQUEST
				builder := flatbuffers.NewBuilder(0)
				Common.RankRoomRequestStart(builder)
				orc := Common.RankRoomRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			case Common.CommandGameUserDistanceRequest:
				mType = protocal.MSG_TYPE_REQUEST
				builder := flatbuffers.NewBuilder(0)
				Common.GameUserDistanceRequestStart(builder)
				orc := Common.GameUserDistanceRequestEnd(builder)
				builder.Finish(orc)
				buf = builder.FinishedBytes()
			default:
				break
			}

			message = protocal.NewImMessage(uint16(mId), uint16(mType), uint16(0), uint16(123), buf)

			break
		case protocal.PACKAGE_TYPE_KICK: // 踢下线
			message, _ = js.Encode()
			break
		case protocal.PACKAGE_TYPE_SYSTEM: // 服务器指令
			js.Set("systemKey", config.SYSTEM_KEY)
			act := receiveMessageSplit[1]
			switch act {
			case "jr":
				userId, _ := strconv.Atoi(receiveMessageSplit[2])
				user := map[string]interface{}{
					"token":        util.GenToken(userId, "latest", config.TOKEN_SECRET_KEY),
					"lat":          float64(-2),
					"lng":          float64(-2),
					"device":       "android",
					"device_token": "testdevicetoken",
					"ip":           "127.0.0.2",
				}
				number := receiveMessageSplit[3]
				js.Set("user", user)
				js.Set("number", number)
			case "stat":
				v, _ := strconv.Atoi(receiveMessageSplit[2])
				js.Set("v", v)
			case "roomInfo":
				fallthrough
			case "gameDetail":
				roomId, _ := strconv.Atoi(receiveMessageSplit[2])
				js.Set("roomId", int64(roomId))
			case "create":
				userId, _ := strconv.Atoi(receiveMessageSplit[2])
				clubId, _ := strconv.Atoi(receiveMessageSplit[3])
				gType, _ := strconv.Atoi(receiveMessageSplit[4])
				round, _ := strconv.Atoi(receiveMessageSplit[5])
				js.Set("userId", userId)
				js.Set("clubId", clubId)
				js.Set("gType", gType)
				js.Set("round", round)
				js.Set("setting", setting)
			case "dismiss":
				userId, _ := strconv.Atoi(receiveMessageSplit[2])
				clubId, _ := strconv.Atoi(receiveMessageSplit[3])
				roomId, _ := strconv.Atoi(receiveMessageSplit[4])
				js.Set("userId", userId)
				js.Set("clubId", clubId)
				js.Set("roomId", int64(roomId))
			case "hf":
				f, _ := strconv.Atoi(receiveMessageSplit[2])
				js.Set("f", f)
			}
			js.Set("act", act)
			message, _ = js.Encode()
			break
		default:
			fmt.Printf("错误的数据包id:%d\n", packageId)
			break
		}

		// 发送消息给服务器
		imPacket := protocal.NewImPacket(packageId, message)
		// fmt.Println("消息发送开始, t: ", util.GetMicrotime(), ", timestamp:", util.GetTimestamp())
		conn.Write(imPacket.Serialize())

		// fmt.Println("消息发送成功, t: ", util.GetMicrotime(), ", timestamp:", util.GetTimestamp())
	}
}

func onMessageRecived(conn *net.TCPConn) {
	for {
		// 读取包内容
		impacket, err := protocal.ReadPacket(conn)

		// 检查解析错误
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				fmt.Println("disconnected")
				os.Exit(0)
			} else {
				// 协议解析错误
				fmt.Println(err.Error())
			}
			break
		}

		if impacket.GetPackage() == protocal.PACKAGE_TYPE_KICK {
			fmt.Println("重复登录，踢下线")
			os.Exit(0)
		} else if impacket.GetPackage() == protocal.PACKAGE_TYPE_HANDSHAKE {
			// 握手成功
			js, _ := simplejson.NewJson(impacket.GetMessage())
			vmap, _ := js.Map()
			fmt.Printf("%#v\n", vmap)

			// 发送ack
			message, _ := simplejson.New().Encode()
			imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, message)
			conn.Write(imPacket.Serialize())

			fmt.Println("handShakeAck....")

			// 开启心跳
			timeInterval, err := js.Get("heartbeat").Int()
			if err != nil {
				timeInterval = 3
			}
			fmt.Println("handShakeAck....", timeInterval)
			go heartBeat(conn, timeInterval)

		} else if impacket.GetPackage() == protocal.PACKAGE_TYPE_HEARTBEAT {
			// fmt.Println("[", util.GetTimestamp(), "]", "收到心跳，长度：", len(impacket.Serialize()))
		} else if impacket.IType == protocal.PACKAGE_TYPE_NONE_MESSAGE_ID {
			js, _ := simplejson.NewJson(impacket.GetMessage())
			vmap, _ := js.Map()
			fmt.Printf("packageId====%#v\n", impacket.GetPackage())
			fmt.Printf("%#v\n", vmap)
		} else {
			//  fmt.Println("消息接受成功, t: ", util.GetMicrotime(), ", timestamp:", util.GetTimestamp())
			// packageId := impacket.GetPackage()
			// mId := impacket.GetMessageId()
			// mType := impacket.GetMessageType()
			// mNumber := impacket.GetMessageNumber()
			body := impacket.GetBody()
			seq := impacket.GetMessageIndex()

			switch impacket.GetMessageId() {
			case Common.CommandJoinRoomResponse:
				code := Common.GetRootAsJoinRoomResponse(body, 0).S2cResult(nil).Code()
				if code < 0 {
					fmt.Printf("joinRoom error, code:%v, msg:%v\n",
						code, string(Common.GetRootAsJoinRoomResponse(body, 0).S2cResult(nil).Msg()))
				} else {
					// 记录roomId
					globalRoomId = Common.GetRootAsJoinRoomResponse(body, 0).RoomInfo(nil).RoomId()
					fmt.Println(
						"joinRoom,",
						"roomId:", Common.GetRootAsJoinRoomResponse(body, 0).RoomInfo(nil).RoomId(),
						", number:", string(Common.GetRootAsJoinRoomResponse(body, 0).RoomInfo(nil).Number()),
						", round:", Common.GetRootAsJoinRoomResponse(body, 0).RoomInfo(nil).Round(),
						",cType:", Common.GetRootAsJoinRoomResponse(body, 0).RoomInfo(nil).RandomRoom(),
					)
				}

			case Common.CommandJoinRoomPush:
				fmt.Printf("加入房间, 用户ID: %d\n", Common.GetRootAsJoinRoomPush(body, 0).RoomUserInfo(nil).UserId())
				fmt.Printf("加入房间, 用户Index: %d\n", Common.GetRootAsJoinRoomPush(body, 0).RoomUserInfo(nil).Index())

				break
			case Common.CommandQuitRoomPush:
				fmt.Printf("退出房间, 用户Id: %d, code:%d\n", Common.GetRootAsQuitRoomPush(body, 0).UserId(), Common.GetRootAsQuitRoomPush(body, 0).Code())

				break
			case Common.CommandDismissRoomPush:
				userId := Common.GetRootAsDismissRoomPush(body, 0).UserId()
				op := Common.GetRootAsDismissRoomPush(body, 0).Op()
				fmt.Printf("解散房间操作, userId:%d, op:%d\n", userId, op)

				break
			case Common.CommandCloseRoomPush:
				fmt.Println("房间已解散, code:", Common.GetRootAsCloseRoomPush(body, 0).Code(), ", 原因:", string(Common.GetRootAsCloseRoomPush(body, 0).Msg()))

				break
			case Common.CommandGameEnterPush: // 游戏进入
				fmt.Println("game enter...")
				break
			case Common.CommandGameInitPush: // 游戏初始化
				// 记录当前游戏局数据
				globalRound = Common.GetRootAsGameInitPush(body, 0).CurrentRound()

				fmt.Println("dealer:", Common.GetRootAsGameInitPush(body, 0).Dealer())
				fmt.Println("dealCount:", Common.GetRootAsGameInitPush(body, 0).DealCount())
				fmt.Println("dice:", Common.GetRootAsGameInitPush(body, 0).DiceBytes())
				fmt.Println("tiles:", Common.GetRootAsGameInitPush(body, 0).TilesBytes())
				break
			case Common.CommandGameReadyPush: // 用户准备
				fmt.Println("用户已准备:", Common.GetRootAsGameReadyPush(body, 0).UserId())
				break
			case Common.CommandOperationPush: // 推送用户可进行的操作
				fmt.Println("------------recieve operation push[seq:", seq, "]----------")
				response := Common.GetRootAsOperationPush(body, 0)
				op := new(Common.Operation)
				for i := 0; i < response.OpLength(); i++ {
					response.Op(op, i)
					fmt.Println("可进行的操作:", codeText[int(op.Op())], "[", op.Op(), "]", ",牌:", op.TilesBytes())
				}
				break
			case Common.CommandUserOperationPush: // 推送用户可进行的操作
				fmt.Println("------------recieve user operation push[seq:", seq, "]----------")
				response := Common.GetRootAsUserOperationPush(body, 0)
				var op *Common.Operation
				op = response.Op(op)

				fmt.Printf("用户操作, opCode:%v,  userId:%v,tiles:%v.\n", codeText[int(op.Op())], response.UserId(), op.TilesBytes())
			case Common.CommandRoomChatPush:
				Common.GetRootAsRoomChatPush(body, 0).UserId()
				userId := Common.GetRootAsRoomChatPush(body, 0).UserId()
				chatId := Common.GetRootAsRoomChatPush(body, 0).ChatId()
				content := string(Common.GetRootAsRoomChatPush(body, 0).Content())

				// fmt.Println("用户id:", Common.GetRootAsRoomChatPush(body, 0).UserId())
				// fmt.Println("chatId:", Common.GetRootAsRoomChatPush(body, 0).ChatId())
				// fmt.Println("memberId:", Common.GetRootAsRoomChatPush(body, 0).MemberId())
				if chatId == config.CHAT_ID_SIGNAL_VERY_STRONGER ||
					chatId == config.CHAT_ID_SIGNAL_STRONGER ||
					chatId == config.CHAT_ID_SIGNAL_NORMAL ||
					chatId == config.CHAT_ID_SIGNAL_WEAK ||
					chatId == config.CHAT_ID_SIGNAL_VERY_WEAK {
					// do nothing
					fmt.Println("[chat]userId:", userId, ",信号强度:", chatId)
				} else {
					fmt.Println("[chat]userId:", userId, ",chatId:", chatId,
						",content:", content)
				}

			case Common.CommandUserOnlinePush:
				fmt.Printf("userId: %d, online:%d\n", Common.GetRootAsUserOnlinePush(body, 0).UserId(), Common.GetRootAsUserOnlinePush(body, 0).Online())

				break
			case Common.CommandUpdateMoneyPush:
				fmt.Printf("钻石变化: %d.\n", Common.GetRootAsUpdateMoneyPush(body, 0).Amount())
				break
			case Common.CommandClientOperationPush:
				fmt.Println("------------recieve client operation push[seq:", seq, "]----------")

				response := Common.GetRootAsClientOperationPush(body, 0)
				operation := new(Common.Operation)
				operation = response.Op(operation)

				fmt.Printf("客户端操作,  opCode:%v,userId:%v, tiles:%v.\n", codeText[int(operation.Op())], response.UserId(), operation.TilesBytes())

				break
			case Common.CommandGameRestorePush:
				response := Common.GetRootAsGameRestorePush(body, 0)
				fmt.Println("结果: ", response.S2cResult(nil).Code())
				gameplayState := new(Common.GamePlayState)
				gameplayState = response.GameplayState(gameplayState)

				// 记录房间id和局数
				globalRoomId = gameplayState.RoomInfo(nil).RoomId()
				globalRound = gameplayState.CurrentRound()
				fmt.Printf("globalUserId:%v, globalRoomId:%v, globalRound:%v\n", globalUserId, globalRoomId, globalRound)

				roomUserInfo := new(Common.RoomUserInfo)
				for i := 0; i < gameplayState.RoomUserListLength(); i++ {
					gameplayState.RoomUserList(roomUserInfo, i)
					fmt.Printf("roomUserInfo, id:%v, 历史积分:%v\n", roomUserInfo.UserId(), roomUserInfo.Score())
				}

				fmt.Println("seq:", gameplayState.Step())
				fmt.Println("游戏状态: ", int(gameplayState.GameStatus()), ",当前局: ", int(gameplayState.CurrentRound()), ",step:", gameplayState.Step())
				fmt.Println("剩余牌数: ", int(gameplayState.WallTileCount()), ",前面抓张数: ", int(gameplayState.DrawFront()), ",后面抓张数 ", int(gameplayState.DrawBehind()))
				fmt.Println("骰子: ", int(gameplayState.Dice(0)), int(gameplayState.Dice(1)), ",庄家: ", int(gameplayState.Dealer()), ",连庄数量: ", int(gameplayState.DealCount()))
				fmt.Println("最后打牌者:", gameplayState.LastPlayerId(), ",当前操作者:", gameplayState.CurrentPlayerId())
				fmt.Println("翻牌鸡: ", int(gameplayState.ChikenDraw()), ",前后鸡: ", int(gameplayState.ChikenRoller()), ",前后鸡位置: ", int(gameplayState.ChikenRollerIndex()))
				fmt.Println("准备状态: ", int(gameplayState.PrepareStatus()), ",定缺状态: ", int(gameplayState.LackStatus()))
				fmt.Println("听的牌: ", gameplayState.TingTilesBytes())
				for i := 0; i < gameplayState.LackedUsersLength(); i++ {
					fmt.Println("已定缺用户:", gameplayState.LackedUsers(i))
				}
				for i := 0; i < gameplayState.PrepareUsersLength(); i++ {
					fmt.Println("已准备用户:", gameplayState.PrepareUsers(i))
				}
				fmt.Println("解散房间剩余时间:", int(gameplayState.DismissRemainTime()))

				mahjongUserInfo := new(Common.MahjongUserInfo_v_2_1_0)
				showcardInfo := new(Common.ShowCard_v_2_1_0)
				for i := 0; i < gameplayState.MahjongUserInfoV210Length(); i++ {
					gameplayState.MahjongUserInfoV210(mahjongUserInfo, i)
					fmt.Println("-------------------------------------------牌局用户id：", mahjongUserInfo.UserId())
					fmt.Println(
						"冲锋幺鸡: ", int(mahjongUserInfo.ChikenChargeBam1()), ",冲锋乌骨鸡: ", int(mahjongUserInfo.ChikenChargeDot8()),
						",冲锋乌骨鸡: ", int(mahjongUserInfo.ChikenChargeDot8()), ",责任鸡:", int(mahjongUserInfo.ChikenResponsibility()),
					)
					fmt.Println("局内分数:", int(mahjongUserInfo.GameScore()), ",报听:", int(mahjongUserInfo.BaoTing()), ",缺:", int(mahjongUserInfo.LackTile()))
					fmt.Println("手牌数量：", mahjongUserInfo.HandTilesCount(), ",牌：", mahjongUserInfo.HandTilesBytes())
					fmt.Println("牌局用户弃牌：", mahjongUserInfo.PlayListBytes())
					fmt.Println("用户选择换的牌：", mahjongUserInfo.ExchangeTilesBytes())

					for j := 0; j < mahjongUserInfo.ShowCardListLength(); j++ {
						mahjongUserInfo.ShowCardList(showcardInfo, j)
						fmt.Println("明牌类型:", showcardInfo.OperationCode())
						fmt.Println("明牌:", showcardInfo.TilesBytes())
					}
				}

				// 读取用户可进行的操作
				length := gameplayState.OperationPushArrayLength()
				operationPush := new(Common.OperationPush)
				op := new(Common.Operation)
				if length > 0 {
					for i := 0; i < length; i++ {
						gameplayState.OperationPushArray(operationPush, i)
						if operationPush != nil {
							for j := 0; j < operationPush.OpLength(); j++ {
								operationPush.Op(op, j)
								fmt.Println("操作:", codeText[int(op.Op())], ",code:", int(op.Op()), ",牌:", op.TilesBytes())
							}
						}
					}
				}

				// 读取结算信息
				settlement := gameplayState.GameSettlementPush(nil)
				if settlement != nil && settlement.SettlementInfoV230Length() > 0 {
					settlementInfo := new(Common.SettlementInfo_v_2_3_0)
					scoreItem := new(Common.ScoreItem_v_2_3_0)
					for i := 0; i < settlement.SettlementInfoV230Length(); i++ {
						settlement.SettlementInfoV230(settlementInfo, i)
						fmt.Println("------------------结算用户id:", int(settlementInfo.UserId()))
						fmt.Println("winWay:", int(settlementInfo.WinWay()))
						fmt.Println("是否报听:", int(settlementInfo.BaoTing()))
						fmt.Println("winStatus:", int(settlementInfo.WinStatus()))
						fmt.Println("本局分数:", int(settlementInfo.TotalScore()))
						fmt.Println("本把累积分数:", int(settlementInfo.GameScore()))
						fmt.Println("听牌状态:", int(settlementInfo.TingStatus()))
						fmt.Println("胡牌状态:", int(settlementInfo.HuStatus()))
						fmt.Println("点炮状态:", int(settlementInfo.PaoStatus()))
						fmt.Println("结算明细=>>>>>>>>>>>")
						for j := 0; j < settlementInfo.ScoreItemsLength(); j++ {
							settlementInfo.ScoreItems(scoreItem, j)
							fmt.Println(fmt.Sprintf("group:%v,typeId:%v,Count:%v,score*scoreCount:%v*:%v,tiles:%v",
								int(scoreItem.Group()), scoreItem.TypeId(), scoreItem.Count(), scoreItem.Score(), scoreItem.ScoreCount(), scoreItem.TilesBytes()))
						}
						fmt.Println("结算明细=<<<<<<<<<<<")
					}
				}

				// 读取鸡牌信息
				if settlement != nil && settlement.ChikenInfoLength() > 0 {
					settlementChikens := new(Common.SettlementChikens)
					chikenInfo := new(Common.ChikenInfo)
					for i := 0; i < settlement.ChikenInfoLength(); i++ {
						settlement.ChikenInfo(settlementChikens, i)
						fmt.Println("--------鸡牌用户id:", int(settlementChikens.UserId()))
						// 读取用户打出去的鸡
						if settlementChikens.PlayChikensLength() > 0 {
							for j := 0; j < settlementChikens.PlayChikensLength(); j++ {
								settlementChikens.PlayChikens(chikenInfo, j)
								fmt.Println(
									"弃牌鸡",
									",tile:", int(chikenInfo.Tile()),
									",isRecharge:", int(chikenInfo.IsRecharge()),
									",isBao:", int(chikenInfo.IsBao()),
									",isGold:", int(chikenInfo.IsGold()),
									",chikenType:", int(chikenInfo.ChikenType()),
									",extra:", string(chikenInfo.Extra()),
								)
							}

						}
						// 读取用户手牌的鸡
						if settlementChikens.HandChikensLength() > 0 {
							for j := 0; j < settlementChikens.HandChikensLength(); j++ {
								settlementChikens.HandChikens(chikenInfo, j)
								fmt.Println(
									"手牌鸡",
									",tile:", int(chikenInfo.Tile()),
									",isRecharge:", int(chikenInfo.IsRecharge()),
									",isBao:", int(chikenInfo.IsBao()),
									",isGold:", int(chikenInfo.IsGold()),
									",chikenType:", int(chikenInfo.ChikenType()),
									",extra:", string(chikenInfo.Extra()),
								)
							}

						}
						// 读取用户明牌中的鸡
						if settlementChikens.ShowCardChikensLength() > 0 {
							for j := 0; j < settlementChikens.ShowCardChikensLength(); j++ {
								settlementChikens.ShowCardChikens(chikenInfo, j)
								fmt.Println(
									"明牌鸡",
									",tile:", int(chikenInfo.Tile()),
									",isRecharge:", int(chikenInfo.IsRecharge()),
									",isBao:", int(chikenInfo.IsBao()),
									",isGold:", int(chikenInfo.IsGold()),
									",chikenType:", int(chikenInfo.ChikenType()),
									",extra:", string(chikenInfo.Extra()),
								)
							}
						}
					}
				}

				// 自动发送restoredone
				mType := protocal.MSG_TYPE_NOTIFY
				builder := flatbuffers.NewBuilder(0)
				Common.GameRestoreDoneNotifyStart(builder)
				orc := Common.GameRestoreDoneNotifyEnd(builder)
				builder.Finish(orc)
				buf := builder.FinishedBytes()
				// 发送消息给服务器
				message := protocal.NewImMessage(uint16(Common.CommandGameRestoreDoneNotify), uint16(mType), uint16(0), uint16(123), buf)
				imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
				conn.Write(imPacket.Serialize())
				break
			case Common.CommandGameRestoreSectionPush:
				response := Common.GetRootAsGameRestoreSectionPush(body, 0)

				dismissUsers := []uint32{}
				for i := 0; i < response.DismissUsersLength(); i++ {
					dismissUsers = append(dismissUsers, response.DismissUsers(i))
				}
				hostingUsers := []uint32{}
				for i := 0; i < response.HostingUsersLength(); i++ {
					hostingUsers = append(hostingUsers, response.HostingUsers(i))
				}
				fmt.Printf("托管用户列表:%v, 解散用户列表:%v, 解散剩余时间:%v\n", hostingUsers, dismissUsers, response.DismissRemainTime())
				fmt.Printf("------未处理消息队列[length:%v]-------\n", response.OperationListLength())

				seqOperation := new(Common.SeqOperation)
				operation := new(Common.Operation)
				for i := 0; i < response.OperationListLength(); i++ {
					wOperation := new(Common.OperationPush)
					uOperation := new(Common.UserOperationPush)
					cOperation := new(Common.ClientOperationPush)
					response.OperationList(seqOperation, i)
					wOperation = seqOperation.OperationPush(wOperation)
					uOperation = seqOperation.UserOperationPush(uOperation)
					cOperation = seqOperation.ClientOperationPush(cOperation)

					if wOperation != nil {
						fmt.Println("------------section restore operation push[seq:", seqOperation.Step(), "]----------")
						for i := 0; i < wOperation.OpLength(); i++ {
							wOperation.Op(operation, i)
							fmt.Println("可进行的操作:", codeText[int(operation.Op())], "[", operation.Op(), "]", ",牌:", operation.TilesBytes())
						}
					}

					if uOperation != nil {
						operation = uOperation.Op(operation)
						fmt.Println("------------section restore user operation push[seq:", seqOperation.Step(), "]----------")
						fmt.Printf("用户操作, opCode:%v,  userId:%v,tiles:%v.\n", codeText[int(operation.Op())], uOperation.UserId(), operation.TilesBytes())
					}

					if cOperation != nil {
						operation = cOperation.Op(operation)
						fmt.Println("------------section restore client operation push[seq:", seqOperation.Step(), "]----------")
						fmt.Printf("客户端操作,  opCode:%v,userId:%v, tiles:%v.\n", codeText[int(operation.Op())], cOperation.UserId(), operation.TilesBytes())
					}
				}

				// 自动发送restoredone
				mType := protocal.MSG_TYPE_NOTIFY
				builder := flatbuffers.NewBuilder(0)
				Common.GameRestoreDoneNotifyStart(builder)
				orc := Common.GameRestoreDoneNotifyEnd(builder)
				builder.Finish(orc)
				buf := builder.FinishedBytes()
				// 发送消息给服务器
				message := protocal.NewImMessage(uint16(Common.CommandGameRestoreDoneNotify), uint16(mType), uint16(0), uint16(123), buf)
				imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
				conn.Write(imPacket.Serialize())
			case Common.CommandGameSettlementPush: // 单局结算
				fmt.Println("收到单局结算信息")

				// 读取结算信息
				settlement := Common.GetRootAsGameSettlementPush(body, 0)
				if settlement != nil && settlement.SettlementInfoV230Length() > 0 {
					settlementInfo := new(Common.SettlementInfo_v_2_3_0)
					scoreItem := new(Common.ScoreItem_v_2_3_0)
					for i := 0; i < settlement.SettlementInfoV230Length(); i++ {
						settlement.SettlementInfoV230(settlementInfo, i)
						fmt.Println("------------------结算用户id:", int(settlementInfo.UserId()))
						fmt.Println("winWay:", int(settlementInfo.WinWay()))
						fmt.Println("是否报听:", int(settlementInfo.BaoTing()))
						fmt.Println("winStatus:", int(settlementInfo.WinStatus()))
						fmt.Println("总分:", int(settlementInfo.TotalScore()))
						fmt.Println("局内分数:", int(settlementInfo.GameScore()))
						fmt.Println("听牌状态:", int(settlementInfo.TingStatus()))
						fmt.Println("胡牌状态:", int(settlementInfo.HuStatus()))
						fmt.Println("点炮状态:", int(settlementInfo.PaoStatus()))
						fmt.Println("结算明细=>>>>>>>>>>>")
						for j := 0; j < settlementInfo.ScoreItemsLength(); j++ {
							settlementInfo.ScoreItems(scoreItem, j)
							fmt.Println(fmt.Sprintf("group:%v,typeId:%v,Count:%v,score*scoreCount:%v*:%v,tiles:%v",
								int(scoreItem.Group()), scoreItem.TypeId(), scoreItem.Count(), scoreItem.Score(), scoreItem.ScoreCount(), scoreItem.TilesBytes()))
						}
						fmt.Println("结算明细=<<<<<<<<<<<")
					}
				}

			case Common.CommandGameHostingPush: // 用户托管
				response := Common.GetRootAsGameHostingPush(body, 0)
				fmt.Println("切换托管状态:", response.HostingStatus())
			case Common.CommandGameAntiCheatingPush: // 防作弊
				response := Common.GetRootAsGameAntiCheatingPush(body, 0)
				for i := 0; i < response.NearUsersLength(); i++ {
					fmt.Println("距离过近的用户列表:", response.NearUsers(i))
				}
				for i := 0; i < response.NoPositionUsersLength(); i++ {
					fmt.Println("未开放位置的用户列表:", response.NoPositionUsers(i))
				}
			case Common.CommandGameUserDistanceResponse: // 用户距离列表
				fmt.Println("收到局内用户距离列表")
				response := Common.GetRootAsGameUserDistanceResponse(body, 0)
				code := response.S2cResult(nil).Code()
				if code < 0 {
					fmt.Printf("request user distance error, code:%v, msg:%v\n",
						code, string(response.S2cResult(nil).Msg()))
				} else {
					userDistance := new(Common.GameUserDistance)
					for i := 0; i < response.DistanceListLength(); i++ {
						response.DistanceList(userDistance, i)
						fmt.Printf("minUserId:%v, maxUserId:%v, distance:%v\n",
							userDistance.MinUserId(), userDistance.MaxUserId(), userDistance.Distance())
					}
				}
			case Common.CommandGameResultPush: // 游戏结果
				response := Common.GetRootAsGameResultPush(body, 0)
				fmt.Printf("游戏结果, 房主id:%v, 是否解散:%v,房间号:%v, 房间类型:%v.\n", response.Host(), response.IsDismiss(), string(response.Number()), response.RandomRoom())
				resultInfo := new(Common.ResultInfo_v_2_3_0)
				for i := 0; i < response.ResultInfoV230Length(); i++ {
					response.ResultInfoV230(resultInfo, i)
					fmt.Printf("userId:%v, score:%v, fromTotalScore:%v, totalScore:%v, 星星变化:%v, 自摸:%v, 接炮:%v, 捉鸡:%v, 暗杠:%v, 明杠:%v, 转弯杠:%v, 点炮:%v, 憨包杠:%v.\n",
						resultInfo.UserId(), resultInfo.Score(), resultInfo.FromTotolScore(), resultInfo.TotalScore(), resultInfo.StarChange(), resultInfo.WinSelfTimes(), resultInfo.WinTimes(),
						resultInfo.KitchenTimes(), resultInfo.KongDarkTimes(), resultInfo.KongTimes(), resultInfo.KongTurnTimes(),
						resultInfo.DianPaoTimes(), resultInfo.KongTurnFreeTimes())
				}
			case Common.CommandGameSkipOperateNoticePush: // 跳过操作
				response := Common.GetRootAsGameSkipOperateNoticePush(body, 0)
				fmt.Printf("处于过[%v]状态, tile:%v.\n", response.Op(), response.Tile())
			default:
				break
			}
		}
	}
}

// 心跳
func heartBeat(conn *net.TCPConn, timeInterval int) {
	for {
		time.Sleep(time.Duration(timeInterval) * time.Millisecond)

		message, _ := simplejson.New().Encode()
		imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HEARTBEAT, message)
		conn.Write(imPacket.Serialize())
		// fmt.Println("[", util.GetTimestamp(), "]", "发送心跳")
	}
}
