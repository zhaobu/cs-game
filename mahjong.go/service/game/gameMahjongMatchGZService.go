package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongMatchGZ 贵州麻将比赛
type MahjongMatchGZ struct {
	Mahjong
}

// NewMahjongMatchGZ 新开一局贵阳麻将
func NewMahjongMatchGZ(room *Room) *MahjongMatchGZ {
	var matchGZ = &MahjongMatchGZ{}

	// 麻将初始化
	matchGZ.Mahjong.begin(room)

	// 初始化积分对照表
	matchGZ.scoreMap = config.MATCH_GZ_SCORE
	// 积分倍数
	if multiple, exists := config.MatchScoreMultipleList[matchGZ.MType]; exists {
		matchGZ.setting.Multiple = multiple
	}

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		matchGZ.TileWall.SetTiles(card.MahjongCards108)
	} else {
		matchGZ.TileWall.SetTiles(card.MahjongCards72)
	}

	return matchGZ
}
