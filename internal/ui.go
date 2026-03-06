package internal

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	frames   []string
	current  int
	mu       sync.Mutex
	message  string
	warnings []string
	done     chan struct{}
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		done:    make(chan struct{}),
	}
}

func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = msg
}

func (s *Spinner) Start() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.Lock()
				frame := s.frames[s.current]
				msg := s.message
				s.current = (s.current + 1) % len(s.frames)
				s.mu.Unlock()

				fmt.Printf("\r\033[K%s %s", frame, msg)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) FlushWarnings() {
	for _, w := range s.warnings {
		fmt.Printf("⚠ %s\n", w)
	}
}

func (s *Spinner) BufferWarn(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.warnings = append(s.warnings, msg)
}

func (s *Spinner) Stop(finalMessage string) {
	s.done <- struct{}{}
	fmt.Printf("\r\033[K✓ %s\n", finalMessage)
	s.FlushWarnings()
}

func (s *Spinner) Fail(finalMessage string) {
	s.done <- struct{}{}
	fmt.Printf("\r\033[K✗ %s\n", finalMessage)
	s.FlushWarnings()
}
