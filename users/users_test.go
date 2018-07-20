package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	l "github.com/daominah/livestream/language"
)

func Test01(t *testing.T) {
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
