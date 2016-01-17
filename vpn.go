package main

import (
	"errors"
	"net"
	"strings"

	"github.com/albertrdixon/gearbox/logger"
)

func findIP(vpn *OpenVPN) (string, error) {
	logger.Debugf("Looking up addresses for interface %q", vpn.Tun)
	inf, er := net.InterfaceByName(vpn.Tun)
	if er != nil {
		return "", er
	}
	addrs, er := inf.Addrs()
	if er != nil {
		return "", er
	}
	if len(addrs) < 1 {
		return "", errors.New("Interface has no addresses")
	}

	logger.Debugf("Addresses for %q: %v", vpn.Tun, addrs)
	bits := strings.SplitN(addrs[0].String(), "/", 2)
	return bits[0], nil
}
