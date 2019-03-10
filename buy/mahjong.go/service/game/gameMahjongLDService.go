package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongLD 两丁拐
type MahjongLD struct {
	Mahjong
}

// NewMahjongLD 创建一个两丁拐麻将玩法
// 新开一局
func NewMahjongLD(room *Room) *MahjongLD {
	mahjongLD := &MahjongLD{}

	// 麻将初始化
	mahjongLD.Mahjong.begin(room)

	// 初始化积分对照表
	// 与三丁拐保持一致
	mahjongLD.scoreMap = config.SD_SCORE

	// 麻将张数
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongLD.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongLD.TileWall.SetTiles(card.MahjongCards72)
	}

	// 两丁需要定缺
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongLD.setting.SetEnableLack()
	}

	return mahjongLD
}
