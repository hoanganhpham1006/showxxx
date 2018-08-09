package car

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/daominah/livestream/games/multiplayer"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/nbackend"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/zglobal"
)

const (
	GAME_CODE      = "car"
	DURATION_MATCH = 10 * time.Second
	DURATION_IDLE  = 3 * time.Second
	NUMBER_OF_CARS = 4
)

func init() {
	_ = fmt.Println
	rand.Seed(time.Now().Unix())
}

type CarGame struct {
	multiplayer.Game
	SharedMatch *CarMatch
}

func (game *CarGame) GetPlayingMatch(userId int64) multiplayer.MatchInterface {
	return game.SharedMatch
}

func (game *CarGame) PeriodicallyCreateMatch() {
	go func() {
		for {
			match := &CarMatch{}
			game.InitMatch(match)
			game.SharedMatch = match
			time.Sleep(DURATION_MATCH + DURATION_IDLE)
		}
	}()
}

func (game *CarGame) GetCurrentMatch() (map[string]interface{}, error) {
	sharedMatch := game.SharedMatch
	if sharedMatch == nil {
		return nil, errors.New("sharedMatch == nil")
	}
	return game.SharedMatch.ToMap(), nil
}

type CarMatch struct {
	multiplayer.Match
	MapUserIdToMapCarToValue map[int64]map[int]float64

	// to calculate turn remaining duration
	StartedTime     time.Time
	MovesLog        []*Move
	WinningCarIndex int
	IsFinished      bool

	ChanMove    chan *Move `json:"-"`
	ChanMoveErr chan error `json:"-"`
}

type Move struct {
	UserId      int64
	CarIndex    int
	BetValue    float64
	CreatedTime time.Time
}

func (match *CarMatch) String() string {
	match.Mutex.Lock()
	defer match.Mutex.Unlock()
	bs, e := json.Marshal(match)
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (match *CarMatch) ToMap() map[string]interface{} {
	s := match.String()
	r := map[string]interface{}{}
	json.Unmarshal([]byte(s), &r)
	return r
}

// command in [COMMAND_MATCH_START, COMMAND_MATCH_UPDATE, COMMAND_MATCH_FINISH]
func (match *CarMatch) UpdateMatch(command string) {
	data := match.ToMap()
	data["Command"] = command
	data["TurnRemainingSeconds"] =
		match.StartedTime.Add(DURATION_MATCH).Sub(time.Now()).Seconds()
	match.Mutex.Lock()
	for uid, _ := range match.MapUserIds {
		nbackend.WriteMapToUserId(uid, nil, data)
	}
	match.Mutex.Unlock()
	zconfig.TPrint("_____________________________________")
	zconfig.TPrint(time.Now(), command, data)
}

func (match *CarMatch) Start() {
	match.Mutex.Lock()
	match.MapUserIdToMapCarToValue = make(map[int64]map[int]float64)
	match.StartedTime = time.Now()
	match.MovesLog = make([]*Move, 0)
	match.ChanMove = make(chan *Move)
	match.ChanMoveErr = make(chan error)
	match.Mutex.Unlock()
	//
	match.UpdateMatch(multiplayer.COMMAND_MATCH_START)
	matchTimeout := time.After(DURATION_MATCH)
LoopWaitingLegalMove:
	for {
		select {
		case move := <-match.ChanMove:
			err := match.MakeMove(move)
			if err == nil {
				match.UpdateMatch(multiplayer.COMMAND_MATCH_UPDATE)
			}
			select {
			case match.ChanMoveErr <- err:
			default:
			}
		case <-matchTimeout:
			break LoopWaitingLegalMove
		}
	}
	// betting duration is over
	match.Mutex.Lock()
	match.IsFinished = true
	match.WinningCarIndex = rand.Intn(NUMBER_OF_CARS)
	for uid, mapCarToValue := range match.MapUserIdToMapCarToValue {
		for carI, value := range mapCarToValue {
			if carI == match.WinningCarIndex {
				winningMoney := 4 * value * zglobal.GameCarPayoutRate
				match.MapUserIdToResultChangedMoney[uid] += winningMoney
				users.ChangeUserMoney(uid, match.MoneyType, winningMoney,
					users.REASON_PLAY_GAME, false)
			}
		}
	}
	for _, changedMoney := range match.MapUserIdToResultChangedMoney {
		match.ResultChangedMoney += changedMoney
	}
	match.Mutex.Unlock()
	//
	match.ResultDetail = match.String()
	match.Game.FinishMatch(match)
	match.UpdateMatch(multiplayer.COMMAND_MATCH_FINISH)
}

func (m *CarMatch) SendMove(data map[string]interface{}) error {
	move := &Move{
		UserId:      misc.ReadInt64(data, "SourceUserId"),
		CarIndex:    int(misc.ReadInt64(data, "CarIndex")),
		BetValue:    misc.ReadFloat64(data, "BetValue"),
		CreatedTime: time.Now()}
	user, err := users.GetUser(move.UserId)
	if user == nil {
		return err
	}
	t := time.After(1 * time.Second)
	select {
	case m.ChanMove <- move:
		t2 := time.After(1 * time.Second)
		select {
		case err := <-m.ChanMoveErr:
			return err
		case <-t2:
			return errors.New("<-m.ChanMoveErr timeout")
		}
	case <-t:
		return errors.New(l.Get(l.M043MovingDurationEnded))
	}
}

func (m *CarMatch) MakeMove(move *Move) error {
	m.AddUserId(move.UserId)
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	if move.CarIndex >= NUMBER_OF_CARS || move.CarIndex <= 0 {
		return errors.New(l.Get(l.M042GameInvalidCarIndex))
	}
	m.MovesLog = append(m.MovesLog, move)
	requiringMoney := move.BetValue
	_, err := users.ChangeUserMoney(move.UserId, m.MoneyType, -requiringMoney,
		users.REASON_PLAY_GAME, true)
	if err != nil {
		return err
	}
	if _, isIn := m.MapUserIdToMapCarToValue[move.UserId]; !isIn {
		m.MapUserIdToMapCarToValue[move.UserId] = make(map[int]float64)
	}
	m.MapUserIdToMapCarToValue[move.UserId][move.CarIndex] += requiringMoney
	m.MapUserIdToResultChangedMoney[move.UserId] -= requiringMoney
	return nil
}
