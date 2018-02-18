package main

import (
	"log"
	"flag"
)

var version = "MUST BE REPLACED WITH ldflags"

func main() {
	cfgPath := flag.String("c", "gproxy.conf", "config file path")
	flag.Parse()
	proxy, err := NewGProxy(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer proxy.Shutdown()
	log.Fatal(proxy.Start())
}
