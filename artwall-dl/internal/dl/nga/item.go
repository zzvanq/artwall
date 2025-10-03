package nga

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zzvanq/artwall-dl/internal/http"
	"github.com/zzvanq/tinymedia/pkg/file"
	"github.com/zzvanq/tinymedia/pkg/meta/codec"
	"github.com/zzvanq/tinymedia/pkg/meta/manager"
)

const (
	maxResString     = "9999,"
	fileNameFormat   = "%s - %s.jpg"
	imgUrlResPattern = `full/%s/0/`
)

var (
	imgUrlRe              = regexp.MustCompile(`c-artwork-media-single__media-element-inner">[\s\S]*?<img src="(.*?)"`)
	imgUrlResRe           = regexp.MustCompile(fmt.Sprintf(imgUrlResPattern, ".*?"))
	titleSectorRe         = regexp.MustCompile(`c-artwork-header__title">([\s\S]*?)</p>\s*?</div`)
	descSectorRe          = regexp.MustCompile(`u-wysiwyg--small">\s*(.*?)(</p>)?\s*</div>`)
	htmlTagRe             = regexp.MustCompile(`<.*?>`)
	htmlParagraphTagEndRe = regexp.MustCompile(`</p>`)
	bigSpacesRe           = regexp.MustCompile(`\s{2,}`)
)

type imgMeta struct {
	author      string
	title       string
	year        string
	description string
}

func downloadItem(ctx context.Context, destination, url string) error {
	pageBody, err := http.GetBody(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch %s failed: %w", url, err)
	}

	meta, err := parseItemMeta(pageBody)
	if err != nil {
		return fmt.Errorf("meta parsing failed for item '%s': %w", url, err)
	}

	imgUrl, err := parseImgUrl(pageBody)
	if err != nil {
		return fmt.Errorf("img url parsing failed for item '%s': %w", url, err)
	}

	if err := downloadAndTagImg(ctx, destination, imgUrl, meta); err != nil {
		return fmt.Errorf("img downloading failed for item '%s': %w", url, err)
	}
	return nil
}

func parseItemMeta(pageBody []byte) (*imgMeta, error) {
	titleSectorMatches := titleSectorRe.FindSubmatch(pageBody)
	if len(titleSectorMatches) < 2 {
		return nil, fmt.Errorf("title of the item is not found")
	}

	titleSector := string(titleSectorMatches[1])
	titleSector = htmlTagRe.ReplaceAllString(titleSector, "")
	titleSector = bigSpacesRe.ReplaceAllString(titleSector, "\n")
	titleSectorSplit := strings.Split(titleSector, "\n")
	if len(titleSectorSplit) != 3 {
		return nil, fmt.Errorf("incorrect title sector format: %s", titleSector)
	}

	title, year, author := titleSectorSplit[0], titleSectorSplit[1], titleSectorSplit[2]
	var description string

	descMatches := descSectorRe.FindSubmatch(pageBody)
	if len(descMatches) >= 2 {
		descSector := string(descMatches[1])
		descTagEndReplaced := htmlParagraphTagEndRe.ReplaceAllString(descSector, "\n")
		description = htmlTagRe.ReplaceAllString(descTagEndReplaced, "")
	}
	return &imgMeta{author, title, year, description}, nil
}

func parseImgUrl(sector []byte) (string, error) {
	imgUrlMatches := imgUrlRe.FindSubmatch(sector)
	if len(imgUrlMatches) < 2 {
		return "", fmt.Errorf("img url not found")
	}
	imgUrlOrigRes := string(imgUrlMatches[1])
	return imgUrlResRe.ReplaceAllString(imgUrlOrigRes, fmt.Sprintf(imgUrlResPattern, maxResString)), nil
}

func downloadAndTagImg(ctx context.Context, destination, imgUrl string, meta *imgMeta) error {
	imgBody, err := http.GetBodyReader(ctx, imgUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch the image: %w", err)
	}
	defer imgBody.Close()

	r, ftype, err := file.ReadFileType(imgBody)
	if err != nil {
		return fmt.Errorf("failed to verify the image type: %w", err)
	}
	if ftype != file.FileTypeJPEG {
		return fmt.Errorf("filetype is not a jpeg")
	}

	m, err := manager.NewMetaManager(r, ftype)
	if err != nil {
		return fmt.Errorf("failed to initialize metadata manager: %w", err)
	}
	fields := map[string]string{
		"title":       meta.title,
		"year":        meta.year,
		"author":      meta.author,
		"description": meta.description,
	}
	if err := m.Insert(codec.TinyMetaGzipVendor, fields); err != nil {
		return fmt.Errorf("failed to set the metadata: %w", err)
	}

	fn := fmt.Sprintf(fileNameFormat, meta.author, meta.title)
	return saveImg(m.FileReader(), destination, fn)
}

func saveImg(r io.Reader, destination string, fileName string) error {
	f, err := os.Create(filepath.Join(destination, fileName))
	if err != nil {
		return fmt.Errorf("failed to create a file: %w", err)
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("failed to fill a file: %w", err)
	}
	return nil
}
