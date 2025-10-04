package dl

import (
	"errors"

	"github.com/zzvanq/artwall-dl/internal/dl/nga"
)

var ErrUnsupportedDownloader = errors.New("unknown downloader")

type Downloader interface {
	Download(int) int
}

func NewDownloader(listUrl, dlKey string, destination string) (Downloader, error) {
	switch dlKey {
	case nga.DlKey:
		return nga.NewNgaDl(listUrl, destination), nil
	default:
		return nil, ErrUnsupportedDownloader
	}
}
