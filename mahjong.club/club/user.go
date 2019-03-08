package club

import (
	"github.com/fwhappy/util"
	"mahjong.club/user"
)

// AddUser 给房间添加一个用户
func (c *Club) AddUser(u *user.User) {
	c.Users.Store(u.ID, u)
}

// DelUser 删除一个房间用户
func (c *Club) DelUser(id int) {
	c.Users.Delete(id)
}

// HasUser 判断用户是否在房间内
func (c *Club) HasUser(id int) bool {
	_, exists := c.Users.Load(id)
	return exists
}

// IsEmpty 判断俱乐部是否为空
func (c *Club) IsEmpty() bool {
	return util.SMapIsEmpty(c.Users)
}
