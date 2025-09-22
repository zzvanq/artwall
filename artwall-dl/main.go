package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
)

type downloader interface {
	download(int, string) int
}

// keys are single bits. 1 = nga, 2 = wallhaven, 4 = ...
var downloadersMap = map[int]downloader{
	1: ngaDl{},
	// 2: wallhavenDl{},
}

var destination = flag.String("d", "", "Destination directory")
var downloaderBits = flag.Int("s", 1, "Sources bitmap")
var amount = flag.Int("n", 1, "Amount to download per source")

func main() {
	validateFlags()

	conf := parseConfig("conf.xml")

	downloaders := getDownloaders(conf.Sources, *downloaderBits)
	if len(downloaders) == 0 {
		log.Fatalln("no downloaders found")
	}

	// TODO: parallelize
	downloaded := 0
	for _, downloader := range downloaders {
		downloaded += downloader.download(*amount, *destination)
	}

	fmt.Println("downloaded ", downloaded, " images")
}

func getDownloaders(sources []source, downloaderBits int) []downloader {
	var downloaders = make([]downloader, len(sources))
	for _, current := range sources {
		if ((downloaderBits) & current.Id) != 0 {
			downloader := downloadersMap[current.Id]
			if current.ListUrl == "" {
				log.Fatal("Source ", current.Id, " is not properly configured")
			}
			downloaders = append(downloaders, downloader)
		}
	}
	return downloaders
}

func validateFlags() {
	if *destination == "" {
		log.Fatalln("destination must be set")
	}

	if _, err := os.Stat(*destination); err != nil {
		log.Fatalln(err)
	}

	if *downloaderBits <= 0 {
		log.Fatalln("sources must be greater than 0")
	}

	if *amount <= 0 {
		log.Fatalln("amount must be greater than 0")
	}
}

type config struct {
	Sources []source `xml:"source"`
}

type source struct {
	Id      int    `xml:"id,attr"`
	ListUrl string `xml:"listUrl"`
}

func parseConfig(configFile string) config {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	var conf config
	err = xml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal(err)
	}
	return conf
}
