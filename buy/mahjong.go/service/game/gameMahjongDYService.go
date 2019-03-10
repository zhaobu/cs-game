package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongDY 都匀麻将结构定义
type MahjongDY struct {
	Mahjong
}

// NewMahjongDY 新开一局贵阳麻将
func NewMahjongDY(room *Room) *MahjongDY {
	var mahjongDY = &MahjongDY{}

	// 麻将初始化
	mahjongDY.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongDY.scoreMap = config.DY_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongDY.setting.SetEnableLack()
	}
	// 不支持双龙七对
	mahjongDY.setting.EnableShuangLongQiDui = false

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongDY.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongDY.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongDY
}
