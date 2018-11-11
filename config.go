package main

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
)

type Credential struct {
	BaseURL  string
	ClientID string
	Username string
	Password string
}

type Config struct {
	Insteon       Credential
	WifiInterface string
	SSID          string
	GarageID      int
	Logfile       string
}

//NewConfig creates a new Configuration object needed
func NewConfig() *Config {
	//config := Config{}
	return &Config{
		Insteon:       Credential{BaseURL: "https://connect.insteon.com/api/v2", Username: "youremail@example.com", Password: "password", ClientID: "api_key_from_insteon"},
		WifiInterface: "wlan0",
		SSID:          "ssid_to_track",
		GarageID:      1118511,
		Logfile:       "/var/log/opensesame.log",
	}
}

//LoadConfig will load up a Config object based on configPath
func LoadConfig(data []byte) *Config {
	var config = new(Config)
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err.Error())
	}
	return config
}

// Write will save the configuration to the given path
func (c *Config) Write(path string) error {
	bytes, err := yaml.Marshal(c)
	if err == nil {
		return ioutil.WriteFile(path, bytes, 0777)
	}
	return err
}
