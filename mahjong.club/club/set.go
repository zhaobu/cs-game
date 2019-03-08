package club

import (
	"sync"

	"github.com/fwhappy/util"
)

// Set 房间集合
type Set struct {
	Mux   *sync.RWMutex
	Clubs *sync.Map
}

// NewSet 生成一个俱乐部列表
func NewSet() *Set {
	s := &Set{}
	s.Clubs = &sync.Map{}
	s.Mux = &sync.RWMutex{}
	return s
}

// Len 俱乐部个数
func (s *Set) Len() int {
	return util.SMapLen(s.Clubs)
}

// Get 根据Id读取俱乐部信息
func (s *Set) Get(id int) (*Club, bool) {
	c, ok := s.Clubs.Load(id)
	if ok {
		return c.(*Club), ok
	}
	return nil, ok
}

// Add 将俱乐部加入到俱乐部集合
func (s *Set) Add(c *Club) {
	s.Clubs.Store(c.ID, c)
}

// Del 将俱乐部从集合移除
func (s *Set) Del(id int) {
	s.Clubs.Delete(id)
}

// IsExists 判断俱乐部是否存在于集合中
func (s *Set) IsExists(id int) bool {
	_, ok := s.Clubs.Load(id)
	return ok
}
