package http

import (
	"errors"
	g "github.com/AlekSi/gophers"
	gnet "github.com/AlekSi/gophers/net"
	"io"
	"net/http"
)

type Client struct {
	*http.Client
}

func NewClient(p *gnet.ConnPool) *Client {
	redirector := func(req *http.Request, via []*http.Request) error {
		return errors.New("gophers: not following redirect")
	}
	return &Client{&http.Client{Transport: NewTransport(p), CheckRedirect: redirector}}
}

func (c *Client) GetDiscardBody(url string) (res *http.Response, bodySize int64, err error) {
	req, err := NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	return c.DoDiscardBody(req)
}

func (c *Client) DoDiscardBody(req *http.Request) (res *http.Response, bodySize int64, err error) {
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
