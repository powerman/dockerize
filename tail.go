package main

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/powerman/tail"
)

func tailFile(ctx context.Context, wg *sync.WaitGroup, path string, dest io.Writer) {
	t := tail.Follow(ctx, tail.LoggerFunc(log.Printf), path)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, err := io.Copy(dest, t); err != nil; _, err = io.Copy(dest, t) {
		}
	}()
}
