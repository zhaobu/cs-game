package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongXY 贵阳麻将结构定义
type MahjongXY struct {
	Mahjong
}

// NewMahjongXY 新开一局贵阳麻将
func NewMahjongXY(room *Room) *MahjongXY {
	var mahjongXY = &MahjongXY{}

	// 麻将初始化
	mahjongXY.Mahjong.begin(room)

	// 初始化积分对照表
	// 与贵阳麻将一致
	mahjongXY.scoreMap = config.XY_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongXY.setting.SetEnableLack()
	}

	// 不支持双龙七对
	mahjongXY.setting.EnableShuangLongQiDui = false
	// 不支持单调
	mahjongXY.setting.EnableDanDiao = false

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongXY.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongXY.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongXY
}
