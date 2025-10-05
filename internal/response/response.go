package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/rmdevio/httpserver/internal/headers"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) Writer {
	return Writer{
		writer: writer,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var statusReason string

	switch statusCode {
	case StatusOk:
		statusReason = "OK"
	case StatusBadRequest:
		statusReason = "Bad Request"
	case StatusInternalServerError:
		statusReason = "Internal Server Error"
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusReason)

	_, err := w.writer.Write([]byte(statusLine))

	return err
}

func (w *Writer) WriteHeaders(h *headers.Headers) error {
	var err error
	h.ForEach(func(name, value string) {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s:%s\r\n", name, value)))
		if err != nil {
			return
		}
	})
	w.writer.Write([]byte("\r\n"))

	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}

func (w *Writer) WriteChunkedBody(reader io.Reader) (int64, error) {
	n, err := io.Copy(w.writer, reader)
	return n, err
}

func GetDefaultHeaders(contentLength int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLength))
	h.Set("Content-Type", "text/html")

	return h
}

func RespondOK() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func RespondBadRequest() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func RespondInternalServerError() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}
