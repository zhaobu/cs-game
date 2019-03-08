package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongKL 凯里麻将
type MahjongKL struct {
	Mahjong
}

// NewMahjongKL 新开一局凯里麻将
func NewMahjongKL(room *Room) *MahjongKL {
	mahjongKL := &MahjongKL{}

	// 麻将初始化
	mahjongKL.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongKL.scoreMap = config.KL_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongKL.setting.SetEnableLack()
	}
	// 支持庄家翻倍
	mahjongKL.setting.EnableDoubleDealer = true
	// 不支持双龙七对
	mahjongKL.setting.EnableShuangLongQiDui = false
	// 不支持单调
	mahjongKL.setting.EnableDanDiao = false

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongKL.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongKL.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongKL
}
