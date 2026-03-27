package service

import (
	"designpatterns/prototype1/prototype"
	"designpatterns/prototype1/registry"
)

type ReportService struct {
	Registry *registry.DocumentRegistry
}

func (s *ReportService) GenerateQuartelyReport(quarter, author string) *prototype.TemplatedDocument {
	// clone the template - don't build from scratch
	doc := s.Registry.Get("quarterly-report").(*prototype.TemplatedDocument)
	doc.Title = quarter + " Quartely Report"
	doc.Content = "Find Results for " + quarter
	doc.Author = author
	return doc
}
