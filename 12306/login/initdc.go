package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/astaxie/beego/httplib"
)

func initDc() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("initDc %s", err.Error())
		}
	}()

	req := httplib.Get(`https://kyfw.12306.cn/otn/confirmPassenger/initDc`)

	for k, v := range tmpCookie {
		req.GetRequest().AddCookie(&http.Cookie{Name: k, Value: v})
	}
	// req.GetRequest().AddCookie(&http.Cookie{Name: "uamtk", Value: tmpCookie["uamtk"]})
	// req.GetRequest().AddCookie(&http.Cookie{Name: "tk", Value: tmpCookie["tk"]})                                           // 就这一个够了
	// req.GetRequest().AddCookie(&http.Cookie{Name: "route", Value: tmpCookie["route"]})                                     // 就这一个够了
	// req.GetRequest().AddCookie(&http.Cookie{Name: "JSESSIONID", Value: tmpCookie["JSESSIONID"]})                           // 就这一个够了
	// req.GetRequest().AddCookie(&http.Cookie{Name: "BIGipServerpool_excater", Value: tmpCookie["BIGipServerpool_excater"]}) // 就这一个够了

	fmt.Println("cookies iniDc: ", req.GetRequest().Cookies())

	req.Param(`_json_att`, ``)

	srcCode, err := req.String()
	if err != nil {
		return err
	}

	fmt.Println(srcCode)
	idx := strings.Index(srcCode, `globalRepeatSubmitToken = '`)
	if idx == -1 {
		return fmt.Errorf("can not find globalRepeatSubmitToken")
	}
	idx += len(`globalRepeatSubmitToken = '`)

	idx2 := strings.Index(srcCode[idx:], `'`)
	if idx2 == -1 {
		return fmt.Errorf("can not find '")
	}

	globalRepeatSubmitToken = srcCode[idx : idx+idx2]
	fmt.Println("token: ", globalRepeatSubmitToken)
	return nil
}
