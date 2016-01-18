package config

import (
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/albertrdixon/gearbox/logger"
	pi "github.com/albertrdixon/transmon/pia"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/pborman/uuid"
)

var (
	conf = &Config{
		PIA:          &PIA{URL: pi.GetPortForwardEndpoint(), ClientID: uuid.New()},
		Transmission: &Transmission{UID: 0, GID: 0},
		OpenVPN:      &OpenVPN{Tun: defaultDevice},
		Timeout:      &duration{Duration: defaultDuration},
	}
)

func Read(file string) (*Config, error) {
	info, er := os.Stat(file)
	if er != nil {
		return conf, er
	}

	return read(file, info)
}

func ReadAndWatch(file string, ctx context.Context) (*Config, error) {
	c, er := Read(file)
	if er != nil {
		return c, er
	}

	go func() {
		logger.Debugf("Watching for config changes: config=%q", file)
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if er, ok := c.update(); er == nil && ok {
					logger.Infof("Config updated")
				}
			}
		}
	}()

	return c, nil
}

func (c *Config) update() (error, bool) {
	info, er := os.Stat(c.file)
	if er != nil {
		return er, false
	}

	if info.ModTime().Equal(c.modTime) {
		return nil, false
	}

	nc, er := read(c.file, info)
	if er == nil {
		c = nc
		return nil, true
	}
	return er, false
}

func read(file string, info os.FileInfo) (*Config, error) {
	logger.Debugf("Reading config from %q", file)
	content, er := ioutil.ReadFile(file)
	if er != nil {
		return nil, er
	}

	c := new(Config)
	if er := yaml.Unmarshal(content, c); er != nil {
		return conf, er
	}

	c.file = file
	c.modTime = info.ModTime()
	return c, mergo.Merge(c, conf)
}

const (
	defaultDuration = 5 * time.Minute
	defaultDevice   = "tun0"
)
