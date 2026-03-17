package builder

type NormalBuilder struct {
	windowType string
	doorType   string
	floor      int
}

func newNormalBuilder() *NormalBuilder {
	return &NormalBuilder{}
}

func (normal *NormalBuilder) setWindowType() {
	normal.windowType = "wooden window"
}

func (normal *NormalBuilder) setDoorType() {
	normal.doorType = "wooden door"
}

func (normal *NormalBuilder) setFloor() {
	normal.floor = 2
}

func (normal *NormalBuilder) getHouse() House {
	return House{
		WindowType: normal.windowType,
		DoorType:   normal.doorType,
		Floor:      normal.floor,
	}
}
