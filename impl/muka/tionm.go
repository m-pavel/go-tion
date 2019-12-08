package muka

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
	addr string

	debug bool

	d *device.Device1

	st *SyncTimeout
}

func (n mTion) Info() string {
	return "github.com/muka/go-bluetooth/api"
}

// New go ble backend
func New(addr string, debug ...bool) tion.Tion {
	nt := mTion{addr: addr}
	nt.st = NewSt()
	if len(debug) > 0 {
		nt.debug = debug[0]
	}
	return &nt
}

func (n *mTion) Connected() bool {
	c, err := n.isConnected()
	if err != nil {
		return false
	}
	return c
}

func (n *mTion) ReadState(timeout time.Duration) (*tion.Status, error) {
	data, err := n.st.Call(timeout, func(stc chan interface{}, ec chan error) {
		if c, err := n.isConnected(); err != nil {
			ec <- err
			return
		} else {
			if !c {
				ec <- errors.New("Not connected")
				return
			}
		}

		var data []byte

		wc, err := n.d.GetCharByUUID(tion.WriteCaract)
		if err != nil {
			ec <- err
			return
		}
		if err := wc.WriteValue(tion.StatusRequest, nil); err != nil {
			ec <- err
			return
		}
		time.Sleep(200 * time.Millisecond)
		rc, err := n.d.GetCharByUUID(tion.ReadCharact)
		if err != nil {
			ec <- err
			return
		}

		if data, err = rc.ReadValue(nil); err != nil {
			ec <- err
			return
		}
		if n.debug {
			log.Printf("RSP: %v\n", data)
		}

		if status, err := tion.FromBytes(data); err != nil {
			ec <- err
		} else {
			stc <- status
		}

	})
	if data != nil {
		return data.(*tion.Status), err
	} else {
		return nil, err
	}
}

func (n *mTion) Update(s *tion.Status, timeout time.Duration) error {
	_, err := n.st.Call(timeout, func(stc chan interface{}, ec chan error) {
		if c, err := n.isConnected(); err != nil {
			ec <- err
			return
		} else {
			if !c {
				ec <- errors.New("Not connected")
				return
			}
		}
		wc, err := n.d.GetCharByUUID(tion.WriteCaract)
		if err != nil {
			ec <- err
			return
		}
		wc.WriteValue(tion.FromStatus(s), nil)
		stc <- nil
	})
	return err
}

func (n *mTion) Connect(timeout time.Duration) error {
	_, err := n.st.Call(timeout, func(stc chan interface{}, ec chan error) {
		if c, err := n.isConnected(); err != nil {
			ec <- err
			return
		} else {
			if c {
				ec <- nil
				return
			}
		}
		ad, err := api.GetDefaultAdapter()
		if err != nil {
			ec <- err
			return
		}
		if n.d, err = ad.GetDeviceByAddress(n.addr); err != nil {
			ec <- err
			return
		}
		if n.d == nil {
			ec <- fmt.Errorf("Device %s not available", n.addr)
			return
		}
		if p, err := n.d.GetPaired(); err != nil {
			ec <- err
			return
		} else {
			if !p {
				ec <- fmt.Errorf("Device %s is not paired. Pair with bluetoothctrl", n.addr)
				return
			}
		}
		if err = n.d.Connect(); err != nil {
			ec <- err
			return
		}
		time.Sleep(time.Second)
		if _, err := n.d.GetDescriptorList(); err != nil {
			ec <- err
			return
		}
		stc <- nil
	})
	return err
}

func (n *mTion) isConnected() (bool, error) {
	if n.d == nil || n.d.Client() == nil {
		return false, nil
	}

	return n.d.GetConnected()
}

func (n *mTion) Disconnect(timeout time.Duration) error {
	if n.d != nil {
		_, err := n.st.Call(timeout, func(stc chan interface{}, ec chan error) {
			defer func() {
				n.d = nil
			}()
			if c, err := n.isConnected(); err != nil {
				ec <- err
				return
			} else {
				if !c {
					ec <- err
					return
				}
			}
			if err := n.d.Disconnect(); err != nil {
				ec <- err
				return
			}
			stc <- nil
		})
		return err
	}
	return nil
}
