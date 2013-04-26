package net

import (
	"math/rand"
	"net"
	"sync"
)

// DNS resolver with cache.
type DNSCache struct {
	m     sync.Mutex
	cache map[string][]string // hostname to IPs
}

var DefaultDNSCache = NewDNSCache()

func NewDNSCache() *DNSCache {
	return &DNSCache{cache: make(map[string][]string)}
}

func (c *DNSCache) ResolveAll(host string) (ips []string, err error) {
	c.m.Lock()
	defer c.m.Unlock()

	ips = c.cache[host]
	if ips != nil {
		return
	}

	ips, err = net.LookupHost(host)
	if err != nil {
		return
	}
	c.cache[host] = ips
	return
}

func (c *DNSCache) Resolve(host string) (ip string, err error) {
	ips, err := c.ResolveAll(host)
	if err != nil {
		return
	}

	ip = ips[rand.Intn(len(ips))]
	return
}
