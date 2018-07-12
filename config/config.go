package config

import (
	"encoding/json"
	"openrasp-cloud/tools"
	"path"
	log "github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	RpcAddress      string `json:"rpc_address"`
	UDPPort         string `json:"udp_port"`
	LogMaxAge       uint64 `json:"log_max_age"`       // 日志保存最长时间，单位/天
	LogRotationTime uint64 `json:"log_rotation_time"` // 日志 rotate 间隔，单位/小时
	LogFileName     string `json:"log_file_name"`
	HeartbeatDelay  uint64 `json:"heartbeat_dalay"` // 心跳时间间隔，单位/秒
	HttpPort        string `json:"http_port"`
	RpcPort         string `json:"rpc_port"`
	MongoAddress    string `json:"mongo_address"`
	MongoUser       string `json:"mongo_user"`
	MongoPassword   string `json:"mongo_password"`
	Token           string `json:"token"`
}

var Conf Config

const (
	UTF8 = string("UTF-8")
)

func NewConfig(configMessage string) *Config {
	Conf = Config{
		RpcAddress:      "localhost:50051",
		UDPPort:         "22223",
		LogMaxAge:       7,
		LogRotationTime: 24,
		LogFileName:     "openrasp_client.log",
		HeartbeatDelay:  60,
		RpcPort:         "50051",
		HttpPort:        "8201",
		MongoAddress:    "localhost:27017",
		MongoUser:       "",
		MongoPassword:   "",
	}
	if configMessage != "" {
		var data = []byte(configMessage)
		json.Unmarshal(data, &Conf)
	}
	configData, _ := json.Marshal(Conf)
	log.Infof("config: %s", configData)
	return &Conf
}

func handleError(message string, err error) {
	if err != nil {
		log.WithError(err).Panic("failed to init config: " + message)
	}
}

func init() {
	currentPath, err := tools.GetCurrentPath()
	handleError("can not get current path", err)
	configDir := path.Join(currentPath, "config")
	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		os.Mkdir(configDir, os.ModePerm)
	}
	configPath := path.Join(configDir, "client.conf")
	_, err = os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	handleError("can not get config file", err)
	data, err := tools.ReadFromFile(configPath)
	handleError("can not read config file", err)
	Conf = *NewConfig(string(data))
}
