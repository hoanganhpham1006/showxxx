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
	_ = zconfig.ProxyPort
}

func Test02(t *testing.T) {
	for i := 0; i < 1; i++ {
		// serverAddr := fmt.Sprintf("ws://localhost%v/ws", zconfig.ProxyPort)
		//		serverAddr := fmt.Sprintf("ws://43.239.221.117%v/ws", zconfig.ProxyPort)
		serverAddr := "ws://tung:2052/ws"
		wsConn, _, e := websocket.DefaultDialer.Dial(serverAddr, nil)
		if e != nil {
			t.Error(e)
			return
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

		//		c.WriteMap(nil, map[string]interface{}{
		//			"Command":  "UserLoginByPassword",
		//			"Username": "daominah",
		//			"Password": "123qwe",
		//		})

		c.WriteMap(nil, map[string]interface{}{
			"method": "auth",
			"data": map[string]interface{}{
				"type":     "auth_player_by_password",
				"username": "daominah",
				"password": "123qwe",
			},
		})
		time.Sleep(500 * time.Millisecond)

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
		//		c.WriteMap(nil, map[string]interface{}{
		//			"Command":  "UserFollowing",
		//			"Key":      "T",
		//			"TargetId": 7,
		//			"UserId":   1,
		//		})
		c.WriteMap(nil, map[string]interface{}{
			"method": "TangkasquChooseBaseMoney",
			"data":   map[string]interface{}{"BaseMoney": 20000},
		})
		c.WriteMap(nil, map[string]interface{}{
			"method": "TangkasquCreateMatch",
		})
		for i := 0; i < 10; i++ {
			c.WriteMap(nil, map[string]interface{}{
				"method": "TangkasquSendMove",
				"data":   map[string]interface{}{"MoveType": "MOVE_BET"},
			})
		}

	}
	select {}
}
