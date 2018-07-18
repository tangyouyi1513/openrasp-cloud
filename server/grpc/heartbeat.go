package grpc

import (
	"openrasp-cloud/proto"
	"fmt"
)

func handleHeartbeat(heartbeat *proto.HeartBeatInfo) {
	fmt.Printf("%+v",heartbeat.SumData)
}

