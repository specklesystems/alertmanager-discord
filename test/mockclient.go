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

type RequestToMockClient struct {
	Url         string
	ContentType string
	Body        io.Reader
}

type MockClientRecorder struct {
	Requests []RequestToMockClient
}

func (mc MockClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return mc.RequestHandler(url, contentType, body)
}

func (mc *MockClientRecorder) NewMockClientWithResponse(statusCode int, errorHandlerShouldReturn error) MockClient {
	return MockClient{
		RequestHandler: func(url, contentType string, requestBody io.Reader) (resp *http.Response, err error) {
			mc.Requests = append(mc.Requests, RequestToMockClient{
				Url:         url,
				ContentType: contentType,
				Body:        requestBody,
			})

			resp = &http.Response{
				StatusCode: statusCode,
				// Body: io.NopCloser(bytes.NewBufferString(responseBody)),
			}
			return resp, errorHandlerShouldReturn
		},
	}
}
