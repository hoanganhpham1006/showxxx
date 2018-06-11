package connections

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/daominah/livestream/zconfig"
)

func Test01(t *testing.T) {
	serverAddr := fmt.Sprintf("ws://localhost%v/ws", zconfig.WebsocketPort)
	wsConn, _, e := websocket.DefaultDialer.Dial(serverAddr, nil)
	if e != nil {
		t.Error(e)
		return
	}
	conn := CreateConnection(wsConn)
	go conn.readPump(nil, nil)
	go conn.writePump(nil)
	time.Sleep(1 * time.Second)
	conn.Write([]byte("hihi"))
	time.Sleep(1 * time.Second)
	//	conn.Close()
	conn.Write([]byte(`{"":"hihi"}`))
	conn.Write([]byte("hihi"))
	select {}
}
