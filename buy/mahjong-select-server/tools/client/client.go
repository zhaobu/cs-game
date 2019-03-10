package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mahjong-select-server/config"
	"net/http"

	"github.com/fwhappy/util"

	fbsInfo "mahjong-select-server/fbs/info"

	flatbuffers "github.com/google/flatbuffers/go"
)

// 参数定义
//var url = "http://localhost:8081/client/selectServer"
//var id = 210499
var url = "http://114.55.227.47:8081/client/selectServer"
var id = 1
var number = "CREATE_ROOM"

func main() {
	// 构建fbs对象
	builder := flatbuffers.NewBuilder(0)
	roomNum := builder.CreateString(number)
	fbsInfo.SelectServerRequestStart(builder)
	fbsInfo.SelectServerRequestAddUserId(builder, uint32(id))
	fbsInfo.SelectServerRequestAddRoomNum(builder, roomNum)
	orc := fbsInfo.SelectServerRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	// 生成token

	responseBuf, _ := doBytesPost(url+"?token="+getToken(), buf)
	response := fbsInfo.GetRootAsSelectServerResponse(responseBuf, 0)
	result := new(fbsInfo.CommonResult)
	result = response.Result(result)
	fmt.Println("code=", result.Code(), ",msg=", string(result.Msg()))
	if result.Code() >= 0 {
		fmt.Println("roomId=", response.RoomId())
		fmt.Println("ip:port=", string(response.Ip()), ":", response.Port())
	}
}

func getToken() string {
	// 1 return "900bZm8rRDhtWcvt0MnKtw2JiYbpLQa_YV5RoPoPeJVLjbPM56gV48x3NVTUvtbXHtH1"
	return "ac68lHjS6EKhUUu_kMs6jbT4XQi9ibvcyqoxKQo46n3mGt2VY2xwMOgF6tcP_FkBgLaIaG6k77-qpVTxpQtr6_42prZse5WewT10O8HA"
	return util.GenToken(id, "", config.TOKEN_SECRET_KEY)
}

func doBytesPost(url string, data []byte) ([]byte, error) {
	body := bytes.NewReader(data)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println("http.NewRequest,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	request.Header.Set("Connection", "Keep-Alive")
	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		log.Println("http.Do failed,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("http.Do failed,[err=%s][url=%s]", err, url)
	}
	return b, err
}
