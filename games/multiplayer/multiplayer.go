// This package provides general feature for a game.
// Game requirements:
//   * user choose moneyType, baseMoney
//      * user or system start a match
//   * player view his recent match's results
//   * player view big wins from all matches
//   * game can have jackpots (users contribute to the jackpot when
//       they play a match)
//   * user can only playing one match at a time, and can get the match detail
//   * user make move to play a match
package multiplayer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	//	l "github.com/daominah/livestream/language"
	//	"github.com/daominah/livestream/misc"
	//	"github.com/daominah/livestream/users"
	//	zc "github.com/daominah/livestream/zconfig"
	//	"github.com/daominah/livestream/games"
	"github.com/daominah/livestream/zdatabase"
)

const (
	COMMAND_MATCH_START  = "COMMAND_MATCH_START"
	COMMAND_MATCH_UPDATE = "COMMAND_MATCH_UPDATE"
	COMMAND_MATCH_FINISH = "COMMAND_MATCH_FINISH"
)

func init() {
	_ = fmt.Println
}

type GameInterface interface {
	// init game fields
	Init(gameCode string, moneyTypeDefault string, baseMoneyDefault float64)
	// set basic match's fields, change MapUidToPlayingMatchId
	InitMatch(match MatchInterface) error
	// call by server, not for client
	FinishMatch(match MatchInterface)
	GetPlayingMatch(userId int64) MatchInterface
}

type Game struct {
	GameCode         string
	MatchCounter     int64
	MoneyTypeDefault string
	BaseMoneyDefault float64
	// protect below maps
	Mutex                  sync.Mutex `json:"-"`
	MapUidToPlayingMatchId map[int64]string
	// map matchId to matchObj
	MapMatches map[string]MatchInterface `json:"-"`
}

func (game *Game) Init(
	gameCode string, moneyTypeDefault string, baseMoneyDefault float64) {
	game.GameCode = gameCode
	game.BaseMoneyDefault = baseMoneyDefault
	matchCounterS := zdatabase.LoadGlobalVar(matchCounterKey(game))
	game.MatchCounter, _ = strconv.ParseInt(matchCounterS, 10, 64)

	game.MapUidToPlayingMatchId = make(map[int64]string)
	game.MapMatches = make(map[string]MatchInterface)
}

func matchCounterKey(game *Game) string {
	return fmt.Sprintf("MatchCounter_%v", game.GameCode)
}

type MatchInterface interface {
	Start()
	// save to database
	Archive() error
	ToMap() map[string]interface{}
	SetGame(game GameInterface)
	SetGameCode(gameCode string)
	SetMoneyType(moneyType string)
	SetMatchId(matchId string)
	AddUserId(userId int64)
	SetStartedTime(t time.Time)
	SetBaseMoney(baseMoney float64)
	GetMatchId() string
	GetUserIds() map[int64]bool
	InitMaps()
	SendMove(data map[string]interface{}) error
}

type Match struct {
	Game        GameInterface `json:"-"`
	GameCode    string
	MatchId     string
	MapUserIds  map[int64]bool
	StartedTime time.Time

	MoneyType string
	BaseMoney float64

	ResultChangedMoney            float64
	MapUserIdToResultChangedMoney map[int64]float64
	ResultDetail                  string

	Mutex sync.Mutex `json:"-"`
}

func (game *Game) InitMatch(match MatchInterface) error {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	game.MatchCounter++
	zdatabase.SaveGlobalVar(matchCounterKey(game), fmt.Sprintf("%v", game.MatchCounter))
	match.SetGame(game)
	match.SetGameCode(game.GameCode)
	match.SetMoneyType(game.MoneyTypeDefault)
	match.SetMatchId(fmt.Sprintf("%v_%010d", game.GameCode, game.MatchCounter))
	match.SetStartedTime(time.Now())
	match.SetBaseMoney(game.BaseMoneyDefault)
	match.InitMaps()
	game.MapMatches[match.GetMatchId()] = match
	//
	go match.Start()
	return nil
}

func (game *Game) FinishMatch(match MatchInterface) {
	match.Archive()
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	for userId, _ := range match.GetUserIds() {
		delete(game.MapUidToPlayingMatchId, userId)
	}
	delete(game.MapMatches, match.GetMatchId())
	return
}

func (game *Game) GetPlayingMatch(userId int64) MatchInterface {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	matchId := game.MapUidToPlayingMatchId[userId]
	match := game.MapMatches[matchId]
	return match
}

// _____________________________________________________________

func (match *Match) Start() {}
func (match *Match) SendMove(data map[string]interface{}) error {
	return errors.New("Virtual func")
}
func (match *Match) String() string {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	bs, e := json.Marshal(match)
	if e != nil {
		return "{}"
	}
	return string(bs)
}
func (match *Match) ToMap() map[string]interface{} {
	s := match.String()
	r := map[string]interface{}{}
	json.Unmarshal([]byte(s), &r)
	return r
}

// _____________________________________________________________

func (match *Match) GetMatchId() string {
	return match.MatchId
}

func (match *Match) GetUserIds() map[int64]bool {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	result := make(map[int64]bool)
	for uid, _ := range match.MapUserIds {
		result[uid] = true
	}
	return result
}

func (match *Match) SetGame(a GameInterface) {
	match.Game = a
}

func (match *Match) SetGameCode(a string) {
	match.GameCode = a
}

func (match *Match) SetMoneyType(a string) {
	match.MoneyType = a
}

func (match *Match) SetMatchId(a string) {
	match.MatchId = a
}

func (match *Match) AddUserId(a int64) {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	match.MapUserIds[a] = true
}

func (match *Match) SetStartedTime(a time.Time) {
	match.StartedTime = a
}

func (match *Match) SetBaseMoney(a float64) {
	match.BaseMoney = a
}

func (match *Match) InitMaps() {
	match.MapUserIds = make(map[int64]bool)
	match.MapUserIdToResultChangedMoney = make(map[int64]float64)
}

func (m *Match) Archive() error {
	_, e := zdatabase.DbPool.Exec(
		`INSERT INTO match_multi (id, game_code, started_time,
            money_type, base_money, result_changed_money, result_detail)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.MatchId, m.GameCode, m.StartedTime,
		m.MoneyType, m.BaseMoney, m.ResultChangedMoney, m.ResultDetail)
	if e != nil {
		return e
	}
	//
	temps := []string{}
	args := []interface{}{}
	i := 0
	for uid, _ := range m.MapUserIds {
		temps = append(temps, fmt.Sprintf("($%v, $%v, $%v)", 3*i+1, 3*i+2, 3*i+3))
		args = append(args, []interface{}{
			m.MatchId, uid, m.MapUserIdToResultChangedMoney[uid]}...)
		i += 1
	}
	queryPart := strings.Join(temps, ", ")
	query := fmt.Sprintf(
		`INSERT INTO match_multi_participant
		    (match_id, user_id, result_changed_money)
		VALUES %v`, queryPart)
	// fmt.Println("query", query)
	_, e = zdatabase.DbPool.Exec(query, args...)
	return e
}
