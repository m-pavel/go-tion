package gatt

import (
	"log"
	"time"

	"sync"

	"github.com/go-errors/errors"
	"github.com/m-pavel/go-gattlib/pkg"
	tion2 "github.com/m-pavel/go-tion/tion"
)

const (
	wchar = "6e400002-b5a3-f393-e0a9-e50e24dcca9e"
	rchar = "6e400003-b5a3-f393-e0a9-e50e24dcca9e"
)

type tion struct {
	g     *gattlib.Gatt
	Addr  string
	debug bool
	mutex *sync.Mutex
}

func New(addr string, debug ...bool) tion2.Tion {
	t := tion{Addr: addr, g: &gattlib.Gatt{}, mutex: &sync.Mutex{}}
	if len(debug) == 1 && debug[0] {
		t.debug = true
	}
	return &t
}

type cRes struct {
	s *tion2.Status
	e error
}

func (t *tion) Connect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.g.Connected() {
		return errors.New("Tion already connected")
	}
	return t.g.Connect(t.Addr)
}
func (t *tion) Disconnect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.g.Connected() {
		return errors.New("Tion already disconnected")
	}
	return t.g.Disconnect()
}

func (t *tion) ReadState(timeout int) (*tion2.Status, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.g.Connected() {
		return nil, errors.New("Tion not connected")
	}

	c1 := make(chan cRes, 1)

	go func() {
		r, e := t.rw()
		c1 <- cRes{e: e, s: r}
	}()

	select {
	case res := <-c1:
		return res.s, res.e
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil, errors.New("Read timeout")
	}
}

func (t *tion) rw() (*tion2.Status, error) {
	if !t.g.Connected() {
		return nil, errors.New("Not connected")
	}
	if err := t.g.Write(wchar, tion2.StatusRequest); err != nil {
		return nil, err
	}
	time.Sleep(time.Second)
	resp, _, err := t.g.Read(rchar)
	if err != nil {
		return nil, err
	}
	if t.debug {
		log.Printf("RSP: %v\n", resp)
	}
	return tion2.FromBytes(resp)
}

func (t *tion) Update(s *tion2.Status, timeout int) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.g.Connected() {
		return errors.New("Tion not connected")
	}

	c1 := make(chan error, 1)

	go func() {
		c1 <- t.g.Write(wchar, tion2.FromStatus(s))
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return errors.New("Write timeout")
	}
}
