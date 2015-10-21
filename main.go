package main

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/logger"
	"github.com/blablacar/green-garden/commands"
)

//go:generate go run compile/info_generate.go
func main() {
	log.Logger = logger.NewLogger()
	commands.Execute()
}
