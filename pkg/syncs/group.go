package syncs

import (
	"sync"
)

// WaitGroup is a custom wrapper around sync.WaitGroup.
// Done for more clear workflow with goroutines.
type WaitGroup struct {
	wg sync.WaitGroup
}

func (w *WaitGroup) Go(f func()) {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		f()
	}()
}

func (w *WaitGroup) Wait() {
	w.wg.Wait()
}
