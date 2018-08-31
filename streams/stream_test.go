package streams

import (
	//	"encoding/json"
	"fmt"
	"testing"
	"time"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/nbackend"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	_ = time.Sleep
	nbackend.InitBackend(nil)

	stream, e := CreateStream(2, "stream cua 2", "anh2")
	if e != nil {
		t.Error(e)
	}
	stream2, e2 := ViewStream(3, 2)
	if e2 != nil {
		t.Error(e2)
	}
	if stream != stream2 {
		t.Error()
	}
	e = FinishStream(3)
	if e.Error() != l.Get(l.M028StreamNotBroadcasting) {
		t.Error()
	}
	_, e3 := ViewStream(6, 2)
	_, e4 := ViewStream(5, 2)
	e5 := ReportStream(5, 2, "Con gái đéo gì cởi trần")
	e6 := StopViewingStream(5)
	//	fmt.Println("stream", stream)
	if e3 != nil || e4 != nil || e5 != nil || e6 != nil {
		t.Error(e3, e4, e5, e6)
	}
	//
	CreateStream(4, "stream cua 4", "anh4")
	ViewStream(5, 4)
	ViewStream(7, 4)
	ViewStream(8, 4)
	ViewStream(9, 4)
	streams := StreamAllSummaries(false)
	if misc.ReadFloat64(streams[0], "BroadcasterId") != 4 ||
		misc.ReadFloat64(streams[0], "NViewers") != 5 {
		t.Error()
	}
	// fmt.Println("streams", streams)
	fmt.Println(MapUserIdToStream[4].ToMap())

	e = FinishStream(2)
	FinishStream(4)
	if e != nil {
		t.Error()
	}
	//	fmt.Println("FinishStream(2) e", e)
	time.Sleep(200 * time.Millisecond)
}
