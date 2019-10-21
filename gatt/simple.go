package gatt

import (
	"log"
	"time"

	"sync"

	"github.com/go-errors/errors"
	"github.com/m-pavel/go-gattlib/pkg"
	tion2 "github.com/m-pavel/go-tion/tion"
)

type tion struct {
	g     *gattlib.Gatt
	Addr  string
	debug bool
	mutex *sync.Mutex
}

// New gattlib backend
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

func (n tion) Connected() bool {
	return n.g.Connected()
}

func (t *tion) Connect(timeout time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.g.Connected() {
		return errors.New("Tion already connected")
	}
	return t.g.Connect(t.Addr)
}
func (t *tion) Disconnect(duration time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.g.Connected() {
		return errors.New("Tion already disconnected")
	}
	return t.g.Disconnect()
}

func (t *tion) ReadState(readtimeout time.Duration) (*tion2.Status, error) {
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
	case <-time.After(readtimeout):
		return nil, errors.New("Read timeout")
	}
}

func (t *tion) rw() (*tion2.Status, error) {
	if !t.g.Connected() {
		return nil, errors.New("Not connected")
	}
	if err := t.g.Write(tion2.WriteCaract, tion2.StatusRequest); err != nil {
		return nil, err
	}
	time.Sleep(2 * time.Second)
	resp, n, err := t.g.Read(tion2.ReadCharact)
	if err != nil {
		return nil, err
	}
	if t.debug {
		log.Printf("RSP [%d]: %v\n", n, resp)
	}
	return tion2.FromBytes(resp)
}

func (t *tion) Update(s *tion2.Status, timeout time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.g.Connected() {
		return errors.New("Tion not connected")
	}

	c1 := make(chan error, 1)

	go func() {
		c1 <- t.g.Write(tion2.WriteCaract, tion2.FromStatus(s))
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return errors.New("Write timeout")
	}
}
