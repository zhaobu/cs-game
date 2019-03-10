package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ticketInfo struct {
	secretStr              string // [0]
	remark                 string // 备注
	train_no               string // 车票号
	station_train_code     string // 车次
	start_station_telecode string // 起始站代号

	end_station_telecode  string // 终点站代号 [5]
	from_station_telecode string // 出发站代号
	to_station_telecode   string // 到达站代号
	start_time            string // 出发时间
	arrive_time           string // 到达时间

	lishi              string // 历时  [10]
	canWebBuy          string // 是否能购买：Y 可以
	yp_info            string //
	start_train_date   string // 出发日期
	train_seat_feature string

	location_code         string // [15]
	from_station_no       string
	to_station_no         string
	is_support_card       string
	controlled_train_flag string

	gg_num string // [20]
	gr_num string // 高级软卧
	qt_num string // 其他
	rw_num string // 软卧
	rz_num string // 软座

	tz_num string // [25]
	wz_num string // 无座
	yb_num string //
	yw_num string // 硬卧
	yz_num string // 硬座

	ze_num   string // [30] 二等座
	zy_num   string // 一等座
	swz_num  string // 商务座
	srrb_num string // 动卧
	yp_ex    string

	seat_types          string // [35]
	exchange_train_flag string
}

// 图片点击坐标
type point struct {
	x, y int
}

var (
	points = []point{
		point{50, 50}, point{120, 50}, point{190, 50}, point{260, 50},
		point{50, 120}, point{120, 120}, point{190, 120}, point{260, 120},
	}
)

// "1,2,...8" 转 50,50,260,120
func stridx2points(str string) string {
	var arr []string
	for _, v := range strings.Split(str, ",") {
		i, err := strconv.Atoi(v)
		if err == nil {
			if i < 1 || i > 8 {
				continue
			}
			p := points[i-1]
			arr = append(arr, strconv.Itoa(p.x))
			arr = append(arr, strconv.Itoa(p.y))
		}
	}
	return strings.Join(arr, ",")
}

// 1,2,...8 转 50,50,260,120
func idx2points(idxs ...int) string {
	var arr []string
	for _, v := range idxs {
		if v < 1 || v > 8 {
			continue
		}
		p := points[v-1]
		arr = append(arr, strconv.Itoa(p.x))
		arr = append(arr, strconv.Itoa(p.y))
	}
	return strings.Join(arr, ",")
}

const (
	// TODO 这个值是否要变化？
	jQuery = `jQuery191084319782487173_1542797282762` // [jQueryXX_ms]
)

func trimjQuery(str string) string {
	const (
		jQueryHead = `/**/` + jQuery + `(`
		jQueryEnd  = `);`
	)
	str = strings.TrimPrefix(str, jQueryHead)
	str = strings.TrimSuffix(str, jQueryEnd)
	return str
}

func saveCookie(fn string, rsp *http.Response) {
	fmt.Printf("--> response cookie %s %v\n", fn, rsp.Cookies())
	for _, n := range rsp.Cookies() {
		tmpCookie[n.Name] = n.Value
	}
	return
}
