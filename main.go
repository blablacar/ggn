package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/green-garden/commands"
)

//go:generate go run compile/version_generate.go
func main() {
	logrus.SetFormatter(&log.BlaFormatter{})
	commands.Execute()
}
