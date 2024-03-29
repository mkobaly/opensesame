package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	cmdline "github.com/galdor/go-cmdline"
	"github.com/sirupsen/logrus"
)

func main() {

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
		fmt.Println("config.yaml not present. One was just created for you. Please edit it accordingly")
		os.Exit(0)
	}

	//read config content and load it
	content, _ := ioutil.ReadFile(cfgPath)
	config := LoadConfig(content)

	options := LoggerOptions{
		Application: "opensesame",
		LogFile:     config.Logfile,
	}
	log := NewLogger(options)

	wiFiFetcher := WiFiReal{}
	manager := NewManager(config, log)
	// //confirm Isteon access
	// err := manager.Authenticate()
	// if err != nil {
	// 	log.Errorf("Error logging into Isteon: %v", err)
	// 	os.Exit(1)
	// }
	log.Info("Scanning...")
	run(log, manager, wiFiFetcher)
}

func run(logger *logrus.Entry, manager *Manager, wifi WiFiFetcher) {
	//loop forever and poll wifi
	//_, _, day := time.Now().Date()
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
		time.Sleep(time.Second * 1)
	}
}
