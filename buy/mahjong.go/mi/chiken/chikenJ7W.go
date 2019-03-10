package chiken

import (
	"mahjong.go/mi/card"
)

// GetJ7W 获取哪些牌属于见7挖
func (mc *MChiken) GetJ7W() []int {
	return []int{card.MAHJONG_DOT7, card.MAHJONG_CRAK7, card.MAHJONG_BAM7}
}
