// +build ppal

package impl

import (
	"log"
	"time"

	"errors"
	"fmt"

	"github.com/m-pavel/go-tion/tion"
	"github.com/paypal/gatt"
)

type nativeTion struct {
	addr   string
	p      gatt.Peripheral
	rc, wc *gatt.Characteristic
	cnct   chan error
	debug  bool
}

// New go gatt backend
func NewTionImpl(addr string, debug bool, options map[string]string) tion.Tion {
	nt := nativeTion{addr: addr}
	nt.cnct = make(chan error)
	nt.debug = debug
	return &nt
}

func (n nativeTion) Info() string {
	return "github.com/paypal/gatt"
}

func (n nativeTion) Connected() bool {
	return true // TODO
}

func (n *nativeTion) ReadState(timeout time.Duration) (*tion.Status, error) {
	if err := n.p.WriteCharacteristic(n.wc, tion.StatusRequest, true); err != nil {
		return nil, err
	}
	for try := 0; try < 5; try++ {
		time.Sleep(2 * time.Second)
		resp, err := n.p.ReadCharacteristic(n.rc)
		if err != nil {
			return nil, err
		}
		if n.debug {
			log.Printf("RSP [%d]: %s\n", len(resp), tion.Bytes(resp))
		}
		log.Println(len(resp))
	}
	//resp, err = n.p.ReadCharacteristic(n.rc)
	//if err != nil {
	//	return nil, err
	//}
	//if n.debug {
	//	log.Printf("RSP [%d]: %v\n", len(resp), resp)
	//}
	//return tion.FromBytes(resp)
	return nil, errors.New("ReadState not implemented")
}

func (n *nativeTion) Update(s *tion.Status, timeout time.Duration) error {
	return errors.New("Update not implemented")
}

func (n *nativeTion) Connect(timeout time.Duration) error {
	var DefaultClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, false),
	}

	d, err := gatt.NewDevice(DefaultClientOptions...)
	if err != nil {
		return fmt.Errorf("C1: %w", err)
	}

	d.Handle(
		gatt.PeripheralDiscovered(n.onPeriphDiscovered),
		gatt.PeripheralConnected(n.onPeriphConnected),
		gatt.PeripheralDisconnected(n.onPeriphDisconnected),
	)
	if err = d.Init(onStateChanged); err != nil {
		return fmt.Errorf("C2: %w", err)
	}

	select {
	case res := <-n.cnct:
		return res
	case <-time.After(timeout):
		return fmt.Errorf("C3: Connect timeout %.2fs", timeout.Seconds())
	}
}

func (n *nativeTion) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if n.debug {
		log.Printf("Discovered %s %s", p.Name(), p.ID())
	}
	if p.ID() == n.addr {
		p.Device().StopScanning()
		p.Device().Connect(p)
	}
}

func (n *nativeTion) onPeriphConnected(p gatt.Peripheral, err error) {
	if err != nil {
		n.cnct <- err
	}
	n.p = p

	services, err := p.DiscoverServices(nil)
	if err != nil {
		n.cnct <- err
	}

	for _, service := range services {
		cs, _ := p.DiscoverCharacteristics(nil, service)

		for _, c := range cs {
			log.Printf("%v %v\n", service.UUID().String(), c.UUID().String())
			if c.UUID().Equal(gatt.MustParseUUID(tion.WriteCaract)) {
				n.wc = c
			}
			if c.UUID().Equal(gatt.MustParseUUID(tion.ReadCharact)) {
				n.rc = c
			}
		}
	}

	if n.wc == nil {
		n.cnct <- errors.New("Unable to find write characteristic")
	}
	if n.rc == nil {
		n.cnct <- errors.New("Unable to find read characteristic")
	}

	if dd, err := p.DiscoverDescriptors(nil, n.rc); err != nil {
		n.cnct <- err
	} else {
		log.Println(dd)
	}

	if err := p.SetNotifyValue(n.rc, n.reqdNotification); err != nil {
		n.cnct <- err
	}
	log.Println("Subscribed")
	n.cnct <- nil
}

func (n *nativeTion) reqdNotification(c *gatt.Characteristic, data []byte, err error) {
	log.Println("AAAA")
	log.Println(err)
	log.Println(data)
}

func (n *nativeTion) onPeriphDisconnected(p gatt.Peripheral, err error) {
	n.p = nil
	n.cnct <- err
}
func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		log.Printf("Scanning with %v...", d)
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (n *nativeTion) Disconnect(time.Duration) error {
	//if n.p != nil {
	//	// TODO
	//}
	return nil
}
