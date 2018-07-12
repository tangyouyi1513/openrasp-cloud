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
	grpc.InitRpc()
	// 获取本机的MAC地址

}
