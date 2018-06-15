package main

import (
	"fmt"
	"net"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: %s", err.Error())
		os.Exit(1)
	}
}

func recvUDPMsg(conn *net.UDPConn) {
	var buf [20]byte

	n, raddr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return
	}

	fmt.Println("msg is ", string(buf[0:n]))

	//WriteToUDP
	//func (c *UDPConn) WriteToUDP(b []byte, addr *UDPAddr) (int, error)
	_, err = conn.WriteToUDP([]byte("nice to see u"), raddr)
	checkError(err)
}

func main() {

	a := "5"
	b := "5"
	fmt.Println(&a, " ", &b)
	//go recvUDPMsg(conn)

}
