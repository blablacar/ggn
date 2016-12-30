package main

import (
	"math/rand"
	"time"

	"github.com/blablacar/ggn/commands"
	_ "github.com/n0rad/go-erlog/register"
)

var BuildCommit string
var BuildVersion string
var BuildTime string

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	commands.Execute(BuildCommit, BuildVersion, BuildTime)
}
