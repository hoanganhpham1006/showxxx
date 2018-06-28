package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/daominah/livestream/connections"
	"github.com/daominah/livestream/zconfig"
)

func Test01(t *testing.T) {
	_ = fmt.Println
}

func Test02(t *testing.T) {
	serverAddr := fmt.Sprintf("ws://localhost%v/ws", zconfig.WebsocketPort)
	wsConn, _, e := websocket.DefaultDialer.Dial(serverAddr, nil)
	if e != nil {
		t.Error(e)
	}
	c := connections.CreateConnection(wsConn)
	c.TestingStart()
	c.WriteMap(nil, map[string]interface{}{
		"Command":    "UserLoginByPassword",
		"Username":   "daominah",
		"Password":   "123qwe",
		"DeviceName": "LinuxMint18",
		"AppName":    "Eclipse",
	})
	time.Sleep(100 * time.Millisecond)
	//
	c.WriteMap(nil, map[string]interface{}{
		"Command":     "UserChangeProfileImage",
		"ImageBase64": []byte("Profile picture 0"),
	})
	//
	c.WriteMap(nil, map[string]interface{}{
		"Command": "RankGetLeaderBoard",
		"RankId":  15,
	})
	//
	c.WriteMap(nil, map[string]interface{}{
		"Command": "RankGetLeaderBoard",
		"RankId":  15,
	})
	//
	c.WriteMap(nil, map[string]interface{}{
		"Command":     "ConversationCreate",
		"RecipientId": 7,
	})
	select {}
}
