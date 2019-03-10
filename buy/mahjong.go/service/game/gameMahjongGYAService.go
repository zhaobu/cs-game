package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongGYA 贵阳麻将（全鸡）结构定义
type MahjongGYA struct {
	Mahjong
}

// NewMahjongGY 新开一局贵阳麻将
func NewMahjongGYA(room *Room) *MahjongGYA {
	var mahjongGYA = &MahjongGYA{}

	// 麻将初始化
	mahjongGYA.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongGYA.scoreMap = config.GY_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongGYA.setting.SetEnableLack()
	}

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongGYA.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongGYA.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongGYA
}
