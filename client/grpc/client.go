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
	registerRetryCount = 100
)

var client pb.OpenRASPClient
var heartbeatStream pb.OpenRASP_HeartBeatClient
var isConnecting = false
var agent *pb.Agent

func InitRpc() {
	conn, err := grpc.Dial(Conf.RpcAddress, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Panicf("failed to connect to address %s : %v", Conf.RpcAddress, err)
	}
	//defer conn.Close()

	client = pb.NewOpenRASPClient(conn)
	register()
	go startHeartbeat()
	//wait := make(chan struct{})
	//<-wait
}

// agent 注册
func register() {
	retryCount := 0
	for retryCount < registerRetryCount {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		agent = getAgent()
		result, err := client.Register(ctx, agent)
		if err != nil {
			log.WithError(err).Error("failed to register to grpc server")
		} else {
			if result.IsSuccess {
				log.Info("success to register to grpc server")
				cancel()
				return
			} else {
				log.WithError(err).Error("failed to register to grpc server")
			}
		}
		cancel()
		retryCount++
		time.Sleep(registerRetryDelay * time.Second)
	}
	if retryCount >= registerRetryCount {
		log.Panic("failed to register to grpc server")
	}
	return
}

// 建立更新推送长连接通道
func connectHeartbeat() {
	if isConnecting {
		return
	}
	isConnecting = true
	defer func() { isConnecting = false }()
	if heartbeatStream != nil {
		heartbeatStream.CloseSend()
	}
	stream, err := client.HeartBeat(context.Background())

	for err != nil {
		log.WithError(err).Error("failed to start heartbeat, will be retry after 10 second")
		stream, err = client.HeartBeat(context.Background())
		time.Sleep(10 * time.Second)
	}
	heartbeatStream = stream
}

// 开始心跳
func startHeartbeat() {

	defer func() {
		if err := recover(); err != nil {
			log.WithError(err.(error)).Error("heart beat failed")
			connectHeartbeat()
		}
	}()

	connectHeartbeat()
	go func() {
		for {
			updateInfo, err := heartbeatStream.Recv()
			if err != nil {
				log.WithError(err).Error("failed to receive update info")
				connectHeartbeat()
			}
			updateRasp(updateInfo)
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
	var err error
	if result != nil {
		// 升级结果返回心跳
		err = heartbeatStream.Send(&pb.HeartBeatInfo{Id: agent.Id, UpdateResult: result})
	} else {
		// 统计信息心跳
		sumData := <-data.Sum
		defer func() {
			data.Sum <- sumData
		}()
		err = heartbeatStream.Send(&pb.HeartBeatInfo{Id: agent.Id, SumData: sumData})
		if err == nil {
			sumData = make(map[string]*pb.SumData)
		}
	}

	if err != nil {
		log.WithError(err).Error("failed to send heartbeat, reconnect heartbeat")
		connectHeartbeat()
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

// 通过 mac 地址和当前文件夹得到id
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

// 更新 rasp 插件和配置
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
