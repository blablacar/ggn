package application

import (
	"github.com/blablacar/ggn/config"
	"os"
)

var CommitHash = ""
var Version = "DEV"
var BuildDate = ""
var PathSkip = 0

func GetUserAndHost() string {
	user := os.Getenv("USER")
	if config.GetConfig().User != "" {
		user = config.GetConfig().User
	}
	hostname, _ := os.Hostname()
	return user + "@" + hostname
}
