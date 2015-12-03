package ggn

import (
	"os"
)

var Home HomeStruct
var CommitHash = ""
var Version = "DEV"
var BuildDate = ""
var PathSkip = 0

func GetUserAndHost() string {
	user := os.Getenv("USER")
	if Home.Config.User != "" {
		user = Home.Config.User
	}
	hostname, _ := os.Hostname()
	return user + "@" + hostname
}
