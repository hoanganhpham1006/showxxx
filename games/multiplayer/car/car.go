package car

import (
	"encoding/json"
	"math/rand"
	//	"fmt"
	"errors"
	"time"

	"github.com/daominah/livestream/games/multiplayer"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/nbackend"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/zglobal"
)

const (
	GAME_CODE_EGG = "car"
	DURATION_TURN = 20 * time.Second
)

func init() {
	rand.Seed(time.Now().Unix())
}

type CarGame struct {
	multiplayer.Game
}

type CarMatch struct {
	multiplayer.Match

	// map hammer to its cost
	MapHammers map[int]float64

	// to calculate turn remaining duration
	TurnStartedTime time.Time
	UserWonMoney    float64
	UserLostMoney   float64
	MovesLog        []*Move

	ChanMove    chan *Move `json:"-"`
	ChanMoveErr chan error `json:"-"`
}

type Move struct {
	UserId      int64
	CarIndex    int
	Bet         float64
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
		match.TurnStartedTime.Add(DURATION_TURN).Sub(time.Now()).Seconds()
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
	match.MapHammers = map[int]float64{
		0: 0,
		1: 1 * match.BaseMoney,
		//		2: 3 * match.BaseMoney,
		//		3: 5 * match.BaseMoney,
		//		4: 15 * match.BaseMoney,
	}
	match.TurnStartedTime = time.Now()
	match.MovesLog = make([]*Move, 0)
	match.ChanMove = make(chan *Move)
	match.ChanMoveErr = make(chan error)
	match.Mutex.Unlock()
	//
	match.UpdateMatch(multiplayer.COMMAND_MATCH_START)
	for i := 0; i < 1; i++ {
		turnTimeout := time.After(DURATION_TURN)
	LoopWaitingLegalMove:
		for {
			select {
			case move := <-match.ChanMove:
				err := match.MakeMove(move)
				if err == nil {
					match.TurnStartedTime = time.Now()
				}
				select {
				case match.ChanMoveErr <- err:
				default:
				}
				if err == nil {
					break LoopWaitingLegalMove
				}
			case <-turnTimeout:
				match.TurnStartedTime = time.Now()
				break LoopWaitingLegalMove
			}
		}
		match.UpdateMatch(multiplayer.COMMAND_MATCH_UPDATE)
	}
	//
	match.ResultChangedMoney = match.UserWonMoney - match.UserLostMoney
	match.ResultDetail = match.String()
	match.Game.FinishMatch(match)
	match.UpdateMatch(multiplayer.COMMAND_MATCH_FINISH)
}

func (m *CarMatch) SendMove(data map[string]interface{}) error {
	move := &Move{
		UserId:      misc.ReadInt64(data, "UserId"),
		CarIndex:    int(misc.ReadInt64(data, "CarIndex")),
		Bet:         misc.ReadFloat64(data, "Bet"),
		CreatedTime: time.Now()}
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
		return errors.New("m.ChanMove <- move timeout")
	}
}

// calc payRate from random number
// input = rand.Intn(10000)
func calcRate(r int) float64 {
	var rate float64
	switch {
	case (0 <= r) && (r < 3000):
		rate = 1.25
	case 3000 <= r && r < 5500:
		rate = 1.1
	case 5500 <= r && r < 7500:
		rate = 0.5
	case 7500 <= r && r < 9000:
		rate = 0.2
	case 9000 <= r && r < 9999:
		rate = 2.1
	case r == 9999:
		rate = 100
	}
	return rate
}

func calcAvarageProfit() float64 {
	ap := float64(0)
	for i := 0; i < 10000; i++ {
		ap += 0.0001 * calcRate(i)
	}
	return ap
}

func (m *CarMatch) MakeMove(move *Move) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	user, err := users.GetUser(move.UserId)
	if user == nil {
		return err
	}
	m.MovesLog = append(m.MovesLog, move)
	requiringMoney := move.Bet
	_, err = users.ChangeUserMoney(move.UserId, m.MoneyType, -requiringMoney,
		users.REASON_PLAY_GAME, true)
	if err != nil {
		return err
	}
	//
	r := rand.Intn(10000)
	rate := calcRate(r)
	wonMoney := rate * requiringMoney
	wonMoney = zglobal.GameEggPayoutRate * wonMoney
	users.ChangeUserMoney(move.UserId, m.MoneyType, wonMoney,
		users.REASON_PLAY_GAME, false)
	m.UserLostMoney += requiringMoney
	m.UserWonMoney += wonMoney
	return nil
}
