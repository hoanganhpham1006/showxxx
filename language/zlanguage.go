package language

import (
	//	"fmt"

	"github.com/daominah/livestream/zconfig"
)

const (
	M001DuplicateUsername = "M001DuplicateUsername"
	M002InvalidLogin      = "M002InvalidLogin"
	M020InvalidSex        = "M020InvalidSex"
	M021InvalidCountry    = "M021InvalidCountry"
	M022InvalidUserId     = "M022InvalidUserId"
	M024InvalidLoginType  = "M024InvalidLoginType"
	M025UserSuspended     = "M025UserSuspended"
	M030InvalidRole       = "M030InvalidRole"

	M012DuplicateTeamName           = "M012DuplicateTeamName"
	M013SetTeamCaptainOutsider      = "M013SetTeamCaptainOutsider"
	M014DuplicateTeamJoiningRequest = "M014DuplicateTeamJoiningRequest"
	M015MemberMultipleTeam          = "M015MemberMultipleTeam"
	M016TeamMultipleCaptain         = "M016TeamMultipleCaptain"
	M017TeamMemberPrivilege         = "M017TeamMemberPrivilege"

	M003ConversationOutsider      = "M003ConversationOutsiderMessage"
	M004ConversationBlockedMember = "M004ConversationBlockedMember"
	M005ConversationInvalidId     = "M005ConversationInvalidId"
	M006ConversationPairUnique    = "M006ConversationPairUnique"
	M007ConversationPairRemove    = "M007ConversationPairRemove"
	M011ConversationModPrivilege  = "M011ConversationModPrivilege"

	M008Disconnected        = "M008Disconnected"
	M009LoggedInDiffDevice  = "M009LoggedInDiffDevice"
	M010CommandNotSupported = "M010CommandNotSupported"

	M018NotEnoughMoney    = "M018NotEnoughMoney"
	M019MoneyTypeNotExist = "M019MoneyTypeNotExist"

	M023StaticServerDown = "M023StaticServerDown"

	M026StreamCreatePrivilege = "M026StreamCreatePrivilege"
	M027StreamBroadcasted     = "M027StreamBroadcasted"
	M028StreamNotBroadcasting = "M028StreamNotBroadcasting"
	M029StreamConcurrentView  = "M029StreamConcurrentView"
)

// map msgName to msgContent
var mapMessages map[string]string

func init() {
	// fmt.Println("zconfig.Language", zconfig.Language)
	if zconfig.Language == zconfig.LANG_ENGLISH {
		mapMessages = mapMessagesEnglish
	} else if zconfig.Language == zconfig.LANG_VIETNAMESE {
		mapMessages = mapMessagesVietnamese
	} else {
		mapMessages = make(map[string]string)
	}
}

// get messageContent from messageName
func Get(msgName string) string {
	return mapMessages[msgName]
}
