package config

// 用户基本信息表
type User struct {
	UserId        int `orm:"pk"`
	Money         int
	GiftMoney     int
	Status        int
	LastLoginTime int64
	RegisterTime  int64
	Nickname      string
	IconUrl       string
	TypeId        int
	Unionid       string
}

// 用户扩展信息表
type UserInfo struct {
	UserId   int `orm:"pk"`
	InfoType int
	Info     string
	Time     int64
}
type UserInfoList map[int]UserInfo

// 用户消费日志
type UserAccountLog struct {
	LogId      int `orm:"pk"`
	UserId     int
	Money      int
	GiftMoney  int
	CreateTime int64
	Sn         string
	ChangeType string
	OrderId    int
}

// 消费日志
type UserConsumeInfo struct {
	Id         int `orm:"pk"`
	Sn         string
	UserId     int
	Num        int
	GiftNum    int
	Note       string
	Ctype      int // 消耗类型
	CreateTime int64
}

// 收入日志
type UserTransInfo struct {
	Id           int `orm:"pk"`
	Sn           string
	UserId       int
	TargetUserId int
	Num          int
	CreateTime   int64
	TransType    int
	DiamondType  int
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 牌局相关的记录
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 游戏信息
type GameInfo struct {
	RoomId        int64  `orm:"pk"` //系统房间id
	RoomNum       string //  房间号
	CreatorUserId int    // 房间创建者
	GameType      int    // 房间类型 0=自己创建 1=随机建立
	MahjongType   int    // 麻将类型
	Setting       string // 房间扩展玩法
	TotalRounds   int    // 总局数
	PlayerCount   int    // 参与人数
	Players       string // 参与者(eg:1,2,3,4)
	CreateTime    int64  // 房间建立时间
	StartTime     int64  // 房间开始时间
	ServerRemote  string // 房间所属服务器
	RoomMaster    int    // 房主id
}

// 游戏结果
type GameResult struct {
	RoomId       int64  `orm:"pk"` // 系统房间id
	RoomNum      string //  房间号
	GameType     int    // 房间类型 0=自己创建 1=随机建立
	MahjongType  int    // 麻将类型
	TotalRounds  int    // 总局数
	Players      string // 参与者(eg:1,2,3,4)
	Scores       string // 玩家积分[{userId:value,score:value},...]
	CreateTime   int64  // 开房时间
	StartTime    int64  // 开始时间
	CompleteTime int64  // 完成时间
	IsDismiss    int    // 是否解散
	DismissUsers string // 解散玩家,第一个是发起者({userId,...})
	PlayRounds   int    // 完成到多少局
}

// 用户游戏结果,输赢记录
type GameUserRecords struct {
	Id         int64 `orm:"pk"`
	RoomId     int64
	UserId     int
	Score      int
	Wins       int
	Loses      int
	CreateTime int64
}

// 游戏每一回合信息
type GameRoundData struct {
	Id            int64 `orm:"pk"`
	RoomId        int64
	Round         int    // 回合数
	Scores        string // 玩家积分[{userId:value,score:value},...]
	Data          string // 游戏数据,用户回放、分析等
	Huang         int    // 是否黄牌
	WinPlayers    string // 胡牌的人
	WinPlayersCnt int    // 胡牌人数
	GoldBam1      int    // 是否金鸡
	GoldDot8      int    // 是否金乌骨鸡
	StartTime     int64  // 本局开始时间
	CompleteTime  int64  // 本局完成时间
}

// 游戏每回合的玩家数据
type GameUserRound struct {
	Id                   int64 `orm:"pk"`
	RoomId               int64
	Round                int // 回合数
	UserId               int
	TingStatus           int   // 听牌状态(0:未叫牌;1:叫牌;2:报听)
	WinStatus            int   // 胡牌状态(0:无作为;1:天胡;2:地胡;3:杠后胡;4:自模胡;5:抢杠胡;6:热炮胡;7:点炮胡)
	WinType              int   // 胡牌牌型
	PaoStatus            int   // 点炮类型(0:未点炮;1:点炮;2:热炮;3:被抢杠)
	ChikenChargeBam1     int   // 是否有冲锋鸡
	ChikenResponsibility int   // 是否责任鸡
	ChikenChargeDot8     int   // 是否有冲锋乌骨
	ChikenBao            int   // 是否包鸡
	ChikenCnt            int   // 用于结算的鸡的个数
	ChikenBam1Cnt        int   // 用于结算的幺鸡个数
	ChikenDot8Cnt        int   // 用于计算的乌骨鸡个数
	KongCnt              int   // 明杠次数
	KongTurnCnt          int   // 转弯杠次数
	KongDarkCnt          int   // 暗杠次数
	PlayCnt              int   // 出牌张数
	DrawCnt              int   // 抓牌张数
	ReplyTimeCnt         int   // 操作时间汇总
	Score                int   // 本局积分
	StartTime            int64 // 本局开始时间
	CompleteTime         int64 // 本局完成时间
}

// 用户游戏次数记录表
type UserGameRoundTimes struct {
	UserId            int    `orm:"pk"` // 用户id
	CreateTimes       int    // 自主创建房间次数
	RandomTimes       int    // 随机创建房间次数
	CreateTimesDetail string // 自助创建次数详情
	RandomTimesDetail string // 随机创建次数详情
	CompleteTimes     int    // 正常完成游戏次数
	CreateUpdateTime  int64  // 创建类型的最后更新时间
	RandomUpdateTime  int64  // 随机类型的最后更新时间
	Reward            int    // 已发放奖励记录
}

// 比赛积分日志
type GameMatchesScore struct {
	Id         int   `orm:"pk"`
	UserId     int   // 用户id
	CurDay     int   // 年月日
	Score      int   // 积
	UpdateTime int64 // 更新时间
}

// 观察员观察的房间
type ObRooms struct {
	RoomId int64 `orm:"pk"` // 房间id
}

// 观察员表
type UserOther struct {
	UserId    int `orm:"pk"` // 用户id
	Status    int // 状态0=正常 1=非正常
	OtherType int // 1=电视端
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 俱乐部
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// Club 俱乐部表
// 服务端暂时用不到太多字段，只选择有需要的才放到结构中去
type Club struct {
	Id              int `orm:"pk"` // 俱乐部id
	ManageUser      int // 俱乐部创建者id
	Fund            int // 俱乐部基金
	EnableOut       int // 淘汰赛是否开启
	Score           int // 淘汰积分值
	AllowCreateroom int // 是否允许会员创建房间
}

// ClubRoom 俱乐部房间
type ClubRoom struct {
	RoomId     int64 `orm:"pk"` // 房间id
	ClubId     int   // 俱乐部id
	CreateTime int64 // 牌局创建时间
}

// ClubUser 俱乐部成员
type ClubUser struct {
	ClubId     int   `orm:"pk"` // 俱乐部id
	UserId     int   // 成员id
	CreateTime int64 // 加入时间
	Type       int   // 关联类型；1、普通会员；2、会长
	Score      int   // 用户麻将馆积分
}

// 俱乐部消费日志
type ClubConsumeLog struct {
	Id         int64  `orm:"pk"`
	RoomId     int64  // 房间id
	ClubId     int    // 俱乐部id
	Diamonds   int    // 消耗钻石
	LogType    string // 消耗类型
	CreateTime int64  // 房间创建时间
}

// 金币场消费日志
type CoinConsumeLog struct {
	Id         int64 `orm:"pk"`
	MatchType  int
	RoomId     int64 // 房间id
	UserId     int   // 用户id
	Coin       int   // 消耗的金币
	CreateTime int64 // 房间创建时间
}
