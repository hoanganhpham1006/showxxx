// this package defines app connection type, wrap websocket.Conn
package nwebsocket

import (
	"encoding/json"
	//	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/daominah/livestream/zconfig"
)

const (
	BLANK_LINE = "________________________________________________________________________________\n"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// protect ConnGUID
var GMutex sync.Mutex
var ConnGUID = int64(0)

type Server struct {
	// limit number of connections per ip address
	LimitNConnsPerIp int
	// limit number of request per second of a connection
	LimitNRequestsPerSecond int
	Mutex                   sync.Mutex
	// map connection.ConnId to it
	MapIdToConnection map[int64]*Connection
	MapIpToNConns     map[string]int
}

// init map fields,
// call server.ListenAndServe after this func
func CreateServer(LimitNConnsPerIp int, LimitNRequestsPerSecond int) *Server {
	server := &Server{
		LimitNConnsPerIp:        LimitNConnsPerIp,
		LimitNRequestsPerSecond: LimitNRequestsPerSecond,
		MapIdToConnection:       make(map[int64]*Connection),
		MapIpToNConns:           make(map[string]int),
	}
	return server
}

// ListenAndServe listens on a tcp port and upgrate connections to websocket,
// already run in a goroutine,
func (server *Server) ListenAndServe(
	port string,
	doAfterReceivingMessage func(connection *Connection, message []byte),
	doAfterClosingConnection func(connection *Connection),
) {
	go func() {
		fmt.Printf("Listening websocket on address host%v/ws\n", port)
		err := http.ListenAndServe(port, nil)
		if err != nil {
			fmt.Printf("Fail to listen websocket on address host%v/ws\n %v\n",
				port, err.Error())
		}
	}()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		_ = ioutil.ReadAll
		//		tPrintf("http connect header: %#v\n", r.Header)
		//		body, _ := ioutil.ReadAll(r.Body)
		//		tPrintf("http connect body: %#v\n", body)
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrader.Upgrade", err.Error())
			return
		}
		c := CreateConnection(ws, server.LimitNRequestsPerSecond)
		server.Mutex.Lock()
		server.MapIdToConnection[c.ConnId] = c
		server.MapIpToNConns[c.Ip()] += 1
		server.Mutex.Unlock()
		go c.ReadPump(doAfterReceivingMessage, doAfterClosingConnection)
		go c.WritePump(doAfterClosingConnection)
		zconfig.TPrint("Connection created: ", c.WsConn.RemoteAddr(), c)
	})
}

// default doAfterClosingConnection
func (server *Server) CleanDisconnection(connection *Connection) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()
	delete(server.MapIdToConnection, connection.ConnId)
	server.MapIpToNConns[connection.Ip()] -= 1
	if server.MapIpToNConns[connection.Ip()] == 0 {
		delete(server.MapIpToNConns, connection.Ip())
	}
}

// help init read/write channels,
// limitNRequest: limit number of requests received from peer in a second
func CreateConnection(wsConn *websocket.Conn, limitNRequest int) *Connection {
	GMutex.Lock()
	ConnGUID += 1
	c := &Connection{ConnId: ConnGUID, UserId: 0, WsConn: wsConn,
		NRequestLastSecond: 0, LimitNRequest: limitNRequest,
		ChanWrite: make(chan []byte), ChanClose: make(chan []byte)}
	GMutex.Unlock()
	go func() {
		for !c.HasLoopsEnded {
			time.Sleep(1 * time.Second)
			// fmt.Printf("Loop ResetNRequestLastSecond ConnId %v NRequest %v\n", c.ConnId, c.NRequestLastSecond)
			c.NRequestLastSecond = 0
		}
	}()
	return c
}

// this app connection type, wrap websocket.Conn
type Connection struct {
	ConnId int64
	WsConn *websocket.Conn `json:"-"`
	// authenticated connection has UserId != 0
	UserId int64
	// for user's login history
	LoginId int64
	// library gorilla/websocket only allows
	// that no more than one goroutine calls the write method

	// number of requests received from peer in a second
	NRequestLastSecond int
	// limit NRequestLastSecond
	LimitNRequest int
	// a connection has 3 loops: ReadPump, WritePump, ResetNRequestLastSecond,
	// if anyone of them ended, other should end too
	HasLoopsEnded bool
	// connection can be closed by ReadPump or WritePump,
	// this field ensures the doAfterClosingConnection can only execute only once.
	HasHandledClosing bool

	ChanWrite chan []byte `json:"-"`
	ChanClose chan []byte `json:"-"`
}

func (c *Connection) String() string {
	bs, e := json.MarshalIndent(c, "", "    ")
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (c *Connection) Ip() string {
	if c.WsConn == nil {
		return "c.WsConn == nil"
	}
	addr := c.WsConn.RemoteAddr().String()
	temp := strings.Index(addr, ":")
	if temp == -1 {
		return `strings.Index(addr, ":") == -1`
	}
	ip := addr[0:temp]
	return ip
}

// doAfterClosingConnection: change MapConnection, update user online record;
// doAfterReceivingMessage: execute and respond to peer's command
func (c *Connection) ReadPump(
	doAfterReceivingMessage func(*Connection, []byte),
	doAfterClosingConnection func(*Connection),
) {
	defer zconfig.TPrint("Connection readPump ended",
		c.WsConn.LocalAddr(), c.WsConn.RemoteAddr(), c)
	defer func() { c.HasLoopsEnded = true }()
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
			zconfig.TPrint("WsConn.ReadMessage err", err)
			c.WsConn.Close()
			if doAfterClosingConnection != nil && c.HasHandledClosing == false {
				c.HasHandledClosing = true
				doAfterClosingConnection(c)
			}
			return
		} else {
			zconfig.TPrintf(
				BLANK_LINE+"Connection readPump %v local %v remote %v:\n%v\n",
				time.Now(), c.WsConn.LocalAddr(), c.WsConn.RemoteAddr(), string(message))
			c.NRequestLastSecond += 1
			if c.NRequestLastSecond > c.LimitNRequest {
				c.Close("Exceed LimitNRequest")
			}
		}
		if doAfterReceivingMessage != nil {
			go doAfterReceivingMessage(c, message)
		}
	}
}

// doAfterClosingConnection: change MapConnection, update  user online record
func (c *Connection) WritePump(doAfterClosingConnection func(*Connection)) {
	defer zconfig.TPrint("Connection writePump ended",
		c.WsConn.LocalAddr(), c.WsConn.RemoteAddr(), c)
	ticker := time.NewTicker(zconfig.WebsocketPingPeriod)
	defer func() { ticker.Stop() }()
	defer func() { c.HasLoopsEnded = true }()
	for {
		var writeErr error
		var msg []byte
		var caseName string
		select {
		case msg = <-c.ChanWrite:
			c.WsConn.SetWriteDeadline(time.Now().Add(zconfig.WebsocketWriteWait))
			writeErr = c.WsConn.WriteMessage(websocket.TextMessage, msg)
			if writeErr == nil {
				zconfig.TPrintf(
					BLANK_LINE+"Connection writePump %v local %v remote %v:\n%v\n",
					time.Now(), c.WsConn.LocalAddr(), c.WsConn.RemoteAddr(), string(msg))
			}
			caseName = "0"
		case <-ticker.C:
			c.WsConn.SetWriteDeadline(time.Now().Add(zconfig.WebsocketWriteWait))
			writeErr = c.WsConn.WriteMessage(websocket.PingMessage, nil)
			caseName = "1"
		case msg = <-c.ChanClose:
			c.WsConn.SetWriteDeadline(time.Now().Add(zconfig.WebsocketWriteWait))
			writeErr = c.WsConn.WriteMessage(websocket.CloseMessage, msg)
			caseName = "2"
		}
		if writeErr != nil {
			c.WsConn.Close()
			if doAfterClosingConnection != nil && c.HasHandledClosing == false {
				c.HasHandledClosing = true
				doAfterClosingConnection(c)
			}
			zconfig.TPrintf("WsConn.WriteMessage writeErr %v msg %v caseName %v\n",
				writeErr, string(msg), caseName)
			return
		}
	}
}

// send close control message
func (c *Connection) Close(reason string) {
	zconfig.TPrint(time.Now(), "Manual Close Connection", c.Ip())
	timeout := time.After(1 * time.Second)
	select {
	case c.ChanClose <- []byte(reason):
	case <-timeout:
	}
}

// run a goroutine to send the message to peer
func (c *Connection) Write(message []byte) {
	go func(c *Connection) {
		timeout := time.After(1 * time.Second)
		select {
		case c.ChanWrite <- message:
		case <-timeout:
			fmt.Println("Write timeout", time.Now(), string(message))
		}
	}(c)
}

// WriteMap jsonDump (data+err),
// then run a goroutine to send the jsonData to peer,
// map data must not have field "Error",
// this func usually is used for reply a command from peer
func (c *Connection) WriteMap(err error, data map[string]interface{}) {
	message := MapToBytes(err, data)
	c.Write(message)
}

//
func MapToBytes(err error, data map[string]interface{}) []byte {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err == nil {
		data["Error"] = ""
	} else {
		data["Error"] = err.Error()
	}
	message, _ := json.Marshal(data)
	return message
}
