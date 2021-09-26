package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listen   = flag.String("listen", ":8083", "")
	keyFile  = flag.String("key", "", "path to coincap.io API key")
	interval = flag.Duration("interval", time.Minute, "")

	apiKey string
)

func init() {
	flag.Parse()

	if *keyFile != "" {
		data, err := ioutil.ReadFile(*keyFile)
		if err != nil {
			panic(err)
		}
		apiKey = string(data)
		log.Printf("Loaded API key from %v", *keyFile)
	} else {
		log.Println("No API key file provided")
	}
}

type workFunc func(*http.Client) (reset time.Time, err error)

func loop(work workFunc) {
	client := new(http.Client)
	timer := time.NewTimer(0)
	for range timer.C {
		reset, err := work(client)
		dt := *interval
		if err != nil && reset != (time.Time{}) {
			dt = -time.Since(reset) + time.Second
			if dt < 0 {
				dt = 0
			}
		}
		log.Printf("Next tick in %v", dt)
		timer.Reset(dt)
	}
}

func main() {
	go loop(func(client *http.Client) (time.Time, error) {
		log.Println("Fetching assets")
		assets, reset, err := Assets(client)
		if err != nil {
			log.Println("Failed to fetch assets:", err)
		}
		log.Printf("Updating %d assets", len(assets))
		updateAssets(assets)
		return reset, err
	})

	go loop(func(client *http.Client) (time.Time, error) {
		log.Println("Fetching markets")
		markets, reset, err := Markets(client)
		if err != nil {
			log.Println("Failed to fetch markets:", err)
		}
		log.Printf("Updating %d markets", len(markets))
		updateMarkets(markets)
		return reset, err
	})

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listen, nil))
}
