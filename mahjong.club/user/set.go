package user

import (
	"sync"

	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"
)

// Set 用户集合
type Set struct {
	Mux   *sync.RWMutex
	Users *sync.Map
}

// NewSet 生成一个用户集合
func NewSet() *Set {
	s := &Set{}
	s.Users = &sync.Map{}
	s.Mux = &sync.RWMutex{}
	return s
}

// Len 用户个数
func (s *Set) Len() int {
	return util.SMapLen(s.Users)
}

// Get 根据Id读取用户信息
func (s *Set) Get(id int) (*User, bool) {
	u, ok := s.Users.Load(id)
	if ok {
		return u.(*User), ok
	}
	return nil, ok
}

// Add 将用户加入到用户集合
func (s *Set) Add(u *User) {
	s.Users.Store(u.ID, u)
}

// Del 将用户从集合移除
func (s *Set) Del(id int) {
	s.Users.Delete(id)
}

// IsExists 判断用户是否存在于集合中
func (s *Set) IsExists(id int) bool {
	_, ok := s.Users.Load(id)
	return ok
}

// SendMessageByID 通过id给用户发送一条消息
func (s *Set) SendMessageByID(id int, imPacket *protocal.ImPacket) bool {
	if u, online := s.Get(id); online {
		u.AppendMessage(imPacket)
		return true
	}
	return false
}
