package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {

	options := LoggerOptions{
		Application: "opensesame",
	}
	log := NewLogger(options)
	cfg := NewConfig()
	mgr := NewManager(cfg, log)
	if mgr.WiFi.lastSeen != 99999 {
		t.Errorf("Expected lastSeen of 99999 but but got: %d", mgr.WiFi.lastSeen)
	}
}

func TestProcessNotActive(t *testing.T) {
	options := LoggerOptions{
		Application: "opensesame",
	}
	log := NewLogger(options)
	cfg := NewConfig()
	mgr := NewManager(cfg, log)
	wifi := []WiFi{
		WiFi{
			SSID:     "foo",
			lastSeen: 500,
		},
		WiFi{
			SSID:     "bar",
			lastSeen: 500,
		},
	}
	mgr.Process(wifi)
	if mgr.WiFiActive == true {
		t.Error("Expected WifiActive false but got true")
	}
}

func TestProcessActive(t *testing.T) {
	options := LoggerOptions{
		Application: "opensesame",
	}
	log := NewLogger(options)
	cfg := NewConfig()
	mgr := NewManager(cfg, log)
	wifi := []WiFi{
		WiFi{
			SSID:     "ssid",
			lastSeen: 500,
		},
		WiFi{
			SSID:     "bar",
			lastSeen: 500,
		},
	}
	mgr.Process(wifi)
	if mgr.WiFiActive == false {
		t.Error("Expected WifiActive true but got false")
	}
}

func TestIsDoorOpen(t *testing.T) {
	options := LoggerOptions{
		Application: "opensesame",
	}
	log := NewLogger(options)
	cfgPath := "./config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return
	}
	content, _ := ioutil.ReadFile(cfgPath)
	cfg := LoadConfig(content)
	mgr := NewManager(cfg, log)

	_, err := mgr.IsDoorOpen()
	assert.NoError(t, err)
}
