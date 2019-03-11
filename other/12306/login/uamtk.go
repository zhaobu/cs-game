package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego/httplib"
)

func authUamtk() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("authUamtk %s", err.Error())
		}
	}()

	req := httplib.Post(`https://kyfw.12306.cn/passport/web/auth/uamtk`)

	req.GetRequest().AddCookie(&http.Cookie{Name: "uamtk", Value: tmpCookie["uamtk"]}) // 就这一个够了

	req.Param(`uamtk`, ``)
	req.Param(`appid`, `excater`)

	rsp, err := req.Response()
	if err != nil {
		return err
	}

	saveCookie("authUamtk", rsp)

	if rsp.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	var uamtkRsp struct {
		ResultMessage string `json:"result_message"`
		ResultCode    int    `json:"result_code"`
		Newapptk      string `json:"newapptk"`
	}

	err = json.Unmarshal([]byte(body), &uamtkRsp)
	if err != nil {
		return err
	}

	if uamtkRsp.ResultCode != 0 {
		return fmt.Errorf("%s", uamtkRsp.ResultMessage)
	}

	newapptk = uamtkRsp.Newapptk

	fmt.Printf("uamtk: %+v\n", uamtkRsp)
	return nil
}
