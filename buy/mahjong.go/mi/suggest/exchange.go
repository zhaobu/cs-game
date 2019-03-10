package suggest

import (
	"mahjong.go/mi/card"
	"mahjong.go/mi/weight"
)

// GetSuggestExchange 根据当前牌型，推荐一张牌
// 通用规则：先打缺，跟ai等级无关
func (ms *MSelector) GetSuggestExchange(handTiles []int, cnt int) []int {
	suggestTiles := []int{}

	// 统计手牌中，万条筒牌的数量
	craks, bams, dots := card.KindCards(handTiles...)
	// 计算万、条、筒的分数
	// 默认一个最大值，不会选
	craksScore, bamsScore, dotsScore := 99999999, 99999999, 99999999
	if len(craks) >= cnt {
		craksScore = weight.GetCardsScore(craks)
	}
	if len(bams) >= cnt {
		bamsScore = weight.GetCardsScore(bams)
	}
	if len(dots) >= cnt {
		dotsScore = weight.GetCardsScore(dots)
	}

	// 找出该换哪个牌
	var kindTiles []int
	if craksScore <= bamsScore && craksScore <= dotsScore {
		kindTiles = craks
	} else if bamsScore < craksScore && bamsScore <= dotsScore {
		kindTiles = bams
	} else {
		kindTiles = dots
	}

	suggestTiles = weight.GetMinWeigthTiles(kindTiles, nil, cnt)

	return suggestTiles
}
