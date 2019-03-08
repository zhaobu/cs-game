package main

import (
	"flag"
	"fmt"
)

var (
	trainDate   = flag.String("date", "2018-11-30", "train date")
	fromStation = flag.String("from", "深圳北", "from station")
	toStation   = flag.String("to", "长沙南", "to station")

	username = flag.String("u", "3015014@qq.com", "12306登陆名称")
	password = flag.String("p", "ly000111", "12306登陆密码")

	secretStr               string
	uamauthclient_tk        string
	newapptk                string
	globalRepeatSubmitToken string

	// 测试用
	tmpCookie = make(map[string]string)
)

func main() {
	flag.Parse()

	var err error

	initShort()
	fmt.Println(leftTicketQuery())

	err = captchaImage64()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = captchaCheck()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = weblogin()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = authUamtk()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = uamauthclient()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = submitOrderRequest()
	if err != nil {
		fmt.Println(err)
		return
	}

	// err = initDc()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
}
