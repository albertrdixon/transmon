package config

import (
	"bytes"
	"time"
)

func (d *duration) UnmarshalJSON(p []byte) error {
	val := bytes.Trim(p, `"`)
	t, er := time.ParseDuration(string(val))
	if er != nil {
		return er
	}
	d.Duration = t
	return nil
}
