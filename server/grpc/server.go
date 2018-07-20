package grpc

import (
	"net"
	"google.golang.org/grpc"

	pb "openrasp-cloud/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	. "openrasp-cloud/config"
	"openrasp-cloud/server/mongo"
)

var raspRpcServer *RaspRpcServer

type RaspRpcServer struct {
	heartbeatCache chan map[string]pb.OpenRASP_HeartBeatServer
}

func (server *RaspRpcServer) HeartBeat(stream pb.OpenRASP_HeartBeatServer) error {
	isSubscribe := false
	for {
		heartbeatInfo, err := stream.Recv()
		if err != nil {
			log.WithError(err).Error("can not read heartbeat")
			break
		}
		if !isSubscribe {
			cache := <-server.heartbeatCache
			cache[heartbeatInfo.Id] = stream
			isSubscribe = true
			server.heartbeatCache <- cache
		}
		handleHeartbeat(heartbeatInfo)
	}

	return nil
}

func (*RaspRpcServer) Register(ctx context.Context, agent *pb.Agent) (response *pb.RegistrationResponse, err error) {
	session := mongo.Clone()
	defer session.Close()

	if agent.Id == "" {
		response = &pb.RegistrationResponse{IsSuccess: false, Message: "agent id can not be empty"}
		return
	}
	if agent.Version == "" {
		response = &pb.RegistrationResponse{IsSuccess: false, Message: "agent version can not be empty"}
		return
	}
	if agent.HostName == "" {
		response = &pb.RegistrationResponse{IsSuccess: false, Message: "agent hostname can not be empty"}
		return
	}
	if agent.Os == "" {
		response = &pb.RegistrationResponse{IsSuccess: false, Message: "agent OS can not be empty"}
		return
	}

	_, mgoErr := session.Upsert("agent", agent, agent)
	if err != nil {
		response = &pb.RegistrationResponse{IsSuccess: false, Message: mgoErr.Error()}
	} else {
		response = &pb.RegistrationResponse{IsSuccess: true, Message: ""}
	}
	return
}

func (*RaspRpcServer) AddRasp(ctx context.Context, rasp *pb.Rasp) (response *pb.AddRaspResponse, err error) {
	session := mongo.Clone()
	defer session.Close()

	if rasp.Id == "" {
		response = &pb.AddRaspResponse{IsSuccess: false, Message: "rasp id can not be empty"}
		return
	}
	if rasp.Version == "" {
		response = &pb.AddRaspResponse{IsSuccess: false, Message: "rasp version can not be empty"}
		return
	}
	if rasp.LanguageVersion == "" {
		response = &pb.AddRaspResponse{IsSuccess: false, Message: "rasp language version can not be empty"}
		return
	}
	if rasp.RaspHome == "" {
		response = &pb.AddRaspResponse{IsSuccess: false, Message: "rasp home can not be empty"}
		return
	}
	if rasp.Type == "" {
		response = &pb.AddRaspResponse{IsSuccess: false, Message: "rasp type can not be empty"}
		return
	}
	if rasp.StartTime == 0 {
		response = &pb.AddRaspResponse{IsSuccess: false, Message: "rasp start time can not be empty"}
		return
	}

	_, mgoErr := session.Upsert("rasp", rasp, rasp)
	if err != nil {
		response = &pb.AddRaspResponse{IsSuccess: false, Message: mgoErr.Error()}
	} else {
		response = &pb.AddRaspResponse{IsSuccess: true, Message: ""}
	}
	return
}

func NewRpcServer() (server *RaspRpcServer) {
	server = &RaspRpcServer{heartbeatCache: make(chan map[string]pb.OpenRASP_HeartBeatServer, 1)}
	server.heartbeatCache <- make(map[string]pb.OpenRASP_HeartBeatServer)
	return
}

func UpdateRasp(allId []string, updateInfo *pb.UpdateInfo) (result map[string]error) {
	result = make(map[string]error)
	cache := <-raspRpcServer.heartbeatCache
	defer func() { raspRpcServer.heartbeatCache <- cache }()
	for _, id := range allId {
		stream := cache[id]
		result[id] = stream.Send(updateInfo)
	}
	return
}

func InitRpc() {
	lis, err := net.Listen("tcp", ":"+Conf.RpcPort)
	if err != nil {
		log.WithError(err).Panicf("grpc failed to listen: %v", err)
	}
	server := grpc.NewServer()
	raspRpcServer = NewRpcServer()
	pb.RegisterOpenRASPServer(server, raspRpcServer)
	if err := server.Serve(lis); err != nil {
		log.WithError(err).Panicf("grpc failed to serve: %v", err)
	}
}
