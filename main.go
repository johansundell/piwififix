package main

import (
	"flag"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	useSyslog := flag.Bool("syslog", false, "Use syslog")
	flag.Parse()
	if *useSyslog {
		logwriter, e := syslog.New(syslog.LOG_NOTICE, "cocsniffer")
		if e == nil {
			log.SetOutput(logwriter)
		}
	}

	site := &Site{"http://192.168.22.131", UNCHECKED}

	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				s, err := site.Status()
				if err != nil || s == DOWN {
					log.Println("Wifi Down")
					card := "wlan0"
					resp, err := restartWifi(card)
					if err != nil {
						log.Println("Failed restart wifi", err)
					} else {
						log.Println("Restarted Wifi", card, resp)
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	close(quit)
	log.Println("Bye ;)")

}

func restartWifi(card string) (string, error) {
	// ifdown wlan0
	cmdName := "ifdown"
	cmdArgs := []string{card}
	cmdOutput, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return "", err
	}

	// ifup --force wlan0
	cmdName = "ifup"
	cmdArgs = []string{"--force", "wlan0"}
	cmdOutput, err = exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return "", err
	}
	return string(cmdOutput), nil
}

type Status int

const (
	UNCHECKED Status = iota
	DOWN
	UP
)

// The Site struct encapsulates the details about the site being monitored.
type Site struct {
	url         string
	last_status Status
}

// Site.Status makes a GET request to a given URL and checks whether or not the
// resulting status code is 200.
func (s Site) Status() (Status, error) {
	resp, err := http.Get(s.url)
	status := s.last_status

	if (err == nil) && (resp.StatusCode == 200) {
		status = UP
	} else {
		status = DOWN
	}

	return status, err
}
