package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/zzvanq/artwall-dl/internal/dl"
)

const configFile = ".conf.xml"

var destination = flag.String("d", "", "Destination directory")
var downloaderBits = flag.Int("s", 0, "Sources bitmap")
var amount = flag.Int("n", 1, "Amount to download per source")

func main() {
	flag.Parse()
	validateFlags()

	conf := dl.ParseConfig(configFile)

	downloaders := dl.GetDownloaders(conf.Sources, *downloaderBits, *destination)
	if len(downloaders) == 0 {
		log.Fatalln("no downloaders found")
	}

	var downloaded atomic.Int64
	var wg sync.WaitGroup
	wg.Add(len(downloaders))
	for _, d := range downloaders {
		go func() {
			defer wg.Done()
			downloaded.Add(int64(d.Download(*amount)))
		}()
	}
	wg.Wait()
	fmt.Printf("downloaded %d images from %d sources\n", downloaded.Load(), len(downloaders))
}

func validateFlags() {
	if *destination == "" {
		log.Fatalln("-d must be set")
	}

	if _, err := os.Stat(*destination); err != nil {
		log.Fatalln(err)
	}

	if *downloaderBits <= 0 {
		log.Fatalln("-s value is not valid")
	}

	if *amount <= 0 {
		log.Fatalln("-n must be greater than 0")
	}
}
