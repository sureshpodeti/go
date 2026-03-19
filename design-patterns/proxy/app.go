package proxy

import "fmt"

type Application struct{}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) HandleRequest(url, method string) (int, string) {
	fmt.Printf("Application: handling %s %s\n", method, url)
	return 200, "success"
}
