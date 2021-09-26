package main

import "time"

type Timestamp int64

func (ts Timestamp) Time() time.Time {
	s := int64(ts / 1000)
	ns := int64(ts-Timestamp(s)*1000) * int64(time.Millisecond)
	return time.Unix(s, ns)
}

func (ts Timestamp) String() string {
	return ts.Time().String()
}
