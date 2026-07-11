package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Spinner struct {
	stop chan struct{}
	done chan struct{}
	last string
}

func StartSpinner(msg string) *Spinner {
	s := &Spinner{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go func() {
		defer close(s.done)
		frames := []string{"|", "/", "-", "\\"}
		i := 0
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-s.stop:
				padded := s.last
				if padded != "" {
					fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", len(padded)+2))
				}
				return
			case <-ticker.C:
				s.last = fmt.Sprintf("%s... %s", msg, frames[i%4])
				fmt.Fprintf(os.Stderr, "\r%s", s.last)
				i++
			}
		}
	}()
	return s
}

func (s *Spinner) Stop() {
	close(s.stop)
	<-s.done
}
