// This package provides general feature for a game.
// Game requirements:
//   * user choose moneyType, baseMoney
//	 * user or system start a match
//   * player view his recent match's results
//   * player view big wins from all matches
//   * game can have jackpots (users contribute to the jackpot when
//	  they play a match)
// Match requirements:
//   * user make move to play a match
//   * user can get match detail everytime
//   * system finish a match
package games

//import (
//	//	"errors"
//	"encoding/json"
//	"fmt"
//	"strconv"
//	"sync"
//	"time"
//
//	//	"github.com/daominah/livestream/users"
//	"github.com/daominah/livestream/misc"
//	"github.com/daominah/livestream/zconfig"
//	"github.com/daominah/livestream/zdatabase"
//)
//
//const (
//	ACTION_CHOOSE_MONEY_TYPE   = "ACTION_CHOOSE_MONEY_TYPE"
//	ACTION_CHOOSE_BASE_MONEY   = "ACTION_CHOOSE_BASE_MONEY"
//	ACTION_CREATE_MATCH        = "ACTION_START_MATCH"
//	ACTION_VIEW_RECENT_MATCHES = "ACTION_VIEW_OLD_MATCHES"
//	ACTION_VIEW_BIG_WINS       = "ACTION_VIEW_BIG_WINS"
//
//	ACTION_MAKE_MOVE        = "ACTION_MAKE_MOVE"
//	ACTION_GET_MATCH_DETAIL = "ACTION_GET_MATCH_DETAIL"
//	ACTION_FINISH_MATCH     = "ACTION_FINISH_MATCH"
//)
//
//func init() {
//	_ = fmt.Println
//	_ = zconfig.Language
//}
//
//type Game struct {
//	GameCode     string
//	MatchCounter int64
//	MapMatches   map[string]*Match
//	//	MapUidToOption        map[int64]Option
//	MapUidToRecentResults map[int64]*misc.LimitedList
//	BigWins               *misc.LimitedList
//	// map jackpot name to jackpot
//	//	Jackpots           map[string]*Jackpot
//	ChanActionReceiver chan *Action
//	Mutex              sync.Mutex
//}
//
//func CreateGame(gameCode string) *Game {
//	game := &Game{
//		GameCode:              gameCode,
//		MapUidToOptions:       make(map[string]*Match),
//		MapUidToBaseMoney:     make(map[int64]float64),
//		MapUidToRecentResults: make(map[int64]*misc.LimitedList),
//		BigWins:               misc.CreateLimitedList(10),
//		ChanActionReceiver:    make(chan *Action),
//	}
//	matchCounterS := zdatabase.LoadGlobalVar(matchCounterKey(game))
//	game.MatchCounter, _ = strconv.ParseInt(matchCounterS, 10, 64)
//
//	return game
//}
//
//func (game *Game) LoopReceiveActions() {
//	for {
//		action := <-game.ChanActionReceiver
//		go func(action *Action) {
//			switch action.ActionName {
//			case ACTION_CREATE_MATCH:
//				//
//			case ACTION_SET_BASE_MONEY:
//				//
//			default:
//				//
//			}
//		}(action)
//	}
//}
//
//func matchCounterKey(game *Game) string {
//	return fmt.Sprintf("MatchCounter_%v_%v", game.GameCode, game.MoneyType)
//}
//
//func CreateMatch() *Match {
//	return &Match{}
//}
//
//type Match struct {
//	GameCode           string
//	MoneyType          string
//	MatchId            string
//	UserId             int64
//	StartedTime        time.Time
//	BaseMoney          float64
//	ResultChangedMoney float64
//	ResultDetail       string
//	Actions            string
//	ChanActionReceiver chan *Action
//}
//
//func (match *Match) LoopReceiveActions() {
//	for {
//		action := <-match.ChanActionReceiver
//		if action.ActionName == ACTION_FINISH_MATCH {
//			//
//			break
//		} else {
//			go func(action *Action) {
//				switch action.ActionName {
//				case ACTION_MAKE_MOVE:
//					//
//				case ACTION_GET_MATCH_DETAIL:
//					//
//				default:
//					//
//				}
//			}(action)
//		}
//	}
//}
//
//type Action struct {
//	ActionName   string
//	UserId       int64
//	Data         map[string]interface{}
//	CreatedTime  time.Time
//	ChanResponse chan error
//}
//
//func (a *Action) String() string {
//	shortObj := map[string]interface{}{"ActionName": a.ActionName,
//		"UserId": a.UserId, "Data": a.Data, "CreatedTime": a.CreatedTime}
//	bs, e := json.Marshal(shortObj)
//	if e != nil {
//		return "{}"
//	}
//	return string(bs)
//}
