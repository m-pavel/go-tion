package tionn

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

func New(addr string, debug ...bool) tion.Tion {
	nt := nativeTion{addr: addr}
	nt.cnct = make(chan error)
	if len(debug) > 0 {
		nt.debug = debug[0]
	}
	return &nt
}

func (n *nativeTion) ReadState(timeout time.Duration) (*tion.Status, error) {
	if err := n.p.WriteCharacteristic(n.wc, tion.StatusRequest, true); err != nil {
		return nil, err
	}
	time.Sleep(2 * time.Second)
	resp, err := n.p.ReadCharacteristic(n.rc)
	if err != nil {
		return nil, err
	}
	if n.debug {
		log.Printf("RSP [%d]: %v\n", n, resp)
	}
	return tion.FromBytes(resp)
}

func (n *nativeTion) Update(s *tion.Status, timeout time.Duration) error {
	return errors.New("not implemented")
}

func (n *nativeTion) Connect(timeout time.Duration) error {
	var DefaultClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, false),
	}

	d, err := gatt.NewDevice(DefaultClientOptions...)
	if err != nil {
		return err
	}

	d.Handle(
		gatt.PeripheralDiscovered(n.onPeriphDiscovered),
		gatt.PeripheralConnected(n.onPeriphConnected),
		gatt.PeripheralDisconnected(n.onPeriphDisconnected),
	)
	if err = d.Init(onStateChanged); err != nil {
		return err
	}

	select {
	case res := <-n.cnct:
		return res
	case <-time.After(timeout):
		return errors.New(fmt.Sprintf("Connect timeout (%d)", timeout))
	}
}

func (n *nativeTion) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
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
			if c.UUID().String() == tion.WRITE_CHARACT {
				n.wc = c
			}
			if c.UUID().String() == tion.READ_CHARACT {
				n.rc = c
			}
		}
	}
	n.cnct <- nil
}

func (n *nativeTion) onPeriphDisconnected(p gatt.Peripheral, err error) {
	n.p = nil
	n.cnct <- err
}
func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		log.Printf("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (n *nativeTion) Disconnect() error {
	if n.p != nil {
		// TODO
	}
	return nil
}
