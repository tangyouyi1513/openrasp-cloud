package grpc

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"openrasp-cloud/client/config"
	pb "openrasp-cloud/proto"
	"os"
	"runtime"
	log "github.com/sirupsen/logrus"
)

var token string

type OpenRaspClientAgent interface {
}

func InitAgent() {
	conn, err := grpc.Dial(config.RpcAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to address %s : %v", config.RpcAddress, err)
	}
	defer conn.Close()

	client := pb.NewOpenRASPClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	registerResult, err := client.Register(ctx, getAgent())
	if err != nil {
		log.Fatalf("failed to register to grpc server")
	}
	token = registerResult.Token
	go startHeartBeat(client)
}

func startHeartBeat(client pb.OpenRASPClient) {

}

//func AddRasp(rasp pb.Rasp) pb.AddRaspResponse {
//
//}
//
//func HeartBeat(info pb.HeartBeatInfo) pb.UpdateInfo {
//
//}

func getAgent() *pb.Agent {
	hostName, err := os.Hostname()
	if err != nil {
		log.Fatalln("failed to get hostname, use empty string instead")
		hostName = ""
	}
	return &pb.Agent{HostName: hostName, Os: runtime.GOOS}
}
