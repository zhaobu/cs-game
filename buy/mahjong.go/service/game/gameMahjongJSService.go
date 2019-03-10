package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongJS 金沙麻将
type MahjongJS struct {
	Mahjong
}

// NewMahjongJS 新开一局金沙麻将
func NewMahjongJS(room *Room) *MahjongJS {
	mahjongJS := &MahjongJS{}
	// 麻将初始化
	mahjongJS.Mahjong.begin(room)
	// 初始化积分对照表
	mahjongJS.scoreMap = config.JS_SCORE
	// 麻将总张数定义
	mahjongJS.TileWall.SetTiles(card.MahjongCards112)
	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 {
		mahjongJS.setting.SetEnableLack()
	}
	// 支持满鸡算清一色
	mahjongJS.setting.EnableFullChiken = true
	// 支持合浦小七对
	mahjongJS.setting.EnableHePu7Dui = true
	// 支持银鸡
	mahjongJS.setting.EnableSilverChiken = true
	// 支持钻石鸡
	mahjongJS.setting.EnableDiamondChiken = true
	// 不支持双龙七对
	mahjongJS.setting.EnableShuangLongQiDui = false
	// 不管对方是不是叫牌，都需要包杠
	mahjongJS.setting.BaoKongNeedTing = false

	return mahjongJS
}
