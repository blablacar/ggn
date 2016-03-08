package main

import (
	"github.com/blablacar/ggn/commands"
	_ "github.com/n0rad/go-erlog/register"
)

var CommitHash string
var GgnVersion string
var BuildDate string

func main() {
	commands.Execute(CommitHash, GgnVersion, BuildDate)
}
