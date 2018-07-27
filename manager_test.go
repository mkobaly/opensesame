package main

import (
	"io/ioutil"
	"testing"
)

func TestNewManager(t *testing.T) {
	cfg := NewConfig()
	mgr := NewManager(cfg)
	if mgr.WiFi.lastSeen != 99999 {
		t.Errorf("Expected lastSeen of 99999 but but got: %d", mgr.WiFi.lastSeen)
	}
}

func TestProcessNotActive(t *testing.T) {
	cfg := NewConfig()
	mgr := NewManager(cfg)
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
	cfg := NewConfig()
	mgr := NewManager(cfg)
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
	content, _ := ioutil.ReadFile("config.yaml")
	config := LoadConfig(content)
	manager := NewManager(config)
	err := manager.Authenticate()
	if err != nil {
		t.Errorf("expected no error but got one %s", err.Error())
	}
	_, err = manager.IsDoorOpen()
	if err != nil {
		t.Errorf("expected no error but got one %s", err.Error())
	}
}
