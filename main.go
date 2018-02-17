package main

import (
	"log"
)

func main() {
	proxy, err := NewGProxy("gproxy.conf")
	if err != nil {
		log.Fatal(err)
	}
	defer proxy.Shutdown()
	log.Fatal(proxy.Start())
}


