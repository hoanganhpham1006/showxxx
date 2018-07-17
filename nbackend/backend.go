package nbackend

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/nwebsocket"
	"github.com/daominah/livestream/zconfig"
)

var GBackend *Backend

type Proxy struct {
	// clients connect to this
	ProxyId int64
	Server  *nwebsocket.Server
	Mutex   sync.Mutex
	// map connection.ConnId to connection
	MapBackendConn    map[int64]*nwebsocket.Connection
	MapUserIdToConnId map[int64]int64
}

func CreateProxy() *Proxy {
	proxy := &Proxy{
		ProxyId:           time.Now().UnixNano(),
		MapBackendConn:    make(map[int64]*nwebsocket.Connection),
		MapUserIdToConnId: make(map[int64]int64),
	}
	return proxy
}

// infinite loop maintains connections to backend
func (proxy *Proxy) ConnectToBackend() {
	serverAddr := fmt.Sprintf("ws://%v%v/ws",
		zconfig.BackendIp, zconfig.BackendPort)
	for {
		proxy.Mutex.Lock()
		if len(proxy.MapBackendConn) < 50 {
			for len(proxy.MapBackendConn) < 50 {
				wsConn, _, e := websocket.DefaultDialer.Dial(serverAddr, nil)
				if e == nil {
					conn := nwebsocket.CreateConnection(wsConn, 999999)
					proxy.MapBackendConn[conn.ConnId] = conn
					go conn.ReadPump(proxy.doAfterReceivingBackendMessage, proxy.doAfterClosingBackendConnection)
					go conn.WritePump(proxy.doAfterClosingBackendConnection)
					go conn.WriteMap(nil, map[string]interface{}{
						"Command": "ProxyConnect",
						"ProxyId": time.Now().UnixNano()})
				} else {
					fmt.Println("ConnectToBackend err", e)
					time.Sleep(1 * time.Second)
				}
			}
			fmt.Println(time.Now(), "Proxy connected to backend.")
		}
		proxy.Mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (proxy *Proxy) WriteToUserId(userId int64, message []byte) {
	proxy.Mutex.Lock()
	connId := proxy.MapUserIdToConnId[userId]
	proxy.Mutex.Unlock()
	if connId != 0 && proxy.Server != nil {
		proxy.Server.Mutex.Lock()
		conn := proxy.Server.MapIdToConnection[connId]
		proxy.Server.Mutex.Unlock()
		if conn != nil {
			conn.Write(message)
		}
	}
}

// this func get the connection correspond to the userId and send data to him,
// except command is UserCreate or UserLogin, need to do more
func (proxy *Proxy) doAfterReceivingBackendMessage(
	connection *nwebsocket.Connection, message []byte) {
	var data map[string]interface{}
	parseTextErr := json.Unmarshal(message, &data)
	if parseTextErr != nil {
		return
	}
	command := misc.ReadString(data, "Command")
	clientConnId := misc.ReadInt64(data, "ConnId")
	userId := misc.ReadInt64(data, "UserId")
	errM := misc.ReadString(data, "Error")
	loginId := misc.ReadInt64(data, "LoginId")

	if (command == "UserLoginByPassword" ||
		command == "UserLoginByCookie") && errM == "" { // successfully logged in
		proxy.Mutex.Lock()
		proxy.MapUserIdToConnId[userId] = clientConnId
		proxy.Mutex.Unlock()
		proxy.Server.Mutex.Lock()
		proxy.Server.MapIdToConnection[clientConnId].UserId = userId
		proxy.Server.MapIdToConnection[clientConnId].LoginId = loginId
		proxy.Server.Mutex.Unlock()
	} else if command == "DisconnectFromServer" {
		proxy.Mutex.Lock()
		connId := proxy.MapUserIdToConnId[userId]
		proxy.Mutex.Unlock()
		if connId != 0 && proxy.Server != nil {
			proxy.Server.Mutex.Lock()
			conn := proxy.Server.MapIdToConnection[connId]
			proxy.Server.Mutex.Unlock()
			if conn != nil {
				conn.Close(l.Get(l.M009LoggedInDiffDevice))
			}
		}
	} else {
		if userId == 0 {
			if proxy.Server != nil {
				proxy.Server.Mutex.Lock()
				conn := proxy.Server.MapIdToConnection[clientConnId]
				proxy.Server.Mutex.Unlock()
				if conn != nil {
					conn.Write(message)
				}
			}
		} else {
			proxy.WriteToUserId(userId, message)
		}
	}
}

func (proxy *Proxy) doAfterClosingBackendConnection(
	connection *nwebsocket.Connection) {
	proxy.Mutex.Lock()
	delete(proxy.MapBackendConn, connection.ConnId)
	proxy.Mutex.Unlock()
}

func (proxy *Proxy) ListenToClients() {
	proxy.Server = nwebsocket.CreateServer(
		zconfig.LimitNConnsPerIp, zconfig.LimitNRequestsPerSecond)
	proxy.Server.ListenAndServe(zconfig.ProxyPort,
		proxy.doAfterReceivingClientMessage,
		proxy.doAfterClosingClientConnection)
}

// get a random connection from the pool (proxy.MapBackendConn)
func (proxy *Proxy) getABackendConnection() *nwebsocket.Connection {
	proxy.Mutex.Lock()
	defer proxy.Mutex.Unlock()
	keys := make([]int64, len(proxy.MapBackendConn))
	i := 0
	for k, _ := range proxy.MapBackendConn {
		keys[i] = k
		i++
	}
	conn := proxy.MapBackendConn[misc.ChoiceInt64s(keys)]
	return conn
}

// add ProxyId/ConnId/UserId/RemoteAddr field to client's data then send to backend,
func (proxy *Proxy) doAfterReceivingClientMessage(
	connection *nwebsocket.Connection, message []byte) {
	var data map[string]interface{}
	parseTextErr := json.Unmarshal(message, &data)
	if parseTextErr != nil {
		return
	}
	data["ProxyId"] = proxy.ProxyId
	data["ConnId"] = connection.ConnId
	data["ClientIpAddr"] = connection.Ip()
	data["UserId"] = connection.UserId
	bs, _ := json.Marshal(data)
	backendConn := proxy.getABackendConnection()
	if backendConn != nil {
		backendConn.Write(bs)
	}
}

// send a Disconnect command to backend,
// clean this connection in proxy
func (proxy *Proxy) doAfterClosingClientConnection(
	connection *nwebsocket.Connection) {
	backendConn := proxy.getABackendConnection()
	if backendConn != nil {
		backendConn.WriteMap(nil, map[string]interface{}{
			"Command": "DisconnectFromClient",
			"UserId":  connection.UserId,
			"LoginId": connection.LoginId,
		})
	}
	connection.LoginId = 0
	proxy.Server.CleanDisconnection(connection)
	proxy.Mutex.Lock()
	delete(proxy.MapUserIdToConnId, connection.UserId)
	proxy.Mutex.Unlock()
}

type Backend struct {
	Server                     *nwebsocket.Server
	Mutex                      sync.Mutex
	MapProxyIdToMapConnections map[int64]map[int64]*nwebsocket.Connection
	MapUserIdToProxyId         map[int64]int64
}

// call this func in main
func InitBackend(
	doAfterReceivingMessage func(
		connection *nwebsocket.Connection, message []byte),
) {
	backend := &Backend{
		MapProxyIdToMapConnections: make(map[int64]map[int64]*nwebsocket.Connection),
		MapUserIdToProxyId:         make(map[int64]int64),
	}
	backend.Server = nwebsocket.CreateServer(999999, 999999)
	backend.Server.ListenAndServe(zconfig.BackendPort, doAfterReceivingMessage,
		backend.Server.CleanDisconnection)
	GBackend = backend
}

// change backend.MapProxyIdToMapConnections
func (backend *Backend) HandleProxyConnect(proxyId int64, conn *nwebsocket.Connection) {
	backend.Mutex.Lock()
	defer backend.Mutex.Unlock()
	if backend.MapProxyIdToMapConnections[proxyId] == nil {
		backend.MapProxyIdToMapConnections[proxyId] = make(map[int64]*nwebsocket.Connection)
	}
	backend.MapProxyIdToMapConnections[proxyId][conn.ConnId] = conn
}

// change MapUserIdToProxyId
func (backend *Backend) HandleClientLogIn(proxyId int64, userId int64) {
	backend.Mutex.Lock()
	defer backend.Mutex.Unlock()
	backend.MapUserIdToProxyId[userId] = proxyId
}

// change MapUserIdToProxyId
func (backend *Backend) HandleClientDisconnect(userId int64) {
	backend.Mutex.Lock()
	defer backend.Mutex.Unlock()
	delete(backend.MapUserIdToProxyId, userId)
}

func (backend *Backend) WriteToProxyId(proxyId int64, message []byte) {
	backend.Mutex.Lock()
	defer backend.Mutex.Unlock()
	if backend.MapProxyIdToMapConnections[proxyId] != nil {
		keys := make([]int64, len(backend.MapProxyIdToMapConnections[proxyId]))
		i := 0
		for k, _ := range backend.MapProxyIdToMapConnections[proxyId] {
			keys[i] = k
			i++
		}
		conn := backend.MapProxyIdToMapConnections[proxyId][misc.ChoiceInt64s(keys)]
		if conn != nil {
			conn.Write(message)
		}
	} else {
		fmt.Println("Invalid proxyId", proxyId)
	}
}

func WriteMapToUserId(userId int64, err error, data map[string]interface{}) {
	GBackend.Mutex.Lock()
	proxyId := GBackend.MapUserIdToProxyId[userId]
	GBackend.Mutex.Unlock()
	message := nwebsocket.MapToBytes(err, data)
	GBackend.WriteToProxyId(proxyId, message)
}

func WriteMapToAll(err error, data map[string]interface{}) {
	GBackend.Mutex.Lock()
	uids := make([]int64, len(GBackend.MapUserIdToProxyId))
	i := 0
	for uid, _ := range GBackend.MapUserIdToProxyId {
		uids[i] = uid
		i++
	}
	GBackend.Mutex.Unlock()
	for _, uid := range uids {
		WriteMapToUserId(uid, err, data)
	}
}
