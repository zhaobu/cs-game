package main

import (
	"flag"
	"fmt"
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"
	"net"

	"github.com/fwhappy/util"
)

var (
	// host = flag.String("host", "114.215.254.129", "server host")
	host = flag.String("host", "0.0.0.0", "server host")
	// host = flag.String("host", "114.55.227.47", "server host")
	port = flag.String("port", "9000", "server port")
	conn *net.TCPConn
	id   int
)

func init() {
	flag.Parse()
}

func main() {
	remote := *host + ":" + *port
	tcpAddr, err := net.ResolveTCPAddr("tcp", remote)
	if err != nil {
		fmt.Println("Error:ResolveTCPAddr:", err.Error())
		return
	}
	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Error:DialTCP:", err.Error())
		return
	}
	defer conn.Close()

	showClientDebug("connect at:%v", remote)

	showUsage()

	// 定义接收消息的协程
	go onRecived()

	// 控制台接收输入
	onInput()
}

func showUsage() {
	fmt.Println("------------------------------------------------")
	fmt.Println("usage:")
	fmt.Println("[握手]", protocal.PACKAGE_TYPE_HANDSHAKE, ".token")
	fmt.Println("[拉取大厅列表]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueListRequest)
	fmt.Println("[报名]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueApplyRequest, ".leagueId")
	fmt.Println("[取消报名]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueCancelRequest)
	fmt.Println("[退赛]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueQuitRequest)
	fmt.Println("[收到比赛结果]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueRaceResultReceivedNotify)
	fmt.Println("[加入俱乐部]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandClubJoinRequest, ".clubID")
	fmt.Println("[退出俱乐部]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandClubQuitRequest, ".clubID")
	fmt.Println("[俱乐部消息]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandClubClubMessageNotify, ".clubID.mType.content")
	fmt.Println("[俱乐部历史消息]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandClubClubMessageListNotify, ".clubID.lastMessageId.limit")
	fmt.Println("[俱乐部房间列表]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandClubI2CRoomListRequest, ".clubID")
	fmt.Println("[游戏激活]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandGameActivateRequest)
	fmt.Println("[关闭俱乐部连接]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandGatewayC2SCloseClubNotify)
	fmt.Println("[关闭联赛大厅连接]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandGatewayC2SCloseLeagueNotify)
	fmt.Println("------------------------------------------------")
	fmt.Println("[加入房间]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandJoinRoomRequest, ".number")
}

// 显示客户端错误
func showClientError(a string, b ...interface{}) {
	fmt.Println("[", util.GetTimestamp(), "]", "[ERROR]", fmt.Sprintf(a, b...))
}

// 显示客户端调试信息
// 显示客户端错误
func showClientDebug(a string, b ...interface{}) {
	fmt.Println("[", util.GetTimestamp(), "]", "[DEBUG]", fmt.Sprintf(a, b...))
}
