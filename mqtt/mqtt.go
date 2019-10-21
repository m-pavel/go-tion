package main

import (
	"encoding/json"
	"flag"
	"log"
	_ "net/http"
	_ "net/http/pprof"

	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/m-pavel/go-hassio-mqtt/pkg"

	"net/http"

	"github.com/m-pavel/go-tion/tion"
	"github.com/m-pavel/go-tion/tionm"
)

const timeout = 7 * time.Second

// TionService instance
type TionService struct {
	t      tion.Tion
	bt     *string
	debug  bool
	fake   *bool
	keepbt *bool
	ss     ghm.SendState
}

// PrepareCommandLineParams for TionService
func (ts *TionService) PrepareCommandLineParams() {
	ts.bt = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")
	ts.fake = flag.Bool("fake", false, "Fake device")
	ts.keepbt = flag.Bool("keepbt", false, "Keep bluetooth connection")

}

// Name of TionService
func (ts TionService) Name() string { return "tion" }

// Init TionService
func (ts *TionService) Init(client MQTT.Client, topic, topicc, topica string, debug bool, ss ghm.SendState) error {
	go func() {
		log.Println(http.ListenAndServe(":7070", nil))
	}()
	if *ts.fake {
		log.Println("Using fake device.")
		ts.t = tion.NewFake()
	} else {
		ts.t = tionm.New(*ts.bt, debug)
	}

	ts.debug = debug
	ts.ss = ss

	if token := client.Subscribe(topicc, 2, ts.control); token.Error() != nil {
		return token.Error()
	}

	if *ts.keepbt {
		return ts.t.Connect(timeout)
	}
	return nil
}
func (ts *TionService) cmStart() error {
	if *ts.keepbt && ts.t.Connected() {
		return nil
	}
	return ts.t.Connect(timeout)
}
func (ts *TionService) cmEnd() error {
	if *ts.keepbt {
		return nil
	}
	return ts.t.Disconnect(timeout)
}

// Do TionService
func (ts TionService) Do() (interface{}, error) {
	if err := ts.cmStart(); err != nil {
		return nil, err
	}
	defer ts.cmEnd()
	s, err := ts.t.ReadState(timeout)
	if err != nil {
		return nil, err
	}

	if ts.debug {
		log.Println(s.BeautyString())
	}
	return tion.RestFromStatus(s), nil
}

func (ts *TionService) control(cli MQTT.Client, msg MQTT.Message) {
	req := tion.RestStatus{}
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println(err)
		return
	}
	if ts.debug {
		log.Println(req)
	}

	if err := ts.cmStart(); err != nil {
		log.Println(err)
		return
	}
	defer ts.cmEnd()

	cs, err := ts.t.ReadState(timeout)
	if err != nil {
		log.Println(err)
		return
	}

	if req.Speed != nil && *req.Speed != cs.Speed {
		cs.Speed = *req.Speed
		err = ts.t.Update(cs, timeout)
		if err != nil {
			log.Println(err)
		} else {
			ts.ss()
			log.Printf("Updated speed to %d by MQTT request\n", *req.Speed)
		}

	} else {
		if req.On != nil {
			if cs.Enabled {
				if !*req.On {
					cs.Enabled = false
					err = ts.t.Update(cs, timeout)
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
				if *req.On {
					cs.Enabled = true
					err = ts.t.Update(cs, timeout)
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
		}
	}
	log.Println("Control done.")
}

// Close TionService
func (ts TionService) Close() error {
	return ts.t.Disconnect(timeout)
}

func main() {
	ghm.NewStub(&TionService{}).Main()
}
