package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"

	shortdescription "github.com/Inuart/wikimedia-exercise"
)

type Config struct {
	Addr        string        `envconfig:"ADDR"`
	ContactInfo string        `envconfig:"CONTACT_INFO" required:"true"`
	CacheSize   int           `envconfig:"CACHE_SIZE"`        // Max amount of results the cache should hold
	CachedTTL   time.Duration `envconfig:"CACHED_RESULT_TTL"` // Time To Live for each cached result
}

func main() {
	var conf Config

	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatal("error parsing env vars:", err)
	}

	descriptor, err := shortdescription.New(shortdescription.Config{
		ContactInfo: conf.ContactInfo,
		CacheSize:   conf.CacheSize,
		CachedTTL:   conf.CachedTTL,
	})
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		log.Fatalf("Unable to listen to the provided address %q: %v", conf.Addr, err)
	}

	log.Println("The shortdescription server will listen at", listener.Addr().String())

	err = http.Serve(listener, descriptor)
	if err != nil {
		log.Fatal(err)
	}
}
