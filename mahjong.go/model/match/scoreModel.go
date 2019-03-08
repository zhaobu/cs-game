package match

import (
	"strconv"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// UpdateDailyScore 更新用户日积分
// 当日记录存在, 更新score值;记录不存在,插入一条记录
// 当score=0时，为了进入排行，插入操作必须要执行, update操作跳过
func UpdateDailyScore(userId, score int) *core.Error {
	// 当前年月日
	curDay, _ := strconv.Atoi(util.GetYMD())

	// 已存在， update ，不存在 insert
	o := core.GetWriter()
	record := new(config.GameMatchesScore)
	qs := o.QueryTable(record)
	err := qs.Filter("user_id", userId).Filter("cur_day", curDay).One(record)

	var dberr error
	if err == orm.ErrNoRows {
		newRecord := new(config.GameMatchesScore)
		newRecord.UserId = userId
		newRecord.CurDay = curDay
		newRecord.Score = score
		newRecord.UpdateTime = util.GetTime()
		_, dberr = o.Insert(newRecord)
	} else {
		// 如果是0分，则不更新
		if score == 0 {
			return nil
		}

		record.Score += score
		record.UpdateTime = util.GetTime()
		_, dberr = o.Update(record, "score", "update_time")
	}

	if dberr != nil {
		return core.NewError(-4, dberr.Error())
	}
	return nil
}
