package nbackend

import (
	"fmt"
	"github.com/daominah/livestream/zconfig"
	"testing"

	"github.com/daominah/livestream/nwebsocket"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	backend := nwebsocket.CreateServer(999999, 999999)
	backend.ListenAndServe(zconfig.BackendPort, nil, nil)
	select {}
}
