package main

import (
	"fmt"
	"os"
	"openrasp-cloud/server/grpc"
)

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	//logger.InitLogger()
	grpc.InitRpc()

}
