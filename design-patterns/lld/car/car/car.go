package car

type CarState interface {
	EngineOff()
	EngineOn()
	Accelerate()
	ApplyBrake()
	RelaseBrake()
}

type CarType string

const (
	SEDAN     CarType = "SEDAN"
	SUV       CarType = "SUV"
	HATCHBACK CarType = "HATCHBACK"
)

type CarColor string

const (
	BLUE  CarColor = "BLUE"
	RED   CarColor = "RED"
	GREEN CarColor = "GREEN"
	WHITE CarColor = "WHITE"
	BLACK CarColor = "BLACK"
)

type Car struct {
	color   CarColor
	state   CarState
	carType CarType

	offState    CarState
	onState     CarState
	movingState CarState
	stopState   CarState
}

func NewCar(colorColor CarColor, carType CarType) *Car {
	car := &Car{
		color:   colorColor,
		carType: carType,
	}

	car.offState = &Off{Car: car}
	car.onState = &On{Car: car}
	car.movingState = &Moving{Car: car}
	car.stopState = &Stop{Car: car}

	car.state = car.offState

	return car
}

func (car *Car) SetState(state CarState) {
	car.state = state
}

func (car *Car) EngineOff() {
	car.state.EngineOff()
}

func (car *Car) EngineOn() {
	car.state.EngineOn()
}

func (car *Car) Accelerate() {
	car.state.Accelerate()
}

func (car *Car) ApplyBrake() {
	car.state.ApplyBrake()
}

func (car *Car) RelaseBrake() {
	car.state.RelaseBrake()
}
