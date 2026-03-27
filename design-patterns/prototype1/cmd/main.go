package main

import (
	"designpatterns/prototype1/prototype"
	"designpatterns/prototype1/registry"
	"designpatterns/prototype1/service"
	"fmt"
)

func main() {
	registry := registry.NewDocumentRegistry()

	registry.Register("quarterly-report", &prototype.TemplatedDocument{
		Document: prototype.Document{
			Title:    "Quartely Report",
			Content:  "",
			Author:   "",
			FontSize: 12,
		},
		TemplateName: "corporate-report",
		Version:      2,
	})

	registry.Register("welcome-letter", &prototype.Document{
		Title: "Welcome Letter", Content: "", Author: "HR Team", FontSize: 14,
	})

	registry.Register("nda", &prototype.TemplatedDocument{
		Document:     prototype.Document{Title: "NDA", Content: "", Author: "Legal Dept", FontSize: 11},
		TemplateName: "legal-standard",
		Version:      5,
	})

	//services get the registry - that's all they need
	reportSvc := &service.ReportService{Registry: registry}
	onBoardingSvc := &service.OnboardingService{Registry: registry}
	legalSvc := &service.LegalService{Registry: registry}

	//Each service clones and customerizes
	report := reportSvc.GenerateQuartelyReport("Q4 2025", "Alice")
	fmt.Printf("Report: %+v\n\n", report)

	welcome := onBoardingSvc.CreateWelcomePacket("Bob")
	fmt.Printf("Welcome: %+v\n\n", welcome)

	nda := legalSvc.CreateNDA("Acme Corp")
	fmt.Printf("NDA: %+v\n\n", nda)

	// Prove they're independent copies
	report2 := reportSvc.GenerateQuartelyReport("Q1 2026", "Charlie")
	fmt.Printf("Report2: %+v\n", report2)
	fmt.Printf("Report1 unchanged: %+v\n", report)

}
