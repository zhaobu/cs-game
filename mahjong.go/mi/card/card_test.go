package card

import (
	"testing"
)

func TestIsSameCards(t *testing.T) {
	if !IsSameSuit([]int{1, 2, 3, 4, 5, 6}...) {
		t.Error("[IsSameSuit]测试失败1")
	}
	if IsSameSuit([]int{1, 2, 3, 4, 5, 6, 11}...) {
		t.Error("[IsSameSuit]测试失败2")
	}
	if IsSameSuit([]int{42, 1, 2, 3, 4, 5, 6}...) {
		t.Error("[IsSameSuit]测试失败3")
	}
	if !IsSameSuit([]int{22, 23, 21, 25}...) {
		t.Error("[IsSameSuit]测试失败4")
	}
}
