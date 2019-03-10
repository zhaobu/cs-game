package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"mahjong.go/config"
	"mahjong.go/mi/protocal"

	simplejson "github.com/bitly/go-simplejson"
)

type server struct {
	name string
	host string
	port string
}

// 初始化参数
var (
	serverList []*server
	wg         sync.WaitGroup
)

func main() {
	// 初始化服务器组
	g1 := []string{"g1", "118.178.190.132"}
	g1Ports := []string{"9000", "8999", "8998", "8997", "8996", "8995"}
	g2 := []string{"g2", "121.43.37.217"}
	g2Ports := []string{"9000", "8999", "8998", "8997", "8996", "8995"}
	g3 := []string{"g3", "118.178.127.24"}
	g3Ports := []string{"9000", "8999", "8998", "8997", "8996", "8995"}

	if len(g1Ports) > 0 {
		for _, port := range g1Ports {
			serverList = append(serverList, &server{"g1-" + g1[1], g1[1], port})
		}
	}

	if len(g2Ports) > 0 {
		for _, port := range g2Ports {
			serverList = append(serverList, &server{"g2-" + g2[1], g2[1], port})
		}
	}

	if len(g3Ports) > 0 {
		for _, port := range g3Ports {
			serverList = append(serverList, &server{"g3-" + g3[1], g3[1], port})
		}
	}

	/*
		// g1
		g1_9000 := &server{"g1-9000", "118.178.190.132", "9000"}
		serverList = append(serverList, g1_9000)
		g1_8999 := &server{"g1-8999", "118.178.190.132", "8999"}
		serverList = append(serverList, g1_8999)
		g1_8998 := &server{"g1-8998", "118.178.190.132", "8998"}
		serverList = append(serverList, g1_8998)
		g1_8997 := &server{"g1-8997", "118.178.190.132", "8997"}
		serverList = append(serverList, g1_8997)
		g1_8996 := &server{"g1-8996", "118.178.190.132", "8996"}
		serverList = append(serverList, g1_8996)
		g1_8995 := &server{"g1-8995", "118.178.190.132", "8995"}
		serverList = append(serverList, g1_8995)
		// g2
		g2_9000 := &server{"g2-9000", "121.43.37.217", "9000"}
		serverList = append(serverList, g2_9000)
		g2_8999 := &server{"g2-8999", "121.43.37.217", "8999"}
		serverList = append(serverList, g2_9000)
		g2_8999 := &server{"g2-8999", "121.43.37.217", "8999"}
		serverList = append(serverList, g2_9000)
		g2_8999 := &server{"g2-8999", "121.43.37.217", "8999"}
		serverList = append(serverList, g2_9000)
		g2_8999 := &server{"g2-8999", "121.43.37.217", "8999"}
		serverList = append(serverList, g2_9000)
		g2_8999 := &server{"g2-8999", "121.43.37.217", "8999"}
		serverList = append(serverList, g2_9000)
		// g3
		g3 := &server{"g3-9000", "118.178.127.24", "9000"}
		// serverList = append(serverList, g1_9000, g1_8999, g2, g3)
		serverList = append(serverList, g1_9000, g2, g3)
	*/

	wg.Add(len(serverList))
	for _, s := range serverList {
		go stat(s)
	}
	wg.Wait()
	fmt.Println("统计完成...")
}

func stat(s *server) {
	defer wg.Done()
	time.Sleep(1 * time.Second)

	// 连接服务器
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.host+":"+s.port)
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

	// 发送
	js := simplejson.New()
	js.Set("systemKey", config.SYSTEM_KEY)
	js.Set("act", "stat")
	message, _ := js.Encode()

	imPacket := protocal.NewImPacket(100, message)
	conn.Write(imPacket.Serialize())
	for {
		// 读取包内容
		impacket, err := protocal.ReadPacket(conn)
		// 检查解析错误
		if err != nil {
			// 协议解析错误
			fmt.Println(err.Error())
			break
		}

		js, _ := simplejson.NewJson(impacket.GetMessage())
		str := ""
		str += "[" + s.name + "]"
		str += "[" + s.host + ":" + s.port + "]"
		str += fmt.Sprintf("%v", js.Get("data"))
		fmt.Println(str)

		break
	}

}
