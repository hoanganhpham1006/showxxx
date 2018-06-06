package main

import (
	"encoding/json"
	"errors"

	"github.com/daominah/livestream/connections"
	l "github.com/daominah/livestream/language"
	m "github.com/daominah/livestream/misc"
)

func serverCommandHandler(connection *connections.Connection, message []byte) {
	var data map[string]interface{}
	parseTextErr := json.Unmarshal(message, &data)
	if parseTextErr != nil {
		connection.WriteMap(errors.New(
			"handleClientMessage json.Unmarshal err: "+parseTextErr.Error()), nil)
		return
	}

	command := m.ReadString(data, "Command")
	// responseData
	var d map[string]interface{}
	// responseError
	var e error

	if connection.UserId == 0 { // not logged in
		switch command {
		case "UserCreate":
			d, e = UserCreate(
				m.ReadString(data, "Username"),
				m.ReadString(data, "Password"))
		case "UserLoginByPassword":
			d, e = UserLoginByPassword(
				connection,
				m.ReadString(data, "Username"),
				m.ReadString(data, "Password"))
		case "UserLoginByCookie":
			d, e = UserLoginByCookie(
				connection,
				m.ReadString(data, "LoginSession"))
		default:
			d = map[string]interface{}{"message": string(message)}
			e = errors.New(l.Get(l.M010CommandNotSupported))
		}
	} else { // logged in
		switch command {
		case "UserDetail":
			d, e = UserDetail(
				m.ReadInt64(data, "UserId"))
		case "UserFollowers":
			d, e = UserFollowers(
				m.ReadInt64(data, "UserId"))
		case "UserFollowing":
			d, e = UserFollowing(
				m.ReadInt64(data, "UserId"))
		case "UserFollow":
			d, e = UserFollow(
				connection.UserId,
				m.ReadInt64(data, "TargetId"))
		case "UserUnfollow":
			d, e = UserUnfollow(
				connection.UserId,
				m.ReadInt64(data, "TargetId"))
		case "UserViewMoneyLog":
			d, e = UserViewMoneyLog(
				connection.UserId,
				m.ReadTime(data, "FromTime"),
				m.ReadTime(data, "ToTime"))

		case "ConversationAllSummaries":
			d, e = ConversationAllSummaries(
				connection.UserId,
				m.ReadString(data, "Filter"), // FILTER_ALL, FILTER_UNREAD, FILTER_PAIR
				int(m.ReadFloat64(data, "NConversation")))
		case "ConversationDetail":
			d, e = ConversationDetail(
				connection.UserId,
				m.ReadInt64(data, "ConversationId"))
		case "ConversationCreateMessage":
			d, e = ConversationCreateMessage(
				m.ReadInt64(data, "ConversationId"),
				connection.UserId,
				m.ReadString(data, "MessageContent"))
		case "ConversationAddMember":
			d, e = ConversationAddMember(
				connection.UserId,
				m.ReadInt64(data, "ConversationId"),
				m.ReadInt64(data, "NewMemberId"),
				m.ReadBool(data, "IsModerator"))
		case "ConversationRemoveMember":
			d, e = ConversationRemoveMember(
				connection.UserId,
				m.ReadInt64(data, "ConversationId"),
				m.ReadInt64(data, "MemberId"))
		case "ConversationBlockMember":
			d, e = ConversationBlockMember(
				connection.UserId,
				m.ReadInt64(data, "ConversationId"),
				m.ReadInt64(data, "MemberId"))
		case "ConversationMute":
			d, e = ConversationMute(
				connection.UserId,
				m.ReadInt64(data, "ConversationId"))
		case "ConversationMarkMessage":
			d, e = ConversationMarkMessage(
				connection.UserId,
				m.ReadInt64(data, "MessageId"),
				m.ReadBool(data, "HasSeen"))
		default:
			d = map[string]interface{}{"message": string(message)}
			e = errors.New(l.Get(l.M010CommandNotSupported))
		}
	}
	connection.WriteMap(e, d)
}
