package main

import (
	"designpatterns/factory-method/constant"
	"designpatterns/factory-method/factory"
)

func main() {
	cloudstorage := factory.GetStorage(constant.AWS)

	cloudstorage.Upload()

	cloudstorage = factory.GetStorage(constant.GCP)
	cloudstorage.Upload()

}
