package main

import (
	"fmt"
	"net"

	"github.com/rmdevio/httpserver/internal/headers"
	"github.com/rmdevio/httpserver/internal/request"
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
		req, err := request.RequestFromReader(conn)
		if err != nil {
			continue
		}

		PrintRequestLine(req.RequestLine)
		PrintHeaders(req.Headers)
		PrintBody(req.Body)

		fmt.Printf("Channel closed for connection: %s\n", conn.RemoteAddr().String())
	}
}

func PrintRequestLine(reqLine request.RequestLine) {
	fmt.Println("Request Line:")
	fmt.Println("- Method:", reqLine.Method)
	fmt.Println("- Target:", reqLine.RequestTarget)
	fmt.Println("- Version:", reqLine.HttpVersion)
}

func PrintHeaders(headers *headers.Headers) {
	fmt.Println("Headers:")
	headers.ForEach(func(name, value string) {
		fmt.Printf("- %s: %s\n", name, value)
	})
}

func PrintBody(body string) {
	fmt.Println("Body:")
	fmt.Println(body)
}