package streams

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"

	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/zconfig"
)

var MapUidToSockio = make(map[int64]*gosocketio.Channel)
var MapSockIpToUid = make(map[string]int64)

//
func toString(err error, data map[string]interface{}) string {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err == nil {
		data["Error"] = ""
	} else {
		data["Error"] = err.Error()
	}
	messageB, _ := json.Marshal(data)
	return string(messageB)
}

func ForwarderListenAndServer() {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		fmt.Println("connection", c.Ip())
		c.Ack("guessid", misc.HashStringToInt64(c.Id()), 5*time.Second)
	})
	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		fmt.Println("disconnection", c.Ip())
		GMutex.Lock()
		uid := MapSockIpToUid[c.Ip()]
		delete(MapSockIpToUid, c.Ip())
		delete(MapUidToSockio, uid)
		GMutex.Unlock()
		err := StopViewingStream(uid)
		_ = err
	})
	server.On("signal", func(c *gosocketio.Channel, dataS string) string {
		fmt.Println("signal", dataS)
		return ""
	})
	server.On("join", func(c *gosocketio.Channel, dataS string) string {
		fmt.Println("join", dataS)
		var data map[string]interface{}
		err := json.Unmarshal([]byte(dataS), &data)
		if err != nil {
			return toString(err, nil)
		}
		broadcasterIdS := misc.ReadString(data, "BroadcasterId")
		broadcasterId, _ := strconv.ParseInt(broadcasterIdS, 10, 64)
		isCreatingStream := misc.ReadBool(data, "IsCreatingStream")
		viewerIdS := misc.ReadString(data, "ViewerId")
		viewerId, _ := strconv.ParseInt(viewerIdS, 10, 64)
		streamName := misc.ReadString(data, "StreamName")
		streamImage := misc.ReadString(data, "StreamImage")
		passwd := misc.ReadString(data, "Password")
		var stream *Stream
		if isCreatingStream {
			stream, err = CreateStream(broadcasterId, streamName, streamImage, passwd)
		} else {
			stream, err = ViewStream(viewerId, broadcasterId, passwd)
		}
		if stream != nil {
			GMutex.Lock()
			MapUidToSockio[viewerId] = c
			MapSockIpToUid[c.Ip()] = viewerId
			GMutex.Unlock()
			return toString(err, stream.ToMap())
		} else {
			return toString(err, nil)
		}
	})
	server.On("exchange", func(c *gosocketio.Channel, dataS string) string {
		fmt.Println("exchange", dataS)
		var data map[string]interface{}
		err := json.Unmarshal([]byte(dataS), &data)
		if err != nil {
			return toString(err, nil)
		}
		recipientIdS := misc.ReadString(data, "to")
		recipientId, _ := strconv.ParseInt(recipientIdS, 10, 64)
		GMutex.Lock()
		conn := MapUidToSockio[recipientId]
		GMutex.Unlock()
		if conn != nil {
			go func(conn *gosocketio.Channel) {
				conn.Ack("exchange", dataS, 5*time.Second)
			}(conn)
		}
		return ""
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)
	go func() {
		fmt.Printf("Listening socketIo on address host%v/socket.io/\n",
			zconfig.WebRTCSignalingPort)
		err := http.ListenAndServe(zconfig.WebRTCSignalingPort, serveMux)
		if err != nil {
			fmt.Printf("Fail to listen socketIo on address host%v/socket.io/\n %v\n",
				zconfig.WebRTCSignalingPort, err.Error())
		}
	}()
}
