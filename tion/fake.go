package tion

type fakeTion struct {
	s Status
}

func NewFake() Tion {
	return &fakeTion{}
}

func (t fakeTion) Connect() error {
	return nil
}
func (t fakeTion) Disconnect() error {
	return nil
}
func (ft fakeTion) ReadState(tmt int) (*Status, error) {
	return &ft.s, nil
}
func (ft *fakeTion) Update(s *Status, tmt int) error {
	ft.s = *s
	return nil
}
