package logger

import (
	"io"
	"log"
)

func LogCloser(c io.Closer) {
	if c != nil {
		if err := c.Close(); err != nil {
			log.Print(err)
		}
	}
}
