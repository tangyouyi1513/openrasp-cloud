package udp

import (
	"net"
	. "openrasp-cloud/config"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"openrasp-cloud/proto"
	"openrasp-cloud/client/grpc"
	sum "openrasp-cloud/client/data"
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
			log.WithError(err.(error)).Error("handle UDP message failed")
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

		id := data["id"].(string)
		isNewRasp := false
		if rasp, ok := sum.RaspBasicData[id]; ok {
			if startTime, ok := data["start_time"]; ok && rasp.StartTime != startTime {
				isNewRasp = true
			}
		} else {
			isNewRasp = true
		}

		if isNewRasp {
			rasp := &proto.Rasp{
				Id:          id,
				Type:        data["type"].(string),
				Version:     data["rasp_version"].(string),
				JavaVersion: data["java_version"].(string),
				RaspHome:    data["rasp_home"].(string),
				StartTime:   data["start_time"].(uint64),
			}

			response, err := grpc.AddRasp(rasp)
			if err != nil || response == nil || !response.IsSuccess {
				log.WithError(err).Error("register rasp error")
				return
			} else {
				sum.RaspBasicData[id] = rasp
			}
		}

		sum.AddData(id, data)
	}
}

func InitUDP() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+Conf.UDPPort)
	checkError(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)
	go recvUDPMsg(conn)
}
