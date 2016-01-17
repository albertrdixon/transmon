package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	is := assert.New(t)

	c, er := readConfig("config.yml.example")
	is.NoError(er)
	is.Equal("username", c.Transmission.User)
	is.Equal("tun3", c.OpenVPN.Tun)
	is.EqualValues(10*time.Minute, c.Timeout.Duration)
}
