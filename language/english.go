package language

var mapMessagesEnglish map[string]string

func init() {
	mapMessagesEnglish = map[string]string{
		M001DuplicateUsername: "Duplicate username",
		M002InvalidLogin:      "Username or password is invalid",

		M003ConversationOutsider:      "Outsider can't send or read message from group",
		M004ConversationBlockedMember: "You have been blocked",
		M005ConversationInvalidId:     "ConversationId is invalid",
		M006ConversationPairUnique:    "Duplicate pair conversation",
		M007ConversationPairRemove:    "Cannot remove member from pair conversation (but you can block)",
		M011ConversationModPrivilege:  "Need moderator privilege",

		M008Disconnected:        "Disconnected from server",
		M009LoggedInDiffDevice:  "Your account was logged in from difference device",
		M010CommandNotSupported: "This command is not supported",
	}
}
