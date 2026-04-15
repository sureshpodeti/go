package waterglass1

type Container interface {
	GetLiquid() Liquid
	GetLiquidVolume() float64
}
