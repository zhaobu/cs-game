package step

import (
	"sort"

	"github.com/fwhappy/util"
	"mahjong.go/mi/win"
)

// GetCardsStep 计算当前手牌所处牌阶
// fixme 暂未考虑7对情形
func GetCardsStep(showTiles []int) int {
	var cardsStep int
	var sortedCards = util.SliceCopy(showTiles)
	// 升序排列
	sort.Ints(sortedCards)

	// 计算不抽对的时候的牌阶
	cardsStep = findSequenceOrTripletCnt(sortedCards)

	// 找到所有的对
	var pos = win.FindPairPos(sortedCards)

	// 遍历所有的对牌，找出离胡最近的抽法，算出牌处于第几阶
	var lastPairTile int // 上次做为对的牌
	for _, v := range pos {
		// 避免有4张同样手牌时，多判断一次
		if sortedCards[v] == lastPairTile {
			continue
		} else {
			lastPairTile = sortedCards[v]
		}
		cards := win.RemovePair(sortedCards, v)
		cnt := findSequenceOrTripletCnt(cards)
		if cnt >= cardsStep {
			cardsStep = cnt + 1
		}
	}

	// 兼容7对
	if len(pos) >= 5 && len(pos) > cardsStep+2 {
		cardsStep = len(pos) - 2
	}

	return cardsStep
}

// 找出排序好的牌中，刻和顺子的数量
func findSequenceOrTripletCnt(sortedCards []int) int {
	// 首先按胡牌的胡牌逻辑取一次，看cnt是多少
	cnt, _ := win.FindSequenceOrTripletCnt(sortedCards)

	cnt2 := findSequenceOrTripletCntPriorityTriplet(sortedCards)

	if cnt > cnt2 {
		return cnt
	}
	return cnt2
}

func findSequenceOrTripletCntPriorityTriplet(sortedCards []int) int {
	tripletCnt, sortedCards := findAndRemoveTriplet(sortedCards)
	// 如果找不到刻，上层函数的计算肯定就是正确的，没必要再去找顺了
	if tripletCnt > 0 {
		shunCnt, _ := win.FindSequenceOrTripletCnt(sortedCards)
		return tripletCnt + shunCnt
	}

	return tripletCnt
}

// 找到刻的个数，并移除刻
func findAndRemoveTriplet(cards []int) (int, []int) {
	m := util.SliceToMap(cards)
	tripletCnt := 0
	for tile, tileCnt := range m {
		if tileCnt < 3 {
			continue
		}
		tripletCnt++
		m[tile] -= 3
	}

	cards = util.MapToSlice(m)
	sort.Ints(cards)
	return tripletCnt, cards
}
