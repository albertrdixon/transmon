package vpn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindIPFake(t *testing.T) {
	is := assert.New(t)

	ip, er := FindIP("foo")
	is.Error(er)
	is.Empty(ip)
}
