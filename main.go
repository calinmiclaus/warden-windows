package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/garden/server"

	Backend "github.com/cloudfoundry-incubator/warden-windows/backend"
)

var listenNetwork = flag.String(
	"listenNetwork",
	"tcp",
	"how to listen on the address (unix, tcp, etc.)",
)

var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:9876",
	"address to listen on",
)

var containerGraceTime = flag.Duration(
	"containerGraceTime",
	0,
	"time after which to destroy idle containers",
)

var containerRootPath = flag.String(
	"containerRootPath",
	"",
	"directory in which to store container root directories",
)

var debug = flag.Bool(
	"debug",
	false,
	"show low-level command output",
)

func main() {
	flag.Parse()

	maxProcs := runtime.NumCPU()
	prevMaxProcs := runtime.GOMAXPROCS(maxProcs)

	log.Println("set GOMAXPROCS to", maxProcs, "was", prevMaxProcs)

	backend := Backend.New(*containerRootPath, nil)

	log.Println("setting up backend")

	log.Println("starting server; listening with", *listenNetwork, "on", *listenAddr)

	graceTime := *containerGraceTime

	logger := cf_lager.New("warden-winodws")
	serverLogger := logger.Session("garden")

	wardenServer := server.New(*listenNetwork, *listenAddr, graceTime, backend, serverLogger)

	err := wardenServer.Start()
	if err != nil {
		log.Fatalln("failed to start:", err)
	}

	signals := make(chan os.Signal, 1)

	go func() {
		<-signals
		log.Println("stopping...")
		wardenServer.Stop()
		os.Exit(0)
	}()

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {}
}
