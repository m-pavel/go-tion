package impl

import (
	"github.com/paypal/gatt/examples/option"

	"github.com/paypal/gatt"
)

func HciInit() error {
	_, err := gatt.NewDevice(option.DefaultClientOptions...)
	return err
}
