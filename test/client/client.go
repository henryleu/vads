package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"vads/hly"
)

var addr = flag.String("addr", "127.0.0.1:6000", "http service address")

//var addr = flag.String("addr", "114.116.110.22:6000", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		select {
		case <-interrupt:
			log.Println("process is interrupted")
			os.Exit(0)
		}
	}()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/websocket/hly/calling"}
	log.Printf("connecting to %s", u.String())
	fn := "../../data/8ef79f2695c811ea.wav"
	log.Printf("detecting %s", fn)
	hly.ClientRequest(u.String(), fn)

}
