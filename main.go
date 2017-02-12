package main

import (
	"flag"
	"io"
	"log"
	"net"
	"sync"
)

var (
	liface  string
	proxy1  string
	proxy2  string
	pattern string
	route   chan string
	wg      sync.WaitGroup
)

func Listen() error {
	lis, err := net.Listen("tcp", liface)
	if err != nil {
		return err
	}
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("Accept error:", err)
		}

		c := <-route
		switch c {
		case "1":
			log.Printf("Proxy 1: %s -> %s", conn.RemoteAddr(), proxy1)
			go forward(conn, proxy1)
		case "2":
			log.Printf("Proxy 2: %s -> %s", conn.RemoteAddr(), proxy2)
			go forward(conn, proxy2)
		}

	}
}

func forward(in net.Conn, iproxy string) {
	out, err := net.Dial("tcp", iproxy)
	defer out.Close()
	defer in.Close()
	if err != nil {
		return
	}

	wg.Add(1)
	go func() {
		io.Copy(out, in)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		io.Copy(in, out)
		wg.Done()
	}()

	wg.Wait()

}
func router() {
	for {
		for _, r := range pattern {
			c := string(r)
			route <- c
		}
	}
}

func main() {
	flag.StringVar(&liface, "l", "", "listen interface format ip:port")
	flag.StringVar(&proxy1, "1", "", "proxy #1 format host:port")
	flag.StringVar(&proxy2, "2", "", "proxy #2 format host:port")
	flag.StringVar(&pattern, "p", "12", "proxy switch pattern. e.g. 12 to alternate proxies or 1121112 as more complex switch pattern")
	flag.Parse()
	route = make(chan string)
	go router()
	Listen()
}
