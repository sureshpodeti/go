package flyweight

type TerroristDress struct {
	Color string
}

func NewTerroristDress() *TerroristDress {
	return &TerroristDress{Color: "red"}
}
