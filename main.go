package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/gearbox/process"
	"github.com/albertrdixon/transmon/config"
	"github.com/albertrdixon/transmon/transmission"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logLevels = []string{"fatal", "error", "warn", "info", "debug"}
	app       = kingpin.New("transmon", "Keep your transmission ports clear!")

	conf  = app.Flag("config", "config file").Short('C').Default("/etc/transmon/config.yml").OverrideDefaultFromEnvar("CONFIG").ExistingFile()
	level = app.Flag("log-level", "log level. One of: fatal, error, warn, info, debug").Short('l').Default("info").OverrideDefaultFromEnvar("LOG_LEVEL").Enum(logger.Levels...)
)

const (
	portInterval    = 1 * time.Hour
	restartInterval = 24 * time.Hour
	checkInterval   = 5 * time.Minute
	cleanInterval   = 30 * time.Minute
)

func workers(conf *config.Config, c context.Context, quit context.CancelFunc) {
	var (
		port    = time.NewTicker(portInterval)
		restart = time.NewTicker(restartInterval)
		check   = time.NewTicker(checkInterval)
	)

	logger.Infof("Port update will run once every hour")
	logger.Infof("VPN restart will run once every day")

	trans, er := process.New("transmission", conf.Transmission.Command, os.Stdout)
	if er != nil {
		quit()
		logger.Fatalf(er.Error())
	}
	vpn, er := process.New("openvpn", conf.OpenVPN.Command, os.Stdout)
	if er != nil {
		quit()
		logger.Fatalf(er.Error())
	}

	if er := startProcesses(trans, vpn, conf, c); er != nil {
		quit()
		logger.Fatalf(er.Error())
	}
	portUpdate(conf, c)

	for {
		select {
		case <-c.Done():
			port.Stop()
			restart.Stop()
			trans.Stop()
			vpn.Stop()
			return
		case t := <-check.C:
			logger.Debugf("Checking transmission port at %v", t)
			if er := portCheck(trans, conf, c); er != nil {
				if er := restartProcesses(trans, vpn, conf, c); er != nil {
					port.Stop()
					restart.Stop()
					trans.Stop()
					vpn.Stop()
					logger.Fatalf(er.Error())
				}
			}
		case t := <-port.C:
			logger.Infof("Update of Transmission port at %v", t)
			if er := portUpdate(conf, c); er != nil {
				if er := restartProcesses(trans, vpn, conf, c); er != nil {
					port.Stop()
					restart.Stop()
					trans.Stop()
					vpn.Stop()
					logger.Fatalf(er.Error())
				}
			}
		case t := <-restart.C:
			logger.Infof("Restarting Transmission and OpenVPN at %v", t)
			if er := restartProcesses(trans, vpn, conf, c); er != nil {
				port.Stop()
				restart.Stop()
				trans.Stop()
				vpn.Stop()
				logger.Fatalf(er.Error())
			}
		}
	}
}

func cleaner(conf *config.Config, c context.Context) {
	var (
		d     = conf.Cleaner.Interval.Duration
		clean = time.NewTicker(d)
	)
	logger.Infof("Torrent cleaner will run once every %v", d)
	for {
		select {
		case <-c.Done():
			clean.Stop()
			return
		case t := <-clean.C:
			logger.Infof("Half hourly torrent cleaning at %v", t)
			er := transmission.
				NewClient(conf.Transmission.URL.String(), conf.Transmission.User, conf.Transmission.Pass).
				CleanTorrents()
			if er != nil {
				logger.Errorf(er.Error())
			}
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	kingpin.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	logger.Configure(*level, "[transmon] ", os.Stdout)
	logger.Infof("Starting transmon version %v", version)

	c, stop := context.WithCancel(context.Background())
	conf, er := config.ReadAndWatch(*conf, c)
	if er != nil {
		logger.Fatalf("Failed to read config: %v", er)
	}

	go workers(conf, c, stop)

	if conf.Cleaner.Enabled {
		go cleaner(conf, c)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-sig:
		logger.Infof("Received interrupt, shutting down...")
		close(sig)
		stop()
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}
}
