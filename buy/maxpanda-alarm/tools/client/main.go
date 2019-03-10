package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fwhappy/util"
)

func main() {
	testAlarm()
	testMailAlarm()
	testMobileAlarm()
}

func testAlarm() {
	url := "http://127.0.0.1:13579/alarm/mail"
	smsParams := make(map[string]interface{})
	smsParams["code"] = 1234
	sp, _ := util.InterfaceToJsonString(smsParams)
	data := make(map[string]string)
	data["target"] = "default"
	data["subject"] = "测试报警subject"
	data["body"] = "测试报警body"
	data["sms_code"] = "SMS_76115014"
	data["sms_params"] = sp
	str, _ := util.InterfaceToJsonString(data)
	fmt.Println("str:", str)
	response := httpPost(url, str)
	fmt.Println("[testMailAlarm]response:", response)
}

func testMailAlarm() {
	url := "http://127.0.0.1:13579/alarm/mail"
	data := make(map[string]string)
	data["target"] = "default"
	data["subject"] = "测试报警subject"
	data["body"] = "测试报警body"
	str, _ := util.InterfaceToJsonString(data)
	fmt.Println("str:", str)
	response := httpPost(url, str)
	fmt.Println("[testAlarm]response:", response)
}

func testMobileAlarm() {
	url := "http://127.0.0.1:13579/alarm/mobile"
	data := make(map[string]interface{})
	smsParams := make(map[string]interface{})
	smsParams["code"] = 1234
	sp, _ := util.InterfaceToJsonString(smsParams)
	data["target"] = "default"
	data["sms_code"] = "SMS_76115014"
	data["sms_params"] = sp
	str, _ := util.InterfaceToJsonString(data)
	fmt.Println("str:", str)
	response := httpPost(url, str)
	fmt.Println("[testMobileAlarm]response:", response)
}

func httpPost(url string, data string) string {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader("*="+data))
	if err != nil {
		fmt.Println("[httpPost.Post]error:", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[httpPost.ReadAll]error:", err)
		return ""
	}
	return string(body)
}
