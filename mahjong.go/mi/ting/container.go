package ting

// MTContainer 用户听牌容器
type MTContainer struct {
	status int           // 听牌状态
	maps   map[int][]int // 听牌列表
}

// NewMTContainer 初始化听牌容器
func NewMTContainer() *MTContainer {
	return &MTContainer{
		maps:   make(map[int][]int),
		status: STATUS_BAD,
	}
}

// IsBaoTing 是否报听
func (mtc *MTContainer) IsBaoTing() bool {
	return mtc.status == STATUS_BAOTING
}

// IsTing 是否听牌
func (mtc *MTContainer) IsTing() bool {
	return mtc.status == STATUS_TING || mtc.status == STATUS_BAOTING
}

// IsBad 是否处于未初始化听牌状态
func (mtc *MTContainer) IsBad() bool {
	return mtc.status == STATUS_BAD
}

// IsNormal 是否处于未叫牌状态
func (mtc *MTContainer) IsNormal() bool {
	return mtc.status == STATUS_NORMAL
}

// SetStatus 设置听牌状态
func (mtc *MTContainer) SetStatus(status int) {
	mtc.status = status
}

// SetMaps 填充听牌列表
func (mtc *MTContainer) SetMaps(m map[int][]int) {
	mtc.maps = m
}

// SetTingTiles 将tiles切片填充到可听牌
func (mtc *MTContainer) SetTingTiles(tiles []int) {
	mtc.Clean()
	mtc.maps[tiles[0]] = tiles[1:]
	mtc.status = STATUS_TING
}

// AppendTingTiles 将tiles切片填充到可听牌
func (mtc *MTContainer) AppendTingTiles(tiles []int) {
	mtc.maps[tiles[0]] = tiles[1:]
}

// GetStatus 获取听牌状态
func (mtc *MTContainer) GetStatus() int {
	return mtc.status
}

// GetMaps 读取听牌列表
func (mtc *MTContainer) GetMaps() map[int][]int {
	return mtc.maps
}

// GetTingTiles 获取当前所有能听的牌
func (mtc *MTContainer) GetTingTiles() []int {
	for _, v := range mtc.maps {
		return v
	}
	return []int{}
}

// Clean 清空听牌列表
func (mtc *MTContainer) Clean() {
	for k := range mtc.maps {
		delete(mtc.maps, k)
	}
}

// SetBaoTing 设置报听状态
func (mtc *MTContainer) SetBaoTing(tile int) {
	mtc.status = STATUS_BAOTING
	mtc.SetTingTile(tile)
}

// SetNormal 设置为未听牌状态
func (mtc *MTContainer) SetNormal() {
	mtc.status = STATUS_NORMAL
	mtc.Clean()
}

// SetTingTile 设置听某张牌
// 如果这张牌不在可听列表，用户状态会被设置为未听牌
func (mtc *MTContainer) SetTingTile(tile int) {
	for k := range mtc.maps {
		if k != tile {
			delete(mtc.maps, k)
		}
	}
}

// SetTingMap 根据用户打出的牌，更新用户的听牌状态
func (mtc *MTContainer) SetTingMap(tile int) {
	if _, ok := mtc.maps[tile]; ok {
		mtc.status = STATUS_TING
		mtc.SetTingTile(tile)
	} else {
		mtc.status = STATUS_NORMAL
		mtc.Clean()
	}
}
