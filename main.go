package main

import (
	"github.com/blablacar/ggn/commands"
	_ "github.com/n0rad/go-erlog/register"
	"math/rand"
	"time"
)

var CommitHash string
var Version string
var BuildTime string

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	commands.Execute(CommitHash, Version, BuildTime)
}
