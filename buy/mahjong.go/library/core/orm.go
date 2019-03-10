package core

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"mahjong.go/config"
)

const (
	MAX_IDLE = 10  // 最大空闲连接
	MAC_CONN = 100 // 最大连接数
)

var (
	ormWriter orm.Ormer
	ormReader orm.Ormer
	ormLogger orm.Ormer
)

// 定义orm结构
func init() {
	// 是否开启sql日志
	orm.Debug = false
	// 用户信息
	orm.RegisterModel(new(config.User), new(config.UserInfo))
	// 消费日志
	orm.RegisterModel(new(config.UserAccountLog), new(config.UserConsumeInfo), new(config.UserTransInfo))
	// 房间日志
	orm.RegisterModel(new(config.GameInfo), new(config.GameUserRecords), new(config.GameRoundData), new(config.GameResult), new(config.GameUserRound))
	// 用户游戏次数
	orm.RegisterModel(new(config.UserGameRoundTimes))
	// 比赛日志
	orm.RegisterModel(new(config.GameMatchesScore))
	// 观察者观察的房间
	orm.RegisterModel(new(config.ObRooms), new(config.UserOther))
	// 俱乐部
	orm.RegisterModel(new(config.Club), new(config.ClubUser), new(config.ClubRoom), new(config.ClubConsumeLog))
	// 金币场
	orm.RegisterModel(new(config.CoinConsumeLog))
}

func LoadOrmConfig() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// 注册写库
	orm.RegisterDataBase("default", "mysql", DBWriter, MAX_IDLE, MAC_CONN)
	// orm.RegisterDataBase("DBWriter", "mysql", DBWriter)

	// 载入读库
	// orm.RegisterDataBase("DBReader", "mysql", DBReader, MAX_IDLE, MAC_CONN)

	// 载入日志库
	// orm.RegisterDataBase("DBLogger", "mysql", DBLogger, MAX_IDLE, MAC_CONN)

	// 自动创建表
	// orm.RunSyncdb("default", false, true)
}

func GetWriter() orm.Ormer {
	/*
		if ormWriter == nil {
			ormWriter = orm.NewOrm()
		}

		return ormWriter
	*/

	return orm.NewOrm()
}

/*
func GetReader() orm.Ormer {
	ormReader = orm.NewOrm()
	ormReader.Using("DBReader")
	return ormReader
}

func GetLogger() orm.Ormer {
	ormLogger = orm.NewOrm()
	ormLogger.Using("DBLogger")

	return ormLogger
}
*/
