package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongQX 黔西麻将
type MahjongQX struct {
	Mahjong
}

// NewMahjongQX 新开一局黔西麻将
func NewMahjongQX(room *Room) *MahjongQX {
	mahjongQX := &MahjongQX{}

	// 麻将初始化
	mahjongQX.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongQX.scoreMap = config.QX_SCORE

	// 支持银鸡
	mahjongQX.setting.EnableSilverChiken = true
	// 包杠不需要对方叫牌
	mahjongQX.setting.BaoKongNeedTing = false
	// 包鸡不需要对方叫牌
	mahjongQX.setting.BaoChikenNeedTing = false

	// 麻将总张数定义
	mahjongQX.TileWall.SetTiles(card.MahjongCards112_FLOWER)

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 {
		mahjongQX.setting.SetEnableLack()
	}

	return mahjongQX
}
