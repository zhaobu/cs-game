package net

import (
	"bytes"
	pbcommon "cy/game/pb/common"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func PushUserRegister(info *pbcommon.UserInfo)(err error) {
	song := make(map[string]interface{})
	song["UserId"] = info.UserID
	song["UnionId"] = info.WxID
	song["HeadUrl"] = info.Profile
	song["WxName"] = info.Name
	song["Time"] =  time.Now().Format("20060102150405")
	song["Sign"] =  md5V(fmt.Sprintf("%d%s%s%s",info.UserID,info.WxID,song["Time"].(string),NetKey))
	bytesData, err := json.Marshal(song)
	if err != nil {
		return fmt.Errorf("推送用户注册信息到 .net  PushUserRegister 序列化参数错误 err:%s",err)
	}
	reader := bytes.NewReader(bytesData)
	url := netAddr + "/Register/RegisterUser"
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return 	fmt.Errorf("推送用户注册信息到 .net  PushUserRegister 错误1 err:%s",err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("推送用户注册信息到 .net  PushUserRegister 错误2 err:%s",err)
	}
	resCode := resp.StatusCode
	if resCode != 200 {
		return 	fmt.Errorf("推送用户注册信息到 .net  PushUserRegister 错误3 resCode:%d",resCode)
	}
	return
}