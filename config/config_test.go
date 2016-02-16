package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	var (
		is = assert.New(t)
	)

	c, er := Read("examples/config.yml")
	is.NoError(er)
	is.Equal("username", c.Transmission.User)
	is.Equal("tun3", c.OpenVPN.Tun)
	is.EqualValues(10*time.Minute, c.Timeout.Duration)
	is.True(c.Cleaner.Enabled)
	is.Equal(3*time.Hour, c.Cleaner.Interval.Duration)
}
