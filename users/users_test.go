package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/nbackend"
	"github.com/daominah/livestream/zglobal"
)

func Test01(t *testing.T) {
	nbackend.InitBackend(nil)
	fmt.Print("")
	cases := [][]string{
		[]string{"Đào Thị Lán", "oThLn"},
		[]string{"tung.daothanhtung@gmail.com", "tungdaothanhtung@gmailcom"},
		[]string{"ta_la_Tung_208^^hohohaha", "ta_la_Tung_208hohohaha"},
		[]string{"Lisbeth Salander", "LisbethSalander"},
	}
	for _, c := range cases {
		name := c[0]
		r := c[1]
		e := NormalizeName(name)
		if r != e {
			t.Error(name, r, e)
		}
	}
}

func Test02(t *testing.T) {
	cases := [][]string{
		[]string{"123qwe", "164f04b29f50874c9330ee60d23a6ff04279c8b21a79afb5721602c6b97e2ac24d7c2070eba5827cab5f3b503bfac26539ec479921c1abadeac4980fcbf3b8a6"},
	}
	for _, c := range cases {
		pw := c[0]
		r := c[1]
		e := HassPassword(pw)
		if r != e {
			t.Error(pw, r, e)
		}
	}
}

func T1est03(t *testing.T) {
	var e error
	var user *User
	var cookie string
	var waitGroup sync.WaitGroup
	user, cookie, e = LoginByPassword("daominah", "123qwe")
	if e != nil {
		t.Error(e)
	}
	user, e = LoginByCookie(cookie)
	if e != nil {
		t.Error(e)
	}
	//	fmt.Println(user.String())
	moneyType := MT_CASH
	nChanges := 1000
	mb := user.MapMoney[moneyType]
	mb1 := MapIdToUser[user.Id].MapMoney[moneyType]
	for i := 0; i < nChanges; i++ {
		waitGroup.Add(1)
		go func() {
			_, e = ChangeUserMoney(user.Id, moneyType, 1, REASON_ADMIN_CHANGE, false)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	ma1 := MapIdToUser[user.Id].MapMoney[moneyType]
	user, _ = LoginByCookie(cookie)
	// fmt.Println(user.String())
	ma := user.MapMoney[moneyType]
	if ma-mb != float64(nChanges) {
		t.Error("ma-mb != nChanges")
	}
	if ma1-mb1 != float64(nChanges) {
		t.Error("ma1-mb1 != nChanges                                                                      ")
	}
}

func Test15(t *testing.T) {
	_, err, moneyLogId :=
		ChangeUserMoney2(1, MT_CASH, 100, "TEST", false)
	if err != nil {
		t.Error()
	}
	// fmt.Println("moneyLogId", moneyLogId)
	if moneyLogId == 0 {
		t.Error()
	}
}

func Test04(t *testing.T) {
	fT, _ := time.Parse(time.RFC3339Nano, "2018-05-28T15:20:25.606000000+07:00")
	tT, _ := time.Parse(time.RFC3339Nano, "2018-05-28T15:20:25.612000000+07:00")
	rs, e := ViewMoneyLog(1, fT, tT)
	prettyBs, _ := json.MarshalIndent(rs, "", "    ")
	_, _ = e, string(prettyBs)
	//	fmt.Println(e, string(prettyBs))
}

func Test05(t *testing.T) {
	id1 := int64(1)
	id2 := int64(2)
	Unfollow(id1, id2)
	u2, _ := LoadUser(id2)
	nFollowers1 := u2.NFollowers
	e1 := Follow(id1, id2)
	e2 := Follow(id1, id2)
	nFollowers2 := u2.NFollowers
	e3 := Unfollow(id1, id2)
	e4 := Unfollow(id1, id2)
	if !((e1 == nil) && (e2 != nil) && (e3 == nil) && (e4 == nil)) {
		t.Error(e1, e2, e3, e4)
	}
	nFollowers3 := u2.NFollowers
	if nFollowers2-nFollowers1 != 1 {
		t.Error(nFollowers2-nFollowers1, " != 1")
	}
	if nFollowers3-nFollowers2 != -1 {
		t.Error(nFollowers3-nFollowers2, " != -1")
	}
	Follow(1, 2)
	if CheckIsFollowing(1, 2) != true {
		t.Error()
	}
	if CheckIsFollowing(4, 2) {
		t.Error()
	}
}

func Test07(t *testing.T) {
	type Case struct {
		UserId   int64
		Username string
		Err      error
	}
	for i, c := range []Case{
		Case{1, "daominah", nil},
		Case{2, "daominah2", nil},
		Case{-1, "", errors.New(l.Get(l.M022InvalidUserId))},
	} {
		realityF1, realityF2 := c.Username, c.Err
		expectationF1, expectationF2 := GetUsernameById(c.UserId)
		// fmt.Println("GetUsernameById", c.UserId, realityF1, realityF2, expectationF1, expectationF2)
		if !((realityF1 == expectationF1) &&
			((realityF2 == nil && expectationF2 == nil) ||
				(realityF2 != nil && expectationF2 != nil) &&
					(realityF2.Error() == expectationF2.Error()))) {
			t.Error(i)
		}
	}
}

func Test08(t *testing.T) {
	type Case struct {
		UserId      int64
		ProfileName string
		Err         error
	}
	for i, c := range []Case{
		Case{1, "Dao Min Ah A1", nil},
		Case{2, "Dao Min Ah B2", nil},
		Case{-1, "", errors.New(l.Get(l.M022InvalidUserId))},
	} {
		realityF1, realityF2 := c.ProfileName, c.Err
		expectationF1, expectationF2 := GetProfilenameById(c.UserId)
		if !((realityF1 == expectationF1) &&
			((realityF2 == nil && expectationF2 == nil) ||
				(realityF2 != nil && expectationF2 != nil) &&
					(realityF2.Error() == expectationF2.Error()))) {
			t.Error(i)
		}
	}
}

func Test09(t *testing.T) {
	CreateTeam("We love MinAh", "image0", "summary0")
	_, e0 := CreateTeam("We love MinAh", "image0", "summary0")
	if e0 == nil || e0.Error() != l.Get(l.M012DuplicateTeamName) {
		t.Error()
	}
	teamId, _ := LoadTeamIdByName("We love MinAh")
	AddTeamMember(teamId, 1)
	AddTeamMember(teamId, 2)
	AddTeamMember(teamId, 3)
	AddTeamMember(teamId, 4)
	e1 := AddTeamMember(teamId, 1)
	if e1 == nil || e1.Error() != l.Get(l.M015MemberMultipleTeam) {
		t.Error(e1)
	}
	SetTeamCaptain(teamId, 1)
	e2 := SetTeamCaptain(teamId, 2)
	if e2 == nil || e2.Error() != l.Get(l.M016TeamMultipleCaptain) {
		t.Error(e2)
	}
	RequestJoinTeam(teamId, 1)
	e3 := RequestJoinTeam(teamId, 1)
	if e3 == nil || e3.Error() != l.Get(l.M014DuplicateTeamJoiningRequest) {
		t.Error(e3)
	}
	e4 := RemoveRequestJoinTeam(teamId, 1)
	if e4 != nil {
		t.Error(e4)
	}
	RequestJoinTeam(teamId, 1)
	rs, e5 := LoadTeamJoiningRequests(teamId)
	// fmt.Println(rs)
	if e5 != nil || len(rs) == 0 {
		t.Error(e5)
	}
}

func Test10(t *testing.T) {
	user, _ := GetUser(1)
	// fmt.Println("user", user.String(), user.ToShortMap())
	if user.TeamId == 0 {
		t.Error()
	}
	team, e := GetTeam(user.TeamId)
	if e != nil {
		t.Error(team, e)
	}
}

func Test11(t *testing.T) {
	u1, _ := GetUser(1)
	u2, _ := GetUser(2)
	ChangeUserMoney(u1.Id, MT_CASH, -u1.MapMoney[MT_CASH], REASON_ADMIN_CHANGE, true)
	ChangeUserMoney(u2.Id, MT_CASH, -u2.MapMoney[MT_CASH], REASON_ADMIN_CHANGE, true)
	ChangeUserMoney(u1.Id, MT_CASH, 300, REASON_ADMIN_CHANGE, true)
	TransferMoney(u1.Id, u2.Id, MT_CASH, 100, REASON_TRANSFER, 0.05)
	if !(u1.MapMoney[MT_CASH] == 200 && u2.MapMoney[MT_CASH] == 95) {
		t.Error()
	}
}

func Test12(t *testing.T) {
	LoadUser(3)
	loginId, e := RecordLogin(3, "1.2.3.4:12345", "deviceName", "appName")
	if e != nil {
		t.Error(e)
	}
	time.Sleep(100 * time.Millisecond)
	e = RecordLogout(loginId)
	if e != nil {
		t.Error(e)
	}
}

func Test13(t *testing.T) {
	var us []map[string]interface{}
	var e error
	us, e = Search("Tùng")
	if e != nil {
		t.Error(e)
	}
	//	fmt.Println("users", us)
	us, e = Search("9")
	if e != nil {
		t.Error(e)
	}
	_ = us
	//fmt.Println("users", us)
	us, e = Search("123")
	if e != nil {
		t.Error(e)
	}
	_ = us
	//	fmt.Println("users", us)
}

func Test14(t *testing.T) {
	e := ChangeUserInfo(11, "Đào Thị Lán", "", SEX_FEMALE, "VN", "Hưng Yên",
		"", "lan.jpg", "smiley dream")
	if e != nil {
		t.Error(e)
	}
	e = ChangeUserInfo(-1, "haha", "", "", "", "", "", "", "")
	if e == nil || e.Error() != l.Get(l.M022InvalidUserId) {
		t.Error()
	}
	e = ChangeUserInfo(11, "haha", "", "", "VN1", "", "", "", "")
	if e == nil || e.Error() != l.Get(l.M021InvalidCountry) {
		t.Error()
	}
	e = ChangeUserInfo(11, "", "", "Nữ", "", "", "", "", "")
	if e == nil || e.Error() != l.Get(l.M020InvalidSex) {
		t.Error()
	}
	e = ChangeUserInfo(12, "Trịnh Thị Vân", "111222333", SEX_FEMALE, "VN",
		"Bắc Ninh", "Vân thiếu máu", "van.png", "cute sexy")
	if e != nil {
		t.Error(e)
	}
}

func TestCharging(t *testing.T) {
	var e error
	user_id := int64(1)
	charging_type := "MinAhCard"
	charging_input := map[string]interface{}{
		"CardVendor": "Viettel", "CardSerial": "456", "CardCode": "123123"}
	chargingId, e := chargingInitDbRow(user_id, charging_type, charging_input)
	// fmt.Println("chargingId", chargingId)
	if e != nil {
		t.Error(e)
	}
	e = chargingSaveThirdPartyResponse(chargingId,
		"http_request", "http_response", float64(10000), "transaction_id_3rd_party",
		true, "")
	if e != nil {
		t.Error(e)
	}
	e = chargingChangeInAppMoney(chargingId, user_id, 12000)
	if e != nil {
		t.Error(e)
	}
}

func TestWithdrawing(t *testing.T) {
	u1, _ := GetUser(1)
	ChangeUserMoney(1, MT_CASH, -u1.MapMoney[MT_CASH], "TestWithdrawing", false)
	ChangeUserMoney(1, MT_CASH, 300000, "TestWithdrawing", false)
	user_id := int64(1)
	withdrawing_type := "HohohahaCard"
	in_app_value := float64(70000)
	vnd_value := float64(50000)
	var waitGroup sync.WaitGroup
	successfulIds := []int64{}
	var counterLock sync.Mutex
	for i := 0; i < 1000; i++ {
		waitGroup.Add(1)
		go func() {
			withdrawingId, e := withdrawingInitDbRow(user_id, withdrawing_type, in_app_value, vnd_value)
			counterLock.Lock()
			if e == nil && withdrawingId != 0 {
				successfulIds = append(successfulIds, withdrawingId)
			}
			counterLock.Unlock()
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	if len(successfulIds) != 4 {
		t.Error()
	}
	// fmt.Println("successfulIds", successfulIds)
	for i := 0; i < 10; i++ {
		waitGroup.Add(1)
		go func() {
			e := withdrawingAdminDeny(successfulIds[0], "Bố thích thế thôi")
			_ = e
			// fmt.Println("withdrawingAdminDeny e", e)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	if u1.MapMoney[MT_CASH] != 90000 {
		t.Error()
	}
	e1 := withdrawingSaveThirdPartyResponse(
		successfulIds[1], "http_request", "http_response", "transaction_id_3rd_party",
		`{"CardCode":"123123","CardSerial":"456","CardVendor":"Viettel"}`,
		true, "")
	e2 := withdrawingSaveThirdPartyResponse(
		successfulIds[2], "http_request", "http_response", "transaction_id_3rd_party",
		"",
		false, "Kho hết thẻ")
	e3 := withdrawingSaveThirdPartyResponse(
		successfulIds[3], "http_request", "http_response", "244728",
		`{"status": "1", "bank_name": "Test Bank", "account": "977", "apikey": "357", "name": "Jon Doe", "created_at": "2018-03-18 19:53:42 Asia/Kuala_Lumpur", "telephone": "", "contract": "323", "currency": "MYR", "amount": "17.16", "transaction": "244728", "item_description": "item_description", "signature": "e5a25eddda57c97d129a9bec5623468c9afd8f5b5169fbfe6fa6cb4bd4b4312b", "item_id": "item_id", "status_message": "Accepted", "email": "user@tmt.com", "bank_account": "00000000000"}`,
		true, "")
	if e1 != nil || e2 != nil || e3 != nil {
		t.Error(e1, e2, e3)
	}
}

func Test16(t *testing.T) {
	zglobal.MoneyIOPaytrustMapBankNameToBankCode = map[string]string{
		"VietinBank":  "5a8d9b3432bc7",
		"BIDV":        "5a8dc25912217",
		"TechComBank": "5a8ee643945a3",
		"SacomBank":   "5a8eec3fc74e6",
		"DongABank":   "5a904bc3775ba"}
	url, err := rPaytrust(1, "TechComBank", 100000, 63)
	_, _ = url, err
	fmt.Println(url, err)
}
