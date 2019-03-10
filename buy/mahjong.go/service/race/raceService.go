package race

import (
	"io/ioutil"
	"net/http"
	"strconv"

	simplejson "github.com/bitly/go-simplejson"
	"mahjong.go/library/core"
)

// GetUserRace 获取用户当前比赛id
func GetUserRace(userId int) (raceId int64) {
	url := core.AppConfig.LeagueGetUserRaceURL + "?userId=" + strconv.Itoa(userId)
	core.Logger.Debug("[getUserRace]url:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[getUserRace]het.Get, url:%v, error:%v", url, err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Error("[getUserRace]het.Get read body, url:%v, error:%v", url, err.Error())
	}
	core.Logger.Debug("[getUserRace]het.Get read body,body:%v", string(body))

	js, _ := simplejson.NewJson(body)
	code, _ := js.Get("code").Int()
	if code < 0 {
		message, _ := js.Get("message").String()
		core.Logger.Error("[getUserRace]userId:%v, code:%v, msg:%v", userId, code, message)
	} else {
		raceId, _ = js.Get("raceId").Int64()
		core.Logger.Info("[getUserRace]userId:%v, raceId:%v", userId, raceId)
	}
	return
}
