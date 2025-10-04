package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zzvanq/artwall-dl/internal/dl"
	"github.com/zzvanq/artwall-dl/internal/dl/nga"
)

var destination = flag.String("d", "", "Destination directory")
var listUrl = flag.String("l", "", "Url to the list of images")
var downloader = flag.String("s", nga.DlKey, "Source")
var amount = flag.Int("n", 1, "Amount to download")

func main() {
	flag.Parse()
	validateFlags()

	dl, err := dl.NewDownloader(*listUrl, *downloader, *destination)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("downloaded %d images\n", dl.Download(*amount))
}

func validateFlags() {
	if *destination == "" {
		log.Fatalln("-d must be set")
	}

	if _, err := os.Stat(*destination); err != nil {
		log.Fatalln(err)
	}

	if *listUrl == "" {
		log.Fatalln("-l must be set")
	}

	if *amount <= 0 {
		log.Fatalln("-n must be greater than 0")
	}
}
