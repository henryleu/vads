package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/henryleu/vads/hly"
)

var addr = flag.String("addr", "localhost:6000", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for {
			select {
			case <-interrupt:
				log.Println("server is interrupted")
				os.Exit(0)
			}
		}
	}()

	http.HandleFunc("/mrcp", hly.HandleMRCP)
	http.HandleFunc("/", hly.Home)
	log.Printf("server is listening on %v\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
