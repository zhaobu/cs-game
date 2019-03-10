package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongAS 贵阳麻将结构定义
type MahjongAS struct {
	Mahjong
}

// NewMahjongAS 新开一局贵阳麻将
func NewMahjongAS(room *Room) *MahjongAS {
	var mahjongAS = &MahjongAS{}

	// 麻将初始化
	mahjongAS.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongAS.scoreMap = config.AS_SCORE

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongAS.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongAS.TileWall.SetTiles(card.MahjongCards72)
	}

	// 安顺麻将，明杠不算通行证
	mahjongAS.setting.EnableKongTXZ = false
	// 不支持双龙七对
	mahjongAS.setting.EnableShuangLongQiDui = false
	// 不支持单调
	mahjongAS.setting.EnableDanDiao = false
	// 不支持包鸡
	mahjongAS.setting.EnableBaoChiken = false
	// 不支持责任鸡
	mahjongAS.setting.EnableResponsibilityBam1 = false
	// 明杠不算热炮
	mahjongAS.setting.EnableKongHotPao = false

	return mahjongAS
}
