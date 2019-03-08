package main

import (
	"flag"
	"fmt"
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/protocal"
	"net"

	"github.com/fwhappy/util"
)

var (
	// host = flag.String("host", "114.215.254.129", "server host")
	host = flag.String("host", "0.0.0.0", "server host")
	// host = flag.String("host", "114.55.227.47", "server host")
	port = flag.String("port", "31212", "server port")
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
	fmt.Println("[握手]", protocal.PACKAGE_TYPE_HANDSHAKE, ".userID")
	fmt.Println("[拉取大厅列表]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueListRequest)
	fmt.Println("[报名]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueApplyRequest, ".leagueId")
	fmt.Println("[取消报名]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueCancelRequest)
	fmt.Println("[退赛]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueQuitRequest)
	fmt.Println("[收到比赛结果]", protocal.PACKAGE_TYPE_DATA, ".", fbsCommon.CommandLeagueRaceResultReceivedNotify)
	fmt.Println("------------------------------------------------")
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
