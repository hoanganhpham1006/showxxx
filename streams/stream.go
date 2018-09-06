package streams

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/daominah/livestream/conversations"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/nbackend"
	//	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/zdatabase"
)

const (
	COMMAND_NEW_VIEWER      = "COMMAND_NEW_VIEWER"
	COMMAND_STREAM_FINISHED = "COMMAND_STREAM_FINISHED"
	// if a viewer joined for a duration that longer than this const:
	// he will be suggested as a relay for later joins
	GOOD_RELAY_DURATION = 30 * time.Second
)

var MapUserIdToStream = make(map[int64]*Stream)

// for read/write MapUserIdToStream, stream.ViewerIds, stream.MapUidToReport
var GMutex sync.Mutex

type Report struct {
	UserId int64
	Reason string
}

type Stream struct {
	BroadcasterId           int64
	StreamName              string
	StreamImage             string
	StartedTime             time.Time
	FinishedTime            time.Time
	ViewerIds               map[int64]bool
	MapUidToReport          map[int64]*Report
	ConversationId          int64
	MapViewerIdToJoinedTime map[int64]time.Time `json:"-"`
	RelayUserId             int64
	Password                string `json:"-"`
}

func (u *Stream) String() string {
	GMutex.Lock()
	defer GMutex.Unlock()
	u.RelayUserId = calcBestRelayUserId(u.MapViewerIdToJoinedTime)
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

// need Mutex.Lock() before call this func
func (u *Stream) writeMapToAllViewer(err error, data map[string]interface{}) {
	for uid, _ := range u.ViewerIds {
		nbackend.WriteMapToUserId(uid, err, data)
	}
}

// return turn the lastest viewer who have joined longer than GOOD_RELAY_DURATION,
// return 0 means we dont have a good relay.
// need to embracing in locker
func calcBestRelayUserId(mapViewerIdToJoinedTime map[int64]time.Time) int64 {
	now := time.Now()
	bestUid := int64(0)
	bestJoinedTime := now.Add(-86400 * time.Second)
	for uid, joinedTime := range mapViewerIdToJoinedTime {
		if joinedTime.After(bestJoinedTime) &&
			joinedTime.Before(now.Add(-GOOD_RELAY_DURATION)) {
			bestUid = uid
			bestJoinedTime = joinedTime
		}
	}
	return bestUid
}

func GetStream(broadcasterId int64) (*Stream, error) {
	GMutex.Lock()
	stream := MapUserIdToStream[broadcasterId]
	GMutex.Unlock()
	if stream == nil {
		return nil, errors.New(l.Get(l.M028StreamNotBroadcasting))
	}
	return stream, nil
}

func CreateStream(userId int64, streamName string, streamImage string, passwd string) (
	*Stream, error) {
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
		BroadcasterId:           userId,
		StreamName:              streamName,
		StreamImage:             streamImage,
		StartedTime:             time.Now(),
		FinishedTime:            zconfig.DefaultFutureTime,
		ViewerIds:               map[int64]bool{userId: true},
		MapUidToReport:          make(map[int64]*Report),
		ConversationId:          conversationId,
		MapViewerIdToJoinedTime: make(map[int64]time.Time),
		Password:                passwd,
	}
	GMutex.Lock()
	MapUserIdToStream[userId] = stream
	GMutex.Unlock()
	user.StatusL1 = users.STATUS_BROADCASTING
	return stream, nil
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
	for uid, _ := range stream.ViewerIds {
		user, _ := users.GetUser(uid)
		if user != nil {
			user.StatusL1 = users.STATUS_ONLINE
		}
	}
	delete(MapUserIdToStream, broadcasterId)
	nViewers := len(stream.ViewerIds)
	nReports := len(stream.MapUidToReport)
	viewersB, _ := json.Marshal(stream.ViewerIds)
	reportsB, _ := json.Marshal(stream.MapUidToReport)
	go func() {
		_, e := zdatabase.DbPool.Exec(
			`INSERT INTO stream_archive
    			(broadcaster_id, started_time, finished_time,
    			n_viewers, n_reports, viewers, reports, conversation_id, 
    			stream_name, stream_image)
        	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			stream.BroadcasterId, stream.StartedTime, stream.FinishedTime,
			nViewers, nReports, string(viewersB), string(reportsB),
			stream.ConversationId, stream.StreamName, stream.StreamImage,
		)
		_ = e
	}()
	return nil
}

func ViewStream(viewerId int64, broadcasterId int64, passwd string) (*Stream, error) {
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
		if stream.ViewerIds[viewerId] {
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
	if passwd != targetStream.Password {
		return nil, errors.New(l.Get(l.M029StreamConcurrentView))
	}
	//
	GMutex.Lock()
	targetStream.ViewerIds[viewerId] = true
	targetStream.MapViewerIdToJoinedTime[viewerId] = time.Now()
	targetStream.writeMapToAllViewer(nil, map[string]interface{}{
		"Command":     COMMAND_NEW_VIEWER,
		"NewViewerId": viewerId,
	})
	GMutex.Unlock()
	viewer.StatusL1 = users.STATUS_WATCHING
	viewer.StatusL2 = fmt.Sprintf(`{"BroadcasterId": %v}`, broadcasterId)
	conversations.AddMember(targetStream.ConversationId, viewerId, false)
	return targetStream, nil
}

// StopViewingStream is FinishStream if viewerId is BroadcasterId
func StopViewingStream(viewerId int64) error {
	viewer, e := users.GetUser(viewerId)
	if viewer == nil {
		return e
	}
	_, targetStream := GetViewingStream(viewerId)
	if targetStream == nil {
		return errors.New(l.Get(l.M028StreamNotBroadcasting))
	}
	//
	GMutex.Lock()
	delete(targetStream.ViewerIds, viewerId)
	delete(targetStream.MapViewerIdToJoinedTime, viewerId)
	GMutex.Unlock()
	viewer.StatusL1 = users.STATUS_ONLINE
	viewer.StatusL2 = "{}"
	conversations.RemoveMember(targetStream.ConversationId, viewerId)
	if viewerId == targetStream.BroadcasterId {
		FinishStream(targetStream.BroadcasterId)
	}
	return nil
}

func ReportStream(viewerId int64, broadcasterId int64, reason string) error {
	GMutex.Lock()
	defer GMutex.Unlock()
	stream := MapUserIdToStream[broadcasterId]
	if stream == nil {
		return errors.New(l.Get(l.M028StreamNotBroadcasting))
	}
	if !stream.ViewerIds[viewerId] {
		return errors.New(l.Get(l.M032StreamOnlyViewerCanReport))
	}
	stream.MapUidToReport[viewerId] = &Report{UserId: viewerId, Reason: reason}
	return nil
}

func GetViewingStream(viewerId int64) (int64, *Stream) {
	viewingStreamId := int64(0)
	var viewingStream *Stream
	GMutex.Lock()
	for _, stream := range MapUserIdToStream {
		if stream.ViewerIds[viewerId] {
			viewingStreamId = stream.BroadcasterId
		}
		if viewingStreamId != 0 {
			viewingStream = stream
			break
		}
	}
	GMutex.Unlock()
	return viewingStreamId, viewingStream
}

type StreamNViewersOrder []*Stream

func (a StreamNViewersOrder) Len() int { return len(a) }
func (a StreamNViewersOrder) Less(i int, j int) bool {
	return len(a[i].ViewerIds) > len(a[j].ViewerIds)
}
func (a StreamNViewersOrder) Swap(i int, j int) { a[i], a[j] = a[j], a[i] }

// if filterReported == true: return only reported streams
func StreamAllSummaries(filterReported bool) []map[string]interface{} {
	GMutex.Lock()
	result := make([]map[string]interface{}, 0)
	temp := make([]*Stream, 0)
	for _, stream := range MapUserIdToStream {
		if !filterReported {
			temp = append(temp, stream)
		} else {
			if len(stream.MapUidToReport) > 0 {
				temp = append(temp, stream)
			}
		}
	}
	sort.Sort(StreamNViewersOrder(temp))
	for _, stream := range temp {
		var temp1 map[string]interface{}
		b, _ := users.GetUser(stream.BroadcasterId)
		if b != nil {
			temp1 = b.ToMap()
		}
		stream.RelayUserId = calcBestRelayUserId(stream.MapViewerIdToJoinedTime)
		result = append(result, map[string]interface{}{
			"BroadcasterId":     stream.BroadcasterId,
			"BroadcasterDetail": temp1,
			"NViewers":          len(stream.ViewerIds),
			"StreamName":        stream.StreamName,
			"StreamImage":       stream.StreamImage,
			"RelayUserId":       stream.RelayUserId,
		})
	}
	GMutex.Unlock()
	return result
}
