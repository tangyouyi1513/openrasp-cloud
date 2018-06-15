package config

import "time"

const (
	RpcAddress      = "localhost:50051"
	HttpAddress     = "http://localhost:8080"
	UDPPort         = ":22223"
	LogMaxAge       = 7 * 24 * time.Hour
	LogRotationTime = 24 * time.Hour
	LogFileName     = "openrasp_client.log"
)
