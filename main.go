package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"satae66.dev/netzeps2022/network/packets"
)

func main() {
	remoteAddr := net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: 6969,
		Zone: "",
	}

	conn, err := net.ListenUDP("udp", &remoteAddr)
	if err != nil {
		panic(err)
	}

	go func() {
	again:
		buf := make([]byte, 512)
		n, _, _, _, err := conn.ReadMsgUDP(buf, nil)
		if err != nil {
			panic(err)
		}
		dataReader := bytes.NewReader(buf[:n])

		hdr, err := packets.ParseHeader(dataReader)
		fmt.Printf("%+v\n", hdr)

		var pkt packets.Packet
		switch hdr.PacketType {
		case packets.Info:
			pkt, err = packets.ParseInfoPacket(dataReader)
			break
		case packets.Data:
			pkt, err = packets.ParseDataPacket(dataReader)
			break
		case packets.Finalize:
			pkt, err = packets.ParseFinalizePacket(dataReader)
			break
		}

		fmt.Printf("%+v\n", pkt)

		goto again
	}()

	t, err := NewTransmitter(512)
	if err != nil {
		panic(err)
	}

	f, err := os.Open("data")
	if err != nil {
		panic(err)
	}

	err = t.SendFileTo(f, &remoteAddr)
	if err != nil {
		panic(err)
	}

	var x string
	_, _ = fmt.Scan(&x)
}
