package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CardInfoResponse struct {
	Data []CardInfoData `json:"data"`
}

type CardInfoData struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	CardImages []CardImageData `json:"card_images"`
}

type CardImageData struct {
	ImageURL      string `json:"image_url"`
	ImageURLSmall string `json:"image_url_small"`
}

func GetCardInfo(cardNames []string) (*CardInfoResponse, error) {
	cardNamesParam := url.QueryEscape(strings.Join(cardNames, "|"))
	response, err := http.Get(fmt.Sprintf("https://db.ygoprodeck.com/api/v7/cardinfo.php?name=%s", cardNamesParam))
	if err != nil {
		return nil, err
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	cardInfoResponse := CardInfoResponse{}
	err = json.Unmarshal(responseBytes, &cardInfoResponse)
	if err != nil {
		return nil, err
	}

	return &cardInfoResponse, nil
}
