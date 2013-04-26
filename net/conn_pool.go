package net

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ConnPool struct {
	m    sync.Mutex
	pool []*Conn
}

var DefaultConnPool = NewConnPool()

var lastConnId int32 = -1

func NewConnPool() *ConnPool {
	return &ConnPool{pool: make([]*Conn, 0, 100)}
}

func (p *ConnPool) Dial(netz string, addr string) (conn net.Conn, err error) {
	p.m.Lock()
	defer p.m.Unlock()

	id := atomic.AddInt32(&lastConnId, 1)
	created := time.Now()
	conn, err = net.Dial(netz, addr)
	established := time.Now()
	if err != nil {
		return
	}

	c := &Conn{Conn: conn, Id: id, s: Stats{Created: created, Established: established}}
	p.pool = append(p.pool, c)
	conn = c
	return
}

func (p *ConnPool) Len() int {
	p.m.Lock()
	defer p.m.Unlock()

	return len(p.pool)
}

func (p *ConnPool) Get(id int32) *Conn {
	p.m.Lock()
	defer p.m.Unlock()

	return p.pool[id]
}

func (p *ConnPool) Each(f func(*Conn)) {
	p.m.Lock()
	defer p.m.Unlock()

	for _, c := range p.pool {
		f(c)
	}
}
