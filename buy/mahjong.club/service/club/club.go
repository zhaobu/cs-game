package club

import (
	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"
	"mahjong.club/club"
	"mahjong.club/config"
	"mahjong.club/core"
	"mahjong.club/hall"
	"mahjong.club/ierror"
	"mahjong.club/message"
	"mahjong.club/sensitive"
	userService "mahjong.club/service/user"
	"mahjong.club/user"

	fbs "mahjong.club/fbs/Common"
)

// Join 加入俱乐部
func Join(userID int, impacket *protocal.ImPacket) *ierror.Error {
	// 解析参数
	request := fbs.GetRootAsClubJoinRequest(impacket.GetBody(), 0)
	clubID := int(request.ClubId())
	if clubID == 0 {
		return ierror.NewError(-10101, "club.Join", "clubID")
	}
	u, online := hall.UserSet.Get(userID)
	if !online {
		return ierror.NewError(-10200, userID)
	}
	// 读取俱乐部信息
	c, isExists := hall.ClubSet.Get(clubID)
	if !isExists {
		c = club.NewClub(clubID)
		hall.ClubSet.Add(c)
	}

	// 将用户添加到俱乐部
	u.Mux.Lock()
	defer u.Mux.Unlock()
	if !util.IntInSlice(clubID, u.Clubs) {
		u.Clubs = append(u.Clubs, clubID)
	}
	core.Logger.Debug("[club.Join]用户所处的俱乐部列表,userID:%v, clubs:%#v", userID, u.Clubs)
	// 添加俱乐部中的用户
	c.AddUser(u)

	// 返回一个加入俱乐部成功的消息
	u.AppendMessage(JoinResponse(impacket.GetMessageNumber(), clubID, nil))

	// 推送俱乐部的房间列表给用户
	go pushRoomList(u, c)
	go pushMessageList(u, c, 0, config.CLUB_MESSAGE_LIST_LIMIT)

	core.Logger.Info("[club.Join]clubID:%v,userID:%v", clubID, userID)

	return nil
}

// Quit 退出俱乐部
func Quit(userID int, impacket *protocal.ImPacket) *ierror.Error {
	// 解析参数
	request := fbs.GetRootAsClubJoinRequest(impacket.GetBody(), 0)
	clubID := int(request.ClubId())
	if clubID == 0 {
		return ierror.NewError(-10101, "club.Quit", "clubID")
	}
	u, online := hall.UserSet.Get(userID)
	if !online {
		return ierror.NewError(-10200, userID)
	}
	// 删除用户属性中的俱乐部房间id
	if util.IntInSlice(clubID, u.Clubs) {
		u.Clubs = util.SliceDel(u.Clubs, clubID)
	}
	core.Logger.Debug("[club.Quit]用户所处的俱乐部列表,userID:%v, clubs:%#v", userID, u.Clubs)

	// 将用户从俱乐部删除
	hall.RemoveClubUser(clubID, u)

	// 发送成功退出俱乐部的通知
	u.AppendMessage(QuitResponse(impacket.GetMessageNumber(), clubID, nil))

	core.Logger.Info("[club.Quit]clubID:%v,userID:%v", clubID, userID)

	return nil
}

/*
// Restore 重载俱乐部房间列表
func Restore(userID int, impacket *protocal.ImPacket) *ierror.Error {
	// 解析参数
	request := fbs.GetRootAsClubJoinRequest(impacket.GetBody(), 0)
	clubID := int(request.ClubId())
	if clubID == 0 {
		return ierror.NewError(-10101, "club.Restore", "clubID")
	}
	u, online := hall.UserSet.Get(userID)
	if !online {
		return ierror.NewError(-10200, userID)
	}
	// 判断用户是否在俱乐部中
	if !util.IntInSlice(clubID, u.Clubs) {
		return ierror.NewError(-10201, userID, clubID)
	}
	// 读取俱乐部信息
	c, isExists := hall.ClubSet.Get(clubID)
	if !isExists {
		return ierror.NewError(-10300, clubID)
	}

	// 推送房间消息
	u.MStatus = false
	go pushRoomList(u, c)

	core.Logger.Info("[club.Restore]clubID:%v, userID:%v", clubID, userID)

	return nil
}

// RestoreDone 重载俱乐部房间完成
func RestoreDone(userID int, impacket *protocal.ImPacket) *ierror.Error {
	u, online := hall.UserSet.Get(userID)
	if online {
		u.MStatus = true
	}
	core.Logger.Info("[club.RestoreDone]userID:%v", userID)
	return nil
}
*/

// 推送俱乐部房间列表给用户
func pushRoomList(u *user.User, c *club.Club) {
	defer util.RecoverPanic()
	u.AppendMessage(ClubRestorePush(c))
}

// SendMessage 发送俱乐部消息
func SendMessage(userID int, impacket *protocal.ImPacket) *ierror.Error {
	request := fbs.GetRootAsClubClubMessageNotify(impacket.GetBody(), 0)
	clubID := int(request.ClubId())
	mType := int(request.MType())
	content := string(request.Content())
	if clubID == 0 {
		return ierror.NewError(-10101, "club.SendMessage", "clubID")
	}
	_, online := hall.UserSet.Get(userID)
	if !online {
		return ierror.NewError(-10200, userID)
	}
	// 读取俱乐部信息
	c, isExists := hall.ClubSet.Get(clubID)
	if !isExists {
		return ierror.NewError(-10300, clubID)
	}
	if len(content) == 0 {
		return ierror.NewError(-10101, "club.SendMessage", "content")
	}
	if mType == fbs.ClubMessageTypeTEXT {
		content = sensitive.Replace(content)
	}
	// 新建一条消息
	sender := message.NewSender(userID)
	sender.Info.AvatarBox = userService.GetUserAvatarBox(userID)
	msg := message.NewMsg()
	msg.MID = c.NextMessageID()
	msg.MType = mType
	msg.Content = content
	msg.Sender = sender
	// 添加消息到历史消息
	c.ML.Add(msg)
	// 推送消息到俱乐部
	hall.SendClubMessage(c, ClubMessagePush(clubID, msg))
	core.Logger.Info("[club.SendMessage]clubId:%v,userId:%v,mType:%v,mId:%v,content:%v", clubID, userID, mType, msg.MID, content)
	return nil
}

// MessageList 获取俱乐部历史消息
func MessageList(userID int, impacket *protocal.ImPacket) *ierror.Error {
	request := fbs.GetRootAsClubClubMessageListNotify(impacket.GetBody(), 0)
	clubID := int(request.ClubId())
	lastMsgID := request.MsgId()
	limit := int(request.Limit())

	if clubID == 0 {
		return ierror.NewError(-10101, "club.MessageList", "clubID")
	}
	u, online := hall.UserSet.Get(userID)
	if !online {
		return ierror.NewError(-10200, userID)
	}
	// 读取俱乐部信息
	c, isExists := hall.ClubSet.Get(clubID)
	if !isExists {
		return ierror.NewError(-10300, clubID)
	}
	if limit == 0 {
		limit = config.CLUB_MESSAGE_LIST_LIMIT
	}
	go pushMessageList(u, c, lastMsgID, limit)
	return nil
}

// 推送俱乐部历史消息给用户
func pushMessageList(u *user.User, c *club.Club, lastMsgID uint64, limit int) {
	msgList := c.ML.GetList(lastMsgID, limit)
	if len(msgList) > 0 {
		u.AppendMessage(ClubMessageListPush(c.ID, msgList))
		core.Logger.Debug("[pushMessageList]clubId:%v,userId:%v,lastMsgId:%v,limit:%v,len:%v", c.ID, u.ID, lastMsgID, limit, len(msgList))
	} else {
		core.Logger.Debug("[pushMessageList]没有更多消息了,clubId:%v,userId:%v,lastMsgId:%v,limit:%v", c.ID, u.ID, lastMsgID, limit)
	}
}
