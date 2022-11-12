package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HttpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type Client struct {
	httpClient HttpClient
	URL        string
}

func NewClient(client HttpClient, url string) *Client {
	return &Client{
		httpClient: client,
		URL:        url,
	}
}

func (dc *Client) PublishMessage(message Out) (*http.Response, error) {
	DOD, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("Error encountered when marshalling object to json. We will not continue posting to Discord. Discord Out object: '%v+'. Error: %w", message, err)
	}

	res, err := dc.httpClient.Post(dc.URL, "application/json", bytes.NewReader(DOD))
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("Error encountered sending POST to '%s'. Error: %w", dc.URL, err)
	}

	return res, nil
}
