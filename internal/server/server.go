package server

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/rmdevio/httpserver/internal/request"
	"github.com/rmdevio/httpserver/internal/response"
)

type Server struct {
	port     uint16
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

type Handler func(w response.Writer, req *request.Request)

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	srv := &Server{
		port:     port,
		handler:  handler,
		listener: listener,
	}

	go srv.listen()

	return srv, nil
}

func (s *Server) listen() {
	for {
		if s.closed.Load() {
			return
		}

		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		fmt.Printf("Accepted new connection: %s\n", conn.RemoteAddr().String())
		go s.handle(conn)
	}
}

func (s *Server) Close() {
	s.listener.Close()
}

func (s *Server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()

	for {
		responseWriter := response.NewWriter(conn)
		req, err := request.RequestFromReader(conn)
		if err != nil {
			responseWriter.WriteStatusLine(response.StatusBadRequest)
			responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
			return
		}

		s.handler(responseWriter, req)

		connHeader := req.Headers.Get("Connection")
		if len(connHeader) != 0 && connHeader == "close" {
			break
		}
	}

	fmt.Printf("Channel closed for connection: %s\n", conn.(net.Conn).RemoteAddr().String())
}
