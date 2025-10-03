package nga

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"

	"github.com/zzvanq/artwall-dl/internal/http"
)

const pageParam = "page"

var (
	itemNoImgRe = regexp.MustCompile(`framed_artwork_fallback\.jpg`)
	itemUrlRe   = regexp.MustCompile(`/artworks/[^"]*`)
)

func parseItemUrls(ctx context.Context, listUrl string) ([]string, string) {
	nextListUrl := buildNextListUrl(listUrl)

	body, err := http.GetBody(ctx, listUrl)
	if err != nil {
		log.Println("parseListUrl: ", err)
		return nil, ""
	}

	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Println("parseListUrl: failed to unmarshal json: ", err)
		return nil, ""
	}

	rows, ok := resp["rows"].([]any)
	if !ok || len(rows) <= 0 {
		log.Println("parseListUrl: no rows found")
		return nil, ""
	}

	urls := make([]string, 0, len(rows))
	for _, row := range rows {
		result, ok := row.(map[string]any)["result"].(string)
		if !ok {
			log.Println("parseListUrl: failed to parse row: ", row)
			continue
		}

		if itemNoImgRe.MatchString(result) {
			continue
		}

		match := itemUrlRe.FindString(result)
		if match == "" {
			log.Println("parseListUrl: failed to parse item url: ", result)
			continue
		}

		urls = append(urls, baseUrl+match)
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
