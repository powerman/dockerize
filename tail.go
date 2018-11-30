package main

import (
	"context"
	"io"
	"log"

	"github.com/powerman/tail"
)

func tailFile(path string, dest io.Writer) {
	t := tail.Follow(context.Background(), tail.LoggerFunc(log.Printf), path)

	go func() {
		for _, err := io.Copy(dest, t); err != nil; _, err = io.Copy(dest, t) {
		}
	}()
}
