package nga

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zzvanq/artwall-dl/internal/http"
	"github.com/zzvanq/tinymedia/pkg/file"
	"github.com/zzvanq/tinymedia/pkg/meta/codec"
	"github.com/zzvanq/tinymedia/pkg/meta/manager"
)

type imgMeta struct {
	author      string
	title       string
	year        string
	description string
}

func downloadItem(ctx context.Context, destination, artUrl string) error {
	pageBody, err := http.GetBody(ctx, artUrl)
	if err != nil {
		return fmt.Errorf("fetch %s failed: %w", artUrl, err)
	}

	meta, err := parseImgMeta(pageBody)
	if err != nil {
		return fmt.Errorf("meta parsing faile for item '%s': %w", artUrl, err)
	}

	imgUrl, err := parseImgUrl(pageBody)
	if err != nil {
		return fmt.Errorf("img url parsing failed for item '%s': %w", artUrl, err)
	}

	if err := downloadAndTagImg(ctx, destination, imgUrl, meta); err != nil {
		return fmt.Errorf("img downloading failed for item '%s': %w", artUrl, err)
	}
	return nil
}

func parseImgMeta(pageBody []byte) (*imgMeta, error) {
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
		descTagEndReplaced := htmlTagEndRe.ReplaceAllString(descSector, "\n")
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

	var buf bytes.Buffer
	tr := io.TeeReader(imgBody, &buf)
	ftype, err := file.ReadFileType(tr)
	if err != nil {
		return fmt.Errorf("failed to verify the image type: %w", err)
	}
	if ftype != file.FileTypeJPEG {
		panic("artWorker: filetype is not a jpeg")
	}

	r := io.MultiReader(&buf, imgBody)
	m, err := manager.NewMetaManager(r)
	if err != nil {
		return fmt.Errorf("failed to initialize metadata manager: %w", err)
	}

	fields := map[string]string{
		"title":       meta.title,
		"year":        meta.year,
		"author":      meta.author,
		"description": meta.description,
	}
	if err := m.Insert(codec.TinyMetaVendor, fields); err != nil {
		return fmt.Errorf("failed to set the metadata: %w", err)
	}

	fn := fmt.Sprintf(fileNameFormat, meta.author, meta.title)
	f, err := os.Create(filepath.Join(destination, fn))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		panic(err)
	}
	return nil
}
