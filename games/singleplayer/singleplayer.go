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
	"fmt"
	"sync"
	"time"

	//	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/zconfig"
)

func init() {
	_ = fmt.Println
	_ = zconfig.Language
}

type Game struct {
	GameCode              string
	MoneyType             string
	MatchCounter          int64
	MapUidToMatch         map[int64]*Match
	MapUidToBaseMoney     map[int64]float64
	MapUidToRecentMatches map[int64]*misc.LimitedList
	BigWins               *misc.LimitedList
	// map jackpot name to jackpot
	//	Jackpots           map[string]*Jackpot
	ChanActionReceiver chan *Action
	Mutex              sync.Mutex
}

type Match struct {
	GameCode           string
	MatchId            string
	StartedTime        time.Time
	BaseMoney          float64
	ResultChangedMoney float64
	ResultDetail       string
	Actions            string
}

type Action struct {
	ActionName   string
	UserId       int64
	Data         map[string]interface{}
	CreatedTime  time.Time
	ChanResponse chan error
}

//*/
