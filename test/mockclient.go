package test

import (
	// "bytes"
	"io"
	"net/http"
)

type mockClientRequestHandler func(url, contentType string, body io.Reader) (resp *http.Response, err error)

type MockClient struct {
	RequestHandler mockClientRequestHandler
}

type MockClientRecorder struct {
	ClientTriggered bool
	Url             string
	ContentType     string
	Body            io.Reader
}

func (mc MockClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return mc.RequestHandler(url, contentType, body)
}

func (mc *MockClientRecorder) NewMockClientWithResponse(statusCode int) MockClient {
	return MockClient{
		RequestHandler: func(url, contentType string, requestBody io.Reader) (resp *http.Response, err error) {
			mc.ClientTriggered = true
			mc.Url = url
			mc.ContentType = contentType
			mc.Body = requestBody

			resp = &http.Response{
				StatusCode: statusCode,
				// Body: io.NopCloser(bytes.NewBufferString(responseBody)),
			}
			return resp, nil
		},
	}
}
