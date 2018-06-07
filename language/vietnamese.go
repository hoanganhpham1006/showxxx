package language

var mapMessagesVietnamese map[string]string

func init() {
	mapMessagesVietnamese = map[string]string{
		M001DuplicateUsername: "Trùng tên đăng nhập",
		M002InvalidLogin:      "Tên đăng nhập hoặc mật khẩu không đúng",

		M012DuplicateTeamName:           "Tên đội bị trùng",
		M013SetTeamCaptainOutsider:      "Chỉ có thành viên mới có thể thành đội trưởng",
		M014DuplicateTeamJoiningRequest: "Bạn đã xin gia nhập đội này rồi, vui lòng chờ",

		M003ConversationOutsider:      "Người ngoài nhóm không được gửi hoặc đọc tin nhắn",
		M004ConversationBlockedMember: "Bạn đã bị chặn gửi tin cho nhóm này",
		M005ConversationInvalidId:     "M005ConversationInvalidId",
		M006ConversationPairUnique:    "M006ConversationPairUnique",
		M007ConversationPairRemove:    "Không thể thoát cuộc nói chuyện riêng (nhưng có thể chặn)",
		M011ConversationModPrivilege:  "Cần quyền quản lí",

		M008Disconnected:        "Mất kết nối đến máy chủ",
		M009LoggedInDiffDevice:  "Tài khoản của bạn bị đăng nhập từ thiết bị khác",
		M010CommandNotSupported: "Câu lệnh không đúng",
	}
}
