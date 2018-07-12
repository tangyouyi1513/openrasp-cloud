package main

import (
	"openrasp-cloud/client/udp"
	"sync"
	"openrasp-cloud/client/grpc"
)

var wg sync.WaitGroup
func main() {
	grpc.InitRpc()
	udp.InitUDP()

	wg.Add(1)
	wg.Wait()
}