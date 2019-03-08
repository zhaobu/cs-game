package model

import (
	"sync"

	"github.com/fwhappy/util"
)

// UserSet 用户集合
type UserSet struct {
	users *sync.Map
}

// NewUserSet 生成一个用户集合
func NewUserSet() *UserSet {
	return &UserSet{users: &sync.Map{}}
}

// Len 用户个数
func (us *UserSet) Len() int {
	return util.SMapLen(us.users)
}

// LoadUsers 获取所有用户集合
func (us *UserSet) LoadUsers() *sync.Map {
	return us.users
}

// Load 根据Id读取用户信息
func (us *UserSet) Load(id int) *User {
	u, ok := us.users.Load(id)
	if ok {
		return u.(*User)
	}
	return nil
}

// Store 存储用户(insert or update)
func (us *UserSet) Store(u *User) {
	us.users.Store(u.UserId, u)
}

// Delete 从集合移除用户
func (us *UserSet) Delete(id int) {
	us.users.Delete(id)
}

// IsExists 判断用户是否存在于集合中
func (us *UserSet) IsExists(id int) bool {
	_, ok := us.users.Load(id)
	return ok
}
