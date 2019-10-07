package tionm

import (
	"errors"
	"time"

	"log"

	"fmt"

	"github.com/m-pavel/go-tion/tion"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/device"
)

type mTion struct {
	addr  string
	cnct  chan error
	debug bool

	d *device.Device1
}

func New(addr string, debug ...bool) tion.Tion {
	nt := mTion{addr: addr}
	nt.cnct = make(chan error)
	if len(debug) > 0 {
		nt.debug = debug[0]
	}
	return &nt
}

func (n *mTion) ReadState(timeout time.Duration) (*tion.Status, error) {
	wc, err := n.d.GetCharByUUID(tion.WRITE_CHARACT)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if err := wc.WriteValue(tion.StatusRequest, nil); err != nil {
		log.Println(err)
		return nil, err
	}
	time.Sleep(200 * time.Millisecond)
	rc, err := n.d.GetCharByUUID(tion.READ_CHARACT)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if data, err := rc.ReadValue(nil); err != nil {
		log.Println(err)
		return nil, err
	} else {
		if n.debug {
			log.Printf("RSP: %v\n", data)
		}
		return tion.FromBytes(data)
	}
}

func (n *mTion) Update(s *tion.Status, timeout time.Duration) error {
	wc, err := n.d.GetCharByUUID(tion.WRITE_CHARACT)
	if err != nil {
		return err
	}

	c1 := make(chan error, 1)
	go func() {
		c1 <- wc.WriteValue(tion.FromStatus(s), nil)
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(timeout):
		return errors.New(fmt.Sprintf("Write timeout %d", timeout))
	}
}

func (n *mTion) Connect(timeout time.Duration) error {
	ad, err := api.GetDefaultAdapter()
	if err != nil {
		return err
	}
	n.d, err = ad.GetDeviceByAddress(n.addr)
	if err != nil {
		return err
	}
	if err = n.d.Connect(); err != nil {
		return err
	}
	return nil
}

func (n *mTion) Disconnect() error {
	if n.d != nil {
		defer func() {
			n.d = nil
		}()
		return n.d.Disconnect()
	}
	return nil
}
