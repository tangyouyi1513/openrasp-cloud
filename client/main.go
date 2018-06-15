package main

import (
	"openrasp-cloud/client/grpc"
	"openrasp-cloud/client/udp"
	"sync"
	"openrasp-cloud/client/tools"
)

var wg sync.WaitGroup

func main() {
	tools.InitLogger()
	grpc.InitAgent()
	udp.InitUDP()

	wg.Add(1)
	wg.Wait()
}
