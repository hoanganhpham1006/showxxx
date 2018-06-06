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
	"github.com/daominah/livestream/users"
)

func UserCreate(username string, password string) (
	map[string]interface{}, error) {
	userId, e := users.CreateUser(username, password)
	res := map[string]interface{}{"UserId": userId}
	return res, e
}

func saveLoggedInConnection(connection *connections.Connection, userObj *users.User) {
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
}

func UserLoginByPassword(connection *connections.Connection, username string, password string) (
	map[string]interface{}, error) {
	userObj, loginSession, err := users.LoginByPassword(username, password)
	if userObj == nil {
		return nil, err
	}
	res := map[string]interface{}{
		"User":         userObj.ToMap(),
		"LoginSession": loginSession,
	}
	saveLoggedInConnection(connection, userObj)
	return res, err
}

func UserLoginByCookie(connection *connections.Connection, login_session string) (
	map[string]interface{}, error) {
	userObj, err := users.LoginByCookie(login_session)
	if userObj == nil {
		return nil, err
	}
	res := map[string]interface{}{"User": userObj.ToMap()}
	saveLoggedInConnection(connection, userObj)
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
func ConversationCreateMessage(
	conversationId int64, senderId int64, messageContent string) (
	map[string]interface{}, error) {
	err := conversations.CreateMessage(conversationId, senderId, messageContent)
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
