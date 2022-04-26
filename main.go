package main

import (
	"net"
	"os"
)

func main() {
	remoteAddr := net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: 6969,
		Zone: "",
	}

	r, err := NewReceiver(512, "down", 10, &remoteAddr)
	if err != nil {
		panic(err)
	}

	fin := make(chan bool, 1)

	go func() {
		err = r.ReceiverFile()
		if err != nil {
			panic(err)
		}
		fin <- true
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

	<-fin
}
