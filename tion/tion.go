package tion

import "time"

type Tion interface {
	Connect(timeout time.Duration) error

	ReadState(timeout time.Duration) (*Status, error)
	Update(s *Status, timeout time.Duration) error

	Disconnect() error
}
