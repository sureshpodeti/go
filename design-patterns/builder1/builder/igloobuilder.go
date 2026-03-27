package builder

import "designpatterns/builder1/product"

type IglooBuilder struct {
	Igloo *product.Igloo
}

func (b *IglooBuilder) Reset() { b.Igloo = &product.Igloo{} }

func (b *IglooBuilder) BuildWalls() { b.Igloo.IceBlocks = 50 }

func (b *IglooBuilder) BuildRoof() { b.Igloo.Domeshape = "hemisphere" }

func (b *IglooBuilder) BuildDoor() { b.Igloo.HasTunnel = true }

func (b *IglooBuilder) GetHouse() *product.Igloo { return b.Igloo }
