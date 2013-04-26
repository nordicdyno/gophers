package main

import (
	"flag"
	gnet "github.com/AlekSi/gophers/net"
	ghttp "github.com/AlekSi/gophers/net/http"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"
)

var (
	nFlag = flag.Int("n", 1, "Number of requests to perform")
	cFlag = flag.Int("c", 1, "Number of multiple requests to make")
)

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(0)

	flag.Usage = func() {
		log.Printf("Usage: gophers [flags] http[s]://hostname[:port]/path")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	u, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to parse URL %q: %s", flag.Arg(0), err)
	}

	req, err := ghttp.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalf("failed to create request for %s: %s", u, err)
	}

	ips, _ := gnet.DefaultDNSCache.ResolveAll(u.Host)
	log.Printf("%s resolves to %v", u.Host, ips)

	log.SetFlags(log.Lmicroseconds)

	requests := make(chan *http.Request, *cFlag)
	responses := make(chan *http.Response, *cFlag)

	for i := 0; i < *cFlag; i++ {
		go func() {
			for req := range requests {
				res, bodySize, err := ghttp.DefaultClient.DoDiscardBody(req)
				if err != nil {
					log.Fatalf("%s: %s", req.URL, err)
				}
				res.ContentLength = bodySize
				responses <- res
			}
		}()
	}

	go func() {
		for i := 0; i < *nFlag; i++ {
			requests <- req
		}
	}()

	for i := 0; i < *nFlag; i++ {
		res := <-responses
		if res.StatusCode != 200 {
			log.Fatalf("%s: %d", res.Request.URL, res.StatusCode)
		}
	}

	log.SetFlags(0)

	var (
		open, closed int
		ctMin        time.Duration = time.Hour
		ctMax, ctSum time.Duration
		rsMax, rsSum float64
	)
	gnet.DefaultConnPool.Each(func(conn *gnet.Conn) {
		if conn.Closed() {
			closed++
		} else {
			open++
		}

		s := conn.Stats()

		d := s.Established.Sub(s.Created)
		if d < ctMin {
			ctMin = d
		}
		if d > ctMax {
			ctMax = d
		}
		ctSum += d

		rs := s.RecvSpeed()
		if rs > rsMax {
			rsMax = rs
		}
		rsSum += rs
	})

	conns := open + closed
	log.Printf("Connections: %d open, %d closed", open, closed)
	log.Printf("Connect time: min %d msec, mean %d msec, max %d msec",
		int(ctMin/time.Millisecond), int(ctSum/time.Millisecond)/conns, int(ctMax/time.Millisecond))
	log.Printf("Receive speed: mean %d bytes/sec, max %d bytes/sec", int(rsSum)/conns, int(rsMax))
}
