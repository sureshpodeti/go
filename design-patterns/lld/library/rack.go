package library

// Rack represents a physical location in the library
// Rack is just a location identifier - it doesn't "own" books
// BookItem will reference which Rack it's located at
type Rack struct {
	Number   string // e.g., "A-101", "B-205"
	Location string // e.g., "Floor 1, Section A"
}

func NewRack(number, location string) *Rack {
	return &Rack{
		Number:   number,
		Location: location,
	}
}
