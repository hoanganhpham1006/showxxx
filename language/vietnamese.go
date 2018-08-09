package language

var mapMessagesVietnamese map[string]string

func init() {
	mapMessagesVietnamese = map[string]string{
		M001DuplicateUsername:     "Trùng tên đăng nhập",
		M002InvalidLogin:          "Đăng nhập không hợp lệ",
		M020InvalidSex:            "Sex must be in [SEX_MALE, SEX_FEMALE, SEX_UNDEFINED]",
		M021InvalidCountry:        "Quốc gia phải có định dạng ISO 3166-1 alpha-2: VN, US, GB,.. ",
		M022InvalidUserId:         "Sai mã nhận dạng người chơi",
		M024InvalidLoginType:      "Sai kiểu đăng nhập",
		M025UserSuspended:         "Tài khoản của bạn đã bị tạm ngưng hoạt động",
		M030InvalidRole:           "Sai kiểu người dùng.",
		M031OperationNotPermitted: "Hành động không được cho phép",

		M012DuplicateTeamName:           "Tên đội bị trùng",
		M013SetTeamCaptainOutsider:      "Chỉ có thành viên mới có thể thành đội trưởng",
		M014DuplicateTeamJoiningRequest: "Bạn đã xin gia nhập đội này rồi, vui lòng chờ",
		M015MemberMultipleTeam:          "Bạn chỉ có thể là thành viên của một đội",
		M016TeamMultipleCaptain:         "Một đội không được có nhiều đội trưởng",
		M017TeamMemberPrivilege:         "Chỉ có đội trưởng có thể thay đổi thành viên của đội",

		M003ConversationOutsider:      "Người ngoài nhóm không được gửi hoặc đọc tin nhắn",
		M004ConversationBlockedMember: "Bạn đã bị chặn gửi tin cho nhóm này",
		M005ConversationInvalidId:     "M005ConversationInvalidId",
		M006ConversationPairUnique:    "M006ConversationPairUnique",
		M007ConversationPairRemove:    "Không thể thoát cuộc nói chuyện riêng (nhưng có thể chặn)",
		M011ConversationModPrivilege:  "Cần quyền quản lí",

		M008Disconnected:        "Mất kết nối đến máy chủ",
		M009LoggedInDiffDevice:  "Tài khoản của bạn bị đăng nhập từ thiết bị khác",
		M010CommandNotSupported: "Câu lệnh không đúng",

		M018NotEnoughMoney:    "Không đủ tiền",
		M019MoneyTypeNotExist: "Loại tiền không tồn tại",

		M038InvalidBankName:         "M038InvalidBankName",
		M039CanOnlyDenyWithdrawOnce: "M039CanOnlyDenyWithdrawOnce",

		M023StaticServerDown: "Máy chủ lưu trữ không hoạt động",

		M026StreamCreatePrivilege:     "Bạn không có quyền stream",
		M027StreamBroadcasted:         "Bạn đang phát sóng rồi",
		M028StreamNotBroadcasting:     "Người bạn muốn xem đang không phát sóng",
		M029StreamConcurrentView:      "Bạn chỉ có thể xem một người tại một thời điểm",
		M032StreamOnlyViewerCanReport: "Phải là người xem mới được báo cáo vi phạm",

		M033GameNeedToChooseMoneyType: "M033GameNeedToChooseMoneyType",
		M034GameInvalidBaseMoney:      "M034GameInvalidBaseMoney",
		M035GameInvalidGameCode:       "M035GameInvalidGameCode",
		M036GameOnlyOneMatchAtATime:   "M036GameOnlyOneMatchAtATime",
		M037GameInvalidMatchId:        "M037GameInvalidMatchId",
		M042GameInvalidCarIndex:       "M042GameInvalidCarIndex",
		M043MovingDurationEnded:       "M043MovingDurationEnded",
	}
}
