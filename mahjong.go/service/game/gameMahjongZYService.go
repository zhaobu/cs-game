package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongZY 遵义麻将
type MahjongZY struct {
	Mahjong
}

// NewMahjongZY 新开一局遵义麻将
func NewMahjongZY(room *Room) *MahjongZY {
	mahjongZY := &MahjongZY{}

	// 麻将初始化
	mahjongZY.Mahjong.begin(room)

	// 初始化积分对照表
	mahjongZY.scoreMap = config.ZY_SCORE

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongZY.TileWall.SetTiles(card.MahjongCards108)
	} else {
		mahjongZY.TileWall.SetTiles(card.MahjongCards72)
	}

	// 遵义麻将允许屁胡
	mahjongZY.setting.SetEnablePinghu()

	// 支持边卡吊、大宽张
	mahjongZY.setting.EnableDaKuanZhang = true
	mahjongZY.setting.EnableBianKaDiao = true
	// 支持独立的杠上开花
	mahjongZY.setting.EnableKongAfterDraw = true

	// 遵义支持定缺
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		mahjongZY.setting.SetEnableLack()
	}

	return mahjongZY
}
