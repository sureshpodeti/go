package builder

import "designpatterns/builder1/product"

type HouseBuilder struct {
	House *product.House
}

func (h *HouseBuilder) Reset() { h.House = &product.House{} }

func (h *HouseBuilder) BuildWalls() { h.House.Bricks = 2000 }

func (h *HouseBuilder) BuildRoof() { h.House.RoofMaterial = "tile" }

func (h *HouseBuilder) BuildDoor() { h.House.DoorType = "wooden" }

func (h *HouseBuilder) GetHouse() *product.House { return h.House }
