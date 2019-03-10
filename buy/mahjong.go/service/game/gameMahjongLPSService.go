package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongLPS 六盘水麻将结构定义
type MahjongLPS struct {
	Mahjong
}

// NewMahjongLPS 新开一局六盘水麻将
func NewMahjongLPS(room *Room) *MahjongLPS {
	var mahjongLPS = &MahjongLPS{}

	// 麻将初始化
	mahjongLPS.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongLPS.scoreMap = config.LPS_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongLPS.setting.SetEnableLack()
	}
	// 不支持双龙七对
	mahjongLPS.setting.EnableShuangLongQiDui = false
	// 报听状态下，允许明杠
	mahjongLPS.setting.EnableBaoTingKong = true

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongLPS.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongLPS.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongLPS
}
