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
	//	for i := 0; i < 1; i++ {
	//	serverAddr := fmt.Sprintf("ws://localhost%v/ws", zconfig.ProxyPort)
	//	serverAddr := fmt.Sprintf("ws://43.239.221.117%v/ws", zconfig.ProxyPort)
	//	serverAddr := fmt.Sprintf("ws://43.239.221.117%v/ws", zconfig.BackendPort)
	//	serverAddr := "ws://149.28.154.5:2052/ws"
	serverAddr := "wss://43.239.221.117:20001/ws"
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
	//	c.WriteMap(nil, map[string]interface{}{
	//		"Command":  "UserLoginByPassword",
	//		"Username": "daominah",
	//		"Password": "123qwe",
	//	})
	c.WriteMap(nil, map[string]interface{}{
		"method": "auth",
		"data": map[string]interface{}{
			"type":     "auth_player_by_password",
			"username": "daominah",
			"password": "123qwe",
		},
	})
	time.Sleep(500 * time.Millisecond)
	//	i := 0
	//	for {
	//		time.Sleep(1000 * time.Millisecond)
	//		i += 1
	//		ds := []map[string]interface{}{
	//			map[string]interface{}{
	//				"CommandId": i,
	//				"Command":   "UserFollowing",
	//				"UserId":    1,
	//			},
	//			map[string]interface{}{
	//				"CommandId": i,
	//				"Command":   "StreamAllSummaries",
	//			},
	//			map[string]interface{}{
	//				"CommandId": i,
	//				"Command":   "RankGetLeaderBoard",
	//				"RankId":    16,
	//			},
	//			map[string]interface{}{
	//				"CommandId": i,
	//				"Command":   "AdnvidGetListVideoCategories",
	//			},
	//			map[string]interface{}{
	//				"CommandId": i,
	//				"Command":   "AdnvidGetListVideos",
	//				"Limit":     10,
	//				"Offset":    0,
	//				"OrderBy":   "id",
	//			},
	//		}
	//		c.WriteMap(nil, ds[i%len(ds)])
	//	}

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

	//	c.WriteMap(nil, map[string]interface{}{"ClientIpAddr": "171.224.91.47", "Command": "ConversationAllSummaries", "CommandId": 4, "ConnId": 5162, "Filter": "FILTER_ALL", "NConversation": 100, "ProxyId": 1535791370987156113, "SourceUserId": 1903})

	//	c.WriteMap(nil, map[string]interface{}{
	//		"method": "TangkasquChooseBaseMoney",
	//		"data":   map[string]interface{}{"BaseMoney": 20000},
	//	})
	//	c.WriteMap(nil, map[string]interface{}{
	//		"method": "DragontigerMatchesHistory",
	//	})
	//	for i := 0; i < 10; i++ {
	//		c.WriteMap(nil, map[string]interface{}{
	//			"method": "TangkasquSendMove",
	//			"data":   map[string]interface{}{"MoveType": "MOVE_BET"},
	//		})
	//	}

	//	}
	select {}
}
