package card

import (
	"sort"

	"github.com/fwhappy/util"
)

// IsSide 是否边张
func IsSide(card int) bool {
	return util.IntInSlice(card, SideCards)
}

// IsSideNeighbor 是否次边张
func IsSideNeighbor(card int) bool {
	return util.IntInSlice(card, SideNeighborCards)
}

// IsSuit 是否普通牌
// 普通牌是指万、筒、条
func IsSuit(card int) bool {
	return card > 0 && card < MAHJONG_DOT_PLACEHOLDER
}

// GetSelfAndNeighborCards 获取自身或者相邻的一张牌, 结果需去重
// 不包括隔张
// 1条、1筒、1万只有自己和上一张
// 九条、九筒、九万只有自己和下一张
// 非万、筒、条 只有自己
func GetSelfAndNeighborCards(cards ...int) []int {
	result := []int{}
	for _, card := range cards {
		result = append(result, card)
		// 非普通牌、只有自身
		if !IsSuit(card) {
			continue
		}
		if util.IntInSlice(card, LeftSideCards) {
			result = append(result, card+1)
		} else if util.IntInSlice(card, RightSideCards) {
			result = append(result, card-1)
		} else {
			result = append(result, card-1, card+1)
		}
	}
	return util.SliceUniqueInt(result)
}

// GetRelationTiles 获取有关联的牌
// 包括自己、相邻的、跳张
func GetRelationTiles(cards ...int) []int {
	result := []int{}
	for _, card := range cards {
		result = append(result, card)
		// 非普通牌、只有自身
		if !IsSuit(card) {
			continue
		}

		if util.IntInSlice(card, LeftSideCards) {
			result = append(result, card+1, card+2)
		} else if util.IntInSlice(card, LeftSideNeighborCards) {
			result = append(result, card+1, card+2, card-1)
		} else if util.IntInSlice(card, RightSideNeighborCards) {
			result = append(result, card-1, card-2, card+1)
		} else if util.IntInSlice(card, RightSideCards) {
			result = append(result, card-1, card-2)
		} else {
			result = append(result, card-1, card-2, card+1, card+2)
		}
	}
	return util.SliceUniqueInt(result)
}

// GetGuTiles 获取孤张
func GetGuTiles(cards ...int) []int {
	result := []int{}

	m := util.SliceToMap(cards)
	for card, cnt := range m {
		if cnt > 1 {
			continue
		}
		isGu := true
		for _, rcard := range GetRelationTiles(card) {
			if rcard == card {
				continue
			}
			if _, exists := m[rcard]; exists {
				isGu = false
				break
			}
		}
		if isGu {
			result = append(result, card)
		}
	}
	return result
}

// GetDiaoTiles 获取吊张
func GetDiaoTiles(cards ...int) []int {
	result := []int{}

	m := util.SliceToMap(cards)
	for card, cnt := range m {
		if cnt > 1 {
			continue
		}
		isDiao := false
		for _, rcard := range GetRelationTiles(card) {
			if rcard == card {
				continue
			}
			if _, exists := m[rcard]; exists && (rcard == card-1 || rcard == card+1) {
				isDiao = false
				break
			}
			if _, exists := m[rcard]; exists && (rcard == card-2 || rcard == card+2) {
				isDiao = true
				continue
			}
		}
		if isDiao {
			result = append(result, card)
		}
	}
	return result
}

// IsSameType 检查两张牌是否同类
func IsSameType(checkCard, lackCard int) bool {
	return checkCard/10 == lackCard/10
}

// IsSameSuit 检索传入的牌，是否全是同类普通牌
func IsSameSuit(cards ...int) bool {
	firstMod := -1
	for _, card := range cards {
		// 非普通牌，退出
		if !IsSuit(card) {
			return false
		}
		if firstMod == -1 {
			firstMod = card / 10
		} else {
			// 非同类型
			if firstMod != card/10 {
				return false
			}
		}
	}
	return true
}

// IsCrak 是否万
func IsCrak(card int) bool {
	return card >= MAHJONG_CRAK1 && card <= MAHJONG_CRAK9
}

// IsBAM 是否条
func IsBAM(card int) bool {
	return card >= MAHJONG_BAM1 && card <= MAHJONG_BAM9
}

// IsDot 是否筒
func IsDot(card int) bool {
	return card >= MAHJONG_DOT1 && card <= MAHJONG_DOT9
}

// IsFlower 是否补花
func IsFlower(card int) bool {
	return card == MAHJONG_RED_FLOWER
}

// GetBehindCardCycle 获取某张牌的下一张牌（循环获取）
func GetBehindCardCycle(card int) int {
	var behind int
	if IsSuit(card) {
		if util.IntInSlice(card, RightSideCards) {
			behind = card - 8
		} else {
			behind = card + 1
		}
	}
	return behind
}

// GetFrontCardCycle 获取某张牌的前一张牌（循环获取）
func GetFrontCardCycle(card int) int {
	var front int
	if IsSuit(card) {
		if util.IntInSlice(card, LeftSideCards) {
			front = card + 8
		} else {
			front = card - 1
		}
	}
	return front
}

// CanWin 判断牌是否支持胡
func CanWin(card int) bool {
	return card != MAHJONG_RED_FLOWER
}

// CanKong 判断牌是否支持杠
func CanKong(card int) bool {
	return card != MAHJONG_RED_FLOWER
}

// KindCards 将牌分类，获取万、条、筒各自有哪些
func KindCards(cards ...int) (craks, bams, dots []int) {
	craks = make([]int, 0)
	bams = make([]int, 0)
	dots = make([]int, 0)

	for _, tile := range cards {
		if IsCrak(tile) {
			craks = append(craks, tile)
		} else if IsBAM(tile) {
			bams = append(bams, tile)
		} else if IsDot(tile) {
			dots = append(dots, tile)
		}
	}

	sort.Ints(craks)
	sort.Ints(bams)
	sort.Ints(dots)
	return
}
