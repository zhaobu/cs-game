package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongZYA 遵义麻将(全鸡麻将)
type MahjongZYA struct {
	Mahjong
}

// NewMahjongZYA 新开一局遵义麻将
func NewMahjongZYA(room *Room) *MahjongZYA {
	mahjongZYA := &MahjongZYA{}

	// 麻将初始化
	mahjongZYA.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongZYA.scoreMap = config.ZY_SCORE

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongZYA.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongZYA.TileWall.SetTiles(card.MahjongCards72)
	}

	// 遵义麻将允许屁胡
	mahjongZYA.setting.SetEnablePinghu()

	// 支持边卡吊、大宽张
	mahjongZYA.setting.EnableDaKuanZhang = true
	mahjongZYA.setting.EnableBianKaDiao = true
	// 支持独立的杠上开花
	mahjongZYA.setting.EnableKongAfterDraw = true

	// 遵义支持定缺
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongZYA.setting.SetEnableLack()
	}

	return mahjongZYA
}
