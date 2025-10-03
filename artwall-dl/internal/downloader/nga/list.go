package nga

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/zzvanq/artwall-dl/internal/http"
)

func parseArtUrls(ctx context.Context, listUrl string) ([]string, string) {
	nextListUrl := buildNextListUrl(listUrl)

	body, err := http.GetBody(ctx, listUrl)
	if err != nil {
		log.Println("parseListUrl: ", err)
		return nil, ""
	}

	body = artsSectorRe.Find(body)
	if body == nil {
		log.Println("parseListUrl: no images sector found")
		return nil, ""
	}

	arts := artSectorRe.FindAll(body, -1)
	urls := make([]string, 0, len(arts))
	for _, art := range arts {
		match := artUrlRe.Find(art)
		urls = append(urls, string(match))
	}

	return urls, nextListUrl
}

func buildNextListUrl(listUrl string) string {
	u, err := url.Parse(listUrl)
	if err != nil {
		panic(fmt.Errorf("failed to parse url: %s \n error: %w", listUrl, err))
	}

	page := 1
	if pageStr := u.Query().Get(pageParam); pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	q := u.Query()
	q.Set(pageParam, strconv.Itoa(page+1))
	u.RawQuery = q.Encode()
	return u.String()
}
