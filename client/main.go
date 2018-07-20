package main

import (
	"openrasp-cloud/client/udp"
	"sync"
	"openrasp-cloud/client/grpc"
	"openrasp-cloud/logger"
)

var wg sync.WaitGroup

func main() {
	logger.InitLogger()
	grpc.InitRpc()
	udp.InitUDP()

	wg.Add(1)
	wg.Wait()
}
