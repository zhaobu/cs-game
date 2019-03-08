package user

// Info 用户共享信息
type Info struct {
	ID       int // 用户id，做一下冗余，方便使用
	Nickname string
	Avatar   string
	Score    int
}

// NewInfo 创建一个新的用户共享信息
func NewInfo(id int) *Info {
	return &Info{ID: id}
}
