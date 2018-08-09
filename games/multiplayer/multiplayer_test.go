package multiplayer

import (
	"testing"

	"github.com/daominah/livestream/users"
)

func Test01(t *testing.T) {
	game := &Game{}
	game.Init("car", users.MT_CASH, 10000)
	match := &Match{}
	game.InitMatch(match)
	for i := int64(1); i < 5; i++ {
		match.AddUserId(i)
	}
	e := match.Archive()
	if e != nil {
		t.Error(e)
	}
}
