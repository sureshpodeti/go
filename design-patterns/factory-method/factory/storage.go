package factory

import (
	"designpatterns/factory-method/constant"
	"designpatterns/factory-method/storage"
)

type Storage interface {
	Upload()
}

func GetStorage(provider string) Storage {
	switch provider {
	case constant.AWS:
		return &storage.S3{}
	case constant.GCP:
		return &storage.GCS{}
	default:
		return &storage.Local{}
	}

}
