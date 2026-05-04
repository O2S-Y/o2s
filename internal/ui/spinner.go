package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Spinner is a tiny, dependency-free CLI spinner used for short
// blocking commands (`o2s files hash`, `o2s sys ping`...).
// For interactive long-running TUIs we prefer Bubble Tea instead.
type Spinner struct {
	frames []string
	delay  time.Duration
	msg    string
	stop   chan struct{}
	wg     sync.WaitGroup
}

func NewSpinner(msg string) *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay:  90 * time.Millisecond,
		msg:    msg,
		stop:   make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	if !IsTTY() {
		fmt.Fprintln(os.Stderr, s.msg+"...")
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Fprint(os.Stderr, "\r\033[K")
				return
			case <-time.After(s.delay):
				th := T()
				frame := th.Title.Render(s.frames[i%len(s.frames)])
				fmt.Fprintf(os.Stderr, "\r\033[K%s %s", frame, s.msg)
				i++
			}
		}
	}()
}

func (s *Spinner) Stop()                  { close(s.stop); s.wg.Wait() }
func (s *Spinner) Update(msg string)      { s.msg = msg }
func (s *Spinner) Successf(format string, a ...any) {
	s.Stop()
	t := T()
	fmt.Fprintln(os.Stderr, t.Success.Render("✔ ")+fmt.Sprintf(format, a...))
}
func (s *Spinner) Failf(format string, a ...any) {
	s.Stop()
	t := T()
	fmt.Fprintln(os.Stderr, t.Danger.Render("✘ ")+fmt.Sprintf(format, a...))
}
