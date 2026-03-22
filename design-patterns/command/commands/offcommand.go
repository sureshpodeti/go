package commands

type OffCommand struct {
	tv Device
}

func NewOffCommand(tv Device) *OffCommand {
	return &OffCommand{tv: tv}
}

func (off *OffCommand) Execute() {
	off.tv.Off()
}
