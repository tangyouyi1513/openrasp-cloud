package udp

import (
	"fmt"
	"net"
	"openrasp-cloud/client/config"
	log "github.com/sirupsen/logrus"
)

func checkError(err error) {
	if err != nil {

	}
}

func recvUDPMsg(conn *net.UDPConn) {
	buf := make([]byte, 2048)
	count, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("msg is ", string(buf[0:count]))
	} else {
		log.WithError(err).Errorf("Receive message failed %s",addr.String())
	}
}

func InitUDP() {
	udpAddr, err := net.ResolveUDPAddr("udp", config.UDPPort)
	checkError(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()
	checkError(err)
	go recvUDPMsg(conn)

}
