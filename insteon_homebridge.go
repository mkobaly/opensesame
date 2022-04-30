package main

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

//InsteonHomeLinkClient will use the rest api exposed by insteon hombridge plugin
//https://github.com/kuestess/homebridge-platform-insteonlocal
type InsteonHomeLinkClient struct {
	URL string
}

func NewInsteonHomeLinkClient(url string) *InsteonHomeLinkClient {
	return &InsteonHomeLinkClient{
		URL: strings.TrimSuffix(url, "/"),
	}
}

func (i *InsteonHomeLinkClient) IOLinkStatus(id string) (isOpen bool, err error) {
	isOpen = false
	client := i.getHttpClient()
	url := i.URL + "/iolinc/" + id + "/sensor_status"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return isOpen, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return isOpen, err
	}
	rawTxt := string(b)
	//sensor is reversed so off = open
	if strings.Contains(rawTxt, "off") {
		isOpen = true
	}
	return isOpen, nil
}

func (i *InsteonHomeLinkClient) ToggleIOLink(id string) error {
	client := i.getHttpClient()
	url := i.URL + "/iolinc/" + id + "/relay_on"
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("non 200 response from ToggleIOLink. Got " + resp.Status)
	}
	defer resp.Body.Close()
	return nil
}

func (i *InsteonHomeLinkClient) getHttpClient() *http.Client {
	tr := &http.Transport{
		DisableCompression: true,
	}
	return &http.Client{Transport: tr}
}
