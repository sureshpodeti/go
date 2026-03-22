package commands

type OnCommand struct {
	tv Device
}

func NewOnCommand(tv Device) *OnCommand {
	return &OnCommand{tv: tv}
}

func (on *OnCommand) Execute() {
	on.tv.On()
}
