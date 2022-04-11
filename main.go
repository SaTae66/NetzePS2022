package main

import (
	"fmt"
	"net"
	"satae66.dev/netzeps2022/network/packets"
)

func main() {
	remoteAddr := net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: 6969,
		Zone: "",
	}

	conn, err := net.DialUDP("udp", &remoteAddr, &remoteAddr)
	if err != nil {
		panic(err)
	}

	//
	header := packets.NewHeader(1, 10, uint8(packets.Info))
	packet := packets.NewInfoPacket(header, 15, "Hello World!")
	//

	fin := make(chan int)

	go func() {
		data := make([]byte, 512)
		_, _, _, _, err := conn.ReadMsgUDP(data, nil)
		if err != nil {
			panic(err)
		}
		pkt, err := packets.ParseInfoPacket(data)
		fmt.Printf("%+v\n", pkt)
		fin <- 0
	}()

	err = send(conn, packet)
	if err != nil {
		panic(err)
	}
	<-fin
}

func send(conn *net.UDPConn, packet packets.Packet) error {
	_, _, err := conn.WriteMsgUDP(packet.ToBytes(), nil, nil)
	if err != nil {
		panic(err)
	}

	return nil
}
