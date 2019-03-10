package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/astaxie/beego/httplib"
)

func captchaCheck() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("captchaCheck %s", err.Error())
		}
	}()

	var stridxs string
	fmt.Println("输入图片序号如 1,2,3,4,5,6,7,8 ")
	fmt.Scanf("%s", &stridxs)

	answer := stridx2points(stridxs)
	fmt.Printf("你选择了 %s\n", answer)

	ms := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

	req := httplib.Get(`https://kyfw.12306.cn/passport/captcha/captcha-check`)

	req.GetRequest().AddCookie(&http.Cookie{Name: "_passport_ct", Value: tmpCookie["_passport_ct"]}) // 就这一个够了

	req.Param(`callback`, jQuery)
	req.Param(`answer`, answer)
	req.Param(`rand`, `sjrand`)
	req.Param(`login_site`, `E`)
	req.Param(`_`, ms) // _好像是自增

	rsp, err := req.Response()
	if err != nil {
		return err
	}

	saveCookie("captchaCheck", rsp)

	if rsp.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	strBody := trimjQuery(string(body))

	var captchaCheckRsp struct {
		ResultMessage string `json:"result_message"`
		ResultCode    string `json:"result_code"`
	}

	err = json.Unmarshal([]byte(strBody), &captchaCheckRsp)
	if err != nil {
		return err
	}

	if captchaCheckRsp.ResultCode != "4" {
		return fmt.Errorf("%s", captchaCheckRsp.ResultMessage)
	}

	fmt.Printf("检查验证码: %+v\n", captchaCheckRsp)
	return nil
}
