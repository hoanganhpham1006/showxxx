package main

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"time"

	//	m "github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/connections"
	"github.com/daominah/livestream/conversations"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/rank"
	"github.com/daominah/livestream/streams"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zglobal"
)

const (
	LOGIN_BY_PASSWORD = "LOGIN_BY_PASSWORD"
	LOGIN_BY_COOKIE   = "LOGIN_BY_COOKIE"
)

func doAfterClosingConnection(c *connections.Connection) {
	if c.UserId != 0 {
		connections.GMutex.Lock()
		delete(connections.MapConnection, c.UserId)
		connections.GMutex.Unlock()

		user, _ := users.GetUser(c.UserId)
		if user != nil {
			user.StatusL1 = users.STATUS_OFFLINE
		}
	}
	if c.LoginId != 0 {
		temp := c.LoginId
		c.LoginId = 0 // evade connection can be RecordLogout 2 times
		users.RecordLogout(temp)
	}
}

func UserCreate(username string, password string) (
	map[string]interface{}, error) {
	userId, e := users.CreateUser(username, password)
	res := map[string]interface{}{"UserId": userId}
	return res, e
}

func UserLogin(loginType string, connection *connections.Connection,
	username string, password string, login_session string,
	deviceName string, appName string) (
	map[string]interface{}, error) {
	var userObj *users.User
	var loginSession string
	var err error
	if loginType == LOGIN_BY_PASSWORD {
		userObj, loginSession, err = users.LoginByPassword(username, password)
	} else if loginType == LOGIN_BY_COOKIE {
		userObj, err = users.LoginByCookie(login_session)
	} else {
		return nil, errors.New(l.Get(l.M024InvalidLoginType))
	}
	if userObj == nil {
		return nil, err
	}
	userObj.StatusL1 = users.STATUS_ONLINE
	//
	loginId, _ := users.RecordLogin(userObj.Id,
		fmt.Sprintf("%v", connection.WsConn.RemoteAddr()), deviceName, appName)
	connection.LoginId = loginId
	res := map[string]interface{}{
		"User":         userObj.ToMap(),
		"LoginSession": loginSession,
	}
	//
	connections.GMutex.Lock()
	oldConn := connections.MapConnection[userObj.Id]
	connections.GMutex.Unlock()
	if oldConn != nil {
		oldConn.WriteMap(nil, map[string]interface{}{
			"Command": "Disconnected",
			"Text": fmt.Sprintf("%v. %v",
				l.Get(l.M008Disconnected), l.Get(l.M009LoggedInDiffDevice)),
		})
		oldConn.Close()
	}
	connection.UserId = userObj.Id
	connections.GMutex.Lock()
	connections.MapConnection[userObj.Id] = connection
	connections.GMutex.Unlock()
	return res, err
}

func UserDetail(userId int64) (
	map[string]interface{}, error) {
	userObj, err := users.GetUser(userId)
	if userObj == nil {
		return nil, err
	}
	res := map[string]interface{}{"User": userObj.ToMap()}
	return res, err
}

func UserFollowers(userId int64) (
	map[string]interface{}, error) {
	followerIds := users.LoadFollowers(userId)
	res := map[string]interface{}{"FollowerIds": followerIds}
	return res, nil
}

func UserFollowing(userId int64) (
	map[string]interface{}, error) {
	followingIds := users.LoadFollowing(userId)
	res := map[string]interface{}{"FollowingIds": followingIds}
	return res, nil
}

func UserFollow(userId int64, targetId int64) (
	map[string]interface{}, error) {
	err := users.Follow(userId, targetId)
	return nil, err
}
func UserUnfollow(userId int64, targetId int64) (
	map[string]interface{}, error) {
	err := users.Unfollow(userId, targetId)
	return nil, err
}
func UserViewMoneyLog(userId int64, fromTime time.Time, toTime time.Time) (
	map[string]interface{}, error) {
	rows, err := users.ViewMoneyLog(userId, fromTime, toTime)
	res := map[string]interface{}{"Rows": rows}
	return res, err
}
func UserSearch(key string) (
	map[string]interface{}, error) {
	rows, err := users.Search(key)
	return map[string]interface{}{"Rows": rows}, err
}
func UserChangeInfo(userId int64, RealName string, NationalId string, Sex string,
	Country string, Address string, ProfileName string, ProfileImage string,
	Summary string) (
	map[string]interface{}, error) {
	e := users.ChangeUserInfo(
		userId, RealName, NationalId, Sex, Country, Address, ProfileName,
		ProfileImage, Summary)
	return nil, e
}
func UserChangeProfileImage(userId int64, newProfileImage []byte) (
	map[string]interface{}, error) {
	profileImagePath, e := users.ChangeUserProfileImage(userId, newProfileImage)
	return map[string]interface{}{"ProfileImagePath": profileImagePath}, e
}
func UserCheckFollowing(userId int64, targetId int64) (
	map[string]interface{}, error) {
	r := users.CheckIsFollowing(userId, targetId)
	return map[string]interface{}{"IsFollowing": r}, nil
}
func RankGetLeaderBoard(rankId int64) (
	map[string]interface{}, error) {
	rows := rank.GetLeaderboard(rankId)
	userRows := make([]map[string]interface{}, 0)
	for _, row := range rows {
		user, _ := users.GetUser(row.UserId)
		if user == nil {
			continue
		}
		userRow := user.ToShortMap()
		userRow["RKey"] = row.RKey
		userRows = append(userRows, userRow)
	}
	return map[string]interface{}{"RankId": rankId, "Rows": userRows}, nil
}
func ConversationAllSummaries(userId int64, filter string, nConversation int) (
	map[string]interface{}, error) {
	convs, err := conversations.UserLoadAllConversations(userId, filter, nConversation)
	res := map[string]interface{}{"Conversations": convs}
	return res, err
}
func ConversationDetail(userId int64, conversationId int64) (
	map[string]interface{}, error) {
	conv, err := conversations.GetConversation(conversationId)
	if conv == nil {
		return nil, err
	}
	conv.Mutex.Lock()
	_, isIn := conv.Members[userId]
	conv.Mutex.Unlock()
	if !isIn {
		return nil, errors.New(l.Get(l.M003ConversationOutsider))
	}
	res := map[string]interface{}{"Conversation": conv.ToMap()}
	return res, nil
}
func ConversationCreate(senderId int64, recipientId int64) (
	map[string]interface{}, error) {
	conversationId, err := conversations.CreateConversation(
		[]int64{senderId, recipientId}, nil, conversations.CONVERSATION_PAIR)
	return map[string]interface{}{"ConversationId": conversationId}, err
}
func ConversationCreateMessage(
	conversationId int64, senderId int64, messageContent string) (
	map[string]interface{}, error) {
	err := conversations.CreateMessage(
		conversationId, senderId, messageContent, conversations.DISPLAY_TYPE_NORMAL)
	return nil, err
}

func ConversationCreateBigMessage(
	conversationId int64, senderId int64, messageContent string) (
	map[string]interface{}, error) {
	_, err := users.ChangeUserMoney(senderId, users.MT_CASH, zglobal.MessageBigCost,
		users.REASON_CHAT_BIG, true)
	if err != nil {
		return nil, err
	}
	err = conversations.CreateMessage(
		conversationId, senderId, messageContent, conversations.DISPLAY_TYPE_BIG)
	return nil, err
}
func ConversationAddMember(
	userId int64, conversationId int64, newMemberId int64, isModerator bool) (
	map[string]interface{}, error) {
	conv, e := conversations.GetConversation(conversationId)
	if conv == nil {
		return nil, e
	}
	conv.Mutex.Lock()
	uObj := conv.Members[userId]
	conv.Mutex.Unlock()
	if uObj == nil {
		return nil, errors.New(l.Get(l.M003ConversationOutsider))
	}
	conversations.AddMember(conversationId, newMemberId, isModerator)
	return nil, nil
}
func ConversationRemoveMember(
	userId int64, conversationId int64, memberId int64) (
	map[string]interface{}, error) {
	conv, e := conversations.GetConversation(conversationId)
	if conv == nil {
		return nil, e
	}
	conv.Mutex.Lock()
	uObj := conv.Members[userId]
	conv.Mutex.Unlock()
	if uObj == nil {
		return nil, errors.New(l.Get(l.M003ConversationOutsider))
	}
	if uObj.IsModerator == false {
		return nil, errors.New(l.Get(l.M011ConversationModPrivilege))
	}
	conversations.RemoveMember(conversationId, memberId)
	return nil, nil
}
func ConversationBlockMember(
	userId int64, conversationId int64, memberId int64) (
	map[string]interface{}, error) {
	return nil, nil
}
func ConversationMute(
	userId int64, conversationId int64) (
	map[string]interface{}, error) {
	return nil, nil
}
func ConversationMarkMessage(userId int64, messageId int64, hasSeen bool) (
	map[string]interface{}, error) {
	e := conversations.UserMarkMessage(userId, messageId, hasSeen)
	return nil, e
}

func Cheer(conversation_id int64, cheerer_id int64, target_user_id int64,
	cheer_type string, val float64, cheer_message string, misc string) (
	map[string]interface{}, error) {
	err := conversations.Cheer(conversation_id, cheerer_id, target_user_id,
		cheer_type, val, cheer_message, misc)
	return nil, err
}

func TeamCreate(
	createrId int64, teamName string, teamImage string, teamSummary string) (
	map[string]interface{}, error) {
	teamId, err := users.CreateTeam(teamName, teamImage, teamSummary)
	if err != nil {
		return nil, err
	}
	err = users.AddTeamMember(teamId, createrId)
	if err != nil {
		return nil, err
	}
	err = users.SetTeamCaptain(teamId, createrId)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"TeamId": teamId}, nil
}
func TeamDetail(teamId int64) (
	map[string]interface{}, error) {
	team, err := users.GetTeam(teamId)
	if err != nil {
		return nil, err
	}
	return team.ToMap(), nil
}
func TeamLoadJoiningRequests(teamId int64) (
	map[string]interface{}, error) {
	rows, err := users.LoadTeamJoiningRequests(teamId)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"JoiningRequest": rows}, nil
}
func TeamRemoveMember(commanderId int64, teamId int64, userId int64) (
	map[string]interface{}, error) {
	team, e := users.GetTeam(teamId)
	if team == nil {
		return nil, e
	}
	var captainId int64
	if team.Captain != nil {
		captainId = team.Captain.Id
	}
	if commanderId != captainId && commanderId != userId {
		return nil, errors.New(l.Get(l.M017TeamMemberPrivilege))
	}
	err := users.RemoveTeamMember(teamId, userId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func TeamRequestJoin(teamId int64, userId int64) (
	map[string]interface{}, error) {
	err := users.RequestJoinTeam(teamId, userId)
	return nil, err
}
func TeamHandleJoiningRequest(teamId int64, userId int64, isAccepted bool) (
	map[string]interface{}, error) {
	defer users.RemoveRequestJoinTeam(teamId, userId)
	var err error
	if isAccepted {
		err = users.AddTeamMember(teamId, userId)
		if err != nil {
			return nil, err
		}
	}
	return nil, err
}

func StreamCreate(userId int64) (map[string]interface{}, error) {
	stream, err := streams.CreateStream(userId)
	if stream == nil {
		return nil, err
	}
	return stream.ToMap(), nil
}

func StreamView(viewerId int64, broadcasterId int64) (
	map[string]interface{}, error) {
	stream, err := streams.ViewStream(viewerId, broadcasterId)
	if stream == nil {
		return nil, err
	}
	return stream.ToMap(), nil
}

func StreamForwardSignaling(
	userId int64, targetUserId int64, data map[string]interface{}) (
	map[string]interface{}, error) {
	if data != nil {
		data["Sender"] = userId
	}
	connections.WriteMapToUserId(targetUserId, nil, data)
	return nil, nil
}
