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
	log.SetFlags(0)
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	log.Printf("GOMAXPROCS = %d", runtime.GOMAXPROCS(-1))
	rand.Seed(time.Now().UnixNano())

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

	start := time.Now()
	log.Printf("Making %d requests with concurrency %d ...", *nFlag, *cFlag)

	go func() {
		for i := 0; i < *nFlag; i++ {
			req, err := ghttp.NewRequest("GET", u.String(), nil)
			if err != nil {
				log.Fatalf("failed to create request for %s: %s", u, err)
			}
			requests <- req
		}
	}()

	lastLog := time.Now()
	for i := 0; i < *nFlag; i++ {
		res := <-responses
		if res.StatusCode != 200 {
			log.Fatalf("%s: %d", res.Request.URL, res.StatusCode)
		}
		if time.Now().Sub(lastLog) > 3*time.Second {
			log.Printf("Made %d requests ...", i)
			lastLog = time.Now()
		}
	}

	stop := time.Now()
	log.Printf("Made %d requests with concurrency %d in %f seconds.", *nFlag, *cFlag, stop.Sub(start).Seconds())
	log.SetFlags(0)

	var (
		open, used, closed int
		ctMin              time.Duration = time.Hour
		ctMax, ctSum       time.Duration
		rsMax, rsSum       float64
	)
	gnet.DefaultConnPool.Each(func(conn *gnet.Conn) {
		s := conn.Stats()

		if s.Established.IsZero() {
			log.Fatalf("Conn %d was not established", conn.Id)
		}

		if s.Recv != 0 && s.Sent != 0 {
			used++
		}
		if conn.Closed() {
			closed++
		} else {
			open++
		}

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

	log.Printf("Connections: %d used, %d open, %d closed", used, open, closed)
	log.Printf("Connect time: min %d msec, mean %d msec, max %d msec",
		int(ctMin/time.Millisecond), int(ctSum/time.Millisecond)/(open+closed), int(ctMax/time.Millisecond))
	log.Printf("Receive speed: mean %d bytes/sec, max %d bytes/sec", int(rsSum)/used, int(rsMax))
}
