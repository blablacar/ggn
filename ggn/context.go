package ggn

import (
	"os"
)

var Home HomeStruct

var CommitHash string
var GgnVersion string
var BuildDate string

func GetUserAndHost() string {
	user := os.Getenv("USER")
	if Home.Config.User != "" {
		user = Home.Config.User
	}
	hostname, _ := os.Hostname()
	return user + "@" + hostname
}
