package main

import (
	"errors"
	"flag"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	useSyslog := flag.Bool("syslog", false, "Use syslog")
	url := flag.String("url", "http://192.168.22.131", "Url to use as check")
	card := flag.String("card", "wlan0", "Network card to restart")
	minutes := flag.Int("minutes", 1, "wait this many minutes between check")
	flag.Parse()
	if *useSyslog {
		logwriter, e := syslog.New(syslog.LOG_NOTICE, "piwififixer")
		if e == nil {
			log.SetOutput(logwriter)
		}
	}
	log.Println("syslog", *useSyslog, "url", *url, "card", *card, "minutes", *minutes)

	ticker := time.NewTicker(time.Duration(*minutes) * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := checkInternet(*url); err != nil {
					log.Println("Wifi Down")

					resp, err := restartWifi(*card)
					if err != nil {
						log.Println("Failed restart wifi", err)
					} else {
						log.Println("Restarted Wifi", *card, resp)
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
	cmdArgs = []string{"--force", card}
	cmdOutput, err = exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		return "", err
	}
	return string(cmdOutput), nil
}

func checkInternet(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Wrong status code " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}
