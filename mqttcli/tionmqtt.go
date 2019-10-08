package mqttcli

import (
	"errors"
	"time"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"encoding/json"
	"fmt"
	"log"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/m-pavel/go-tion/tion"
)

type mqttTion struct {
	url  string
	user string
	pass string
	ca   string

	topic  string
	topicc string
	topica string
	cli    MQTT.Client

	debug bool

	state  *tion.RestStatus
	update chan *tion.RestStatus
}

func New(url, user, pass string, ca string, topic, topica, topicc string, dbg bool) tion.Tion {
	mqt := mqttTion{url: url, user: user, pass: pass, ca: ca, topic: topic, topica: topica, topicc: topicc, debug: dbg}
	mqt.update = make(chan *tion.RestStatus)
	return &mqt
}

func (mqt *mqttTion) Connect(timeout time.Duration) error {
	opts := MQTT.NewClientOptions().AddBroker(mqt.url)

	opts.SetClientID(fmt.Sprintf("mqtt-tion-cli-%d", time.Now().Unix()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	if mqt.user != "" {
		opts.Username = mqt.user
		opts.Password = mqt.pass
	}

	if mqt.ca != "" {
		tlscfg := tls.Config{}
		tlscfg.RootCAs = x509.NewCertPool()
		var b []byte
		var err error
		if b, err = ioutil.ReadFile(mqt.ca); err != nil {
			return err
		}
		if ok := tlscfg.RootCAs.AppendCertsFromPEM(b); !ok {
			return errors.New("failed to parse root certificate")
		}
		opts.SetTLSConfig(&tlscfg)
	}

	mqt.cli = MQTT.NewClient(opts)
	if token := mqt.cli.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	mqt.cli.Subscribe(mqt.topic, 1, mqt.handleState)
	mqt.cli.Subscribe(mqt.topica, 1, mqt.handleAlive)
	return nil
}

func (mqt *mqttTion) handleState(cli MQTT.Client, msg MQTT.Message) {
	if mqt.debug {
		log.Printf("Got MQTT status update %v", msg)
	}
	mqt.state = &tion.RestStatus{}
	if err := json.Unmarshal(msg.Payload(), mqt.state); err != nil {
		mqt.state = nil
		log.Println(err)
	}
	mqt.update <- mqt.state
}
func (mqt *mqttTion) handleAlive(cli MQTT.Client, msg MQTT.Message) {
	if mqt.debug {
		log.Printf("Got MQTT alive update %v", msg)
	}
}

func (mqt *mqttTion) ReadState(timeout time.Duration) (*tion.Status, error) {
	if mqt.state == nil {
		select {
		case <-mqt.update:
			break
		case <-time.After(timeout):
			log.Printf("Timeout %d reached.\n", timeout)

		}
		if mqt.state == nil {
			return nil, errors.New("Not recieved")
		}
	}
	return tion.StatusFromRest(mqt.state), nil
}
func (mqt *mqttTion) Update(s *tion.Status, timeout time.Duration) error {
	b, err := json.Marshal(tion.RestFromStatus(s))
	if err != nil {
		return err
	}
	if token := mqt.cli.Publish(mqt.topicc, 1, false, b); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mqt *mqttTion) Disconnect() error {
	mqt.cli.Disconnect(200)
	return nil
}