package main

import (
	"fmt"
	//	"time"

	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	//	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/connections"
	"github.com/daominah/livestream/zdatabase"
	_ "github.com/daominah/livestream/zglobal"
)

func init() {
	fmt.Println("")
	_ = zdatabase.DbPool
}

func main() {
	// app profile
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	runtime.SetBlockProfileRate(1)

	//	zdatabase.InitTables()
	connections.ListenAndServe(serverCommandHandler)
	select {}
}
