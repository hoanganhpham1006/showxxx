package main

import (
	"encoding/json"
	"errors"

	"github.com/daominah/livestream/connections"
	l "github.com/daominah/livestream/language"
	m "github.com/daominah/livestream/misc"
)

func doAfterReceivingMessage(connection *connections.Connection, message []byte) {
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
				m.ReadString(data, "Password"),
				m.ReadString(data, "DeviceName"),
				m.ReadString(data, "AppName"))
		case "UserLoginByCookie":
			d, e = UserLoginByCookie(
				connection,
				m.ReadString(data, "LoginSession"),
				m.ReadString(data, "DeviceName"),
				m.ReadString(data, "AppName"))

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
		case "UserCheckFollowing":
			d, e = UserCheckFollowing(
				connection.UserId,
				m.ReadInt64(data, "TargetId"),
			)
		case "UserViewMoneyLog":
			d, e = UserViewMoneyLog(
				connection.UserId,
				m.ReadTime(data, "FromTime"),
				m.ReadTime(data, "ToTime"))
		case "UserSearch":
			d, e = UserSearch(
				m.ReadString(data, "Key"),
			)
		case "UserChangeInfo":
			d, e = UserChangeInfo(
				connection.UserId,
				m.ReadString(data, "RealName"),
				m.ReadString(data, "NationalId"),
				m.ReadString(data, "Sex"),
				m.ReadString(data, "Country"),
				m.ReadString(data, "Address"),
				m.ReadString(data, "ProfileName"),
				m.ReadString(data, "ProfileImage"),
				m.ReadString(data, "Summary"),
			)
		case "UserChangeProfileImage":
			d, e = UserChangeProfileImage(
				connection.UserId,
				m.ReadBytes(data, "ImageBase64"),
			)

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
		case "ConversationCreateBigMessage":
			d, e = ConversationCreateBigMessage(
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

		case "Cheer":
			d, e = Cheer(
				m.ReadInt64(data, "ConversationId"),
				connection.UserId,
				m.ReadInt64(data, "TargetUserId"),
				m.ReadString(data, "CheerType"), // CHEER_FOR_TEAM, CHEER_FOR_USER
				m.ReadFloat64(data, "Value"),
				m.ReadString(data, "CheerMessage"),
				m.ReadString(data, "Misc"), // json, ex: {"Description": "9x Mangoes"}
			)

		case "TeamCreate":
			d, e = TeamCreate(
				connection.UserId,
				m.ReadString(data, "TeamName"),
				m.ReadString(data, "TeamImage"),
				m.ReadString(data, "TeamSummary"),
			)
		case "TeamDetail":
			d, e = TeamDetail(
				m.ReadInt64(data, "TeamId"),
			)
		case "TeamLoadJoiningRequests":
			d, e = TeamLoadJoiningRequests(
				m.ReadInt64(data, "TeamId"),
			)
		case "TeamRemoveMember":
			d, e = TeamRemoveMember(
				connection.UserId,
				m.ReadInt64(data, "TeamId"),
				m.ReadInt64(data, "UserId"),
			)
		case "TeamRequestJoin":
			d, e = TeamRequestJoin(
				m.ReadInt64(data, "TeamId"),
				connection.UserId,
			)
		case "TeamHandleJoiningRequest":
			d, e = TeamHandleJoiningRequest(
				m.ReadInt64(data, "TeamId"),
				m.ReadInt64(data, "UserId"),
				m.ReadBool(data, "IsAccepted"),
			)

		default:
			d = map[string]interface{}{"message": string(message)}
			e = errors.New(l.Get(l.M010CommandNotSupported))
		}
	}
	if d == nil {
		d = map[string]interface{}{}
	}
	d["Command"] = command
	connection.WriteMap(e, d)
}
