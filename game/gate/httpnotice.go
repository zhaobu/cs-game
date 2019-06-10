package main

import (
	"bytes"
	"crypto/md5"
	pbcommon "cy/game/pb/common"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	NetKey string = "eo05Efekb*1sMuM6"
)

func md5V(str string) string  {
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

//用户注册需向 .net 那边推送消息
func PushUserRegister(info *pbcommon.UserInfo) {
	song := make(map[string]interface{})
	song["UserId"] = info.UserID
	song["UnionId"] = info.WxID
	song["HeadUrl"] = info.Profile
	song["WxName"] = info.Name
	song["Time"] =  time.Now().Format("20060102150405")
	song["Sign"] =  md5V(fmt.Sprintf("%d%s%s%s",info.UserID,info.WxID,song["Time"].(string),NetKey))
	bytesData, err := json.Marshal(song)
	if err != nil {
		log.Errorf("推送用户注册信息到 .net  PushUserRegister 序列化参数错误 err:%s",err)
		return
	}
	reader := bytes.NewReader(bytesData)
	url := *netAddr + "/Register/RegisterUser"
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Errorf("推送用户注册信息到 .net  PushUserRegister 错误1 err:%s",err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Errorf("推送用户注册信息到 .net  PushUserRegister 错误2 err:%s",err)
		return
	}
	resCode := resp.StatusCode
	resp.Body.Close()
	if resCode != 200 {
		log.Errorf("推送用户注册信息到 .net  PushUserRegister 错误3 resCode:%d",resCode)
		return
	}
}
