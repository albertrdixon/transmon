package transmission

import (
	"strconv"

	"github.com/Pallinder/go-randomdata"
)

func even(n int) bool {
	return (n % 2) == 0
}

func tag() int {
	n := randomdata.StringNumberExt(3, "", 3)
	rtn, _ := strconv.Atoi(n)
	return rtn
}
