package main

import (
	"net"
	"satae66.dev/netzeps2022/network"
	"satae66.dev/netzeps2022/network/packets"
)

func main() {
	remoteAddr := net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: 6969,
		Zone: "",
	}

	conn, err := net.DialUDP("udp", nil, &remoteAddr)
	if err != nil {
		panic(err)
	}

	//
	header := network.NewHeader(1, 10, uint8(packets.Info))
	packet := packets.NewInfoPacket(header, 15, "Hello World!")
	//

	err = send(conn, packet)
	if err != nil {
		panic(err)
	}
}

func send(conn *net.UDPConn, packet packets.Packet) error {
	_, _, err := conn.WriteMsgUDP(packet.ToBytes(), nil, nil)
	if err != nil {
		panic(err)
	}

	return nil
}
