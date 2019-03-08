package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego/httplib"
)

func weblogin() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("weblogin %s", err.Error())
		}
	}()

	req := httplib.Post(`https://kyfw.12306.cn/passport/web/login`)

	req.GetRequest().AddCookie(&http.Cookie{Name: "_passport_ct", Value: tmpCookie["_passport_ct"]}) // 就这一个够了

	req.Param(`username`, *username)
	req.Param(`password`, *password)
	req.Param(`appid`, `excater`)

	rsp, err := req.Response()
	if err != nil {
		return err
	}

	saveCookie("weblogin", rsp)

	if rsp.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	var webloginRsp struct {
		ResultMessage string `json:"result_message"`
		ResultCode    int    `json:"result_code"`
		Uamtk         string `json:"uamtk"`
	}

	err = json.Unmarshal(body, &webloginRsp)
	if err != nil {
		return err
	}

	if webloginRsp.ResultCode != 0 {
		return fmt.Errorf("%s", webloginRsp.ResultMessage)
	}

	fmt.Printf("登陆结果: %+v\n", webloginRsp)
	return nil
}
