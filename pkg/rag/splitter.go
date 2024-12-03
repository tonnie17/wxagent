package rag

import (
	"bufio"
	"os"
	"strings"
)

func processTextFile(fileName string) (chan string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	out := make(chan string)
	go func() {
		defer f.Close()
		defer close(out)

		var buffer strings.Builder
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				out <- buffer.String()
				buffer.Reset()
			} else {
				buffer.WriteString(line + "\n")
			}
		}

		if scanner.Err() != nil {
			return
		}

		if last := strings.TrimSpace(buffer.String()); last != "" {
			out <- last
		}
	}()

	return out, nil
}
