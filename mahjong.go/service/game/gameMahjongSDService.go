package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongSD 三丁拐
type MahjongSD struct {
	Mahjong
}

// NewMahjongSD 新开一局三丁拐
func NewMahjongSD(room *Room) *MahjongSD {
	mahjongSD := &MahjongSD{}

	// 麻将初始化
	mahjongSD.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongSD.scoreMap = config.SD_SCORE

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongSD.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongSD.TileWall.SetTiles(card.MahjongCards72)
	}

	// 三丁拐支持定缺
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongSD.setting.SetEnableLack()
	}

	return mahjongSD
}
