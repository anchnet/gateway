package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anchnet/gateway/g"
	"github.com/anchnet/gateway/http"
	"github.com/anchnet/gateway/receiver"
	"github.com/anchnet/gateway/sender"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	// global config
	g.ParseConfig(*cfg)

	sender.Start()
	receiver.Start()

	// http
	http.Start()

	select {}
}
