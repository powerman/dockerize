package main

import (
	"io"
	"log"

	"golang.org/x/net/context"

	"github.com/powerman/tail"
)

func tailFile(ctx context.Context, file string, dest io.Writer) {
	defer wg.Done()

	t := tail.Follow(ctx, tail.LoggerFunc(log.Printf), file)

	for _, err := io.Copy(dest, t); err != nil; _, err = io.Copy(dest, t) {
	}
}
