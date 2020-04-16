package main

import (
	"encoding/json"
	"log"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/m-pavel/go-tion/tion"
)

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
	defer func() {
		if err := ts.cmEnd(); err != nil {
			log.Println(err)
		}
	}()

	cs, err := ts.t.ReadState(timeout)
	if err != nil {
		log.Println(err)
		return
	}

	needupdate := false
	// 1. Speed
	if req.Speed != nil && *req.Speed != cs.Speed {
		cs.Speed = *req.Speed
		needupdate = true
	}
	// 2. Heater
	if req.Heater != nil {
		if cs.HeaterEnabled != *req.Heater {
			cs.HeaterEnabled = *req.Heater
			needupdate = true
		}
	}
	// 3. Temp
	if req.Target != nil {
		if cs.TempTarget != *req.Target {
			cs.TempTarget = *req.Target
			needupdate = true
		}
	}
	// 4. Gate
	if req.Gate != nil {
		if cs.GateStatus() != *req.Gate {
			cs.SetGateStatus(*req.Gate)
			needupdate = true
		}
	}
	// 5. Sound
	if req.Sound != nil {
		if cs.SoundEnabled != *req.Sound {
			cs.SoundEnabled = *req.Sound
			needupdate = true
		}
	}
	// 77. State
	if req.On != nil {
		if cs.Enabled {
			if !*req.On {
				cs.Enabled = false
				needupdate = true
			} else {
				log.Println("Already on")
			}
		} else {
			if *req.On {
				cs.Enabled = true
				needupdate = true
			} else {
				log.Println("Already off")
			}
		}
	}

	if req.FilterRemains != nil {
		cs.FiltersRemains = *req.FilterRemains
		needupdate = true
	}

	if needupdate {
		if err = ts.t.Update(cs, timeout); err != nil {
			log.Println(err)
		} else {
			if err := ts.ctx.SendState(); err != nil {
				log.Println(err)
			} else {
				log.Println("Made update by MQTT request")
			}
		}
	}

	log.Println("Control done.")
}
