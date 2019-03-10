package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongBJ 毕节麻将
type MahjongBJ struct {
	Mahjong
}

// NewMahjongBJ 新开一局毕节麻将
func NewMahjongBJ(room *Room) *MahjongBJ {
	mahjongBJ := &MahjongBJ{}

	// 麻将初始化
	mahjongBJ.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongBJ.scoreMap = config.BJ_SCORE

	// 是否需要定缺
	if room.setting.GetSettingPlayerCnt() != 4 && room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongBJ.setting.SetEnableLack()
	}
	// 毕节麻将支持平胡自摸
	mahjongBJ.setting.EnablePinghuZimo = true
	// 毕节麻将支持杠幺鸡
	mahjongBJ.setting.EnableBaoKongBam1 = true

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongBJ.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongBJ.TileWall.SetTiles(card.MahjongCards72)
	}

	return mahjongBJ
}
