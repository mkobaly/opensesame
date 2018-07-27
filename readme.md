
# OpenSesame

Utility to automatically open up your garage when your car rolls up.

I am running this on a raspberry pi.


## Prerequisits

- Insteon Home hub with Public API key (developer key) [link](https://www.insteon.com/developer/)
- IO Link to control garage door [link](https://www.smarthome.com/insteon-74551-garage-door-control-status-kit.html)
- Car that has wifi

## Configuration

- clientId - This is your developer api key from Insteon
- You will need to find your insteon device Id for your garage
- ssid - is the ssid of your car's wifi

```yaml
insteon:
  baseurl: "https://connect.insteon.com/api/v2"
  clientid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxxxxxxxxxxx.xxxxxxxx
  username: your.insteon.hub.username
  password: your.insteon.hub.passowrd
wifiinterface: wlan0
ssid: testtett
garageid: 34343434

```

## Building

Using govendor to vendor dependencies. To get up and running


```sh
go get github.com/mkobaly/insteon
govendor sync
./build.sh
```

## Modes of operation and rules

### Open Garage

- wifi recently showed up (last seen < XXX)
- garage door shut
- signal < 90 ?? (probably not needed)


### Close garage

- wifi lost or last seen > XXX
- garage door open
- signal > 90


## Insteon status

get_sensor_status = 100 (Door closed)