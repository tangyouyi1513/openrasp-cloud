package udp

import (
	"net"
	. "openrasp-cloud/config"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	sum "openrasp-cloud/client/data"
	"fmt"
	"openrasp-cloud/proto"
	"openrasp-cloud/client/grpc"
)

func checkError(err error) {
	if err != nil {
		log.WithError(err).Panicf("init udp failed")
	}
}

func recvUDPMsg(conn *net.UDPConn) {
	defer conn.Close()
	for {
		handleUDPMsg(conn)
	}
}

func handleUDPMsg(conn *net.UDPConn) {

	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(error); ok {
				log.WithError(err).Error("handle UDP message failed")
			}
		}
	}()

	buf := make([]byte, 2048)
	count, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.WithError(err).Errorf("receive message failed %s", addr.String())
	} else if count > 0 {
		data := make(map[string]interface{})
		err = json.Unmarshal(buf[0:count], &data)
		if err != nil {
			log.WithError(err).Errorf("decode json failed")
			return
		}
		fmt.Println(string(buf), len(buf))
		basicInfo := data["info"].(map[string]interface{})
		id := basicInfo["id"].(string)
		isNewRasp := false
		startTime := uint64(basicInfo["start_time"].(float64))
		if rasp, ok := sum.RaspBasicData[id]; ok {
			if rasp.StartTime != startTime {
				isNewRasp = true
			}
		} else {
			isNewRasp = true
		}

		if isNewRasp {
			rasp := &proto.Rasp{
				Id:              id,
				Type:            basicInfo["type"].(string),
				Version:         basicInfo["rasp_version"].(string),
				LanguageVersion: basicInfo["language_version"].(string),
				RaspHome:        basicInfo["rasp_home"].(string),
				StartTime:       startTime,
			}

			response, err := grpc.AddRasp(rasp)
			if err != nil || response == nil || !response.IsSuccess {
				log.WithError(err).Error("register rasp error")
				return
			} else {
				sum.RaspBasicData[id] = rasp
			}
		}

		sumData := data["sum"].(map[string]interface{})
		sum.AddData(id, sumData)
	}
}

func InitUDP() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+Conf.UDPPort)
	checkError(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)
	go recvUDPMsg(conn)
}
