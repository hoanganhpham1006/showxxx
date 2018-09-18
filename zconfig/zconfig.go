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
	// PostgresAddress        = "127.0.0.1:12354"
	PostgresInitTablesFile = "/Users/phamhoanganh/go/src/github.com/daominah/livestream/zdatabase/init.sql"

	BackendIp   = "127.0.0.1"
	BackendPort = ":20004"

	IsDeveloping       = true
	BackendProfilePort = ":20000"

	ProxyPort           = ":20001"
	AdminToolPort       = ":20002"
	WebRTCSignalingPort = ":20003"
	IPNPort             = ":20005"

	StaticHost         = "127.0.0.1"
	StaticUploadPort   = ":20891"
	StaticUploadPath   = "/hohohaha"
	StaticDownloadPort = "20892" // this port dont have ":" before the number
	StaticFolder       = "/Users/phamhoanganh/go/src/github.com/daominah/livestream_static"

	//
	// you rarely need to config below vars

	WebsocketMaxMessageSize = int64(65536)
	WebsocketWriteWait      = 60 * time.Second
	WebsocketReadWait       = 60 * time.Second
	WebsocketPingPeriod     = WebsocketReadWait * 9 / 10

	LimitNConnsPerIp        = 10 // limit number of connections per ip address
	LimitNRequestsPerSecond = 10 // limit number of request per second of a connection

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

func TPrint(a ...interface{}) {
	if IsDeveloping {
		fmt.Println(a...)
	}
}

func TPrintf(format string, a ...interface{}) {
	if IsDeveloping {
		fmt.Printf(format, a...)
	}
}
