package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego/httplib"
)

func uamauthclient() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("uamauthclient %s", err.Error())
		}
	}()

	req := httplib.Post(`https://exservice.12306.cn/excater/uamauthclient`)

	req.GetRequest().AddCookie(&http.Cookie{Name: "uamtk", Value: tmpCookie["uamtk"]}) // 就这一个够了

	req.Param(`tk`, newapptk)

	rsp, err := req.Response()
	if err != nil {
		return err
	}
	for _, v := range rsp.Cookies() {
		if v.Name == "tk" {
			fmt.Println("uamauthclient_tk:", v.Value)
			uamauthclient_tk = v.Value
		}
	}
	saveCookie("uamauthclient", rsp)

	if rsp.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	var uamauthclientRsp struct {
		ResultMessage string `json:"result_message"`
		ResultCode    int    `json:"result_code"`
		Apptk         string `json:"apptk"`
		Username      string `json:"username"`
	}

	err = json.Unmarshal(body, &uamauthclientRsp)
	if err != nil {
		return err
	}

	fmt.Printf("uamauthclient: %+v\n", uamauthclientRsp)
	return nil
}
