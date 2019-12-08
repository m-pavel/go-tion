package fake

import (
	"time"

	"github.com/m-pavel/go-tion/tion"
)

type fakeTion struct {
	s tion.Status
}

// NewFake backend
func NewFake() tion.Tion {
	return &fakeTion{}
}

func (ft fakeTion) Info() string {
	return "fake"
}

func (ft fakeTion) Connected() bool {
	return true
}
func (ft fakeTion) Connect(time.Duration) error {
	return nil
}
func (ft fakeTion) Disconnect(time.Duration) error {
	return nil
}
func (ft fakeTion) ReadState(time.Duration) (*tion.Status, error) {
	return &ft.s, nil
}
func (ft *fakeTion) Update(s *tion.Status, tm time.Duration) error {
	ft.s = *s
	return nil
}
