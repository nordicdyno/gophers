package net

import (
	"time"
)

// TODO replace with pointers, set atomically

// Connection statistics.
type Stats struct {
	Recv, Sent                   uint64    // bytes received and sent
	FirstRecv, FirstSend         time.Time // time of first non-empty receive and send operation
	LastRecv, LastSend           time.Time // time of last non-empty receive and send operation
	Created, Established, Closed time.Time // time of object creation, connection establishment and closing
}

func (s *Stats) SendSpeed() float64 {
	if s.Sent == 0 {
		return 0
	}
	return float64(s.Sent) / (float64(s.LastSend.Sub(s.FirstSend)) / float64(time.Second))
}

func (s *Stats) RecvSpeed() float64 {
	if s.Recv == 0 {
		return 0
	}
	return float64(s.Recv) / (float64(s.LastRecv.Sub(s.FirstRecv)) / float64(time.Second))
}
