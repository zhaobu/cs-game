package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/astaxie/beego/httplib"
)

func leftTicketQuery() error {
	fromStationCode := ""
	toStationCode := ""
	var find bool

	fromStationCode, find = cn2Code[*fromStation]
	if !find {
		return fmt.Errorf("没有这个地址 %s\n", *fromStation)
	}

	toStationCode, find = cn2Code[*toStation]
	if !find {
		return fmt.Errorf("没有这个地址 %s\n", *toStation)
	}

	fmt.Printf("fromStationCode:%s toStationCode:%s\n", fromStationCode, toStationCode)

	req := httplib.Get(`https://kyfw.12306.cn/otn/leftTicket/query`)
	req.Param(`leftTicketDTO.train_date`, *trainDate)
	req.Param(`leftTicketDTO.from_station`, fromStationCode)
	req.Param(`leftTicketDTO.to_station`, toStationCode)
	req.Param(`purpose_codes`, `ADULT`)

	str, err := req.String()
	if err != nil {
		return err
	}

	var queryRsp struct {
		Data struct {
			Flag   string   `json:"flag"`
			Result []string `json:"result"`
		} `json:"data"`
		Httpstatus int    `json:"httpstatus"`
		Messages   string `json:"messages"`
		Status     bool   `json:"status"`
	}

	if err = json.Unmarshal([]byte(str), &queryRsp); err != nil {
		return err
	}

	if queryRsp.Httpstatus == 200 && queryRsp.Status {
		fmt.Println("查询成功!")
	}

	var tickInfoSli []ticketInfo

	for _, v := range queryRsp.Data.Result {
		spl := strings.Split(v, "|")
		if len(spl) != 37 {
			continue
		}
		var info ticketInfo

		info.secretStr = spl[0]
		info.remark = spl[1]
		info.train_no = spl[2]
		info.station_train_code = spl[3]
		info.start_station_telecode = spl[4]

		info.end_station_telecode = spl[5]
		info.from_station_telecode = spl[6]
		info.to_station_telecode = spl[7]
		info.start_time = spl[8]
		info.arrive_time = spl[9]

		info.lishi = spl[10]
		info.canWebBuy = spl[11]
		info.yp_info = spl[12]
		info.start_train_date = spl[13]
		info.train_seat_feature = spl[14]

		info.location_code = spl[15]
		info.from_station_no = spl[16]
		info.to_station_no = spl[17]
		info.is_support_card = spl[18]
		info.controlled_train_flag = spl[19]

		info.gg_num = spl[20]
		info.gr_num = spl[21]
		info.qt_num = spl[22]
		info.rw_num = spl[23]
		info.rz_num = spl[24]

		info.tz_num = spl[25]
		info.wz_num = spl[26]
		info.yb_num = spl[27]
		info.yw_num = spl[28]
		info.yz_num = spl[29]

		info.ze_num = spl[30]
		info.zy_num = spl[31]
		info.swz_num = spl[32]
		info.srrb_num = spl[33]
		info.yp_ex = spl[34]

		info.seat_types = spl[35]
		info.exchange_train_flag = spl[36]

		if info.canWebBuy != "Y" {
			continue
		}

		if info.ze_num == "无" && info.zy_num == "无" && info.swz_num == "无" {
			continue
		}

		tickInfoSli = append(tickInfoSli, info)

	}

	sort.Slice(tickInfoSli, func(i, j int) bool {
		return tickInfoSli[i].start_time < tickInfoSli[j].start_time
	})

	if len(tickInfoSli) > 0 {
		fmt.Printf("查询到票 %+v\n", tickInfoSli[0])
		secretStr = tickInfoSli[0].secretStr
	}

	return nil
}
