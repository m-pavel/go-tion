package tionn

import (
	"log"

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

func (n *nativeTion) ReadState(timeout int) (*tion.Status, error) {
	return nil, nil
}

func (n *nativeTion) Update(s *tion.Status, timeout int) error {
	return nil
}

func (n *nativeTion) Connect() error {
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
	return <-n.cnct
}

func (n *nativeTion) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	fmt.Println(a.LocalName)
	fmt.Println(p)
	fmt.Println(a)

	if a.LocalName == "AAA" {
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
	log.Printf("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Printf("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (n *nativeTion) Disconnect() error {
	return nil
}
