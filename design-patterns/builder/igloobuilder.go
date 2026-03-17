package builder

type IglooBuilder struct {
	windowType string
	doorType   string
	floor      int
}

func newIglooBuilder() *IglooBuilder {
	return &IglooBuilder{}
}

func (igloo *IglooBuilder) setWindowType() {
	igloo.windowType = "snow window"
}

func (igloo *IglooBuilder) setDoorType() {
	igloo.doorType = "snow door"
}

func (igloo *IglooBuilder) setFloor() {
	igloo.floor = 1

}

func (igloo *IglooBuilder) getHouse() House {
	return House{
		WindowType: igloo.windowType,
		DoorType:   igloo.doorType,
		Floor:      igloo.floor,
	}

}
