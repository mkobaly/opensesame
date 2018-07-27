package main

import (
	"errors"
	"github.com/mkobaly/insteon"
	"time"
)

type WifiStateChange struct {
	IsActive bool
}

var errCommandFailed = errors.New("Execuing command failed. Please retry")

type Manager struct {
	client *insteon.Client
	config *Config
	//IsDoorOpen       bool
	WiFi             *WiFi
	WiFiActive       bool
	LastChecked      time.Time
	LastAuthenticate time.Time
	logger           *Logger
}

func NewManager(config *Config) *Manager {

	logger, _ := New("manager", 1)
	return &Manager{
		client: insteon.New(config.Insteon.BaseURL),
		WiFi:   &WiFi{SSID: config.SSID, lastSeen: 99999},
		config: config,
		logger: logger,
	}
}

func (m *Manager) Process(wifi []WiFi) *WifiStateChange {
	curActive := m.WiFiActive
	for _, r := range wifi {
		if r.SSID == m.config.SSID {
			m.WiFi.lastSeen = r.lastSeen
			m.WiFi.macAddress = r.macAddress
			m.WiFi.signal = r.signal
			m.logger.Infof("Wifi: %+v", r)
			if r.lastSeen < 8000 {
				if curActive == false {
					m.WiFiActive = true
					return &WifiStateChange{IsActive: true}
				}
				return nil
			}
		}
	}
	if curActive == true {
		m.logger.Info("Setting wifi to NOT active")
		m.WiFiActive = false
		return &WifiStateChange{IsActive: false}
	}
	return nil
}

func (m *Manager) Rebalance() bool {
	isOpen := false
	err := withRetry(3, m.logger, func() error {
		o, e := m.IsDoorOpen()
		isOpen = o
		return e
	})
	m.logger.Infof("Rebalance Door is open: %v", isOpen)
	if err != nil {
		m.logger.Errorf("IsDoorOpen Error %v", err)
		return false
	}
	if m.WiFiActive {
		m.logger.Info("WiFi Active")
		if !isOpen {
			m.logger.Info("Toggling Door")
			err := m.ToggleDoor()
			if err != nil {
				m.logger.Errorf("ToggleDoor Error %v", err)
			}
			return true
		}
	} else {
		m.logger.Info("WiFi not active")
		if isOpen {
			m.logger.Info("Toggling Door")
			err := m.ToggleDoor()
			if err != nil {
				m.logger.Errorf("ToggleDoor Error %v", err)
			}
			return true
		}
	}
	return false
}

func (m *Manager) IsDoorOpen() (bool, error) {
	resp, err := m.client.SendCommand("get_sensor_status", m.config.GarageID)
	if err != nil {
		m.logger.Errorf("IsDoorOpen->SendCommand error %s\n", err.Error())
		return false, err
	}

	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 1)
		m.logger.Infof("IsDoorOpen->Loop Count %d", i)
		cs, err := m.client.CommandStatus(resp.ID)
		if err != nil {
			m.logger.Errorf("IsDoorOpen->CommandStatus error %s", err.Error())
			m.logger.StackAsError("test")
			continue
		}
		m.logger.Infof("Door cmd status: %v", cs.Status)
		if cs.Status == "succeeded" {
			level := cs.Response["level"].(float64)
			m.logger.Infof("Door sensor level: %v (100 = closed, 0 = open)", level)
			return level == 0, nil
		}
		if cs.Status == "failed" {
			return false, errCommandFailed
		}
	}
	return false, errors.New("Unable to determine door state")
}

func (m *Manager) Authenticate() error {
	m.logger.Info("AUthenticating..")
	err := m.client.Authenticate(m.config.Insteon.ClientID, m.config.Insteon.Username, m.config.Insteon.Password)
	if err == nil {
		m.LastAuthenticate = time.Now()
	}
	return err
}

func (m *Manager) RefreshToken() error {
	m.logger.Info("Refreshing token..")
	err := m.client.RefreshToken()
	if err == nil {
		m.LastAuthenticate = time.Now()
	}
	return err
}

func (m *Manager) ToggleDoor() error {
	_, err := m.client.SendCommand("on", m.config.GarageID)
	if err != nil {
		return err
	}
	return nil
}

func withRetry(retries int, logger *Logger, f func() error) (err error) {
	for i := 0; i < retries; i++ {
		err = f()
		if err != nil {
			logger.Errorf("Retry: %v / %v, error: %v", i, retries, err)
		} else {
			return
		}
	}
	return
}
