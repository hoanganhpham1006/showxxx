// Package zconfig contains global consts
package zconfig

import (
	"fmt"
	"time"
)

const (
	PostgresUsername     = "vic_user"
	PostgresPassword     = "123qwe"
	PostgresDatabaseName = "live_stream"
	PostgresAddress      = "127.0.0.1:5432"

	PostgresInitTablesFile = "/home/tungdt/go/src/github.com/daominah/livestream/zdatabase/init.sql"

	IsDeveloping = true

	WebsocketPort = ":20001"
	HttpPort      = ":20002"
	SocketIoPort  = ":20003"

	ProfilePort = ":20000"

	StaticHost         = "127.0.0.1"
	StaticUploadPort   = ":20891"
	StaticUploadPath   = "/hohohaha"
	StaticDownloadPort = "20892" // this port dont have ":" before the number
	StaticFolder       = "/home/tungdt/go/src/github.com/daominah/livestream_static"

	//
	// you rarely need to config below vars

	WebsocketMaxMessageSize = int64(65536)
	WebsocketWriteWait      = 60 * time.Second
	WebsocketReadWait       = 60 * time.Second
	WebsocketPingPeriod     = WebsocketReadWait * 9 / 10

	LANG_VIETNAMESE = "LANG_VIETNAMESE"
	LANG_ENGLISH    = "LANG_ENGLISH"
)

var Language = LANG_ENGLISH
var DefaultFutureTime, _ = time.Parse(time.RFC3339, "9999-01-01T00:00:00+07:00")

var Test = int64(5)

func init() {
	_ = fmt.Println
	// fmt.Println("DefaultFutureTime", DefaultFutureTime)
	Test += 5
}
