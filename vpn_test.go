package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindIPFake(t *testing.T) {
	is := assert.New(t)

	ip, er := findIP(&OpenVPN{Tun: "foo"})
	is.Error(er)
	is.Empty(ip)
}
