package service

import (
	"designpatterns/prototype1/prototype"
	"designpatterns/prototype1/registry"
)

type OnboardingService struct {
	Registry *registry.DocumentRegistry
}

func (o *OnboardingService) CreateWelcomePacket(employee string) *prototype.Document {
	doc := o.Registry.Get("welcome-letter").(*prototype.Document)
	doc.Title = "Welcome. " + employee
	doc.Content = "We're glad to have you on the team." + employee
	return doc
}
