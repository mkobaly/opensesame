package main

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

type WifiStateChange struct {
	IsActive bool
}

var errCommandFailed = errors.New("Execuing command failed. Please retry")

type Manager struct {
	//client *insteon.Client
	insteon Insteon
	config  *Config
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
		//insteon: NewInsteonClient(config.Insteon.BaseURL, config.Insteon.Username, config.Insteon.Password),
		insteon: NewInsteonHomeLinkClient(config.Insteon.BaseURL),
		WiFi:    &WiFi{SSID: config.SSID, lastSeen: 99999},
		config:  config,
		logger:  logger,
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
			if r.lastSeen < 4000 {
				if !curActive {
					m.WiFiActive = true
					m.logger.Info("WiFi SSID found")
					return &WifiStateChange{IsActive: true}
				}
				return nil
			}
		}
	}
	if curActive {
		m.logger.Info("WiFi SSID lost")
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
	m.logger.WithField("isOpen", isOpen).WithField("wifi_active", m.WiFiActive).Info("Rebalancing Door")
	if err != nil {
		m.logger.Errorf("IsDoorOpen Error %v", err)
		return false
	}
	if m.WiFiActive {
		//m.logger.Info("WiFi is active")
		if !isOpen {
			m.logger.Info("Toggling Door")
			err := m.ToggleDoor()
			if err != nil {
				m.logger.Errorf("ToggleDoor Error %v", err)
			}
			return true
		}
	} else {
		//m.logger.Info("WiFi is not active")
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

	isOpen, err := m.insteon.IOLinkStatus(m.config.GarageID)
	if err == nil {
		return isOpen, err
	}
	m.logger.Errorf("IsDoorOpen->IOLinkStatus error %s\n", err.Error())

	//we need to loop because the commandStatus could be pending
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 1)
		m.logger.Infof("IsDoorOpen->Loop Count %d", i)
		isOpen, err = m.insteon.IOLinkStatus(m.config.GarageID)
		if err == nil {
			return isOpen, err
		}
		m.logger.Errorf("IsDoorOpen->IOLinkStatus error %s\n", err.Error())
	}
	return false, errors.New("Unable to determine door state")
}

//Authenticate will authenticate with Insteon API
// func (m *Manager) Authenticate() error {
// 	m.logger.Info("Authenticating..")
// 	err := m.client.Authenticate(m.config.Insteon.ClientID, m.config.Insteon.Username, m.config.Insteon.Password)
// 	if err == nil {
// 		m.LastAuthenticate = time.Now()
// 	}
// 	return err
// }

//RefreshToken will refresh the authentication token for Insteon API
// func (m *Manager) RefreshToken() error {
// 	m.logger.Info("Refreshing token..")
// 	err := m.client.RefreshToken()
// 	if err == nil {
// 		m.LastAuthenticate = time.Now()
// 	}
// 	return err
// }

//ToggleDoor will toggle the door to open or close
func (m *Manager) ToggleDoor() error {
	err := m.insteon.ToggleIOLink(m.config.GarageID)
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

// func withRetryDeley(retries int, delay int, logger *logrus.Entry, f func() error) (err error) {
// 	for i := 0; i < retries; i++ {
// 		err = f()
// 		if err != nil {
// 			logger.Errorf("Retry: %v / %v, error: %v", i, retries, err)
// 		} else {
// 			return
// 		}
// 	}
// 	return
// }
