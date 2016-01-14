package main

import (
	"github.com/blablacar/ggn/commands"
	_ "github.com/n0rad/go-erlog/register"
)

//go:generate go run compile/version_generate.go
func main() {
	commands.Execute()
}
