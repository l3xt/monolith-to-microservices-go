package utils

import (
	"bytes"
	"io"
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
