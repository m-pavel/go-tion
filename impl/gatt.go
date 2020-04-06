// +build gatt

package impl

import (
	"log"
	"time"

	"sync"

	"github.com/go-errors/errors"
	"github.com/m-pavel/go-gattlib/pkg"
	"github.com/m-pavel/go-tion/tion"
)

type gattTion struct {
	g     *gattlib.Gatt
	Addr  string
	debug bool
	mutex *sync.Mutex
}

// New gattlib backend
func NewTionImpl(addr string, debug bool, options map[string]string) tion.Tion {
	return &gattTion{Addr: addr, g: &gattlib.Gatt{}, mutex: &sync.Mutex{}, debug: debug}
}

func (t gattTion) Info() string {
	return "github.com/m-pavel/go-tion/tion"
}

type cRes struct {
	s *tion.Status
	e error
}

func (t gattTion) Connected() bool {
	return t.g.Connected()
}

func (t *gattTion) Connect(timeout time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.g.Connected() {
		return errors.New("Tion already connected")
	}
	return t.g.Connect(t.Addr)
}
func (t *gattTion) Disconnect(duration time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.g.Connected() {
		return errors.New("Tion already disconnected")
	}
	return t.g.Disconnect()
}

func (t *gattTion) ReadState(readtimeout time.Duration) (*tion.Status, error) {
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

func (t *gattTion) rw() (*tion.Status, error) {
	if !t.g.Connected() {
		return nil, errors.New("Not connected")
	}
	if err := t.g.Write(tion.WriteCaract, tion.StatusRequest); err != nil {
		return nil, err
	}
	time.Sleep(2 * time.Second)
	resp, n, err := t.g.Read(tion.ReadCharact)
	if err != nil {
		return nil, err
	}
	if t.debug {
		log.Printf("RSP [%d]: %v\n", n, resp)
	}
	return tion.FromBytes(resp)
}

func (t *gattTion) Update(s *tion.Status, timeout time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.g.Connected() {
		return errors.New("Tion not connected")
	}

	c1 := make(chan error, 1)

	go func() {
		c1 <- t.g.Write(tion.WriteCaract, tion.FromStatus(s))
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return errors.New("Write timeout")
	}
}
