package client

import "fmt"

type Client interface {
	Get(url string) string
}

type httpClient struct{}

func (c *httpClient) Get(url string) string {
	return fmt.Sprintf("http response from %s", url)
}

type mockClient struct{}

func (m *mockClient) Get(url string) string {
	return fmt.Sprintf("mock response from %s", url)
}

func NewClient(useMock bool) Client {
	if useMock {
		return &mockClient{}
	}
	return &httpClient{}
}
