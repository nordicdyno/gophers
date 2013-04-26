package net

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

// Connection with statistics.
type Conn struct {
	net.Conn
	Id     int32
	closed int32
	s      Stats
}

// check interfaces
var (
	_ net.Conn     = &Conn{}
	_ fmt.Stringer = &Conn{}
)

func Dial(net, addr string) (net.Conn, error) {
	return DefaultConnPool.Dial(net, addr)
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	now := time.Now()
	n64 := uint64(n)
	res := atomic.AddUint64(&c.s.Recv, n64)
	if res > 0 && res == n64 {
		c.s.FirstRecv = now
	}
	if n64 > 0 {
		c.s.LastRecv = now
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	now := time.Now()
	n64 := uint64(n)
	res := atomic.AddUint64(&c.s.Sent, n64)
	if res > 0 && res == n64 {
		c.s.FirstSend = now
	}
	if n64 > 0 {
		c.s.LastSend = now
	}
	return
}

func (c *Conn) Close() error {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.s.Closed = time.Now()
	}
	return c.Conn.Close()
}

func (c *Conn) String() string {
	o := "open"
	if atomic.LoadInt32(&c.closed) == 1 {
		o = "closed"
	}
	s := c.Stats()
	return fmt.Sprintf("Conn #%d (%s): sent %d, recv %d", c.Id, o, s.Sent, s.Recv)
}

func (c *Conn) Stats() *Stats {
	s := Stats{
		Recv: atomic.LoadUint64(&c.s.Recv), Sent: atomic.LoadUint64(&c.s.Sent),
		FirstRecv: c.s.FirstRecv, FirstSend: c.s.FirstSend,
		LastRecv: c.s.LastRecv, LastSend: c.s.LastSend,
		Created: c.s.Created, Established: c.s.Established, Closed: c.s.Closed,
	}
	return &s
}

func (c *Conn) ResetStats() {
	atomic.StoreUint64(&c.s.Recv, 0)
	atomic.StoreUint64(&c.s.Sent, 0)
	c.s.FirstRecv = time.Time{}
	c.s.FirstSend = time.Time{}
	c.s.LastRecv = time.Time{}
	c.s.LastSend = time.Time{}
	return
}
