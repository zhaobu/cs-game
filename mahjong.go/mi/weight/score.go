package weight

import (
	"sort"

	"github.com/fwhappy/util"
	"mahjong.go/mi/win"
)

// GetCardsScore 获取牌型得分
// 评估得分 = 牌数得分 + 牌型得分
// 牌数得分 = 每张牌3分，有几张牌加计分
// 牌型得分 = 杠10分>刻子7分>顺子3分>对子3分>挨张2分
// 牌型得分中一张牌只能计算一次；
func GetCardsScore(cards []int) (score int) {
	if len(cards) > 0 {
		score += GetCardsNumScore(cards)
		score += GetCardsTypeScore(cards)
	}
	return
}

// GetCardsNumScore 计算牌数得分
// 牌数得分 = 每张牌3分，有几张牌加计分
func GetCardsNumScore(cards []int) int {
	return len(cards) * 3
}

// GetCardsTypeScore 计算牌型分
// 牌型得分 = 杠10分>刻子7分>顺子3分>对子3分>挨张2分
func GetCardsTypeScore(cards []int) int {
	score := 0
	m := util.SliceToMap(cards)
	for tile, cnt := range m {
		// 抽杠
		if cnt == 4 {
			score += 10
			delete(m, tile)
		}
		// 抽刻
		if cnt == 3 {
			score += 7
			delete(m, tile)
		}
	}
	// 抽顺子
	s := util.MapToSlice(m)
	sort.Ints(s)
	for {
		find := win.FindAndRemoveSequence(&s)
		if find {
			score += 3
		} else {
			break
		}
	}

	// 抽对子
	m = util.SliceToMap(s)
	for tile, cnt := range m {
		// 抽对
		if cnt == 2 {
			score += 3
			delete(m, tile)
		}
	}
	// 挨张加分
	s = util.MapToSlice(m)
	for i := 0; i < len(s)-1; i++ {
		if s[i]+1 == s[i+1] {
			score += 2
			i++
		}
	}

	return score
}
