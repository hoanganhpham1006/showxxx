package streams

import (
	"fmt"
	"testing"
	"time"

	l "github.com/daominah/livestream/language"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	stream, e := CreateStream(2)
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
	fmt.Println("FinishStream(3) e", e)
	if e.Error() != l.Get(l.M028StreamNotBroadcasting) {
		t.Error()
	}
	_, e3 := ViewStream(4, 2)
	_, e4 := ViewStream(5, 2)
	e5 := ReportStream(5, 2, "Con gái đéo gì cởi trần")
	e6 := StopViewingStream(5, 2)
	fmt.Println("stream", stream)
	if e3 != nil || e4 != nil || e5 != nil || e6 != nil {
		t.Error(e3, e4, e5, e6)
	}
	//
	e = FinishStream(2)
	if e != nil {
		t.Error()
	}
	fmt.Println("FinishStream(2) e", e)
	time.Sleep(200 * time.Microsecond)
}
