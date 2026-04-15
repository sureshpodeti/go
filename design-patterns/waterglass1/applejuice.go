package waterglass1

type AppleJuice struct {
	priceForML float64
}

func (aplj *AppleJuice) GetUnitPrice() float64 {
	return aplj.priceForML
}
