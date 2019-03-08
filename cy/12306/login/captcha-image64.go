package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/astaxie/beego/httplib"
)

func captchaImage64() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("captchaImage64 %s", err.Error())
		}
	}()

	ms := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

	u := `https://kyfw.12306.cn/passport/captcha/captcha-image64`
	req := httplib.Get(u)
	req.Param(`login_site`, `E`)
	req.Param(`module`, `login`)
	req.Param(`rand`, `sjrand`)
	req.Param(ms, ``)
	req.Param(`_`, ms) // _好像是自增
	req.Param(`callback`, jQuery)

	rsp, err := req.Response()
	if err != nil {
		return err
	}

	saveCookie("captchaImage64", rsp)

	if rsp.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	strBody := trimjQuery(string(body))

	var captchaImage64Rsp struct {
		ResultMessage string `json:"result_message"`
		ResultCode    string `json:"result_code"`
		Image         string `json:"image"`
	}

	err = json.Unmarshal([]byte(strBody), &captchaImage64Rsp)
	if err != nil {
		return nil
	}

	if captchaImage64Rsp.ResultCode != "0" {
		err = fmt.Errorf("生成验证码失败 %s", captchaImage64Rsp.ResultMessage)
		return err
	}

	// 保存图片
	imageData, err := base64.StdEncoding.DecodeString(captchaImage64Rsp.Image)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("captcha.jpg", imageData, 0666)
	if err != nil {
		return err
	}

	return err
}
