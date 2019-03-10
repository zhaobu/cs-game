package game

import (
	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/mi/card"
	"mahjong.go/mi/suggest"
	"mahjong.go/mi/win"
)

// 这里的检查能不能胡，是不算地龙的
func checkHu(hCard []int, sCard []*card.ShowCard) bool {
	return win.CanWin(hCard, []int{})
}

func qingCheck(pai []int, sCard []*card.ShowCard) bool {
	// 普通牌类型数量
	suitMaps := make(map[int]bool)
	// 判断手牌
	for _, tile := range pai {
		if util.IntInSlice(tile, card.QingNotSuitCards) {
			continue
		} else if card.IsSuit(tile) {
			mod := tile / 10
			suitMaps[mod] = true
			if len(suitMaps) > 1 {
				return false
			}
		} else {
			return false
		}
	}
	// 判断明牌
	for _, j := range sCard {
		for _, tile := range j.GetTiles() {
			if card.IsSuit(tile) {
				mod := tile / 10
				suitMaps[mod] = true
				if len(suitMaps) > 1 {
					return false
				}
			}
		}
	}
	return true
}

func getLackCount(lack int, pai []int) (count int) {
	if lack == 0 {
		return 0
	}
	for _, j := range pai {
		if IsSameTileKind(lack, j) {
			count = count + 1
		}
	}
	return count
}

func shuangLong7Dui(pai []int, tile int) bool {
	if len(pai) != 14 {
		return false
	}
	var p = win.FindPairPos(pai)
	if len(p) != 7 {
		return false
	}

	var longCount = 0
	var longFlag = false
	for i := 1; i < len(p); i++ {
		if pai[p[i]] == pai[p[i-1]] {
			longCount++
			if pai[p[i]] == tile {
				longFlag = true
			}
		}
	}
	return longCount >= 2 && longFlag
}

// 是否合浦七对
// pai是已经经过升序排列的数组了
func hepu7Dui(pai []int) bool {
	if len(pai) != 14 {
		return false
	}
	var p = win.FindPairPos(pai)
	if len(p) != 7 {
		return false
	}

	var tripleCount = 0
	for i := 0; i < len(p); {
		if pai[i] == pai[i+2]-1 && pai[i] == pai[i+4]-2 {
			tripleCount++
			i += 6
		} else {
			i += 2
		}
	}

	return tripleCount == 2
}

func long7Dui(pai []int, tile int) bool {
	if len(pai) != 14 {
		return false
	}
	var p = win.FindPairPos(pai)
	if len(p) != 7 {
		return false
	}

	for i := 1; i < len(p); i++ {
		if pai[p[i]] == pai[p[i-1]] && pai[p[i]] == tile {
			return true
		}
	}
	return false
}

// 是否地七对
func di7Dui(pai []int, sCard []*card.ShowCard, tile int) bool {
	if len(pai) != 11 || len(sCard) != 1 || !sCard[0].IsPongTile(tile) {
		return false
	}
	var pos = win.FindPairPos(pai)
	return len(pos) == 5
}

// 是否七对
func allDui(pai []int, sCard []*card.ShowCard) bool {
	var p = win.FindPairPos(pai)
	if len(p) == 7 {
		return true
	}
	return false
}

// 因为牌已经胡了
func danDiao(pai []int) bool {
	return len(pai) == 2 && pai[0] == pai[1]
}

// 大对
func daDui(pai []int) bool {
	var pairCount = 0
	var tmp = util.SliceCopy(pai)
	var count int
	for len(tmp) != 0 {
		tmp, count = delSameFromHead(tmp)
		if count == 2 {
			pairCount++
		} else if count == 3 {
			continue
		} else {
			return false
		}
	}
	return pairCount == 1
}

func delSameFromHead(pai []int) ([]int, int) {
	var end int
	for end = 0; end < len(pai); end++ {
		if pai[end] != pai[0] {
			break
		}
	}
	pai = pai[end:]
	return pai, end
}

// 判断能不能明杠
func canKong(opList []*Operation, tileMap *card.CMap, tile int) ([]*Operation, bool) {
	if !card.CanKong(tile) {
		return opList, false
	}
	if tileMap.GetTileCnt(tile) != 3 {
		return opList, false
	}

	opList = append(opList, NewOperation(fbsCommon.OperationCodeKONG, []int{tile}))
	return opList, true
}

// 检查用户有没有暗杠
func canKongDark(opList []*Operation, tileMap *card.CMap, lack int) ([]*Operation, bool) {
	var hasKongDark = false
	for _, tile := range tileMap.GetUnique() {
		if tileMap.GetTileCnt(tile) == 4 {
			// 不支持缺 或者 不是缺
			if lack == 0 || !IsSameTileKind(lack, tile) {
				opList = append(opList, NewOperation(fbsCommon.OperationCodeKONG_DARK, []int{tile}))
				hasKongDark = true
			}
		}
	}

	return opList, hasKongDark
}

// 检查用户是否可以转弯杠
// 如果杠的牌和lastDrawTile不相同，表示为憨包杠
func canKongTurn(opList []*Operation, showTileList []*card.ShowCard, tileMap *card.CMap) ([]*Operation, bool) {
	var hasKongTurn = false
	for _, showCard := range showTileList {
		if showCard.IsPong() && tileMap.GetTileCnt(showCard.GetTile()) > 0 {
			hasKongTurn = true
			var opCode int
			if showCard.GetTile() == tileMap.GetLastAdd() {
				// 转弯杠
				opCode = fbsCommon.OperationCodeKONG_TURN
			} else {
				// 憨包杠
				opCode = fbsCommon.OperationCodeKONG_TURN_FREE
			}
			opList = append(opList, NewOperation(opCode, []int{showCard.GetTile()}))
		}
	}

	return opList, hasKongTurn
}

// 检查用户的操作列表，是否需要添加一个“取消”的操作，若需要，往操作列表中添加一个“取消”操作
func canPassCancel(opList []*Operation) []*Operation {
	var chooseLen = 0
	for _, op := range opList {
		if op.OperationCode != fbsCommon.OperationCodeTING &&
			op.OperationCode != fbsCommon.OperationCodeDRAW &&
			op.OperationCode != fbsCommon.OperationCodeDRAW_AFTER_KONG &&
			op.OperationCode != fbsCommon.OperationCodePLAY &&
			op.OperationCode != fbsCommon.OperationCodePLAY_SUGGEST &&
			op.OperationCode != fbsCommon.OperationCodeROBOT_PLAY_SUGGEST &&
			op.OperationCode != fbsCommon.OperationCodeTING_PLAY_SUGGEST {
			chooseLen++
			break
		}
	}
	if chooseLen > 0 {
		opList = append(opList, NewOperation(fbsCommon.OperationCodePASS_CANCEL, nil))
	}
	return opList
}

// IsSameTileKind 判断两张牌是否同一花色
// 如果非万、条、筒，需要判断两张牌值是否相同
func IsSameTileKind(tileA, tileB int) bool {
	if card.IsCrak(tileA) {
		return card.IsCrak(tileB)
	} else if card.IsBAM(tileA) {
		return card.IsBAM(tileB)
	} else if card.IsDot(tileA) {
		return card.IsDot(tileB)
	}
	return tileA == tileB
}

// GetSuggestLack 获取推荐的定缺
// 优先选择牌最少的那一门
// 如果牌数相同, 则按照万、条、筒的顺序进行选择
func GetSuggestLack(handTiles []int) int {
	ms := suggest.NewMSelector()
	ms.SetHandTilesSlice(handTiles)
	return ms.GetSuggestLack()
}

// 获取包杠的积分类型
func getBaoKongScoreType(code int) int {
	switch code {
	case fbsCommon.OperationCodeKONG:
		return config.SCORE_TYPE_BAO_KONG
	case fbsCommon.OperationCodeKONG_TURN:
		return config.SCORE_TYPE_BAO_KONG_TURN
	case fbsCommon.OperationCodeKONG_DARK:
		return config.SCORE_TYPE_BAO_KONG_DARK
	default:
		return 0
	}
}
