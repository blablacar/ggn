package main

import (
	"github.com/blablacar/green-garden/commands"
)

//go:generate go run compile/info_generate.go
func main() {
	commands.Execute()
}
