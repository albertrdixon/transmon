package pia

import (
	"encoding/json"
	"net/http"
	ur "net/url"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/gearbox/url"
)

var endpoint string

const defaultEndpoint = `https://www.privateinternetaccess.com/vpninfo/port_forward_assignment`

func GetPortForwardEndpoint() *url.URL {
	ep := endpoint
	if ep == "" {
		ep = defaultEndpoint
	}
	u, _ := url.Parse(ep)
	return u
}

func SetPortForwardEndpoint(u *url.URL) {
	endpoint = u.String()
}

func RequestPort(ip, user, pass, id string) (int, error) {
	values := ur.Values{}
	values.Add("user", user)
	values.Add("pass", pass)
	values.Add("client_id", id)
	values.Add("local_ip", ip)

	ep := GetPortForwardEndpoint().String()
	logger.Debugf("POST %v", ep)
	resp, er := http.PostForm(ep, values)
	if er != nil {
		return 0, er
	}

	defer resp.Body.Close()
	port := new(response)
	er = json.NewDecoder(resp.Body).Decode(port)
	return port.Port, er
}
