package request

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"

	"github.com/rmdevio/httpserver/internal/headers"
)

type parserState int

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string

	state parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	HTTP_VERSION = "1.1"

	StateInit parserState = iota
	StateParseHeader
	StateParseBody
	StateDone
)

var (
	ErrRequestLineNotFound  = errors.New("request line not found")
	ErrInvalidRequestLine   = errors.New("invalid request line")
	ErrMalformedHttpVersion = errors.New("malformed http version")
	ErrInvalidHttpVersion   = errors.New("invalid http version")
	ErrInvalidHttpMethod    = errors.New("invalid http method")

	methodRegex   = "^[A-Z]+$"
	crlfSeparator = []byte("\r\n")
)

func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == HTTP_VERSION
}

func (r *RequestLine) ValidMethod() bool {
	match, _ := regexp.MatchString(methodRegex, r.Method)
	return match
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

func getIntHeader(headers *headers.Headers, name string, defaultValue int) int {
	if valueStr := headers.Get(name); valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return defaultValue
		}

		return value
	} else {
		return defaultValue
	}
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	idx := bytes.Index(data, crlfSeparator)
	if idx == -1 {
		return RequestLine{}, 0, nil
	}

	reqLine := data[:idx]
	read := idx + len(crlfSeparator)

	parts := bytes.Split(reqLine, []byte(" "))
	if len(parts) != 3 {
		return RequestLine{}, 0, ErrInvalidRequestLine
	}

	var requestLine RequestLine

	// Check method contains uppercase alphabetic characters only
	requestLine.Method = string(parts[0])
	if !requestLine.ValidMethod() {
		return RequestLine{}, 0, ErrInvalidHttpMethod
	}

	// Set request target
	requestLine.RequestTarget = string(parts[1])

	// Check if HTTP version is 1.1
	httpVersionParts := bytes.Split(parts[2], []byte("/"))
	if len(httpVersionParts) != 2 {
		return RequestLine{}, 0, ErrMalformedHttpVersion
	}

	requestLine.HttpVersion = string(httpVersionParts[1])
	if !requestLine.ValidHTTP() {
		return RequestLine{}, 0, ErrInvalidHttpVersion
	}

	return requestLine, read, nil
}

func (r *Request) hasBody() bool {
	length := getIntHeader(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break
		}

		switch r.state {
		case StateInit:
			reqLine, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = reqLine
			read += n

			r.state = StateParseHeader
		case StateParseHeader:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateParseBody
				} else {
					r.state = StateDone
				}
			}

		case StateParseBody:
			length := getIntHeader(r.Headers, "content-length", 0)
			if length == 0 {
				panic("not implemented")
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer
		}
	}

	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
