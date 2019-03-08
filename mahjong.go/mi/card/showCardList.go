package card

import (
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/mi/oc"
)

// ShowCardList 明牌列表
type ShowCardList struct {
	showCards []*ShowCard
}

// NewShowCardList 新建一个明牌列表
func NewShowCardList() *ShowCardList {
	sl := &ShowCardList{
		showCards: make([]*ShowCard, 0),
	}
	return sl
}

func (sl *ShowCardList) string() string {
	s := ""
	for _, sc := range sl.showCards {
		s += sc.string()
	}
	return s
}

// Len 明牌长度
func (sl *ShowCardList) Len() int {
	return len(sl.showCards)
}

// GetAll 获取明牌列表
func (sl *ShowCardList) GetAll() []*ShowCard {
	return sl.showCards
}

// GetAllTiles 获取明牌的所有牌
func (sl *ShowCardList) GetAllTiles() []int {
	tiles := []int{}
	for _, sc := range sl.showCards {
		tiles = append(tiles, sc.GetTiles()...)
	}
	return tiles
}

// GetTileCnt 获取明牌中某张牌的数量
// 暂时没有吃，所以忽略吃的情况
func (sl *ShowCardList) GetTileCnt(tile int) int {
	var cnt = 0
	for _, sc := range sl.showCards {
		if sc.GetTile() == tile {
			cnt += sc.GetTilesLen()
		}
	}
	return cnt
}

// GetLast 获取最后一个明牌
func (sl *ShowCardList) GetLast() *ShowCard {
	if sl.Len() == 0 {
		return nil
	}
	return sl.showCards[sl.Len()-1]
}

// Add 添加一个明牌到明牌列表
func (sl *ShowCardList) Add(sc *ShowCard) {
	sl.showCards = append(sl.showCards, sc)
}

// AddFlower 添加一个花到明牌
func (sl *ShowCardList) AddFlower(target, tile, tileCnt int) {
	var sc *ShowCard
	for _, s := range sl.GetAll() {
		if s.IsFlower(tile) {
			sc = s
			break
		}
	}
	tiles := make([]int, tileCnt)
	for i := 0; i < tileCnt; i++ {
		tiles[i] = tile
	}
	if sc == nil {
		sc = NewShowCard(fbsCommon.OperationCodeFLOWER_CHANGE, target, tiles, false)
		sl.Add(sc)
	} else {
		sc.tiles = append(sc.tiles, tiles...)
	}
}

// DelKongTurnTile 删除一张转弯杠的牌
func (sl *ShowCardList) DelKongTurnTile(tile int) {
	for _, sc := range sl.showCards {
		if oc.IsKongTurnOperation(sc.GetOpCode()) && sc.GetTile() == tile && sc.GetTilesLen() == 4 {
			sc.ModifyQiangKong()
		}
	}
}

// HasPongOrKongTile 是否碰了或者杠了某张牌
func (sl *ShowCardList) HasPongOrKongTile(tile int) bool {
	for _, sc := range sl.showCards {
		if sc.IsPongTile(tile) || sc.IsKongTile(tile) {
			return true
		}
	}
	return false
}

// SetPongToKongTurn 将某张牌的碰设置为转弯杠
// 收费的是转弯杠
// 不收费的是憨包杠
func (sl *ShowCardList) SetPongToKongTurn(tile, opCode int) {
	var free = false
	if opCode == fbsCommon.OperationCodeKONG_TURN_FREE {
		free = true
	}
	for _, sc := range sl.showCards {
		if sc.IsPongTile(tile) {
			sc.ModifyPongToKong(opCode, free)
			return
		}
	}
}
