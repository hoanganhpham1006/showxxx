package conversations

import (
	"encoding/json"
	"fmt"
	"testing"

	l "github.com/daominah/livestream/language"
)

func T1est01(t *testing.T) {
	_ = fmt.Println
	_ = json.MarshalIndent
	_ = l.Get
	var e error
	cid, e := CreateConversation([]int64{1, 3}, []int64{1}, CONVERSATION_GROUP)
	if e != nil {
		t.Error(e)
	}
	e = CreateMessage(cid, 1, "Tao 1 day")
	if e != nil {
		t.Error(e)
	}
	conv, _ := GetConversation(cid)
	conv.Mutex.Lock()
	if !((len(conv.Messages) == 1) && (len(conv.Members) == 2)) {
		t.Error()
	}
	conv.Mutex.Unlock()

	e = CreateMessage(cid, 3, "Uh nghe roi, tao 3 nay")
	conv.Mutex.Lock()
	if !(len(conv.Messages) == 2) {
		t.Error()
	}
	conv.Mutex.Unlock()

	e = CreateMessage(cid, 2, "Chung may thi tham cai deo gi the")
	if (e == nil) || (e.Error() != l.Get(l.M003ConversationOutsider)) {
		t.Error(e)
	}
	CreateMessage(cid, 3, "Ke bon tao")

	CreateConversation([]int64{1, 2}, []int64{}, CONVERSATION_PAIR)
	_, e = CreateConversation([]int64{1, 2}, []int64{}, CONVERSATION_PAIR)
	if (e == nil) || (e.Error() != l.Get(l.M006ConversationPairUnique)) {
		t.Error(e)
	}
	cid = LoadConversationPairId(1, 2)
	if cid == 0 {
		t.Error()
	}
	e = CreateMessage(cid, 1, "1 chat rieng voi 2")
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
