package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/albertrdixon/gearbox/logger"
)

type piaResponse struct {
	Port int `json:"port"`
}

func requestPort(ip string, p *PIA) (int, error) {
	values := url.Values{}
	values.Add("user", p.User)
	values.Add("pass", p.Pass)
	values.Add("client_id", p.ClientID)
	values.Add("local_ip", ip)

	logger.Debugf("POST %v", p.URL)
	resp, er := http.PostForm(p.URL.String(), values)
	if er != nil {
		return 0, er
	}

	defer resp.Body.Close()
	port := new(piaResponse)
	er = json.NewDecoder(resp.Body).Decode(port)
	return port.Port, er
}
