package proxy

import "fmt"

type Nginx struct {
	app Server
}

func NewNginx(app Server) *Nginx {
	return &Nginx{app: app}
}

func (n *Nginx) HandleRequest(url, method string) (int, string) {
	fmt.Printf("Nginx: checking rate limit for %s %s\n", method, url)
	fmt.Printf("Nginx: filtering URL %s\n", url)
	fmt.Printf("Nginx: forwarding request to application\n")
	return n.app.HandleRequest(url, method)
}
