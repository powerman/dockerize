package main

import (
	"io"
	"log"
	"os"

	"golang.org/x/net/context"

	"github.com/powerman/dockerize/pkg/tail"
)

const (
	stdioBufSize = 4096
)

// here is some syntaxic sugar inspired by the Tomas Senart's video,
// it allows me to inline the Reader interface
type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

func tailFile(ctx context.Context, file string, dest *os.File) {
	defer wg.Done()

	t, err := tail.NewTail(file)
	if err != nil {
		log.Printf("cannot tail file %s: %s", file, err)
		return
	}

	for {
		// Copy will call the Reader and Writer interface multiple time, in order
		// to copy by chunk (avoiding loading the whole file in memory).
		// I insert the ability to cancel before read time as it is the earliest
		// possible in the call process.
		buf := make([]byte, stdioBufSize)
		_, err = io.CopyBuffer(dest, readerFunc(func(p []byte) (int, error) {
			// golang non-blocking channel: https://gobyexample.com/non-blocking-channel-operations
			select {

			// if context has been canceled
			case <-ctx.Done():
				// stop process and propagate "context canceled" error
				return 0, ctx.Err()
			default:
				// otherwise just run default io.Reader implementation
				return t.Read(p)
			}
		}), buf)
		if err != nil {
			log.Printf("error while tailing file %s: %s", file, err)
			return
		}
	}
}
