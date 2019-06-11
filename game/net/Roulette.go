package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var Bureau int

type GetConditionResultEntity struct {
	Bureau int
	Recharge int
}

type GetConditionResult struct {
	Entity GetletteNumResultEntity
	Code int
	Msg string
	IsSuccess bool
}

type GetletteNumResultEntity struct {
	Bureau int
}

type GetletteNumResult struct {
	Entity GetletteNumResultEntity
	Code int
	Msg string
	IsSuccess bool
}

func GetCondition()(err error) {
	song := make(map[string]interface{})
	song["Time"] =  time.Now().Format("20060102150405")
	song["Sign"] =  md5V(fmt.Sprintf("%s%s",song["Time"].(string),NetKey))
	bytesData, err := json.Marshal(song)
	if err != nil {
		return fmt.Errorf("推送用户注册信息到 .net  GetCondition 序列化参数错误 err:%s",err)
	}
	reader := bytes.NewReader(bytesData)
	url := netAddr +"/Roulette/GetCondition"
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return 	fmt.Errorf("推送用户注册信息到 .net  GetCondition 错误1 err:%s",err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("推送用户注册信息到 .net  GetCondition 错误2 err:%s",err)
	}
	Result := &GetConditionResult{}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return fmt.Errorf("推送用户注册信息到 .net  GetCondition 错误3 err:%s",err)
	}
	err = json.Unmarshal(respBytes,Result)
	if err != nil{
		return fmt.Errorf("推送用户注册信息到 .net  GetCondition 错误4 err:%s",err)
	}
	if Result.Code != 0 {
		return fmt.Errorf("推送用户注册信息到 .net  GetCondition 错误5 err:%s",string(respBytes))
	}
	Bureau = Result.Entity.Bureau
	return
}

func GetletteNum(uId uint64,playnum int)(err error){
	song := make(map[string]interface{})
	song["UserId"] =  uId
	song["Type"] =  1
	song["Number"] =  playnum
	song["Time"] =  time.Now().Format("20060102150405")
	song["Sign"] =  md5V(fmt.Sprintf("%d%d%s%s",uId,1,song["Time"].(string),NetKey))
	bytesData, err := json.Marshal(song)
	if err != nil {
		return fmt.Errorf("推送用户注册信息到 .net  GetletteNum 序列化参数错误 err:%s",err)
	}
	reader := bytes.NewReader(bytesData)
	url := netAddr +"/Roulette/GetletteNum"
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return 	fmt.Errorf("推送用户注册信息到 .net  GetletteNum 错误1 err:%s",err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("推送用户注册信息到 .net  GetletteNum 错误2 err:%s",err)
	}
	Result := &GetletteNumResult{}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return fmt.Errorf("推送用户注册信息到 .net  GetletteNum 错误3 err:%s",err)
	}
	err = json.Unmarshal(respBytes,Result)
	if err != nil{
		return fmt.Errorf("推送用户注册信息到 .net  GetletteNum 错误4 err:%s",err)
	}
	if Result.Code == -5 {
		Bureau = Result.Entity.Bureau
	}
	if Result.Code != 0 ||  Result.Code != -5 {
		return fmt.Errorf("推送用户注册信息到 .net  GetletteNum 错误5 err:%s",string(respBytes))
	}
	return
}