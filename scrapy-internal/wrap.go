package scrapy_internal

import (
	"sync"
)

type WaitGroupWrap struct {
	sync.WaitGroup
}

func (w *WaitGroupWrap) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

func Once(f func()) {
	once.Do(f)
}
