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
	for i := 0; i < 1; i++ {
		//		serverAddr := fmt.Sprintf("ws://localhost%v/ws", zconfig.ProxyPort)
		serverAddr := fmt.Sprintf("ws://43.239.221.117%v/ws", zconfig.ProxyPort)
		wsConn, _, e := websocket.DefaultDialer.Dial(serverAddr, nil)
		if e != nil {
			t.Error(e)
		}
		c := nwebsocket.CreateConnection(wsConn, 999999)
		go c.ReadPump(nil, nil)
		go c.WritePump(nil)
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
			"Command":  "UserLoginByPassword",
			"Username": "daominah",
			"Password": "123qwe",
		})
		time.Sleep(200 * time.Millisecond)
		//		c.WriteMap(nil, map[string]interface{}{
		//			"Command":        "ConversationCreateMessage",
		//			"ConversationId": 1,
		//			"MessageContent": "hohohaha",
		//		})
		//		c.WriteMap(nil, map[string]interface{}{
		//			"Command":  "UserFollow",
		//			"Key":      "V",
		//			"TargetId": 7,
		//		})
		c.WriteMap(nil, map[string]interface{}{
			"Command":  "UserFollowing",
			"Key":      "T",
			"TargetId": 7,
			"UserId":   1,
		})

	}
	select {}
}
