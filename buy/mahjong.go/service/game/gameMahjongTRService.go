package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongTR 铜仁麻将结构定义
type MahjongTR struct {
	Mahjong
}

// NewMahjongTR 新开一局贵阳麻将
func NewMahjongTR(room *Room) *MahjongTR {
	var mahjongTR = &MahjongTR{}

	// 麻将初始化
	mahjongTR.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongTR.scoreMap = config.TR_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongTR.setting.SetEnableLack()
	}
	// 不支持双龙七对
	mahjongTR.setting.EnableShuangLongQiDui = false
	// 不支持单调
	mahjongTR.setting.EnableDanDiao = false

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongTR.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongTR.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongTR
}
