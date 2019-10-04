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
	addr string
	cnct chan error
}

func New(addr string, debug ...bool) tion.Tion {
	nt := nativeTion{addr: addr}
	nt.cnct = make(chan error)
	return &nt
}

func (n *nativeTion) ReadState(timeout time.Duration) (*tion.Status, error) {
	return nil, nil
}

func (n *nativeTion) Update(s *tion.Status, timeout time.Duration) error {
	return nil
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
	n.cnct <- err
}

func (n *nativeTion) onPeriphDisconnected(p gatt.Peripheral, err error) {
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
	return nil
}
