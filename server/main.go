package main

import (
	"fmt"
	"os"
	"openrasp-cloud/server/grpc"
	"openrasp-cloud/logger"
)

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	logger.InitLogger()
	grpc.InitRpc()

}
