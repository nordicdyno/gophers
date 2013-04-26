package http

import (
	gnet "github.com/AlekSi/gophers/net"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewTransport(p *gnet.ConnPool) *http.Transport {
	return &http.Transport{Dial: p.Dial, MaxIdleConnsPerHost: 1 << 30, ResponseHeaderTimeout: time.Hour}
}

func NewRequest(method, urlStr string, body io.Reader) (req *http.Request, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return
	}

	host := u.Host
	port := "80"
	if u.Scheme == "https" {
		port = "443"
	}
	if strings.Contains(host, ":") {
		host, port, err = net.SplitHostPort(host)
		if err != nil {
			return
		}
	}

	ip, err := gnet.DefaultDNSCache.Resolve(host)
	if err != nil {
		return
	}

	u.Host = net.JoinHostPort(ip, port)
	req, err = http.NewRequest(method, u.String(), body)
	if err != nil {
		return
	}
	req.Host = net.JoinHostPort(host, port)

	return
}
