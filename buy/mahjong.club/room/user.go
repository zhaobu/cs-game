package room

// User 房间用户信息
type User struct {
	ID        int    // 用户id
	Index     int    // 房间位置
	Avatar    string // 用户头像
	Nickname  string // 用户昵称
	AvatarBox int    // 头像框
}

// NewUser 生成一个房间用户信息的对象
func NewUser(id, index int, avatar, nickname string, avatarBox int) *User {
	return &User{
		ID:        id,
		Index:     index,
		Avatar:    avatar,
		Nickname:  nickname,
		AvatarBox: avatarBox,
	}
}
