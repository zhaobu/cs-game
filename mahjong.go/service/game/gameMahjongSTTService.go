package game

import (
	"mahjong.go/config"
	"mahjong.go/mi/card"
)

// MahjongSST 72张玩法
type MahjongSST struct {
	Mahjong
}

// NewMahjongSTT 创建一个72张麻将玩法
// 新开一局
func NewMahjongSTT(room *Room) *MahjongSST {
	mahjongSST := &MahjongSST{}

	// 麻将初始化
	mahjongSST.Mahjong.begin(room)

	// 初始化积分对照表
	// 与三丁拐保持一致
	mahjongSST.scoreMap = config.SD_SCORE

	// 72张
	mahjongSST.TileWall.SetTiles(card.MahjongCards72)

	return mahjongSST
}
