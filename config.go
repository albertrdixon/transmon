package main

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/gearbox/url"
	"github.com/ghodss/yaml"
)

type Config struct {
	Timeout      *duration     `json:"timeout,omitempty"`
	PIA          *PIA          `json:"pia"`
	Transmission *Transmission `json:"transmission"`
	OpenVPN      *OpenVPN      `json:"openvpn"`
}

type PIA struct {
	User     string   `json:"user"`
	Pass     string   `json:"password"`
	ClientID string   `json:"client_id"`
	URL      *url.URL `json:"url"`
}

type Transmission struct {
	Command string   `json:"command"`
	UID     int      `json:"uid"`
	GID     int      `json:"gid"`
	URL     *url.URL `json:"url"`
	User    string   `json:"rpc_username"`
	Pass    string   `json:"rpc_password"`
}

type OpenVPN struct {
	Tun     string `json:"tun_device"`
	Command string `json:"command"`
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalJSON(p []byte) error {
	val := bytes.Trim(p, `"`)
	t, er := time.ParseDuration(string(val))
	if er != nil {
		return er
	}
	d.Duration = t
	return nil
}

func readConfig(file string) (*Config, error) {
	logger.Debugf("Reading config from %q", file)
	content, er := ioutil.ReadFile(file)
	if er != nil {
		return nil, er
	}

	c := new(Config)
	c.Timeout = &duration{Duration: 5 * time.Minute}
	return c, yaml.Unmarshal(content, c)
}
