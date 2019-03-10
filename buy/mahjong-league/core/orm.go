package core

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

const (
	MAX_IDLE = 10  // 最大空闲连接
	MAC_CONN = 100 // 最大连接数
)

// 定义orm结构
func init() {
	// 是否开启sql日志
	orm.Debug = false
}

// LoadOrmConfig 加载orm配置
func LoadOrmConfig() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// 注册写库
	orm.RegisterDataBase("default", "mysql", DBWriter, MAX_IDLE, MAC_CONN)
}

// GetWriter 获取写库实例
func GetWriter() orm.Ormer {
	return orm.NewOrm()
}
