package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestFindIPReal(t *testing.T) {
// 	is := assert.New(t)
// 	must := require.New(t)

// 	// tests := []struct{
// 	//   vpn *VPN
// 	//   success bool
// 	// }{
// 	//   {&VPN{Tun: "foo"}, false},
// 	//   {&VPN{Tun:}}
// 	// }

// 	infi, er := net.Interfaces()
// 	must.NoError(er)

// 	for _, inf := range infi {
// 		ip, er := findIP(&OpenVPN{Tun: inf.Name})
// 		is.NoError(er, "Interface %s: error %v", inf.Name, er)
// 		is.NotEmpty(ip, "Interface %s: ip %v", inf.Name, ip)
// 	}
// }

func TestFindIPFake(t *testing.T) {
	is := assert.New(t)

	ip, er := findIP(&OpenVPN{Tun: "foo"})
	is.Error(er)
	is.Empty(ip)
}
