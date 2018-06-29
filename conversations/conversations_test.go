package conversations

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zglobal"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	_ = json.MarshalIndent
	_ = l.Get
	var e error
	cid, e := CreateConversation([]int64{1, 3}, []int64{1}, CONVERSATION_GROUP)
	if e != nil {
		t.Error(e)
	}
	e = CreateMessage(cid, 1, "Tao 1 day", DISPLAY_TYPE_NORMAL)
	if e != nil {
		t.Error(e)
	}
	conv, _ := GetConversation(cid)
	conv.Mutex.Lock()
	if !((len(conv.Messages) == 1) && (len(conv.Members) == 2)) {
		t.Error()
	}
	conv.Mutex.Unlock()
	fmt.Println(conv.String())

	e = CreateMessage(cid, 3, "Uh nghe roi, tao 3 nay", DISPLAY_TYPE_NORMAL)
	conv.Mutex.Lock()
	if !(len(conv.Messages) == 2) {
		t.Error()
	}
	conv.Mutex.Unlock()

	e = CreateMessage(cid, 2, "Chung may thi tham cai deo gi the", DISPLAY_TYPE_NORMAL)
	if (e == nil) || (e.Error() != l.Get(l.M003ConversationOutsider)) {
		t.Error(e)
	}
	CreateMessage(cid, 3, "Ke bon tao", DISPLAY_TYPE_NORMAL)

	cid1, _ := CreateConversation([]int64{1, 2}, []int64{}, CONVERSATION_PAIR)
	cid2, e := CreateConversation([]int64{1, 2}, []int64{}, CONVERSATION_PAIR)
	//	fmt.Println("cid2", cid2)
	if (e == nil) || (e.Error() != l.Get(l.M006ConversationPairUnique)) ||
		(cid1 != cid2) {
		t.Error(e)
	}
	cid = LoadConversationPairId(1, 2)
	if cid == 0 {
		t.Error()
	}
	e = CreateMessage(cid, 1, "1 chat rieng voi 2", DISPLAY_TYPE_NORMAL)
	if e != nil {
		t.Error(e)
	}

}

func Test02(t *testing.T) {
	shortedConvs, e := UserLoadAllConversations(1, "", 2)
	if e != nil {
		t.Error()
	}
	temp, _ := json.MarshalIndent(shortedConvs, "", "    ")
	_ = temp
	if len(shortedConvs) != 2 {
		t.Error()
	}

	cid := LoadConversationPairId(1, 2)
	if cid == 0 {
		t.Error()
	}
	BlockMember(cid, 1, true)
	conv, e := GetConversation(cid)
	if e != nil {
		t.Error()
	}
	if conv.Members[1].IsBlocked != true {
		t.Error()
	}
	BlockMember(cid, 1, false)
	e = MuteMember(cid, 1, true)
	if e != nil {
		t.Error()
	}
	if conv.Members[1].IsMute != true {
		t.Error()
	}
	e = RemoveMember(cid, 1)
	if e == nil || e.Error() != l.Get(l.M007ConversationPairRemove) {
		t.Error()
	}
}

func Test03(t *testing.T) {
	zglobal.CheerTax = 0.2
	zglobal.CheerTeamMainProfit = 0.85
	zglobal.CheerTeamCaptainProfit = 0.05
	cid, _ := CreateConversation([]int64{1, 2, 3, 4, 5, 6}, nil, CONVERSATION_GROUP)
	for _, uid := range []int64{1, 2, 3, 4, 5, 6} {
		user, _ := users.GetUser(uid)
		users.ChangeUserMoney(
			uid, users.MT_CASH, -user.MapMoney[users.MT_CASH], "test", false)
	}
	users.ChangeUserMoney(5, users.MT_CASH, 100, "test", false)
	e := Cheer(cid, 5, 2, CHEER_FOR_TEAM, 1250, "love to mr2", "9x tangoes")
	if e == nil || e.Error() != l.Get(l.M018NotEnoughMoney) {
		t.Error(e)
	}
	users.ChangeUserMoney(5, users.MT_CASH, 1200, "test", false)
	e = Cheer(cid, 5, 2, CHEER_FOR_TEAM, 1250, "love to mr2", "9x tangoes")
	// after run users_test: 1,2,3,4 in a team, 1 is captain
	u1, _ := users.GetUser(1)
	u2, _ := users.GetUser(2)
	u3, _ := users.GetUser(3)
	u4, _ := users.GetUser(4)
	epsilon := 0.001
	if !(math.Abs(u1.MapMoney[users.MT_CASH]-75) < epsilon &&
		math.Abs(u2.MapMoney[users.MT_CASH]-875) < epsilon &&
		math.Abs(u3.MapMoney[users.MT_CASH]-25) < epsilon &&
		math.Abs(u4.MapMoney[users.MT_CASH]-25) < epsilon) {
		t.Error(u1.MapMoney[users.MT_CASH], u2.MapMoney[users.MT_CASH],
			u3.MapMoney[users.MT_CASH], u4.MapMoney[users.MT_CASH])
	}
}
