package card

import (
	"sort"
	"sync"

	"github.com/fwhappy/util"
)

// CMap 牌面=>数量
type CMap struct {
	Mux         *sync.RWMutex
	tiles       map[int]int // 手牌列表
	lastAdd     int         // 最后添加的牌
	initTiles   []int       // 初始化的牌
	drawTileCnt int         // 总抓牌数
}

// NewCMap 初始化一个TileMap
func NewCMap() *CMap {
	return &CMap{
		Mux:   &sync.RWMutex{},
		tiles: make(map[int]int),
	}
}

// Len 当前手牌数量
func (cm *CMap) Len() int {
	var l = 0
	for _, cnt := range cm.tiles {
		l += cnt
	}
	return l
}

// SetTiles 初始化手牌
func (cm *CMap) SetTiles(tiles []int) {
	cm.Mux.Lock()
	defer cm.Mux.Unlock()
	for _, tile := range tiles {
		cm.tiles[tile]++
	}
}

// GetTileMap 读取所有牌的列表
// 这里不会主动加锁，在外面用的话，如果用于range，需要手动加锁
func (cm *CMap) GetTileMap() map[int]int {
	return cm.tiles
}

// GetInitTiles 获取初始化的手牌
func (cm *CMap) GetInitTiles() []int {
	return cm.initTiles
}

// InitTiles 初始化手牌
func (cm *CMap) InitTiles(tiles []int) {
	// 记录初始手牌
	cm.initTiles = util.SliceCopy(tiles)
	// 初始化抓牌张数
	cm.drawTileCnt = len(tiles)
	// 添加到手牌
	cm.Mux.Lock()
	defer cm.Mux.Unlock()
	for _, tile := range tiles {
		cm.tiles[tile]++
	}
	cm.lastAdd = tiles[len(tiles)-1]
}

// AddTile 添加手牌
func (cm *CMap) AddTile(tile, cnt int) {
	cm.Mux.Lock()
	defer cm.Mux.Unlock()
	// 增加手牌
	cm.tiles[tile] += cnt
	// 设置最后抓的牌
	cm.lastAdd = tile
	// 记录总抓牌张数
	cm.drawTileCnt++

}

// DelTile 删除手牌
func (cm *CMap) DelTile(tile, cnt int) bool {
	cm.Mux.Lock()
	defer cm.Mux.Unlock()
	if cm.tiles[tile] > cnt {
		cm.tiles[tile] -= cnt
	} else if cm.tiles[tile] == cnt {
		delete(cm.tiles, tile)
	} else {
		return false
	}
	return true
}

// ToSlice 转成slice
func (cm *CMap) ToSlice() []int {
	cm.Mux.RLock()
	defer cm.Mux.RUnlock()
	tiles := []int{}
	for tile, cnt := range cm.tiles {
		for i := 0; i < cnt; i++ {
			tiles = append(tiles, tile)
		}
	}
	sort.Ints(tiles)
	return tiles
}

// ToSortedSlice 转成slice并排序
func (cm *CMap) ToSortedSlice() []int {
	tiles := cm.ToSlice()
	sort.Ints(tiles)
	return tiles
}

// GetUnique 获取独立的牌
func (cm *CMap) GetUnique() []int {
	cm.Mux.RLock()
	defer cm.Mux.RUnlock()
	tiles := []int{}
	for tile := range cm.tiles {
		tiles = append(tiles, tile)
	}
	return tiles
}

// GetTileCnt 获取某张牌的数量
func (cm *CMap) GetTileCnt(tile int) int {
	cm.Mux.RLock()
	defer cm.Mux.RUnlock()
	return cm.tiles[tile]
}

// GetLastAdd 获取最后追加的牌
func (cm *CMap) GetLastAdd() int {
	return cm.lastAdd
}

// GetDrawTileCnt 获取总抓牌数量
func (cm *CMap) GetDrawTileCnt() int {
	return cm.drawTileCnt
}

// IsPlayStatus 是否处于待出牌状态
func (cm *CMap) IsPlayStatus() bool {
	l := cm.Len()
	return util.IntInSlice(l, []int{2, 5, 8, 11, 14})
}
