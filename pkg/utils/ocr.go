package utils

import (
	"io"

	"github.com/otiai10/gosseract/v2"
)

func RunTesseract(r io.Reader) (string, error) {
	client := gosseract.NewClient()

	client.SetLanguage("ita", "eng")
	client.SetPageSegMode(gosseract.PSM_SPARSE_TEXT)
	client.SetVariable("load_system_dawg", "0")
	client.SetVariable("load_freq_dawg", "0")
	client.SetVariable("load_punc_dawg", "0")
	client.SetVariable("load_number_dawg", "0")

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
