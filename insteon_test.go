package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInsteonStatusBuffer(t *testing.T) {

	_, err := NewInsteonStatusBuffer("3488DD", "02623488DD0F19010602503488DD39E088210000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028")
	assert.NoError(t, err)
}

func TestNewInsteonStatusBufferWithWrongDevice(t *testing.T) {
	_, err := NewInsteonStatusBuffer("3488AD", "02623488DD0F19010602503488DD39E088210000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028")
	assert.Error(t, err)
	assert.Equal(t, "Status is not for this device. Got 3488DD wanted 3488AD", err.Error())
}

func TestGetBuffer(t *testing.T) {
	cfgPath := "./config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return
	}
	content, _ := ioutil.ReadFile(cfgPath)
	cfg := LoadConfig(content)

	client := NewInsteonClient(cfg.Insteon.BaseURL, cfg.Insteon.Username, cfg.Insteon.Password)
	result, err := client.bufferStatus()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestIOLinkStatus(t *testing.T) {
	cfgPath := "./config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return
	}
	content, _ := ioutil.ReadFile(cfgPath)
	cfg := LoadConfig(content)

	client := NewInsteonClient(cfg.Insteon.BaseURL, cfg.Insteon.Username, cfg.Insteon.Password)
	_, err := client.IOLinkStatus(cfg.GarageID)
	assert.NoError(t, err)
}

func TestToggleIoLink(t *testing.T) {
	cfgPath := "./config.yaml"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return
	}
	content, _ := ioutil.ReadFile(cfgPath)
	cfg := LoadConfig(content)

	client := NewInsteonClient(cfg.Insteon.BaseURL, cfg.Insteon.Username, cfg.Insteon.Password)
	err := client.ToggleIOLink(cfg.GarageID)
	assert.NoError(t, err)
}
