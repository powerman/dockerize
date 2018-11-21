// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tail implements "tail -F" functionality following rotated logs
package tail

import (
	"bufio"
	"io"
	"log"
	"os"
	"sync"
	"syscall"
	"time"
)

// Tail is an io.ReadCloser with "tail -F" behaviour.
type Tail struct {
	reader     *bufio.Reader
	readerErr  error
	readerLock sync.RWMutex
	filename   string
	file       *os.File
	stat       os.FileInfo
}

const (
	defaultRetryInterval = 200 * time.Millisecond
	maxRetryInterval     = 30 * time.Second
)

// NewTail starts opens the given file and watches it for deletion/rotation
func NewTail(filename string) (*Tail, error) {
	t := &Tail{
		filename: filename,
	}
	// Initialize readerErr as io.EOF, so that the reader can work properly
	// during initialization.
	t.readerErr = io.EOF
	go t.watchLoop()
	return t, nil
}

// Read implements the io.Reader interface for Tail
func (t *Tail) Read(p []byte) (int, error) {
	t.readerLock.RLock()
	defer t.readerLock.RUnlock()
	if t.readerErr != nil {
		return 0, t.readerErr
	}
	return t.reader.Read(p)
}

var _ io.ReadCloser = &Tail{}

// Close stops watching and closes the file
func (t *Tail) Close() error {
	t.file.Close()
	return nil
}

func (t *Tail) attemptOpen() error {
	t.readerLock.Lock()
	defer t.readerLock.Unlock()
	t.readerErr = nil
	attempt := 0
	var lastErr error
	for interval := defaultRetryInterval; ; interval *= 2 {
		attempt++
		var err error
		t.file, err = os.OpenFile(t.filename, os.O_RDONLY|syscall.O_NONBLOCK, 0600)
		if err == nil {
			err = syscall.SetNonblock(int(t.file.Fd()), false)
		}

		if err == nil {
			t.stat, err = t.file.Stat()
		}

		if err == nil {
			// TODO: not interested in old events?
			// t.file.Seek(0, os.SEEK_END)
			t.reader = bufio.NewReader(t.file)
			return nil
		}
		lastErr = err
		log.Printf("can't open %s: %v", t.filename, err)

		if interval >= maxRetryInterval {
			break
		}
		time.Sleep(interval)
	}
	t.readerErr = lastErr
	return lastErr
}

func (t *Tail) watchLoop() {
	for {
		err := t.watchFile()
		if err != nil {
			log.Printf("tail failed on %s: %v", t.filename, err)
			break
		}
	}
}

func (t *Tail) watchFile() error {
	err := t.attemptOpen()
	if err != nil {
		return err
	}
	defer t.file.Close()

	for {
		time.Sleep(2 * time.Second)
		stat, err := os.Stat(t.filename)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Cannot stat file %s: %v", t.filename, err)
		}

		if err == nil {
			tstat := t.stat.Sys().(*syscall.Stat_t)
			st := stat.Sys().(*syscall.Stat_t)

			if tstat.Dev != st.Dev || tstat.Ino != st.Ino {
				log.Printf("Log file %s moved/deleted", t.filename)
				t.readerLock.Lock()
				defer t.readerLock.Unlock()
				t.readerErr = io.EOF
				return nil
			}
		}
	}
}
