package main

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Wifi represents a wifi access point
type WiFi struct {
	macAddress string
	SSID       string
	signal     int
	lastSeen   int
}

type WiFiFetcher interface {
	Fetch(wifiInterface string) []WiFi
}

type WiFiReal struct {
}

func (w WiFiReal) Fetch(wifiInterface string) []WiFi {
	var results []WiFi

	s, _ := RunCommand(10*time.Second, "sudo /sbin/iw dev "+wifiInterface+" scan -u")
	var current WiFi
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "BSS") {
			mac := strings.Split(strings.Split(line, "(")[0], "BSS")[1]
			mac = strings.ToLower(mac)
			mac = strings.TrimSpace(mac)
			current = WiFi{macAddress: mac}
		}
		if strings.HasPrefix(line, "signal:") {
			sig := strings.Split(line, "signal:")[1]
			sig = strings.Split(sig, ".")[0]
			sig = strings.TrimSpace(sig)
			var err error
			signal, err := strconv.Atoi(sig)
			if err != nil {
				signal = -100
			}
			current.signal = signal
		}
		if strings.HasPrefix(line, "last seen:") {
			ls := strings.Split(line, "last seen:")[1]
			ls = strings.TrimSpace(ls)
			ls = strings.Split(ls, " ")[0]
			var err error
			lastSeen, err := strconv.Atoi(ls)
			if err != nil {
				lastSeen = 999
			}
			current.lastSeen = lastSeen
		}
		if strings.HasPrefix(line, "SSID:") {
			ssid := strings.Split(line, "SSID:")[1]
			ssid = strings.TrimSpace(ssid)
			current.SSID = ssid
			results = append(results, current)
		}

	}
	return results
}

func RunCommand(tDuration time.Duration, commands string) (string, string) {
	command := strings.Fields(commands)
	cmd := exec.Command(command[0])
	if len(command) > 0 {
		cmd = exec.Command(command[0], command[1:]...)
	}
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(tDuration):
		if err := cmd.Process.Kill(); err != nil {
			log.Println("failed to kill: ", err)
		}
		log.Printf("%s killed as timeout reached\n", commands)
	case err := <-done:
		if err != nil {
			log.Printf("%s: %s\n", err.Error(), commands)
		}
	}
	return strings.TrimSpace(outb.String()), strings.TrimSpace(errb.String())
}
