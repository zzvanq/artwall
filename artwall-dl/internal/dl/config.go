package dl

import (
	"encoding/xml"
	"log"
	"os"
)

type Config struct {
	Sources []Source `xml:"source"`
}

type Source struct {
	Id      int    `xml:"id,attr"`
	ListUrl string `xml:"listUrl"`
}

func ParseConfig(configFile string) Config {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	var conf Config
	if err = xml.Unmarshal(data, &conf); err != nil {
		log.Fatal(err)
	}
	return conf
}
