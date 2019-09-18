package tion

type Tion interface {
	ReadState(timeout int) (*Status, error)
	Update(s *Status, timeout int) error
	Connect() error
	Disconnect() error
}
