package headers

import (
	"bytes"
	"errors"
	"strings"
)

var (
	crlfSeparator = []byte("\r\n")
)

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: make(map[string]string),
	}
}

func isValidToken(field string) bool {
	for _, ch := range field {
		found := false
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= 0 && ch <= 9 {
			found = true
		}

		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}

	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", errors.New("malformed header")
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", errors.New("malformed field name")
	}

	return string(name), string(value), nil

}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	lcName := strings.ToLower(name)
	if existingVal, ok := h.headers[lcName]; ok {
		h.headers[lcName] = strings.Join([]string{existingVal, value}, ",")
	} else {
		h.headers[lcName] = value
	}
}

func (h *Headers) Replace(name, value string) {
	h.headers[strings.ToLower(name)] = value
}

func (h *Headers) ForEach(cb func(n, v string)) {
	for name, val := range h.headers {
		cb(name, val)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], crlfSeparator)
		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(crlfSeparator)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		if !isValidToken(name) {
			return 0, false, errors.New("malformed header name")
		}
		read += idx + len(crlfSeparator)
		h.Set(name, value)
	}

	return read, done, nil
}
