package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/rmdevio/httpserver/internal/headers"
	"github.com/rmdevio/httpserver/internal/request"
	"github.com/rmdevio/httpserver/internal/response"
	"github.com/rmdevio/httpserver/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := response.RespondOK()
		statusCode := response.StatusOk

		if req.RequestLine.RequestTarget == "/yourproblem" {
			body = response.RespondBadRequest()
			statusCode = response.StatusBadRequest
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body = response.RespondInternalServerError()
			statusCode = response.StatusInternalServerError
		} else if req.RequestLine.RequestTarget == "/httpbin/stream/100" {
			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				body = response.RespondInternalServerError()
				statusCode = response.StatusInternalServerError
			} else {
				defer res.Body.Close()
				w.WriteStatusLine(response.StatusOk)

				h.Remove("Content-Length")
				h.Set("Transfer-Encoding", "chunked")
				h.Replace("Content-Type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(h)

				fullBody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}

					fullBody = append(fullBody, data[:n]...)
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n\r\n"))
				tailers := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				tailers.Set("X-Content-SHA256", toStr(out[:]))
				tailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))

				w.WriteHeaders(tailers)
				return
			}
		} else if req.RequestLine.RequestTarget == "/video" {
			file, err := os.Open("./assets/video.mp4")
			if err != nil {
				h.Replace("Content-Type", "text/html")
				body = response.RespondInternalServerError()
				statusCode = response.StatusInternalServerError
			} else {
				defer file.Close()

				info, err := file.Stat()
				if err != nil {
					h.Replace("Content-Type", "text/html")
					body = response.RespondInternalServerError()
					statusCode = response.StatusInternalServerError
				} else {
					h.Replace("Content-Type", "video/mp4")
					h.Replace("Content-Length", strconv.FormatInt(info.Size(), 10))

					w.WriteStatusLine(response.StatusOk)
					w.WriteHeaders(h)

					w.WriteChunkedBody(file)
					return
				}
			}
		}

		h.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteStatusLine(statusCode)
		w.WriteHeaders(h)
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

func toStr(byteSlice []byte) string {
	out := ""
	for _, b := range byteSlice {
		out += fmt.Sprintf("%02x", b)
	}

	return string(out)
}
