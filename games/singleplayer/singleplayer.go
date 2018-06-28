// Single player game: only one player create a match and play, others cant join
// the match.
// Example single player games: slot (pay money to spin, receive random reward).
// Requirements:
//  * player choose option before create a match (base money, ..)
//  * player send moves to play the match
//  * player view his old match's results
//  * player view big wins from the other's matches
//  * optional jackpots (players contribute to the jackpot when they play a match)
package singleplayer

///*

import (
	//	"errors"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	//	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/zdatabase"
)

const (
	ACTION_CREATE_MATCH   = "ACTION_CREATE_MATCH"
	ACTION_SET_BASE_MONEY = "ACTION_SET_BASE_MONEY"

	ACTION_MAKE_MOVE        = "ACTION_MAKE_MOVE"
	ACTION_GET_MATCH_DETAIL = "ACTION_GET_MATCH_DETAIL"
	ACTION_FINISH_MATCH     = "ACTION_FINISH_MATCH"
)

func init() {
	_ = fmt.Println
	_ = zconfig.Language
}

func CreateGame(gameCode string, moneyType string) *Game {
	game := &Game{
		GameCode:              gameCode,
		MoneyType:             moneyType,
		MapUidToMatch:         make(map[int64]*Match),
		MapUidToBaseMoney:     make(map[int64]float64),
		MapUidToRecentResults: make(map[int64]*misc.LimitedList),
		BigWins:               misc.CreateLimitedList(10),
		ChanActionReceiver:    make(chan *Action),
	}
	matchCounterS := zdatabase.LoadGlobalVar(matchCounterKey(game))
	game.MatchCounter, _ = strconv.ParseInt(matchCounterS, 10, 64)

	return game
}

type Game struct {
	GameCode              string
	MoneyType             string
	MatchCounter          int64
	MapUidToMatch         map[int64]*Match
	MapUidToBaseMoney     map[int64]float64
	MapUidToRecentResults map[int64]*misc.LimitedList
	BigWins               *misc.LimitedList
	// map jackpot name to jackpot
	//	Jackpots           map[string]*Jackpot
	ChanActionReceiver chan *Action
	Mutex              sync.Mutex
}

func (game *Game) LoopReceiveActions() {
	for {
		action := <-game.ChanActionReceiver
		go func(action *Action) {
			switch action.ActionName {
			case ACTION_CREATE_MATCH:
				//
			case ACTION_SET_BASE_MONEY:
				//
			default:
				//
			}
		}(action)
	}
}

func matchCounterKey(game *Game) string {
	return fmt.Sprintf("MatchCounter_%v_%v", game.GameCode, game.MoneyType)
}

func CreateMatch() *Match {
	return &Match{}
}

type Match struct {
	GameCode           string
	MoneyType          string
	MatchId            string
	UserId             int64
	StartedTime        time.Time
	BaseMoney          float64
	ResultChangedMoney float64
	ResultDetail       string
	Actions            string
	ChanActionReceiver chan *Action
}

func (match *Match) LoopReceiveActions() {
	for {
		action := <-match.ChanActionReceiver
		if action.ActionName == ACTION_FINISH_MATCH {
			//
			break
		} else {
			go func(action *Action) {
				switch action.ActionName {
				case ACTION_MAKE_MOVE:
					//
				case ACTION_GET_MATCH_DETAIL:
					//
				default:
					//
				}
			}(action)
		}
	}
}

type Action struct {
	ActionName   string
	UserId       int64
	Data         map[string]interface{}
	CreatedTime  time.Time
	ChanResponse chan error
}

func (a *Action) String() string {
	shortObj := map[string]interface{}{"ActionName": a.ActionName,
		"UserId": a.UserId, "Data": a.Data, "CreatedTime": a.CreatedTime}
	bs, e := json.Marshal(shortObj)
	if e != nil {
		return "{}"
	}
	return string(bs)
}

//*/
