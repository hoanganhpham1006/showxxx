package main

import (
	"encoding/json"
	"errors"

	l "github.com/daominah/livestream/language"
	m "github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/nbackend"
	"github.com/daominah/livestream/nwebsocket"
)

func doAfterReceivingMessage(connection *nwebsocket.Connection, message []byte) {
	var data map[string]interface{}
	parseTextErr := json.Unmarshal(message, &data)
	if parseTextErr != nil {
		connection.WriteMap(errors.New(
			"handleClientMessage json.Unmarshal err: "+parseTextErr.Error()), nil)
		return
	}

	proxyId := m.ReadInt64(data, "ProxyId")
	userId := m.ReadInt64(data, "SourceUserId")
	clientConnId := m.ReadInt64(data, "ConnId")
	clientIp := m.ReadString(data, "ClientIpAddr")

	command := m.ReadString(data, "Command")
	// unique id has been created by client that help to identify response
	// belong to what request
	commandId := m.ReadInt64(data, "CommandId")

	// responseData
	var d map[string]interface{}
	// responseError
	var e error

	if userId == 0 { // not logged in
		switch command {
		case "ProxyConnect": // exclusive command of proxies
			nbackend.GBackend.HandleProxyConnect(proxyId, connection)
		case "UserCreate":
			d, e = UserCreate(
				m.ReadString(data, "Username"),
				m.ReadString(data, "Password"))
		case "UserLoginByPassword":
			d, e = UserLogin(
				LOGIN_BY_PASSWORD,
				proxyId, clientConnId, clientIp,
				m.ReadString(data, "Username"),
				m.ReadString(data, "Password"),
				"",
				m.ReadString(data, "DeviceName"),
				m.ReadString(data, "AppName"))
		case "UserLoginByCookie":
			d, e = UserLogin(
				LOGIN_BY_COOKIE,
				proxyId, clientConnId, clientIp,
				"",
				"",
				m.ReadString(data, "LoginSession"),
				m.ReadString(data, "DeviceName"),
				m.ReadString(data, "AppName"))

		case "RankGetLeaderBoard":
			d, e = RankGetLeaderBoard(
				m.ReadInt64(data, "RankId"), // RANK_RECEIVED_CASH_DAY   = int64(3), RANK_RECEIVED_CASH_WEEK  = int64(4), RANK_RECEIVED_CASH_MONTH = int64(5), RANK_RECEIVED_CASH_ALL   = int64(6), , RANK_SENT_CASH_DAY   = int64(7), RANK_SENT_CASH_WEEK  = int64(8), RANK_SENT_CASH_MONTH = int64(9), RANK_SENT_CASH_ALL   = int64(10), , RANK_PURCHASED_CASH_DAY   = int64(11), RANK_PURCHASED_CASH_WEEK  = int64(12), RANK_PURCHASED_CASH_MONTH = int64(13), RANK_PURCHASED_CASH_ALL   = int64(14), , RANK_N_FOLLOWERS_WEEK = int64(15), RANK_N_FOLLOWERS_ALL  = int64(16)
			)
		case "StreamAllSummaries":
			d, e = StreamAllSummaries()
		case "UserDetail":
			d, e = UserDetail(
				m.ReadInt64(data, "UserId"))
		case "UserSearch":
			d, e = UserSearch(
				m.ReadString(data, "Key"),
			)
		case "AdnvidGetListVideos2":
			d, e = AdnvidGetListVideos2(
				m.ReadInt64(data, "Limit"),
				m.ReadInt64(data, "Offset"),
				m.ReadString(data, "OrderBy"),
			)
		case "AdnvidGetListAds":
			d, e = AdnvidGetListAds(
				m.ReadInt64(data, "Limit"),
				m.ReadInt64(data, "Offset"),
				m.ReadString(data, "OrderBy"),
			)

		default:
			d = map[string]interface{}{"message": string(message)}
			e = errors.New("Not logged in." + l.Get(l.M010CommandNotSupported))
		}
	} else { // logged in
		switch command {
		case "DisconnectFromClient": // exclusive command of proxies
			HandleClientDisconnect(userId, m.ReadInt64(data, "LoginId"))
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
				userId,
				m.ReadInt64(data, "TargetId"))
		case "UserUnfollow":
			d, e = UserUnfollow(
				userId,
				m.ReadInt64(data, "TargetId"))
		case "UserCheckFollowing":
			d, e = UserCheckFollowing(
				userId,
				m.ReadInt64(data, "TargetId"),
			)
		case "UserViewMoneyLog":
			d, e = UserViewMoneyLog(
				userId,
				m.ReadTime(data, "FromTime"),
				m.ReadTime(data, "ToTime"))
		case "UserSearch":
			d, e = UserSearch(
				m.ReadString(data, "Key"),
			)
		case "UserChangeInfo":
			d, e = UserChangeInfo(
				userId,
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
				userId,
				m.ReadBytes(data, "ImageBase64"),
			)
		case "UploadFile":
			d, e = UploadFile(
				m.ReadBytes(data, "FileBase64"),
			)

		case "RankGetLeaderBoard":
			d, e = RankGetLeaderBoard(
				m.ReadInt64(data, "RankId"), // RANK_RECEIVED_CASH_DAY   = int64(3), RANK_RECEIVED_CASH_WEEK  = int64(4), RANK_RECEIVED_CASH_MONTH = int64(5), RANK_RECEIVED_CASH_ALL   = int64(6), , RANK_SENT_CASH_DAY   = int64(7), RANK_SENT_CASH_WEEK  = int64(8), RANK_SENT_CASH_MONTH = int64(9), RANK_SENT_CASH_ALL   = int64(10), , RANK_PURCHASED_CASH_DAY   = int64(11), RANK_PURCHASED_CASH_WEEK  = int64(12), RANK_PURCHASED_CASH_MONTH = int64(13), RANK_PURCHASED_CASH_ALL   = int64(14), , RANK_N_FOLLOWERS_WEEK = int64(15), RANK_N_FOLLOWERS_ALL  = int64(16)
			)

		case "ConversationAllSummaries":
			d, e = ConversationAllSummaries(
				userId,
				m.ReadString(data, "Filter"), // FILTER_ALL, FILTER_UNREAD, FILTER_PAIR
				int(m.ReadFloat64(data, "NConversation")))
		case "ConversationDetail":
			d, e = ConversationDetail(
				userId,
				m.ReadInt64(data, "ConversationId"))
		case "ConversationCreate":
			d, e = ConversationCreate(
				userId,
				m.ReadInt64(data, "RecipientId"),
			)
		case "ConversationCreateMessage":
			d, e = ConversationCreateMessage(
				m.ReadInt64(data, "ConversationId"),
				userId,
				m.ReadString(data, "MessageContent"))
		case "ConversationCreateBigMessage":
			d, e = ConversationCreateBigMessage(
				m.ReadInt64(data, "ConversationId"),
				userId,
				m.ReadString(data, "MessageContent"))
		case "ConversationAddMember":
			d, e = ConversationAddMember(
				userId,
				m.ReadInt64(data, "ConversationId"),
				m.ReadInt64(data, "NewMemberId"),
				m.ReadBool(data, "IsModerator"))
		case "ConversationRemoveMember":
			d, e = ConversationRemoveMember(
				userId,
				m.ReadInt64(data, "ConversationId"),
				m.ReadInt64(data, "MemberId"))
		case "ConversationBlockMember":
			d, e = ConversationBlockMember(
				userId,
				m.ReadInt64(data, "ConversationId"),
				m.ReadInt64(data, "MemberId"))
		case "ConversationMute":
			d, e = ConversationMute(
				userId,
				m.ReadInt64(data, "ConversationId"))
		case "ConversationMarkMessage":
			d, e = ConversationMarkMessage(
				userId,
				m.ReadInt64(data, "MessageId"),
				m.ReadBool(data, "HasSeen"))

		case "Cheer":
			d, e = Cheer(
				m.ReadInt64(data, "ConversationId"),
				userId,
				m.ReadInt64(data, "TargetUserId"),
				m.ReadString(data, "CheerType"), // CHEER_FOR_TEAM, CHEER_FOR_USER
				m.ReadFloat64(data, "Value"),
				m.ReadString(data, "CheerMessage"),
				m.ReadString(data, "Misc"), // json, ex: {"Description": "9x Mangoes"}
			)
		case "MoneyCharge":
			d, e = MoneyCharge(
				userId,
				data,
				m.ReadString(data, "ChargingType"),
				m.ReadString(data, "CardVendor"),
				m.ReadString(data, "CardSerial"),
				m.ReadString(data, "CardCode"),
				m.ReadString(data, "BankName"),
				m.ReadFloat64(data, "BankVndValue"),
			)
		case "MoneyWithdraw":
			d, e = MoneyWithdraw(
				userId,
				data,
				m.ReadString(data, "WithdrawingType"),
				m.ReadString(data, "VndValue"),
			)

		case "ConversationGifts":
			d, e = ConversationGifts()

		case "TeamCreate":
			d, e = TeamCreate(
				userId,
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
				userId,
				m.ReadInt64(data, "TeamId"),
				m.ReadInt64(data, "UserId"),
			)
		case "TeamRequestJoin":
			d, e = TeamRequestJoin(
				m.ReadInt64(data, "TeamId"),
				userId,
			)
		case "TeamHandleJoiningRequest":
			d, e = TeamHandleJoiningRequest(
				m.ReadInt64(data, "TeamId"),
				m.ReadInt64(data, "UserId"),
				m.ReadBool(data, "IsAccepted"),
			)

			//		case "StreamCreate":
			//			d, e = StreamCreate(
			//				userId,
			//				m.ReadString(data, "StreamName"),
			//				m.ReadString(data, "StreamImage"),
			//			)
			//		case "StreamFinish":
			//			_ = 1
			//		case "StreamView":
			//			d, e = StreamView(
			//				userId,
			//				m.ReadInt64(data, "BroadcasterId"),
			//			)
			//		case "StreamStopViewing":
			//			_ = 1
		case "StreamAllSummaries":
			d, e = StreamAllSummaries()
		case "StreamGetMyViewing":
			d, e = StreamGetMyViewing(userId)
		case "StreamReport":
			d, e = StreamReport(
				userId,
				m.ReadInt64(data, "BroadcasterId"),
				m.ReadString(data, "Reason"),
			)

			//		case "SGameChooseMoneyType":
			//			d, e = SGameChooseMoneyType(
			//				m.ReadString(data, "GameCode"), // "egg", ..
			//				userId,
			//				m.ReadString(data, "MoneyType"), // MT_CASH
			//			)
		case "SGameChooseBaseMoney":
			d, e = SGameChooseBaseMoney(
				m.ReadString(data, "GameCode"), // "egg", ..
				userId,
				m.ReadFloat64(data, "BaseMoney"), // 100, 1000, 2000,..
			)
		case "SGameGetPlayingMatch":
			d, e = SGameGetPlayingMatch(
				m.ReadString(data, "GameCode"),
				userId,
			)

		case "SGameEggSendMove":
			d, e = SGameEggSendMove(userId, data)

		case "MGameCarGetCurrentMatch":
			d, e = MGameCarGetCurrentMatch()
		case "MGameCarSendMove":
			d, e = MGameCarSendMove(userId, data,
				m.ReadInt64(data, "CarIndex"),
				m.ReadFloat64(data, "BetValue"),
			)

		case "AdnvidGetListVideoCategories":
			d, e = AdnvidGetListVideoCategories()
		case "AdnvidGetListVideos":
			d, e = AdnvidGetListVideos(
				userId,
				m.ReadInt64(data, "Limit"),
				m.ReadInt64(data, "Offset"),
				m.ReadString(data, "OrderBy"),
			)
		case "AdnvidGetVideoInfoById":
			d, e = AdnvidGetVideoInfoById(
				userId,
				m.ReadInt64(data, "VideoId"),
			)
		case "AdnvidBuyVideo":
			d, e = AdnvidBuyVideo(
				userId,
				m.ReadInt64(data, "VideoId"),
			)
		case "AdnvidGetListAds":
			d, e = AdnvidGetListAds(
				m.ReadInt64(data, "Limit"),
				m.ReadInt64(data, "Offset"),
				m.ReadString(data, "OrderBy"),
			)
		case "AdnvidGetAdById":
			d, e = AdnvidGetAdById(
				m.ReadInt64(data, "AdId"),
			)

		default:
			d = map[string]interface{}{"message": string(message)}
			e = errors.New("Logged in. " + l.Get(l.M010CommandNotSupported))
		}
	}
	if d == nil {
		d = map[string]interface{}{}
	}
	d["Command"] = command
	d["CommandId"] = commandId
	d["ConnId"] = clientConnId
	d["SourceUserId"] = userId
	connection.WriteMap(e, d)
}
