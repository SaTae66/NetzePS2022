package main

import (
	"net"
)

func main() {
	remoteAddr := net.UDPAddr{
		IP:   net.ParseIP("10.3.3.140"),
		Port: 6969,
		Zone: "",
	}
	/*
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
	*/

	r, err := NewReceiver(512, "down", 10, &remoteAddr)
	if err != nil {
		panic(err)
	}

	err = r.ReceiverFile()
	if err != nil {
		panic(err)
	}
}
