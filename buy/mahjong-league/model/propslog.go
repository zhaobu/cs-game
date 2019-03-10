package model

import (
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

func init() {
	orm.RegisterModel(new(PropsLog))
}

// PropsLog 退费日志
type PropsLog struct {
	Id         int64 `orm:"pk"`
	UserId     int   // 用户id
	EntityId   int   // 实体id
	Price      int   // 数量
	SceneId    int   // 场景id
	CreateTime int64 // 创建时间
}

// LogPropsLog 记录退费日志
func LogPropsLog(userId int, entityId, amount int) error {
	defer util.RecoverPanic()

	log := new(PropsLog)
	log.UserId = userId
	log.EntityId = entityId
	log.Price = amount
	log.SceneId = 4
	log.CreateTime = util.GetTime()
	if _, dberr := core.GetWriter().Insert(log); dberr != nil {
		core.Logger.Error("[LogPropsLog]入库失败,err:%v", dberr.Error())
		return dberr
	}
	return nil
}
