package language

import (
	"fmt"

	"github.com/daominah/livestream/zconfig"
)

const (
	M001DuplicateUsername = "M001DuplicateUsername"
	M002InvalidLogin      = "M002InvalidLogin"

	M012DuplicateTeamName           = "M012DuplicateTeamName"
	M013SetTeamCaptainOutsider      = "M013SetTeamCaptainOutsider"
	M014DuplicateTeamJoiningRequest = "M014DuplicateTeamJoiningRequest"

	M003ConversationOutsider      = "M003ConversationOutsiderMessage"
	M004ConversationBlockedMember = "M004ConversationBlockedMember"
	M005ConversationInvalidId     = "M005ConversationInvalidId"
	M006ConversationPairUnique    = "M006ConversationPairUnique"
	M007ConversationPairRemove    = "M007ConversationPairRemove"
	M011ConversationModPrivilege  = "M011ConversationModPrivilege"

	M008Disconnected        = "M008Disconnected"
	M009LoggedInDiffDevice  = "M009LoggedInDiffDevice"
	M010CommandNotSupported = "M010CommandNotSupported"
)

// map msgName to msgContent
var mapMessages map[string]string

func init() {
	fmt.Println("zconfig.Language", zconfig.Language)
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
