package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongRH 仁怀麻将(全鸡麻将)
type MahjongRH struct {
	Mahjong
}

// NewMahjongRH 新开一局遵义麻将
func NewMahjongRH(room *Room) *MahjongRH {
	m := &MahjongRH{}

	// 麻将初始化
	m.Mahjong.begin(room)

	// 初始化积分对照表
	m.scoreMap = config.RH_SCORE

	// 麻将总张数定义
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		m.TileWall.SetTiles(card.MahjongCards108)
	} else {
		m.TileWall.SetTiles(card.MahjongCards72)
	}

	m.setting.SetEnablePinghu()                // 允许屁胡
	m.setting.EnableDaKuanZhang = true         // 支持边卡吊
	m.setting.EnableBianKaDiao = true          // 支持大宽张
	m.setting.EnableKongAfterDraw = true       // 支持独立的杠上开花
	m.setting.EnableSilverChiken = true        // 支持银鸡
	m.setting.EnableInitLack = true            // 支持原缺
	m.setting.EnableCharge = false             // 不支持冲锋鸡
	m.setting.EnableResponsibilityBam1 = false // 不支持责任鸡
	// 遵义支持定缺
	if room.setting.GetSettingTileCnt() == config.MAHJONG_TILE_CNT_108 {
		m.setting.SetEnableLack()
	}

	return m
}
