package main

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type InsteonStatusBuffer struct {
	PassThrough     string
	ID              string
	Flags           string
	CMD1            string
	CMD2            string
	Echo            string
	InsteonReceived string
	From            string
	To              string
	HopCount        string
	Delta           string
	OnLevel         string
}

func NewInsteonStatusBuffer(deviceId string, value string) (*InsteonStatusBuffer, error) {
	buff := &InsteonStatusBuffer{
		PassThrough:     value[0:4],
		ID:              value[4:10],
		Flags:           value[10:12],
		CMD1:            value[12:14],
		CMD2:            value[14:16],
		Echo:            value[16:18],
		InsteonReceived: value[18:22],
		From:            value[22:28],
		To:              value[28:34],
		HopCount:        value[34:36],
		Delta:           value[36:38],
		OnLevel:         value[38:40],
	}
	if buff.PassThrough != "0262" {
		return nil, errors.New("Passthrough has wrong value " + buff.PassThrough)
	}
	if buff.ID != deviceId {
		return nil, errors.New("Status is not for this device. Got " + buff.ID + " wanted " + deviceId)
	}
	if buff.Echo != "06" {
		return nil, errors.New("Echo has wrong value " + buff.Echo)
	}
	if buff.InsteonReceived != "0250" {
		return nil, errors.New("InsteonReceived has wrong value " + buff.InsteonReceived)
	}

	return buff, nil
}

type Insteon interface {
	IOLinkStatus(id string) (isOpen bool, err error)
	ToggleIOLink(id string) error
}

//InsteonClient is talking to actual insteon hub. There can only be one process talking
//to the hub at a time so if this is used you can't use homelink for example
type InsteonClient struct {
	URL      string
	Username string
	Password string
}

type buffResponse struct {
	XMLName xml.Name `xml:"response"`
	BS      string
}

func NewInsteonClient(url string, username string, password string) *InsteonClient {
	return &InsteonClient{
		URL:      strings.TrimSuffix(url, "/"),
		Username: username,
		Password: password,
	}
}

func (i *InsteonClient) bufferStatus() (result string, err error) {
	client := i.getHttpClient()
	url := i.URL + "/buffstatus.xml"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(i.Username, i.Password)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var buffResp buffResponse
	err = xml.Unmarshal(data, &buffResp)
	if err != nil {
		return "", err
	}

	return buffResp.BS, nil
}

func (i *InsteonClient) IOLinkStatus(id string) (isOpen bool, err error) {

	isOpen = false
	client := i.getHttpClient()
	url := i.URL + "/3?0262" + id + "0F1901=I=3"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(i.Username, i.Password)
	resp, err := client.Do(req)
	if err != nil {
		return isOpen, err
	}
	defer resp.Body.Close()

	rawStatus, err := i.bufferStatus()
	if err != nil {
		return isOpen, err
	}
	isb, err := NewInsteonStatusBuffer(id, rawStatus)
	if err != nil {
		return isOpen, err
	}
	isOpen = isb.OnLevel == "00"
	return isOpen, nil
}

func (i *InsteonClient) ToggleIOLink(id string) error {

	client := i.getHttpClient()
	url := i.URL + "/3?0262" + id + "0F11FF=I=3"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(i.Username, i.Password)

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

func (i *InsteonClient) getHttpClient() *http.Client {
	tr := &http.Transport{
		DisableCompression: true,
	}
	return &http.Client{Transport: tr}
}
