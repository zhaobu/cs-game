package game

import "testing"

func TestHepu7Dui(t *testing.T) {
	if !hepu7Dui([]int{1, 1, 2, 2, 3, 3, 7, 7, 8, 8, 9, 9, 11, 11}) {
		t.Error("[TestHepu7Dui]失败1")
	}
	if hepu7Dui([]int{1, 1, 2, 2, 7, 7, 8, 8, 9, 9, 11, 11, 12, 12}) {
		t.Error("[TestHepu7Dui]失败2")
	}
	if hepu7Dui([]int{1, 1, 2, 2, 3, 3, 7, 7, 8, 8, 9, 9, 11, 12}) {
		t.Error("[TestHepu7Dui]失败3")
	}
}

func BenchmarkHepu7Dui(b *testing.B) {
	b.StopTimer()
	tiles := []int{1, 1, 2, 2, 3, 3, 7, 7, 8, 8, 9, 9, 11, 11}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		hepu7Dui(tiles)
	}
}
