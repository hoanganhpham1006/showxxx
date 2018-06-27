package streams

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/daominah/livestream/connections"
	"github.com/daominah/livestream/conversations"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zconfig"
)

const (
	COMMAND_NEW_VIEWER = "COMMAND_NEW_VIEWER"
)

var MapUserIdToStream = make(map[int64]*Stream)
var GMutex sync.Mutex

type Stream struct {
	BroadcasterId  int64
	StartedTime    time.Time
	FinishedTime   time.Time
	ViewerIds      []int64
	ConversationId int64
	Mutex          sync.Mutex
}

func (u *Stream) ToString() string {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	bs, e := json.MarshalIndent(u, "", "    ")
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (u *Stream) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	s := u.ToString()
	json.Unmarshal([]byte(s), &result)
	return result
}

func (u *Stream) writeMapToAllViewer(err error, data map[string]interface{}) {
	u.Mutex.Lock()
	for _, uid := range u.ViewerIds {
		connections.WriteMapToUserId(uid, err, data)
	}
	u.Mutex.Unlock()
}

func CreateStream(userId int64) (*Stream, error) {
	user, e := users.GetUser(userId)
	if user == nil {
		return nil, errors.New(l.Get(l.M022InvalidUserId))
	}
	if user.Role != users.ROLE_BROADCASTER {
		return nil, errors.New(l.Get(l.M026StreamCreatePrivilege))
	}
	GMutex.Lock()
	oldStream := MapUserIdToStream[userId]
	GMutex.Unlock()
	if oldStream != nil {
		return nil, errors.New(l.Get(l.M027StreamBroadcasted))
	}
	conversationId, e := conversations.CreateConversation(
		[]int64{userId}, []int64{userId}, conversations.CONVERSATION_GROUP)
	if e != nil {
		return nil, e
	}
	stream := &Stream{
		BroadcasterId:  userId,
		StartedTime:    time.Now(),
		FinishedTime:   zconfig.DefaultFutureTime,
		ViewerIds:      make([]int64, 0),
		ConversationId: conversationId,
	}
	GMutex.Lock()
	MapUserIdToStream[userId] = stream
	GMutex.Unlock()
	user.StatusL1 = users.STATUS_BROADCASTING
	return stream, nil
}

func ViewStream(viewerId int64, broadcasterId int64) (*Stream, error) {
	GMutex.Lock()
	defer GMutex.Unlock()
	targetStream := MapUserIdToStream[broadcasterId]
	if targetStream == nil {
		return nil, errors.New(l.Get(l.M028StreamNotBroadcasting))
	}
	viewingStreamId := int64(0)
	for _, stream := range MapUserIdToStream {
		stream.Mutex.Lock()
		if misc.FindInt64InSlice(viewerId, stream.ViewerIds) != -1 {
			viewingStreamId = stream.BroadcasterId
		}
		stream.Mutex.Unlock()
		if viewingStreamId != 0 {
			break
		}
	}
	if viewingStreamId != 0 {
		return nil, fmt.Errorf("%v: %v", l.Get(l.M029StreamConcurrentView), viewingStreamId)
	}
	//
	targetStream.Mutex.Lock()
	targetStream.ViewerIds = append(targetStream.ViewerIds, viewerId)
	targetStream.Mutex.Unlock()
	conversations.AddMember(targetStream.ConversationId, viewerId, false)
	targetStream.writeMapToAllViewer(nil, map[string]interface{}{
		"Command":     COMMAND_NEW_VIEWER,
		"NewViewerId": viewerId,
	})
	return targetStream, nil
}
