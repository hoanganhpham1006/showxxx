package singleplayer

import (
	"testing"
)

func Test01(t *testing.T) {
	game := &Game{}
	game.Init("slot", 10000)
	match := &Match{}
	game.InitMatch(2, match)
	e := match.Archive()
	if e != nil {
		t.Error(e)
	}
}
