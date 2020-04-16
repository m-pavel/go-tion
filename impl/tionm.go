// +build muka

package impl

import (
	"errors"
	"strings"
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
func NewTionImpl(addr string, debug bool, options map[string]string) tion.Tion {
	nt := mTion{addr: addr}
	nt.st = NewSt()
	nt.debug = debug
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
				ec <- errors.New("R1: Not connected")
				return
			}
		}

		var data []byte

		wc, err := n.d.GetCharByUUID(tion.WriteCaract)
		if err != nil {
			ec <- fmt.Errorf("R2: %w", err)
			return
		}
		if err := wc.WriteValue(tion.StatusRequest, nil); err != nil {
			ec <- fmt.Errorf("R3: %w", err)
			return
		}
		time.Sleep(200 * time.Millisecond)
		rc, err := n.d.GetCharByUUID(tion.ReadCharact)
		if err != nil {
			ec <- fmt.Errorf("R4: %w", err)
			return
		}

		if data, err = rc.ReadValue(nil); err != nil {
			ec <- fmt.Errorf("R5: %w", err)
			return
		}
		if n.debug {
			log.Printf("RSP: %v\n", data)
		}

		if status, err := tion.FromBytes(data); err != nil {
			ec <- fmt.Errorf("R6: %w", err)
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
			ec <- fmt.Errorf("U1: %w", err)
			return
		} else {
			if !c {
				ec <- errors.New("U1: Not connected")
				return
			}
		}
		wc, err := n.d.GetCharByUUID(tion.WriteCaract)
		if err != nil {
			ec <- fmt.Errorf("U2: %w", err)
			return
		}
		if err := wc.WriteValue(tion.FromStatus(s), nil); err != nil {
			ec <- fmt.Errorf("U3: %w", err)
			return
		}
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
			ec <- fmt.Errorf("C1: %w", err)
			return
		}
		if n.d, err = ad.GetDeviceByAddress(n.addr); err != nil {
			ec <- fmt.Errorf("C2: %w", err)
			return
		}
		if n.d == nil {
			ec <- fmt.Errorf("C3: Device %s not available", n.addr)
			return
		}
		if p, err := n.d.GetPaired(); err != nil {
			ec <- fmt.Errorf("C4: %w", err)
			return
		} else {
			if !p {
				ec <- fmt.Errorf("C5: Device %s is not paired. Pair with bluetoothctrl", n.addr)
				return
			}
		}
		if err = n.d.Connect(); err != nil {
			ec <- fmt.Errorf("C6: %w", err)
			return
		}
		time.Sleep(time.Second)
		if _, err := n.d.GetDescriptorList(); err != nil {
			ec <- fmt.Errorf("C7: %w", err)
			return
		}
		stc <- nil
	})
	return err
}

func (n *mTion) Connect_(timeout time.Duration) error {
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
			ec <- fmt.Errorf("C1: %w", err)
			return
		}
		chn, cancel, err := ad.OnDeviceDiscovered()
		if err != nil {
			ec <- fmt.Errorf("C2: %w", err)
			return
		}
		err = ad.StartDiscovery()
		if err != nil {
			ec <- fmt.Errorf("C3: %w", err)
			return
		}

		defer func() {
			err := ad.StopDiscovery()
			if err != nil {
				log.Println(err)
			}
			cancel()
		}()

		spl := strings.Split(n.addr, ":")
		pathsuffix := fmt.Sprintf("%s_%s_%s_%s_%s_%s", spl[0], spl[1], spl[2], spl[3], spl[4], spl[5])

		for {
			select {
			case dev := <-chn:
				if strings.HasSuffix(string(dev.Path), pathsuffix) {
					log.Printf("Found %s", dev.Path)
					break
				}
			case <-time.After(timeout):
				ec <- fmt.Errorf("C4: discovery timout %.2fs", timeout.Seconds())
				return
			}
		}
		ad.GetGattManager()
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
				ec <- fmt.Errorf("D1: %w", err)
				return
			} else {
				if !c {
					ec <- err
					return
				}
			}
			if err := n.d.Disconnect(); err != nil {
				ec <- fmt.Errorf("D2: %w", err)
				return
			}
			stc <- nil
		})
		return err
	}
	return nil
}
