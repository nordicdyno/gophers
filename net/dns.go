package net

import (
	"math/rand"
	"net"
	"sync"
)

// DNS resolver with cache.
type DNSCache struct {
	SkipIPv4 bool
	SkipIPv6 bool
	m        sync.Mutex
	cache    map[string][]net.IP
}

var DefaultDNSCache = NewDNSCache()

func NewDNSCache() *DNSCache {
	return &DNSCache{SkipIPv6: true, cache: make(map[string][]net.IP)}
}

func (c *DNSCache) ResolveAll(host string) (ips []net.IP, err error) {
	c.m.Lock()
	defer c.m.Unlock()

	ips = c.cache[host]
	if ips != nil {
		return
	}

	aips, err := net.LookupIP(host)
	if err != nil {
		return
	}

	ips = make([]net.IP, 0, len(aips))
	for _, ip := range aips {
		if ip.To4() != nil {
			if c.SkipIPv4 {
				continue
			}
		} else {
			if c.SkipIPv6 {
				continue
			}
		}
		ips = append(ips, ip)
	}
	c.cache[host] = ips
	return
}

func (c *DNSCache) Resolve(host string) (ip net.IP, err error) {
	ips, err := c.ResolveAll(host)
	if err != nil {
		return
	}

	ip = ips[rand.Intn(len(ips))]
	return
}
