package nwebsocket

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	_ "github.com/daominah/livestream/zconfig"
)

func Test01(t *testing.T) {
	port := ":11111"
	server := CreateServer(10, 10)
	server.ListenAndServe(port, nil, nil)
	serverAddr := fmt.Sprintf("ws://localhost%v/ws", port)
	//	serverAddr = fmt.Sprintf("ws://43.239.221.117%v/ws", zconfig.WebsocketPort)
	wsConn, _, e := websocket.DefaultDialer.Dial(serverAddr, nil)
	if e != nil {
		t.Error(e)
		return
	}
	conn := CreateConnection(wsConn, 999999)
	go conn.ReadPump(nil, nil)
	go conn.WritePump(nil)
	conn.Write([]byte("hihi"))
	conn.Write([]byte("hihi"))
	conn.Write([]byte("hihi"))
	time.Sleep(1 * time.Second)
	conn.Write([]byte("hihi"))
	conn.Write([]byte("hihi"))
	fmt.Println("conn", conn)
	select {}
}

func Test02(t *testing.T) {

}
