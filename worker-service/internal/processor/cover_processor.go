package processor

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

type CoverProcessor struct {
	coverWidth  int
	coverHeight int
	thumbWidth  int
	thumbHeight int
}

func NewCoverProcessor(cw, ch, tw, th int) *CoverProcessor {
	return &CoverProcessor{
		coverWidth:  cw,
		coverHeight: ch,
		thumbWidth:  tw,
		thumbHeight: th,
	}
}

func (p *CoverProcessor) CreateCover(in io.Reader) ([]byte, error) {
	res, err := p.processImage(in, p.coverWidth, p.coverHeight)
	if err != nil {
		return nil, fmt.Errorf("CoverProcessor.CreateCover: process image: %w", err)
	}

	return res, nil
}

func (p *CoverProcessor) CreateThumbnail(in io.Reader) ([]byte, error) {
	res, err := p.processImage(in, p.thumbWidth, p.thumbHeight)
	if err != nil {
		return nil, fmt.Errorf("CoverProcessor.CreateThumbnail: process image: %w", err)
	}

	return res, nil
}

// Внутренняя приватная функция, инкапсулирующая логику библиотеки imaging
func (p *CoverProcessor) processImage(in io.Reader, width, height int) ([]byte, error) {
	img, err := imaging.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("imaging decode failed: %w", err)
	}

	// Делаем ресайз с сохранением пропорций
	resizedImg := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	// Кодируем обратно в JPEG и пишем в буфер
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("jpeg encode failed: %w", err)
	}

	return buf.Bytes(), nil
}
