package main

import (
	"flag"
	"fmt"
	"log"

	"time"

	"github.com/m-pavel/go-tion/mqttcli"
	"github.com/m-pavel/go-tion/tion"
	"github.com/m-pavel/go-tion/tionm"
)

type cliDevice struct {
	device           *string
	mqtt             *string
	mqttUser         *string
	mqttPass         *string
	mqttCa           *string
	mqttTopic        *string
	mqttAvalTopic    *string
	mqttControlTopic *string
	debug            *bool
	timeout          time.Duration
}

func main() {
	device := cliDevice{}
	device.device = flag.String("device", "", "bt addr")
	device.mqtt = flag.String("mqtt", "", "MQTT addr")
	device.mqttUser = flag.String("mqtt-user", "", "MQTT user")
	device.mqttPass = flag.String("mqtt-pass", "", "MQTT password")
	device.mqttCa = flag.String("mqtt-ca", "", "MQTT ca")
	device.mqttTopic = flag.String("mqtt-t", "", "MQTT status topic")
	device.mqttAvalTopic = flag.String("mqtt-ta", "", "MQTT availability topic")
	device.mqttControlTopic = flag.String("mqtt-tc", "", "MQTT control topic")
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
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	device.timeout = time.Duration(*timeoutp) * time.Second
	if *device.device == "" && !*scanp && *device.mqtt == "" {
		log.Fatal("Device address or MQTT is mandatory")
	}

	if *on {
		deviceCall(&device, func(t tion.Tion, s *tion.Status) error {
			s.Enabled = true
			return t.Update(s, device.timeout)
		}, "Turned on")

		return
	}
	if *off {
		deviceCall(&device, func(t tion.Tion, s *tion.Status) error {
			s.Enabled = false
			return t.Update(s, device.timeout)
		}, "Turned off")
		return
	}
	if *temp != 0 {
		deviceCall(&device, func(t tion.Tion, s *tion.Status) error {
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
		deviceCall(&device, func(t tion.Tion, s *tion.Status) error {
			s.Speed = int8(*speed)
			return t.Update(s, device.timeout)
		}, fmt.Sprintf("Speed updated to %d", *speed))
		return
	}

	if *gate != "" {
		deviceCall(&device, func(t tion.Tion, s *tion.Status) error {
			s.SetGateStatus(*gate)
			return t.Update(s, device.timeout)
		}, fmt.Sprintf("Gate set to %s", *gate))
		return
	}

	if *sound != "" {
		deviceCall(&device, func(t tion.Tion, s *tion.Status) error {
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
		deviceCall(&device, func(t tion.Tion, s *tion.Status) error {
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
		return
	}

	if *status {
		t := newDevice(&device)
		if err := t.Connect(device.timeout); err != nil {
			log.Printf("Connect error: %v\n", err)
			return
		}
		defer t.Disconnect(device.timeout)
		state, err := t.ReadState(device.timeout)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(state.BeautyString())

	}
}

func deviceCall(device *cliDevice, cb func(tion.Tion, *tion.Status) error, succ string) error {
	t := newDevice(device)
	if err := t.Connect(device.timeout); err != nil {
		return err
	}
	defer t.Disconnect(device.timeout)
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
	if *device.device != "" {
		return tionm.New(*device.device, *device.debug)
	}
	if *device.mqtt != "" {
		return mqttcli.New(*device.mqtt, *device.mqttUser, *device.mqttPass, *device.mqttCa, *device.mqttTopic, *device.mqttAvalTopic, *device.mqttControlTopic, *device.debug)
	}
	log.Panic("Unable to create device")
	return nil
}

//func scan() {
//	gattlib.Scan(func(ad ble.Advertisement) {
//		log.Printf("%s %s", ad.Addr(), ad.LocalName())
//	}, 5)
//	time.Sleep(10 * time.Second)
//}
