/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
	This package is a FORK of https://github.com/kubernetes-sigs/kind/blob/master/pkg/log/status.go
	See above license
*/

// Package fidget implements CLI functionality for bored users waiting for results
package fidget

import (
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

// These are frames for the default "spinner" using unicode. These
// are meant for macOS and Linux terminals that by default support unicode.
var unicodeSpinnerFrames = []string{
	"◐",
	"◓",
	"◑",
	"◒",
}

// These are the spinner using ASCII. We revert to these frames
// for Windows terminals that don't support unicode.
var asciiSpinnerFrames = []string{
	"<",
	"^",
	">",
	"v",
}

// Spinner is a simple and efficient CLI loading spinner used by kind
// It is simplistic and assumes that the line length will not change.
// It is best used indirectly via log.Status (see parent package)
type Spinner struct {
	frames []string
	stop   chan struct{}
	ticker *time.Ticker
	writer io.Writer
	mu     *sync.Mutex
	// protected by mu
	prefix string
	suffix string
	start  time.Time
}

// NewSpinner initializes and returns a new Spinner that will write to
func NewSpinner(w io.Writer) *Spinner {

	frames := unicodeSpinnerFrames
	if runtime.GOOS == "windows" {
		frames = asciiSpinnerFrames
	}

	return &Spinner{
		frames: frames,
		stop:   make(chan struct{}, 1),
		ticker: time.NewTicker(time.Millisecond * 200),
		mu:     &sync.Mutex{},
		writer: w,
	}
}

// SetPrefix sets the prefix to print before the spinner
func (s *Spinner) SetPrefix(prefix string) {
	s.mu.Lock()
	s.prefix = prefix
	s.mu.Unlock()
}

// SetSuffix sets the suffix to print after the spinner
func (s *Spinner) SetSuffix(suffix string) {
	s.mu.Lock()
	s.suffix = suffix
	s.mu.Unlock()
}

// Start starts the spinner running
func (s *Spinner) Start() {
	s.start = time.Now()
	go func() {
		for {
			for _, frame := range s.frames {
				select {
				case <-s.stop:
					return
				case <-s.ticker.C:
					func() {
						s.mu.Lock()
						defer s.mu.Unlock()
						fmt.Fprintf(s.writer, "\r%s%s%s", s.prefix, frame, s.suffix)
					}()
				}
			}
		}
	}()
}

// Stop signals the spinner to stop
func (s *Spinner) Stop() {
	s.stop <- struct{}{}
}

// TimeSpent returns the seconds spent since the spinner first started
func (s *Spinner) TimeSpent() time.Duration {
	currentTime := time.Now()

	if currentTime.Sub(s.start) < time.Second {
		return currentTime.Sub(s.start).Round(time.Millisecond)
	}

	return currentTime.Sub(s.start).Round(time.Second)
}
