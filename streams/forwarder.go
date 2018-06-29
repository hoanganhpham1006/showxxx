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

var MapUidToConnection = make(map[int64]*gosocketio.Channel)

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
	})
	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		fmt.Println("disconnection", c.Ip())
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
		var stream *Stream
		if isCreatingStream {
			stream, err = CreateStream(broadcasterId)
		} else {
			stream, err = ViewStream(viewerId, broadcasterId)
		}
		if stream != nil {
			GMutex.Lock()
			MapUidToConnection[viewerId] = c
			GMutex.Unlock()
			return toString(err, stream.ToMap())
		} else {
			return toString(err, nil)
		}
	})
	server.On("exchange", func(c *gosocketio.Channel, dataS string) string {
		var data map[string]interface{}
		err := json.Unmarshal([]byte(dataS), &data)
		if err != nil {
			return toString(err, nil)
		}
		recipientIdS := misc.ReadString(data, "to")
		recipientId, _ := strconv.ParseInt(recipientIdS, 10, 64)
		GMutex.Lock()
		conn := MapUidToConnection[recipientId]
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
			zconfig.SocketIoPort)
		err := http.ListenAndServe(zconfig.SocketIoPort, serveMux)
		if err != nil {
			fmt.Printf("Fail to listen socketIo on address host%v/socket.io/\n %v\n",
				zconfig.SocketIoPort, err.Error())
		}
	}()
}
