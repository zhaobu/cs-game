package ob

import (
	"github.com/astaxie/beego/orm"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// RecordObRoom 记录观察员观察过的房间id
func RecordObRoom(roomId int64) *core.Error {
	record := config.ObRooms{RoomId: roomId}

	// 记录已存在，不再执行
	err := core.GetWriter().Read(&record)
	if err == orm.ErrNoRows {
		if _, dberr := core.GetWriter().Insert(&record); dberr != nil {
			core.Logger.Error("记录观察员观察的房间失败, roomId:%v", roomId)
			return core.NewError(-4, dberr.Error())
		}
	}
	return nil
}

// IsObservers 是否观察员
func IsObservers(userId int) bool {

	o := core.GetWriter()
	record := new(config.UserOther)
	qs := o.QueryTable(record)
	err := qs.Filter("user_id", userId).Filter("status", 0).Filter("other_type", 1).One(record)
	if err == orm.ErrNoRows || record.OtherType != 1 {
		return false
	}
	return true
}
