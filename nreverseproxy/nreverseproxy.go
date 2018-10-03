package main

import (
	"github.com/daominah/livestream/nbackend"
)

func main() {
	proxy := nbackend.CreateProxy()
	go proxy.ConnectToBackend()
	isTls, certFile, keyFile := true, "cert.pem", "key.pem"
	go proxy.ListenToClients(isTls, certFile, keyFile)
	select {}
}
