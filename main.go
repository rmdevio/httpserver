package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	workingDir, _ := os.Getwd()
	file, err := os.Open(filepath.Join(workingDir, "messages.txt"))
	if err != nil {
		panic(err)
	}

	linesChan := getLinesThroughChannel(file)
	for line := range linesChan {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesThroughChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
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
