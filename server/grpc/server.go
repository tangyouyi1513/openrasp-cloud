package grpc

import (
	"net"
	"google.golang.org/grpc"

	pb "openrasp-cloud/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	. "openrasp-cloud/config"
	"fmt"
)

var rpcServer *OpenraspServer

type OpenraspServer struct {
	heartbeatCache map[string]*pb.OpenRASP_SubscribeUpdateServer
}

func (*OpenraspServer) HeartBeat(ctx context.Context,info *pb.HeartBeatInfo) (response *pb.HeartBeatResponse, err error) {
	fmt.Println("heartbeat")
	handleHeartbeat(&info)
	return
}

func (*OpenraspServer) SubscribeUpdate(agent *pb.Agent, stream pb.OpenRASP_SubscribeUpdateServer) error {

	handleHeartbeat(&stream)
	return nil
}

func (*OpenraspServer) Register(ctx context.Context, agent *pb.Agent) (response *pb.RegistrationResponse, err error) {
	println(agent.Id)
	response = &pb.RegistrationResponse{IsSuccess: true, Message: ""}
	return
}

func (*OpenraspServer) AddRasp(context.Context, *pb.Rasp) (response *pb.AddRaspResponse, err error) {
	return
}

func InitRpc() {
	lis, err := net.Listen("tcp", ":"+Conf.RpcPort)
	if err != nil {
		log.WithError(err).Panicf("grpc failed to listen: %v", err)
	}
	server := grpc.NewServer()
	rpcServer = &OpenraspServer{}
	pb.RegisterOpenRASPServer(server, rpcServer)
	if err := server.Serve(lis); err != nil {
		log.WithError(err).Panicf("grpc failed to serve: %v", err)
	}
}
