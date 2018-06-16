package language

var mapMessagesEnglish map[string]string

func init() {
	mapMessagesEnglish = map[string]string{
		M001DuplicateUsername: "Duplicate username",
		M002InvalidLogin:      "Username or password is invalid",

		M012DuplicateTeamName:           "Duplicate team name",
		M013SetTeamCaptainOutsider:      "Can only set team's member as captain",
		M014DuplicateTeamJoiningRequest: "Duplicate team joining request",
		M015MemberMultipleTeam:          "User can only belong to one team",
		M016TeamMultipleCaptain:         "Team can only have one captain",
		M017TeamMemberPrivilege:         "Only team's captain can manage members",

		M003ConversationOutsider:      "Outsider can't send or read message from group",
		M004ConversationBlockedMember: "You have been blocked",
		M005ConversationInvalidId:     "M005ConversationInvalidId",
		M006ConversationPairUnique:    "M006ConversationPairUnique",
		M007ConversationPairRemove:    "Cannot remove member from pair conversation (but you can block)",
		M011ConversationModPrivilege:  "Need moderator privilege",

		M008Disconnected:        "Disconnected from server",
		M009LoggedInDiffDevice:  "Your account was logged in from difference device",
		M010CommandNotSupported: "This command is not supported",

		M018NotEnoughMoney:    "Not enough money",
		M019MoneyTypeNotExist: "Money type does not exist",
	}
}