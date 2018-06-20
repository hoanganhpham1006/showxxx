// Package zconfig contains global consts
package zconfig

import (
	"time"
)

const (
	PostgresUsername     = "vic_user"
	PostgresPassword     = "123qwe"
	PostgresDatabaseName = "live_stream"
	PostgresAddress      = "127.0.0.1:5432"

	PostgresInitTablesFile = "/home/tungdt/go/src/github.com/daominah/livestream/zdatabase/init.sql"

	IsDeveloping = true

	WebsocketPort = ":2052"
	HttpPort      = ":2082"

	StaticHost         = "127.0.0.1"
	StaticUploadPort   = ":20891"
	StaticDownloadPort = "20892" // this port dont have ":" before the number
	StaticFolder       = "/home/tungdt/go/src/github.com/daominah/livestream_static"

	//
	// you rarely need to config below vars

	WebsocketMaxMessageSize = int64(8192)
	WebsocketWriteWait      = 60 * time.Second
	WebsocketReadWait       = 60 * time.Second
	WebsocketPingPeriod     = WebsocketReadWait * 9 / 10

	LANG_VIETNAMESE = "LANG_VIETNAMESE"
	LANG_ENGLISH    = "LANG_ENGLISH"
)

var Language = LANG_ENGLISH
