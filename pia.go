package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/albertrdixon/gearbox/logger"
)

// type pia struct {
// 	*http.Request
// 	timeout time.Duration
// }

// func newPIA(timeout time.Duration, url *url.URL, body string) (*pia, error) {
// 	req, er := http.NewRequest("POST", url.String(), strings.NewReader(body))
// 	if er != nil {
// 		return nil, er
// 	}
// 	return &pia{
// 		Request: req,
// 		timeout: timeout,
// 	}, nil
// }

// func (p *pia) getPort() (int, error) {

// }

type piaResponse struct {
	Port int `json:"port"`
}

func requestPort(ip string, p *PIA) (int, error) {
	values := url.Values{}
	values.Add("user", p.User)
	values.Add("pass", p.Pass)
	values.Add("client_id", p.ClientID)
	values.Add("local_ip", ip)

	logger.Debugf("POST %v form=%v", p.URL, values.Encode())
	resp, er := http.PostForm(p.URL.String(), values)
	if er != nil {
		return 0, er
	}

	defer resp.Body.Close()
	// body, _ := ioutil.ReadAll(resp.Body)
	// logger.Debugf("Response: [%d] %v", resp.StatusCode, string(body))
	port := new(piaResponse)
	er = json.NewDecoder(resp.Body).Decode(port)
	return port.Port, er
}
