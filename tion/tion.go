package tion

import "time"

// BLE Characteristics
const (
	WriteCaract = "6e400002-b5a3-f393-e0a9-e50e24dcca9e"
	ReadCharact = "6e400003-b5a3-f393-e0a9-e50e24dcca9e"
)

// Tion brazer interface
type Tion interface {
	Connect(timeout time.Duration) error

	ReadState(timeout time.Duration) (*Status, error)
	Update(s *Status, timeout time.Duration) error

	Disconnect() error
}
