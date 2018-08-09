package car

import (
	"time"
	//	"encoding/json"
	"fmt"
	"testing"

	"github.com/daominah/livestream/nbackend"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zglobal"
)

func Test02(t *testing.T) {
	zglobal.GameCarPayoutRate = 0.9
	nbackend.InitBackend(nil)

	for uid := int64(1); uid < 6; uid++ {
		user, _ := users.GetUser(uid)
		users.ChangeUserMoney(uid, users.MT_CASH, -user.MapMoney[users.MT_CASH], "", false)
		users.ChangeUserMoney(uid, users.MT_CASH, 10000, "", false)
	}
	game := &CarGame{}
	game.Init(GAME_CODE, users.MT_CASH, 100)
	var match *CarMatch
	go func() {
		for {
			match = &CarMatch{}
			game.InitMatch(match)
			time.Sleep(6000 * time.Millisecond)
		}
	}()
	go func() {
		i := int64(0)
		for {
			time.Sleep(700 * time.Millisecond)
			err := match.SendMove(map[string]interface{}{
				"UserId":   i%5 + 1,
				"CarIndex": i % NUMBER_OF_CARS,
				"BetValue": i * 100,
			})
			fmt.Println("SendMove", err)
			i += 1
		}
	}()
	select {}
}
