package egg

import (
	"time"
	//	"encoding/json"
	"fmt"
	"testing"

	"github.com/daominah/livestream/nbackend"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zglobal"
)

func Test01(t *testing.T) {
	epsilon := float64(0.001)
	ap := calcAvarageProfit()
	if !((1-epsilon < ap) && (ap < 1+epsilon)) {
		t.Error(ap)
	}
}

func Test02(t *testing.T) {
	zglobal.GameEggPayoutRate = 0.9
	nbackend.InitBackend(nil)

	user, _ := users.GetUser(1)
	users.ChangeUserMoney(1, users.MT_CASH, -user.MapMoney[users.MT_CASH], "", false)
	users.ChangeUserMoney(1, users.MT_CASH, 10000, "", false)
	game := &EggGame{}
	game.Init(GAME_CODE_EGG, 1000)
	for {
		time.Sleep(200 * time.Millisecond)
		match := &EggMatch{}
		game.InitMatch(1, match)
		time.Sleep(100 * time.Millisecond)
		go func() {
			for {
				time.Sleep(100 * time.Millisecond)
				err := match.SendMove(map[string]interface{}{"HammerIndex": 1})
				fmt.Println("SendMove", err)
			}
		}()
	}
	select {}
}
