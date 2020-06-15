package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/henryleu/vads/test/app"
)

var addr = flag.String("addr", "localhost:6000", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		select {
		case <-interrupt:
			log.Println("server is interrupted")
			os.Exit(0)
		}
	}()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/mrcp"}
	log.Printf("connecting to %s", u.String())
	// fn := "../../data/8ef79f2695c811ea.wav"
	fn := "../../data/0ebb1c6895c611ea.wav"
	log.Printf("detecting %s", fn)
	app.ClientRequest(u.String(), fn)
}
