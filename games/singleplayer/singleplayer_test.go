package singleplayer

import (
	"testing"

	"github.com/daominah/livestream/users"
)

func Test01(t *testing.T) {
	game := &Game{}
	game.Init("slot", users.MT_CASH, 10000)
	match := &Match{}
	game.InitMatch(2, match)
	e := match.Archive()
	if e != nil {
		t.Error(e)
	}
}
