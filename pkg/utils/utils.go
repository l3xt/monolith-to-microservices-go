package utils

import (
	"bytes"
	"io"
	"mime"
	"net/http"
)

func DetectContentType(file io.Reader) (string, io.Reader, error) {
	// Буфер 512 байт
	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", nil, err
	}
	// Обрезаем буфер до n
	contentType := http.DetectContentType(buffer[:n])

	// Склеиваем обратно
	newReader := io.MultiReader(bytes.NewReader(buffer[:n]), file)
	return contentType, newReader, nil
}

func MapContentTypeToExt(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
			return exts[0]
		}
		return ""
	}
}
