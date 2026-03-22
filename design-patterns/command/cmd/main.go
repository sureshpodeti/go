package main

import (
	"designpatterns/command"
	"designpatterns/command/commands"
	"designpatterns/command/device"
)

func main() {
	tv := device.NewTv()

	// offCommand := commands.NewOffCommand(tv)
	onCommand := commands.NewOnCommand(tv)

	button := command.NewButton(onCommand)

	button.Press()
}
