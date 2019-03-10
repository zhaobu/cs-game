package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongGFT 杠翻天玩法
type MahjongGFT struct {
	Mahjong
}

// NewMahjongGFT 新开一局贵阳麻将
func NewMahjongGFT(room *Room) *MahjongGFT {
	var mahjong = &MahjongGFT{}

	// 麻将初始化
	mahjong.Mahjong.begin(room)

	// 初始化积分对照表
	mahjong.scoreMap = config.GY_SCORE

	// 杠后加倍
	mahjong.setting.EnableKongDouble = true

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjong.setting.SetEnableLack()
	}

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjong.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjong.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjong
}
