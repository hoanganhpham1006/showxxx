package main

import (
	//	"encoding/json"
	"errors"
	//	"fmt"
	"time"

	"github.com/daominah/livestream/conversations"
	//	"github.com/daominah/livestream/games/singleplayer"
	"github.com/daominah/livestream/games/singleplayer/egg"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/nbackend"
	//	"github.com/daominah/livestream/nwebsocket"
	"github.com/daominah/livestream/rank"
	"github.com/daominah/livestream/streams"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zglobal"
)

const (
	LOGIN_BY_PASSWORD = "LOGIN_BY_PASSWORD"
	LOGIN_BY_COOKIE   = "LOGIN_BY_COOKIE"
)

func HandleClientDisconnect(userId int64, loginId int64) {
	nbackend.GBackend.HandleClientDisconnect(userId)
	user, _ := users.GetUser(userId)
	if user != nil {
		user.StatusL1 = users.STATUS_OFFLINE
	}
	if loginId != 0 {
		users.RecordLogout(loginId)
	}
}

func UserCreate(username string, password string) (
	map[string]interface{}, error) {
	userId, e := users.CreateUser(username, password)
	res := map[string]interface{}{"CreatedUserId": userId}
	return res, e
}

func UserLogin(loginType string,
	proxyId int64, clientConnId int64, clientIp string,
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
	loginId, _ := users.RecordLogin(userObj.Id, clientIp, deviceName, appName)
	res := map[string]interface{}{
		"User":         userObj.ToMap(),
		"UserId":       userObj.Id,
		"LoginId":      loginId,
		"LoginSession": loginSession,
	}
	//
	nbackend.WriteMapToUserId(userObj.Id, nil, map[string]interface{}{
		"Command": "DisconnectFromServer",
		"UserId":  userObj.Id,
	})
	nbackend.GBackend.HandleClientLogIn(proxyId, userObj.Id)
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
func UploadFile(file []byte) (
	map[string]interface{}, error) {
	imgPath, e := users.UploadFile(file)
	if e != nil {
		return nil, errors.New(l.Get(l.M023StaticServerDown))
	}
	return map[string]interface{}{"FilePath": imgPath}, nil
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

func ConversationGifts() (map[string]interface{}, error) {
	gifts, e := conversations.LoadGiftsList()
	if e != nil {
		return nil, e
	}
	return map[string]interface{}{"Gifts": gifts}, nil
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

func StreamCreate(userId int64, streamName string, streamImage string) (
	map[string]interface{}, error) {
	stream, err := streams.CreateStream(userId, streamName, streamImage)
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
func StreamAllSummaries() (
	map[string]interface{}, error) {
	rows := streams.StreamAllSummaries(false)
	return map[string]interface{}{"Streams": rows}, nil
}
func StreamGetMyViewing(viewerId int64) (
	map[string]interface{}, error) {
	_, stream := streams.GetViewingStream(viewerId)
	if stream == nil {
		return nil, nil
	}
	return stream.ToMap(), nil
}
func StreamReport(viewerId int64, broadcasterId int64, reason string) (
	map[string]interface{}, error) {
	err := streams.ReportStream(viewerId, broadcasterId, reason)
	return nil, err
}

func SGameChooseMoneyType(gameCode string, userId int64, moneyType string) (
	map[string]interface{}, error) {
	game := MapSGames[gameCode]
	if game == nil {
		return nil, errors.New(l.Get(l.M035GameInvalidGameCode))
	}
	game.ChooseMoneyType(userId, moneyType)
	return nil, nil
}

func SGameChooseBaseMoney(gameCode string, userId int64, baseMoney float64) (
	map[string]interface{}, error) {
	game := MapSGames[gameCode]
	if game == nil {
		return nil, errors.New(l.Get(l.M035GameInvalidGameCode))
	}
	game.ChooseBaseMoney(userId, baseMoney)
	return nil, nil
}
func SGameGetPlayingMatch(gameCode string, userId int64) (
	map[string]interface{}, error) {
	game := MapSGames[gameCode]
	if game == nil {
		return nil, errors.New(l.Get(l.M035GameInvalidGameCode))
	}
	match := game.GetPlayingMatch(userId)
	if match == nil {
		return nil, errors.New(l.Get(l.M037GameInvalidMatchId))
	}
	return match.ToMap(), nil
}

func SGameEggSendMove(
	userId int64, data map[string]interface{}, args ...interface{}) (
	map[string]interface{}, error) {
	game := MapSGames[egg.GAME_CODE_EGG]
	if game == nil {
		return nil, errors.New(l.Get(l.M035GameInvalidGameCode))
	}
	match := game.GetPlayingMatch(userId)
	if match == nil {
		return nil, errors.New(l.Get(l.M037GameInvalidMatchId))
	}
	err := match.SendMove(data)
	return nil, err
}
func SGameEggCreateMatch(userId int64) (
	map[string]interface{}, error) {
	game := MapSGames[egg.GAME_CODE_EGG]
	if game == nil {
		return nil, errors.New(l.Get(l.M035GameInvalidGameCode))
	}
	match := &egg.EggMatch{}
	err := game.InitMatch(userId, match)
	return nil, err
}
