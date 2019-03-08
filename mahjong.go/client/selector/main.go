package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/fwhappy/util"
	"mahjong.go/library/core"
	"mahjong.go/mi/card"
	"mahjong.go/mi/step"
	"mahjong.go/mi/suggest"
	"mahjong.go/mi/ting"
)

var cards = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9,
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
}
var steps = [][]int{
	[]int{1, 1},
	[]int{2, 2},
	[]int{3, 3},
	[]int{4, 4},
	[]int{5, 5},
	[]int{6, 6},
	[]int{7, 7},
	[]int{8, 8},
	[]int{9, 9},
	[]int{11, 11},
	[]int{12, 12},
	[]int{13, 13},
	[]int{14, 14},
	[]int{15, 15},
	[]int{16, 16},
	[]int{17, 17},
	[]int{18, 18},
	[]int{19, 19},
	[]int{21, 21},
	[]int{22, 22},
	[]int{23, 23},
	[]int{24, 24},
	[]int{25, 25},
	[]int{26, 26},
	[]int{27, 27},
	[]int{28, 28},
	[]int{29, 29},
	[]int{1, 1, 1},
	[]int{2, 2, 2},
	[]int{3, 3, 3},
	[]int{4, 4, 4},
	[]int{5, 5, 5},
	[]int{6, 6, 6},
	[]int{7, 7, 7},
	[]int{8, 8, 8},
	[]int{9, 9, 9},
	[]int{11, 11, 11},
	[]int{12, 12, 12},
	[]int{13, 13, 13},
	[]int{14, 14, 14},
	[]int{15, 15, 15},
	[]int{16, 16, 16},
	[]int{17, 17, 17},
	[]int{18, 18, 18},
	[]int{19, 19, 19},
	[]int{21, 21, 21},
	[]int{22, 22, 22},
	[]int{23, 23, 23},
	[]int{24, 24, 24},
	[]int{25, 25, 25},
	[]int{26, 26, 26},
	[]int{27, 27, 27},
	[]int{28, 28, 28},
	[]int{29, 29, 29},
	[]int{1, 2, 3},
	[]int{2, 3, 4},
	[]int{3, 4, 5},
	[]int{4, 5, 6},
	[]int{5, 6, 7},
	[]int{6, 7, 8},
	[]int{7, 8, 9},
	[]int{11, 12, 13},
	[]int{12, 13, 14},
	[]int{13, 14, 15},
	[]int{14, 15, 16},
	[]int{15, 16, 17},
	[]int{16, 17, 18},
	[]int{17, 18, 19},
	[]int{21, 22, 23},
	[]int{22, 23, 24},
	[]int{23, 24, 25},
	[]int{24, 25, 26},
	[]int{25, 26, 27},
	[]int{26, 27, 28},
	[]int{27, 28, 29},
}

type suggestInfo struct {
	PlayTile int         `json:"tile"`    // 出牌
	Suggest  bool        `json:"suggest"` // 是否推荐
	Best     bool        `json:"best"`    // 是否最优
	Gu       bool        `json:"gu"`      // 是否孤张
	Diao     bool        `json:"diao"`    // 是否吊张
	Bian     bool        `json:"bian"`    // 是否边张
	FETiles  map[int]int `json:"feTiles"` // 一类有效牌列表, tile => 张数
	FECount  int         `json:"feCount"` // 一类有效牌总张数
	SETiles  map[int]int `json:"seTiles"` // 二类有效牌列表, tile => 张数
	SECount  int         `json:"seCount"` // 二类有效牌总张数 tile => 张数
	Score    int         `json:"-"`       // 有效分

}

func init() {
	// 解析url参数
	flag.Parse()
}

var (
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "local", "env")
	// 监听端口
	port = flag.Int("port", 8085, "port")
)

func main() {
	showDebug("选牌服务开启，listen remote : %v, env:%s", *port, *env)
	defer showDebug("选牌服务已关闭，listen remote : %v, env:%s", *port, *env)

	// 初始化基础配置
	core.LoadAppConfig(core.GetConfigFile("app.toml", *env, "conf"))

	// 测试
	http.HandleFunc("/suggest", getSuggest)

	err := http.ListenAndServe(fmt.Sprintf(":%v", *port), nil)
	if err != nil {
		showError("ListenAndServe error, err:%v", err)
		return
	}
}

func getSuggest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	r.ParseForm()
	showDebug("收到客户端请求: %#v", r.Form)

	from := r.Form.Get("from")
	if from != "js" {
		token := r.Form.Get("token")
		// 验证token
		err := verifyToken(token)
		if err != nil {
			data := make(map[string]interface{})
			data["errcode"] = -1
			data["errmsg"] = err.Error()
			b, _ := json.Marshal(data)
			showDebug("data:%v", string(b))
			w.Write(b)
			return
		}
	}

	tilestr := r.Form.Get("tiles")
	discardTilestr := r.Form.Get("discard_tiles")
	tileStep, _ := strconv.Atoi(r.Form.Get("step"))
	tiles := []int{}
	for _, v := range strings.Split(tilestr, ",") {
		tile, _ := strconv.Atoi(v)
		if util.IntInSlice(tile, card.MahjongCards108) {
			tiles = append(tiles, tile)
		}
	}
	discardTiles := []int{}
	for _, v := range strings.Split(discardTilestr, ",") {
		tile, _ := strconv.Atoi(v)
		if util.IntInSlice(tile, card.MahjongCards108) {
			discardTiles = append(discardTiles, tile)
		}
	}

	// 如果未输入牌，则自动根据牌阶自动生成牌
	if len(tiles) == 0 {
		tiles = autoTiles(tileStep)
	}
	sort.Ints(tiles)
	tileStep = step.GetCardsStep(tiles)
	showDebug("tiles=:%+v, step:%v", tiles, tileStep)

	// 推荐列表
	suggestResult := make([]*suggestInfo, 0)

	// 独立的牌
	uniqueTiles := util.SliceUniqueInt(tiles)
	sort.Ints(uniqueTiles)
	showDebug("uniqueTiles:%v", uniqueTiles)

	// 有关系的牌
	relationTiles := card.GetRelationTiles(uniqueTiles...)
	showDebug("relationTiles:%v", relationTiles)

	// 获取孤张
	guTiles := card.GetGuTiles(tiles...)
	sort.Ints(guTiles)
	showDebug("guTiles:%v", guTiles)

	// 或取吊张
	diaoTiles := card.GetDiaoTiles(tiles...)
	sort.Ints(diaoTiles)
	showDebug("diaoTiles:%v", diaoTiles)

	tingMap := ting.GetTingMap(tiles, nil)
	showDebug("tingMap:%#v", tingMap)

	// if len(tingMap) > 0 {
	// 	// 组织叫牌数据
	// 	for _, tile := range uniqueTiles {
	// 		if v, ok := tingMap[tile]; ok {
	// 			sort.Ints(v)
	// 			info := &suggestInfo{
	// 				PlayTile: tile,
	// 				Suggest:  true,
	// 				Gu:       util.IntInSlice(tile, guTiles),
	// 				Diao:     util.IntInSlice(tile, diaoTiles),
	// 				Bian:     util.IntInSlice(tile, card.SideCards),
	// 				RTiles:   v,
	// 			}
	// 			suggestResult = append(suggestResult, info)
	// 		}
	// 	}
	// } else {
	// 开始选牌
	startTime := util.GetMillsTime()
	ms := suggest.NewMSelector()
	ms.SetTiles(card.MahjongCards108)
	ms.SetHandTilesSlice(tiles)
	ms.SetDiscardTilesSlice(discardTiles)
	ms.CalcRemaimTiles()

	// 计算推荐出牌
	_, firstEffectsMap, secondEffectsMap := ms.GetSuggestProgress()
	suggestTile, _ := ms.GetSuggest()
	// showDebug("suggestTile:%#v", suggestTile)
	// showDebug("firstEffectsMap:%#v", firstEffectsMap)
	// showDebug("secondEffectsMap:%#v", secondEffectsMap)

	// 有效分= feCount * 1000 + seCount
	maxScore := 0
	// 组织叫牌数据
	for _, playTile := range uniqueTiles {
		firstEffects, ok := firstEffectsMap[playTile]
		if !ok {
			continue
		}

		// 一类有效牌
		feCount := 0
		feTiles := make(map[int]int)
		for _, tile := range firstEffects {
			feTiles[tile] = ms.GetRemainTilesCnt([]int{tile})
			feCount += feTiles[tile]
		}

		// 二类有效牌
		seCount := 0
		seTiles := make(map[int]int)
		if secondEffects, ok := secondEffectsMap[playTile]; ok {
			for _, tile := range secondEffects {
				seTiles[tile] = ms.GetRemainTilesCnt([]int{tile})
				seCount += seTiles[tile]
			}
		}

		isSuggest := (playTile == suggestTile)
		info := &suggestInfo{
			PlayTile: playTile,
			Suggest:  false,
			Best:     isSuggest,
			Gu:       util.IntInSlice(playTile, guTiles),
			Diao:     util.IntInSlice(playTile, diaoTiles),
			Bian:     util.IntInSlice(playTile, card.SideCards),
			FETiles:  feTiles,
			FECount:  feCount,
			SETiles:  seTiles,
			SECount:  seCount,
			Score:    feCount*1000 + seCount,
		}
		if info.Score > maxScore {
			maxScore = info.Score
		}
		suggestResult = append(suggestResult, info)
	}

	// 将有效分=最大分的牌，设置为推荐
	for _, v := range suggestResult {
		if v.Score == maxScore {
			v.Suggest = true
		}
	}

	costTime := util.GetMillsTime() - startTime

	data := make(map[string]interface{})
	data["errcode"] = 0
	data["errmsg"] = ""
	data["step"] = tileStep
	data["ting"] = len(tingMap) > 0
	sort.Ints(tiles)
	showDebug("tiles:%v", tiles)
	data["tiles"] = tiles
	data["discardTiles"] = discardTiles
	data["guTiles"] = guTiles
	data["diaoTiles"] = diaoTiles
	data["suggest"] = suggestResult
	data["cost_time"] = costTime
	b, _ := json.Marshal(data)
	showDebug("data:%v", string(b))
	w.Write(b)
}

// 生成一个几阶的牌
func autoTiles(tileStep int) []int {
	if tileStep == 99 {
		tiles := make([]int, 14)
		allTiles := util.ShuffleSliceInt(card.MahjongCards108)
		copy(tiles, allTiles[:14])
		return tiles
	}

	for {
		tiles := make([]int, 0, 14)
		if tileStep > 0 && tileStep != 99 {
			for i := 0; i < tileStep; i++ {
				k := util.RandIntn(len(steps))
				tiles = append(tiles, steps[k]...)
			}
		}
		for i := len(tiles); i < 14; i++ {
			k := util.RandIntn(len(cards))
			tiles = append(tiles, cards[k])
		}

		// 检查列表是否合理
		if !checkTiles(tiles) {
			continue
		}
		if tileStep != 99 {
			if step.GetCardsStep(tiles) != tileStep {
				continue
			}
		}
		return tiles
	}
}

func checkTiles(tiles []int) bool {
	if !util.IntInSlice(len(tiles), []int{2, 5, 8, 11, 14}) {
		return false
	}
	m := make(map[int]int)
	for _, tile := range tiles {
		m[tile]++
		if m[tile] > 4 {
			return false
		}
	}
	return true
}

// 验证token
func verifyToken(token string) error {
	// 如果未配置验签地址，则默认是正确的
	if len(core.AppConfig.SelectorTokenVerifyURL) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s?token=%s", core.AppConfig.SelectorTokenVerifyURL, token)
	showDebug("[verifyToken]url:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		showError("[httpGet]het.Get, url:%v, error:%v", url, err.Error())
		return errors.New("请求验证token地址失败")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		showError("[httpGet]het.Get read body, url:%v, error:%v", url, err.Error())
		return errors.New("等待验证token返回结果失败")
	}

	showDebug("[verifyToken]response:%v", string(body))

	if string(body) != "0" {
		return errors.New("验证token失败")
	}

	return nil
}

// 显示客户端错误
func showError(a string, b ...interface{}) {
	fmt.Println("[", util.GetTimestamp(), "]", "[ERROR]", fmt.Sprintf(a, b...))
}

// 显示客户端调试信息
// 显示客户端错误
func showDebug(a string, b ...interface{}) {
	fmt.Println("[", util.GetTimestamp(), "]", "[DEBUG]", fmt.Sprintf(a, b...))
}
