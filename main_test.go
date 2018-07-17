package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/daominah/livestream/nwebsocket"
	"github.com/daominah/livestream/zconfig"
)

func Test01(t *testing.T) {
	_ = time.Sleep
	_ = fmt.Println
}

func Test02(t *testing.T) {
	for i := 0; i < 100000; i++ {
		serverAddr := fmt.Sprintf("ws://localhost%v/ws", zconfig.WebsocketPort)
		//	serverAddr := fmt.Sprintf("ws://43.239.221.117%v/ws", zconfig.WebsocketPort)
		wsConn, _, e := websocket.DefaultDialer.Dial(serverAddr, nil)
		if e != nil {
			t.Error(e)
		}
		c := nwebsocket.CreateConnection(wsConn)
		c.TestingStart()
		//	for i := 50; i < 1000; i++ {
		//		c.WriteMap(nil, map[string]interface{}{
		//			"Command":    "UserCreate",
		//			"Username":   fmt.Sprintf("thuy%v", i),
		//			"Password":   "123qwe",
		//			"DeviceName": "LinuxMint18",
		//			"AppName":    "Eclipse",
		//		})
		//	}

		c.WriteMap(nil, map[string]interface{}{
			"Command":      "UserLoginByCookie",
			"LoginSession": "7b226c6f67696e54696d65223a22323031382d30372d30325430393a31363a33302e3738303231353334362b30373a3030222c22757365724964223a223134227d",
		})
		// time.Sleep(200 * time.Millisecond)
		c.WriteMap(nil, map[string]interface{}{
			"Command":        "ConversationCreateMessage",
			"ConversationId": 1,
			"MessageContent": "hohohaha",
		})
	}

	select {}
}
