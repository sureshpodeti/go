package proxy

type Server interface {
	HandleRequest(url, method string) (int, string)
}
