package conversations

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/daominah/livestream/connections"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zdatabase"
	"github.com/lib/pq"
)

const (
	CONVERSATION_PAIR  = "CONVERSATION_PAIR"
	CONVERSATION_GROUP = "CONVERSATION_GROUP"

	FILTER_ALL    = "FILTER_ALL"
	FILTER_UNREAD = "FILTER_UNREAD"
	FILTER_PAIR   = "FILTER_PAIR"

	N_MESSAGE_DEFAULT      = 3
	N_CONVERSATION_DEFAULT = 3

	DISPLAY_TYPE_NORMAL = "DISPLAY_TYPE_NORMAL"
	DISPLAY_TYPE_BIG    = "DISPLAY_TYPE_BIG"
	DISPLAY_TYPE_CHEER  = "DISPLAY_TYPE_CHEER"

	CHEER_FOR_USER = "CHEER_FOR_USER"
	CHEER_FOR_TEAM = "CHEER_FOR_TEAM"

	COMMAND_NEW_MESSAGE  = "NewMessage"
	COMMAND_SEEN_MESSAGE = "SeenMessage"
)

// map conversationId to Conversation object
var MapConversations map[int64]*Conversation

// map messageId to Message object
var MapMessages map[int64]*Message
var GMutex sync.Mutex

var TestVar int

func init() {
	_ = fmt.Println
	TestVar += 1
	MapConversations = make(map[int64]*Conversation)
	MapMessages = make(map[int64]*Message)
}

type Conversation struct {
	Id   int64
	Name string
	Type string
	// map userId to memberInConversation
	Members  map[int64]*Member
	Messages []*Message `json:"-"`
	Mutex    sync.Mutex
}

type Member struct {
	ConversationId int64
	UserId         int64
	IsModerator    bool
	IsBlocked      bool
	IsMute         bool
}

type Message struct {
	MessageId      int64
	ConversationId int64
	SenderId       int64
	MessageContent string
	DisplayType    string
	CreatedTime    time.Time
	Recipients     map[int64]*Recipient
	Mutex          sync.Mutex
}

type Recipient struct {
	MessageId   int64
	RecipientId int64
	HasSeen     bool
	SeenTime    time.Time
}

// json format, msgs excluded to avoid concurrent rw map
func (c *Conversation) String() string {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	bs, e := json.MarshalIndent(c, "", "    ")
	if e != nil {
		return "{}"
	}
	return string(bs)
}

//
func (u *Conversation) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	s := u.String()
	json.Unmarshal([]byte(s), &result)
	msgs := make([]map[string]interface{}, 0)
	u.Mutex.Lock()
	for _, msg := range u.Messages {
		msgs = append(msgs, msg.ToMap())
	}
	u.Mutex.Unlock()
	return result
}

// for show all conv of 1 user
func (c *Conversation) ToShortMap(userId int64) map[string]interface{} {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	if len(c.Messages) >= 1 {
		lastMsg := c.Messages[len(c.Messages)-1]
		lastMsgSenderN, _ := users.GetProfilenameById(lastMsg.SenderId)
		hasSeen := false
		lastMsg.Mutex.Lock()
		recipient := lastMsg.Recipients[userId]
		hasSeen = recipient.HasSeen
		lastMsg.Mutex.Unlock()
		result := map[string]interface{}{
			"Id":             c.Id,
			"Name":           c.Name,
			"LastMsgSender":  lastMsgSenderN,
			"LastMsgTime":    lastMsg.CreatedTime,
			"LastMsgContent": lastMsg.MessageContent,
			"HasSeen":        hasSeen,
		}
		return result
	} else {
		result := map[string]interface{}{
			"Id":   c.Id,
			"Name": c.Name,
		}
		return result
	}
}

func (m *Message) String() string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	bs, e := json.MarshalIndent(m, "", "    ")
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (u *Message) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	s := u.String()
	json.Unmarshal([]byte(s), &result)
	return result
}

// return conversationId, error
func CreateConversation(
	memberIds []int64, moderatorIds []int64, conversationType string) (
	int64, error) {
	memberIds = misc.SortedInt64s(memberIds)
	names := []string{}
	for _, uid := range memberIds {
		name, _ := users.GetProfilenameById(uid)
		names = append(names, name)
	}
	conversationName := strings.Join(names, ", ")
	var pairKey sql.NullString
	if conversationType == CONVERSATION_PAIR {
		if len(memberIds) != 2 {
			return 0, errors.New("CONVERSATION_PAIR len(memberIds) != 2")
		}
		pairKey.Scan(fmt.Sprintf("%v_%v", memberIds[0], memberIds[1]))
	}
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO conversation
		    (conversation_type, pair_key, name)
		VALUES ($1, $2, $3) RETURNING id`,
		conversationType, pairKey, conversationName)
	var cid int64
	e := row.Scan(&cid)
	if e != nil {
		pqErr, isOk := e.(*pq.Error)
		if !isOk {
			return 0, e
		} else {
			if pqErr.Code.Name() == "unique_violation" {
				oldRow := zdatabase.DbPool.QueryRow(
					`SELECT id FROM  conversation 
            		WHERE pair_key = $1`,
					pairKey)
				var oldConvId int64
				err := oldRow.Scan(&oldConvId)
				if err != nil {
					return 0, err
				}
				return oldConvId, errors.New(l.Get(l.M006ConversationPairUnique))
			} else {
				return 0, e
			}
		}
	}
	//
	temps := []string{}
	args := []interface{}{}
	for i, uid := range memberIds {
		isMod := misc.FindInt64InSlice(uid, moderatorIds) != -1
		if conversationType == CONVERSATION_PAIR {
			isMod = true
		}
		temps = append(temps, fmt.Sprintf("($%v, $%v, $%v)", 3*i+1, 3*i+2, 3*i+3))
		args = append(args, []interface{}{cid, uid, isMod}...)
	}
	queryPart := strings.Join(temps, ", ")
	zdatabase.DbPool.Exec(fmt.Sprintf(
		`INSERT INTO conversation_member
		    (conversation_id, user_id, is_moderator)
		VALUES %v`, queryPart),
		args...)
	//
	LoadConversation(cid, N_MESSAGE_DEFAULT)
	return cid, nil
}

// Adding member to a PAIR will create a new GROUP,
// return newConversationId, error
func AddMember(conversationId int64, newMemberId int64, isModerator bool) (
	int64, error) {
	conversation, e := GetConversation(conversationId)
	if e != nil {
		return 0, e
	}
	var newConvId = conversationId
	if conversation.Type == CONVERSATION_PAIR {
		conversation.Mutex.Lock()
		memberIds := []int64{}
		for uid, _ := range conversation.Members {
			memberIds = append(memberIds, uid)
		}
		conversation.Mutex.Unlock()
		newConvId, e = CreateConversation(memberIds, []int64{}, CONVERSATION_GROUP)
		if e != nil {
			return 0, e
		}
	}
	//
	_, e = zdatabase.DbPool.Exec(
		`INSERT INTO conversation_member
		    (conversation_id, user_id, is_moderator)
		VALUES ($1, $2, $3)`,
		newConvId, newMemberId, isModerator)
	if e != nil {
		return 0, e
	}
	newConv, e := GetConversation(newConvId)
	if e != nil {
		return 0, e
	}
	//
	newConv.Mutex.Lock()
	newConv.Members[newMemberId] = &Member{
		ConversationId: newConv.Id, UserId: newMemberId,
		IsBlocked: false, IsModerator: isModerator, IsMute: false}
	newConv.Mutex.Unlock()
	return newConvId, nil
}

// cannot remove member in PAIR
func RemoveMember(conversationId int64, memberId int64) error {
	conversation, e := GetConversation(conversationId)
	if e != nil {
		return e
	}
	if conversation.Type == CONVERSATION_PAIR {
		return errors.New(l.Get(l.M007ConversationPairRemove))
	}
	conversation.Mutex.Lock()
	delete(conversation.Members, memberId)
	conversation.Mutex.Unlock()
	zdatabase.DbPool.Exec(
		`DELETE FROM conversation_member
	    WHERE user_id = $1`,
		memberId)
	return nil
}

// can block or unblock
func BlockMember(conversationId int64, memberId int64, isBlock bool) error {
	conversation, e := GetConversation(conversationId)
	if e != nil {
		return e
	}
	conversation.Mutex.Lock()
	member := conversation.Members[memberId]
	conversation.Mutex.Unlock()
	if member == nil {
		return errors.New(l.Get(l.M003ConversationOutsider))
	}
	member.IsBlocked = isBlock
	zdatabase.DbPool.Exec(
		`UPDATE conversation_member
	    SET is_blocked = $3
	    WHERE conversation_id = $1 AND user_id = $2`,
		conversationId, memberId, isBlock)
	return nil
}

// can mute or unmute
func MuteMember(conversationId int64, memberId int64, isMute bool) error {
	conversation, e := GetConversation(conversationId)
	if e != nil {
		return e
	}
	conversation.Mutex.Lock()
	member := conversation.Members[memberId]
	conversation.Mutex.Unlock()
	if member == nil {
		return errors.New(l.Get(l.M003ConversationOutsider))
	}
	member.IsMute = isMute
	zdatabase.DbPool.Exec(
		`UPDATE conversation_member
	    SET is_mute = $3
	    WHERE conversation_id = $1 AND user_id = $2`,
		conversationId, memberId, isMute)
	return nil
}

// member send a message to a conversation
func CreateMessage(
	conversationId int64, senderId int64, messageContent string,
	displayType string) error {
	conversation, e := GetConversation(conversationId)
	if e != nil {
		return e
	}
	conversation.Mutex.Lock()
	sender := conversation.Members[senderId]
	conversation.Mutex.Unlock()
	if sender == nil {
		return errors.New(l.Get(l.M003ConversationOutsider))
	}
	if sender.IsBlocked {
		return errors.New(l.Get(l.M004ConversationBlockedMember))
	}
	//
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO conversation_message
	        (conversation_id, sender_id, message_content, display_type)
	    VALUES ($1, $2, $3, $4) RETURNING message_id`,
		conversationId, senderId, messageContent, displayType)
	var mid int64
	e = row.Scan(&mid)
	if e != nil {
		return errors.New("CreateMessage Insert message: " + e.Error())
	}
	//
	temps := []string{}
	args := []interface{}{}
	i := 0
	for uid, _ := range conversation.Members {
		temps = append(temps, fmt.Sprintf("($%v, $%v)", 2*i+1, 2*i+2))
		args = append(args, []interface{}{mid, uid}...)
		i += 1
	}
	queryPart := strings.Join(temps, ", ")
	_, e = zdatabase.DbPool.Exec(fmt.Sprintf(
		`INSERT INTO conversation_message_recipient
		    (message_id, recipient_id)
		VALUES %v`, queryPart),
		args...)
	if e != nil {
		return errors.New("CreateMessage Insert recipient " + e.Error())
	}
	//
	msg, e := LoadMessage(mid)
	if e != nil {
		return errors.New("CreateMessage LoadMessage " + e.Error())
	}
	conversation.Mutex.Lock()
	conversation.Messages = append(conversation.Messages, msg)
	conversation.Mutex.Unlock()
	//
	conversation.Mutex.Lock()
	for _, member := range conversation.Members {
		if !member.IsMute {
			connections.WriteMapToUserId(member.UserId, nil,
				map[string]interface{}{
					"Command":    COMMAND_NEW_MESSAGE,
					"NewMessage": msg.ToMap(),
				})
		}
	}
	conversation.Mutex.Unlock()
	return nil
}

// set a message is read or unread
func UserMarkMessage(userId int64, messageId int64, hasSeen bool) error {
	now := time.Now()
	_, e := zdatabase.DbPool.Exec(
		`UPDATE conversation_message_recipient
		SET has_seen = $1, seen_time = $2
		WHERE message_id = $3 AND recipient_id = $4`,
		hasSeen, now, messageId, userId)
	//
	GMutex.Lock()
	message := MapMessages[messageId]
	if message != nil {
		message.Mutex.Lock()
		if message.Recipients[userId] != nil {
			message.Recipients[userId].HasSeen = hasSeen
			message.Recipients[userId].SeenTime = now
		}
		message.Mutex.Unlock()
	}
	GMutex.Unlock()
	//
	conversation, e := GetConversation(message.ConversationId)
	if e != nil {
		return e
	}
	for _, member := range conversation.Members {
		if !member.IsMute {
			connections.WriteMapToUserId(member.UserId, nil,
				map[string]interface{}{
					"Command":        COMMAND_SEEN_MESSAGE,
					"ChangedMessage": message.ToMap(),
				})
		}
	}
	return e
}
