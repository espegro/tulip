package main

import (
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// BloomFilter
type BloomFilter struct {
	Name     string `json:"name"`     // Filter name
	Bitset   []uint `json:"bitset"`   // Bloom filter buckets
	Rounds   uint   `json:"rounds"`   // Number of hash values
	Elements []uint `json:"elements"` // Number of elements in the filter
	Size     uint   `json:"size"`     // Size of the bloom filter
	Decay    uint   `json:"decay"`    // Decay speed
	Max      uint   `json:"max"`      // Max bucket value
}

// Global config struct
type Config struct {
	FilterStateName string
	Port            string
	ListenAddress   string
}

// Global config
var config = new(Config)

// Global map of filters
var filters = make(map[string]BloomFilter)

// Global mutex for map
var maplock = &sync.Mutex{}

func main() {
	// Set up logging to syslog
	logwriter, err := syslog.New(syslog.LOG_NOTICE, "tulip")
	if err == nil {
		log.SetOutput(logwriter)
	}

	// Config
	config_state := flag.String("state", "tulip-filterstate.json", "statefile-name")
	config_port := flag.String("port", "8080", "listen port")
	config_addr := flag.String("address", "127.0.0.1", "listen address")
	flag.Parse()
	if len(*config_state) > 0 {
		config.FilterStateName = *config_state
	}
	if len(*config_port) > 0 {
		config.Port = *config_port
	}
	if len(*config_addr) > 0 {
		config.ListenAddress = *config_addr
	}
	log.Printf("Starting tulip listen:%v:%v - state: %v\n", *config_addr, *config_port, *config_state)
	fmt.Printf("Starting tulip listen:%v:%v - state: %v\n", *config_addr, *config_port, *config_state)

	// Check write permissions on filterstate
	touchfile(config.FilterStateName)
	file, err := os.OpenFile(config.FilterStateName, os.O_WRONLY, 0666)
	if err != nil {
		if os.IsPermission(err) {
			log.Println("Unable to write to ", config.FilterStateName)
			log.Println(err)
			fmt.Println("Unable to write to ", config.FilterStateName)
			fmt.Println(err)
			os.Exit(1)
		}
	}
	file.Close()

	// Load state
	loadfilterstate(config.FilterStateName)

	// Catch signals - save state
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Signal caught, saving state to file: %v\n", config.FilterStateName)
		savefilterstate(config.FilterStateName)
		os.Exit(1)
	}()

	// Set up handlers
	mux := httprouter.New()
	mux.GET("/hello/:name/:action", hello)
	mux.POST("/bloom/new/:name/:size/:hash/:decay/:max", newhandler)
	mux.POST("/bloom/add/:name/:value", addhandler)
	mux.POST("/bloom/addifnotset/:name/:value", addifnotsethandler)
	mux.POST("/bloom/poster/:name", posterhandler)
	mux.GET("/bloom/test/:name/:value", testhandler)
	mux.POST("/bloom/reset/:name", resethandler)
	mux.POST("/bloom/destroy/:name", destroyhandler)
	mux.GET("/bloom/list", listhandler)
	mux.GET("/bloom/debug/:name", debughandler)
	mux.GET("/bloom/info/:name", infohandler)
	mux.POST("/bloom/save", savehandler)
	mux.POST("/bloom/load", loadhandler)

	// Start server process
	server := http.Server{
		Addr:    config.ListenAddress + ":" + config.Port,
		Handler: mux,
	}
	server.ListenAndServe()
}
