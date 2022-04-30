package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomebridgeIOLinkStatus(t *testing.T) {
	cfgPath := "./config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return
	}
	content, _ := ioutil.ReadFile(cfgPath)
	cfg := LoadConfig(content)

	client := NewInsteonHomeLinkClient(cfg.Insteon.BaseURL)
	_, err := client.IOLinkStatus(cfg.GarageID)
	assert.NoError(t, err)
}

func TestHomebridgeToggleIoLink(t *testing.T) {
	cfgPath := "./config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return
	}
	content, _ := ioutil.ReadFile(cfgPath)
	cfg := LoadConfig(content)

	client := NewInsteonHomeLinkClient(cfg.Insteon.BaseURL)
	err := client.ToggleIOLink(cfg.GarageID)
	assert.NoError(t, err)
}
