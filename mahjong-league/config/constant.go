package config

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* DUMMY
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	TOKEN_SECRET_KEY       = "Tq-TqRIqf8fuck" // token密钥
	SYSTEM_KEY             = "fkmaxrmxxoo"          // 系统消息密钥
	MODEL_REFRESH_INTERVAL = int64(10)           // 模板数据读取间隔
)

// 钻石类型
const (
	DIAMOND_TYPE_MONEY      = 1
	DIAMOND_TYPE_GIFT_MONEY = 2
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* TABLE
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	TABLE_LEAGUE_LIST         = "league_list"
	TABLE_LEAGUE_REWARDS      = "league_rewards"
	TABLE_LEAGUE_RACE         = "league_race"
	TABLE_LEAGUE_RACE_USER    = "league_race_user"
	TABLE_LEAGUE_RACE_RANK    = "league_race_rank"
	TABLE_LEAGUE_RACE_ROOM    = "league_room"
	TABLE_LEAGUE_USER_REWARDS = "league_user_rewards"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* USER
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 心跳间隔：秒
const HEART_BEAT_SECOND = 3

// 握手超时间隔
const HANDSHAKE_TIMEOUT = 30

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* LEAGUE
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 比赛开始前多久不允许退赛
const RACE_FOBBIDEN_GIVEUP_SECOND = int64(30)
const ROOM_TYPE_LEAGUE = 7 // 联赛场

const RACE_USER_SCORE_BASE = 1000 // 用户基础积分

// 比赛类型
const (
	LEAGUE_TYPE_MONEY  = 1 // 钻石赛
	LEAGUE_TYPE_BONUS  = 2 // 红包赛
	LEAGUE_TYPE_CALLS  = 3 // 话费赛
	LEAGUE_TYPE_REWARD = 4 // 大奖赛
)

// 比赛开始条件
const (
	LEAGUE_START_CONDITION_TEMP  = 1 // 非定时赛(满员即开)
	LEAGUE_START_CONDITION_FIXED = 2 // 定时赛
)

// 循环模式
const (
	LEAGUE_CYCLE_NONE  = 0 // 非循环赛
	LEAGUE_CYCLE_DAY   = 1 // 日循环赛
	LEAGUE_CYCLE_WEEK  = 2 // 周循环赛
	LEAGUE_CYCLE_MONTH = 3 // 月循环赛
)

// 比赛上架状态
const (
	LEAGUE_STATUS_CLOSE = 0 // 非上架
	LEAGUE_STATUS_OPEN  = 1 // 上架
)

// 比赛状态
const (
	RACE_STATUS_SIGNUP        = 0 // 报名中
	RACE_STATUS_PLAN          = 1 // 排赛中
	RACE_STATUS_PLAY          = 2 // 比赛中
	RACE_STATUS_SETTLEMENT    = 3 // 结算中
	RACE_STATUS_FINISH        = 4 // 已结束(正常逻辑完成后介绍)
	RACE_STATUS_DISMISS       = 5 // 被解散（人数不足解散）
	RACE_STATUS_DISMISS_FORCE = 6 // 后台解散
)

// 比赛用户状态
const (
	RACE_USER_STATUS_SIGNUP  = 0 // 已报名
	RACE_USER_STATUS_GIVEUP  = 1 // 已退赛
	RACE_USER_STATUS_FAIL    = 2 // 已淘汰
	RACE_USER_STATUS_DISMISS = 5 // 被解散（人数不足解散）
)

// 比赛用户房间状态
const (
	RACE_USER_GIVEUP_STATUS_FORBID = 0 // 不允许退赛
	RACE_USER_GIVEUP_STATUS_ALLOW  = 1 // 允许退赛
)

// 比赛房间状态
const (
	RACE_ROOM_STATUS_NORMAL  = 0 // 正常
	RACE_ROOM_STATUS_FINISH  = 1 // 正常结束
	RACE_ROOM_STATUS_DISMISS = 2 // 异常结束
)

// 房间结束code
const (
	DISMISS_ROOM_CODE_FINISH = 1 // 牌局结束
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 消费类型
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	MONEY_CHANGE_TYPE_XF = "xf"  // 消费钻石
	MONEY_CHANGE_TYPE_TF = "tf1" // 退还房费
)
const (
	MONEY_CONSUME_TYPE_LEAGUE = 10 // 联赛报名收费
)
const (
	MONEY_TRANS_TYPE_LEAGUE = 26 // 联赛报名退费
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* CACHE KEY
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 联赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	CACHE_KEY_LEAGUE_RACE_SCORES   = "LEAGUE:RACE:SCORES:%v"
	CACHE_KEY_HALL_ROBOT_ROOM_LIST = "HALL:ROBOT:ROOM:LIST" // 机器人房间列表
	CACHE_KEY_GIVEUP_MESSAGE_QUEUE = "LEAGUE:EXIT:QUEUE"
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
	ENTITY_MODULE_GOLD         = 5 // 金币
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 用户扩展信息类型定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
const (
	USER_INFO_TYPE_SCORE_COIN = 37 // 金币场累计积分
	USER_INFO_TYPE_ITEMS      = 44 // 用户道具,{itemId:count, ... }
)
