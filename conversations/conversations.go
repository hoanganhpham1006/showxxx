package conversations

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/nbackend"
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
	DISPLAY_TYPE_ADMIN  = "DISPLAY_TYPE_ADMIN"

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
	SenderIds []int64
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
	Gift					 Gift
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
	result["Messages"] = msgs
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
		if recipient != nil {
			hasSeen = recipient.HasSeen
		}
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

// read data from database to MapConversations
func LoadConversation(cid int64, nMsgLimit int) (*Conversation, error) {
	row := zdatabase.DbPool.QueryRow(
		`SELECT name, conversation_type FROM conversation WHERE id = $1`,
		cid)
	var name, conversation_type string
	e := row.Scan(&name, &conversation_type)
	if e != nil {
		return nil, errors.New(l.Get(l.M005ConversationInvalidId))
	}
	conversation := &Conversation{
		Id: cid, Name: name, Type: conversation_type}
	//
	conversation.Members = make(map[int64]*Member)
	rows, e := zdatabase.DbPool.Query(
		`SELECT user_id, is_moderator, is_blocked, is_mute 
	    FROM conversation_member WHERE conversation_id = $1`,
		cid)
	if e != nil {
		return nil, errors.New("LoadConversation Members:" + e.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var user_id int64
		var is_moderator, is_blocked, is_mute bool
		e = rows.Scan(&user_id, &is_moderator, &is_blocked, &is_mute)
		if e != nil {
			return nil, e
		}
		conversation.Members[user_id] = &Member{
			ConversationId: cid, UserId: user_id,
			IsBlocked: is_blocked, IsModerator: is_moderator, IsMute: is_mute}
	}
	//
	conversation.Messages = make([]*Message, 0)
	rows2, e := zdatabase.DbPool.Query(
		`SELECT message_id
	    FROM conversation_message WHERE conversation_id = $1
	    ORDER BY created_time DESC LIMIT $2`,
		cid, N_MESSAGE_DEFAULT)
	if e != nil {
		return nil, errors.New("LoadConversation Messages:" + e.Error())
	}
	defer rows2.Close()
	for rows2.Next() {
		var message_id int64
		e = rows2.Scan(&message_id)
		if e != nil {
			return nil, e
		}
		msg, e := GetMessage(message_id)
		if e != nil {
			return nil, errors.New("LoadConversation LoadMessage:" + e.Error())
		}
		conversation.Messages = append([]*Message{msg}, conversation.Messages...)
	}

	rows3, e := zdatabase.DbPool.Query(
		`SELECT DISTINCT sender_id
	    FROM conversation_message WHERE conversation_id = $1`,
		cid)
	if e != nil {
		return nil, errors.New("LoadConversation LoadSenderIds:" + e.Error())
	}

	defer rows3.Close()
	for rows3.Next() {
		var sender_id int64
		e = rows3.Scan(&sender_id)
		if e != nil {
			return nil, e
		}
		conversation.SenderIds = append([]int64{sender_id}, conversation.SenderIds...)
	}
	//
	GMutex.Lock()
	MapConversations[cid] = conversation
	GMutex.Unlock()
	return conversation, nil
}

// try to read data in ram,
// if cant: read data from database
func GetConversation(id int64) (*Conversation, error) {
	GMutex.Lock()
	c := MapConversations[id]
	GMutex.Unlock()
	if c != nil {
		return c, nil
	} else {
		return LoadConversation(id, N_MESSAGE_DEFAULT)
	}
}

// return 0 if does not exist
func LoadConversationPairId(uid1 int64, uid2 int64) int64 {
	var pairKey sql.NullString
	pairKey.Scan(fmt.Sprintf("%v_%v", uid1, uid2))
	row := zdatabase.DbPool.QueryRow(
		`SELECT id FROM conversation WHERE pair_key = $1`,
		pairKey)
	var cid int64
	row.Scan(&cid)
	return cid
}

// read data from database to MapMessages
func LoadMessage(mid int64) (*Message, error) {
	row := zdatabase.DbPool.QueryRow(
		`SELECT conversation_id, sender_id, message_content,
    		created_time, display_type, gift_id
        FROM conversation_message
        WHERE message_id = $1`,
		mid,
	)
	var conversation_id, sender_id int64
	var gift_id sql.NullInt64
	var message_content, display_type string
	var created_time time.Time
	e := row.Scan(&conversation_id, &sender_id, &message_content,
		&created_time, &display_type, &gift_id)
	if e != nil {
		return nil, errors.New("LoadMessage:" + e.Error())
	}

	var gift Gift
	
	if gift_id.Valid {
		for _, tmp := range GiftList {
			if tmp.Id == gift_id.Int64 {
				gift = tmp
				break
			}
		}
	}

	msg := &Message{
		MessageId: mid, ConversationId: conversation_id, SenderId: sender_id,
		MessageContent: message_content, CreatedTime: created_time,
		DisplayType: display_type, Recipients: make(map[int64]*Recipient),
	}
	msg.Gift = gift

	rows3, e := zdatabase.DbPool.Query(
		`SELECT recipient_id, has_seen, seen_time
		FROM conversation_message_recipient
        WHERE message_id = $1 `,
		mid)
	if e != nil {
		return nil, errors.New("LoadMessage Recipients:" + e.Error())
	}
	defer rows3.Close()
	for rows3.Next() {
		var recipient_id int64
		var has_seen bool
		var seen_time time.Time
		e = rows3.Scan(&recipient_id, &has_seen, &seen_time)
		msg.Mutex.Lock()
		msg.Recipients[recipient_id] =
			&Recipient{MessageId: mid, RecipientId: recipient_id,
				HasSeen: has_seen, SeenTime: seen_time}
		msg.Mutex.Unlock()
	}
	//
	GMutex.Lock()
	MapMessages[mid] = msg
	GMutex.Unlock()
	return msg, nil
}

// try to read data in ram,
// if cant: read data from database
func GetMessage(id int64) (*Message, error) {
	GMutex.Lock()
	m := MapMessages[id]
	GMutex.Unlock()
	if m != nil {
		return m, nil
	} else {
		return LoadMessage(id)
	}
}

// return list conversationIds for user,
// filter: FILTER_ALL, FILTER_UNREAD, FILTER_PAIR
// sort by last message
func UserLoadConversationIds(
	userId int64, filter string, nConversation int) (
	[]int64, error) {
	rows, e := zdatabase.DbPool.Query(
		`SELECT conversation.id, conversation.conversation_type,
        	MAX(conversation_message.created_time) AS last_msg_time
        FROM conversation
            JOIN conversation_member
                ON conversation.id = conversation_member.conversation_id
            JOIN conversation_message
                ON conversation.id = conversation_message.conversation_id
        WHERE user_id = $1
        GROUP BY conversation.id
        ORDER BY last_msg_time DESC limit $2`,
		userId, nConversation)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	conversationIds := make([]int64, 0)
	for rows.Next() {
		var cid int64
		var conversation_type string
		var max_created_time time.Time
		e := rows.Scan(&cid, &conversation_type, &max_created_time)
		if e != nil {
			return nil, e
		}
		conversationIds = append(conversationIds, cid)
	}
	return conversationIds, nil
}

// return conversations in short form for user,
// filter: FILTER_ALL, FILTER_UNREAD, FILTER_PAIR
// sort by last message
func UserLoadAllConversations(userId int64, filter string, nConversation int) (
	[]map[string]interface{}, error) {
	cids, e := UserLoadConversationIds(userId, filter, nConversation)
	if e != nil {
		return nil, e
	}
	result := make([]map[string]interface{}, 0)
	for _, cid := range cids {
		conv, _ := GetConversation(cid)
		if conv != nil {
			result = append(result, conv.ToShortMap(userId))
		}
	}
	return result, nil
}

// return conversationId, error
func CreateConversation(
	memberIds []int64, moderatorIds []int64, conversationType string) (
	int64, error) {
	memberIds = misc.SortedInt64s(memberIds)
	names := []string{}
	for _, uid := range memberIds {
		name, e := users.GetProfilenameById(uid)
		if e != nil {
			return 0, e
		}
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
	if conversation == nil {
		return e
	}
	conversation.Mutex.Lock()
	sender := conversation.Members[senderId]
	conversation.Mutex.Unlock()
	senderU, _ := users.GetUser(senderId)
	if senderU == nil {
		return errors.New(l.Get(l.M022InvalidUserId))
	}
	if sender == nil && senderU.Role != users.ROLE_ADMIN {
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
	conversation.Messages = append(conversation.Messages, msg)
	hasSenderId := false
	for _, tmp := range conversation.SenderIds {
		if senderId == tmp {
			hasSenderId = true
			break
		}
	}
	conversation.Mutex.Lock()
	if !hasSenderId {
		conversation.SenderIds = append(conversation.SenderIds, senderId)
	}
	conversation.Mutex.Unlock()
	//
	conversation.Mutex.Lock()
	for _, member := range conversation.Members {
		if !member.IsMute {
			nbackend.WriteMapToUserId(member.UserId, nil,
				map[string]interface{}{
					"Command":    COMMAND_NEW_MESSAGE,
					"NewMessage": msg.ToMap(),
				})
		}
	}
	conversation.Mutex.Unlock()
	return nil
}

func CreateCheerMessage(
	conversationId int64, senderId int64, messageContent string,
	giftId int64, displayType string) error {
	fmt.Println("CHEER GIFT MESS GIFT_ID: ", giftId)
	conversation, e := GetConversation(conversationId)
	if conversation == nil {
		return e
	}
	conversation.Mutex.Lock()
	sender := conversation.Members[senderId]
	conversation.Mutex.Unlock()
	senderU, _ := users.GetUser(senderId)
	if senderU == nil {
		return errors.New(l.Get(l.M022InvalidUserId))
	}
	if sender == nil && senderU.Role != users.ROLE_ADMIN {
		return errors.New(l.Get(l.M003ConversationOutsider))
	}
	if sender.IsBlocked {
		return errors.New(l.Get(l.M004ConversationBlockedMember))
	}
	//
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO conversation_message
	        (conversation_id, sender_id, message_content, gift_id, display_type)
	    VALUES ($1, $2, $3, $4, $5) RETURNING message_id`,
		conversationId, senderId, messageContent, giftId ,displayType)
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
			nbackend.WriteMapToUserId(member.UserId, nil,
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
			nbackend.WriteMapToUserId(member.UserId, nil,
				map[string]interface{}{
					"Command":        COMMAND_SEEN_MESSAGE,
					"ChangedMessage": message.ToMap(),
				})
		}
	}
	return e
}
