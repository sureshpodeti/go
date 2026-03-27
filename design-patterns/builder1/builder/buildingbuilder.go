package builder

type BuildingBuilder interface {
	Reset()
	BuildWalls()
	BuildRoof()
	BuildDoor()
}
