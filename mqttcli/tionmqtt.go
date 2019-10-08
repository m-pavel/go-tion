package mqttcli

import (
	"time"

	"github.com/m-pavel/go-tion/tion"
)

type mqttTion struct {
}

func New(url, user, pass string) tion.Tion {
	mqt := mqttTion{}
	return &mqt
}

func (mqt *mqttTion) Connect(timeout time.Duration) error {
	return nil
}

func (mqt *mqttTion) ReadState(timeout time.Duration) (*tion.Status, error) {
	return nil, nil
}
func (mqt *mqttTion) Update(s *tion.Status, timeout time.Duration) error {
	return nil
}

func (mqt *mqttTion) Disconnect() error {
	return nil
}
