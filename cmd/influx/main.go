package main

import (
	"flag"
	"log"
	_ "net/http"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	"github.com/m-pavel/go-tion/impl"

	"fmt"

	"net/http"

	"github.com/influxdata/influxdb1-client/v2"

	"github.com/m-pavel/go-tion/tion"
	"github.com/sevlyar/go-daemon"
)

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func main() {
	var logf = flag.String("log", "influxexport.log", "log")
	var pid = flag.String("pid", "influxexport.pid", "pid")
	var notdaemonize = flag.Bool("n", false, "Do not do to background.")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var iserver = flag.String("influx", "http://localhost:8086", "Influx DB endpoint")
	var device = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")
	var interval = flag.Int("interval", 30, "Interval secons")
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	cntxt := &daemon.Context{
		PidFileName: *pid,
		PidFilePerm: 0644,
		LogFileName: *logf,
		LogFilePerm: 0640,
		WorkDir:     "/tmp",
		Umask:       027,
		Args:        os.Args,
	}

	if !*notdaemonize && len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %v", err)
		}
		if err := daemon.SendCommands(d); err != nil {
			log.Println(err)
		}
		return
	}

	if !*notdaemonize {
		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal(err)
		}
		if d != nil {
			return
		}
	}

	daemonf(*iserver, *device, *interval)

}

func daemonf(iserver, device string, interval int) {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	var err error
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: iserver,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	t := impl.NewTionImpl(device, false, nil)

	erinr := 0
	for {
		select {
		case <-stop:
			log.Println("Exiting")
			break
		case <-time.After(time.Duration(interval) * time.Second):
			s, err := t.ReadState(7)
			if err != nil {
				log.Println(err)
				erinr++
			} else {
				reportInflux(cli, s)
				erinr = 0
			}
			if erinr == 10 {
				done <- struct{}{}
				return
			}
		}
	}
}

func reportInflux(i client.Client, s *tion.Status) {
	status := 0
	if s.Enabled {
		status = 1
	}
	point, err := client.NewPoint("tion",
		map[string]string{
			"gate":   s.GateStatus(),
			"on":     fmt.Sprintf("%v", s.Enabled),
			"heater": fmt.Sprintf("%v", s.HeaterEnabled),
		},
		map[string]interface{}{
			"out":    s.TempIn,
			"in":     s.TempOut,
			"tgt":    s.TempTarget,
			"spd":    s.Speed,
			"status": status,
		},
		time.Now())
	if err != nil {
		log.Printf("Insert data error: %v", err)
		return
	}
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "tion",
		Precision: "s",
	})
	if err != nil {
		log.Printf("Insert data error: %v", err)
		return
	}
	bp.AddPoint(point)
	err = i.Write(bp)
	if err != nil {
		log.Printf("Insert data error: %v", err)
	}
}

func termHandler(sig os.Signal) error {
	log.Println("Terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}
