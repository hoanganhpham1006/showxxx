package main

import (
	"github.com/daominah/livestream/nbackend"
)

func main() {
	proxy := nbackend.CreateProxy()
	go proxy.ConnectToBackend()
	go proxy.ListenToClients()
	select {}
}
