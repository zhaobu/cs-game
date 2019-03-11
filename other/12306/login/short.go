package main

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/httplib"
)

var (
	cn2Code = make(map[string]string)
	code2Cn = make(map[string]string)
)

func initShort() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			panic(r)
		}
	}()

	req := httplib.Get(`https://kyfw.12306.cn/otn/resources/js/framework/station_name.js?station_version=1.9043`)
	str, err := req.String()
	if err != nil {
		fmt.Println(err)
		return
	}

	str = strings.TrimPrefix(str, "var station_names ='")
	str = strings.TrimSuffix(str, "';")

	strsli := strings.Split(str, "|")
	s := len(strsli)
	fmt.Println("size ", s)

	for i := 0; i < (s - 1); i += 5 {
		//fmt.Println(strsli[i+1], " -> ", strsli[i+2])
		cn2Code[strsli[i+1]] = strsli[i+2]
		code2Cn[strsli[i+2]] = strsli[i+1]
	}
}
