package net

import (
	"errors"
	g "github.com/AlekSi/gophers"
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

func NewClient(p *gnet.ConnPool) *http.Client {
	redirector := func(req *http.Request, via []*http.Request) error {
		return errors.New("gophers: not following redirect")
	}
	return &http.Client{Transport: NewTransport(p), CheckRedirect: redirector}
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

func GetDiscardBody(c *http.Client, url string) (res *http.Response, bodySize int64, err error) {
	req, err := NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	return DoDiscardBody(c, req)
}

func DoDiscardBody(c *http.Client, req *http.Request) (res *http.Response, bodySize int64, err error) {
	res, err = c.Do(req)
	if err != nil {
		return
	}

	if res.Body != nil {
		defer res.Body.Close()

		if res.ContentLength != 0 {
			size := res.ContentLength
			if size < 0 || size > g.MB {
				size = g.MB
			}
			buf := make([]byte, size)

			var n int
			for err == nil {
				n, err = res.Body.Read(buf)
				bodySize += int64(n)
			}

			if err == io.EOF {
				if res.ContentLength > 0 && res.ContentLength == bodySize {
					err = nil
				}
			}
		}
	}
	return
}
