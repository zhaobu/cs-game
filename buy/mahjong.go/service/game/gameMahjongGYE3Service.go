package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

type MahjongGYE3 struct {
	Mahjong
}

func NewMahjongGYE3(room *Room) *MahjongGYE3 {
	var m = &MahjongGYE3{}

	// 麻将初始化
	m.Mahjong.begin(room)

	// 初始化积分对照表
	m.scoreMap = config.GY_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		m.setting.SetEnableLack()
	}

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		m.TileWall.SetTiles(card.MahjongCards108)
	} else {
		m.TileWall.SetTiles(card.MahjongCards72)
	}

	return m
}
