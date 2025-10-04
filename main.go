package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("error while accepting connection: %s", err)
			continue
		}

		fmt.Printf("Accepted new connection: %s\n", conn.RemoteAddr().String())
		linesChan := getLinesThroughChannel(conn)
		for line := range linesChan {
			fmt.Printf("read: %s\n", line)
		}
		fmt.Printf("Channel closed for connection: %s\n", conn.RemoteAddr().String())
	}

}

func getLinesThroughChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer close(out)

		var bufStr string
		for {
			buf := make([]byte, 8)
			n, err := f.Read(buf)
			if err != nil {
				break
			}

			data := buf[:n]
			if idx := bytes.IndexByte(data, '\n'); idx != -1 {
				bufStr += string(data[:idx])
				data = data[idx+1:]
				out <- bufStr
				bufStr = ""
			}

			bufStr += string(data)
		}

		if len(bufStr) != 0 {
			out <- bufStr
		}
	}()

	return out
}
