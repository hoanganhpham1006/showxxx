package main

import (
	"fmt"
	"time"

	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/daominah/livestream/connections"
	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/zdatabase"
	//	"github.com/daominah/livestream/zglobal"
	"github.com/daominah/livestream/admintool"
	"github.com/daominah/livestream/games/singleplayer"
	"github.com/daominah/livestream/games/singleplayer/egg"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/rank"
	"github.com/daominah/livestream/streams"
)

// read only map
var MapSGames = make(map[string]singleplayer.GameInterface)

func init() {
	fmt.Println("")
	_ = zdatabase.DbPool
}

func main() {
	// app profile: memory, inused objects, goroutines...
	go func() {
		log.Println(http.ListenAndServe("localhost"+zconfig.ProfilePort, nil))
	}()
	runtime.SetBlockProfileRate(1)

	// Create tables in database. The second call should return duplicate errors.
	zdatabase.InitTables()

	//
	MapSGames[egg.GAME_CODE_EGG] = &egg.EggGame{}
	for gameCode, game := range MapSGames {
		game.Init(gameCode)
	}

	// listen to clients
	connections.ListenAndServe(doAfterReceivingMessage, doAfterClosingConnection)
	streams.ForwarderListenAndServer()
	admintool.ListenAndServe()

	// reset rank leaderboard
	go func() {
		for {
			durToNextReset := misc.NextDay00().Sub(time.Now())
			time.Sleep(durToNextReset)
			rank.Reset(rank.RANK_PURCHASED_CASH_DAY)
			rank.Reset(rank.RANK_RECEIVED_CASH_DAY)
			rank.Reset(rank.RANK_SENT_CASH_DAY)
		}
	}()
	go func() {
		for {
			durToNextReset := misc.NextWeek00().Sub(time.Now())
			time.Sleep(durToNextReset)
			rank.Reset(rank.RANK_PURCHASED_CASH_WEEK)
			rank.Reset(rank.RANK_RECEIVED_CASH_WEEK)
			rank.Reset(rank.RANK_SENT_CASH_WEEK)
			rank.Reset(rank.RANK_N_FOLLOWERS_WEEK)
		}
	}()
	go func() {
		for {
			durToNextReset := misc.NextMonth00().Sub(time.Now())
			time.Sleep(durToNextReset)
			rank.Reset(rank.RANK_PURCHASED_CASH_MONTH)
			rank.Reset(rank.RANK_RECEIVED_CASH_MONTH)
			rank.Reset(rank.RANK_SENT_CASH_MONTH)
		}
	}()

	//
	fmt.Println("main hohohaha")
	select {}
}
