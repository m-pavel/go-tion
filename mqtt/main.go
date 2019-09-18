package main

import (
	"encoding/json"
	"flag"
	"log"
	_ "net/http"
	_ "net/http/pprof"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/m-pavel/go-tion/tion"
	"github.com/m-pavel/go-tion/gatt"
	"github.com/m-pavel/go-hassio-mqtt/pkg"
)

type Request struct {
	Gate   string `json:"gate"`
	On     bool   `json:"on"`
	Heater bool   `json:"heater"`
	Sound  bool   `json:"sound"`
	Out    int8   `json:"temp_out"`
	In     int8   `json:"temp_in"`
	Target int8   `json:"temp_target"`
	Speed  int8   `json:"speed"`
}

type TionService struct {
	t     tion.Tion
	bt    *string
	debug bool
	fake  *bool
	ss    ghm.SendState
}

func (ts *TionService) PrepareCommandLineParams() {
	ts.bt = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")
	ts.fake = flag.Bool("fake", false, "Fake device")
}
func (ts TionService) Name() string { return "tion" }

func (ts *TionService) Init(client MQTT.Client, topic, topicc, topica string, debug bool, ss ghm.SendState) error {
	if *ts.fake {
		log.Println("Using fake device.")
		ts.t = tion.NewFake()
	} else {
		ts.t = gatt.New(*ts.bt, debug)
	}

	ts.debug = debug
	ts.ss = ss

	if token := client.Subscribe(topicc, 2, ts.control); token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (ts TionService) Do() (interface{}, error) {
	ts.t.Connect()
	defer ts.t.Disconnect()
	s, err := ts.t.ReadState(7)
	if err != nil {
		return nil, err
	}

	if ts.debug {
		log.Println(s.BeautyString())
	}
	return &Request{
		Gate:   s.GateStatus(),
		On:     s.Enabled,
		Heater: s.HeaterEnabled,
		Out:    s.TempOut,
		In:     s.TempIn,
		Target: s.TempTarget,
		Speed:  s.Speed,
		Sound:  s.SoundEnabled,
	}, nil
}

func (ts *TionService) control(cli MQTT.Client, msg MQTT.Message) {
	req := Request{}
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println(err)
		return
	}
	if ts.debug {
		log.Println(req)
	}

	ts.t.Connect()
	defer ts.t.Disconnect()
	cs, err := ts.t.ReadState(7)
	if err != nil {
		log.Println(err)
		return
	}

	if cs.Enabled {
		if !req.On {
			cs.Enabled = false
			err = ts.t.Update(cs, 7)
			if err != nil {
				log.Println(err)
			} else {
				ts.ss()
				log.Println("Turned off by MQTT request")
			}
		} else {
			log.Println("Already on")
		}
	} else {
		if req.On {
			cs.Enabled = true
			err = ts.t.Update(cs, 7)
			if err != nil {
				log.Println(err)
			} else {
				ts.ss()
				log.Println("Turned on  by MQTT request")
			}
		} else {
			log.Println("Already off")
		}
	}
	log.Println("Control done.")
}

func (ts TionService) Close() error {
	return nil
}

func main() {
	ghm.NewStub(&TionService{}).Main()
}
