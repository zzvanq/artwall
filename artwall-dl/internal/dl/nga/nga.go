package nga

// national gallery of art

import (
	"context"
	"log"
	"sync"
	"time"
)

const baseUrl = "https://www.nga.gov"

const (
	DlKey             = "nga"
	itemWorkersAmount = 10
	dlTimeout         = 45 * time.Second
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

	workers := min(amount, itemWorkersAmount)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			dl.itemWorker(ctx, urlsCh, resCh)
		}()
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	go dl.itemUrlsWorker(ctx, amount, urlsCh)

	downloaded := 0
	for range resCh {
		downloaded++
		if downloaded >= amount {
			return downloaded
		}
	}

	return downloaded
}

func (dl NgaDl) itemUrlsWorker(ctx context.Context, amount int, urlsCh chan<- string) {
	defer func() {
		close(urlsCh)
	}()

	var itemUrls []string
	itemsLeft := amount
	nextListUrl := dl.listUrl
	for {
		itemUrls, nextListUrl = parseItemUrls(ctx, nextListUrl)
		if len(itemUrls) == 0 {
			return
		}

		for _, url := range itemUrls {
			if itemsLeft == 0 {
				return
			}
			select {
			case <-ctx.Done():
				return
			case urlsCh <- url:
				itemsLeft -= 1
			}
		}
	}
}

func (dl NgaDl) itemWorker(ctx context.Context, urlsCh <-chan string, resCh chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case url, ok := <-urlsCh:
			if !ok {
				return
			}
			if err := downloadItem(ctx, dl.destination, url); err != nil {
				log.Println("itemWorker: ", err)
				continue
			}
			resCh <- struct{}{}
		}
	}
}
