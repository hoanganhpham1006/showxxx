package conversations

import (
	"database/sql"
	//	"encoding/json"
	"errors"
	"fmt"
	//	"strings"
	//	"sync"
	"time"

	l "github.com/daominah/livestream/language"
	//	"github.com/daominah/livestream/misc"
	//	"github.com/daominah/livestream/user"
	"github.com/daominah/livestream/zdatabase"
	//	"github.com/lib/pq"
)

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
		`SELECT conversation_id, sender_id,message_content,created_time 
        FROM conversation_message
        WHERE message_id = $1`,
		mid,
	)
	var conversation_id, sender_id int64
	var message_content string
	var created_time time.Time
	e := row.Scan(&conversation_id, &sender_id, &message_content, &created_time)
	if e != nil {
		return nil, errors.New("LoadMessage:" + e.Error())
	}
	msg := &Message{
		MessageId: mid, ConversationId: conversation_id, SenderId: sender_id,
		MessageContent: message_content, CreatedTime: created_time,
		Recipients: make(map[int64]*Recipient),
	}
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
		msg.Recipients[recipient_id] =
			&Recipient{MessageId: mid, RecipientId: recipient_id,
				HasSeen: has_seen, SeenTime: seen_time}
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
			result = append(result, conv.ToShortMap())
		}
	}
	return result, nil
}
