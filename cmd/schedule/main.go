package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/m-pavel/go-tion/impl"

	"github.com/m-pavel/go-tion/tion"

	"fmt"

	"github.com/gorhill/cronexpr"

	"github.com/sevlyar/go-daemon"
)

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

const timeout = 7 * time.Second
const status = "%2d | %20s | %7s | %6s | %5s | %4s | %5s | %4s | %s |\n"
const statush = "ID |       SCHEDULE       | ENABLED | HEATER | SOUND | TEMP | SPEED | GATE |         NEXT RUN         |\n"

func main() {
	var logf = flag.String("log", "schedule.log", "log file name")
	var pid = flag.String("pid", "schedule.pid", "pid")
	var notdaemonize = flag.Bool("n", false, "Do not do to background.")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var db = flag.String("db", "schedule.db", "Schedule db")
	var device = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")

	var prepare = flag.Bool("prepare", false, "Prepare database")

	var list = flag.Bool("list", false, "list")
	var del = flag.Int("del", -1, "Delete entry with given ID")

	var on = flag.Bool("on", false, "On")
	var off = flag.Bool("off", false, "Off")
	var schedule = flag.String("schedule", "", "Add schedule")
	var temp = flag.Int("temp", -1, "Temperature target")
	var heater = flag.String("heater", "", "Heater")
	var sound = flag.String("sound", "", "Sound")
	var gate = flag.String("gate", "", "indoor|mixed|outdoor")
	var speed = flag.Int("speed", -1, "speed")

	var repeat = flag.Int("repeat", 3, "repeat")
	var summer = flag.Bool("summer", false, "summer")
	var winter = flag.Bool("winter", false, "winter")

	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	dao, err := New(*db)
	if err != nil {
		log.Println(err)
		stop <- struct{}{}
		return
	}
	defer dao.Close()

	if *prepare {
		err = dao.Prepare()
		if err != nil {
			log.Println(err)
		}
		return
	}

	if *list {
		s, err := dao.GetSchedules()
		if err != nil {
			log.Println(err)
		}
		log.Println(statush)

		fi := func(v *int) string {
			if v == nil {
				return "n/a"
			}
			return fmt.Sprintf("%d", *v)
		}
		fg := func(v *int) string {
			if v == nil {
				return "n/a"
			}
			return tion.GateStatus(int8(*v))
		}
		for _, sch := range s {
			expr := cronexpr.MustParse(sch.Value).Next(time.Now())
			fmt.Printf(status, sch.ID, sch.Value, fb(sch.Enabled), fb(sch.Heater), fb(sch.Sound), fi(sch.Temp), fi(sch.Speed), fg(sch.Gate), expr.Format("Mon Jan _2 15:04:05 2006"))
		}
		return
	}
	if *del != -1 {
		err = dao.Delete(*del)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if *schedule != "" {
		var enb, htr, snd *bool
		var trueAddr, falseAddr bool
		trueAddr = true
		falseAddr = false
		if *heater == "on" {
			htr = &trueAddr
		}
		if *heater == "off" {
			htr = &falseAddr
		}
		if *sound == "on" {
			snd = &trueAddr
		}
		if *sound == "off" {
			snd = &falseAddr
		}
		if *on {
			enb = &trueAddr
		}
		if *off {
			enb = &falseAddr
		}
		var gt *int
		if *gate != "" {
			s := tion.Status{}
			s.SetGateStatus(*gate)
			iv := int(s.Gate)
			gt = &iv
		}
		if *temp == -1 {
			temp = nil
		}
		if *speed == -1 {
			speed = nil
		}

		if err = dao.Add(*schedule, enb, htr, snd, gt, speed, temp); err != nil {
			log.Println(err)
		}
		return
	}

	if *summer {
		if err := dao.UpdateHeater(false); err != nil {
			log.Println(err)
		}
	}
	if *winter {
		if err := dao.UpdateHeater(true); err != nil {
			log.Println(err)
		}
	}

	log.Println("Running daemon")
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)

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

	daemonf(*device, dao, *repeat)

}

func fb(v *bool) string {
	if v == nil {
		return "n/a"
	}
	if *v {
		return "on"
	}
	return "off"
}

func daemonf(device string, dao *Dao, repeat int) {
	sch, err := dao.GetSchedules()
	if err != nil {
		log.Println(err)
		return
	}
	for i := range sch {
		go func(s Schedule) {
			for {
				expr := cronexpr.MustParse(s.Value).Next(time.Now())
				mins := time.Until(expr) / time.Minute
				log.Printf("Next time for %d (%s) is %s in %d minute(s).\n", s.ID, fb(s.Enabled), expr.Format("Mon Jan _2 15:04:05 2006"), mins)
				select {
				case <-stop:
					log.Println("Exiting")
					break
				case <-time.After(time.Until(expr)):
					log.Printf("Executing %d\n", s.ID)
					for i := 0; i < repeat; i++ {
						err := execute(s, device, 5, 5*time.Second)
						if err != nil {
							log.Println(err)
						} else {
							break
						}
					}
				}
			}
		}(sch[i])
	}
	done <- struct{}{}
}

func execute(s Schedule, device string, retry int, interval time.Duration) error {
	t := impl.NewTionImpl(device, false, nil)
	if err := t.Connect(timeout); err != nil {
		return err
	}
	defer func() {
		if err := t.Disconnect(timeout); err != nil {
			log.Println(err)
		}
	}()
	ts, err := t.ReadState(timeout)
	if err != nil {
		return err
	}

	if s.Enabled != nil {
		ts.Enabled = *s.Enabled
	}
	if s.Gate != nil {
		ts.Gate = int8(*s.Gate)
	}
	if s.Temp != nil {
		ts.TempTarget = int8(*s.Temp)
	}
	if s.Speed != nil {
		ts.Speed = int8(*s.Speed)
	}
	if s.Heater != nil {
		ts.HeaterEnabled = *s.Heater
	}
	if s.Sound != nil {
		ts.SoundEnabled = *s.Sound
	}
	log.Printf("Device request %v\n", ts)

	i := 0
	for ; i < retry; i++ {
		err = t.Update(ts, timeout)
		if err == nil {
			break
		}
		time.Sleep(interval)
	}
	if err != nil {
		log.Printf("Device update failed after %d retries with error %v\n", i, err)
	} else {
		log.Printf("Device updated after %d retries.\n", i)
	}
	return err
}

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}
