package game

import (
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	userModel "mahjong.go/model/user"
)

// EntityCheckEnough 检查entityId对应的资源是否足够
func EntityCheckEnough(userId, entityId, require int) bool {
	_, module, _ := util.ParseEntityID(entityId)
	var had int
	switch module {
	case config.ENTITY_MODULE_DIAMOND_ALL:
		u := userModel.GetUser(userId)
		had = userModel.GetMoney(u)
	case config.ENTITY_MODULE_DIAMOND:
		u := userModel.GetUser(userId)
		had = u.Money
	case config.ENTITY_MODULE_DIAMOND_FREE:
		u := userModel.GetUser(userId)
		had = u.GiftMoney
	default:
		core.Logger.Error("未支持的entity, userId:%v, module:%v", userId, module)
	}

	return require >= had
}
