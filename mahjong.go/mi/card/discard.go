package card

// DiscardContainer 弃牌列表
type DiscardContainer struct {
	tiles             []int // 当前弃牌列表
	lastDiscardTile   int   // 最后打出的弃牌（有可能被抓走）
	discardRemoveFlag bool  // 是否允许拿走最后打的牌，防止在一炮多响时，弃牌会被拿多张
	playedTiles       []int // 所有出过的牌
}

// NewDiscardContainer 初始化弃牌容器
func NewDiscardContainer() *DiscardContainer {
	return &DiscardContainer{
		tiles:             make([]int, 0),
		playedTiles:       make([]int, 0),
		discardRemoveFlag: true,
	}
}

// SetRemoveFlag 设置是否允许删除最后一张牌的标志
func (dc *DiscardContainer) SetRemoveFlag(flag bool) {
	dc.discardRemoveFlag = flag
}

// GetTiles 获取所有弃牌
func (dc *DiscardContainer) GetTiles() []int {
	return dc.tiles
}

// AppendTile 添加弃牌
func (dc *DiscardContainer) AppendTile(tile int) {
	// 增加当前弃牌
	dc.tiles = append(dc.tiles, tile)
	// 记录出过的牌
	dc.playedTiles = append(dc.playedTiles, tile)
	// 加了弃牌之后，自动设置为可拿走弃牌
	dc.discardRemoveFlag = true
}

// DelLastTile 删除最后一张弃牌
func (dc *DiscardContainer) DelLastTile() bool {
	// 不允许删除弃牌
	tLen := len(dc.tiles)
	if tLen == 0 || !dc.discardRemoveFlag {
		return false
	}
	dc.tiles = append([]int{}, dc.tiles[:tLen-1]...)
	// 加了弃牌之后，自动设置为不可拿走弃牌
	dc.discardRemoveFlag = false
	return true
}

// Len 弃牌长度
func (dc *DiscardContainer) Len() int {
	return len(dc.tiles)
}

// GetTileCnt 获取某张弃牌的张数
func (dc *DiscardContainer) GetTileCnt(tile int) int {
	var cnt = 0
	for _, card := range dc.tiles {
		if card == tile {
			cnt++
		}
	}
	return cnt
}

// GetPlayedLen 获取总的出牌数量
func (dc *DiscardContainer) GetPlayedLen() int {
	return len(dc.playedTiles)
}

// GetTilePlayedCnt 获取某张牌的出牌张数
func (dc *DiscardContainer) GetTilePlayedCnt(tile int) int {
	var cnt = 0
	for _, card := range dc.playedTiles {
		if card == tile {
			cnt++
		}
	}
	return cnt
}

// HasPlayedTile 是否出过某张牌
func (dc *DiscardContainer) HasPlayedTile(tile int) bool {
	for _, discard := range dc.playedTiles {
		if discard == tile {
			return true
		}
	}
	return false
}
