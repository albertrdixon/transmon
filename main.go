package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/cenkalti/backoff"
	"github.com/pborman/uuid"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logLevels = []string{"fatal", "error", "warn", "info", "debug"}
	app       = kingpin.New("transmon", "Keep your transmission ports clear!")

	conf  = app.Flag("config", "config file").Short('C').Default("/etc/transmon/config.yml").OverrideDefaultFromEnvar("CONFIG").ExistingFile()
	level = app.Flag("log-level", "log level. One of: fatal, error, warn, info, debug").Short('l').Default("info").OverrideDefaultFromEnvar("LOG_LEVEL").Enum(logger.Levels...)
)

func portUpdate(c *Config, ctx context.Context) error {
	ip, er := getIP(c.OpenVPN, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return er
	}
	logger.Infof("%v: inet %v", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA, c.Timeout.Duration, ctx)
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
			t := newTransmissionClient(c.Transmission.URL.String(), c.Transmission.User, c.Transmission.Pass)
			return t.updatePort(port)
		case <-ctx.Done():
			return nil
		}
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = c.Timeout.Duration
	return backoff.RetryNotify(operation, b, notify)
}

func getPort(ip string, pia *PIA, timeout time.Duration, c context.Context) (int, error) {
	var port int
	notify := func(e error, w time.Duration) {
		logger.Errorf("Failed to get port from %v (retry in %v): %v", pia.URL, w, e)
	}
	fn := func() error {
		select {
		default:
			p, er := requestPort(ip, pia)
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

func getIP(vpn *OpenVPN, timeout time.Duration, c context.Context) (string, error) {
	var address string
	notify := func(e error, w time.Duration) {
		logger.Errorf("Failed to get IP for %q (retry in %v): %v", vpn.Tun, w, e)
	}
	fn := func() (er error) {
		select {
		default:
			address, er = findIP(vpn)
			return
		case <-c.Done():
			return
		}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = timeout
	return address, backoff.RetryNotify(fn, b, notify)
}

func runTransAndVPN(c *Config, ctx context.Context) (*command, *command, error) {
	t, er := newCommand("transmission", c.Transmission.Command)
	if er != nil {
		return nil, nil, er
	}
	v, er := newCommand("openvpn", c.OpenVPN.Command)
	if er != nil {
		return nil, nil, er
	}
	t.SetUser(c.Transmission.UID, c.Transmission.GID)

	logger.Infof(`Starting openvpn: %s`, c.OpenVPN.Command)
	if er := v.Execute(ctx); er != nil {
		return nil, nil, er
	}

	ip, er := getIP(c.OpenVPN, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return nil, nil, er
	}
	logger.Infof("%v: inet %v", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA, c.Timeout.Duration, ctx)
	if er != nil || ctx.Err() != nil {
		return nil, nil, er
	}
	logger.Infof("New transmission port: %d", port)

	if er := updateTransmissionConfig(c.Transmission.Config, ip, port); er != nil {
		return nil, nil, er
	}

	logger.Infof("Starting transmission: %s", c.Transmission.Command)
	if er := t.Execute(ctx); er != nil {
		return nil, nil, er
	}

	return t, v, nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	kingpin.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	logger.Configure(*level, "[transmon] ", os.Stdout)
	logger.Infof("Starting transmon version %v", version)

	config, er := readConfig(*conf)
	if er != nil {
		logger.Fatalf("Failed to read config: %v", er)
	}
	if config.PIA.ClientID == "" {
		config.PIA.ClientID = uuid.New()
	}

	logger.Infof("Port update will run once every hour")
	port := time.NewTicker(1 * time.Hour)
	logger.Infof("VPN restart will run once every day")
	restart := time.NewTicker(24 * time.Hour)
	c, stop := context.WithCancel(context.Background())

	go func(q context.CancelFunc) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-sig:
			logger.Infof("Received interrupt, shutting down...")
			q()
		}
	}(stop)

	trans, vpn, er := runTransAndVPN(config, c)
	if er != nil {
		logger.Fatalf(er.Error())
	}
	portUpdate(config, c)

	go func() {
		for {
			select {
			case <-c.Done():
				trans.Stop()
				vpn.Stop()
				return
			case t := <-port.C:
				logger.Infof("Hourly update of transmission port at %v", t)
				if er := portUpdate(config, c); er != nil {
					trans.Stop()
					vpn.Stop()
					trans, vpn, er = runTransAndVPN(config, c)
					if er != nil {
						logger.Errorf(er.Error())
					}
				}
			case t := <-restart.C:
				logger.Infof("Daily restart of vpn and transmission at %v", t)
				trans.Stop()
				vpn.Stop()
				trans, vpn, er = runTransAndVPN(config, c)
				if er != nil {
					logger.Errorf(er.Error())
				}
			}
		}
	}()

	select {
	case <-c.Done():
		port.Stop()
		restart.Stop()
		trans.Stop()
		vpn.Stop()
		time.Sleep(50 * time.Millisecond)
		os.Exit(0)
	}
}
