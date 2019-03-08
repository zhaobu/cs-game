package wall

import (
	"fmt"

	"github.com/fwhappy/util"
)

// WExpect 期望的牌的容器
type WExpect struct {
	DisabledIndex []int
}

// SetExpectDisabled 设置哪些牌不能被期待
func (w *Wall) SetExpectDisabled(index int) {
	if index > len(w.expect.DisabledIndex)-1 {
		fmt.Printf("[wall.SetExpectDisabled]index out of range, index:%v, len(tiles):%v, len(w.expect.DisabledIndex):%v", index, len(w.tiles), len(w.expect.DisabledIndex))
	} else {
		w.expect.DisabledIndex[index] = 1
	}

}

// StatExpect 计算需要的牌中，可期待的有多少张，总数有多少张
func (w *Wall) StatExpect(tiles []int) (canExpect, total int) {
	index := w.getForwardNextIndex()
	for i := index; i < len(w.tiles); i++ {
		if w.expect.isDisabled(i) {
			continue
		}
		total++
		if util.IntInSlice(w.GetTile(i), tiles) {
			canExpect++
		}
	}
	return
}

// GetExpectTiles 获取所有未被抓和未被看过的牌
func (w *Wall) GetExpectTiles() []int {
	tiles := []int{}
	index := w.getForwardNextIndex()
	for i := index; i < len(w.tiles); i++ {
		if w.expect.isDisabled(i) {
			continue
		}
		tiles = append(tiles, w.GetTile(i))
	}
	return tiles
}

// 初始化，
func (we *WExpect) initializtion(length int) {
	we.DisabledIndex = make([]int, length, length)
}

func (we *WExpect) isDisabled(index int) bool {
	return we.DisabledIndex[index] == 1
}

// ForwardExpect 设置期望从前面抓到哪些牌
func (w *Wall) ForwardExpect(tiles []int) {
	index := w.getForwardNextIndex()
	// 如果下一张被禁止交换，直接返回
	if w.expect.isDisabled(index) {
		return
	}

	// 如果已经是最后一张了，直接返回
	if w.RemainLength() == 1 {
		return
	}

	for i := index + 1; i < len(w.tiles); i++ {
		if w.expect.isDisabled(i) {
			continue
		}
		if util.IntInSlice(w.GetTile(i), tiles) {
			w.exchange(index, i)
		}
	}
}

// ForwardUnexpect 设置期望从前面不要抓到哪些牌
func (w *Wall) ForwardUnexpect(tiles []int) {
	index := w.getForwardNextIndex()

	// 如果下一张被禁止交换，直接返回
	if w.expect.isDisabled(index) {
		return
	}

	// 如果已经是最后一张了，直接返回
	if w.RemainLength() == 1 {
		return
	}

	for i := index + 1; i < len(w.tiles); i++ {
		if w.expect.isDisabled(i) {
			continue
		}
		if !util.IntInSlice(w.GetTile(i), tiles) {
			if index != i {
				w.exchange(index, i)
			}
			break
		}
	}
}

// BackwardExpect 设置期望从后面抓到哪些牌
func (w *Wall) BackwardExpect(tiles []int) {
	index := w.getBackwordNextIndex()

	// 如果下一张被禁止交换，直接返回
	if w.expect.isDisabled(index) {
		return
	}

	// 如果已经是最后一张了，直接返回
	if w.RemainLength() == 1 {
		return
	}

	for i := w.getForwardNextIndex(); i < len(w.tiles); i++ {
		if w.expect.isDisabled(i) {
			continue
		}

		if util.IntInSlice(w.GetTile(i), tiles) {
			if index != i {
				w.exchange(index, i)
			}
			break
		}
	}
}

// BackwardUnexpect 设置期望从前面不要抓到哪些牌
func (w *Wall) BackwardUnexpect(tiles []int) {
	index := w.getBackwordNextIndex()

	// 如果下一张被禁止交换，直接返回
	if w.expect.isDisabled(index) {
		return
	}

	// 如果已经是最后一张了，直接返回
	if w.RemainLength() == 1 {
		return
	}

	for i := w.getForwardNextIndex(); i < len(w.tiles); i++ {
		if w.expect.isDisabled(i) {
			continue
		}
		if !util.IntInSlice(w.GetTile(i), tiles) {
			if index != i {
				w.exchange(index, i)
			}
			break
		}
	}
}
