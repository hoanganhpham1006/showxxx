package connections

import (
	"encoding/json"
	//	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/daominah/livestream/zconfig"
)

// map userId to connection
var MapConnection map[int64]*Connection
var GMutex sync.Mutex

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func init() {
	_ = ioutil.ReadAll
}

func tPrint(a ...interface{}) {
	if zconfig.IsDeveloping {
		fmt.Println(a...)
	}
}

func tPrintf(format string, a ...interface{}) {
	if zconfig.IsDeveloping {
		fmt.Printf(format, a...)
	}
}

// ListenAndServe listens on a tcp port and upgrate connections to websocket,
// already run in a goroutine
func ListenAndServe(
	serverCommandHandler func(connection *Connection, message []byte)) {
	go func() {
		fmt.Printf("Listening http message on address host%v/ws\n",
			zconfig.WebsocketPort)
		err := http.ListenAndServe(zconfig.WebsocketPort, nil)
		if err != nil {
			fmt.Printf("Fail to listen http message on address host%v/ws\n %v\n",
				zconfig.WebsocketPort, err.Error())
		}
	}()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		//		tPrintf("http connect header: %#v\n", r.Header)
		//		body, _ := ioutil.ReadAll(r.Body)
		//		tPrintf("http connect body: %#v\n", body)
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrader.Upgrade", err.Error())
			return
		}
		c := CreateConnection(ws)
		go c.readPump(serverCommandHandler)
		go c.writePump()
		tPrint("Connection created: ", c)
	})
}

// help init channels
func CreateConnection(wsConn *websocket.Conn) *Connection {
	c := &Connection{UserId: 0, WsConn: wsConn,
		ChanWrite: make(chan []byte), ChanClose: make(chan bool)}
	return c
}

// this app connection type, wrap websocket.Conn
type Connection struct {
	WsConn *websocket.Conn
	// authenticated connection has UserId != 0
	UserId int64
	// library gorilla/websocket only allows
	// that no more than one goroutine calls the write method
	ChanWrite chan []byte
	ChanClose chan bool
}

// messageHandler is what to do after receive message from peer,
// messageHandler == nil: just print message
func (c *Connection) readPump(messageHandler func(*Connection, []byte)) {
	defer tPrint("Connection readPump ended", c)
	c.WsConn.SetReadLimit(zconfig.WebsocketMaxMessageSize)
	c.WsConn.SetReadDeadline(time.Now().Add(zconfig.WebsocketReadWait))
	c.WsConn.SetPongHandler(func(string) error {
		c.WsConn.SetReadDeadline(time.Now().Add(zconfig.WebsocketReadWait))
		return nil
	})
	for {
		messageType, message, err := c.WsConn.ReadMessage()
		_ = messageType //
		if err != nil {
			tPrint("WsConn.ReadMessage err", err)
			c.WsConn.Close()
			return
		} else {
			tPrintf("Connection readPump message %v:\n%v\n",
				time.Now(), string(message))
		}
		if messageHandler != nil {
			go messageHandler(c, message)
		}
	}
}

// ensuring no more than one goroutine calls the write method for the connection,
// so you send message to connection.ChanWrite
//   instead of directly call connection.WsConn.WriteMessage
func (c *Connection) writePump() {
	defer tPrint("Connection writePump ended", c)
	ticker := time.NewTicker(zconfig.WebsocketPingPeriod)
	defer func() { ticker.Stop() }()
	for {
		var writeErr error
		select {
		case message := <-c.ChanWrite:
			c.WsConn.SetWriteDeadline(time.Now().Add(zconfig.WebsocketWriteWait))
			writeErr = c.WsConn.WriteMessage(websocket.TextMessage, message)
			if writeErr == nil {
				tPrintf("Connection writePump message %v:\n%v\n",
					time.Now(), string(message))
			}
		case <-ticker.C:
			c.WsConn.SetWriteDeadline(time.Now().Add(zconfig.WebsocketWriteWait))
			writeErr = c.WsConn.WriteMessage(websocket.PingMessage, nil)
		case <-c.ChanClose:
			c.WsConn.SetWriteDeadline(time.Now().Add(zconfig.WebsocketWriteWait))
			writeErr = c.WsConn.WriteMessage(websocket.CloseMessage, nil)
		}
		if writeErr != nil {
			c.WsConn.Close()
			tPrint("WsConn.WriteMessage err", writeErr)
			return
		}
	}
}

// send close control message
func (c *Connection) Close() {
	timeout := time.After(1 * time.Second)
	select {
	case c.ChanClose <- true:
	case <-timeout:
	}
}

// run a goroutine to send the message to peer
func (c *Connection) Write(message []byte) {
	go func() {
		timeout := time.After(1 * time.Second)
		select {
		case c.ChanWrite <- message:
		case <-timeout:
		}
	}()
}

// WriteMap jsonDump (data+err),
// then run a goroutine to send the jsonData to peer,
// map data must not have field "Error",
// this func usually is used for reply a command from peer
func (c *Connection) WriteMap(err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err == nil {
		data["Error"] = ""
	} else {
		data["Error"] = err.Error()
	}
	message, _ := json.Marshal(data)
	go func() {
		timeout := time.After(1 * time.Second)
		select {
		case c.ChanWrite <- message:
		case <-timeout:
		}
	}()
}

// WriteMapToUserId jsonDump (data+err),
// then run a goroutine to send the jsonData to user,
// map data must not have field "Error",
// this func usually is used for reply a command from user
func WriteMapToUserId(userId int64, err error, data map[string]interface{}) {
	GMutex.Lock()
	conn := MapConnection[userId]
	GMutex.Unlock()
	if conn != nil {
		conn.WriteMap(err, data)
	}
}
