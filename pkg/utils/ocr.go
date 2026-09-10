package utils

import (
	"io"

	"github.com/otiai10/gosseract/v2"
)

func RunTesseract(r io.Reader) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	imgBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	if err := client.SetImageFromBytes(imgBytes); err != nil {
		return "", err
	}

	textOut, err := client.Text()
	if err != nil {
		return "", err
	}

	return textOut, nil
}

// func RunTesseract(r io.Reader) ([]byte, error) {
// 	client := gosseract.NewClient()
// 	defer client.Close()

// 	imgBytes, err := io.ReadAll(r)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := client.SetImageFromBytes(imgBytes); err != nil {
// 		return nil, err
// 	}

// 	textOut, err := client.Text()
// 	if err != nil {
// 		return nil, err
// 	}

// 	pr, pw := io.Pipe()

// 	go func() {
// 		b64 := base64.NewEncoder(base64.RawStdEncoding, pw)

// 		zw, err := zstd.NewWriter(b64)
// 		if err != nil {
// 			pw.CloseWithError(fmt.Errorf("zstd writer: %w", err))
// 			return
// 		}

// 		if _, err := zw.Write([]byte(textOut)); err != nil {
// 			pw.CloseWithError(fmt.Errorf("zstd write: %w", err))
// 			return
// 		}

// 		if err := zw.Close(); err != nil {
// 			pw.CloseWithError(fmt.Errorf("zstd close: %w", err))
// 			return
// 		}

// 		if err := b64.Close(); err != nil {
// 			pw.CloseWithError(fmt.Errorf("base64 close: %w", err))
// 			return
// 		}

// 		pw.Close()
// 	}()

// 	compressed, err := io.ReadAll(pr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return compressed, nil
// }
