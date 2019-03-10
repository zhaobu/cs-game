package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"
	fbs "mahjong.club/fbs/Common"
)

var (
	//host = flag.String("host", "114.215.254.129", "server host")
	host = flag.String("host", "0.0.0.0", "server host")
	// host = flag.String("host", "114.55.227.47", "server host")
	port = flag.String("port", "38438", "server port")
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
	fmt.Println("[下线]", protocal.PACKAGE_TYPE_KICK)
	fmt.Println("[加入俱乐部]", protocal.PACKAGE_TYPE_DATA, ".", fbs.CommandClubJoinRequest, ".clubID")
	fmt.Println("[退出俱乐部]", protocal.PACKAGE_TYPE_DATA, ".", fbs.CommandClubQuitRequest, ".clubID")
	fmt.Println("[俱乐部消息]", protocal.PACKAGE_TYPE_DATA, ".", fbs.CommandClubClubMessageNotify, ".clubID.mType.content")
	fmt.Println("[俱乐部历史消息]", protocal.PACKAGE_TYPE_DATA, ".", fbs.CommandClubClubMessageListNotify, ".clubID.lastMessageId.limit")
	fmt.Println("[俱乐部房间列表]", protocal.PACKAGE_TYPE_DATA, ".", fbs.CommandClubI2CRoomListRequest, ".clubID")
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
