package googletokenvalidator

import (
	"strconv"
	"strings"
	"time"
)

type JsonTimestamp time.Time
type JsonBool bool

func (jt *JsonTimestamp) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	f, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return err
	}

	ts := time.Unix(0, int64(f*float64(time.Second/time.Nanosecond)))
	*jt = JsonTimestamp(ts)

	return nil
}

func (jb *JsonBool) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	value, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	*jb = JsonBool(value)

	return nil
}
