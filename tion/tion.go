package tion

import "time"

const (
	WRITE_CHARACT = "6e400002-b5a3-f393-e0a9-e50e24dcca9e"
	READ_CHARACT  = "6e400003-b5a3-f393-e0a9-e50e24dcca9e"
)

type Tion interface {
	Connect(timeout time.Duration) error

	ReadState(timeout time.Duration) (*Status, error)
	Update(s *Status, timeout time.Duration) error

	Disconnect() error
}
