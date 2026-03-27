package service

import (
	"designpatterns/prototype1/prototype"
	"designpatterns/prototype1/registry"
)

type LegalService struct {
	Registry *registry.DocumentRegistry
}

func (l *LegalService) CreateNDA(partyName string) *prototype.TemplatedDocument {
	doc := l.Registry.Get("nda").(*prototype.TemplatedDocument)
	doc.Title = "NDA - " + partyName
	doc.Content = "Non-discloure agreement between company and " + partyName
	return doc

}
