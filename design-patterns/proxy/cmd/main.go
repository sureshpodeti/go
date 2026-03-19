package main

import (
	"designpatterns/proxy"
	"fmt"
)

func main() {
	app := proxy.NewApplication()
	nginx := proxy.NewNginx(app)

	code, body := nginx.HandleRequest("/api/users", "GET")
	fmt.Printf("Response: %d %s\n\n", code, body)

	code, body = nginx.HandleRequest("/api/orders", "POST")
	fmt.Printf("Response: %d %s\n", code, body)
}
