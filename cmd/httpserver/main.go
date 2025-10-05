package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/rmdevio/httpserver/internal/request"
	"github.com/rmdevio/httpserver/internal/response"
	"github.com/rmdevio/httpserver/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w response.Writer, req *request.Request) {
		headers := response.GetDefaultHeaders(0)
		body := response.RespondOK()
		statusCode := response.StatusOk

		if req.RequestLine.RequestTarget == "/yourproblem" {
			body = response.RespondBadRequest()
			statusCode = response.StatusBadRequest
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body = response.RespondInternalServerError()
			statusCode = response.StatusInternalServerError
		}

		headers.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteStatusLine(statusCode)
		w.WriteHeaders(headers)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
