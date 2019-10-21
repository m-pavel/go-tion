package tion

import "time"

type fakeTion struct {
	s Status
}

// NewFake backend
func NewFake() Tion {
	return &fakeTion{}
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
func (ft fakeTion) ReadState(time.Duration) (*Status, error) {
	return &ft.s, nil
}
func (ft *fakeTion) Update(s *Status, tm time.Duration) error {
	ft.s = *s
	return nil
}
