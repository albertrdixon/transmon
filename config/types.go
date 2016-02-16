package config

import (
	"time"

	"github.com/albertrdixon/gearbox/url"
)

type Config struct {
	Timeout      *duration `json:"timeout,omitempty"`
	Cleaner      *Cleaner
	PIA          *PIA          `json:"pia"`
	Transmission *Transmission `json:"transmission"`
	OpenVPN      *OpenVPN      `json:"openvpn"`
	modTime      time.Time
	file         string
}

type Cleaner struct {
	Enabled  bool
	Interval *duration
}

type PIA struct {
	User     string   `json:"username"`
	Pass     string   `json:"password"`
	ClientID string   `json:"client_id"`
	URL      *url.URL `json:"url"`
}

type Transmission struct {
	Command          string `json:"command"`
	UID              int    `json:"uid"`
	GID              int    `json:"gid"`
	Config           string `json:"config"`
	*TransmissionRPC `json:"rpc"`
}

type TransmissionRPC struct {
	URL  *url.URL `json:"url"`
	User string   `json:"username"`
	Pass string   `json:"password"`
}

type OpenVPN struct {
	Tun     string `json:"device"`
	Command string `json:"command"`
}

type duration struct {
	time.Duration
}
