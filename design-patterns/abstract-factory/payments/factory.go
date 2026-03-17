package payments

type Processor interface {
	Process(amount float64) string
}

type Refunder interface {
	Refund(amount float64) string
}

type ReceiptGenerator interface {
	Generate(amount float64) string
}

type Factory interface {
	CreateProcessor() Processor
	CreateRefunder() Refunder
	CreateReceiptGenerator() ReceiptGenerator
}
