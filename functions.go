package main

import (
	"time"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/transmon/config"
	"github.com/albertrdixon/transmon/pia"
	"github.com/albertrdixon/transmon/process"
	"github.com/albertrdixon/transmon/transmission"
	"github.com/albertrdixon/transmon/vpn"
	"github.com/cenkalti/backoff"
	"golang.org/x/net/context"
)

func portUpdate(c *config.Config, ctx context.Context) error {
	ip, er := getIP(c.OpenVPN.Tun, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return er
	}
	logger.Infof("%v: inet %v", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA.User, c.PIA.Pass, c.PIA.ClientID, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return er
	}

	logger.Infof("New transmission port: %d", port)
	notify := func(e error, w time.Duration) {
		logger.Debugf("Failed to update transmission port: %v", er)
	}
	operation := func() error {
		select {
		default:
			t := transmission.NewRawClient(c.Transmission.URL.String(), c.Transmission.User, c.Transmission.Pass)
			return t.UpdatePort(port)
		case <-ctx.Done():
			return nil
		}
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = c.Timeout.Duration
	return backoff.RetryNotify(operation, b, notify)
}

func startProcesses(c *config.Config, ctx context.Context) (*process.Process, *process.Process, error) {
	t, er := process.New("transmission", c.Transmission.Command)
	if er != nil {
		return nil, nil, er
	}
	v, er := process.New("openvpn", c.OpenVPN.Command)
	if er != nil {
		return nil, nil, er
	}
	t.SetUser(c.Transmission.UID, c.Transmission.GID)

	logger.Infof(`Starting openvpn: %s`, c.OpenVPN.Command)
	if er := v.Execute(ctx); er != nil {
		return nil, nil, er
	}

	ip, er := getIP(c.OpenVPN.Tun, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return nil, nil, er
	}
	logger.Infof("%v: inet %v", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA.User, c.PIA.Pass, c.PIA.ClientID, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return nil, nil, er
	}
	logger.Infof("New transmission port: %d", port)

	if er := transmission.UpdateSettings(c.Transmission.Config, ip, port); er != nil {
		return nil, nil, er
	}

	logger.Infof("Starting transmission: %s", c.Transmission.Command)
	if er := t.Execute(ctx); er != nil {
		return nil, nil, er
	}

	return t, v, nil
}

func getPort(ip, user, pass, id string, timeout time.Duration, c context.Context) (int, error) {
	var port int
	notify := func(e error, w time.Duration) {
		logger.Errorf("Failed to get port from PIA (retry in %v): %v", w, e)
	}
	fn := func() error {
		select {
		default:
			p, er := pia.RequestPort(ip, user, pass, id)
			if er != nil {
				return er
			}
			port = p
			return nil
		case <-c.Done():
			return nil
		}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = timeout
	return port, backoff.RetryNotify(fn, b, notify)
}

func getIP(dev string, timeout time.Duration, c context.Context) (string, error) {
	var address string
	notify := func(e error, w time.Duration) {
		logger.Errorf("Failed to get IP for %q (retry in %v): %v", dev, w, e)
	}
	fn := func() (er error) {
		select {
		default:
			address, er = vpn.FindIP(dev)
			return
		case <-c.Done():
			return
		}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = timeout
	return address, backoff.RetryNotify(fn, b, notify)
}
