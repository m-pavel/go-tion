package main

import (
	"flag"
	"fmt"
	"log"

	"time"

	"github.com/m-pavel/go-tion/tion"
	"github.com/m-pavel/go-tion/tionn"
)

const timeout = 7 * time.Second

func main() {
	var device = flag.String("device", "", "bt addr")
	var status = flag.Bool("status", true, "Request status")
	var scanp = flag.Bool("scan", false, "Perform scan")
	var debug = flag.Bool("debug", false, "Debug")
	var on = flag.Bool("on", false, "Turn on")
	var off = flag.Bool("off", false, "Turn off")
	var temp = flag.Int("temp", 0, "Set target temperature")
	var speed = flag.Int("speed", 0, "Set speed")
	var sound = flag.String("sound", "", "Sound on|off")
	var heater = flag.String("heater", "", "Heater on|off")
	var gate = flag.String("gate", "", "Set gate position(indoor|outdoor|mixed)")
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	if *device == "" && !*scanp {
		log.Fatal("Device address is mandatory")
	}

	if *on {
		deviceCall(*device, *debug, func(t tion.Tion, s *tion.Status) error {
			s.Enabled = true
			return t.Update(s, timeout)
		}, "Turned on")

		return
	}
	if *off {
		deviceCall(*device, *debug, func(t tion.Tion, s *tion.Status) error {
			s.Enabled = false
			return t.Update(s, timeout)
		}, "Turned off")
		return
	}
	if *temp != 0 {
		deviceCall(*device, *debug, func(t tion.Tion, s *tion.Status) error {
			s.TempTarget = int8(*temp)
			return t.Update(s, timeout)
		}, fmt.Sprintf("Target temperature updated to %d", *temp))
		return
	}

	if *speed != 0 {
		if *speed <= 0 || *speed > 6 {
			log.Println("Speed range 1..6")
			return
		}
		deviceCall(*device, *debug, func(t tion.Tion, s *tion.Status) error {
			s.Speed = int8(*speed)
			return t.Update(s, timeout)
		}, fmt.Sprintf("Speed updated to %d", *speed))
		return
	}

	if *gate != "" {
		deviceCall(*device, *debug, func(t tion.Tion, s *tion.Status) error {
			s.SetGateStatus(*gate)
			return t.Update(s, timeout)
		}, fmt.Sprintf("Gate set to %s", *gate))
		return
	}

	if *sound != "" {
		deviceCall(*device, *debug, func(t tion.Tion, s *tion.Status) error {
			if *sound == "on" {
				s.SoundEnabled = true
			} else {
				s.SoundEnabled = false
			}
			return t.Update(s, timeout)
		}, fmt.Sprintf("Sound set to %s", *sound))
		return
	}

	if *heater != "" {
		deviceCall(*device, *debug, func(t tion.Tion, s *tion.Status) error {
			if *heater == "on" {
				s.HeaterEnabled = true
			} else {
				s.HeaterEnabled = false
			}
			return t.Update(s, timeout)
		}, fmt.Sprintf("Heater set to %s", *heater))
		return
	}

	if *scanp {
		//scan()
		panic("Not supported")
		return
	}

	if *status {
		t := newDevice(*device, *debug)
		if err := t.Connect(timeout); err != nil {
			log.Printf("Connect error: %v\n", err)
			return
		}
		defer t.Disconnect()
		state, err := t.ReadState(timeout)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(state.BeautyString())
		time.Sleep(10 * time.Second)
	}
}

func deviceCall(addr string, dbg bool, cb func(tion.Tion, *tion.Status) error, succ string) error {
	t := newDevice(addr, dbg)
	if err := t.Connect(timeout); err != nil {
		return err
	}
	defer t.Disconnect()
	s, err := t.ReadState(timeout)
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

func newDevice(addr string, dbg bool) tion.Tion {
	return tionn.New(addr, dbg)
}

//func scan() {
//	gattlib.Scan(func(ad ble.Advertisement) {
//		log.Printf("%s %s", ad.Addr(), ad.LocalName())
//	}, 5)
//	time.Sleep(10 * time.Second)
//}
