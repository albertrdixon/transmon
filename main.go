package main

import (
	"os"
	"time"

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

func portUpdate(c *Config) {
	ip, er := getIP(c.OpenVPN, c.Timeout.Duration)
	if er != nil {
		logger.Errorf(er.Error())
		return
	}
	logger.Infof("IP for %v is %v", c.OpenVPN.Tun, ip)

	port, er := getPort(ip, c.PIA, c.Timeout.Duration)
	if er != nil {
		logger.Errorf(er.Error())
		return
	}
	logger.Infof("Port from PIA is %d", port)

	logger.Infof("Updating transmission port now")
	logger.Debugf("url=%s user=%s pass=%s", c.Transmission.URL.String(), c.Transmission.User, "*****")
	t := newTransmissionClient(c.Transmission.URL.String(), c.Transmission.User, c.Transmission.Pass)
	if er := t.updatePort(8888); er != nil {
		logger.Errorf(er.Error())
		return
	}
}

func getPort(ip string, pia *PIA, timeout time.Duration) (int, error) {
	var port int
	notify := func(e error, w time.Duration) {
		logger.Errorf("Failed to get port from %v (retry in %v): %v", pia.URL, w, e)
	}
	fn := func() error {
		p, er := requestPort(ip, pia)
		if er != nil {
			return er
		}
		port = p
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = timeout
	return port, backoff.RetryNotify(fn, b, notify)
}

func getIP(vpn *OpenVPN, timeout time.Duration) (string, error) {
	var address string
	notify := func(e error, w time.Duration) {
		logger.Errorf("Failed to get IP for %q (retry in %v): %v", vpn.Tun, w, e)
	}
	fn := func() (er error) {
		address, er = findIP(vpn)
		return
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = timeout
	return address, backoff.RetryNotify(fn, b, notify)
}

func main() {
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

	portUpdate(config)
	// c, stop := context.WithCancel(context.Background())

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	// select {
	// case <-sig:
	// 	logger.Infof("Received interrupt, shutting down.")
	// 	stop()
	// 	time.Sleep(100 * time.Millisecond)
	// 	os.Exit(0)
	// }
}
