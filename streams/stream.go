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
	"github.com/daominah/livestream/zdatabase"
)

const (
	COMMAND_NEW_VIEWER      = "COMMAND_NEW_VIEWER"
	COMMAND_STREAM_FINISHED = "COMMAND_STREAM_FINISHED"
)

var MapUserIdToStream = make(map[int64]*Stream)

// for read/write MapUserIdToStream, stream.ViewerIds, stream.MapUidToReport
var GMutex sync.Mutex

type Report struct {
	UserId int64
	Reason string
}

type Stream struct {
	BroadcasterId  int64
	StartedTime    time.Time
	FinishedTime   time.Time
	ViewerIds      []int64
	MapUidToReport map[int64]*Report
	ConversationId int64
}

func (u *Stream) String() string {
	GMutex.Lock()
	defer GMutex.Unlock()
	bs, e := json.MarshalIndent(u, "", "    ")
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (u *Stream) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	s := u.String()
	json.Unmarshal([]byte(s), &result)
	return result
}

func (u *Stream) writeMapToAllViewer(err error, data map[string]interface{}) {
	GMutex.Lock()
	for _, uid := range u.ViewerIds {
		connections.WriteMapToUserId(uid, err, data)
	}
	GMutex.Unlock()
}

func CreateStream(userId int64) (*Stream, error) {
	user, e := users.GetUser(userId)
	if user == nil {
		return nil, errors.New(l.Get(l.M022InvalidUserId))
	}
	//	if user.Role != users.ROLE_BROADCASTER {
	//		return nil, errors.New(l.Get(l.M026StreamCreatePrivilege))
	//	}
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
		ViewerIds:      []int64{userId},
		ConversationId: conversationId,
	}
	GMutex.Lock()
	MapUserIdToStream[userId] = stream
	GMutex.Unlock()
	user.StatusL1 = users.STATUS_BROADCASTING
	return stream, nil
}

func ViewStream(viewerId int64, broadcasterId int64) (*Stream, error) {
	viewer, e := users.GetUser(viewerId)
	if viewer == nil {
		return nil, e
	}
	GMutex.Lock()
	targetStream := MapUserIdToStream[broadcasterId]
	GMutex.Unlock()
	if targetStream == nil {
		return nil, errors.New(l.Get(l.M028StreamNotBroadcasting))
	}
	viewingStreamId := int64(0)
	var viewingStream *Stream
	GMutex.Lock()
	for _, stream := range MapUserIdToStream {
		if misc.FindInt64InSlice(viewerId, stream.ViewerIds) != -1 {
			viewingStreamId = stream.BroadcasterId
		}
		if viewingStreamId != 0 {
			viewingStream = stream
			break
		}
	}
	GMutex.Unlock()
	if viewingStreamId != 0 {
		return viewingStream, fmt.Errorf("%v: %v", l.Get(l.M029StreamConcurrentView))
	}
	//
	GMutex.Lock()
	targetStream.ViewerIds = append(targetStream.ViewerIds, viewerId)
	GMutex.Unlock()
	viewer.StatusL1 = users.STATUS_WATCHING
	viewer.StatusL2 = fmt.Sprintf(`{"BroadcasterId": %v}`, broadcasterId)
	conversations.AddMember(targetStream.ConversationId, viewerId, false)
	targetStream.writeMapToAllViewer(nil, map[string]interface{}{
		"Command":     COMMAND_NEW_VIEWER,
		"NewViewerId": viewerId,
	})
	return targetStream, nil
}

func FinishStream(broadcasterId int64) error {
	GMutex.Lock()
	defer GMutex.Unlock()
	stream := MapUserIdToStream[broadcasterId]
	if stream == nil {
		return errors.New(l.Get(l.M028StreamNotBroadcasting))
	}
	stream.writeMapToAllViewer(nil, map[string]interface{}{
		"Command":       COMMAND_STREAM_FINISHED,
		"BroadcasterId": stream.BroadcasterId,
	})
	stream.FinishedTime = time.Now()
	// TODO: users.MT_BROADCAST_DURATION
	delete(MapUserIdToStream, broadcasterId)
	nViewers := len(stream.ViewerIds)
	nReports := len(stream.MapUidToReport)
	viewersB, _ := json.Marshal(stream.ViewerIds)
	reportsB, _ := json.Marshal(stream.MapUidToReport)
	go func() {
		zdatabase.DbPool.Exec(
			`INSERT INTO stream_archive
    			(broadcaster_id, started_time, finished_time,
    			n_viewers, n_reports, viewers, reports, conversation_id)
        	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			stream.BroadcasterId, stream.StartedTime, stream.FinishedTime,
			nViewers, nReports, string(viewersB), string(reportsB),
			stream.ConversationId,
		)
	}()
	return nil
}
