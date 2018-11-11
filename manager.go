package main

import (
	"errors"
	"github.com/mkobaly/insteon"
	"github.com/sirupsen/logrus"
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
	logger           *logrus.Entry
}

func NewManager(config *Config, logger *logrus.Entry) *Manager {
	//logger, _ := New("manager", 1)
	return &Manager{
		client: insteon.New(config.Insteon.BaseURL),
		WiFi:   &WiFi{SSID: config.SSID, lastSeen: 99999},
		config: config,
		logger: logger,
	}
}

//Process will cycle through all wifi's found and see if it
//matches the one wifi SSID we are interested in. If so it will
//report back a WifiStateChange
func (m *Manager) Process(wifi []WiFi) *WifiStateChange {
	curActive := m.WiFiActive
	for _, r := range wifi {
		if r.SSID == m.config.SSID {
			m.WiFi.lastSeen = r.lastSeen
			m.WiFi.macAddress = r.macAddress
			m.WiFi.signal = r.signal
			m.logger.WithFields(logrus.Fields{"Signal": r.signal, "LastSeen": r.lastSeen}).Info("Wifi Signal")
			if r.lastSeen < 8000 {
				if curActive == false {
					m.WiFiActive = true
					m.logger.Info("Setting wifi to active")
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

//Rebalance will trigger a change based on door status and wifi status
func (m *Manager) Rebalance() bool {
	m.logger.Info("Inside rebalance")
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

//IsDoorOpen will get the insteon status of the garage door sensor
func (m *Manager) IsDoorOpen() (bool, error) {
	resp, err := m.client.SendCommand("get_sensor_status", m.config.GarageID)
	if err != nil {
		m.logger.Errorf("IsDoorOpen->SendCommand error %s\n", err.Error())
		return false, err
	}

	//we need to loop because the commandStatus could be pending
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 1)
		m.logger.Infof("IsDoorOpen->Loop Count %d", i)
		cs, err := m.client.CommandStatus(resp.ID)
		if err != nil {
			m.logger.Errorf("IsDoorOpen->CommandStatus error %s", err.Error())
			//m.logger.StackAsError("test")
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

//Authenticate will authenticate with Insteon API
func (m *Manager) Authenticate() error {
	m.logger.Info("Authenticating..")
	err := m.client.Authenticate(m.config.Insteon.ClientID, m.config.Insteon.Username, m.config.Insteon.Password)
	if err == nil {
		m.LastAuthenticate = time.Now()
	}
	return err
}

//RefreshToken will refresh the authentication token for Insteon API
func (m *Manager) RefreshToken() error {
	m.logger.Info("Refreshing token..")
	err := m.client.RefreshToken()
	if err == nil {
		m.LastAuthenticate = time.Now()
	}
	return err
}

//ToggleDoor will toggle the door to open or close
func (m *Manager) ToggleDoor() error {
	_, err := m.client.SendCommand("on", m.config.GarageID)
	if err != nil {
		return err
	}
	return nil
}

func withRetry(retries int, logger *logrus.Entry, f func() error) (err error) {
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
