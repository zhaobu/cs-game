package config

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* DUMMY
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	TOKEN_SECRET_KEY = "Tq-TqRIqf8fuck" // token密钥
	SYSTEM_KEY       = "fkmaxrmxxoo"          // 系统消息密钥
)

// 距离多少米之内被认为是作弊(单位:米)
const CHEAT_DISTANCE_LIMIT = 50

// 心跳间隔：秒
const HEART_BEAT_SECOND = 3

// 游戏版本默认值(最新版)
const GAME_VERSION_DEFAULT = "latest"

// 重连日志有效期
const RESTORE_LOG_EXPIRE_SECOND = 172800 // 两天
// 重连类型
const RESTORE_LOG_TYPE_RECONNECT = 0 // 重连
const RESTORE_LOG_TYPE_KICK = 1      // 下线

// 金币上限
const COIN_UPPER_LIMIT = 99999999

// 钻石类型
const (
	DIAMOND_TYPE_MONEY      = 1
	DIAMOND_TYPE_GIFT_MONEY = 2
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	ROOM_NUMBER_LENGTH = 6 // 房间号长度
	// 房间状态
	ROOM_STATUS_CREATING   = 0  // 组建中
	ROOM_STATUS_PALYING    = 10 // 游戏中
	ROOM_STATUS_SETTLEMENT = 20 // 结算中
	ROOM_STATUS_COMPLETED  = 30 // 已结束
	// 房间类型
	ROOM_TYPE_CREATE     = 0 // 自主创建
	ROOM_TYPE_RAND       = 1 // 随机组建
	ROOM_TYPE_MATCH      = 2 // 比赛房间
	ROOM_TYPE_TV         = 3 // 电视端房间
	ROOM_TYPE_CLUB       = 4 // 俱乐部房间
	ROOM_TYPE_CLUB_MATCH = 5 // 俱乐部淘汰赛房间
	ROOM_TYPE_COIN       = 6 // 金币场
	ROOM_TYPE_LEAGUE     = 7 // 联赛场
	ROOM_TYPE_RANK       = 8 // 排位赛
	// 房间模式
	ROOM_CREATE_MODE_USER   = 0 // 用户自主创建
	ROOM_CREATE_MODE_CLUB   = 1 // 馆主创建
	ROOM_CREATE_MODE_SYSTEM = 2 // 系统自动创建

	// 随机房间默认局数
	ROOM_RANDOM_ROUND = 4
	ROOM_MATCH_ROUND  = 4
	ROOM_COIN_ROUND   = 1
	// 房间自动解散间隔
	ROOM_DISMISS_AOTO_ALLOW_INTERVAL = 60 // 倒计时60秒后，房间自动解散
	// 房间解散回应
	ROOM_DISMISS_APPLY = -1 // 申请解散房间
	ROOM_DISMISS_ALLOW = 0  // 同意解散房间
	ROOM_DISMISS_DENY  = 1  // 拒绝解散房间
	// 房间解散标志
	ROOM_DISMISS_FALG_YES = 0 // 解散房间
	ROOM_DISMISS_FALG_NO  = 1 // 未解散房间
	// 自主创建房间过期时间设置
	ROOM_TIMEOUT_SECOND = int64(7200) // 两小时
	// 随机创建房间过期时间设置
	ROOM_RANDOM_TIMEOUT_SECOND = int64(300) // 秒
	// 房间相关的cache key 的有效期，防止无限期存在于内存中
	ROOM_EXPIRE_SECOND = 7200 // 两小时
	// 房间满员到开始这段时间，准备的超时时间
	ROOM_FIRST_READY_TIMEOUT_SECOND = 10
	// 房间自动准备时间间隔
	ROOM_AUTO_HOSTING_INTERVAL_SECOND = int64(15)
	// 房间付费类型
	ROOM_PAY_TYPE_USER = 0 // 用户付费
	ROOM_PAY_TYPE_CLUB = 1 // 俱乐部付费
	// 金币场不换桌时，多少秒后允许其他用户加入
	ROOM_COIN_ALLOW_OTHER_USER_INTERVAL = int64(10)
)

// 准备标志
const (
	ROOM_READY_NO  = 0 // 准备中
	ROOM_READY_YES = 1 // 已准备
)

// 房间解散原因
const (
	DISMISS_ROOM_CODE_OFFSET           = 10100 // code与langId的偏移量，langId = code + DISMISS_ROOM_CODE_OFFSET
	DISMISS_ROOM_CODE_FINISH           = 1     // 牌局结束
	DISMISS_ROOM_CODE_MONEY_NOT_ENOUGH = 2     // 房费不足
	DISMISS_ROOM_CODE_APPLY            = 3     // 用户申请解散
	DISMISS_ROOM_CODE_TIMEOUT          = 4     // 超时解散
	DISMISS_ROOM_CODE_HOST_LEAVE       = 5     // 房主退出
	DISMISS_ROOM_CODE_OB_QUIT          = 7     // 观察员退出解散
	DISMISS_ROOM_CODE_CLUB_DISMISS     = 8     // 俱乐部馆主解散
)

// 用户退出原因
const (
	QUIT_ROOM_CODE_INITIATIVE = 1 // 主动退出
	QUIT_ROOM_CODE_KICK       = 2 // 被退出
	QUIT_ROOM_CODE_TIMEOUT    = 4 // 超时退出
)

// 用户托管状态
const (
	ROOM_USER_HOSTING_YES = 1 // 已托管
	ROOM_USER_HOSTING_NO  = 0 // 未托管
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 麻将定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 初始麻将张数定义
const (
	MAHJONG_INIT_TILE_CNT_DEALER    = 14 // 庄家
	MAHJONG_INIT_TILE_CNT_NONDEALER = 13 // 非庄家
)

// 麻将牌张数
const (
	MAHJONG_TILE_CNT_72  = 72  // 72 张
	MAHJONG_TILE_CNT_108 = 108 // 108张
	MAHJONG_TILE_CNT_112 = 112 // 112张
)

// 服务端牌局进程定义
const (
	MAHJONG_PROGRESS_CREATE_TEAM = 0 // 组队中
	MAHJONG_PROGRESS_INIT        = 1 // 初始化
	MAHJONG_PROGRESS_FLOWER      = 2 // 换牌
	MAHJONG_PROGRESS_EXCHANGE    = 3 // 换牌
	MAHJONG_PROGRESS_LACK        = 4 // 定缺
	MAHJONG_PROGRESS_INIT_TING   = 5 // 定缺
	MAHJONG_PROGRESS_PLAY        = 6 // 游戏
	MAHJONG_PROGRESS_SETTLE      = 7 // 结算
)

// 计算可进行操作时的牌类型
const (
	MAHJONG_OPERATION_CALC_PLAY        = 0 // 别人出了牌
	MAHJONG_OPERATION_CALC_DROW        = 1 // 自己抓了牌
	MAHJONG_OPERATION_CALC_KONG_TURN   = 2 // 代表转弯杠
	MAHJONG_OPERATION_CALC_BEFORE_DRAW = 3 // 抓牌前
)

// 返回胡的方式，1:天胡;2:地胡;3:杠后胡;4:自模胡;5:抢杠胡;6:热炮胡;7:点炮胡
const (
	HU_WAY_TIAN       = 1
	HU_WAY_DI         = 2
	HU_WAY_KONG_DRAW  = 3
	HU_WAY_DRAW       = 4
	HU_WAY_QIANG_KONG = 5
	HU_WAY_RE_PAO     = 6
	HU_WAY_PAO        = 7
)

// 胡牌类型
const (
	HU_TYPE_SHUANG_LONG_7DUI = 1
	HU_TYPE_LONG_7DUI        = 2
	HU_TYPE_7DUI             = 3
	HU_TYPE_DIQIDUI          = 4
	HU_TYPE_DANDIAO          = 5
	HU_TYPE_DADUI            = 6
	HU_TYPE_BIANKADIAO       = 7
	HU_TYPE_DAKUANZHANG      = 8
	HU_TYPE_PI               = 9
	HU_TYPE_KONG_DRAW        = 10
	HU_TYPE_HEPU_7DUI        = 11
)

// 杀报时，胡牌类型清一色位移值
const SHABAO_HU_TYPE_OFFSET = 10000

// 杀报时，胡牌类型清一色位移值
const SHABAO_HU_TYPE_QING_OFFSET = 10000

// 听牌
const (
	USER_STATUS_BAD     = -1 // 未初始化，无法知道ting的信息，是否能胡牌要根据算法来算
	USER_STATUS_NORMAL  = 0  // 无听
	USER_STATUS_TING    = 1  // 听
	USER_STATUS_BAOTING = 2  // 报听
)

// 听的占位值，在发给用户ting的operation时，如果slice[0]等于此值，表示用户处于听牌状态，听的牌为slice[1:]
const OPERATION_CODE_TING_PLACEHOLDER = 127

// 用户输赢鸡的状态
const (
	USER_SETTLEMENT_CHIKEN_STATUS_NO   = 0 // 不输不赢
	USER_SETTLEMENT_CHIKEN_STATUS_WIN  = 1 // 赢
	USER_SETTLEMENT_CHIKEN_STATUS_LOSE = 2 // 包
)

// 鸡的类型
const (
	CHIKEN_TYPE_BAM1       = 1    // 幺鸡(2的0次方)
	CHIKEN_TYPE_DOT8       = 2    // 乌骨鸡(2的1次方)
	CHIKEN_TYPE_DRAW       = 4    // 翻牌鸡(2的2次方)
	CHIKEN_TYPE_FB         = 8    // 前后鸡(2的3次方)
	CHIKEN_TYPE_SELF       = 16   // 本鸡鸡(2的4次方)
	CHIKEN_TYPE_WEEK       = 32   // 星期鸡(2的5次方)
	CHIKEN_TYPE_TUMBLING   = 64   // 滚筒鸡(2的6次方)
	CHIKEN_TYPE_SILVER     = 128  // 银鸡(2的7次方)
	CHIKEN_TYPE_DIAMOND    = 256  // 钻石鸡(2的8次方)
	CHIKEN_TYPE_PAPO       = 512  // 爬坡鸡(2的9次方)
	CHIKEN_TYPE_FLOWER_RED = 1024 // 补花鸡(2的10次方)
)

// 房间操作托管回应间隔
const HOSTING_OPERATION_WAIT_TIME_1 = 1
const HOSTING_OPERATION_WAIT_TIME_3 = 3
const HOSTING_OPERATION_WAIT_TIME_5 = 6
const HOSTING_OPERATION_WAIT_TIME_10 = 10

// 回放数据有效期
const GAME_PLAYBACK_DATA_EXPIRE = 172800       // 缩减版:48小时
const GAME_PLAYBACK_DATA_INTACT_EXPIRE = 43200 // 完整版:12小时

// 回放oss文件名
const OSS_PLAYBACK_VERSION_FILE_NAME = "version-%v-%v"
const OSS_PLAYBACK_SIMPLE_FILE_NAME = "data-%v-%v"
const OSS_PLAYBACK_INTACT_FILE_NAME = "intact-%v-%v"

// 换牌方向
const EXCHANGE_DIRECTION_OPPOSITE = 0         // 对面
const EXCHANGE_DIRECTION_CLOCKWISE = 1        // 顺时针
const EXCHANGE_DIRECTION_COUNTERCLOCKWISE = 2 // 逆时针

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 缓存键值的相关定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 用户
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_USER_ROOM_ID              = "USER:ROOMID:%d"
	CACHE_KEY_USER_RESOTRE_LIST         = "USER:RESTORE:LIST:%v:%v"
	CACHE_KEY_RANDOM_PAYED              = "USER:RANDOM:PAYED:%v:%v"
	CACHE_KEY_H5_SHARE                  = "H5:SHARE:PLAY:%v"
	CACHE_KEY_USER_LAST_PLAY            = "USER:LAST:PLAY"
	CACHE_KEY_USER_UNREAD_GAME_RESULT   = "USER:UNREAD:GAME:RESULT:%v"
	CACHE_KEY_LAST_WIN_STATUS           = "USER:LAST:WIN:STATUS:%v"
	CACHE_KEY_USER_AVATAR_BOX           = "USER:AVATAR:%v"          // 用户头像框
	CACHE_KEY_USER_MEMBER_LEVEL         = "USER:MEMBER:LEVEL:%v"    // 用户会员等级
	CACHE_KEY_USER_MEMBER_LEVEL_ADD_EXP = "SEASON:MEMBER:ADDED:EXP" // 用户会员等级带来的经验加成
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_ROOM_BUILDER       = "ROOM:BUILDER"
	CACHE_KEY_ROOM_NUMBER_ID     = "ROOM:NUMBER:ID:%s"
	CACHE_KEY_ROOM_ID_NUMBER     = "ROOM:ID:NUMBER:%d"
	CACHE_KEY_ROOM_REMOTE        = "ROOM:REMOTE:%d"
	CACHE_KEY_ROOM_PRICE         = "GAME:PRICE"
	CACHE_KEY_ROOM_COIN_CONFIG   = "COIN:MATCH:%d:%d"
	CACHE_KEY_ROOM_CLUB_GROWTH   = "PUSH:AGENT:GROWTH"
	CACHE_KEY_CLUB_PUSH_LIST     = "PUSH:CLUB:LIST"
	CACHE_KEY_ROOM_RESULT        = "ROOM:RESULT:%v" // 房间游戏结果
	CACHE_KEY_ROOM_RESULT_EXPIRE = 7200             // 房间游戏结果有效期
)

// 房间用户额外好牌率
const (
	CACHE_KEY_DRAW_EXPECT_EXTRA_RATE = "DRAW:EXPECT:EXTRA:RATE" // hash结构, key=用户类型:房间类型:子id, value=1~100
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 大厅
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_REMOTE_USER_CNT    = "HALL:REMOTE:USER:CNT"
	CACHE_KEY_REMOTE_ACTION_TIME = "HALL:REMOTE:ACTION:TIME"
	CACHE_KEY_HALL_ROOM_IDS      = "HALL:ROOM:IDS:%s"
	// CACHE_KEY_HALL_ROBOT_ROOM_LIST = "HALL:ROBOT:ROOM:LIST:%s"
	CACHE_KEY_HALL_ROBOT_ROOM_LIST = "HALL:ROBOT:ROOM:LIST" // 新的键名
	CACHE_KEY_REMOTE_VERSION       = "HALL:REMOTE:VERSION"
	CACHE_KEY_HALL_RESOTRE_LIST    = "HALL:RESTORE:LIST:%v:%v:%v"
	CACHE_KEY_HALL_RESOTRE_COUNT   = "HALL:RESTORE:COUNT"
	CACHE_KEY_COIN_USER_CNT        = "HALL:COIN:USER:CNT"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 回放
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_PLAYBACK_VERSION     = "PLAYBACK:VERSION:%d:%d"
	CACHE_KEY_PLAYBACK_DATA        = "PLAYBACK:DATA:%d:%d"
	CACHE_KEY_PLAYBACK_DATA_INTACT = "PLAYBACK:DATA:INTACT:%d:%d"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* push
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// push队列, lpush、rpop
const (
	CACHE_KEY_PUSH_QUEUE_LIST         = "PUSH:QUEUE:LIST"
	CACHE_KEY_PUSH_QUEUE_LIST_ANDROID = "PUSH:QUEUE:LIST:ANDROID"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* TV端
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_OB_IPS = "TERMINAL:OP:IPS"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 金币场
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_GAME_LOG           = "GAME:LOG:%v:%v:%v"
	CACHE_KEY_COIN_RANK_PROVINCE = "COIN:RANK:PROVINCE"  // 省排名
	CACHE_KEY_COIN_RANK_CITY     = "COIN:RANK:CITY:%v"   // 市排名, 参数:city
	CACHE_KEY_COIN_RANK_FRIEND   = "COIN:RANK:FRIEND:%v" // 好友排名, 参数:userId
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 联赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_LEAGUE_RACE_SCORES = "LEAGUE:RACE:SCORES:%v"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 排位赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_SEASON_INFO                       = "SEASON:INFO"
	CACHE_KEY_SEASON_CARD_CONSUME               = "SEASON:CARD:%v"
	CACHE_KEY_RANK_PROVINCE                     = "SEASON:RANK:PROVINCE:%v"             // 省排名，参数:season
	CACHE_KEY_RANK_CITY                         = "SEASON:RANK:CITY:%v:%v"              // 市排名, 参数:season、city
	CACHE_KEY_RANK_FRIEND                       = "SEASON:RANK:FRIEND:%v:%v"            // 好友排名, 参数:season、userId
	CACHE_KEY_SEASON_USER                       = "SEASON:USER:%v"                      // 用户本赛季段位等级，参数:%v 用户星星等级
	CACHE_KEY_SEASON_SIGN                       = "SEASON:SIGN"                         // 赛季段位报名人数
	CACHE_KEY_RANK_UPGRADE                      = "SEASON:RANK:UPGRADE:%v"              // 用户排位赛排名上升
	CACHE_KEY_RANK_PLAY_TIMES                   = "SEASON:RANK:PLAY:TIMES:%v:%v"        // 用户游戏次数, 按天
	CACHE_KEY_RANK_TOGETHER_TIMES               = "SEASON:RANK:TOGETHER:TIMES:%v:%v:%v" // 同时游戏次数，按天
	CACHE_KEY_RANK_WINNING_STREAK               = "SEASON:RANK:USER:WINNING:STREAK:%v"  // 用户连胜次数, 参数:seasonId
	CACHE_KEY_RANK_SEASON_GAME_WIN              = "SEASON:GAME:WIN"                     // 用户连胜奖励
	CACHE_KEY_RANK_SEASON_PROVINCE_RANK_REWARDS = "SEASON:TEN:REWARD"                   // 赛季省排名前10奖励
	CACHE_KEY_RANK_SEASON_GRADE_REWARDS         = "SEASON:INFO:REWARD"                  // 段位对应的奖励
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 好友
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_USER_FRIEND = "SEASON:FRIEND:%v"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 数据库表的相关定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 表名
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	TABLE_NAME_USER      = "user"
	TABLE_NAME_USER_INFO = "user_info"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 用户扩展信息类型定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	USER_INFO_TYPE_GENDER           = 2  // 性别
	USER_INFO_TYPE_DEVICE_CODE      = 5  // device_code
	USER_INFO_TYPE_CITY             = 6  // 城市
	USER_INFO_TYPE_SCORE            = 12 // 用户积分
	USER_INFO_TYPE_SCORE_RANDOM     = 20 // 随机组局累计积分
	USER_INFO_TYPE_SCORE_MATCH      = 21 // 比赛累计积分
	USER_INFO_TYPE_SCORE_COIN       = 37 // 金币场累计积分
	USER_INFO_TYPE_PUNISHMENT_FLAG  = 22 // 用户需受惩罚标志
	USER_INFO_TYPE_PUNISHMENT_TIMES = 23 // 用户被惩罚次数
	USER_INFO_TYPE_GAME_CITY        = 41 // 用户城市
	USER_INFO_TYPE_RANK_CARDS       = 42 // 用户排位赛参赛卡张数
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 消费类型
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	MONEY_CHANGE_TYPE_XF = "xf" // 消费钻石
	MONEY_CHANGE_TYPE_TF = "tf" // 退还房费
)
const (
	MONEY_CONSUME_TYPE_CREATE = 3 // 创建房间房费
	MONEY_CONSUME_TYPE_MATCH  = 4 // 比赛房间房费
	MONEY_CONSUME_TYPE_RANDOM = 5 // 随机房间房费
	MONEY_CONSUME_TYPE_CLUB   = 7 // 俱乐部房间房费
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 消息id
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 推送消息id配置
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	PUSH_CHAT_ID_NEXT_GAME          = 10000 // 每一局游戏开始
	PUSH_CHAT_ID_ROOM_DISMISS       = 10001 // 房间解散
	PUSH_CHAT_ID_ROOM_DISMISS_APPLY = 10002 // 申请解散
	PUSH_CHAT_ID_ROOM_JOIN          = 10003 // 用户加入房间
	PUSH_CHAT_ID_ROOM_QUIT          = 10004 // 用户退出房间
	PUSH_CHAT_ID_ROOM_FULL          = 10005 // 房间满了
	PUSH_CHAT_ID_ROOM_DISMISS_AGREE = 10006 // 同意解散房间
	PUSH_CHAT_ID_ROOM_DISMISS_DENY  = 10007 // 拒绝解散房间
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 通知消息id配置
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CHAT_ID_YY_JOIN              = int16(1000) // 加入实时语音聊天
	CHAT_ID_YY_QUIT              = int16(1001) // 退出实时语音聊天
	CHAT_ID_SIGNAL_VERY_STRONGER = int16(1002) // 信号强度：非常强
	CHAT_ID_SIGNAL_STRONGER      = int16(1003) // 信号强度：强
	CHAT_ID_SIGNAL_NORMAL        = int16(1004) // 信号强度：普通
	CHAT_ID_SIGNAL_WEAK          = int16(1005) // 信号强度：弱
	CHAT_ID_SIGNAL_VERY_WEAK     = int16(1006) // 信号强度：非常弱
	CHAT_ID_VOICE_ID             = int16(1100) // 语音通知
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 联赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 比赛房间状态
const (
	RACE_ROOM_STATUS_NORMAL  = 0 // 正常
	RACE_ROOM_STATUS_FINISH  = 1 // 正常结束
	RACE_ROOM_STATUS_DISMISS = 2 // 异常结束
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* entity
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	ENTITY_ID_CONSUME = 1 // 消耗
	ENTITY_ID_GET     = 2 // 获得
)

const (
	ENTITY_MODULE_DIAMOND_ALL  = 1 // 双钻
	ENTITY_MODULE_DIAMOND      = 2 // 金钻
	ENTITY_MODULE_DIAMOND_FREE = 3 // 银钻
	ENTITY_MODULE_ITEM         = 4 // 道具
)
