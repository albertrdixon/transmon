package main

import (
	"time"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/gearbox/process"
	"github.com/albertrdixon/transmon/config"
	"github.com/albertrdixon/transmon/pia"
	"github.com/albertrdixon/transmon/transmission"
	"github.com/albertrdixon/transmon/vpn"
	"github.com/cenkalti/backoff"
	"golang.org/x/net/context"
)

func portCheck(t *process.Process, c *config.Config, ctx context.Context) error {
	if transmission.
		NewRawClient(c.Transmission.URL.String(), c.Transmission.User, c.Transmission.Pass).
		CheckPort() {
		return nil
	}

	logger.Infof("Transmission port not open, stopping transmission")
	t.Stop()

	ip, er := getIP(c.OpenVPN.Tun, c.Timeout.Duration, ctx)
	if er != nil {
		return er
	}
	logger.Infof("New bind ip: (%s) %s", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA.User, c.PIA.Pass, c.PIA.ClientID, c.Timeout.Duration, ctx)
	if er != nil {
		return er
	}
	logger.Infof("New peer port: %d", port)

	if er := transmission.UpdateSettings(c.Transmission.Config, ip, port); er != nil {
		return er
	}

	logger.Infof("Starting transmission")
	go t.ExecuteAndRestart(ctx)
	return nil
}

func portUpdate(c *config.Config, ctx context.Context) error {
	ip, er := getIP(c.OpenVPN.Tun, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return er
	}
	logger.Infof("New bind ip: (%s) %s", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA.User, c.PIA.Pass, c.PIA.ClientID, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return er
	}

	logger.Infof("New peer port: %d", port)
	notify := func(e error, w time.Duration) {
		logger.Debugf("Failed to update transmission port: %v", er)
	}
	operation := func() error {
		select {
		default:
			return transmission.
				NewRawClient(c.Transmission.URL.String(), c.Transmission.User, c.Transmission.Pass).
				UpdatePort(port)
		case <-ctx.Done():
			return nil
		}
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = c.Timeout.Duration
	return backoff.RetryNotify(operation, b, notify)
}

func restartProcesses(t, v *process.Process, c *config.Config, ctx context.Context) error {
	var (
		notify = func(e error, t time.Duration) {
			logger.Errorf("Failed to restart processes (retry in %v): %v", t, e)
		}
		operation = func() error {
			t.Stop()
			v.Stop()
			return startProcesses(t, v, c, ctx)
		}
		b = backoff.NewExponentialBackOff()
	)
	b.MaxElapsedTime = 30 * time.Minute
	b.MaxInterval = 10 * time.Second
	return backoff.RetryNotify(operation, b, notify)
}

func startProcesses(t, v *process.Process, c *config.Config, ctx context.Context) error {
	t.SetUser(uint32(c.Transmission.UID), uint32(c.Transmission.GID))

	logger.Infof("Starting openvpn")
	go v.ExecuteAndRestart(ctx)

	ip, er := getIP(c.OpenVPN.Tun, c.Timeout.Duration, ctx)
	if er != nil {
		return er
	}
	logger.Infof("New bind ip: (%s) %s", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA.User, c.PIA.Pass, c.PIA.ClientID, c.Timeout.Duration, ctx)
	if er != nil {
		return er
	}
	logger.Infof("New peer port: %d", port)

	if er := transmission.UpdateSettings(c.Transmission.Config, ip, port); er != nil {
		return er
	}

	logger.Infof("Starting transmission")
	go t.ExecuteAndRestart(ctx)

	return nil
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
