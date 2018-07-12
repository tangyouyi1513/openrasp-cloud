package grpc

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "openrasp-cloud/proto"
	"os"
	"runtime"
	log "github.com/sirupsen/logrus"
	. "openrasp-cloud/config"
	"openrasp-cloud/client/data"
	"sort"
	"openrasp-cloud/tools"
	"crypto/md5"
	"encoding/hex"
	"path"
	"io/ioutil"
	"fmt"
)

const (
	// 注册失败重试时间间隔
	registerRetryDelay = 5
	// 注册失败重试次数
	registerRetryCount = 3
)

var client pb.OpenRASPClient
var updateStream pb.OpenRASP_SubscribeUpdateClient
var isSubscribing = false
var agent *pb.Agent

func InitRpc() {
	conn, err := grpc.Dial(Conf.RpcAddress, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Panicf("failed to connect to address %s : %v", Conf.RpcAddress, err)
	}
	defer conn.Close()

	client = pb.NewOpenRASPClient(conn)
	register()
	subscribeUpdate()
	go startHeartbeat()
	wait := make(chan struct{})
	<-wait
}

// agent 注册
func register() {

	defer func() {
		if err := recover(); err != nil {
			log.WithError(err.(error)).Panicf("failed to register to grpc server")
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	retryCount := 0
	for retryCount < registerRetryCount {
		agent = getAgent()
		result, err := client.Register(ctx, agent)
		if err != nil {
			log.WithError(err).Panicf("failed to register to grpc server")
		} else {
			if result.IsSuccess {
				break
			} else {
				log.WithError(err).Panicf("failed to register to grpc server")
			}
		}
		retryCount++
		time.Sleep(registerRetryDelay * time.Second)
	}

	return
}

// 订阅更新推送
func subscribeUpdate() {
	defer func() {
		if err := recover(); err != nil {
			log.WithError(err.(error)).Error("heart beat failed")
			connectUpgrade()
		}
	}()
	connectUpgrade()
	go func() {
		updateInfo, err := updateStream.Recv()
		if err != nil {
			log.WithError(err).Error("failed to receive update info")
			connectUpgrade()
		}
		updateRasp(updateInfo)
	}()
}

// 建立更新推送长连接通道
func connectUpgrade() {
	if isSubscribing {
		return
	}
	isSubscribing = true
	defer func() { isSubscribing = false }()
	if updateStream != nil {
		updateStream.CloseSend()
	}
	stream, err := client.SubscribeUpdate(context.Background(), agent)

	for err != nil {
		log.WithError(err).Error("failed to start heartbeat, will be retry after 10 second")
		stream, err = client.SubscribeUpdate(context.Background(), agent)
		time.Sleep(10 * time.Second)
	}
	updateStream = stream
}

// 开始心跳
func startHeartbeat() {
	defer func() {
		if err := recover(); err != nil {
			log.WithError(err.(error)).Error("heart beat failed")
		}
	}()

	for {
		Heartbeat(nil)
		time.Sleep(time.Duration(Conf.HeartbeatDelay) * time.Second)
	}
}

// 添加 rasp
func AddRasp(rasp *pb.Rasp) (*pb.AddRaspResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return client.AddRasp(ctx, rasp)
}

// 心跳
func Heartbeat(result *pb.UpdateResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var err error
	var response *pb.HeartBeatResponse
	if result != nil {
		// 升级结果返回心跳
		response, err = client.HeartBeat(ctx, &pb.HeartBeatInfo{Id: agent.Id, UpdateResult: result})
	} else {
		// 统计信息心跳
		sumData := <-data.Sum
		defer func() {
			data.Sum <- sumData
		}()
		response, err = client.HeartBeat(ctx, &pb.HeartBeatInfo{Id: agent.Id, SumData: sumData})
		if err != nil {
			return
		}
	}

	if response != nil {
		if !response.IsSuccess {
			log.Errorf("failed to send heartbeat, reconnect heartbeat %s", response.Message)
		} else {
			log.Info("heartbeat success")
		}
	} else {
		log.WithError(err).Error("failed to send heartbeat, reconnect heartbeat")
	}

	if err != nil {
		log.WithError(err).Error("failed to send heartbeat, reconnect heartbeat")
	}

	return
}

func getAgent() *pb.Agent {
	hostName, err := os.Hostname()
	if err != nil {
		log.WithError(err).Error("failed to get hostname, use empty string instead")
		hostName = ""
	}
	return &pb.Agent{HostName: hostName, Os: runtime.GOOS, Id: getId()}
}

func getId() (id string) {
	allMac, err := tools.GetMacAddrs()
	if err != nil {
		log.WithError(err).Panic("failed to get mac address")
	}
	sort.Sort(sort.StringSlice(allMac))
	id = ""
	for _, mac := range allMac {
		id += mac
	}
	currentPath, err := tools.GetCurrentPath()
	if err != nil {
		log.WithError(err).Panic("failed to get current path")
	}
	id += currentPath
	md5Bytes := md5.Sum([]byte(id))
	id = string(hex.EncodeToString(md5Bytes[:]))
	return
}

func updateRasp(info *pb.UpdateInfo) {
	if info != nil {
		if info.RaspId != nil {
			for _, id := range info.RaspId {
				if rasp, ok := data.RaspBasicData[id]; ok {
					exist, err := tools.PathExists(rasp.RaspHome)
					if err != nil {
						log.WithError(err).Errorf("failed to update rasp: %s", id)
						continue
					}
					if !exist {
						log.Errorf("failed to update rasp %s, the rasp home does not exist", id)
						continue
					}
					var errMessage string
					if rasp.Type == "java" {
						err, errMessage = updateJava(info, rasp)
					} else if rasp.Type == "php" {
						err, errMessage = updatePHP(info, rasp)
					} else {
						log.Errorf("failed to update rasp %s, the rasp type does not support: %s", id, rasp.Type)
						continue
					}
					if err != nil {
						Heartbeat(&pb.UpdateResult{
							info.Id,
							false,
							errMessage,
						})
					}
				} else {
					log.Errorf("failed to update rasp %s, the rasp does not exist", id)
				}
			}
		}
	}
}

func updateJava(info *pb.UpdateInfo, rasp *pb.Rasp) (err error, errorMessage string) {
	err = nil
	confDirPath := path.Join(rasp.RaspHome, "conf")
	if exist, _ := tools.PathExists(confDirPath); !exist {
		err = os.Mkdir(confDirPath, os.ModePerm)
		if err != nil {
			errorMessage = fmt.Sprintf("failed to create config dir for rasp: %s", rasp.Id)
			return
		}
	}

	if info.PropertiesConf != nil && info.PropertiesConf.Content != nil {
		confFilePath := path.Join(confDirPath, "rasp.properties")
		err = ioutil.WriteFile(confFilePath, info.PropertiesConf.Content, 0666)
		if err != nil {
			errorMessage = fmt.Sprintf("failed to write rasp.properties for rasp: %s", rasp.Id)
			return
		}

	}

	pluginDirPath := path.Join(rasp.RaspHome, "plugins")
	err = os.Remove(pluginDirPath)
	if err != nil {
		errorMessage = fmt.Sprintf("failed to remove plugin dir for rasp: %s", rasp.Id)
		return
	}
	if exist, _ := tools.PathExists(pluginDirPath); !exist {
		err = os.Mkdir(pluginDirPath, os.ModePerm)
		if err != nil {
			errorMessage = fmt.Sprintf("failed to create plugin dir for rasp: %s", rasp.Id)
			return
		}
	}

	if info.Plugins != nil {
		for _, plugin := range info.Plugins {
			pluginFilePath := path.Join(pluginDirPath, plugin.Name)
			err = ioutil.WriteFile(pluginFilePath, plugin.Content, 0666)
			if err != nil {
				errorMessage = fmt.Sprintf("failed to write plugin %s for rasp: %s", plugin.Name, rasp.Id)
				return
			}
		}
	}
	return
}

func updatePHP(info *pb.UpdateInfo, rasp *pb.Rasp) (err error, errorMessage string) {
	// TODO: PHP

	return
}
