package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		panic(err)
	}

	udpConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error while reading input: %s\n", err)
			return
		}

		_, err = udpConn.Write([]byte(str))
		if err != nil {
			fmt.Printf("error while writing to connection: %s\n", err)
			return
		}
	}
}
