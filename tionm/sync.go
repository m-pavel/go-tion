package tionm

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type SyncTimeout struct {
	m sync.Mutex
}

func NewSt() *SyncTimeout {
	st := SyncTimeout{m: sync.Mutex{}}
	return &st
}

type Callback func(chan interface{}, chan error)

func (sm *SyncTimeout) Call(timeout time.Duration, callback Callback) (interface{}, error) {
	dc := make(chan interface{}, 1)
	ec := make(chan error, 1)

	go func() {
		sm.m.Lock()
		defer sm.m.Unlock()
		callback(dc, ec)
	}()

	select {
	case data := <-dc:
		return data, nil
	case err := <-ec:
		log.Println(err)
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("Calltimeout %f sec.", timeout.Seconds())
	}
}
