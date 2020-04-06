package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/m-pavel/go-tion/impl"

	"time"

	"github.com/m-pavel/go-tion/tion"
)

type cliDevice struct {
	device  *string
	options map[string]string

	debug   *bool
	timeout time.Duration
}

func main() {
	device := cliDevice{options: make(map[string]string)}
	device.device = flag.String("device", "", "BT (or MQTT) address")
	mqttUser := flag.String("mqtt-user", "", "MQTT user")
	mqttPass := flag.String("mqtt-pass", "", "MQTT password")
	mqttCa := flag.String("mqtt-ca", "", "MQTT ca")
	mqttTopic := flag.String("mqtt-t", "", "MQTT status topic")
	mqttAvalTopic := flag.String("mqtt-ta", "", "MQTT availability topic")
	mqttControlTopic := flag.String("mqtt-tc", "", "MQTT control topic")
	var status = flag.Bool("status", true, "Request status")
	var scanp = flag.Bool("scan", false, "Perform scan")
	device.debug = flag.Bool("debug", false, "Debug")
	var on = flag.Bool("on", false, "Turn on")
	var off = flag.Bool("off", false, "Turn off")
	var temp = flag.Int("temp", 0, "Set target temperature")
	var speed = flag.Int("speed", 0, "Set speed")
	var sound = flag.String("sound", "", "Sound on|off")
	var heater = flag.String("heater", "", "Heater on|off")
	var gate = flag.String("gate", "", "Set gate position(indoor|outdoor|mixed)")
	var timeoutp = flag.Int("timeout", 7, "Timeout seconds")
	flag.Parse()
	device.options["mqtt-user"] = *mqttUser
	device.options["mqtt-password"] = *mqttPass
	device.options["mqtt-ca"] = *mqttCa
	device.options["mqtt-topic"] = *mqttTopic
	device.options["mqtt-topic-a"] = *mqttAvalTopic
	device.options["mqtt-topic-c"] = *mqttControlTopic

	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	device.timeout = time.Duration(*timeoutp) * time.Second
	if *device.device == "" {
		log.Fatal("Device address or MQTT is mandatory")
	}

	if *on {
		deviceCallLog(&device, func(t tion.Tion, s *tion.Status) error {
			s.Enabled = true
			return t.Update(s, device.timeout)
		}, "Turned on")

		return
	}
	if *off {
		deviceCallLog(&device, func(t tion.Tion, s *tion.Status) error {
			s.Enabled = false
			return t.Update(s, device.timeout)
		}, "Turned off")
		return
	}
	if *temp != 0 {
		deviceCallLog(&device, func(t tion.Tion, s *tion.Status) error {
			s.TempTarget = int8(*temp)
			return t.Update(s, device.timeout)
		}, fmt.Sprintf("Target temperature updated to %d", *temp))
		return
	}

	if *speed != 0 {
		if *speed <= 0 || *speed > 6 {
			log.Println("Speed range 1..6")
			return
		}
		deviceCallLog(&device, func(t tion.Tion, s *tion.Status) error {
			s.Speed = int8(*speed)
			return t.Update(s, device.timeout)
		}, fmt.Sprintf("Speed updated to %d", *speed))
		return
	}

	if *gate != "" {
		deviceCallLog(&device, func(t tion.Tion, s *tion.Status) error {
			s.SetGateStatus(*gate)
			return t.Update(s, device.timeout)
		}, fmt.Sprintf("Gate set to %s", *gate))
		return
	}

	if *sound != "" {
		deviceCallLog(&device, func(t tion.Tion, s *tion.Status) error {
			if *sound == "on" {
				s.SoundEnabled = true
			} else {
				s.SoundEnabled = false
			}
			return t.Update(s, device.timeout)
		}, fmt.Sprintf("Sound set to %s", *sound))
		return
	}

	if *heater != "" {
		deviceCallLog(&device, func(t tion.Tion, s *tion.Status) error {
			if *heater == "on" {
				s.HeaterEnabled = true
			} else {
				s.HeaterEnabled = false
			}
			return t.Update(s, device.timeout)
		}, fmt.Sprintf("Heater set to %s", *heater))
		return
	}

	if *scanp {
		//scan()
		panic("Not supported")
	}

	if *status {
		t := newDevice(&device)
		log.Printf("Using implementation %s\n", t.Info())
		if err := t.Connect(device.timeout); err != nil {
			log.Printf("Connect error: %v\n", err)
			return
		}
		defer func() {
			if err := t.Disconnect(device.timeout); err != nil {
				log.Println(err)
			}
		}()
		state, err := t.ReadState(device.timeout)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(state.BeautyString())

	}
}

func deviceCallLog(device *cliDevice, cb func(tion.Tion, *tion.Status) error, succ string) {
	if err := deviceCall(device, cb, succ); err != nil {
		log.Println(err)
	}
}

func deviceCall(device *cliDevice, cb func(tion.Tion, *tion.Status) error, succ string) error {
	t := newDevice(device)
	if err := t.Connect(device.timeout); err != nil {
		return err
	}
	defer func() {
		if err := t.Disconnect(device.timeout); err != nil {
			log.Println(err)
		}
	}()
	s, err := t.ReadState(device.timeout)
	if err != nil {
		return err
	}

	err = cb(t, s)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(succ)
	}
	return err
}

func newDevice(device *cliDevice) tion.Tion {
	return impl.NewTionImpl(*device.device, *device.debug, device.options)
}

//func scan() {
//	gattlib.Scan(func(ad ble.Advertisement) {
//		log.Printf("%s %s", ad.Addr(), ad.LocalName())
//	}, 5)
//	time.Sleep(10 * time.Second)
//}
