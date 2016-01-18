package vpn

import (
	"errors"
	"net"
	"strings"

	"github.com/albertrdixon/gearbox/logger"
)

func FindIP(dev string) (string, error) {
	logger.Debugf("Looking up addresses for interface %q", dev)
	inf, er := net.InterfaceByName(dev)
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

	logger.Debugf("Addresses for %q: %v", dev, addrs)
	bits := strings.SplitN(addrs[0].String(), "/", 2)
	return bits[0], nil
}
