package dl

import (
	"log"

	"github.com/zzvanq/artwall-dl/internal/dl/nga"
)

type Downloader interface {
	Download(int) int
}

// keys are single bits. 1 = nga, 2 = wallhaven, 4 = ...
var downloadersMap = map[int]func(Source, string) Downloader{
	1: func(src Source, destination string) Downloader { return nga.NewNgaDl(src.ListUrl, destination) },
	// 2: func(src source) downloader { return wallhavenDl{src.ListUrl, *destination} },
	// 4: func(src source) downloader { return imgurDl{src.ListUrl, *destination} },
}

func GetDownloaders(sources []Source, downloaderBits int, destination string) []Downloader {
	var downloaders = make([]Downloader, 0, len(sources))
	for _, src := range sources {
		if ((downloaderBits) & src.Id) != 0 {
			if src.ListUrl == "" {
				log.Fatal("Source ", src.Id, " is not properly configured")
			}
			dl, ok := downloadersMap[src.Id]
			if !ok {
				continue
			}
			downloaders = append(downloaders, dl(src, destination))
		}
	}
	return downloaders
}
