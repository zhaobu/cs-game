package main

import (
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego/httplib"
)

func submitOrderRequest() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("submitOrderRequest %s", err.Error())
		}
	}()

	req := httplib.Post(`https://kyfw.12306.cn/otn/leftTicket/submitOrderRequest`)

	req.Param(`secretStr`, secretStr)
	req.Param(`train_date`, *trainDate)
	req.Param(`back_train_date`, *trainDate)
	req.Param(`tour_flag`, `dc`)
	req.Param(`purpose_codes`, `ADULT`)
	req.Param(`query_from_station_name`, *fromStation)
	req.Param(`query_to_station_name`, *toStation)
	req.Param(`undefined`, ``)

	rsp, err := req.Response()
	if err != nil {
		return err
	}

	saveCookie("submitOrderRequest", rsp)

	if rsp.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	fmt.Println("submitOrderRequest", string(body))
	return nil
}
