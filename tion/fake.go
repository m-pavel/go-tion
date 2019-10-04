package tion

import "time"

type fakeTion struct {
	s Status
}

func NewFake() Tion {
	return &fakeTion{}
}

func (t fakeTion) Connect(time.Duration) error {
	return nil
}
func (t fakeTion) Disconnect() error {
	return nil
}
func (ft fakeTion) ReadState(time.Duration) (*Status, error) {
	return &ft.s, nil
}
func (ft *fakeTion) Update(s *Status, tm time.Duration) error {
	ft.s = *s
	return nil
}
