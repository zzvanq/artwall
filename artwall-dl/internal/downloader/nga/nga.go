package nga

// national gallery of art

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"
)

const (
	maxResString     = "9999,"
	fileNameFormat   = "%s - %s"
	imgUrlResPattern = `full/%s/0/`
	pageParam        = "page"
)

var artWorkersAmount = 10
var dlTimeout = 30 * time.Second

var (
	artsSectorRe  = regexp.MustCompile(`aws-filter-display[\s\S]+?js-a11y`)
	artSectorRe   = regexp.MustCompile(`/full/(?!full).*?/0/default\.jpg.*?u-hover-list__title" data-part="link"`)
	artUrlRe      = regexp.MustCompile(`/artworks/[^"]*`)
	imgUrlRe      = regexp.MustCompile(`c-artwork-media-single__media-element-inner">.*?<img src="(.*?)"`)
	imgUrlResRe   = regexp.MustCompile(fmt.Sprintf(imgUrlResPattern, ".*?"))
	titleSectorRe = regexp.MustCompile(`c-artwork-header__title">(.*?)</p>\s*?</div`)
	descSectorRe  = regexp.MustCompile(`u-wysiwyg--small">\s*(.*?)(</p>)?\s*</div>`)
	htmlTagRe     = regexp.MustCompile(`<.*?>`)
	htmlTagEndRe  = regexp.MustCompile(`</.*?>`)
	bigSpacesRe   = regexp.MustCompile(`\n{2,}`)
)

type NgaDl struct {
	listUrl     string
	destination string
}

func NewNgaDl(listUrl, destination string) NgaDl {
	return NgaDl{listUrl, destination}
}

func (dl NgaDl) Download(amount int) int {
	urlsCh := make(chan string, amount)
	resCh := make(chan struct{}, amount)

	ctx, cancel := context.WithTimeout(context.Background(), dlTimeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(artWorkersAmount)
	for range artWorkersAmount {
		go func() {
			defer wg.Done()
			dl.artWorker(ctx, urlsCh, resCh)
		}()
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	go dl.artUrlsWorker(ctx, amount, urlsCh)

	downloaded := 0
	for range resCh {
		downloaded++
		if downloaded >= amount {
			return downloaded
		}
	}

	return downloaded
}

func (dl NgaDl) artUrlsWorker(ctx context.Context, amount int, urlsCh chan<- string) {
	defer func() { close(urlsCh) }()

	var artUrls []string
	artsLeft := amount
	nextListUrl := dl.listUrl
	for {
		// TODO: parallelize parsing of urls. e.g. first 3 pages
		artUrls, nextListUrl = parseArtUrls(ctx, nextListUrl)
		if len(artUrls) == 0 {
			return
		}

		for _, url := range artUrls {
			if artsLeft == 0 {
				return
			}
			select {
			case <-ctx.Done():
				return
			case urlsCh <- url:
				artsLeft -= 1
			}
		}
	}
}

func (dl NgaDl) artWorker(ctx context.Context, artUrlsCh <-chan string, resCh chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case artUrl := <-artUrlsCh:
			if err := downloadItem(ctx, artUrl, dl.destination); err != nil {
				log.Println("artWorker: ", err)
				continue
			}
			resCh <- struct{}{}
		}
	}
}
