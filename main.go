package main

import (
	"bytes"
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
	header := packets.NewHeader(1, 22, uint8(packets.Data))
	//packet := packets.NewInfoPacket(15, "Hello World!")
	packet := packets.NewFinalizePacket([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 1, 2, 3, 4, 5, 6, 7})
	built := append(header.ToBytes(), packet.ToBytes()...)
	//

	fin := make(chan int)

	go func() {
		buf := make([]byte, 512)
		n, _, _, _, err := conn.ReadMsgUDP(buf, nil)
		if err != nil {
			panic(err)
		}
		dataReader := bytes.NewReader(buf[:n])

		//
		hdr, err := packets.ParseHeader(dataReader)
		fmt.Printf("%+v\n", hdr)
		pkt, err := packets.ParseFinalizePacket(dataReader)
		fmt.Printf("%+v\n", pkt)
		//

		fin <- 0
	}()

	err = send(conn, built)
	if err != nil {
		panic(err)
	}
	<-fin
}

func send(conn *net.UDPConn, data []byte) error {
	_, _, err := conn.WriteMsgUDP(data, nil, nil)
	if err != nil {
		panic(err)
	}

	return nil
}
