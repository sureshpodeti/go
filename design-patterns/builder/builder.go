package builder

import "designpatterns/builder/constant"

type IBuilder interface {
	setWindowType()
	setDoorType()
	setFloor()
	getHouse() House
}

func GetBuilder(builderType string) IBuilder {
	var builder IBuilder
	switch builderType {
	case constant.IGLOO:
		builder = newIglooBuilder()
	case constant.NORMAL:
		builder = newNormalBuilder()
	}
	return builder
}
