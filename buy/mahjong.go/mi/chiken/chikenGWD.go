package chiken

import (
	"mahjong.go/mi/card"
)

// GetGWD 获取哪些牌属于高挖弹
func (mc *MChiken) GetGWD() []int {
	return []int{card.MAHJONG_CRAK9, card.MAHJONG_CRAK7, card.MAHJONG_BAM8}
}
