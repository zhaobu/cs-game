package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongGY 贵阳麻将结构定义
type MahjongGY struct {
	Mahjong
}

// NewMahjongGY 新开一局贵阳麻将
func NewMahjongGY(room *Room) *MahjongGY {
	var mahjongGY = &MahjongGY{}

	// 麻将初始化
	mahjongGY.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongGY.scoreMap = config.GY_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongGY.setting.SetEnableLack()
	}

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongGY.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongGY.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongGY
}
