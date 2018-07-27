package main

import (
	cmdline "github.com/galdor/go-cmdline"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	logger, _ := New("main", 1)
	//using config file to hold configuration
	cmdline := cmdline.New()
	cmdline.AddOption("c", "config", "config.yaml", "Path to configuration file")
	cmdline.Parse(os.Args)

	cfgPath := "./config.yaml"
	if cmdline.IsOptionSet("c") {
		cfgPath = cmdline.OptionValue("c")
	}

	//no config file so create one
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		config := NewConfig()
		config.Write(cfgPath)
		logger.Notice("config.yaml not present. One was just created for you. Please edit it accordingly")
		os.Exit(0)
	}

	//read config content and load it
	content, _ := ioutil.ReadFile(cfgPath)
	config := LoadConfig(content)

	wiFiFetcher := WiFiReal{}
	manager := NewManager(config)
	//confirm Isteon access
	err := manager.Authenticate()
	if err != nil {
		logger.Errorf("Error logging into Isteon: %v", err)
		os.Exit(1)
	}
	logger.Info("Scanning...")
	run(logger, manager, wiFiFetcher)
}

func run(logger *Logger, manager *Manager, wifi WiFiFetcher) {
	//loop forever and poll wifi
	_, _, day := time.Now().Date()
	for {
		wifi := wifi.Fetch(manager.config.WifiInterface)
		stateChange := manager.Process(wifi)
		if stateChange != nil {
			logger.Info("Rabalancing...")
			changeHappened := manager.Rebalance()
			if changeHappened {
				//toggling garage takes some time so need to wait
				time.Sleep(time.Second * 10)
			}
		}
		time.Sleep(time.Second * 4)
		_, _, curDay := time.Now().Date()
		if curDay != day {
			logger.Info("Refreshing token")
			manager.RefreshToken()
			day = curDay
		}
	}
}
